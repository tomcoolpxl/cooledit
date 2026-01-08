package ui

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"cooledit/internal/core"
	"cooledit/internal/term"
)

func newTestUI(w, h int) (*UI, *FakeScreen) {
	screen := NewFakeScreen(w, h)
	editor := core.NewEditor()
	ui := New(screen, editor)
	ui.showMenubar = false // Disable by default for tests to match old layout assumptions
	return ui, screen
}

func updateTestLayout(ui *UI, w, h int) {
	ui.layout = ComputeLayout(w, h, ui.mode, ui.showMenubar)
}

func draw(ui *UI) {
	w, h := ui.screen.Size()
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

		if e.Key == term.KeyRune && e.Rune == 'c' && (e.Modifiers&term.ModCtrl) != 0 {
			if ui.handleCtrlC(e) {
				ui.quitNow = true // Simulate Run loop exit
				return
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
	ui, screen := newTestUI(20, 5)
	
typeString(ui, "foo bar foo")
	dispatch(ui, term.KeyEvent{Key: term.KeyHome, Modifiers: term.ModCtrl}) // Go to start
	
	// Ctrl+F
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'f', Modifiers: term.ModCtrl})
	draw(ui) // updates layout to prompt
	
typeString(ui, "foo")
	dispatch(ui, term.KeyEvent{Key: term.KeyEnter})
	
draw(ui) // updates layout back to normal, draws cursor
	
	// Cursor should be at 0,0
	if screen.cursorX != 0 || screen.cursorY != 0 {
		t.Fatalf("expected cursor at (0,0), got (%d,%d)", screen.cursorX, screen.cursorY)
	}
	
	// F3 (Next)
	dispatch(ui, term.KeyEvent{Key: term.KeyF3})
	draw(ui)
	
	// Should find second "foo" at 0, 8
	if screen.cursorX != 8 {
		t.Fatalf("expected cursor at (8,0) after F3, got (%d,%d)", screen.cursorX, screen.cursorY)
	}
	
	// Shift+F3 (Prev)
	dispatch(ui, term.KeyEvent{Key: term.KeyF3, Modifiers: term.ModShift})
	draw(ui)
	
	// Should find first "foo" at 0, 0
	if screen.cursorX != 0 {
		t.Fatalf("expected cursor at (0,0) after Shift+F3, got (%d,%d)", screen.cursorX, screen.cursorY)
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
	
	// Screen should show help text (not empty)
	if screen.Cell(0, 0) != 'c' { // "cooledit - help"
		t.Fatalf("expected help text, got %q", screen.Cell(0, 0))
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
	ui.showMenubar = true
	
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

func TestCtrlCExpiration(t *testing.T) {
	ui, screen := newTestUI(40, 5)
	
	// First Ctrl+C
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'c', Modifiers: term.ModCtrl})
	if !ui.ctrlCArmed {
		t.Fatalf("expected ctrlCArmed true")
	}
	draw(ui)
	// Check message "Press Ctrl+C again..." (P)
	if screen.Cell(0, 3) != 'P' {
		t.Fatalf("expected status message")
	}
	
	// Force expire
	ui.ctrlCUntil = time.Now().Add(-1 * time.Second)
	
	// Second Ctrl+C (should act as First because expired)
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'c', Modifiers: term.ModCtrl})
	
	if !ui.ctrlCArmed {
		t.Fatalf("expected ctrlCArmed true (re-armed)")
	}
	// Should NOT have quit (quitNow false)
	if ui.quitNow {
		t.Fatalf("should not quit if expired")
	}
	
	// Third Ctrl+C (immediate) -> Quit
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'c', Modifiers: term.ModCtrl})
	if !ui.quitNow {
		t.Fatalf("expected quitNow true")
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
	ui.showMenubar = true
	
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
