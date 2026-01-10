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

package core

import (
	"strings"
	"testing"
)

func TestBracketMatcherIsBracket(t *testing.T) {
	m := NewBracketMatcher()

	brackets := []rune{'(', ')', '[', ']', '{', '}', '<', '>'}
	for _, b := range brackets {
		if !m.IsBracket(b) {
			t.Errorf("Expected %q to be a bracket", b)
		}
	}

	nonBrackets := []rune{'a', '1', ' ', '+', '-', '*', '/', '"', '\''}
	for _, b := range nonBrackets {
		if m.IsBracket(b) {
			t.Errorf("Expected %q to NOT be a bracket", b)
		}
	}
}

func TestBracketMatcherIsOpenBracket(t *testing.T) {
	m := NewBracketMatcher()

	openBrackets := []rune{'(', '[', '{', '<'}
	for _, b := range openBrackets {
		if !m.IsOpenBracket(b) {
			t.Errorf("Expected %q to be an opening bracket", b)
		}
	}

	closeBrackets := []rune{')', ']', '}', '>'}
	for _, b := range closeBrackets {
		if m.IsOpenBracket(b) {
			t.Errorf("Expected %q to NOT be an opening bracket", b)
		}
	}
}

func TestBracketMatcherIsCloseBracket(t *testing.T) {
	m := NewBracketMatcher()

	closeBrackets := []rune{')', ']', '}', '>'}
	for _, b := range closeBrackets {
		if !m.IsCloseBracket(b) {
			t.Errorf("Expected %q to be a closing bracket", b)
		}
	}

	openBrackets := []rune{'(', '[', '{', '<'}
	for _, b := range openBrackets {
		if m.IsCloseBracket(b) {
			t.Errorf("Expected %q to NOT be a closing bracket", b)
		}
	}
}

func TestBracketMatcherGetMatchingBracket(t *testing.T) {
	m := NewBracketMatcher()

	tests := []struct {
		input    rune
		expected rune
	}{
		{'(', ')'},
		{')', '('},
		{'[', ']'},
		{']', '['},
		{'{', '}'},
		{'}', '{'},
		{'<', '>'},
		{'>', '<'},
		{'a', 0},
		{' ', 0},
	}

	for _, tc := range tests {
		result := m.GetMatchingBracket(tc.input)
		if result != tc.expected {
			t.Errorf("GetMatchingBracket(%q) = %q, expected %q", tc.input, result, tc.expected)
		}
	}
}

// Helper to convert string to lines
func toLines(s string) [][]rune {
	parts := strings.Split(s, "\n")
	lines := make([][]rune, len(parts))
	for i, p := range parts {
		lines[i] = []rune(p)
	}
	return lines
}

func TestFindMatchSimpleParentheses(t *testing.T) {
	m := NewBracketMatcher()
	lines := toLines("(a)")

	// Cursor on '(' should find ')'
	matchLine, matchCol, found := m.FindMatch(lines, 0, 0, nil)
	if !found {
		t.Fatal("Expected to find match for '('")
	}
	if matchLine != 0 || matchCol != 2 {
		t.Errorf("Expected match at (0, 2), got (%d, %d)", matchLine, matchCol)
	}

	// Cursor on ')' should find '('
	matchLine, matchCol, found = m.FindMatch(lines, 0, 2, nil)
	if !found {
		t.Fatal("Expected to find match for ')'")
	}
	if matchLine != 0 || matchCol != 0 {
		t.Errorf("Expected match at (0, 0), got (%d, %d)", matchLine, matchCol)
	}
}

func TestFindMatchSimpleBrackets(t *testing.T) {
	m := NewBracketMatcher()
	lines := toLines("[a]")

	matchLine, matchCol, found := m.FindMatch(lines, 0, 0, nil)
	if !found || matchLine != 0 || matchCol != 2 {
		t.Errorf("Expected match at (0, 2), got (%d, %d, %v)", matchLine, matchCol, found)
	}
}

func TestFindMatchSimpleBraces(t *testing.T) {
	m := NewBracketMatcher()
	lines := toLines("{a}")

	matchLine, matchCol, found := m.FindMatch(lines, 0, 0, nil)
	if !found || matchLine != 0 || matchCol != 2 {
		t.Errorf("Expected match at (0, 2), got (%d, %d, %v)", matchLine, matchCol, found)
	}
}

func TestFindMatchSimpleAngleBrackets(t *testing.T) {
	m := NewBracketMatcher()
	lines := toLines("<T>")

	matchLine, matchCol, found := m.FindMatch(lines, 0, 0, nil)
	if !found || matchLine != 0 || matchCol != 2 {
		t.Errorf("Expected match at (0, 2), got (%d, %d, %v)", matchLine, matchCol, found)
	}
}

func TestFindMatchNested(t *testing.T) {
	m := NewBracketMatcher()
	lines := toLines("((a))")

	// Outer '(' at position 0 should match outer ')' at position 4
	matchLine, matchCol, found := m.FindMatch(lines, 0, 0, nil)
	if !found || matchLine != 0 || matchCol != 4 {
		t.Errorf("Expected match at (0, 4), got (%d, %d, %v)", matchLine, matchCol, found)
	}

	// Inner '(' at position 1 should match inner ')' at position 3
	matchLine, matchCol, found = m.FindMatch(lines, 0, 1, nil)
	if !found || matchLine != 0 || matchCol != 3 {
		t.Errorf("Expected match at (0, 3), got (%d, %d, %v)", matchLine, matchCol, found)
	}

	// Inner ')' at position 3 should match inner '('
	matchLine, matchCol, found = m.FindMatch(lines, 0, 3, nil)
	if !found || matchLine != 0 || matchCol != 1 {
		t.Errorf("Expected match at (0, 1), got (%d, %d, %v)", matchLine, matchCol, found)
	}
}

func TestFindMatchMixedBrackets(t *testing.T) {
	m := NewBracketMatcher()
	lines := toLines("{[()]}")

	// '{' at 0 should match '}' at 5
	matchLine, matchCol, found := m.FindMatch(lines, 0, 0, nil)
	if !found || matchLine != 0 || matchCol != 5 {
		t.Errorf("Expected match at (0, 5), got (%d, %d, %v)", matchLine, matchCol, found)
	}

	// '[' at 1 should match ']' at 4
	matchLine, matchCol, found = m.FindMatch(lines, 0, 1, nil)
	if !found || matchLine != 0 || matchCol != 4 {
		t.Errorf("Expected match at (0, 4), got (%d, %d, %v)", matchLine, matchCol, found)
	}

	// '(' at 2 should match ')' at 3
	matchLine, matchCol, found = m.FindMatch(lines, 0, 2, nil)
	if !found || matchLine != 0 || matchCol != 3 {
		t.Errorf("Expected match at (0, 3), got (%d, %d, %v)", matchLine, matchCol, found)
	}
}

func TestFindMatchMultiLine(t *testing.T) {
	m := NewBracketMatcher()
	lines := toLines("func main() {\n\tprintln()\n}")

	// '{' at line 0, col 12 should match '}' at line 2, col 0
	matchLine, matchCol, found := m.FindMatch(lines, 0, 12, nil)
	if !found {
		t.Fatal("Expected to find match for '{'")
	}
	if matchLine != 2 || matchCol != 0 {
		t.Errorf("Expected match at (2, 0), got (%d, %d)", matchLine, matchCol)
	}

	// '}' at line 2, col 0 should match '{' at line 0, col 12
	matchLine, matchCol, found = m.FindMatch(lines, 2, 0, nil)
	if !found {
		t.Fatal("Expected to find match for '}'")
	}
	if matchLine != 0 || matchCol != 12 {
		t.Errorf("Expected match at (0, 12), got (%d, %d)", matchLine, matchCol)
	}
}

func TestFindMatchUnmatched(t *testing.T) {
	m := NewBracketMatcher()

	// Unmatched opening bracket
	lines := toLines("(a")
	_, _, found := m.FindMatch(lines, 0, 0, nil)
	if found {
		t.Error("Expected no match for unmatched '('")
	}

	// Unmatched closing bracket
	lines = toLines("a)")
	_, _, found = m.FindMatch(lines, 0, 1, nil)
	if found {
		t.Error("Expected no match for unmatched ')'")
	}
}

func TestFindMatchWithSkipFunc(t *testing.T) {
	m := NewBracketMatcher()
	lines := toLines("(\")\")")  // ( " ) " )
	// Position:      0 1 2 3 4

	// Without skip function, '(' at 0 should find ')' at position 2
	matchLine, matchCol, found := m.FindMatch(lines, 0, 0, nil)
	if !found || matchLine != 0 || matchCol != 2 {
		t.Errorf("Without skip: expected match at (0, 2), got (%d, %d, %v)", matchLine, matchCol, found)
	}

	// With skip function that skips positions 1-3 (inside string quotes),
	// '(' at 0 should find ')' at position 4 (the real closing bracket)
	skipFunc := func(line, col int) bool {
		return col >= 1 && col <= 3
	}
	matchLine, matchCol, found = m.FindMatch(lines, 0, 0, skipFunc)
	if !found || matchLine != 0 || matchCol != 4 {
		t.Errorf("With skip: expected match at (0, 4), got (%d, %d, %v)", matchLine, matchCol, found)
	}
}

func TestFindMatchNotOnBracket(t *testing.T) {
	m := NewBracketMatcher()
	lines := toLines("abc")

	// Cursor on 'a' should not find any match
	_, _, found := m.FindMatch(lines, 0, 0, nil)
	if found {
		t.Error("Expected no match when cursor is not on a bracket")
	}
}

func TestFindMatchOutOfBounds(t *testing.T) {
	m := NewBracketMatcher()
	lines := toLines("()")

	// Line out of bounds
	_, _, found := m.FindMatch(lines, 5, 0, nil)
	if found {
		t.Error("Expected no match for out-of-bounds line")
	}

	// Column out of bounds
	_, _, found = m.FindMatch(lines, 0, 10, nil)
	if found {
		t.Error("Expected no match for out-of-bounds column")
	}

	// Negative indices
	_, _, found = m.FindMatch(lines, -1, 0, nil)
	if found {
		t.Error("Expected no match for negative line")
	}

	_, _, found = m.FindMatch(lines, 0, -1, nil)
	if found {
		t.Error("Expected no match for negative column")
	}
}

func TestFindMatchEmptyLines(t *testing.T) {
	m := NewBracketMatcher()
	var lines [][]rune

	_, _, found := m.FindMatch(lines, 0, 0, nil)
	if found {
		t.Error("Expected no match for empty lines")
	}
}

func TestFindMatchGoCode(t *testing.T) {
	m := NewBracketMatcher()
	code := `func main() {
	if x > 0 {
		fmt.Println("hello")
	}
}`
	lines := toLines(code)

	// Outer '{' at line 0, should match '}' at line 4
	matchLine, matchCol, found := m.FindMatch(lines, 0, 12, nil)
	if !found {
		t.Fatal("Expected to find match for outer '{'")
	}
	if matchLine != 4 {
		t.Errorf("Expected match at line 4, got line %d", matchLine)
	}

	// Inner '{' at line 1, should match '}' at line 3
	matchLine, matchCol, found = m.FindMatch(lines, 1, 10, nil)
	if !found {
		t.Fatal("Expected to find match for inner '{'")
	}
	if matchLine != 3 || matchCol != 1 {
		t.Errorf("Expected match at (3, 1), got (%d, %d)", matchLine, matchCol)
	}
}
