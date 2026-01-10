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
	"testing"
	"time"

	"cooledit/internal/core"
	"cooledit/internal/term"
)

// waitForSearch waits for any pending debounced search to complete
func waitForSearch(ui *UI) {
	if ui.searchDebounceTimer != nil {
		time.Sleep(200 * time.Millisecond) // Wait longer than the 150ms debounce
	}
}

// TestSearchTermPersistence verifies that search terms persist across searches.
func TestSearchTermPersistence(t *testing.T) {
	ui, _ := newTestUI(80, 24)

	// Insert some text
	for _, r := range "hello world\nhello universe" {
		ui.editor.Apply(core.CmdInsertRune{Rune: r}, 10)
	}
	ui.editor.Apply(core.CmdMoveHome{}, 10)

	// First search
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'f', Modifiers: term.ModCtrl})
	if ui.mode != ModeSearch {
		t.Fatal("should be in search mode")
	}

	// Type "hello"
	for _, r := range "hello" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}

	// Exit search
	dispatch(ui, term.KeyEvent{Key: term.KeyEscape})
	if ui.mode != ModeNormal {
		t.Fatal("should be in normal mode")
	}

	// Verify lastSearchQuery was saved
	if ui.lastSearchQuery != "hello" {
		t.Errorf("expected lastSearchQuery='hello', got '%s'", ui.lastSearchQuery)
	}

	// Enter search again - should pre-fill with last query
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'f', Modifiers: term.ModCtrl})
	if string(ui.searchQuery) != "hello" {
		t.Errorf("expected searchQuery to be pre-filled with 'hello', got '%s'", string(ui.searchQuery))
	}

	dispatch(ui, term.KeyEvent{Key: term.KeyEscape})
}

// TestReplaceTermPersistence verifies that replace terms persist across replacements.
func TestReplaceTermPersistence(t *testing.T) {
	ui, _ := newTestUI(80, 24)

	// Insert text with multiple matches
	for _, r := range "foo bar foo baz" {
		ui.editor.Apply(core.CmdInsertRune{Rune: r}, 10)
	}
	ui.editor.Apply(core.CmdMoveHome{}, 10)

	// Search for "foo"
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'f', Modifiers: term.ModCtrl})
	for _, r := range "foo" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}
	
	// Wait for debounced search to complete
	waitForSearch(ui)

	// Press 'r' to enter replace prompt
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'r'})
	if ui.mode != ModePrompt || ui.promptKind != PromptReplaceWith {
		t.Fatal("should be in replace prompt")
	}

	// Type replacement text "qux"
	for _, r := range "qux" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}

	// Press Enter to replace
	dispatch(ui, term.KeyEvent{Key: term.KeyEnter})

	// Verify lastReplaceTerm was saved
	if ui.lastReplaceTerm != "qux" {
		t.Errorf("expected lastReplaceTerm='qux', got '%s'", ui.lastReplaceTerm)
	}

	// Enter another search and try replace again - should pre-fill with last replace term
	dispatch(ui, term.KeyEvent{Key: term.KeyEscape}) // Exit any mode
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'f', Modifiers: term.ModCtrl})
	// Clear the pre-filled query ("foo")
	for len(ui.searchQuery) > 0 {
		dispatch(ui, term.KeyEvent{Key: term.KeyBackspace})
	}
	// Use "foo" again since it has matches
	for _, r := range "foo" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}
	
	// Wait for debounced search
	waitForSearch(ui)

	// Press 'r' to enter replace prompt
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'r'})
	if ui.mode != ModePrompt {
		t.Fatalf("should be in prompt mode, got mode %d", ui.mode)
	}
	if ui.promptKind != PromptReplaceWith {
		t.Fatalf("should be in replace prompt, got promptKind %d", ui.promptKind)
	}
	if string(ui.promptText) != "qux" {
		t.Errorf("expected promptText to be pre-filled with 'qux', got '%s' (lastReplaceTerm='%s')", string(ui.promptText), ui.lastReplaceTerm)
	}

	dispatch(ui, term.KeyEvent{Key: term.KeyEscape})
}

// TestCaseSensitivityPersistence verifies that case sensitivity setting persists.
func TestCaseSensitivityPersistence(t *testing.T) {
	ui, _ := newTestUI(80, 24)

	// Insert text with mixed case
	for _, r := range "Hello hello HELLO" {
		ui.editor.Apply(core.CmdInsertRune{Rune: r}, 10)
	}
	ui.editor.Apply(core.CmdMoveHome{}, 10)

	// Enter search
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'f', Modifiers: term.ModCtrl})
	for _, r := range "hello" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}
	
	// Wait for debounced search
	waitForSearch(ui)

	// Initially should be case-insensitive (default) - 3 matches
	session := ui.editor.GetSearchSession()
	if session == nil {
		t.Fatal("search session should exist")
	}
	if len(session.Matches) != 3 {
		t.Errorf("expected 3 matches (case-insensitive), got %d", len(session.Matches))
	}

	// Toggle case sensitivity with Alt+C
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'c', Modifiers: term.ModAlt})
	
	// Wait for search to re-execute
	waitForSearch(ui)

	// Should now be case-sensitive - only 1 match (exact "hello")
	session = ui.editor.GetSearchSession()
	if session == nil {
		t.Fatal("search session should exist")
	}
	if len(session.Matches) != 1 {
		t.Errorf("expected 1 match (case-sensitive), got %d", len(session.Matches))
	}

	// Exit search
	dispatch(ui, term.KeyEvent{Key: term.KeyEscape})

	// Verify case sensitivity persisted in editor's search state
	searchState := ui.editor.SearchState()
	if !searchState.CaseSensitive {
		t.Error("case sensitivity should be true after toggle")
	}

	// Enter search again - case sensitivity should still be on
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'f', Modifiers: term.ModCtrl})
	// Clear the pre-filled query
	for len(ui.searchQuery) > 0 {
		dispatch(ui, term.KeyEvent{Key: term.KeyBackspace})
	}
	for _, r := range "hello" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}
	
	// Wait for search
	waitForSearch(ui)

	session = ui.editor.GetSearchSession()
	if session == nil {
		t.Fatal("search session should exist")
	}
	if len(session.Matches) != 1 {
		t.Errorf("expected 1 match (case sensitivity persisted), got %d", len(session.Matches))
	}

	dispatch(ui, term.KeyEvent{Key: term.KeyEscape})
}

// TestWholeWordPersistence verifies that whole word setting persists.
func TestWholeWordPersistence(t *testing.T) {
	ui, _ := newTestUI(80, 24)

	// Insert text with partial and whole word matches
	for _, r := range "cat catch cathedral" {
		ui.editor.Apply(core.CmdInsertRune{Rune: r}, 10)
	}
	ui.editor.Apply(core.CmdMoveHome{}, 10)

	// Enter search
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'f', Modifiers: term.ModCtrl})
	for _, r := range "cat" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}
	
	// Wait for search
	waitForSearch(ui)

	// Initially should match all occurrences (not whole word) - 3 matches
	session := ui.editor.GetSearchSession()
	if session == nil {
		t.Fatal("search session should exist")
	}
	if len(session.Matches) != 3 {
		t.Errorf("expected 3 matches (not whole word), got %d", len(session.Matches))
	}

	// Toggle whole word with Alt+W
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'w', Modifiers: term.ModAlt})
	
	// Wait for search to re-execute
	waitForSearch(ui)

	// Should now match only whole word "cat" - 1 match
	session = ui.editor.GetSearchSession()
	if session == nil {
		t.Fatal("search session should exist")
	}
	if len(session.Matches) != 1 {
		t.Errorf("expected 1 match (whole word), got %d", len(session.Matches))
	}

	// Exit search
	dispatch(ui, term.KeyEvent{Key: term.KeyEscape})

	// Verify whole word persisted in editor's search state
	searchState := ui.editor.SearchState()
	if !searchState.WholeWord {
		t.Error("whole word should be true after toggle")
	}

	// Enter search again - whole word should still be on
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'f', Modifiers: term.ModCtrl})
	// Clear the pre-filled query
	for len(ui.searchQuery) > 0 {
		dispatch(ui, term.KeyEvent{Key: term.KeyBackspace})
	}
	for _, r := range "cat" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}
	
	// Wait for search
	waitForSearch(ui)

	session = ui.editor.GetSearchSession()
	if session == nil {
		t.Fatal("search session should exist")
	}
	if len(session.Matches) != 1 {
		t.Errorf("expected 1 match (whole word persisted), got %d", len(session.Matches))
	}

	dispatch(ui, term.KeyEvent{Key: term.KeyEscape})
}

// TestSearchHistoryPersistence verifies that search history persists and can be navigated.
func TestSearchHistoryPersistence(t *testing.T) {
	ui, _ := newTestUI(80, 24)

	for _, r := range "test data" {
		ui.editor.Apply(core.CmdInsertRune{Rune: r}, 10)
	}
	ui.editor.Apply(core.CmdMoveHome{}, 10)

	// Perform first search (using words without special command letters)
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'f', Modifiers: term.ModCtrl})
	// Clear any pre-filled query
	for len(ui.searchQuery) > 0 {
		dispatch(ui, term.KeyEvent{Key: term.KeyBackspace})
	}
	for _, r := range "hello" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}
	dispatch(ui, term.KeyEvent{Key: term.KeyEscape})

	// Perform second search
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'f', Modifiers: term.ModCtrl})
	// Clear pre-filled query
	for len(ui.searchQuery) > 0 {
		dispatch(ui, term.KeyEvent{Key: term.KeyBackspace})
	}
	for _, r := range "world" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}
	dispatch(ui, term.KeyEvent{Key: term.KeyEscape})

	// Perform third search
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'f', Modifiers: term.ModCtrl})
	// Clear pre-filled query
	for len(ui.searchQuery) > 0 {
		dispatch(ui, term.KeyEvent{Key: term.KeyBackspace})
	}
	for _, r := range "test" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}
	dispatch(ui, term.KeyEvent{Key: term.KeyEscape})

	// Enter search and navigate history
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'f', Modifiers: term.ModCtrl})

	// Press Up to go to previous query (test)
	dispatch(ui, term.KeyEvent{Key: term.KeyUp})
	if string(ui.searchQuery) != "test" {
		t.Errorf("expected 'test', got '%s'", string(ui.searchQuery))
	}

	// Press Up again to go to world
	dispatch(ui, term.KeyEvent{Key: term.KeyUp})
	if string(ui.searchQuery) != "world" {
		t.Errorf("expected 'world', got '%s'", string(ui.searchQuery))
	}

	// Press Up again to go to hello
	dispatch(ui, term.KeyEvent{Key: term.KeyUp})
	if string(ui.searchQuery) != "hello" {
		t.Errorf("expected 'hello', got '%s'", string(ui.searchQuery))
	}

	// Press Down to go back to world
	dispatch(ui, term.KeyEvent{Key: term.KeyDown})
	if string(ui.searchQuery) != "world" {
		t.Errorf("expected 'world', got '%s'", string(ui.searchQuery))
	}

	dispatch(ui, term.KeyEvent{Key: term.KeyEscape})
}
