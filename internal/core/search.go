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

import "strings"

type Direction int

const (
	SearchForward Direction = iota
	SearchBackward
)

type SearchState struct {
	LastQuery string
}

func (s *SearchState) SetQuery(q string) {
	s.LastQuery = q
}

// Search finds the next occurrence of query starting from (line, col).
// Returns (line, col, true) if found, otherwise (-1, -1, false).
// If forward=true, searches forward (wrapping around? maybe not for now).
// For now: no wrap, linear scan.
func Search(lines [][]rune, query string, startLine, startCol int, dir Direction) (int, int, bool) {
	if query == "" {
		return -1, -1, false
	}

	if dir == SearchForward {
		return searchForward(lines, query, startLine, startCol)
	} else {
		return searchBackward(lines, query, startLine, startCol)
	}
}

func searchForward(lines [][]rune, query string, startLine, startCol int) (int, int, bool) {
	// Check current line starting from startCol
	// But if startCol is in middle of line, we need to match carefully.
	// Simplest: Convert line to string, search.
	
	// Scan lines
	for i := startLine; i < len(lines); i++ {
		lineStr := string(lines[i])
		
		startIdx := 0
		if i == startLine {
			startIdx = startCol
			// If we are at end of line, move to next
			if startIdx >= len(lineStr) {
				continue
			}
		}
		
		// Optimization: simple string search in the substring
		matchIdx := strings.Index(lineStr[startIdx:], query)
		if matchIdx != -1 {
			return i, startIdx + matchIdx, true
		}
	}
	
	return -1, -1, false
}

func searchBackward(lines [][]rune, query string, startLine, startCol int) (int, int, bool) {
	// Scan lines backwards
	for i := startLine; i >= 0; i-- {
		lineStr := string(lines[i])
		
		endIdx := len(lineStr)
		if i == startLine {
			endIdx = startCol
		}
		
		// We want the *last* occurrence that starts before endIdx.
		// LastIndex gives last occurrence in the whole string.
		// We search in lineStr[:endIdx] ?
		// Wait, LastIndex of "abc" in "abcabc" is 3.
		// If cursor is at 4.
		// We want to search in substring.
		
		searchSpace := lineStr
		if i == startLine {
			if endIdx > len(lineStr) {
				endIdx = len(lineStr)
			}
			searchSpace = lineStr[:endIdx]
		}
		
		matchIdx := strings.LastIndex(searchSpace, query)
		if matchIdx != -1 {
			return i, matchIdx, true
		}
	}
	return -1, -1, false
}