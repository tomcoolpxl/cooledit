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
	"time"

	"cooledit/internal/config"
	"cooledit/internal/core"
	"cooledit/internal/syntax"
	"cooledit/internal/term"
	"cooledit/internal/theme"
)

// SearchHistory maintains a history of search queries for navigation.
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
func (h *SearchHistory) Reset() {
	h.index = -1
	h.tempQuery = ""
}

type UIMode int

const (
	ModeNormal UIMode = iota
	ModeMessage
	ModePrompt
	ModeHelp
	ModeAbout
	ModeMenu
	ModeFindReplace // Legacy mode, being replaced by ModeSearch
	ModeSearch      // Unified incremental search mode
	ModeVimCommand
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

	// remember last search/replace terms
	lastFindTerm    string
	lastReplaceTerm string

	// find/replace mode state
	replacingAll bool

	// Edit mode state
	insertMode bool // true = insert, false = replace/overwrite

	// Vim command mode
	vimCommand []rune

	// Unified search mode (ModeSearch)
	searchQuery         []rune         // Current search query being typed
	searchHistory       *SearchHistory // Search history for up/down navigation
	searchDebounceTimer *time.Timer    // Timer for search debouncing
	searchIsSearching   bool           // True when search is executing (debouncing)
	lastSearchQuery     string         // Last executed search query

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
				res := u.editor.Apply(cmd, u.layout.Viewport.H)
				if res.Message != "" {
					u.enterMessage(res.Message)
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
		res := u.editor.Apply(item.Command, u.layout.Viewport.H)
		if res.Message != "" {
			u.enterMessage(res.Message)
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

func (u *UI) handleFindReplaceKey(e term.KeyEvent) bool {
	switch e.Key {
	case term.KeyEscape:
		// Exit find/replace mode
		u.mode = ModeNormal
		u.editor.ClearSelection()
		return true

	case term.KeyF3:
		// Find next
		if e.Modifiers == term.ModShift {
			res := u.editor.Apply(core.CmdFindPrev{}, u.layout.Viewport.H)
			if res.Message != "" && res.Message[:9] == "Not found" {
				u.mode = ModeNormal
				u.enterMessage(res.Message)
			}
		} else {
			res := u.editor.Apply(core.CmdFindNext{}, u.layout.Viewport.H)
			if res.Message != "" && res.Message[:9] == "Not found" {
				u.mode = ModeNormal
				u.enterMessage(res.Message)
			}
		}
		return true

	case term.KeyRune:
		switch e.Rune {
		case 'n', 'N':
			// Find next
			res := u.editor.Apply(core.CmdFindNext{}, u.layout.Viewport.H)
			if res.Message != "" && res.Message[:9] == "Not found" {
				u.mode = ModeNormal
				u.enterMessage(res.Message)
			}
			return true

		case 'p', 'P':
			// Find previous
			res := u.editor.Apply(core.CmdFindPrev{}, u.layout.Viewport.H)
			if res.Message != "" && res.Message[:9] == "Not found" {
				u.mode = ModeNormal
				u.enterMessage(res.Message)
			}
			return true

		case 'r', 'R':
			// Replace current match - need to prompt for replacement text
			u.mode = ModePrompt
			u.promptKind = PromptReplaceWith
			u.promptLabel = "Replace with: "
			if u.lastReplaceTerm != "" {
				u.promptText = []rune(u.lastReplaceTerm)
			} else {
				u.promptText = nil
			}
			return true

		case 'a', 'A':
			// Replace all - need to prompt for replacement text
			u.replacingAll = true
			u.mode = ModePrompt
			u.promptKind = PromptReplaceWith
			u.promptLabel = "Replace all with: "
			if u.lastReplaceTerm != "" {
				u.promptText = []rune(u.lastReplaceTerm)
			} else {
				u.promptText = nil
			}
			return true

		case 'q', 'Q':
			// Quit find mode
			u.mode = ModeNormal
			u.editor.ClearSelection()
			return true

		default:
			// CRITICAL: Explicitly consume all other runes to prevent key leakage
			// In find/replace mode, only the above keys (n/p/r/a/q) are valid
			// Any other character should be ignored, not inserted into editor
			return true
		}

	default:
		// Handle any other key types (arrows, function keys, etc.)
		// Consume them to prevent unexpected behavior
		return true
	}

	// NOTE: This line should never be reached due to default cases above,
	// but kept as final safety net
	return true
}

// enterSearch enters the unified search mode (ModeSearch).
// If text is currently selected, it will be used as the initial search query.
// Otherwise, the last search query will be used if available.
func (u *UI) enterSearch() {
	// Stop any existing debounce timer
	if u.searchDebounceTimer != nil {
		u.searchDebounceTimer.Stop()
		u.searchDebounceTimer = nil
	}

	// Pre-fill from selection if available
	if u.editor.HasSelection() {
		sl, sc, el, ec := u.editor.GetSelectionRange()
		// Only pre-fill if selection is on a single line
		if sl == el {
			lines := u.editor.Lines()
			if sl < len(lines) && ec <= len(lines[sl]) {
				selectedText := lines[sl][sc:ec]
				u.searchQuery = make([]rune, len(selectedText))
				copy(u.searchQuery, selectedText)
			}
		}
	} else if u.lastSearchQuery != "" {
		// Use last search query
		u.searchQuery = []rune(u.lastSearchQuery)
	} else {
		// Start with empty query
		u.searchQuery = nil
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
			u.editor.EnsureVisible(match.Line, u.layout.Viewport.H)
		}
	}

	// Trigger redraw
	u.screen.PushEvent(term.RedrawEvent{})
}

// nextSearchMatch moves to the next search match.
func (u *UI) nextSearchMatch() {
	session := u.editor.GetSearchSession()
	if session == nil || !session.HasMatches() {
		return
	}

	session.NextMatch()
	match := session.GetCurrentMatch()
	if match != nil {
		u.editor.SetSelection(match.Line, match.Col, match.Length)
		u.editor.EnsureVisible(match.Line, u.layout.Viewport.H)
	}
	u.screen.PushEvent(term.RedrawEvent{})
}

// prevSearchMatch moves to the previous search match.
func (u *UI) prevSearchMatch() {
	session := u.editor.GetSearchSession()
	if session == nil || !session.HasMatches() {
		return
	}

	session.PrevMatch()
	match := session.GetCurrentMatch()
	if match != nil {
		u.editor.SetSelection(match.Line, match.Col, match.Length)
		u.editor.EnsureVisible(match.Line, u.layout.Viewport.H)
	}
	u.screen.PushEvent(term.RedrawEvent{})
}

// searchHistoryPrev navigates backwards in search history.
func (u *UI) searchHistoryPrev() {
	currentQuery := string(u.searchQuery)
	prevQuery := u.searchHistory.Prev(currentQuery)
	u.searchQuery = []rune(prevQuery)
	u.performSearch()
}

// searchHistoryNext navigates forwards in search history.
func (u *UI) searchHistoryNext() {
	currentQuery := string(u.searchQuery)
	nextQuery := u.searchHistory.Next(currentQuery)
	u.searchQuery = []rune(nextQuery)
	u.performSearch()
}

// handleSearchKey handles key events in unified search mode (ModeSearch).
// This function must handle ALL keys to prevent key leakage to the editor.
// Returns true for all keys to indicate they were handled.
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
			}
			// Consume all other Ctrl combinations
			return true
		} else {
			// Handle regular runes
			switch e.Rune {
			case 'n', 'N':
				// Next match (if Shift is held, treat as regular character)
				if e.Modifiers == term.ModShift {
					// 'N' - add to search
					u.searchQuery = append(u.searchQuery, e.Rune)
					u.performSearch()
				} else {
					// 'n' - next match
					u.nextSearchMatch()
				}
				return true
			case 'p', 'P':
				// Previous match (if Shift is held, treat as regular character)
				if e.Modifiers == term.ModShift {
					// 'P' - add to search
					u.searchQuery = append(u.searchQuery, e.Rune)
					u.performSearch()
				} else {
					// 'p' - previous match
					u.prevSearchMatch()
				}
				return true
			case 'r', 'R':
				// Replace current match - enter replace prompt
				// TODO: Implement replace prompt integration
				// For now, just treat as regular character
				u.searchQuery = append(u.searchQuery, e.Rune)
				u.performSearch()
				return true
			case 'a', 'A':
				// Replace all - enter replace all prompt
				// TODO: Implement replace all prompt integration
				// For now, just treat as regular character
				u.searchQuery = append(u.searchQuery, e.Rune)
				u.performSearch()
				return true
			case 'q', 'Q':
				// Exit search (if Shift is held, treat as regular character)
				if e.Modifiers == term.ModShift {
					// 'Q' - add to search
					u.searchQuery = append(u.searchQuery, e.Rune)
					u.performSearch()
				} else {
					// 'q' - exit search
					u.exitSearch()
				}
				return true
			default:
				// Add character to search query
				u.searchQuery = append(u.searchQuery, e.Rune)
				u.performSearch()
				return true
			}
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

	case "q":
		// Quit if no unsaved changes
		if u.editor.Modified() {
			u.enterMessage("No write since last change (use :q! to override)")
		} else {
			u.quitNow = true
		}

	case "q!":
		// Quit without saving
		u.quitNow = true

	case "wq":
		// Save and quit
		res := u.editor.Apply(core.CmdSave{}, u.layout.Viewport.H)
		if res.Message != "" {
			u.enterMessage(res.Message)
		}
		if !u.editor.Modified() {
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
