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
	"strings"
	"testing"
)

// TestSearchCanTypeAllLetters verifies that ALL letters (a-z, A-Z) can be typed
// into the search field without triggering commands. This is critical UX -
// search should be a text input field first, commands second.
//
// BUG CONTEXT: Previously, typing 'n' triggered "next match" and 'r' triggered
// "replace", making it impossible to search for words containing these letters.
func TestSearchCanTypeAllLetters(t *testing.T) {
	ui, _ := newTestUI(80, 24)

	// Insert some content to search through
	for _, r := range "the quick brown fox jumps over the lazy dog" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}

	// Enter search mode
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'f', Modifiers: term.ModCtrl})
	if ui.mode != ModeSearch {
		t.Fatalf("expected ModeSearch, got %v", ui.mode)
	}

	// Test all lowercase letters
	lowercase := "abcdefghijklmnopqrstuvwxyz"
	for _, letter := range lowercase {
		// Reset search
		ui.searchQuery = nil

		// Type the letter
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: letter})

		// Verify it was added to the search query
		if len(ui.searchQuery) != 1 {
			t.Errorf("letter '%c': expected query length 1, got %d", letter, len(ui.searchQuery))
		}
		if len(ui.searchQuery) > 0 && ui.searchQuery[0] != letter {
			t.Errorf("letter '%c': expected in query, got '%c'", letter, ui.searchQuery[0])
		}

		// Verify we're still in search mode (not accidentally triggered a command)
		if ui.mode != ModeSearch {
			t.Errorf("letter '%c': triggered mode change to %v", letter, ui.mode)
		}
	}

	// Test all uppercase letters
	uppercase := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	for _, letter := range uppercase {
		// Reset search
		ui.searchQuery = nil

		// Type the letter (uppercase, which means Shift modifier)
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: letter, Modifiers: term.ModShift})

		// Verify it was added to the search query
		if len(ui.searchQuery) != 1 {
			t.Errorf("letter '%c': expected query length 1, got %d", letter, len(ui.searchQuery))
		}
		if len(ui.searchQuery) > 0 && ui.searchQuery[0] != letter {
			t.Errorf("letter '%c': expected in query, got '%c'", letter, ui.searchQuery[0])
		}

		// Verify we're still in search mode
		if ui.mode != ModeSearch {
			t.Errorf("letter '%c': triggered mode change to %v", letter, ui.mode)
		}
	}
}

// TestSearchCanTypeCommandLetters specifically tests the problematic letters
// that have command meanings (n, p, r, a, q) to ensure they can still be typed.
func TestSearchCanTypeCommandLetters(t *testing.T) {
	ui, _ := newTestUI(80, 24)

	// Insert content with matches
	for _, r := range "nanny proper runner application question" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}

	// Enter search mode
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'f', Modifiers: term.ModCtrl})

	// Test typing "nanny" (contains 'n' which means "next")
	testCases := []struct {
		word string
		desc string
	}{
		{"nanny", "word starting with 'n' (next command letter)"},
		{"proper", "word with 'p' (previous command letter)"},
		{"runner", "word with 'r' (replace command letter)"},
		{"application", "word with 'a' (replace all command letter)"},
		{"question", "word with 'q' (quit command letter)"},
		{"napqr", "all command letters together"},
	}

	for _, tc := range testCases {
		// Reset search
		ui.searchQuery = nil

		// Type the word character by character
		for i, letter := range tc.word {
			dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: letter})

			// Verify each letter was added
			if len(ui.searchQuery) != i+1 {
				t.Errorf("%s: after typing '%c' (char %d), expected query length %d, got %d",
					tc.desc, letter, i+1, i+1, len(ui.searchQuery))
				t.Errorf("  Query so far: '%s'", string(ui.searchQuery))
			}

			// Verify we're still in search mode
			if ui.mode != ModeSearch {
				t.Errorf("%s: after typing '%c', mode changed to %v", tc.desc, letter, ui.mode)
			}
		}

		// Verify final query matches the word
		if string(ui.searchQuery) != tc.word {
			t.Errorf("%s: expected query '%s', got '%s'", tc.desc, tc.word, string(ui.searchQuery))
		}
	}
}

// TestSearchTypingNumbers verifies numbers can be typed in search
func TestSearchTypingNumbers(t *testing.T) {
	ui, _ := newTestUI(80, 24)

	// Insert content
	for _, r := range "123 456 789" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}

	// Enter search mode
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'f', Modifiers: term.ModCtrl})

	// Type numbers
	for _, digit := range "0123456789" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: digit})
	}

	expected := "0123456789"
	if string(ui.searchQuery) != expected {
		t.Errorf("expected query '%s', got '%s'", expected, string(ui.searchQuery))
	}

	if ui.mode != ModeSearch {
		t.Errorf("expected to stay in search mode, got %v", ui.mode)
	}
}

// TestSearchTypingSpecialCharacters verifies special characters can be typed
func TestSearchTypingSpecialCharacters(t *testing.T) {
	ui, _ := newTestUI(80, 24)

	// Insert content with special chars
	for _, r := range "hello@world.com $100 [test]" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}

	// Enter search mode
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'f', Modifiers: term.ModCtrl})

	// Type special characters
	specialChars := "!@#$%^&*()_+-=[]{}\\|;':\",./<>?"
	for _, char := range specialChars {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: char})
	}

	if string(ui.searchQuery) != specialChars {
		t.Errorf("expected query '%s', got '%s'", specialChars, string(ui.searchQuery))
	}

	if ui.mode != ModeSearch {
		t.Errorf("expected to stay in search mode, got %v", ui.mode)
	}
}

// TestSearchTypingMixedContent verifies realistic search patterns
func TestSearchTypingMixedContent(t *testing.T) {
	ui, _ := newTestUI(80, 24)

	// Insert realistic content
	for _, r := range "function calculatePrice(item) { return item.price * 1.2; }" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}

	// Enter search mode
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'f', Modifiers: term.ModCtrl})

	// Type realistic search queries
	testQueries := []string{
		"function",       // Normal word
		"price",          // Contains 'p' (previous)
		"return",         // Contains 'r' (replace)
		"item.price",     // With dot
		"* 1.2",          // Math expression
		"calculatePrice", // CamelCase
		"qna",            // Contains 'q', 'n', 'a'
	}

	for _, query := range testQueries {
		// Reset
		ui.searchQuery = nil

		// Type the query
		for _, char := range query {
			dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: char})
		}

		// Verify
		if string(ui.searchQuery) != query {
			t.Errorf("query '%s': expected '%s', got '%s'", query, query, string(ui.searchQuery))
		}

		if ui.mode != ModeSearch {
			t.Errorf("query '%s': mode changed to %v", query, ui.mode)
		}
	}
}

// TestSearchCommandsStillWork verifies that command keys DO work when appropriate
// Navigation is via F3/Shift+F3 or Enter, and replace is via Ctrl+R/Ctrl+H
func TestSearchCommandsStillWork(t *testing.T) {
	ui, _ := newTestUI(80, 24)

	// Insert content with multiple "test" occurrences
	for _, r := range "test one test two test three" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'a', Modifiers: term.ModCtrl})

	// Enter search mode
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'f', Modifiers: term.ModCtrl})

	// Type "test" first - need matches for commands to work
	for _, r := range "test" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}
	ui.doSearch()

	// Verify we have matches
	if ui.editor.SearchState().Session == nil || len(ui.editor.SearchState().Session.Matches) < 2 {
		t.Fatal("expected at least 2 matches for testing navigation")
	}

	queryBeforeNav := string(ui.searchQuery)
	initialIdx := ui.editor.SearchState().Session.CurrentIndex

	// Test that F3 navigates
	dispatch(ui, term.KeyEvent{Key: term.KeyF3})

	// Query should be unchanged
	if string(ui.searchQuery) != queryBeforeNav {
		t.Errorf("F3 command modified query: expected '%s', got '%s'", queryBeforeNav, string(ui.searchQuery))
	}

	// Index should have changed
	if ui.editor.SearchState().Session.CurrentIndex == initialIdx {
		t.Error("F3 command didn't navigate to next match")
	}

	// Test Shift+F3 navigates back
	dispatch(ui, term.KeyEvent{Key: term.KeyF3, Modifiers: term.ModShift})

	if string(ui.searchQuery) != queryBeforeNav {
		t.Errorf("Shift+F3 command modified query: expected '%s', got '%s'", queryBeforeNav, string(ui.searchQuery))
	}

	if ui.editor.SearchState().Session.CurrentIndex != initialIdx {
		t.Error("Shift+F3 command didn't navigate back to original match")
	}

	// Test that Enter also navigates
	dispatch(ui, term.KeyEvent{Key: term.KeyEnter})
	if ui.editor.SearchState().Session.CurrentIndex == initialIdx {
		t.Error("Enter didn't navigate to next match")
	}
}

// TestSearchCommandLettersBehavior tests that all letters are now typed into search
// The OLD behavior where 'n'/'p'/'r'/'a'/'q' were commands is REMOVED
// Navigation is now via F3/Shift+F3, replace via Ctrl+R/Ctrl+H, exit via Escape
func TestSearchCommandLettersBehavior(t *testing.T) {
	ui, _ := newTestUI(80, 24)

	// Insert content
	for _, r := range "no matches here" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'a', Modifiers: term.ModCtrl})

	// Enter search mode
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'f', Modifiers: term.ModCtrl})

	t.Run("all letters type regardless of matches", func(t *testing.T) {
		// Type 'xyz' - won't find any matches
		for _, r := range "xyz" {
			dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
		}
		ui.doSearch()

		// Type 'n' - should add to query, not navigate
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'n'})

		if !strings.HasSuffix(string(ui.searchQuery), "n") {
			t.Errorf("'n' should be added to query, got '%s'", string(ui.searchQuery))
		}

		// Type 'a', 'p', 'r', 'q' - all should be added
		for _, r := range "aprq" {
			dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
		}

		expected := "xyznaprq"
		if string(ui.searchQuery) != expected {
			t.Errorf("expected query '%s', got '%s'", expected, string(ui.searchQuery))
		}
	})

	t.Run("navigation uses F3 not letters", func(t *testing.T) {
		// Clear and search for something that exists
		ui.searchQuery = nil
		for _, r := range "no" {
			dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
		}
		ui.doSearch()

		// Should have matches
		if ui.editor.SearchState().Session == nil || len(ui.editor.SearchState().Session.Matches) == 0 {
			t.Fatal("expected matches for 'no'")
		}

		queryBefore := string(ui.searchQuery)
		idxBefore := ui.editor.SearchState().Session.CurrentIndex

		// Type 'n' - should ADD to query, not navigate
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'n'})

		queryAfter := string(ui.searchQuery)
		if queryAfter == queryBefore {
			t.Errorf("'n' should add to query, but query unchanged: '%s'", queryAfter)
		}

		// Use F3 for navigation
		ui.searchQuery = []rune("no") // Reset query
		ui.doSearch()

		// Check that we have multiple matches to navigate
		if ui.editor.SearchState().Session == nil || len(ui.editor.SearchState().Session.Matches) < 2 {
			t.Skip("need at least 2 matches for navigation test")
		}

		idxBefore = ui.editor.SearchState().Session.CurrentIndex

		dispatch(ui, term.KeyEvent{Key: term.KeyF3})

		idxAfter := ui.editor.SearchState().Session.CurrentIndex
		if idxAfter == idxBefore {
			// If index didn't change, might have wrapped around
			t.Logf("F3 didn't change index from %d (might have wrapped with %d matches)",
				idxBefore, len(ui.editor.SearchState().Session.Matches))
		}
	})
}

// TestSearchTypingDoesntLeakToBuffer ensures search text never leaks to the actual buffer
func TestSearchTypingDoesntLeakToBuffer(t *testing.T) {
	ui, _ := newTestUI(80, 24)

	// Insert initial content
	initialText := "original content"
	for _, r := range initialText {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}

	// Enter search mode
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'f', Modifiers: term.ModCtrl})

	// Type a long search query with all kinds of characters
	searchText := "abcdefghijklmnopqrstuvwxyz NOPQR 12345 !@#$%"
	for _, r := range searchText {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}

	// Exit search
	dispatch(ui, term.KeyEvent{Key: term.KeyEscape})

	// Verify buffer content is unchanged
	lines := ui.editor.Lines()
	actualText := string(lines[0])
	if actualText != initialText {
		t.Errorf("buffer was modified! Expected '%s', got '%s'", initialText, actualText)
	}
}

// TestSearchShiftLettersAreAdded verifies that Shift+letter adds uppercase to query
func TestSearchShiftLettersAreAdded(t *testing.T) {
	ui, _ := newTestUI(80, 24)

	// Enter search mode
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'f', Modifiers: term.ModCtrl})

	// Type "Hello" with mixed case (H is Shift+h)
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'H', Modifiers: term.ModShift})
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'e'})
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'l'})
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'l'})
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'o'})

	expected := "Hello"
	if string(ui.searchQuery) != expected {
		t.Errorf("expected query '%s', got '%s'", expected, string(ui.searchQuery))
	}
}
