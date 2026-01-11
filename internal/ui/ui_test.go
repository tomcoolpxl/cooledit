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
	"os"
	"path/filepath"
	"testing"
	"time"

	"cooledit/internal/config"
	"cooledit/internal/core"
	"cooledit/internal/term"
)

func newTestUI(w, h int) (*UI, *FakeScreen) {
	screen := NewFakeScreen(w, h)
	editor := core.NewEditor(nil)
	cfg := config.Default()
	ui := New(screen, editor, cfg)
	ui.showMenubar = false // Disable by default for tests to match old layout assumptions
	return ui, screen
}

func updateTestLayout(ui *UI, w, h int) {
	ui.layout = ComputeLayout(w, h, ui.mode, ui.showMenubar, true)
}

func draw(ui *UI) {
	w, h := ui.screen.Size()

	// Check message timeout (same as in Run())
	if ui.mode == ModeMessage && time.Now().After(ui.messageUntil) {
		ui.mode = ModeNormal
	}

	updateTestLayout(ui, w, h)
	ui.draw()
}

// dispatch simulates the UI main loop event handling
func dispatch(ui *UI, ev term.Event) {
	w, h := ui.screen.Size()
	updateTestLayout(ui, w, h)

	switch e := ev.(type) {
	case term.KeyEvent:
		if ui.mode == ModeHelp {
			ui.mode = ModeNormal
			return
		}

		if ui.mode == ModePrompt {
			if ui.handlePromptKey(e) {
				return
			}
		}

		if ui.mode == ModeSearch {
			if ui.handleSearchKey(e) {
				return
			}
		}

		if ui.mode == ModeMenu {
			if ui.handleMenuKey(e) {
				return
			}
		}

		if e.Key == term.KeyF10 {
			ui.toggleMenuFocus()
			return
		}

		if e.Key == term.KeyEscape {
			if ui.mode == ModeMessage {
				ui.mode = ModeNormal
			}
			return
		}

		cmd := ui.translateKey(e)
		if cmd != nil {
			res := ui.editor.Apply(cmd, ui.layout.Viewport.H)
			if res.Message != "" {
				ui.enterMessage(res.Message)
			}
		}
	}
}

func typeString(ui *UI, s string) {
	for _, r := range s {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}
}

func TestTypingShowsCursorAndText(t *testing.T) {
	ui, screen := newTestUI(20, 5)

	typeString(ui, "a")
	draw(ui)

	if !screen.cursorVisible {
		t.Fatalf("cursor should be visible")
	}

	if got := screen.Cell(0, 0); got != 'a' {
		t.Fatalf("expected 'a' at (0,0), got %q", got)
	}

	if screen.cursorX != 1 || screen.cursorY != 0 {
		t.Fatalf("expected cursor at (1,0), got (%d,%d)", screen.cursorX, screen.cursorY)
	}
}

func TestCtrlSSaveAsPrompt(t *testing.T) {
	ui, screen := newTestUI(40, 5)

	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 's', Modifiers: term.ModCtrl})
	draw(ui)

	// Prompt is at h-2 (row 3)
	row := 3
	if got := screen.Cell(0, row); got != 'S' {
		t.Fatalf("expected Save as prompt at row %d, got %q", row, got)
	}
}

func TestMessageExpiresToNormalStatus(t *testing.T) {
	ui, screen := newTestUI(40, 5)

	ui.enterMessage("File saved")
	draw(ui)

	row := 3
	if screen.Cell(0, row) != 'F' {
		t.Fatalf("expected message visible at row %d", row)
	}

	ui.messageUntil = time.Now().Add(-1 * time.Second)
	draw(ui)

	if ui.mode != ModeNormal {
		t.Fatalf("expected mode to revert to normal")
	}
}

func TestMessageBarClearsAfterExpiry(t *testing.T) {
	ui, screen := newTestUI(40, 5)

	// Type some text to have content
	typeString(ui, "test")
	draw(ui)

	// Show a message
	ui.enterMessage("Undo")
	draw(ui)

	// Message should be visible on row 3 (h-2)
	messageRow := 3
	if screen.Cell(0, messageRow) != 'U' {
		t.Fatalf("expected 'U' from message at row %d", messageRow)
	}
	if screen.Cell(1, messageRow) != 'n' {
		t.Fatalf("expected 'n' from message at row %d", messageRow)
	}

	// Expire the message
	ui.messageUntil = time.Now().Add(-1 * time.Second)
	draw(ui)

	// Message bar should be cleared (no lingering characters)
	// The row that was the message bar should now show spaces or content from viewport
	// Since the viewport will expand, that row should be part of viewport or empty
	for x := 0; x < 10; x++ {
		ch := screen.Cell(x, messageRow)
		// Should not contain message text anymore
		if x == 0 && ch == 'U' {
			t.Fatalf("message 'U' still visible after expiry at row %d", messageRow)
		}
		if x == 1 && ch == 'n' {
			t.Fatalf("message 'n' still visible after expiry at row %d", messageRow)
		}
	}

	// Mode should be back to normal
	if ui.mode != ModeNormal {
		t.Fatalf("expected mode to be ModeNormal after message expiry")
	}
}

func TestCtrlQCleanQuitSetsFlag(t *testing.T) {
	ui, _ := newTestUI(40, 5)

	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'q', Modifiers: term.ModCtrl})
	if !ui.quitNow {
		t.Fatalf("Ctrl+Q on clean editor should set quitNow")
	}
}

func TestCtrlFEnterSearch(t *testing.T) {
	ui, _ := newTestUI(40, 5)

	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'f', Modifiers: term.ModCtrl})

	// Updated to expect ModeSearch (unified search mode)
	if ui.mode != ModeSearch {
		t.Fatalf("expected mode to be ModeSearch, got %d", ui.mode)
	}

	draw(ui)
	// The status bar should show search-related info
	// Just verify we're in the right mode for now
}

func TestPromptInteraction(t *testing.T) {
	ui, screen := newTestUI(40, 5)

	// Test search mode typing instead of old prompt-based find
	ui.enterSearch()
	draw(ui) // update layout to Search mode

	typeString(ui, "abc")

	if string(ui.searchQuery) != "abc" {
		t.Fatalf("expected search query 'abc', got %q", string(ui.searchQuery))
	}

	dispatch(ui, term.KeyEvent{Key: term.KeyBackspace})
	if string(ui.searchQuery) != "ab" {
		t.Fatalf("expected search query 'ab' after backspace, got %q", string(ui.searchQuery))
	}

	dispatch(ui, term.KeyEvent{Key: term.KeyEscape})
	if ui.mode != ModeNormal {
		t.Fatalf("expected mode Normal after Escape")
	}

	draw(ui)
	if screen.Cell(0, 3) != ' ' {
		t.Fatalf("expected status row to be clear")
	}
}

func TestQuitFlowWithUnsavedChanges(t *testing.T) {
	ui, _ := newTestUI(40, 5)

	typeString(ui, "x")
	if !ui.editor.Modified() {
		t.Fatalf("editor should be modified")
	}

	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'q', Modifiers: term.ModCtrl})

	if ui.mode != ModePrompt || ui.promptKind != PromptQuitConfirm {
		t.Fatalf("expected quit confirmation prompt")
	}

	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'n'})
	if !ui.quitNow {
		t.Fatalf("expected quitNow to be true after 'n'")
	}
}

func TestStatusBarCursorPosition(t *testing.T) {
	ui, screen := newTestUI(40, 5)

	typeString(ui, "a")
	dispatch(ui, term.KeyEvent{Key: term.KeyEnter})
	typeString(ui, "b")
	// Cursor at (1, 1)

	draw(ui)

	found := false
	for x := 0; x < 30; x++ {
		if screen.Cell(x, 4) == 'L' && screen.Cell(x+1, 4) == 'n' && screen.Cell(x+3, 4) == '2' {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("could not find 'Ln 2' in status bar")
	}
}

func TestUndoRedoUI(t *testing.T) {
	ui, _ := newTestUI(40, 5)

	typeString(ui, "a")

	// Ctrl+Z
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'z', Modifiers: term.ModCtrl})

	if len(ui.editor.Lines()[0]) != 0 {
		t.Fatalf("undo should have cleared the line")
	}

	// Ctrl+Y
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'y', Modifiers: term.ModCtrl})

	if string(ui.editor.Lines()[0]) != "a" {
		t.Fatalf("redo should have restored 'a'")
	}
}

func TestViewportScrolling(t *testing.T) {
	ui, screen := newTestUI(20, 5)

	// Insert 6 lines: 1, 2, 3, 4, 5, 6
	for i := 1; i <= 6; i++ {
		typeString(ui, "x")
		dispatch(ui, term.KeyEvent{Key: term.KeyEnter})
	}

	draw(ui)

	// Let's make lines distinct.
	ui, screen = newTestUI(20, 5)
	typeString(ui, "1")
	dispatch(ui, term.KeyEvent{Key: term.KeyEnter})
	typeString(ui, "2")
	dispatch(ui, term.KeyEvent{Key: term.KeyEnter})
	typeString(ui, "3")
	dispatch(ui, term.KeyEvent{Key: term.KeyEnter})
	typeString(ui, "4")
	dispatch(ui, term.KeyEvent{Key: term.KeyEnter})
	typeString(ui, "5")

	draw(ui)

	// Row 0 should show Line 1 ("2")
	if screen.Cell(0, 0) != '2' {
		t.Fatalf("expected scroll to show '2' at top, got %q", screen.Cell(0, 0))
	}

	// Move up to top
	dispatch(ui, term.KeyEvent{Key: term.KeyHome, Modifiers: term.ModCtrl})
	draw(ui)

	// Row 0 should show Line 0 ("1")
	if screen.Cell(0, 0) != '1' {
		t.Fatalf("expected scroll to top to show '1', got %q", screen.Cell(0, 0))
	}
}

func TestSearchUIIntegration(t *testing.T) {
	ui, _ := newTestUI(20, 5)

	typeString(ui, "foo bar foo")
	dispatch(ui, term.KeyEvent{Key: term.KeyHome, Modifiers: term.ModCtrl}) // Go to start

	// Ctrl+F - now enters ModeSearch (unified search mode)
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'f', Modifiers: term.ModCtrl})
	draw(ui)

	// Type search query - search happens in real-time now
	typeString(ui, "foo")
	// Force immediate search execution for testing (bypasses debounce)
	ui.doSearch()
	draw(ui)

	// Should be in unified search mode
	if ui.mode != ModeSearch {
		t.Fatalf("expected ModeSearch, got mode %d", ui.mode)
	}

	// Cursor should be at start of first match (0,0)
	line, col := ui.editor.Cursor()
	if line != 0 || col != 0 {
		t.Fatalf("expected cursor at (0,0), got (%d,%d)", line, col)
	}

	// F3 (Next) - test navigation in unified search mode
	dispatch(ui, term.KeyEvent{Key: term.KeyF3})
	draw(ui)
	// Should find second "foo" at 0, 8
	line, col = ui.editor.Cursor()
	if line != 0 || col != 8 {
		t.Fatalf("expected cursor at (0,8) after next, got (%d,%d)", line, col)
	}

	// Shift+F3 (Prev) in search mode
	dispatch(ui, term.KeyEvent{Key: term.KeyF3, Modifiers: term.ModShift})
	draw(ui)

	// Should find first "foo" at 0, 0
	line, col = ui.editor.Cursor()
	if line != 0 || col != 0 {
		t.Fatalf("expected cursor at (0,0) after prev, got (%d,%d)", line, col)
	}

	// Escape to quit search mode
	dispatch(ui, term.KeyEvent{Key: term.KeyEscape})

	if ui.mode != ModeNormal {
		t.Fatalf("expected ModeNormal after quit, got mode %d", ui.mode)
	}
}

func TestNavigationKeys(t *testing.T) {
	ui, screen := newTestUI(20, 5)
	typeString(ui, "line1")
	dispatch(ui, term.KeyEvent{Key: term.KeyEnter})
	typeString(ui, "line2")

	dispatch(ui, term.KeyEvent{Key: term.KeyHome})
	draw(ui)
	if screen.cursorX != 0 {
		t.Fatalf("Home failed")
	}

	dispatch(ui, term.KeyEvent{Key: term.KeyEnd})
	draw(ui)
	if screen.cursorX != 5 { // "line2" len 5
		t.Fatalf("End failed")
	}

	dispatch(ui, term.KeyEvent{Key: term.KeyUp})
	draw(ui)
	// line1|
	if screen.cursorY != 0 || screen.cursorX != 5 {
		t.Fatalf("Up failed: (%d,%d)", screen.cursorX, screen.cursorY)
	}

	dispatch(ui, term.KeyEvent{Key: term.KeyLeft})
	draw(ui)
	if screen.cursorX != 4 {
		t.Fatalf("Left failed")
	}

	dispatch(ui, term.KeyEvent{Key: term.KeyRight})
	draw(ui)
	if screen.cursorX != 5 {
		t.Fatalf("Right failed")
	}
}

func TestMenubarRendering(t *testing.T) {
	ui, screen := newTestUI(40, 5)
	ui.showMenubar = true // Enable for this test

	draw(ui)

	// Row 0 should be menubar: " File  Edit ..."
	if screen.Cell(1, 0) != 'F' || screen.Cell(2, 0) != 'i' || screen.Cell(3, 0) != 'l' || screen.Cell(4, 0) != 'e' {
		t.Fatalf("menubar not rendered correctly")
	}

	// Viewport starts at row 1
	ui.editor.Apply(core.CmdInsertRune{Rune: 'x'}, 4)
	draw(ui)

	if screen.Cell(0, 1) != 'x' {
		t.Fatalf("expected viewport content at row 1")
	}
}

func TestEnterKeyCursorPosition(t *testing.T) {
	ui, _ := newTestUI(20, 5)
	ui.showMenubar = false

	// Fill buffer:
	// aaaa
	// bbbb
	typeString(ui, "aaaa")
	dispatch(ui, term.KeyEvent{Key: term.KeyEnter})
	typeString(ui, "bbbb")

	// Test cursor position after Enter
	row, col := ui.editor.Cursor()
	if row != 1 || col != 4 {
		t.Fatalf("expected cursor at (1,4), got (%d,%d)", row, col)
	}
}

func TestHelpMode(t *testing.T) {
	ui, screen := newTestUI(20, 5)

	dispatch(ui, term.KeyEvent{Key: term.KeyF1})
	draw(ui)

	if ui.mode != ModeHelp {
		t.Fatal("expected Help mode")
	}

	// Screen should show help text - first line now has title
	firstLineChar := screen.Cell(2, 0)
	if firstLineChar != 'c' { // "cooledit"
		t.Fatalf("expected help title on first line, got %q", firstLineChar)
	}

	// Any key exits
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'x'})
	draw(ui)

	if ui.mode != ModeNormal {
		t.Fatal("expected Normal mode after keypress")
	}
}

func TestMenuNavigationWrapping(t *testing.T) {
	ui, _ := newTestUI(40, 5)

	// Activate menu
	dispatch(ui, term.KeyEvent{Key: term.KeyF10})

	if ui.menubar.SelectedMenuIndex != 0 {
		t.Fatalf("expected menu 0")
	}

	// Left from 0 -> Last
	dispatch(ui, term.KeyEvent{Key: term.KeyLeft})
	if ui.menubar.SelectedMenuIndex != len(ui.menubar.Menus)-1 {
		t.Fatalf("expected last menu, got %d", ui.menubar.SelectedMenuIndex)
	}

	// Right from Last -> 0
	dispatch(ui, term.KeyEvent{Key: term.KeyRight})
	if ui.menubar.SelectedMenuIndex != 0 {
		t.Fatalf("expected menu 0 after wrap right")
	}
}

func TestLayoutBounds(t *testing.T) {
	ui, screen := newTestUI(10, 3) // Too small (w<16 or h<4)

	draw(ui)

	// Should show warning "Screen too small" at 0,0
	if screen.Cell(0, 0) != 'S' {
		t.Fatalf("expected 'S'creen too small warning")
	}
}

func TestExecuteMenuItems(t *testing.T) {
	ui, _ := newTestUI(40, 10)

	// Activate Menu
	dispatch(ui, term.KeyEvent{Key: term.KeyF10})

	// File Menu (Index 0). "Quit" is Item 2.
	ui.menubar.SelectedItemIndex = 2

	// Execute (Enter)
	dispatch(ui, term.KeyEvent{Key: term.KeyEnter})

	// Action for Quit is u.startQuitFlow()
	// Should enter Prompt mode (PromptQuitConfirm)
	// IF clean -> quitNow. IF modified -> prompt.
	// New editor is clean. So quitNow should be true.

	if !ui.quitNow {
		t.Fatalf("expected Quit action to set quitNow")
	}
}

func TestStatusBarMiniHelp(t *testing.T) {
	// Test that mini-help appears in center of status bar on wide terminal
	ui, screen := newTestUI(100, 5)

	draw(ui)

	// Look for mini-help text (F1, Ctrl+Q, etc.) in the middle section
	statusBarY := 4
	foundF1 := false
	foundCtrlQ := false

	for x := 10; x < 80; x++ {
		if screen.Cell(x, statusBarY) == 'F' && screen.Cell(x+1, statusBarY) == '1' {
			foundF1 = true
		}
		if screen.Cell(x, statusBarY) == 'C' && screen.Cell(x+1, statusBarY) == 't' &&
			screen.Cell(x+2, statusBarY) == 'r' && screen.Cell(x+3, statusBarY) == 'l' {
			// Check if it's Ctrl+Q or Ctrl+S
			if x+6 < 80 && screen.Cell(x+5, statusBarY) == 'Q' {
				foundCtrlQ = true
			}
		}
	}

	if !foundF1 {
		t.Fatalf("expected 'F1' in status bar mini-help")
	}
	if !foundCtrlQ {
		t.Fatalf("expected 'Ctrl+Q' in status bar mini-help")
	}
}

func TestStatusBarMiniHelpNarrowTerminal(t *testing.T) {
	// Test that mini-help is truncated on narrow terminal
	ui, screen := newTestUI(50, 5)

	draw(ui)

	// On narrow terminal, should have F1 but maybe not all items
	statusBarY := 4
	foundF1 := false

	for x := 0; x < 50; x++ {
		if screen.Cell(x, statusBarY) == 'F' && screen.Cell(x+1, statusBarY) == '1' {
			foundF1 = true
			break
		}
	}

	if !foundF1 {
		t.Fatalf("expected at least 'F1' in narrow status bar")
	}
}

func TestStatusBarFindReplaceMode(t *testing.T) {
	// Test that status bar shows search options in unified search mode
	ui, _ := newTestUI(80, 5)

	// Type some text and search for it
	typeString(ui, "hello world")
	dispatch(ui, term.KeyEvent{Key: term.KeyHome, Modifiers: term.ModCtrl})

	// Enter unified search mode
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'f', Modifiers: term.ModCtrl})
	draw(ui)

	typeString(ui, "hello")
	// No need to press Enter - search is incremental
	draw(ui)

	// Should be in ModeSearch (unified search mode)
	if ui.mode != ModeSearch {
		t.Fatalf("expected ModeSearch, got mode %d", ui.mode)
	}

	// Status bar should show search info (Find:, match count, etc.)
	// For now, just verify we're in the right mode
	// Status bar rendering tests will be added in Phase 3
}

func TestHelpScreenWideTerminal(t *testing.T) {
	// Test two-column layout on wide terminal
	ui, screen := newTestUI(90, 25)

	dispatch(ui, term.KeyEvent{Key: term.KeyF1})
	draw(ui)

	if ui.mode != ModeHelp {
		t.Fatal("expected Help mode")
	}

	// Should have "cooledit" title at top
	if screen.Cell(2, 0) != 'c' {
		t.Fatalf("expected 'c' from cooledit title")
	}

	// Should have content in both left and right columns
	// Left should have "MENU & HELP" around line 2
	foundMenuHelp := false
	for x := 0; x < 40; x++ {
		if screen.Cell(x, 2) == 'M' && screen.Cell(x+1, 2) == 'E' &&
			screen.Cell(x+2, 2) == 'N' && screen.Cell(x+3, 2) == 'U' {
			foundMenuHelp = true
			break
		}
	}

	// Right column should have "SEARCH" section
	foundSearch := false
	for x := 40; x < 90; x++ {
		if screen.Cell(x, 2) == 'S' && screen.Cell(x+1, 2) == 'E' &&
			screen.Cell(x+2, 2) == 'A' && screen.Cell(x+3, 2) == 'R' &&
			screen.Cell(x+4, 2) == 'C' && screen.Cell(x+5, 2) == 'H' {
			foundSearch = true
			break
		}
	}

	if !foundMenuHelp {
		t.Fatalf("expected 'MENU' section in left column")
	}
	if !foundSearch {
		t.Fatalf("expected 'SEARCH' section in right column")
	}
}

func TestHelpScreenNarrowTerminal(t *testing.T) {
	// Test single-column layout on narrow terminal
	ui, screen := newTestUI(60, 25)

	dispatch(ui, term.KeyEvent{Key: term.KeyF1})
	draw(ui)

	if ui.mode != ModeHelp {
		t.Fatal("expected Help mode")
	}

	// Should have title
	if screen.Cell(2, 0) != 'c' {
		t.Fatalf("expected 'c' from cooledit title")
	}

	// Should have content in single column
	foundMenuHelp := false
	for y := 0; y < 25; y++ {
		for x := 0; x < 40; x++ {
			if screen.Cell(x, y) == 'M' && screen.Cell(x+1, y) == 'E' &&
				screen.Cell(x+2, y) == 'N' && screen.Cell(x+3, y) == 'U' {
				foundMenuHelp = true
				break
			}
		}
		if foundMenuHelp {
			break
		}
	}

	if !foundMenuHelp {
		t.Fatalf("expected 'MENU' section in single column layout")
	}
}

func TestPromptOverwrite(t *testing.T) {
	ui, screen := newTestUI(40, 5)

	// Create dummy file
	dir := t.TempDir()
	path := filepath.Join(dir, "existing.txt")
	os.WriteFile(path, []byte("old"), 0644)

	// Trigger Save As
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 's', Modifiers: term.ModCtrl})

	// Type path
	typeString(ui, path)
	dispatch(ui, term.KeyEvent{Key: term.KeyEnter})

	// Should be in PromptOverwrite
	if ui.promptKind != PromptOverwrite {
		t.Fatalf("expected prompt overwrite")
	}
	draw(ui)
	// Check label "Overwrite existing file? (y/n) "
	if screen.Cell(0, 3) != 'O' {
		t.Fatalf("expected Overwrite prompt")
	}

	// Press 'n' -> Back to Save As
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'n'})
	if ui.promptKind != PromptSaveAs {
		t.Fatalf("expected back to prompt save as")
	}

	// Press Enter again (to overwrite prompt)
	dispatch(ui, term.KeyEvent{Key: term.KeyEnter})

	// Press 'y' -> Save
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'y'})
	if ui.mode != ModeNormal && ui.mode != ModeMessage {
		t.Fatalf("expected normal or message mode after save, got %v", ui.mode)
	}

	// Verify file overwritten
	// Editor buffer is empty (newTestEditor). So file should be empty.
	content, _ := os.ReadFile(path)
	if len(content) != 0 { // Should be empty (plus newline?)
		// Save adds newline if 1 empty line
		// Editor has 1 line [""] -> Save writes "\n" (if EOL is \n) or nothing?
		// Save logic loops lines. 1 line. content "" -> writes "".
		// if i < len-1 (0 < 0) -> no newline.
		// Result empty file.
		if len(content) > 1 { // allow 1 byte newline if any
			t.Fatalf("expected empty file, got len %d: %q", len(content), string(content))
		}
	}
}

func TestDeleteKey(t *testing.T) {
	ui, screen := newTestUI(20, 5)

	// Type "hello"
	typeString(ui, "hello")
	draw(ui)

	// Move home
	dispatch(ui, term.KeyEvent{Key: term.KeyHome})

	// Delete first character ('h')
	dispatch(ui, term.KeyEvent{Key: term.KeyDelete})
	draw(ui)

	// Should show "ello"
	if screen.Cell(0, 0) != 'e' {
		t.Fatalf("expected 'e' at (0,0) after delete, got %q", screen.Cell(0, 0))
	}
	if screen.Cell(1, 0) != 'l' {
		t.Fatalf("expected 'l' at (1,0) after delete, got %q", screen.Cell(1, 0))
	}

	// Cursor should stay at column 0
	if screen.cursorX != 0 || screen.cursorY != 0 {
		t.Fatalf("expected cursor at (0,0), got (%d,%d)", screen.cursorX, screen.cursorY)
	}
}

func TestDeleteKeyMergesLines(t *testing.T) {
	ui, screen := newTestUI(20, 5)

	// Type "abc" + newline + "def"
	typeString(ui, "abc")
	dispatch(ui, term.KeyEvent{Key: term.KeyEnter})
	typeString(ui, "def")
	draw(ui)

	// Move up to end of first line
	dispatch(ui, term.KeyEvent{Key: term.KeyUp})

	// Delete at end of line (should merge with next)
	dispatch(ui, term.KeyEvent{Key: term.KeyDelete})
	draw(ui)

	// Should show "abcdef" on single line
	if screen.Cell(0, 0) != 'a' || screen.Cell(1, 0) != 'b' || screen.Cell(2, 0) != 'c' {
		t.Fatalf("expected 'abc' at start of line")
	}
	if screen.Cell(3, 0) != 'd' || screen.Cell(4, 0) != 'e' || screen.Cell(5, 0) != 'f' {
		t.Fatalf("expected 'def' after 'abc'")
	}
}

func TestDeleteKeyOnEmptyLine(t *testing.T) {
	ui, _ := newTestUI(20, 5)

	// Create empty line + "test"
	dispatch(ui, term.KeyEvent{Key: term.KeyEnter})
	typeString(ui, "test")

	// Move up to empty line
	dispatch(ui, term.KeyEvent{Key: term.KeyUp})

	// Delete on empty line should merge with next
	dispatch(ui, term.KeyEvent{Key: term.KeyDelete})

	// Should have single line with "test"
	lines := ui.editor.Lines()
	if len(lines) != 1 {
		t.Fatalf("expected 1 line after delete on empty line, got %d", len(lines))
	}
	if string(lines[0]) != "test" {
		t.Fatalf("expected 'test', got %q", string(lines[0]))
	}
}

func TestDeleteWithSelection(t *testing.T) {
	ui, screen := newTestUI(20, 5)

	// Type "hello"
	typeString(ui, "hello")

	// Move home and select first 3 chars
	dispatch(ui, term.KeyEvent{Key: term.KeyHome})
	dispatch(ui, term.KeyEvent{Key: term.KeyRight, Modifiers: term.ModShift})
	dispatch(ui, term.KeyEvent{Key: term.KeyRight, Modifiers: term.ModShift})
	dispatch(ui, term.KeyEvent{Key: term.KeyRight, Modifiers: term.ModShift})

	// Delete selection
	dispatch(ui, term.KeyEvent{Key: term.KeyDelete})
	draw(ui)

	// Should show "lo"
	if screen.Cell(0, 0) != 'l' {
		t.Fatalf("expected 'l' at (0,0) after delete selection, got %q", screen.Cell(0, 0))
	}
	if screen.Cell(1, 0) != 'o' {
		t.Fatalf("expected 'o' at (1,0) after delete selection, got %q", screen.Cell(1, 0))
	}
}

func TestDeleteKeyUndo(t *testing.T) {
	ui, screen := newTestUI(20, 5)

	// Type "test"
	typeString(ui, "test")
	dispatch(ui, term.KeyEvent{Key: term.KeyHome})

	// Delete 't'
	dispatch(ui, term.KeyEvent{Key: term.KeyDelete})
	draw(ui)

	// Should show "est"
	if screen.Cell(0, 0) != 'e' {
		t.Fatalf("expected 'e' after delete")
	}

	// Undo
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'z', Modifiers: term.ModCtrl})
	draw(ui)

	// Should show "test" again
	if screen.Cell(0, 0) != 't' {
		t.Fatalf("expected 't' after undo, got %q", screen.Cell(0, 0))
	}
	if screen.Cell(1, 0) != 'e' {
		t.Fatalf("expected 'e' at position 1 after undo")
	}
}

func TestToggleLineNumbersSavesConfig(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.toml")

	// Override ConfigPath
	origConfigPath := config.ConfigPath
	defer func() { config.ConfigPath = origConfigPath }()
	config.ConfigPath = func() (string, error) {
		return configFile, nil
	}

	ui, _ := newTestUI(80, 24)
	ui.showLineNumbers = true

	// Toggle line numbers off
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'l', Modifiers: term.ModCtrl})

	if ui.showLineNumbers {
		t.Error("Line numbers should be toggled off")
	}

	// Load config and verify it was saved
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load saved config: %v", err)
	}

	if cfg.Editor.LineNumbers {
		t.Error("Config should have LineNumbers=false after toggle")
	}
}

func TestToggleSoftWrapSavesConfig(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.toml")

	// Override ConfigPath
	origConfigPath := config.ConfigPath
	defer func() { config.ConfigPath = origConfigPath }()
	config.ConfigPath = func() (string, error) {
		return configFile, nil
	}

	ui, _ := newTestUI(80, 24)
	ui.softWrap = true

	// Toggle soft wrap off
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'w', Modifiers: term.ModCtrl})

	if ui.softWrap {
		t.Error("Soft wrap should be toggled off")
	}

	// Load config and verify it was saved
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load saved config: %v", err)
	}

	if cfg.Editor.SoftWrap {
		t.Error("Config should have SoftWrap=false after toggle")
	}
}

func TestSoftWrapRendering(t *testing.T) {
	ui, screen := newTestUI(20, 10) // Narrow terminal to force wrapping
	ui.softWrap = true

	// Type a long line that will wrap
	typeString(ui, "This is a very long line that should wrap across multiple screen lines")
	draw(ui)

	// With wrap enabled, text should appear on multiple lines
	// First line should have "This is a very long"
	firstLine := ""
	for x := 0; x < 20; x++ {
		if screen.Cell(x, 0) == 0 {
			break
		}
		firstLine += string(screen.Cell(x, 0))
	}

	if len(firstLine) == 0 {
		t.Error("First line should have content with soft wrap enabled")
	}

	// Second line should have continuation
	secondLine := ""
	for x := 0; x < 20; x++ {
		r := screen.Cell(x, 1)
		if r == 0 || r == ' ' {
			break
		}
		secondLine += string(r)
	}

	if len(secondLine) == 0 {
		t.Error("Second line should have wrapped content")
	}
}

func TestSoftWrapVsNoWrap(t *testing.T) {
	// Test that wrap off uses horizontal scrolling
	ui, screen := newTestUI(20, 10)
	ui.softWrap = false

	typeString(ui, "This is a very long line that should scroll horizontally")
	draw(ui)

	// With wrap disabled, cursor should be at end of visible area
	// and we should see only first 20 chars
	firstLine := ""
	for x := 0; x < 20; x++ {
		r := screen.Cell(x, 0)
		if r == 0 {
			break
		}
		firstLine += string(r)
	}

	// Second line should be empty (no wrap)
	secondLineEmpty := true
	for x := 0; x < 20; x++ {
		r := screen.Cell(x, 1)
		if r != 0 && r != ' ' {
			secondLineEmpty = false
			break
		}
	}

	if !secondLineEmpty {
		t.Error("Second line should be empty with soft wrap disabled")
	}
}

func TestInsertMode(t *testing.T) {
	ui, _ := newTestUI(80, 24)

	// Should start in insert mode
	if !ui.insertMode {
		t.Error("Should start in insert mode")
	}

	// Type "hello"
	typeString(ui, "hello")

	// Result should be "hello"
	lines := ui.editor.Lines()
	if len(lines) == 0 || string(lines[0]) != "hello" {
		t.Errorf("Expected 'hello', got %q", string(lines[0]))
	}
}

func TestReplaceMode(t *testing.T) {
	ui, _ := newTestUI(80, 24)

	// Type "hello"
	typeString(ui, "hello")

	// Move to start
	dispatch(ui, term.KeyEvent{Key: term.KeyHome})

	// Toggle to replace mode
	dispatch(ui, term.KeyEvent{Key: term.KeyInsert})

	if ui.insertMode {
		t.Error("Should be in replace mode after Insert key")
	}

	// Type "HELLO" - should overwrite
	typeString(ui, "HELLO")

	// Result should be "HELLO" (replaced all 5 chars)
	lines := ui.editor.Lines()
	if len(lines) == 0 || string(lines[0]) != "HELLO" {
		t.Errorf("Expected 'HELLO', got %q", string(lines[0]))
	}
}

func TestReplaceModeAtEndOfLine(t *testing.T) {
	ui, _ := newTestUI(80, 24)

	// Type "hi"
	typeString(ui, "hi")

	// Toggle to replace mode
	dispatch(ui, term.KeyEvent{Key: term.KeyInsert})

	// Type " there" at end - should insert since nothing to replace
	typeString(ui, " there")

	lines := ui.editor.Lines()
	if len(lines) == 0 || string(lines[0]) != "hi there" {
		t.Errorf("Expected 'hi there', got %q", string(lines[0]))
	}
}

func TestInsertKeyToggle(t *testing.T) {
	ui, _ := newTestUI(80, 24)

	if !ui.insertMode {
		t.Error("Should start in insert mode")
	}

	// Toggle to replace
	dispatch(ui, term.KeyEvent{Key: term.KeyInsert})
	if ui.insertMode {
		t.Error("Should be in replace mode")
	}

	// Toggle back to insert
	dispatch(ui, term.KeyEvent{Key: term.KeyInsert})
	if !ui.insertMode {
		t.Error("Should be back in insert mode")
	}
}

func TestStatusBarReplaceModeIndicator(t *testing.T) {
	ui, screen := newTestUI(80, 24)

	// In insert mode, status bar should not show REPLACE
	draw(ui)

	statusLine := ""
	for x := 0; x < 80; x++ {
		r := screen.Cell(x, 23)
		if r != 0 {
			statusLine += string(r)
		}
	}

	if contains(statusLine, "REPLACE") {
		t.Error("Status bar should not show REPLACE in insert mode")
	}

	// Toggle to replace mode
	dispatch(ui, term.KeyEvent{Key: term.KeyInsert})
	draw(ui)

	statusLine = ""
	for x := 0; x < 80; x++ {
		r := screen.Cell(x, 23)
		if r != 0 {
			statusLine += string(r)
		}
	}

	if !contains(statusLine, "REPLACE") {
		t.Error("Status bar should show REPLACE in replace mode")
	}

	// Toggle back to insert mode
	dispatch(ui, term.KeyEvent{Key: term.KeyInsert})
	draw(ui)

	statusLine = ""
	for x := 0; x < 80; x++ {
		r := screen.Cell(x, 23)
		if r != 0 {
			statusLine += string(r)
		}
	}

	if contains(statusLine, "REPLACE") {
		t.Error("Status bar should not show REPLACE after toggling back to insert mode")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && func() bool {
		for i := 0; i <= len(s)-len(substr); i++ {
			if s[i:i+len(substr)] == substr {
				return true
			}
		}
		return false
	}()
}

func TestLiteralTabRendering(t *testing.T) {
	ui, screen := newTestUI(40, 10)

	// Insert "hello", then a literal tab, then "world"
	typeString(ui, "hello")
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'i', Modifiers: term.ModCtrl}) // Ctrl+I for literal tab
	typeString(ui, "world")

	draw(ui)

	// With default TabWidth=4, "hello\tworld" should render as:
	// "hello   world" (3 spaces to reach next tab stop at column 8)
	// Col 0-4: hello (5 chars)
	// Col 5-7: 3 spaces (to reach column 8, which is next multiple of 4 after 5)
	// Col 8+: world

	// Check that we see the expected spacing
	if screen.Cell(0, 0) != 'h' {
		t.Errorf("Expected 'h' at column 0, got %c", screen.Cell(0, 0))
	}

	// Column 5-7 should be spaces (tab expansion)
	for col := 5; col < 8; col++ {
		if screen.Cell(col, 0) != ' ' {
			t.Errorf("Expected space at column %d (tab expansion), got %c", col, screen.Cell(col, 0))
		}
	}

	// Column 8 should be 'w' (start of "world")
	if screen.Cell(8, 0) != 'w' {
		t.Errorf("Expected 'w' at column 8, got %c", screen.Cell(8, 0))
	}

	// Verify the buffer actually contains a tab character
	lines := ui.editor.Lines()
	if len(lines) < 1 {
		t.Fatal("Expected at least one line")
	}
	line := lines[0]
	if len(line) != 11 { // hello (5) + tab (1) + world (5) = 11
		t.Errorf("Expected line length 11, got %d", len(line))
	}
	if line[5] != '\t' {
		t.Errorf("Expected tab character at position 5, got %c", line[5])
	}
}

func TestLiteralTabAtBeginning(t *testing.T) {
	ui, screen := newTestUI(40, 10)

	// Type some text first
	typeString(ui, "world")

	// Move to beginning of line
	dispatch(ui, term.KeyEvent{Key: term.KeyHome})

	// Insert a literal tab at the beginning
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'i', Modifiers: term.ModCtrl}) // Ctrl+I

	draw(ui)

	// With TabWidth=4, a tab at the beginning should render as 4 spaces
	// Col 0-3: spaces (tab expansion)
	// Col 4+: world

	// Check tab expansion at beginning
	for col := 0; col < 4; col++ {
		if screen.Cell(col, 0) != ' ' {
			t.Errorf("Expected space at column %d (tab at beginning), got %c", col, screen.Cell(col, 0))
		}
	}

	// Column 4 should be 'w' (start of "world")
	if screen.Cell(4, 0) != 'w' {
		t.Errorf("Expected 'w' at column 4, got %c", screen.Cell(4, 0))
	}

	// Verify buffer contains tab + world
	lines := ui.editor.Lines()
	if len(lines) < 1 {
		t.Fatal("Expected at least one line")
	}
	line := lines[0]
	if len(line) != 6 { // tab (1) + world (5) = 6
		t.Errorf("Expected line length 6, got %d", len(line))
	}
	if line[0] != '\t' {
		t.Errorf("Expected tab character at position 0, got %c", line[0])
	}
	if string(line[1:]) != "world" {
		t.Errorf("Expected 'world' after tab, got %s", string(line[1:]))
	}
}

func TestTabExpansionDebug(t *testing.T) {
	ui, screen := newTestUI(40, 10)

	// Insert just a tab
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'i', Modifiers: term.ModCtrl})

	draw(ui)

	// Check what's in the buffer
	lines := ui.editor.Lines()
	if len(lines) < 1 {
		t.Fatal("Expected at least one line")
	}

	t.Logf("Buffer line length: %d", len(lines[0]))
	if len(lines[0]) > 0 {
		t.Logf("First character: %q (code %d)", lines[0][0], lines[0][0])
	}

	// Check what's on screen
	screenLine := ""
	for x := 0; x < 10; x++ {
		r := screen.Cell(x, 0)
		if r == 0 {
			screenLine += "·"
		} else if r == ' ' {
			screenLine += "·"
		} else {
			screenLine += string(r)
		}
	}
	t.Logf("Screen rendering (· = space/empty): %s", screenLine)

	// Tab at position 0 should render as 4 spaces with TabWidth=4
	spaceCount := 0
	for x := 0; x < 4; x++ {
		if screen.Cell(x, 0) == ' ' {
			spaceCount++
		}
	}
	t.Logf("Space count in first 4 columns: %d", spaceCount)

	if spaceCount != 4 {
		t.Errorf("Expected 4 spaces for tab at beginning, got %d", spaceCount)
	}
}

func TestLiteralTabWithSoftWrap(t *testing.T) {
	// Test tab rendering when soft-wrap is enabled
	ui, screen := newTestUI(40, 10)
	ui.softWrap = true // Enable soft wrap
	ui.editor.TabWidth = 4

	// Type "hello" and insert a tab
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'h'})
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'e'})
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'l'})
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'l'})
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'o'})

	// Insert literal tab with Ctrl+I
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'i', Modifiers: term.ModCtrl})

	// Type "world"
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'w'})
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'o'})
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'r'})
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'l'})
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'd'})

	draw(ui)

	// Verify buffer contains "hello\tworld"
	lines := ui.editor.Lines()
	if len(lines) < 1 {
		t.Fatal("Expected at least one line")
	}

	line := lines[0]
	expected := []rune("hello\tworld")
	if len(line) != len(expected) {
		t.Errorf("Expected line length %d, got %d", len(expected), len(line))
	}

	if string(line) != string(expected) {
		t.Errorf("Expected line 'hello\\tworld', got %q", string(line))
	}

	// With soft-wrap enabled, verify tab renders as spaces
	// "hello" = 5 chars, tab should expand from column 5 to 8 (next multiple of 4)
	// So screen should show: "hello   world"
	//                        01234567890123

	screenLine := make([]rune, 0, 40)
	for x := 0; x < 13; x++ {
		r := screen.Cell(x, 0)
		if r == 0 {
			break
		}
		screenLine = append(screenLine, r)
	}

	expectedScreen := "hello   world" // 5 chars + 3 spaces (tab expansion) + 5 chars
	if string(screenLine) != expectedScreen {
		t.Errorf("With soft-wrap, expected screen to show %q, got %q", expectedScreen, string(screenLine))
	}

	// Verify the tab character at position 5 in buffer
	if line[5] != '\t' {
		t.Errorf("Expected tab character at position 5 in buffer, got %q", line[5])
	}

	// Verify spaces in the tab expansion area (columns 5-7 on screen)
	for col := 5; col < 8; col++ {
		if screen.Cell(col, 0) != ' ' {
			t.Errorf("Expected space at column %d (tab expansion with soft-wrap), got %c", col, screen.Cell(col, 0))
		}
	}
}

// TestMenuSelectedStyleDistinction verifies that menu selected items have
// a different background than the editor background for all themes
func TestMenuSelectedStyleDistinction(t *testing.T) {
	ui, _ := newTestUI(80, 24)

	// Test the default theme specifically, as it had the issue where
	// menu selected background was the same as editor background
	ui.theme = ui.config.GetTheme("default")

	editorStyle := ui.getEditorStyle()
	menuSelectedStyle := ui.getMenuSelectedStyle()
	dropdownSelectedStyle := ui.getDropdownSelectedStyle()

	// For default theme, menu selected items should use subtle dark grey background
	// to distinguish from the editor background (ColorDefault) with minimal contrast
	if ui.isDefaultTheme() {
		if menuSelectedStyle.Background != "#3A3A3A" {
			t.Errorf("Default theme menu selected style should use #3A3A3A background, got %v", menuSelectedStyle.Background)
		}
		if dropdownSelectedStyle.Background != "#3A3A3A" {
			t.Errorf("Default theme dropdown selected style should use #3A3A3A background, got %v", dropdownSelectedStyle.Background)
		}
		if editorStyle.Background == menuSelectedStyle.Background {
			t.Errorf("Default theme menu selected background should differ from editor background")
		}
	}
}
