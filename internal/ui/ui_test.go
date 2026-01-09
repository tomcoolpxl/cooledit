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
	ui.layout = ComputeLayout(w, h, ui.mode, ui.showMenubar)
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

		if ui.mode == ModeFindReplace {
			if ui.handleFindReplaceKey(e) {
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

	case term.MouseEvent:
		ui.handleMouseEvent(e)
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

func TestCtrlFEnterFind(t *testing.T) {
	ui, screen := newTestUI(40, 5)

	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'f', Modifiers: term.ModCtrl})

	if ui.mode != ModePrompt || ui.promptKind != PromptFind {
		t.Fatalf("expected mode to be PromptFind")
	}

	draw(ui)
	row := 3
	if got := screen.Cell(0, row); got != 'F' {
		t.Fatalf("expected Find prompt, got %q", got)
	}
}

func TestPromptInteraction(t *testing.T) {
	ui, screen := newTestUI(40, 5)

	ui.enterFind()
	draw(ui) // update layout to Prompt mode

	typeString(ui, "abc")

	if string(ui.promptText) != "abc" {
		t.Fatalf("expected prompt text 'abc', got %q", string(ui.promptText))
	}

	dispatch(ui, term.KeyEvent{Key: term.KeyBackspace})
	if string(ui.promptText) != "ab" {
		t.Fatalf("expected prompt text 'ab' after backspace, got %q", string(ui.promptText))
	}

	dispatch(ui, term.KeyEvent{Key: term.KeyEscape})
	if ui.mode != ModeNormal {
		t.Fatalf("expected mode Normal after Escape")
	}

	draw(ui)
	if screen.Cell(0, 3) != ' ' {
		t.Fatalf("expected prompt row to be clear")
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

	// Ctrl+F
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'f', Modifiers: term.ModCtrl})
	draw(ui) // updates layout to prompt

	typeString(ui, "foo")
	dispatch(ui, term.KeyEvent{Key: term.KeyEnter})

	draw(ui) // Now in ModeFindReplace with selection

	// Should be in find/replace mode
	if ui.mode != ModeFindReplace {
		t.Fatalf("expected ModeFindReplace, got mode %d", ui.mode)
	}

	// Cursor should be at start of first match (0,0)
	line, col := ui.editor.Cursor()
	if line != 0 || col != 0 {
		t.Fatalf("expected cursor at (0,0), got (%d,%d)", line, col)
	}

	// F3 or N (Next) - test N key in find/replace mode
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'n'})
	draw(ui)

	// Should find second "foo" at 0, 8
	line, col = ui.editor.Cursor()
	if line != 0 || col != 8 {
		t.Fatalf("expected cursor at (0,8) after next, got (%d,%d)", line, col)
	}

	// P (Prev) in find/replace mode
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'p'})
	draw(ui)

	// Should find first "foo" at 0, 0
	line, col = ui.editor.Cursor()
	if line != 0 || col != 0 {
		t.Fatalf("expected cursor at (0,0) after prev, got (%d,%d)", line, col)
	}

	// Q to quit find mode
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'q'})

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

func TestMouseClickMovesCursor(t *testing.T) {
	ui, screen := newTestUI(20, 5)
	ui.showMenubar = false

	// Fill buffer:
	// aaaa
	// bbbb
	typeString(ui, "aaaa")
	dispatch(ui, term.KeyEvent{Key: term.KeyEnter})
	typeString(ui, "bbbb")

	// Click on (2, 1) -> Row 1, Col 2 (zero-based) -> 'b' at index 2
	// Viewport starts at (0,0) without menubar
	dispatch(ui, term.MouseEvent{X: 2, Y: 1, Button: term.MouseLeft})
	draw(ui)

	if screen.cursorY != 1 || screen.cursorX != 2 {
		t.Fatalf("expected cursor at (2,1), got (%d,%d)", screen.cursorX, screen.cursorY)
	}

	// Verify doc position logic
	row, col := ui.editor.Cursor()
	if row != 1 || col != 2 {
		t.Fatalf("expected doc cursor (1,2), got (%d,%d)", row, col)
	}
}

func TestMouseWheelScrolling(t *testing.T) {
	ui, screen := newTestUI(20, 5)

	// 10 lines
	for i := 0; i < 10; i++ {
		typeString(ui, "x")
		dispatch(ui, term.KeyEvent{Key: term.KeyEnter})
	}

	// Cursor at bottom.
	dispatch(ui, term.KeyEvent{Key: term.KeyHome, Modifiers: term.ModCtrl})
	draw(ui)

	if screen.Cell(0, 0) != 'x' {
		t.Fatal("expected 'x' at top")
	}

	// Scroll Down (simulates moving viewport down)
	dispatch(ui, term.MouseEvent{Button: term.MouseWheelDown})
	draw(ui)

	// Cursor should have moved down (CmdMoveDown called 3 times)
	row, _ := ui.editor.Cursor()
	if row != 3 {
		t.Fatalf("expected cursor row 3 after scroll down, got %d", row)
	}

	// Scroll Up
	dispatch(ui, term.MouseEvent{Button: term.MouseWheelUp})
	draw(ui)

	row, _ = ui.editor.Cursor()
	if row != 0 {
		t.Fatalf("expected cursor row 0 after scroll up, got %d", row)
	}
}

func TestMouseMenuInteraction(t *testing.T) {
	ui, _ := newTestUI(40, 10)
	ui.showMenubar = true
	draw(ui)

	// Click "File" (0,0 to 6,0 approx)
	dispatch(ui, term.MouseEvent{X: 2, Y: 0, Button: term.MouseLeft})
	draw(ui)

	if ui.mode != ModeMenu || !ui.menubar.Active {
		t.Fatal("expected menu mode active")
	}
	if ui.menubar.SelectedMenuIndex != 0 { // File is index 0
		t.Fatal("expected File menu selected")
	}

	// Click outside (in viewport) -> should close menu
	dispatch(ui, term.MouseEvent{X: 0, Y: 5, Button: term.MouseLeft})
	draw(ui)

	if ui.mode != ModeNormal {
		t.Fatal("expected normal mode after click outside")
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
	if firstLineChar != 'C' { // "CoolEdit"
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

func TestMouseEdgeCases(t *testing.T) {
	ui, _ := newTestUI(40, 5)
	ui.showMenubar = true
	draw(ui)

	// Click Status Bar (Row 4) - should do nothing (no crash)
	dispatch(ui, term.MouseEvent{X: 0, Y: 4, Button: term.MouseLeft})
	if ui.mode != ModeNormal {
		t.Fatalf("status bar click changed mode")
	}

	// Enter prompt mode
	ui.enterFind()
	draw(ui)
	// Now Prompt is at Row 3.
	dispatch(ui, term.MouseEvent{X: 0, Y: 3, Button: term.MouseLeft})

	if ui.mode != ModePrompt {
		t.Fatalf("prompt click exited prompt mode")
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
	// Test that status bar shows replace options in find/replace mode
	ui, screen := newTestUI(80, 5)

	// Type some text and search for it
	typeString(ui, "hello world")
	dispatch(ui, term.KeyEvent{Key: term.KeyHome, Modifiers: term.ModCtrl})

	// Enter find mode
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'f', Modifiers: term.ModCtrl})
	draw(ui)

	typeString(ui, "hello")
	dispatch(ui, term.KeyEvent{Key: term.KeyEnter})
	draw(ui)

	// Should be in ModeFindReplace
	if ui.mode != ModeFindReplace {
		t.Fatalf("expected ModeFindReplace, got mode %d", ui.mode)
	}

	// Status bar should show [R]eplace options
	statusBarY := 4
	foundReplace := false

	for x := 0; x < 30; x++ {
		if screen.Cell(x, statusBarY) == '[' && screen.Cell(x+1, statusBarY) == 'R' &&
			screen.Cell(x+2, statusBarY) == ']' {
			foundReplace = true
			break
		}
	}

	if !foundReplace {
		t.Fatalf("expected '[R]eplace' in status bar during find/replace mode")
	}
}

func TestHelpScreenWideTerminal(t *testing.T) {
	// Test two-column layout on wide terminal
	ui, screen := newTestUI(90, 25)

	dispatch(ui, term.KeyEvent{Key: term.KeyF1})
	draw(ui)

	if ui.mode != ModeHelp {
		t.Fatal("expected Help mode")
	}

	// Should have "CoolEdit" title at top
	if screen.Cell(2, 0) != 'C' {
		t.Fatalf("expected 'C' from CoolEdit title")
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
	if screen.Cell(2, 0) != 'C' {
		t.Fatalf("expected 'C' from CoolEdit title")
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
