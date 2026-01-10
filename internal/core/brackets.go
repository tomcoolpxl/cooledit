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

// MaxBracketSearchLines is the maximum number of lines to search for a matching bracket.
// This prevents slowdown on very large files.
const MaxBracketSearchLines = 1000

// BracketMatcher finds matching brackets in text.
type BracketMatcher struct {
	openBrackets  map[rune]rune // Maps opening bracket to closing bracket
	closeBrackets map[rune]rune // Maps closing bracket to opening bracket
}

// NewBracketMatcher creates a new BracketMatcher with standard bracket pairs.
func NewBracketMatcher() *BracketMatcher {
	m := &BracketMatcher{
		openBrackets:  make(map[rune]rune),
		closeBrackets: make(map[rune]rune),
	}

	// Define bracket pairs: (), [], {}, <>
	pairs := [][2]rune{
		{'(', ')'},
		{'[', ']'},
		{'{', '}'},
		{'<', '>'},
	}

	for _, pair := range pairs {
		m.openBrackets[pair[0]] = pair[1]
		m.closeBrackets[pair[1]] = pair[0]
	}

	return m
}

// IsBracket returns true if the rune is any bracket character.
func (m *BracketMatcher) IsBracket(r rune) bool {
	_, isOpen := m.openBrackets[r]
	_, isClose := m.closeBrackets[r]
	return isOpen || isClose
}

// IsOpenBracket returns true if the rune is an opening bracket.
func (m *BracketMatcher) IsOpenBracket(r rune) bool {
	_, ok := m.openBrackets[r]
	return ok
}

// IsCloseBracket returns true if the rune is a closing bracket.
func (m *BracketMatcher) IsCloseBracket(r rune) bool {
	_, ok := m.closeBrackets[r]
	return ok
}

// GetMatchingBracket returns the matching bracket for the given bracket.
// Returns 0 if the rune is not a bracket.
func (m *BracketMatcher) GetMatchingBracket(r rune) rune {
	if match, ok := m.openBrackets[r]; ok {
		return match
	}
	if match, ok := m.closeBrackets[r]; ok {
		return match
	}
	return 0
}

// SkipFunc is a function that returns true if a position should be skipped
// (e.g., because it's inside a string or comment).
type SkipFunc func(line, col int) bool

// FindMatch finds the matching bracket for the bracket at the given position.
// Returns the position of the matching bracket and whether a match was found.
// If skipFunc is provided, positions where skipFunc returns true are skipped.
func (m *BracketMatcher) FindMatch(lines [][]rune, line, col int, skipFunc SkipFunc) (matchLine, matchCol int, found bool) {
	if line < 0 || line >= len(lines) {
		return 0, 0, false
	}
	if col < 0 || col >= len(lines[line]) {
		return 0, 0, false
	}

	ch := lines[line][col]

	if m.IsOpenBracket(ch) {
		return m.findMatchForward(lines, line, col, ch, m.openBrackets[ch], skipFunc)
	}

	if m.IsCloseBracket(ch) {
		return m.findMatchBackward(lines, line, col, ch, m.closeBrackets[ch], skipFunc)
	}

	return 0, 0, false
}

// findMatchForward searches forward from the given position for the matching closing bracket.
func (m *BracketMatcher) findMatchForward(lines [][]rune, startLine, startCol int, openBracket, closeBracket rune, skipFunc SkipFunc) (int, int, bool) {
	depth := 1
	linesSearched := 0

	for lineIdx := startLine; lineIdx < len(lines) && linesSearched < MaxBracketSearchLines; lineIdx++ {
		line := lines[lineIdx]
		startColIdx := 0
		if lineIdx == startLine {
			startColIdx = startCol + 1 // Start after the opening bracket
		}

		for colIdx := startColIdx; colIdx < len(line); colIdx++ {
			// Skip if inside string/comment
			if skipFunc != nil && skipFunc(lineIdx, colIdx) {
				continue
			}

			ch := line[colIdx]
			if ch == openBracket {
				depth++
			} else if ch == closeBracket {
				depth--
				if depth == 0 {
					return lineIdx, colIdx, true
				}
			}
		}
		linesSearched++
	}

	return 0, 0, false
}

// findMatchBackward searches backward from the given position for the matching opening bracket.
func (m *BracketMatcher) findMatchBackward(lines [][]rune, startLine, startCol int, closeBracket, openBracket rune, skipFunc SkipFunc) (int, int, bool) {
	depth := 1
	linesSearched := 0

	for lineIdx := startLine; lineIdx >= 0 && linesSearched < MaxBracketSearchLines; lineIdx-- {
		line := lines[lineIdx]
		endColIdx := len(line) - 1
		if lineIdx == startLine {
			endColIdx = startCol - 1 // Start before the closing bracket
		}

		for colIdx := endColIdx; colIdx >= 0; colIdx-- {
			// Skip if inside string/comment
			if skipFunc != nil && skipFunc(lineIdx, colIdx) {
				continue
			}

			ch := line[colIdx]
			if ch == closeBracket {
				depth++
			} else if ch == openBracket {
				depth--
				if depth == 0 {
					return lineIdx, colIdx, true
				}
			}
		}
		linesSearched++
	}

	return 0, 0, false
}
