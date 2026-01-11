// Copyright (C) 2026 Tom Cool
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package ui

import (
	"path/filepath"
	"time"

	"cooledit/internal/autosave"
	"cooledit/internal/config"
	"cooledit/internal/core"
	"cooledit/internal/fileio"
	"cooledit/internal/syntax"
	"cooledit/internal/term"
	"cooledit/internal/theme"
)

// SearchHistory maintains a history of search queries for navigation.
// This allows users to navigate through their previous searches using up/down arrow keys,
// similar to command-line history navigation.
//
// Features:
//   - Stores up to maxSize queries (typically 20)
//   - Automatically deduplicates consecutive identical queries
//   - Supports bidirectional navigation (up/down)
//   - Preserves current query when navigating (tempQuery)
//
// Usage pattern:
//  1. User types query and presses Enter
//  2. Add() is called to save query to history
//  3. In next search, user presses Up to navigate to previous query
//  4. First Prev() call stores current query in tempQuery
//  5. Subsequent Prev() calls move backwards through history
//  6. Next() calls move forwards, eventually returning to tempQuery
//  7. Reset() clears navigation state
type SearchHistory struct {
	queries   []string // Recent search queries (most recent last)
	index     int      // Current position in history (-1 if not navigating)
	maxSize   int      // Maximum number of queries to remember
	tempQuery string   // Temporary storage for current query when navigating
}

// NewSearchHistory creates a new search history with the given max size.
func NewSearchHistory(maxSize int) *SearchHistory {
	return &SearchHistory{
		queries: make([]string, 0),
		index:   -1,
		maxSize: maxSize,
	}
}

// Add adds a query to the history. Ignores empty queries and duplicates of the last query.
func (h *SearchHistory) Add(query string) {
	if query == "" {
		return
	}
	// Don't add if it's the same as the last query
	if len(h.queries) > 0 && h.queries[len(h.queries)-1] == query {
		return
	}
	h.queries = append(h.queries, query)
	// Trim if over limit
	if len(h.queries) > h.maxSize {
		h.queries = h.queries[1:]
	}
	// Reset navigation
	h.index = -1
}

// Prev returns the previous query in history, or empty string if at the beginning.
// On first call, stores the current query temporarily.
func (h *SearchHistory) Prev(currentQuery string) string {
	if len(h.queries) == 0 {
		return currentQuery
	}

	// First time navigating up? Store current query and start from end
	if h.index == -1 {
		h.tempQuery = currentQuery
		h.index = len(h.queries) - 1
		return h.queries[h.index]
	}

	// Already navigating - move backwards if possible
	if h.index > 0 {
		h.index--
		return h.queries[h.index]
	}

	// Already at the beginning
	return h.queries[h.index]
}

// Next returns the next query in history, or the original query if at the end.
func (h *SearchHistory) Next(currentQuery string) string {
	if h.index == -1 {
		// Not currently navigating
		return currentQuery
	}

	h.index++
	if h.index >= len(h.queries) {
		// Past the end - return to temp query and stop navigating
		h.index = -1
		return h.tempQuery
	}

	return h.queries[h.index]
}

// Reset stops navigation and clears temporary state.
// Reset resets the navigation state, clearing the temporary query and index.
// This should be called when starting a new search or when the user types a character
// (indicating they want to stop navigating history and start fresh).
func (h *SearchHistory) Reset() {
	h.index = -1
	h.tempQuery = ""
}

type UIMode int

// UI MODE STATE MACHINE
//
// The UI operates in different modes, each handling keys differently:
//
// ┌──────────────┐
// │  ModeNormal  │◄─────────────────────────────────┐
// └──────┬───────┘                                   │
//
//	│                                           │
//	├── Ctrl+F ──────────────────────────►┌────┴─────────┐
//	│                                      │  ModeSearch  │
//	│            Esc, Q, Enter (on empty)  └──────────────┘
//	│                                           │
//	├── : (vim mode) ─────────────────►┌───────┴──────────┐
//	│                                   │ ModeVimCommand   │
//	│                         Esc/Enter └──────────────────┘
//	│                                           │
//	├── Ctrl+Q, Ctrl+S, etc. ──────►┌──────────┴────────┐
//	│                                │   ModePrompt      │
//	│                       Esc/Enter└───────────────────┘
//	│                                           │
//	├── F1 ──────────────────────────►┌────────┴───────┐
//	│                                  │   ModeHelp     │
//	│                             Esc  └────────────────┘
//	│                                           │
//	├── Alt+A ────────────────────────►┌───────┴────────┐
//	│                                  │  ModeAbout     │
//	│                             Esc  └────────────────┘
//	│                                           │
//	└── (various) ────────────────────►┌───────┴────────┐
//	                                   │  ModeMessage   │
//	                          (timeout)└────────────────┘
//
// KEY HANDLING RULES:
// 1. Each mode MUST handle ALL keys (return true) to prevent leakage to editor
// 2. Mode-specific handlers take precedence over global handlers
// 3. Escape should always provide a way to return to ModeNormal
// 4. State cleanup MUST happen on mode exit
//
// INVARIANTS:
// - Only one mode is active at a time
// - Mode transitions are atomic
// - Each mode is responsible for its own cleanup
// - No state should leak between modes
const (
	ModeNormal UIMode = iota
	ModeMessage
	ModePrompt
	ModeHelp
	ModeAbout
	ModeMenu
	ModeSearch // Unified incremental search mode
	ModeVimCommand
	ModeRecovery // Autosave recovery prompt
)

type UI struct {
	screen  term.Screen
	editor  *core.Editor
	menubar *Menubar

	mode        UIMode
	layout      Layout
	showMenubar bool

	// message mode
	message      string
	messageUntil time.Time
	messageTimer *time.Timer

	// prompt mode
	promptKind  PromptKind
	promptLabel string
	promptText  []rune

	// used by overwrite prompt
	pendingPath string

	// used by quit flow
	quitAfterSave bool
	quitNow       bool

	// replace review mode
	replaceFindTerm string
	replaceWithTerm string

	// Session state persistence: These persist across searches within the editor session
	// They are not saved to disk (session-only memory)
	lastFindTerm    string // Last search term used (persists across mode switches)
	lastReplaceTerm string // Last replace term used (persists across mode switches)

	// find/replace mode state
	replacingAll bool

	// Edit mode state
	insertMode bool // true = insert, false = replace/overwrite

	// Vim command mode
	vimCommand []rune

	// Unified search mode (ModeSearch)
	// Session state: These persist for the duration of the editor session
	searchQuery          []rune         // Current search query being typed
	searchQueryPreFilled bool           // True if query was pre-filled from selection (cleared on first keystroke)
	searchHistory        *SearchHistory // Search history for up/down navigation (persists in session)
	searchDebounceTimer  *time.Timer    // Timer for search debouncing
	searchIsSearching    bool           // True when search is executing (debouncing)
	lastSearchQuery      string         // Last executed search query (persists in session)

	// Features
	showLineNumbers bool
	showStatusBar   bool
	softWrap        bool
	showWhitespace  bool

	// Theme
	theme *theme.Theme

	// Syntax highlighting
	syntaxHighlighting bool
	currentLanguage    string            // Current detected/selected language ("" = auto)
	syntaxCache        *syntax.LineCache // Token cache for current file

	// Configuration
	config *config.Config

	// Bracket matching
	bracketMatcher    *core.BracketMatcher
	bracketMatchState *BracketMatchState

	// Autosave
	autosaveManager   *autosave.Manager
	recoveryCandidate *autosave.RecoveryCandidate // Set during recovery prompt
}

// BracketMatchState holds the current bracket match information for highlighting
type BracketMatchState struct {
	CursorLine  int  // Line of bracket under cursor
	CursorCol   int  // Column of bracket under cursor
	MatchLine   int  // Line of matching bracket
	MatchCol    int  // Column of matching bracket
	HasMatch    bool // True if a match was found
	IsOnBracket bool // True if cursor is on a bracket character
}

func New(screen term.Screen, editor *core.Editor, cfg *config.Config) *UI {
	// Set editor tab width from config
	editor.TabWidth = cfg.Editor.TabWidth

	u := &UI{
		screen:             screen,
		editor:             editor,
		menubar:            NewMenubar(),
		mode:               ModeNormal,
		showMenubar:        false,
		showStatusBar:      cfg.UI.ShowStatusBar,
		showWhitespace:     cfg.Editor.ShowWhitespace,
		insertMode:         true, // Always start in insert mode
		syntaxHighlighting: cfg.Editor.SyntaxHighlighting,
		currentLanguage: func() string {
			if cfg.UI.Language == "" {
				return "auto"
			}
			return cfg.UI.Language
		}(),
		config:         cfg,
		theme:          cfg.GetCurrentTheme(),
		bracketMatcher: core.NewBracketMatcher(),
		searchHistory:  NewSearchHistory(20), // Remember last 20 searches
	}

	// Initialize syntax highlighting
	u.initSyntaxHighlighter()

	// Initialize autosave manager
	u.initAutosave()

	return u
}

func (u *UI) SetOptions(lineNumbers, softWrap bool) {
	u.showLineNumbers = lineNumbers
	u.softWrap = softWrap
}

// SetStatusBarVisibility sets the statusbar visibility
func (u *UI) SetStatusBarVisibility(show bool) {
	u.showStatusBar = show
}

// saveConfig persists the current settings to the config file
func (u *UI) saveConfig() {
	if u.config == nil {
		return
	}

	// Update config with current values
	u.config.Editor.LineNumbers = u.showLineNumbers
	u.config.Editor.SoftWrap = u.softWrap
	u.config.Editor.SyntaxHighlighting = u.syntaxHighlighting
	u.config.Editor.ShowWhitespace = u.showWhitespace
	u.config.UI.ShowStatusBar = u.showStatusBar
	// Only save "auto" state to config, not specific languages
	if u.currentLanguage == "auto" || u.currentLanguage == "" {
		u.config.UI.Language = "auto"
	} else {
		// Don't update config for specific language selections
		// Keep existing auto state
		if u.config.UI.Language == "" {
			u.config.UI.Language = "auto"
		}
	}

	// Save to file (ignore errors - don't interrupt user)
	_ = config.Save(u.config)
}

// initSyntaxHighlighter initializes or reinitializes the syntax highlighter
func (u *UI) initSyntaxHighlighter() {
	if !u.syntaxHighlighting {
		u.syntaxCache = nil
		return
	}

	// Determine language (auto-detect or use manual override)
	lang := u.currentLanguage
	if lang == "" || lang == "auto" || lang == "Auto" {
		// Auto-detect from file
		path := u.editor.File().Path
		var firstLine []rune
		if lines := u.editor.Lines(); len(lines) > 0 {
			firstLine = lines[0]
		}
		lang = syntax.DetectLanguage(path, firstLine)
	}

	if lang == "" {
		u.syntaxCache = nil
		return
	}

	u.syntaxCache = syntax.NewLineCache(lang)
}

// ToggleSyntaxHighlighting toggles syntax highlighting on/off
func (u *UI) ToggleSyntaxHighlighting() {
	u.syntaxHighlighting = !u.syntaxHighlighting
	u.initSyntaxHighlighter()
	u.saveConfig()

	if u.syntaxHighlighting {
		u.enterMessage("Syntax highlighting enabled")
	} else {
		u.enterMessage("Syntax highlighting disabled")
	}
}

// initAutosave initializes the autosave manager
func (u *UI) initAutosave() {
	if u.config == nil {
		return
	}

	idleTimeout := time.Duration(u.config.Autosave.IdleTimeout) * time.Second
	minInterval := time.Duration(u.config.Autosave.MinInterval) * time.Second

	u.autosaveManager = autosave.NewManager(u.config.Autosave.Enabled, idleTimeout, minInterval)

	// Set up state provider callback
	u.autosaveManager.SetStateProvider(func() autosave.AutosaveState {
		file := u.editor.File()
		return autosave.AutosaveState{
			Lines:    u.editor.Lines(),
			Path:     file.Path,
			EOL:      file.EOL,
			Encoding: file.Encoding,
			Modified: u.editor.Modified(),
		}
	})

	// Set up error callback (silent - don't interrupt user)
	u.autosaveManager.SetErrorCallback(func(err error) {
		// Silently ignore autosave errors - don't interrupt user
		// Could optionally show a subtle indicator in status bar
	})
}

// ToggleAutosave toggles autosave on/off
func (u *UI) ToggleAutosave() {
	if u.autosaveManager == nil {
		return
	}

	enabled := !u.autosaveManager.IsEnabled()
	u.autosaveManager.SetEnabled(enabled)

	// Update config
	if u.config != nil {
		u.config.Autosave.Enabled = enabled
		_ = config.Save(u.config)
	}

	if enabled {
		u.enterMessage("Autosave enabled")
	} else {
		u.enterMessage("Autosave disabled")
	}
}

// IsAutosaveEnabled returns whether autosave is enabled
func (u *UI) IsAutosaveEnabled() bool {
	if u.autosaveManager == nil {
		return false
	}
	return u.autosaveManager.IsEnabled()
}

// notifyAutosaveEdit should be called when the buffer is modified
func (u *UI) notifyAutosaveEdit() {
	if u.autosaveManager != nil {
		u.autosaveManager.NotifyEdit()
	}
}

// ClearAutosaveForCurrentFile removes the autosave file for the current file
// Called after a successful save
func (u *UI) ClearAutosaveForCurrentFile() {
	if u.autosaveManager != nil {
		file := u.editor.File()
		if file.Path != "" {
			_ = u.autosaveManager.ClearAutosave(file.Path)
		}
	}
}

// CheckForRecovery checks if there's an autosave to recover for the given path.
// If found, enters recovery mode and returns true.
func (u *UI) CheckForRecovery(targetPath string) bool {
	if targetPath == "" {
		return false
	}

	candidate, err := autosave.FindRecoveryCandidate(targetPath)
	if err != nil || candidate == nil {
		return false
	}

	// Enter recovery mode
	u.recoveryCandidate = candidate
	u.mode = ModeRecovery
	return true
}

// handleRecoveryKey handles keyboard input during recovery prompt
func (u *UI) handleRecoveryKey(e term.KeyEvent) bool {
	if u.recoveryCandidate == nil {
		u.mode = ModeNormal
		return true
	}

	switch e.Key {
	case term.KeyEscape:
		// Escape = same as Open Original
		u.doRecoveryAction(autosave.RecoveryOpenOriginal)
		return true

	case term.KeyRune:
		switch e.Rune {
		case 'r', 'R':
			u.doRecoveryAction(autosave.RecoveryRecover)
			return true
		case 'o', 'O':
			u.doRecoveryAction(autosave.RecoveryOpenOriginal)
			return true
		case 'd', 'D':
			u.doRecoveryAction(autosave.RecoveryDiscard)
			return true
		}
	}

	return true // Consume all keys in recovery mode
}

// doRecoveryAction performs the chosen recovery action
func (u *UI) doRecoveryAction(action autosave.RecoveryAction) {
	if u.recoveryCandidate == nil {
		u.mode = ModeNormal
		return
	}

	result, err := autosave.PerformRecovery(u.recoveryCandidate, action)
	candidate := u.recoveryCandidate
	u.recoveryCandidate = nil
	u.mode = ModeNormal

	if err != nil {
		u.enterMessage("Recovery failed: " + err.Error())
		return
	}

	if result == nil {
		return
	}

	switch action {
	case autosave.RecoveryRecover:
		// Load autosave content into editor
		if result.Lines != nil {
			// Create a fake FileData to load
			u.loadRecoveredContent(result)
			u.enterMessage("Recovered from autosave (unsaved changes restored)")
		}

	case autosave.RecoveryOpenOriginal:
		// Just continue with normal file load - caller will handle it
		u.enterMessage("Opened original file (autosave kept)")

	case autosave.RecoveryDiscard:
		// Autosave already deleted by PerformRecovery
		u.enterMessage("Autosave discarded")
	}

	// Update autosave manager path
	if u.autosaveManager != nil {
		u.autosaveManager.UpdatePath(candidate.Meta.OriginalPath)
	}
}

// loadRecoveredContent loads recovered content into the editor
func (u *UI) loadRecoveredContent(result *autosave.RecoveredFile) {
	// We need to create a FileData-like structure and load it
	// Since editor.LoadFile expects FileData, we'll use a direct approach
	fd := &fileio.FileData{
		Lines:    result.Lines,
		Path:     result.Path,
		BaseName: filepath.Base(result.Path),
		EOL:      result.EOL,
		Encoding: result.Encoding,
	}

	u.editor.LoadFile(fd)

	// Mark as modified since this is recovered unsaved content
	// We do this by making a trivial change and undoing it
	// Actually, we can just insert and delete a space
	// But a cleaner way is to have the editor support marking as modified
	// For now, we'll leave it as-is since the content differs from disk

	// Re-initialize syntax highlighting for new file
	u.initSyntaxHighlighter()
}

// GetRecoveryCandidate returns the current recovery candidate (for rendering)
func (u *UI) GetRecoveryCandidate() *autosave.RecoveryCandidate {
	return u.recoveryCandidate
}

// SwitchLanguage changes the syntax highlighting language and saves to config
func (u *UI) SwitchLanguage(lang string) {
	if lang == "Auto" || lang == "auto" {
		lang = "auto"
	}
	u.currentLanguage = lang
	u.initSyntaxHighlighter()
	u.saveConfig()

	if lang == "auto" || lang == "" {
		u.enterMessage("Language: Auto")
	} else {
		u.enterMessage("Language: " + lang)
	}
}

// SwitchLanguageWithoutSavingConfig changes the language for the session only
// Config only stores "auto" or "off" state, not specific languages
func (u *UI) SwitchLanguageWithoutSavingConfig(lang string) {
	u.currentLanguage = lang
	u.initSyntaxHighlighter()
	u.enterMessage("Language: " + lang)
}

// GetCurrentLanguage returns the current language for display
func (u *UI) GetCurrentLanguage() string {
	if u.currentLanguage != "" && u.currentLanguage != "auto" && u.currentLanguage != "Auto" {
		return u.currentLanguage
	}

	// If syntax cache is active, return its detected language
	if u.syntaxCache != nil {
		lang := u.syntaxCache.Language()
		if lang != "" {
			return lang
		}
	}

	return "Plain"
}

// IsSyntaxHighlightingEnabled returns whether syntax highlighting is enabled
func (u *UI) IsSyntaxHighlightingEnabled() bool {
	return u.syntaxHighlighting
}

// InvalidateSyntaxLine marks a line's syntax cache as stale
func (u *UI) InvalidateSyntaxLine(lineNum int) {
	if u.syntaxCache != nil {
		u.syntaxCache.InvalidateLine(lineNum)
	}
}

// InvalidateSyntaxFromLine invalidates all cached lines from the given line onwards
func (u *UI) InvalidateSyntaxFromLine(lineNum int) {
	if u.syntaxCache != nil {
		u.syntaxCache.InvalidateFromLine(lineNum)
	}
}

func (u *UI) Run() error {
	defer func() {
		// Clean up timer on exit
		if u.messageTimer != nil {
			u.messageTimer.Stop()
		}
		// Stop autosave manager
		if u.autosaveManager != nil {
			u.autosaveManager.Stop()
		}
	}()

	for {
		if u.quitNow {
			return nil
		}

		w, h := u.screen.Size()

		// Check if message has expired before computing layout
		if u.mode == ModeMessage && time.Now().After(u.messageUntil) {
			u.mode = ModeNormal
		}

		u.layout = ComputeLayout(w, h, u.mode, u.showMenubar, u.showStatusBar)

		// Update bracket match state before drawing
		u.updateBracketMatch()

		u.draw()

		ev := u.screen.PollEvent()
		if ev == nil {
			continue
		}

		switch e := ev.(type) {
		case term.RedrawEvent:
			// Just continue to redraw
			continue

		case term.KeyEvent:
			if u.mode == ModeHelp {
				u.mode = ModeNormal
				continue
			}

			if u.mode == ModeAbout {
				u.mode = ModeNormal
				continue
			}

			if u.mode == ModeRecovery {
				if u.handleRecoveryKey(e) {
					continue
				}
			}

			if u.mode == ModePrompt {
				if u.handlePromptKey(e) {
					continue
				}
			}

			if u.mode == ModeVimCommand {
				if u.handleVimCommandKey(e) {
					continue
				}
			}

			if u.mode == ModeSearch {
				if u.handleSearchKey(e) {
					continue
				}
			}

			if u.mode == ModeMenu {
				if u.handleMenuKey(e) {
					continue
				}
				// Don't pass unhandled keys to editor when menu is active
				continue
				continue
			}

			if e.Key == term.KeyEscape {
				if u.mode == ModeMessage {
					u.mode = ModeNormal
					continue
				}
				if u.mode == ModeNormal {
					u.toggleMenuFocus()
					continue
				}
				continue
			}

			// Secret vim command mode
			if u.mode == ModeNormal && e.Key == term.KeyRune && e.Rune == ':' {
				u.mode = ModeVimCommand
				u.vimCommand = nil
				continue
			}

			cmd := u.translateKey(e)
			if cmd != nil {
				// Track if this is a save command
				_, isSave := cmd.(core.CmdSave)
				_, isSaveAs := cmd.(core.CmdSaveAs)

				res := u.editor.Apply(cmd, u.layout.Viewport.H)
				if res.Message != "" {
					u.enterMessage(res.Message)
				}

				// Handle post-command actions
				if (isSave || isSaveAs) && !u.editor.Modified() {
					// Clear autosave on successful save
					u.ClearAutosaveForCurrentFile()
				} else {
					// Notify autosave manager of potential buffer modification
					u.notifyAutosaveEdit()
				}
			}
		}
	}
}

func (u *UI) toggleMenuFocus() {
	if !u.showMenubar {
		u.showMenubar = true
		u.mode = ModeMenu
		u.menubar.Active = true
		return
	}

	// If visible, hide it
	u.showMenubar = false
	u.mode = ModeNormal
	u.menubar.Active = false
}

func (u *UI) handleMenuKey(e term.KeyEvent) bool {
	switch e.Key {
	case term.KeyEscape:
		u.mode = ModeNormal
		u.menubar.Active = false
		u.showMenubar = false
		return true
	case term.KeyLeft:
		u.menubar.PrevMenu()
		return true
	case term.KeyRight:
		u.menubar.NextMenu()
		return true
	case term.KeyUp:
		u.menubar.PrevItem()
		u.adjustMenuScroll()
		return true
	case term.KeyDown:
		u.menubar.NextItem()
		u.adjustMenuScroll()
		return true
	case term.KeyEnter:
		u.executeMenuItem()
		return true
	case term.KeyRune:
		// Secret vim command mode from menu
		if e.Rune == ':' {
			u.mode = ModeVimCommand
			u.menubar.Active = false
			u.showMenubar = false
			u.vimCommand = nil
			return true
		}
		// First check for top-level menu shortcuts (always available)
		for i, menu := range u.menubar.Menus {
			if menu.ShortcutKey != 0 && (e.Rune == menu.ShortcutKey || e.Rune == menu.ShortcutKey-32) {
				u.menubar.SelectedMenuIndex = i
				u.menubar.SelectedItemIndex = 0
				u.adjustMenuScroll()
				return true
			}
		}
		// Then check for menu item shortcut keys in current menu
		menu := u.menubar.Menus[u.menubar.SelectedMenuIndex]
		for i, item := range menu.Items {
			if item.ShortcutKey != 0 && (e.Rune == item.ShortcutKey || e.Rune == item.ShortcutKey-32) { // case insensitive
				u.menubar.SelectedItemIndex = i
				u.executeMenuItem()
				return true
			}
		}
	}
	return false
}

func (u *UI) executeMenuItem() {
	menu := u.menubar.Menus[u.menubar.SelectedMenuIndex]
	item := menu.Items[u.menubar.SelectedItemIndex]

	// Skip separators and readonly items
	if item.IsSeparator || item.IsReadOnly {
		return
	}

	// Exit menu mode
	u.mode = ModeNormal
	u.menubar.Active = false
	u.showMenubar = false

	if item.Action != nil {
		item.Action(u)
	} else if item.Command != nil {
		// Track if this is a save command
		_, isSave := item.Command.(core.CmdSave)
		_, isSaveAs := item.Command.(core.CmdSaveAs)

		res := u.editor.Apply(item.Command, u.layout.Viewport.H)
		if res.Message != "" {
			u.enterMessage(res.Message)
		}

		// Handle post-command actions
		if (isSave || isSaveAs) && !u.editor.Modified() {
			// Clear autosave on successful save
			u.ClearAutosaveForCurrentFile()
		} else {
			// Notify autosave manager of potential buffer modification
			u.notifyAutosaveEdit()
		}
	}
}

func (u *UI) adjustMenuScroll() {
	// Ensure selected item is visible within scroll window
	availableHeight := u.layout.Height - 1 // menubar at top
	if availableHeight < 1 {
		availableHeight = 1
	}

	selectedIdx := u.menubar.SelectedItemIndex
	scrollOffset := u.menubar.ScrollOffset

	// If selected item is above visible area, scroll up
	if selectedIdx < scrollOffset {
		u.menubar.ScrollOffset = selectedIdx
	}

	// If selected item is below visible area, scroll down
	if selectedIdx >= scrollOffset+availableHeight {
		u.menubar.ScrollOffset = selectedIdx - availableHeight + 1
	}
}

func (u *UI) translateKey(e term.KeyEvent) core.Command {
	switch {
	case e.Key == term.KeyF1:
		u.mode = ModeHelp
		return nil

	case e.Key == term.KeyF11:
		u.showStatusBar = !u.showStatusBar
		u.saveConfig()
		return nil

	case e.Key == term.KeyInsert:
		u.insertMode = !u.insertMode
		return nil

	case e.Key == term.KeyRune && e.Rune == 'l' && (e.Modifiers&term.ModCtrl) != 0:
		u.showLineNumbers = !u.showLineNumbers
		u.saveConfig()
		return nil

	case e.Key == term.KeyRune && e.Rune == 'w' && (e.Modifiers&term.ModCtrl) != 0:
		u.softWrap = !u.softWrap
		u.saveConfig()
		return nil

	case e.Key == term.KeyRune && e.Rune == 'h' && (e.Modifiers&term.ModCtrl) != 0:
		u.ToggleSyntaxHighlighting()
		return nil

	case e.Key == term.KeyRune && e.Rune == 'q' && (e.Modifiers&term.ModCtrl) != 0:
		u.startQuitFlow()
		return nil

	case e.Key == term.KeyRune && e.Rune == 's' && e.Modifiers == term.ModCtrl:
		if u.editor.File().Path == "" {
			u.enterSaveAs(false)
			return nil
		}
		return core.CmdSave{}

	case e.Key == term.KeyRune && e.Rune == 's' &&
		(e.Modifiers&(term.ModCtrl|term.ModShift)) == (term.ModCtrl|term.ModShift):
		u.enterSaveAs(false)
		return nil

	case e.Key == term.KeyRune && e.Rune == 'a' && e.Modifiers == term.ModCtrl:
		return core.CmdSelectAll{}

	case e.Key == term.KeyRune && e.Rune == 'c' && e.Modifiers == term.ModCtrl:
		return core.CmdCopy{}

	case e.Key == term.KeyRune && e.Rune == 'x' && e.Modifiers == term.ModCtrl:
		return core.CmdCut{}

	case (e.Key == term.KeyDelete && e.Modifiers == term.ModShift):
		return core.CmdCut{}

	case e.Key == term.KeyRune && e.Rune == 'v' && e.Modifiers == term.ModCtrl:
		return core.CmdPaste{}

	case (e.Key == term.KeyInsert && e.Modifiers == term.ModShift):
		return core.CmdPaste{}

	case e.Key == term.KeyRune && e.Rune == 'z' && (e.Modifiers&term.ModCtrl) != 0:
		return core.CmdUndo{}

	case e.Key == term.KeyRune && e.Rune == 'y' && (e.Modifiers&term.ModCtrl) != 0:
		return core.CmdRedo{}

	case e.Key == term.KeyRune && e.Rune == 'z' && (e.Modifiers&(term.ModCtrl|term.ModShift)) == (term.ModCtrl|term.ModShift):
		return core.CmdRedo{}

	case e.Key == term.KeyRune && e.Rune == 'f' && (e.Modifiers&term.ModCtrl) != 0:
		// Enter unified search mode (ModeSearch)
		u.enterSearch()
		return nil

	case e.Key == term.KeyRune && e.Rune == 'g' && (e.Modifiers&term.ModCtrl) != 0:
		u.enterGoToLine()
		return nil

	case e.Key == term.KeyRune && e.Rune == 'b' && (e.Modifiers&term.ModCtrl) != 0:
		return core.CmdJumpToMatchingBracket{}

	case e.Key == term.KeyF3:
		if e.Modifiers == term.ModShift {
			return core.CmdFindPrev{}
		}
		return core.CmdFindNext{}

	case e.Key == term.KeyRune && e.Modifiers == 0:
		if u.insertMode {
			return core.CmdInsertRune{Rune: e.Rune}
		} else {
			return core.CmdReplaceRune{Rune: e.Rune}
		}

	case e.Key == term.KeyEnter:
		return core.CmdInsertNewline{}

	case e.Key == term.KeyBackspace:
		return core.CmdBackspace{}

	case e.Key == term.KeyTab && e.Modifiers == 0:
		return core.CmdTab{}

	case e.Key == term.KeyRune && e.Rune == 'i' && (e.Modifiers&term.ModCtrl) != 0:
		return core.CmdInsertLiteralTab{}

	case e.Key == term.KeyDelete && e.Modifiers == 0:
		return core.CmdDelete{}

	case e.Key == term.KeyLeft && (e.Modifiers&term.ModCtrl) != 0:
		return core.CmdMoveWordLeft{Select: e.Modifiers&term.ModShift != 0}
	case e.Key == term.KeyRight && (e.Modifiers&term.ModCtrl) != 0:
		return core.CmdMoveWordRight{Select: e.Modifiers&term.ModShift != 0}
	case e.Key == term.KeyLeft:
		return core.CmdMoveLeft{Select: e.Modifiers&term.ModShift != 0}
	case e.Key == term.KeyRight:
		return core.CmdMoveRight{Select: e.Modifiers&term.ModShift != 0}
	case e.Key == term.KeyUp:
		return core.CmdMoveUp{Select: e.Modifiers&term.ModShift != 0}
	case e.Key == term.KeyDown:
		return core.CmdMoveDown{Select: e.Modifiers&term.ModShift != 0}

	case e.Key == term.KeyPageUp:
		return core.CmdPageUp{Select: e.Modifiers&term.ModShift != 0}
	case e.Key == term.KeyPageDown:
		return core.CmdPageDown{Select: e.Modifiers&term.ModShift != 0}

	case e.Key == term.KeyHome && (e.Modifiers&term.ModCtrl) != 0:
		return core.CmdFileStart{Select: e.Modifiers&term.ModShift != 0}
	case e.Key == term.KeyEnd && (e.Modifiers&term.ModCtrl) != 0:
		return core.CmdFileEnd{Select: e.Modifiers&term.ModShift != 0}

	case e.Key == term.KeyHome:
		return core.CmdMoveHome{Select: e.Modifiers&term.ModShift != 0}
	case e.Key == term.KeyEnd:
		return core.CmdMoveEnd{Select: e.Modifiers&term.ModShift != 0}
	}

	return nil
}

// enterSearch enters the unified search mode (ModeSearch).
// If text is currently selected, it will be used as the initial search query.
// Otherwise, the last search query will be used if available.
//
// STATE MACHINE TRANSITION:
// - FROM: ModeNormal (typical), ModePrompt (after canceling a prompt), ModeHelp, ModeAbout, ModeMessage
// - TO: ModeSearch
// - Guards: None (can always enter search from any mode)
// - Side effects: Stops any pending debounce timers, pre-fills query, starts search session
//
// EDGE CASES HANDLED:
// - Already in search mode: Re-enters search (resets state)
// - Empty query: Allowed, shows "..." placeholder
// - Selection pre-fill: Only for single-line selections
func (u *UI) enterSearch() {
	// Guard: If already in search mode, this is a no-op (already handled by calling code)
	// But we proceed anyway to reset the search state if needed

	// Stop any existing debounce timer
	if u.searchDebounceTimer != nil {
		u.searchDebounceTimer.Stop()
		u.searchDebounceTimer = nil
	}

	// Pre-fill from selection if available
	u.searchQueryPreFilled = false
	if u.editor.HasSelection() {
		sl, sc, el, ec := u.editor.GetSelectionRange()
		// Only pre-fill if selection is on a single line
		if sl == el {
			lines := u.editor.Lines()
			if sl < len(lines) && ec <= len(lines[sl]) {
				selectedText := lines[sl][sc:ec]
				u.searchQuery = make([]rune, len(selectedText))
				copy(u.searchQuery, selectedText)
				u.searchQueryPreFilled = true // Mark as pre-filled so first keystroke replaces
			}
		}
	} else if u.lastSearchQuery != "" {
		// Use last search query
		u.searchQuery = []rune(u.lastSearchQuery)
		u.searchQueryPreFilled = false // Last query is not "pre-filled" in the UX sense
	} else {
		// Start with empty query
		u.searchQuery = nil
		u.searchQueryPreFilled = false
	}

	// Reset search history navigation
	u.searchHistory.Reset()

	// Enter search mode
	u.mode = ModeSearch
	u.searchIsSearching = false

	// Perform initial search if we have a query
	if len(u.searchQuery) > 0 {
		u.performSearch()
	}
}

// exitSearch exits the search mode and returns to normal mode.
// Cleans up the search session and resets state.
//
// STATE MACHINE TRANSITION:
//   - FROM: ModeSearch
//   - TO: ModeNormal
//   - Guards: None (always safe to exit)
//   - Side effects: Stops debounce timers, saves query to history, ends editor search session,
//     clears selection, resets all search state
//
// CLEANUP PERFORMED:
// - Stops any pending debounce timer
// - Saves non-empty queries to search history
// - Ends the editor's search session (clears highlights, match data)
// - Clears any active selection
// - Resets all UI search state variables
func (u *UI) exitSearch() {
	// Stop any debounce timer
	if u.searchDebounceTimer != nil {
		u.searchDebounceTimer.Stop()
		u.searchDebounceTimer = nil
	}

	// Save the search query to history
	if len(u.searchQuery) > 0 {
		queryStr := string(u.searchQuery)
		u.searchHistory.Add(queryStr)
		u.lastSearchQuery = queryStr
	}

	// End the search session in the editor
	u.editor.EndSearchSession()

	// Clear selection if any
	u.editor.ClearSelection()

	// Reset state
	u.searchQuery = nil
	u.searchIsSearching = false
	u.searchHistory.Reset()

	// Return to normal mode
	u.mode = ModeNormal
}

// performSearch executes the search with debouncing.
// This is called whenever the search query changes.
//
// DEBOUNCING STRATEGY:
// - Waits 150ms after last keystroke before executing search
// - Shows "Searching..." indicator for queries longer than 20 characters
// - Prevents excessive search operations during typing
//
// THREAD SAFETY:
// - Uses time.AfterFunc which runs in a separate goroutine
// - Screen events are thread-safe via PushEvent
//
// EDGE CASES HANDLED:
// - Multiple rapid calls: Previous timer is stopped, new timer started
// - Long queries: Shows searching indicator immediately
// - Empty queries: Handled by doSearch (clears session)
func (u *UI) performSearch() {
	// Stop any existing timer
	if u.searchDebounceTimer != nil {
		u.searchDebounceTimer.Stop()
	}

	// Show "Searching..." indicator immediately if query is long
	if len(u.searchQuery) > 20 {
		u.searchIsSearching = true
		u.screen.PushEvent(term.RedrawEvent{})
	}

	// Debounce: wait 150ms after last keystroke
	u.searchDebounceTimer = time.AfterFunc(150*time.Millisecond, func() {
		u.doSearch()
	})
}

// doSearch performs the actual search without debouncing.
//
// SEARCH EXECUTION:
// - Clears search session if query is empty
// - Starts/updates search session with current query
// - Moves cursor to first match and selects it
// - Ensures match is visible in viewport
// - Triggers screen redraw
//
// EDGE CASES HANDLED:
// - Empty query: Clears session, no error
// - No matches: Session remains active (error state shown in status bar)
// - First search vs. subsequent: StartSearchSession handles both
// - Match visibility: Uses EnsureVisible to scroll to match
func (u *UI) doSearch() {
	u.searchIsSearching = false

	query := string(u.searchQuery)
	if query == "" {
		// Empty query - clear search session
		u.editor.EndSearchSession()
		u.screen.PushEvent(term.RedrawEvent{})
		return
	}

	// Start or update the search session
	u.editor.StartSearchSession(query)

	// Get the session
	session := u.editor.GetSearchSession()
	if session == nil {
		u.screen.PushEvent(term.RedrawEvent{})
		return
	}

	// If we have matches, move to the first one
	if session.HasMatches() {
		match := session.GetCurrentMatch()
		if match != nil {
			// Move cursor to the match and select it
			u.editor.SetSelection(match.Line, match.Col, match.Length)
			// Make sure the match is visible
			u.editor.EnsureVisible(u.layout.Viewport.W, u.layout.Viewport.H)
		}
	}

	// Trigger redraw
	u.screen.PushEvent(term.RedrawEvent{})
}

// nextSearchMatch moves to the next search match.
// If no matches exist or no session is active, this is a no-op.
// Updates the editor selection to highlight the new match and ensures it's visible.
func (u *UI) nextSearchMatch() {
	session := u.editor.GetSearchSession()
	if session == nil || !session.HasMatches() {
		return
	}

	session.NextMatch()
	match := session.GetCurrentMatch()
	if match != nil {
		u.editor.SetSelection(match.Line, match.Col, match.Length)
		u.editor.EnsureVisible(u.layout.Viewport.W, u.layout.Viewport.H)
	}
	u.screen.PushEvent(term.RedrawEvent{})
}

// prevSearchMatch moves to the previous search match.
// If no matches exist or no session is active, this is a no-op.
// Updates the editor selection to highlight the new match and ensures it's visible.
func (u *UI) prevSearchMatch() {
	session := u.editor.GetSearchSession()
	if session == nil || !session.HasMatches() {
		return
	}

	session.PrevMatch()
	match := session.GetCurrentMatch()
	if match != nil {
		u.editor.SetSelection(match.Line, match.Col, match.Length)
		u.editor.EnsureVisible(u.layout.Viewport.W, u.layout.Viewport.H)
	}
	u.screen.PushEvent(term.RedrawEvent{})
}

// searchHistoryPrev navigates backwards in search history.
// On first call, stores the current query and moves to the most recent historical query.
// On subsequent calls, continues moving backwards through history.
// If already at the beginning, stays at the oldest query.
func (u *UI) searchHistoryPrev() {
	currentQuery := string(u.searchQuery)
	prevQuery := u.searchHistory.Prev(currentQuery)
	u.searchQuery = []rune(prevQuery)
	u.performSearch()
}

// searchHistoryNext navigates forwards in search history.
// Moves to the next query in history. If already at the end, returns to the
// original query that was being typed (stored in tempQuery).
func (u *UI) searchHistoryNext() {
	currentQuery := string(u.searchQuery)
	nextQuery := u.searchHistory.Next(currentQuery)
	u.searchQuery = []rune(nextQuery)
	u.performSearch()
}

// handleSearchKey handles key events in unified search mode (ModeSearch).
// This function must handle ALL keys to prevent key leakage to the editor.
// Returns true for all keys to indicate they were handled.
//
// KEY BINDINGS:
// - Escape: Exit search mode
// - Enter: Move to next match
// - Backspace: Delete character from query (or exit if query is empty)
// - Up/Down: Navigate search history
// - F3 / Shift+F3: Next/previous match
// - F1: Show help
// - Alt+C: Toggle case sensitivity
// - Alt+W: Toggle whole word matching
// - Ctrl+V: Paste into search query
// - Ctrl+R: Replace current match (if matches exist)
// - Ctrl+H: Replace all matches (if matches exist)
// - Any printable character: Add to search query
//
// EDGE CASES HANDLED:
// - Empty query + backspace: Exits search mode (intuitive)
// - Ctrl+R/H with no matches: Does nothing (no error)
// - All navigation/command keys: Consumed to prevent leakage
//
// CRITICAL: This function MUST return true for ALL keys to prevent leakage to editor buffer.
// UX PRINCIPLE: Search is a TEXT INPUT FIELD first. All letters/numbers/symbols are typed.
// Commands require modifier keys (Ctrl/Alt) or function keys (F3, Escape, Enter).
func (u *UI) handleSearchKey(e term.KeyEvent) bool {
	switch e.Key {
	case term.KeyEscape:
		// Exit search mode
		u.exitSearch()
		return true

	case term.KeyEnter:
		// Move to next match (same as 'n' or F3)
		u.nextSearchMatch()
		return true

	case term.KeyBackspace:
		// Remove character from search query
		if len(u.searchQuery) > 0 {
			u.searchQuery = u.searchQuery[:len(u.searchQuery)-1]
			u.searchQueryPreFilled = false // Clear pre-filled flag
			u.performSearch()
		} else {
			// Backspace on empty search = exit (intuitive)
			u.exitSearch()
		}
		return true

	case term.KeyUp:
		// Navigate search history backwards
		u.searchHistoryPrev()
		return true

	case term.KeyDown:
		// Navigate search history forwards
		u.searchHistoryNext()
		return true

	case term.KeyF3:
		// Next/previous match
		if e.Modifiers == term.ModShift {
			u.prevSearchMatch()
		} else {
			u.nextSearchMatch()
		}
		return true

	case term.KeyF1:
		// Show search help overlay
		u.mode = ModeHelp
		return true

	case term.KeyRune:
		if e.Modifiers == term.ModAlt {
			switch e.Rune {
			case 'c', 'C':
				// Toggle case sensitivity (Alt+C matches VS Code)
				u.editor.ToggleCaseSensitivity()
				u.performSearch()
				return true
			case 'w', 'W':
				// Toggle whole word matching
				u.editor.ToggleWholeWord()
				u.performSearch()
				return true
			}
			// Consume all other Alt combinations
			return true
		} else if e.Modifiers == term.ModCtrl {
			switch e.Rune {
			case 'v', 'V':
				// Paste into search (helpful for long patterns)
				clipboard := &SystemClipboard{}
				if text, err := clipboard.Get(); err == nil {
					u.searchQuery = append(u.searchQuery, []rune(text)...)
					u.performSearch()
				}
				return true
			case 'f', 'F':
				// Ctrl+F in search mode - do nothing (already in search)
				return true
			case 'r', 'R':
				// Ctrl+R: Replace current match
				searchState := u.editor.SearchState()
				if searchState != nil && searchState.Session != nil && len(searchState.Session.Matches) > 0 {
					u.lastFindTerm = string(u.searchQuery)
					u.mode = ModePrompt
					u.promptKind = PromptReplaceWith
					u.promptLabel = "Replace with: "
					u.promptText = []rune(u.lastReplaceTerm)
					u.replacingAll = false
				}
				return true
			case 'h', 'H':
				// Ctrl+H: Replace all (matches VS Code and other editors)
				searchState := u.editor.SearchState()
				if searchState != nil && searchState.Session != nil && len(searchState.Session.Matches) > 0 {
					u.lastFindTerm = string(u.searchQuery)
					u.enterReplaceAllConfirm()
				}
				return true
			}
			// Consume all other Ctrl combinations
			return true
		} else {
			// Handle regular runes - ALL characters should be added to search query
			// This is the correct UX: a search field is primarily for TEXT INPUT
			// Commands (next, prev, replace) are accessed via function keys or Ctrl shortcuts

			// If query was pre-filled from selection, replace it on first keystroke
			if u.searchQueryPreFilled {
				u.searchQuery = []rune{e.Rune}
				u.searchQueryPreFilled = false
			} else {
				u.searchQuery = append(u.searchQuery, e.Rune)
			}
			u.performSearch()
			return true
		}

	case term.KeyDelete:
		// Delete key - consume but don't do anything special
		return true

	case term.KeyHome:
		// Home key - consume but don't do anything special
		return true

	case term.KeyEnd:
		// End key - consume but don't do anything special
		return true

	case term.KeyLeft:
		// Left arrow - consume but don't do anything special
		return true

	case term.KeyRight:
		// Right arrow - consume but don't do anything special
		return true

	case term.KeyPageUp:
		// Page up - consume but don't do anything special
		return true

	case term.KeyPageDown:
		// Page down - consume but don't do anything special
		return true

	case term.KeyTab:
		// Tab - consume but don't do anything special
		return true

	default:
		// CRITICAL: Consume ALL other keys to prevent leakage
		// This includes any key type we didn't explicitly handle above
		return true
	}

	// NOTE: This line should never be reached due to default case above,
	// but kept as final safety net
	return true
}

func (u *UI) enterMessage(msg string) {
	u.mode = ModeMessage
	u.message = msg
	u.messageUntil = time.Now().Add(2 * time.Second)

	// Cancel any existing timer
	if u.messageTimer != nil {
		u.messageTimer.Stop()
	}

	// Start a new timer to inject a redraw event when message expires
	u.messageTimer = time.AfterFunc(2*time.Second, func() {
		// Push a redraw event to trigger screen update
		u.screen.PushEvent(term.RedrawEvent{})
	})
}

// SwitchTheme changes the current theme and saves the config
func (u *UI) SwitchTheme(themeName string) {
	// Load the new theme
	newTheme := u.config.GetTheme(themeName)
	if newTheme == nil {
		u.enterMessage("Theme not found: " + themeName)
		return
	}

	// Update the theme
	u.theme = newTheme
	u.config.UI.Theme = themeName

	// Save the config
	if err := config.Save(u.config); err != nil {
		u.enterMessage("Failed to save config: " + err.Error())
		return
	}

	u.enterMessage("Theme changed to: " + themeName)
	u.screen.PushEvent(term.RedrawEvent{})
}

// GetCurrentThemeName returns the name of the current theme
func (u *UI) GetCurrentThemeName() string {
	return u.config.UI.Theme
}

func (u *UI) handleVimCommandKey(e term.KeyEvent) bool {
	switch e.Key {
	case term.KeyEscape:
		u.mode = ModeNormal
		u.vimCommand = nil
		return true

	case term.KeyBackspace:
		if len(u.vimCommand) > 0 {
			u.vimCommand = u.vimCommand[:len(u.vimCommand)-1]
		}
		return true

	case term.KeyEnter:
		u.executeVimCommand()
		return true

	case term.KeyRune:
		u.vimCommand = append(u.vimCommand, e.Rune)
		return true
	}
	return false
}

func (u *UI) executeVimCommand() {
	cmd := string(u.vimCommand)
	u.mode = ModeNormal
	u.vimCommand = nil

	switch cmd {
	case "w", "w!":
		// Save file
		res := u.editor.Apply(core.CmdSave{}, u.layout.Viewport.H)
		if res.Message != "" {
			u.enterMessage(res.Message)
		}
		if !u.editor.Modified() {
			// Clear autosave on successful save
			u.ClearAutosaveForCurrentFile()
		}

	case "q":
		// Quit if no unsaved changes
		if u.editor.Modified() {
			u.enterMessage("No write since last change (use :q! to override)")
		} else {
			// Clear autosave on clean quit
			u.ClearAutosaveForCurrentFile()
			u.quitNow = true
		}

	case "q!":
		// Quit without saving (keep autosave for potential recovery)
		u.quitNow = true

	case "wq":
		// Save and quit
		res := u.editor.Apply(core.CmdSave{}, u.layout.Viewport.H)
		if res.Message != "" {
			u.enterMessage(res.Message)
		}
		if !u.editor.Modified() {
			// Clear autosave on successful save
			u.ClearAutosaveForCurrentFile()
			u.quitNow = true
		}

	default:
		u.enterMessage("Unknown command: :" + cmd)
	}
}

// updateBracketMatch updates the bracket match state based on the current cursor position
func (u *UI) updateBracketMatch() {
	u.bracketMatchState = nil

	line, col := u.editor.Cursor()
	lines := u.editor.Lines()

	if line >= len(lines) || col >= len(lines[line]) {
		return
	}

	ch := lines[line][col]
	if !u.bracketMatcher.IsBracket(ch) {
		return
	}

	// Check if cursor position is in string/comment (skip matching there)
	if u.isInStringOrComment(line, col) {
		return
	}

	matchLine, matchCol, found := u.bracketMatcher.FindMatch(
		lines, line, col,
		u.isInStringOrComment,
	)

	u.bracketMatchState = &BracketMatchState{
		CursorLine:  line,
		CursorCol:   col,
		MatchLine:   matchLine,
		MatchCol:    matchCol,
		HasMatch:    found,
		IsOnBracket: true,
	}
}

// isInStringOrComment returns true if the position is inside a string or comment
func (u *UI) isInStringOrComment(line, col int) bool {
	if u.syntaxCache == nil {
		return false // No syntax info, don't skip anything
	}

	lines := u.editor.Lines()
	if line >= len(lines) {
		return false
	}

	tokens := u.syntaxCache.GetTokens(line, lines[line])
	tokenType := syntax.GetTokenAt(tokens, col)
	return tokenType == syntax.TokenString || tokenType == syntax.TokenComment
}
