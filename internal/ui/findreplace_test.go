package ui

import (
	"cooledit/internal/core"
	"cooledit/internal/term"
	"testing"
)

// TestFindReplaceModePreventsCharacterInsertion verifies that keys like 'n', 'p', 'q'
// in find/replace mode don't insert characters into the editor
func TestFindReplaceModePreventsCharacterInsertion(t *testing.T) {
	ui, _ := newTestUI(80, 24)
	
	// Insert text with multiple "test" occurrences
	for _, r := range "test one test two test three" {
		ui.editor.Apply(core.CmdInsertRune{Rune: r}, 10)
	}
	ui.editor.Apply(core.CmdMoveHome{}, 10)
	
	// Get initial content
	initialLines := ui.editor.Lines()
	initialText := string(initialLines[0])
	
	// Enter find mode
	ui.enterFind()
	
	// Type "test"
	for _, r := range "test" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}
	
	// Press Enter - should enter find/replace mode
	dispatch(ui, term.KeyEvent{Key: term.KeyEnter})
	
	if ui.mode != ModeFindReplace {
		t.Fatalf("should be in find/replace mode, got mode %d", ui.mode)
	}
	
	// Press 'n' for next - should NOT insert 'n' into editor
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'n'})
	
	// Verify content hasn't changed - no 'n' inserted
	currentLines := ui.editor.Lines()
	currentText := string(currentLines[0])
	if currentText != initialText {
		t.Fatalf("'n' key inserted text!\nExpected: %q\nGot: %q", initialText, currentText)
	}
	
	// Press 'p' for previous - should NOT insert 'p'
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'p'})
	
	// Verify no characters inserted
	currentLines = ui.editor.Lines()
	currentText = string(currentLines[0])
	if currentText != initialText {
		t.Fatalf("'p' key inserted text!\nExpected: %q\nGot: %q", initialText, currentText)
	}
	
	// Press 'q' to quit - should NOT insert 'q'
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'q'})
	
	// Should be back in normal mode
	if ui.mode != ModeNormal {
		t.Fatalf("should be in normal mode after 'q', got mode %d", ui.mode)
	}
	
	// Final verification - text unchanged
	currentLines = ui.editor.Lines()
	currentText = string(currentLines[0])
	if currentText != initialText {
		t.Fatalf("text changed during find/replace!\nExpected: %q\nGot: %q", initialText, currentText)
	}
}
