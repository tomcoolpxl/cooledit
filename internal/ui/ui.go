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
	"cooledit/internal/term"
	"cooledit/internal/theme"
)

type UIMode int

const (
	ModeNormal UIMode = iota
	ModeMessage
	ModePrompt
	ModeHelp
	ModeAbout
	ModeMenu
	ModeFindReplace
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

	// Features
	showLineNumbers bool
	showStatusBar   bool
	softWrap        bool

	// Theme
	theme *theme.Theme

	// Configuration
	config *config.Config
}

func New(screen term.Screen, editor *core.Editor, cfg *config.Config) *UI {
	// Set editor tab width from config
	editor.TabWidth = cfg.Editor.TabWidth

	return &UI{
		screen:        screen,
		editor:        editor,
		menubar:       NewMenubar(),
		mode:          ModeNormal,
		showMenubar:   false,
		showStatusBar: cfg.UI.ShowStatusBar,
		insertMode:    true, // Always start in insert mode
		config:        cfg,
		theme:         cfg.GetCurrentTheme(),
	}
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
	u.config.UI.ShowStatusBar = u.showStatusBar

	// Save to file (ignore errors - don't interrupt user)
	_ = config.Save(u.config)
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
		u.enterFind()
		return nil

	case e.Key == term.KeyRune && e.Rune == 'g' && (e.Modifiers&term.ModCtrl) != 0:
		u.enterGoToLine()
		return nil

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
		}
	}

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
