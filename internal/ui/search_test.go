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
	"cooledit/internal/term"
	"testing"
)

func TestUnifiedSearchBasic(t *testing.T) {
	ui, _ := newTestUI(80, 24)

	// Insert some text
	for _, r := range "hello world" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}

	// Enter search mode
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'f', Modifiers: term.ModCtrl})
	if ui.mode != ModeSearch {
		t.Fatalf("expected ModeSearch, got %v", ui.mode)
	}

	// Type search query
	for _, r := range "hello" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}

	// Verify search query
	if string(ui.searchQuery) != "hello" {
		t.Errorf("expected search query 'hello', got '%s'", string(ui.searchQuery))
	}

	// Exit search
	dispatch(ui, term.KeyEvent{Key: term.KeyEscape})
	if ui.mode != ModeNormal {
		t.Fatalf("expected ModeNormal after escape, got %v", ui.mode)
	}
}

func TestUnifiedSearchRealTime(t *testing.T) {
	ui, _ := newTestUI(80, 24)

	// Insert test content with multiple matches
	for _, r := range "test testing tester" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}

	// Move to start
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'a', Modifiers: term.ModCtrl})

	// Enter search mode
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'f', Modifiers: term.ModCtrl})

	// Type search query character by character
	// Each character should trigger a search
	for _, r := range "test" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}

	// Wait for debounce (in real UI) - in test, search happens immediately
	ui.doSearch()

	// Verify we have matches
	if ui.editor.SearchState().Session == nil {
		t.Fatal("expected active search session")
	}

	matches := ui.editor.SearchState().Session.Matches
	if len(matches) != 3 {
		t.Errorf("expected 3 matches for 'test', got %d", len(matches))
	}
}

func TestUnifiedSearchNoLeakage(t *testing.T) {
	ui, _ := newTestUI(80, 24)

	// Start with some original text
	for _, r := range "original text" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}

	// Move to start
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'a', Modifiers: term.ModCtrl})

	// Save original text
	originalText := string(ui.editor.Lines()[0])

	// Enter search mode
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'f', Modifiers: term.ModCtrl})
	if ui.mode != ModeSearch {
		t.Fatal("should be in search mode")
	}

	// Type search query
	for _, r := range "test" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}

	// Press keys that could potentially leak (n, p, a, q, etc.)
	testKeys := []rune{'n', 'p', 'q', 'x', 'y', 'z', 'a', 'b', 'c'}
	for _, r := range testKeys {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}

	// Exit search
	dispatch(ui, term.KeyEvent{Key: term.KeyEscape})

	// Verify text is unchanged
	lines := ui.editor.Lines()
	text := string(lines[0])
	if text != originalText {
		t.Errorf("text modified! Expected '%s', got '%s'", originalText, text)
	}
}

func TestSearchCaseToggle(t *testing.T) {
	ui, _ := newTestUI(80, 24)

	// Insert mixed case text
	for _, r := range "Hello hello HELLO" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'a', Modifiers: term.ModCtrl})

	// Enter search mode
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'f', Modifiers: term.ModCtrl})

	// Default should be case-insensitive
	if ui.editor.SearchState().CaseSensitive {
		t.Error("expected default case-insensitive search")
	}

	// Type search
	for _, r := range "hello" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}

	// Trigger search
	ui.doSearch()

	// Should find all 3 matches (case-insensitive)
	if ui.editor.SearchState().Session != nil && len(ui.editor.SearchState().Session.Matches) != 3 {
		t.Errorf("expected 3 matches (case-insensitive), got %d", len(ui.editor.SearchState().Session.Matches))
	}

	// Toggle case sensitivity with Alt+C
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'c', Modifiers: term.ModAlt})

	// Should now be case-sensitive
	if !ui.editor.SearchState().CaseSensitive {
		t.Error("expected case-sensitive after toggle")
	}

	// Trigger search again
	ui.doSearch()

	// Should find only 1 match (case-sensitive)
	if ui.editor.SearchState().Session != nil && len(ui.editor.SearchState().Session.Matches) != 1 {
		t.Errorf("expected 1 match (case-sensitive), got %d", len(ui.editor.SearchState().Session.Matches))
	}

	// Toggle back
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'c', Modifiers: term.ModAlt})
	if ui.editor.SearchState().CaseSensitive {
		t.Error("expected case-insensitive after second toggle")
	}
}

func TestSearchWholeWordToggle(t *testing.T) {
	ui, _ := newTestUI(80, 24)

	// Insert text with partial matches
	for _, r := range "test testing tester" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'a', Modifiers: term.ModCtrl})

	// Enter search mode
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'f', Modifiers: term.ModCtrl})

	// Default should be whole word off
	if ui.editor.SearchState().WholeWord {
		t.Error("expected default whole word off")
	}

	// Type search
	for _, r := range "test" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}

	// Trigger search
	ui.doSearch()

	// Should find 3 matches (including partial)
	if ui.editor.SearchState().Session != nil && len(ui.editor.SearchState().Session.Matches) != 3 {
		t.Errorf("expected 3 matches (with partials), got %d", len(ui.editor.SearchState().Session.Matches))
	}

	// Toggle whole word with Alt+W
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'w', Modifiers: term.ModAlt})

	// Should now be whole word only
	if !ui.editor.SearchState().WholeWord {
		t.Error("expected whole word after toggle")
	}

	// Trigger search again
	ui.doSearch()

	// Should find 0 matches (no standalone "test" word)
	if ui.editor.SearchState().Session != nil && len(ui.editor.SearchState().Session.Matches) != 0 {
		t.Errorf("expected 0 matches (whole word only), got %d", len(ui.editor.SearchState().Session.Matches))
	}
}

func TestSearchNavigationWhileTyping(t *testing.T) {
	ui, _ := newTestUI(80, 24)

	// Insert text with multiple matches
	for _, r := range "hello world hello again hello" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'a', Modifiers: term.ModCtrl})

	// Enter search mode
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'f', Modifiers: term.ModCtrl})

	// Type search
	for _, r := range "hello" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}
	ui.doSearch()

	// Should have 3 matches
	if ui.editor.SearchState().Session == nil {
		t.Fatal("expected search session")
	}
	if len(ui.editor.SearchState().Session.Matches) != 3 {
		t.Errorf("expected 3 matches, got %d", len(ui.editor.SearchState().Session.Matches))
	}

	// Navigate to next match with 'n'
	initialIndex := ui.editor.SearchState().Session.CurrentIndex
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'n'})
	
	newIndex := ui.editor.SearchState().Session.CurrentIndex
	if newIndex == initialIndex {
		t.Error("expected current index to change after 'n'")
	}

	// Navigate to previous match with 'p'
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'p'})
	
	if ui.editor.SearchState().Session.CurrentIndex != initialIndex {
		t.Error("expected to return to initial index after 'p'")
	}
}

func TestSearchFromSelection(t *testing.T) {
	ui, _ := newTestUI(80, 24)

	// Insert text
	for _, r := range "hello world" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}

	// Select "hello" (first 5 characters)
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'a', Modifiers: term.ModCtrl}) // Move to start
	for i := 0; i < 5; i++ {
		dispatch(ui, term.KeyEvent{Key: term.KeyRight, Modifiers: term.ModShift})
	}

	// Enter search mode with Ctrl+F
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'f', Modifiers: term.ModCtrl})

	// Search query should be pre-filled with selection
	if string(ui.searchQuery) != "hello" {
		t.Errorf("expected search query pre-filled with 'hello', got '%s'", string(ui.searchQuery))
	}
}

func TestSearchEscape(t *testing.T) {
	ui, _ := newTestUI(80, 24)

	// Enter search mode
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'f', Modifiers: term.ModCtrl})
	if ui.mode != ModeSearch {
		t.Fatal("should be in search mode")
	}

	// Type something
	for _, r := range "test" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}

	// Press Escape
	dispatch(ui, term.KeyEvent{Key: term.KeyEscape})

	// Should exit to normal mode
	if ui.mode != ModeNormal {
		t.Errorf("expected ModeNormal after escape, got %v", ui.mode)
	}

	// Search session should be ended
	if ui.editor.SearchState().Session != nil {
		t.Error("expected search session to be nil after escape")
	}

	// But last query should persist
	if ui.lastSearchQuery != "test" {
		t.Errorf("expected last search query to persist, got '%s'", ui.lastSearchQuery)
	}
}

func TestSearchErrorState(t *testing.T) {
	ui, _ := newTestUI(80, 24)

	// Insert text without the search term
	for _, r := range "hello world" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'a', Modifiers: term.ModCtrl})

	// Enter search mode
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'f', Modifiers: term.ModCtrl})

	// Type search for something that doesn't exist
	for _, r := range "xyz" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}
	ui.doSearch()

	// Should still be in search mode
	if ui.mode != ModeSearch {
		t.Errorf("expected to stay in ModeSearch, got %v", ui.mode)
	}

	// Session should exist but have no matches
	if ui.editor.SearchState().Session == nil {
		t.Fatal("expected search session to exist")
	}
	if len(ui.editor.SearchState().Session.Matches) != 0 {
		t.Errorf("expected 0 matches, got %d", len(ui.editor.SearchState().Session.Matches))
	}

	// User can continue typing to correct the search
	dispatch(ui, term.KeyEvent{Key: term.KeyBackspace})
	dispatch(ui, term.KeyEvent{Key: term.KeyBackspace})
	dispatch(ui, term.KeyEvent{Key: term.KeyBackspace})
	for _, r := range "hello" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}
	ui.doSearch()

	// Now should have matches
	if len(ui.editor.SearchState().Session.Matches) == 0 {
		t.Error("expected to find matches after correcting search")
	}
}

func TestSearchEmptyQuery(t *testing.T) {
	ui, _ := newTestUI(80, 24)

	// Insert text
	for _, r := range "hello world" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}

	// Enter search mode
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'f', Modifiers: term.ModCtrl})

	// Search with empty query (just press enter)
	dispatch(ui, term.KeyEvent{Key: term.KeyEnter})

	// Should handle gracefully (either stay in search or show message)
	// According to plan, Enter on non-empty query moves to next match
	// Empty query behavior: stay in search mode
	if ui.mode != ModeSearch {
		t.Error("expected to stay in search mode with empty query")
	}
}

func TestSearchBackspaceOnEmpty(t *testing.T) {
	ui, _ := newTestUI(80, 24)

	// Insert text
	for _, r := range "hello world" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}

	// Enter search mode
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'f', Modifiers: term.ModCtrl})

	if ui.mode != ModeSearch {
		t.Fatal("should be in search mode")
	}

	// Press backspace on empty query
	dispatch(ui, term.KeyEvent{Key: term.KeyBackspace})

	// Should exit search mode (intuitive behavior)
	if ui.mode != ModeNormal {
		t.Errorf("expected to exit search mode, got %v", ui.mode)
	}
}

func TestSearchF1Help(t *testing.T) {
	ui, _ := newTestUI(80, 24)

	// Enter search mode
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'f', Modifiers: term.ModCtrl})

	// Press F1 for help
	dispatch(ui, term.KeyEvent{Key: term.KeyF1})

	// Should show help mode
	if ui.mode != ModeHelp {
		t.Errorf("expected ModeHelp after F1, got %v", ui.mode)
	}

	// Press any key to exit help
	dispatch(ui, term.KeyEvent{Key: term.KeyEscape})

	// Should return to normal mode (not search mode, as help exits to normal)
	if ui.mode != ModeNormal {
		t.Errorf("expected ModeNormal after exiting help, got %v", ui.mode)
	}
}

func TestSearchF3Navigation(t *testing.T) {
	ui, _ := newTestUI(80, 24)

	// Insert text with multiple matches
	for _, r := range "test one test two test three" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'a', Modifiers: term.ModCtrl})

	// Enter search mode
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'f', Modifiers: term.ModCtrl})

	// Type search
	for _, r := range "test" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}
	ui.doSearch()

	if ui.editor.SearchState().Session == nil {
		t.Fatal("expected search session")
	}

	initialIndex := ui.editor.SearchState().Session.CurrentIndex

	// Press F3 to go to next match
	dispatch(ui, term.KeyEvent{Key: term.KeyF3})
	
	if ui.editor.SearchState().Session.CurrentIndex <= initialIndex {
		t.Error("expected F3 to move to next match")
	}

	// Press Shift+F3 to go to previous match
	dispatch(ui, term.KeyEvent{Key: term.KeyF3, Modifiers: term.ModShift})
	
	if ui.editor.SearchState().Session.CurrentIndex != initialIndex {
		t.Error("expected Shift+F3 to return to previous match")
	}
}

func TestSearchEnterMovesToNext(t *testing.T) {
	ui, _ := newTestUI(80, 24)

	// Insert text with multiple matches
	for _, r := range "find me find me find me" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'a', Modifiers: term.ModCtrl})

	// Enter search mode
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'f', Modifiers: term.ModCtrl})

	// Type search
	for _, r := range "find" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}
	ui.doSearch()

	if ui.editor.SearchState().Session == nil {
		t.Fatal("expected search session")
	}

	initialIndex := ui.editor.SearchState().Session.CurrentIndex

	// Press Enter to move to next match
	dispatch(ui, term.KeyEvent{Key: term.KeyEnter})

	// Should move to next match (or wrap around)
	newIndex := ui.editor.SearchState().Session.CurrentIndex
	if newIndex == initialIndex && len(ui.editor.SearchState().Session.Matches) > 1 {
		t.Error("expected Enter to move to next match")
	}

	// Should stay in search mode
	if ui.mode != ModeSearch {
		t.Errorf("expected to stay in ModeSearch, got %v", ui.mode)
	}
}


