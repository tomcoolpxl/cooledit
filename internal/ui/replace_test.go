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
	"cooledit/internal/core"
	"cooledit/internal/term"
	"strings"
	"testing"
)

func TestReplaceCurrent(t *testing.T) {
	ui, _ := newTestUI(80, 24)

	// Insert text with matches
	for _, r := range "hello world hello" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'a', Modifiers: term.ModCtrl})

	// Enter search mode
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'f', Modifiers: term.ModCtrl})

	// Search for "hello"
	for _, r := range "hello" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}
	ui.doSearch()

	// Press Ctrl+R to replace current match
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'r', Modifiers: term.ModCtrl})

	// Should enter prompt mode for replacement
	if ui.mode != ModePrompt {
		t.Errorf("expected ModePrompt after Ctrl+R, got %v", ui.mode)
	}

	// Type replacement text
	for _, r := range "hi" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}

	// Press Enter to confirm
	dispatch(ui, term.KeyEvent{Key: term.KeyEnter})

	// Should return to search mode
	if ui.mode != ModeSearch {
		t.Errorf("expected ModeSearch after replacement, got %v", ui.mode)
	}

	// Verify replacement happened
	lines := ui.editor.Lines()
	text := string(lines[0])
	if !strings.Contains(text, "hi") || strings.Count(text, "hello") != 1 {
		t.Errorf("expected one 'hello' replaced with 'hi', got '%s'", text)
	}
}

func TestReplaceAllConfirmationDialog(t *testing.T) {
	ui, _ := newTestUI(80, 24)

	// Insert text with multiple matches
	for _, r := range "test test test" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'a', Modifiers: term.ModCtrl})

	// Enter search mode
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'f', Modifiers: term.ModCtrl})

	// Search for "test"
	for _, r := range "test" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}
	ui.doSearch()

	// Press Ctrl+H to replace all
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'h', Modifiers: term.ModCtrl})

	// Should show confirmation prompt
	if ui.mode != ModePrompt {
		t.Errorf("expected ModePrompt for replace all, got %v", ui.mode)
	}

	// Type replacement
	for _, r := range "exam" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}

	// Press Enter to start replace all
	dispatch(ui, term.KeyEvent{Key: term.KeyEnter})

	// This should trigger confirmation (ideally)
	// For now, verify the prompt was shown
}

func TestReplaceAllUndo(t *testing.T) {
	ui, _ := newTestUI(80, 24)

	// Insert text
	for _, r := range "foo bar foo baz foo" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}

	// Save original text
	originalText := string(ui.editor.Lines()[0])

	// Move to start
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'a', Modifiers: term.ModCtrl})

	// Perform replace all via command
	res := ui.editor.Apply(core.CmdReplaceAll{Find: "foo", Replace: "bar"}, 10)

	// Verify replacements happened
	if !strings.Contains(res.Message, "Replaced 3 occurrences") {
		t.Errorf("expected replace all message, got: %s", res.Message)
	}

	lines := ui.editor.Lines()
	text := string(lines[0])
	expected := "bar bar bar baz bar"
	if text != expected {
		t.Errorf("expected '%s', got '%s'", expected, text)
	}

	// Undo should revert ALL replacements in one operation
	ui.editor.Apply(core.CmdUndo{}, 10)

	lines = ui.editor.Lines()
	text = string(lines[0])
	if text != originalText {
		t.Errorf("expected '%s' after undo, got '%s'", originalText, text)
	}
}

func TestReplaceWithEmptyString(t *testing.T) {
	ui, _ := newTestUI(80, 24)

	// Insert text
	for _, r := range "hello world" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'a', Modifiers: term.ModCtrl})

	// Perform replace with empty string (deletion)
	res := ui.editor.Apply(core.CmdReplaceAll{Find: "hello ", Replace: ""}, 10)

	// Verify deletion happened
	if !strings.Contains(res.Message, "Replaced") {
		t.Errorf("expected replace message, got: %s", res.Message)
	}

	lines := ui.editor.Lines()
	text := string(lines[0])
	if text != "world" {
		t.Errorf("expected 'world', got '%s'", text)
	}
}

func TestReplaceAllCancel(t *testing.T) {
	ui, _ := newTestUI(80, 24)

	// Insert text
	for _, r := range "test test test" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}

	// Save original text
	originalText := string(ui.editor.Lines()[0])

	// Move to start
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'a', Modifiers: term.ModCtrl})

	// Enter search mode
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'f', Modifiers: term.ModCtrl})

	// Search
	for _, r := range "test" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}
	ui.doSearch()

	// Press Ctrl+H for replace all
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'h', Modifiers: term.ModCtrl})

	// Should be in prompt
	if ui.mode != ModePrompt {
		t.Fatal("expected prompt mode")
	}

	// Type replacement
	for _, r := range "exam" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}

	// Press Escape to cancel
	dispatch(ui, term.KeyEvent{Key: term.KeyEscape})

	// Should return to search mode without replacing
	if ui.mode != ModeSearch {
		t.Errorf("expected ModeSearch after cancel, got %v", ui.mode)
	}

	// Text should be unchanged
	lines := ui.editor.Lines()
	text := string(lines[0])
	if text != originalText {
		t.Errorf("expected text unchanged after cancel, got '%s'", text)
	}
}

func TestReplaceStatusMessage(t *testing.T) {
	ui, _ := newTestUI(80, 24)

	// Insert text
	for _, r := range "foo foo foo" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'a', Modifiers: term.ModCtrl})

	// Perform replace all
	res := ui.editor.Apply(core.CmdReplaceAll{Find: "foo", Replace: "bar"}, 10)

	// Verify message includes count and undo hint
	if !strings.Contains(res.Message, "Replaced 3 occurrences") {
		t.Errorf("expected count in message, got: %s", res.Message)
	}

	if !strings.Contains(res.Message, "Ctrl+Z to undo") {
		t.Errorf("expected undo hint in message, got: %s", res.Message)
	}
}

func TestReplacePreservesCase(t *testing.T) {
	ui, _ := newTestUI(80, 24)

	// Insert mixed case text
	for _, r := range "Hello hello HELLO" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'a', Modifiers: term.ModCtrl})

	// Case-sensitive replace should only replace exact matches
	ui.editor.SearchState().CaseSensitive = true
	res := ui.editor.Apply(core.CmdReplaceAll{Find: "hello", Replace: "hi"}, 10)

	// Should only replace the lowercase "hello"
	if !strings.Contains(res.Message, "Replaced 1 occurrence") {
		t.Errorf("expected 1 replacement (case-sensitive), got: %s", res.Message)
	}

	lines := ui.editor.Lines()
	text := string(lines[0])
	expected := "Hello hi HELLO"
	if text != expected {
		t.Errorf("expected '%s', got '%s'", expected, text)
	}
}

func TestReplaceCaseInsensitive(t *testing.T) {
	ui, _ := newTestUI(80, 24)

	// Insert mixed case text
	for _, r := range "Hello hello HELLO" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'a', Modifiers: term.ModCtrl})

	// Case-insensitive replace should replace all variants
	ui.editor.SearchState().CaseSensitive = false
	res := ui.editor.Apply(core.CmdReplaceAll{Find: "hello", Replace: "hi"}, 10)

	// Should replace all 3 occurrences
	if !strings.Contains(res.Message, "Replaced 3 occurrences") {
		t.Errorf("expected 3 replacements (case-insensitive), got: %s", res.Message)
	}

	lines := ui.editor.Lines()
	text := string(lines[0])
	expected := "hi hi hi"
	if text != expected {
		t.Errorf("expected '%s', got '%s'", expected, text)
	}
}

func TestReplaceNoMatches(t *testing.T) {
	ui, _ := newTestUI(80, 24)

	// Insert text
	for _, r := range "hello world" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'a', Modifiers: term.ModCtrl})

	// Try to replace something that doesn't exist
	res := ui.editor.Apply(core.CmdReplaceAll{Find: "xyz", Replace: "abc"}, 10)

	// Should report no matches
	if !strings.Contains(res.Message, "Not found") && !strings.Contains(res.Message, "No matches found") {
		t.Errorf("expected 'Not found' or 'No matches found' message, got: %s", res.Message)
	}

	// Text should be unchanged
	lines := ui.editor.Lines()
	text := string(lines[0])
	if text != "hello world" {
		t.Errorf("expected text unchanged, got '%s'", text)
	}
}

func TestReplaceAtEndOfLine(t *testing.T) {
	ui, _ := newTestUI(80, 24)

	// Insert text ending with search term
	for _, r := range "hello world test" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'a', Modifiers: term.ModCtrl})

	// Replace term at end of line
	res := ui.editor.Apply(core.CmdReplaceAll{Find: "test", Replace: "exam"}, 10)

	// Should succeed
	if !strings.Contains(res.Message, "Replaced 1 occurrence") {
		t.Errorf("expected success, got: %s", res.Message)
	}

	lines := ui.editor.Lines()
	text := string(lines[0])
	if text != "hello world exam" {
		t.Errorf("expected 'hello world exam', got '%s'", text)
	}
}

func TestReplaceMultipleLines(t *testing.T) {
	ui, _ := newTestUI(80, 24)

	// Insert multi-line text
	for _, r := range "test line one" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}
	dispatch(ui, term.KeyEvent{Key: term.KeyEnter})
	for _, r := range "test line two" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}
	dispatch(ui, term.KeyEvent{Key: term.KeyEnter})
	for _, r := range "test line three" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}

	// Move to start
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'a', Modifiers: term.ModCtrl})

	// Replace "test" across all lines
	res := ui.editor.Apply(core.CmdReplaceAll{Find: "test", Replace: "exam"}, 10)

	// Should replace in all 3 lines
	if !strings.Contains(res.Message, "Replaced 3 occurrences") {
		t.Errorf("expected 3 replacements, got: %s", res.Message)
	}

	// Verify each line
	lines := ui.editor.Lines()
	if string(lines[0]) != "exam line one" {
		t.Errorf("line 0: expected 'exam line one', got '%s'", string(lines[0]))
	}
	if string(lines[1]) != "exam line two" {
		t.Errorf("line 1: expected 'exam line two', got '%s'", string(lines[1]))
	}
	if string(lines[2]) != "exam line three" {
		t.Errorf("line 2: expected 'exam line three', got '%s'", string(lines[2]))
	}
}
