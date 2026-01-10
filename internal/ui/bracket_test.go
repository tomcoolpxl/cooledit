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

	"cooledit/internal/config"
	"cooledit/internal/core"
)

// moveCursorTo moves the cursor to the specified position using editor commands
func moveCursorTo(editor *core.Editor, targetLine, targetCol int) {
	// First go to start of file
	editor.Apply(core.CmdFileStart{}, 100)

	// Move down to target line
	for i := 0; i < targetLine; i++ {
		editor.Apply(core.CmdMoveDown{}, 100)
	}

	// Move to start of line
	editor.Apply(core.CmdMoveHome{}, 100)

	// Move right to target column
	for i := 0; i < targetCol; i++ {
		editor.Apply(core.CmdMoveRight{}, 100)
	}
}

// TestBracketMatchingOnBracket tests that bracket match state is created when cursor is on a bracket
func TestBracketMatchingOnBracket(t *testing.T) {
	screen := NewFakeScreen(80, 24)
	editor := newTestEditorWithContent("(a)", "test.txt")

	cfg := config.Default()
	ui := New(screen, editor, cfg)

	// Move cursor to the opening bracket at column 0
	moveCursorTo(editor, 0, 0)
	ui.updateBracketMatch()

	if ui.bracketMatchState == nil {
		t.Fatal("Expected bracket match state to be created when cursor is on '('")
	}

	if !ui.bracketMatchState.IsOnBracket {
		t.Error("Expected IsOnBracket to be true")
	}

	if !ui.bracketMatchState.HasMatch {
		t.Error("Expected HasMatch to be true for '(' in '(a)'")
	}

	if ui.bracketMatchState.MatchLine != 0 || ui.bracketMatchState.MatchCol != 2 {
		t.Errorf("Expected match at (0, 2), got (%d, %d)",
			ui.bracketMatchState.MatchLine, ui.bracketMatchState.MatchCol)
	}
}

// TestBracketMatchingNotOnBracket tests that no state is created when cursor is not on a bracket
func TestBracketMatchingNotOnBracket(t *testing.T) {
	screen := NewFakeScreen(80, 24)
	editor := newTestEditorWithContent("(a)", "test.txt")

	cfg := config.Default()
	ui := New(screen, editor, cfg)

	// Move cursor to 'a' at column 1
	moveCursorTo(editor, 0, 1)
	ui.updateBracketMatch()

	if ui.bracketMatchState != nil {
		t.Error("Expected no bracket match state when cursor is not on a bracket")
	}
}

// TestBracketMatchingUnmatched tests that unmatched brackets are detected
func TestBracketMatchingUnmatched(t *testing.T) {
	screen := NewFakeScreen(80, 24)
	editor := newTestEditorWithContent("(a", "test.txt") // Unmatched bracket

	cfg := config.Default()
	ui := New(screen, editor, cfg)

	// Move cursor to the opening bracket
	moveCursorTo(editor, 0, 0)
	ui.updateBracketMatch()

	if ui.bracketMatchState == nil {
		t.Fatal("Expected bracket match state for unmatched bracket")
	}

	if !ui.bracketMatchState.IsOnBracket {
		t.Error("Expected IsOnBracket to be true")
	}

	if ui.bracketMatchState.HasMatch {
		t.Error("Expected HasMatch to be false for unmatched '('")
	}
}

// TestBracketMatchingClosingBracket tests matching from closing bracket back to opening
func TestBracketMatchingClosingBracket(t *testing.T) {
	screen := NewFakeScreen(80, 24)
	editor := newTestEditorWithContent("(a)", "test.txt")

	cfg := config.Default()
	ui := New(screen, editor, cfg)

	// Move cursor to the closing bracket at column 2
	moveCursorTo(editor, 0, 2)
	ui.updateBracketMatch()

	if ui.bracketMatchState == nil {
		t.Fatal("Expected bracket match state when cursor is on ')'")
	}

	if !ui.bracketMatchState.HasMatch {
		t.Error("Expected HasMatch to be true for ')' in '(a)'")
	}

	if ui.bracketMatchState.MatchLine != 0 || ui.bracketMatchState.MatchCol != 0 {
		t.Errorf("Expected match at (0, 0), got (%d, %d)",
			ui.bracketMatchState.MatchLine, ui.bracketMatchState.MatchCol)
	}
}

// TestBracketMatchingNested tests nested bracket matching
func TestBracketMatchingNested(t *testing.T) {
	screen := NewFakeScreen(80, 24)
	editor := newTestEditorWithContent("((a))", "test.txt")

	cfg := config.Default()
	ui := New(screen, editor, cfg)

	// Test outer opening bracket at column 0
	moveCursorTo(editor, 0, 0)
	ui.updateBracketMatch()

	if ui.bracketMatchState == nil {
		t.Fatal("Expected bracket match state")
	}

	if ui.bracketMatchState.MatchCol != 4 {
		t.Errorf("Outer '(' at 0 should match ')' at 4, got match at %d",
			ui.bracketMatchState.MatchCol)
	}

	// Test inner opening bracket at column 1
	moveCursorTo(editor, 0, 1)
	ui.updateBracketMatch()

	if ui.bracketMatchState == nil {
		t.Fatal("Expected bracket match state")
	}

	if ui.bracketMatchState.MatchCol != 3 {
		t.Errorf("Inner '(' at 1 should match ')' at 3, got match at %d",
			ui.bracketMatchState.MatchCol)
	}
}

// TestBracketMatchingMultiLine tests multi-line bracket matching
func TestBracketMatchingMultiLine(t *testing.T) {
	screen := NewFakeScreen(80, 24)
	editor := newTestEditorWithContent("func main() {\n\tprintln()\n}", "test.go")

	cfg := config.Default()
	ui := New(screen, editor, cfg)

	// Find the '{' at line 0, column 12
	moveCursorTo(editor, 0, 12)
	ui.updateBracketMatch()

	if ui.bracketMatchState == nil {
		t.Fatal("Expected bracket match state")
	}

	if !ui.bracketMatchState.HasMatch {
		t.Error("Expected HasMatch to be true")
	}

	if ui.bracketMatchState.MatchLine != 2 || ui.bracketMatchState.MatchCol != 0 {
		t.Errorf("'{' at (0, 12) should match '}' at (2, 0), got (%d, %d)",
			ui.bracketMatchState.MatchLine, ui.bracketMatchState.MatchCol)
	}
}

// TestBracketMatchingInString tests that brackets in strings are ignored
func TestBracketMatchingInString(t *testing.T) {
	screen := NewFakeScreen(80, 24)
	// Code: x = "(" (bracket inside string should be ignored)
	editor := newTestEditorWithContent("x = \"(\"", "test.py")

	cfg := config.Default()
	cfg.Editor.SyntaxHighlighting = true
	ui := New(screen, editor, cfg)

	// Move cursor to the '(' inside the string at column 5
	moveCursorTo(editor, 0, 5)
	ui.updateBracketMatch()

	// Bracket inside string should be ignored - no match state
	if ui.bracketMatchState != nil {
		t.Error("Expected no bracket match state for bracket inside string")
	}
}

// TestBracketMatchingInComment tests that brackets in comments are ignored
func TestBracketMatchingInComment(t *testing.T) {
	screen := NewFakeScreen(80, 24)
	// Code: // comment ( bracket in comment
	editor := newTestEditorWithContent("// comment (", "test.go")

	cfg := config.Default()
	cfg.Editor.SyntaxHighlighting = true
	ui := New(screen, editor, cfg)

	// Move cursor to the '(' in comment at column 11
	moveCursorTo(editor, 0, 11)
	ui.updateBracketMatch()

	// Bracket inside comment should be ignored
	if ui.bracketMatchState != nil {
		t.Error("Expected no bracket match state for bracket inside comment")
	}
}

// TestGetBracketStyleMatched tests getBracketStyle returns correct style for matched brackets
func TestGetBracketStyleMatched(t *testing.T) {
	screen := NewFakeScreen(80, 24)
	editor := newTestEditorWithContent("(a)", "test.txt")

	cfg := config.Default()
	cfg.UI.Theme = "dark" // Use dark theme for consistent colors
	ui := New(screen, editor, cfg)

	// Move cursor to opening bracket
	moveCursorTo(editor, 0, 0)
	ui.updateBracketMatch()

	// Check style for cursor bracket position
	style := ui.getBracketStyle(0, 0)
	if style == nil {
		t.Fatal("Expected bracket style for cursor position")
	}

	// Check style for match position
	matchStyle := ui.getBracketStyle(0, 2)
	if matchStyle == nil {
		t.Fatal("Expected bracket style for match position")
	}

	// Check non-bracket position returns nil
	otherStyle := ui.getBracketStyle(0, 1)
	if otherStyle != nil {
		t.Error("Expected nil style for non-bracket position")
	}
}

// TestGetBracketStyleUnmatched tests getBracketStyle returns unmatched style
func TestGetBracketStyleUnmatched(t *testing.T) {
	screen := NewFakeScreen(80, 24)
	editor := newTestEditorWithContent("(", "test.txt") // Unmatched

	cfg := config.Default()
	cfg.UI.Theme = "dark"
	ui := New(screen, editor, cfg)

	// Move cursor to unmatched bracket
	moveCursorTo(editor, 0, 0)
	ui.updateBracketMatch()

	style := ui.getBracketStyle(0, 0)
	if style == nil {
		t.Fatal("Expected bracket style for unmatched bracket")
	}

	// Style should use unmatch background color
	if style.Background != ui.theme.Editor.BracketUnmatchBg {
		t.Errorf("Expected unmatch background color %v, got %v",
			ui.theme.Editor.BracketUnmatchBg, style.Background)
	}
}

// TestBracketMatchingAllTypes tests all bracket types work
func TestBracketMatchingAllTypes(t *testing.T) {
	tests := []struct {
		content string
		col     int
	}{
		{"(a)", 0}, // Parentheses
		{"[a]", 0}, // Square brackets
		{"{a}", 0}, // Curly braces
		{"<a>", 0}, // Angle brackets
	}

	for _, tc := range tests {
		screen := NewFakeScreen(80, 24)
		editor := newTestEditorWithContent(tc.content, "test.txt")

		cfg := config.Default()
		ui := New(screen, editor, cfg)

		moveCursorTo(editor, 0, tc.col)
		ui.updateBracketMatch()

		if ui.bracketMatchState == nil {
			t.Errorf("Content %q: Expected bracket match state", tc.content)
			continue
		}

		if !ui.bracketMatchState.HasMatch {
			t.Errorf("Content %q: Expected HasMatch to be true", tc.content)
		}

		if ui.bracketMatchState.MatchCol != 2 {
			t.Errorf("Content %q: Expected match at column 2, got %d",
				tc.content, ui.bracketMatchState.MatchCol)
		}
	}
}

// TestBracketMatchingMixed tests mixed bracket types
func TestBracketMatchingMixed(t *testing.T) {
	screen := NewFakeScreen(80, 24)
	editor := newTestEditorWithContent("{[()]}", "test.txt")

	cfg := config.Default()
	ui := New(screen, editor, cfg)

	// Test '{' at 0 matches '}' at 5
	moveCursorTo(editor, 0, 0)
	ui.updateBracketMatch()
	if ui.bracketMatchState == nil || ui.bracketMatchState.MatchCol != 5 {
		t.Errorf("'{' at 0 should match '}' at 5")
	}

	// Test '[' at 1 matches ']' at 4
	moveCursorTo(editor, 0, 1)
	ui.updateBracketMatch()
	if ui.bracketMatchState == nil || ui.bracketMatchState.MatchCol != 4 {
		t.Errorf("'[' at 1 should match ']' at 4")
	}

	// Test '(' at 2 matches ')' at 3
	moveCursorTo(editor, 0, 2)
	ui.updateBracketMatch()
	if ui.bracketMatchState == nil || ui.bracketMatchState.MatchCol != 3 {
		t.Errorf("'(' at 2 should match ')' at 3")
	}
}

// TestBracketMatchInitialized tests that bracketMatcher is initialized
func TestBracketMatchInitialized(t *testing.T) {
	screen := NewFakeScreen(80, 24)
	editor := newTestEditorWithContent("test", "test.txt")

	cfg := config.Default()
	ui := New(screen, editor, cfg)

	if ui.bracketMatcher == nil {
		t.Error("Expected bracketMatcher to be initialized")
	}
}

// TestBracketStyleUsesCorrectMatchColor tests that matched brackets use match color
func TestBracketStyleUsesCorrectMatchColor(t *testing.T) {
	screen := NewFakeScreen(80, 24)
	editor := newTestEditorWithContent("(a)", "test.txt")

	cfg := config.Default()
	cfg.UI.Theme = "dark"
	ui := New(screen, editor, cfg)

	moveCursorTo(editor, 0, 0)
	ui.updateBracketMatch()

	style := ui.getBracketStyle(0, 0)
	if style == nil {
		t.Fatal("Expected bracket style")
	}

	// Should use match background color since bracket is matched
	if style.Background != ui.theme.Editor.BracketMatchBg {
		t.Errorf("Expected match background color %v, got %v",
			ui.theme.Editor.BracketMatchBg, style.Background)
	}
}
