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

// TestFindReplaceModeComprehensiveKeyLeakage tests that ALL keys are properly consumed
// in find/replace mode and none leak through to the editor buffer
func TestFindReplaceModeComprehensiveKeyLeakage(t *testing.T) {
	ui, _ := newTestUI(80, 24)

	// Insert initial text
	for _, r := range "search term here" {
		ui.editor.Apply(core.CmdInsertRune{Rune: r}, 10)
	}
	ui.editor.Apply(core.CmdMoveHome{}, 10)

	// Get initial content
	initialLines := ui.editor.Lines()
	initialText := string(initialLines[0])

	// Enter find mode and search for "term"
	ui.enterFind()
	for _, r := range "term" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}
	dispatch(ui, term.KeyEvent{Key: term.KeyEnter})

	if ui.mode != ModeFindReplace {
		t.Fatalf("should be in find/replace mode, got mode %d", ui.mode)
	}

	// Test ALL printable ASCII characters that should be consumed
	testKeys := []rune{
		// Letters (excluding valid command keys n, p, r, a, q which are tested separately)
		'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm',
		'o', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z',
		'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M',
		'O', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z',
		// Numbers
		'0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
		// Common symbols
		'!', '@', '#', '$', '%', '^', '&', '*', '(', ')', '-', '_', '=', '+',
		'[', ']', '{', '}', '\\', '|', ';', ':', '\'', '"', ',', '.', '<', '>', '/', '?',
		'`', '~', ' ',
	}

	// Press each key and verify none insert into editor
	for _, r := range testKeys {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})

		// Verify content hasn't changed
		currentLines := ui.editor.Lines()
		currentText := string(currentLines[0])
		if currentText != initialText {
			t.Fatalf("Key %q inserted text!\nExpected: %q\nGot: %q", r, initialText, currentText)
		}
	}

	// Should still be in find/replace mode (unless we hit command keys)
	if ui.mode != ModeFindReplace {
		t.Fatalf("should still be in find/replace mode after non-command keys, got mode %d", ui.mode)
	}

	// Test special key types (arrows, delete, backspace, etc.)
	specialKeys := []term.KeyEvent{
		{Key: term.KeyUp},
		{Key: term.KeyDown},
		{Key: term.KeyLeft},
		{Key: term.KeyRight},
		{Key: term.KeyBackspace},
		{Key: term.KeyDelete},
		{Key: term.KeyTab},
		{Key: term.KeyHome},
		{Key: term.KeyEnd},
		{Key: term.KeyPageUp},
		{Key: term.KeyPageDown},
	}

	for _, ke := range specialKeys {
		dispatch(ui, ke)

		// Verify content hasn't changed
		currentLines := ui.editor.Lines()
		currentText := string(currentLines[0])
		if currentText != initialText {
			t.Fatalf("Special key %v inserted text!\nExpected: %q\nGot: %q", ke.Key, initialText, currentText)
		}
	}

	// Exit cleanly
	dispatch(ui, term.KeyEvent{Key: term.KeyEscape})

	// Final verification
	finalLines := ui.editor.Lines()
	finalText := string(finalLines[0])
	if finalText != initialText {
		t.Fatalf("Final text changed!\nExpected: %q\nGot: %q", initialText, finalText)
	}
}

// TestReplaceAllConfirmation tests the replace-all confirmation dialog workflow
func TestReplaceAllConfirmation(t *testing.T) {
	ui, _ := newTestUI(80, 24)

	// Insert text with multiple occurrences
	for _, r := range "foo bar foo baz foo" {
		ui.editor.Apply(core.CmdInsertRune{Rune: r}, 10)
	}
	ui.editor.Apply(core.CmdMoveHome{}, 10)

	// Verify initial text
	lines := ui.editor.Lines()
	text := string(lines[0])
	if text != "foo bar foo baz foo" {
		t.Fatalf("initial text wrong: %q", text)
	}

	// Use CmdReplaceAll directly first to verify it works
	res := ui.editor.Apply(core.CmdReplaceAll{Find: "foo", Replace: "XXX"}, 10)
	t.Logf("Direct CmdReplaceAll result: %s", res.Message)

	lines = ui.editor.Lines()
	text = string(lines[0])
	t.Logf("After direct CmdReplaceAll: %q", text)

	expected := "XXX bar XXX baz XXX"
	if text != expected {
		t.Fatalf("direct replace-all failed: expected %q, got %q", expected, text)
	}

	// Undo it
	ui.editor.Apply(core.CmdUndo{}, 10)
	lines = ui.editor.Lines()
	text = string(lines[0])
	if text != "foo bar foo baz foo" {
		t.Fatalf("undo failed: expected %q, got %q", "foo bar foo baz foo", text)
	}

	// Now test the full UI flow...
	t.Log("Testing full UI flow with confirmation dialog")

	// Enter find mode and search for "foo"
	ui.enterFind()
	for _, r := range "foo" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}
	dispatch(ui, term.KeyEvent{Key: term.KeyEnter})

	if ui.mode != ModeFindReplace {
		t.Fatalf("should be in find/replace mode, got mode %d", ui.mode)
	}

	t.Logf("lastFindTerm: %q", ui.lastFindTerm)

	// Press 'a' for replace all
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'a'})

	if ui.mode != ModePrompt || ui.promptKind != PromptReplaceAllConfirm {
		t.Fatalf("should show confirmation, got mode %d, promptKind %d", ui.mode, ui.promptKind)
	}

	t.Logf("Confirmation prompt: %q", ui.promptLabel)

	// Confirm
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'y'})

	if ui.mode != ModePrompt || ui.promptKind != PromptReplaceWith {
		t.Fatalf("should prompt for replacement, got mode %d, promptKind %d", ui.mode, ui.promptKind)
	}

	t.Logf("replacingAll flag: %v", ui.replacingAll)

	// Type "XXX"
	for _, r := range "XXX" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}

	t.Logf("promptText before Enter: %q", string(ui.promptText))

	// Execute
	dispatch(ui, term.KeyEvent{Key: term.KeyEnter})

	// Check result
	lines = ui.editor.Lines()
	text = string(lines[0])
	t.Logf("After UI replace-all: %q", text)

	if text != expected {
		t.Fatalf("UI replace-all failed: expected %q, got %q", expected, text)
	}
}
