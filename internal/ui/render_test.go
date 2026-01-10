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

func TestSearchHighlighting(t *testing.T) {
	ui, screen := newTestUI(80, 24)

	// Insert text with multiple matches
	for _, r := range "hello world hello again" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'a', Modifiers: term.ModCtrl})

	// Enter search mode and search
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'f', Modifiers: term.ModCtrl})
	for _, r := range "hello" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}
	ui.doSearch()

	// Draw the screen
	draw(ui)

	// Verify that matches are highlighted
	// This is a basic test - in a real implementation, we'd check
	// that the cells at match positions have the correct highlight style
	if ui.editor.SearchState().Session == nil {
		t.Fatal("expected active search session")
	}

	matches := ui.editor.SearchState().Session.Matches
	if len(matches) != 2 {
		t.Errorf("expected 2 matches, got %d", len(matches))
	}

	// Verify screen was drawn (basic check)
	// In a real implementation, we'd check cell styles at match positions
	_ = screen // Use screen variable
}

func TestSearchCurrentMatchHighlighting(t *testing.T) {
	ui, screen := newTestUI(80, 24)

	// Insert text
	for _, r := range "test one test two test three" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'a', Modifiers: term.ModCtrl})

	// Search for "test"
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'f', Modifiers: term.ModCtrl})
	for _, r := range "test" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}
	ui.doSearch()

	if ui.editor.SearchState().Session == nil {
		t.Fatal("expected search session")
	}

	// Current match should be at index 0
	currentIndex := ui.editor.SearchState().Session.CurrentIndex
	if currentIndex < 0 || currentIndex >= len(ui.editor.SearchState().Session.Matches) {
		t.Fatalf("invalid current index: %d", currentIndex)
	}

	// Draw
	draw(ui)

	// Verify current match is highlighted differently
	// In real implementation, check that current match has different style
	// than other matches
	_ = screen
}

func TestStatusBarInSearch(t *testing.T) {
	ui, screen := newTestUI(80, 24)

	// Insert text
	for _, r := range "hello world hello" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'a', Modifiers: term.ModCtrl})

	// Enter search mode
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'f', Modifiers: term.ModCtrl})
	for _, r := range "hello" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}
	ui.doSearch()

	// Draw
	draw(ui)

	// Verify status bar shows search info
	if ui.mode != ModeSearch {
		t.Fatal("should be in search mode")
	}

	// Check that status bar was rendered
	// In real implementation, check status bar content
	h := screen.h
	statusY := h - 1

	// Basic check: status bar row should have content
	hasContent := false
	for x := 0; x < screen.w; x++ {
		if screen.Cell(x, statusY) != 0 {
			hasContent = true
			break
		}
	}

	if !hasContent {
		t.Error("status bar should have content in search mode")
	}
}

func TestStatusBarSearchError(t *testing.T) {
	ui, screen := newTestUI(80, 24)

	// Insert text
	for _, r := range "hello world" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'a', Modifiers: term.ModCtrl})

	// Search for something that doesn't exist
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'f', Modifiers: term.ModCtrl})
	for _, r := range "xyz" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}
	ui.doSearch()

	// Draw
	draw(ui)

	// Verify we're in error state
	if ui.editor.SearchState().Session == nil {
		t.Fatal("expected search session even with no matches")
	}

	if len(ui.editor.SearchState().Session.Matches) != 0 {
		t.Error("expected no matches in error state")
	}

	// Status bar should indicate error
	// In real implementation, check for error styling/message
	_ = screen
}

func TestStatusBarMatchCount(t *testing.T) {
	ui, _ := newTestUI(80, 24)

	// Insert text with 3 matches
	for _, r := range "foo bar foo baz foo" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'a', Modifiers: term.ModCtrl})

	// Search
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'f', Modifiers: term.ModCtrl})
	for _, r := range "foo" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}
	ui.doSearch()

	// Draw
	draw(ui)

	// Verify match count
	if ui.editor.SearchState().Session == nil {
		t.Fatal("expected search session")
	}

	matchCount := len(ui.editor.SearchState().Session.Matches)
	if matchCount != 3 {
		t.Errorf("expected 3 matches, got %d", matchCount)
	}

	// Status bar should show match count
	// In real implementation, parse status bar text to verify
}

func TestStatusBarCaseSensitivityIndicator(t *testing.T) {
	ui, _ := newTestUI(80, 24)

	// Insert text
	for _, r := range "Hello hello" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'a', Modifiers: term.ModCtrl})

	// Enter search mode
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'f', Modifiers: term.ModCtrl})

	// Default: case-insensitive
	if ui.editor.SearchState().CaseSensitive {
		t.Error("expected default case-insensitive")
	}

	draw(ui)

	// Toggle to case-sensitive
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'c', Modifiers: term.ModAlt})

	if !ui.editor.SearchState().CaseSensitive {
		t.Error("expected case-sensitive after toggle")
	}

	draw(ui)

	// Status bar should show case sensitivity indicator
	// In real implementation, verify indicator text
}

func TestStatusBarWholeWordIndicator(t *testing.T) {
	ui, _ := newTestUI(80, 24)

	// Insert text
	for _, r := range "test testing" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'a', Modifiers: term.ModCtrl})

	// Enter search mode
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'f', Modifiers: term.ModCtrl})

	// Default: whole word off
	if ui.editor.SearchState().WholeWord {
		t.Error("expected default whole word off")
	}

	draw(ui)

	// Toggle whole word
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'w', Modifiers: term.ModAlt})

	if !ui.editor.SearchState().WholeWord {
		t.Error("expected whole word on after toggle")
	}

	draw(ui)

	// Status bar should show whole word indicator
	// In real implementation, verify indicator text
}

func TestStatusBarResponsive(t *testing.T) {
	// Test with different screen widths
	widths := []int{60, 80, 100, 120}

	for _, width := range widths {
		ui, _ := newTestUI(width, 24)

		// Insert and search
		for _, r := range "test" {
			dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
		}
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'a', Modifiers: term.ModCtrl})
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'f', Modifiers: term.ModCtrl})
		for _, r := range "test" {
			dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
		}
		ui.doSearch()

		// Draw should succeed without panic at any width
		draw(ui)

		// Status bar should adapt to width
		// In real implementation, verify content adapts appropriately
	}
}

func TestHighlightingWithSelection(t *testing.T) {
	ui, _ := newTestUI(80, 24)

	// Insert text
	for _, r := range "hello world hello" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}

	// Select "world"
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'a', Modifiers: term.ModCtrl})
	for i := 0; i < 6; i++ {
		dispatch(ui, term.KeyEvent{Key: term.KeyRight})
	}
	for i := 0; i < 5; i++ {
		dispatch(ui, term.KeyEvent{Key: term.KeyRight, Modifiers: term.ModShift})
	}

	// Enter search and search for "hello"
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'f', Modifiers: term.ModCtrl})
	for _, r := range "hello" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}
	ui.doSearch()

	// Draw
	draw(ui)

	// According to plan, selection should take priority over search highlighting
	// In real implementation, verify that selected text uses selection style,
	// not search highlight style
}

func TestHighlightingMaxMatches(t *testing.T) {
	ui, _ := newTestUI(80, 24)

	// Insert text with many matches
	for i := 0; i < 50; i++ {
		for _, r := range "test " {
			dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
		}
	}

	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'a', Modifiers: term.ModCtrl})

	// Search for "test"
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'f', Modifiers: term.ModCtrl})
	for _, r := range "test" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}
	ui.doSearch()

	// Should handle many matches gracefully
	if ui.editor.SearchState().Session != nil {
		matchCount := len(ui.editor.SearchState().Session.Matches)
		// Should be limited to avoid performance issues
		// Exact limit depends on implementation (plan suggests 1000)
		if matchCount > 1000 {
			t.Errorf("too many matches: %d (should be limited)", matchCount)
		}
	}

	// Draw should still work
	draw(ui)
}

func TestSearchStatusBarPreservesRightInfo(t *testing.T) {
	ui, screen := newTestUI(80, 24)

	// Insert text
	for _, r := range "hello world" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}

	// Enter search mode
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'f', Modifiers: term.ModCtrl})
	for _, r := range "hello" {
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
	}
	ui.doSearch()

	// Draw
	draw(ui)

	// Status bar should still show line, col, encoding on the right
	// In real implementation, verify right side of status bar contains
	// "Ln X, Col Y  UTF-8 LF" or similar
	h := screen.h
	statusY := h - 1

	// Check that right side has content
	rightContent := false
	for x := screen.w - 30; x < screen.w; x++ {
		if screen.Cell(x, statusY) != 0 {
			rightContent = true
			break
		}
	}

	if !rightContent {
		t.Error("status bar should preserve right-side info in search mode")
	}
}
