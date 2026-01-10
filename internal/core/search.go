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
	LastQuery     string
	CaseSensitive bool           // Session-level case sensitivity preference
	WholeWord     bool           // Session-level whole word preference
	Session       *SearchSession // Active search session (nil when not searching)
}

// SearchSession represents an active search session with real-time results.
type SearchSession struct {
	Query          string  // Current search term
	CaseSensitive  bool    // Case sensitivity for this search
	WholeWord      bool    // Whole word matching for this search
	Matches        []Match // All match positions in current buffer
	CurrentIndex   int     // Index of currently selected match (-1 if none)
	LastReplaceStr string  // Last replacement string used
}

func (s *SearchState) SetQuery(q string) {
	s.LastQuery = q
}

// Search finds the next occurrence of query starting from (line, col).
// Returns (line, col, true) if found, otherwise (-1, -1, false).
// caseSensitive controls whether the search is case-sensitive.
func Search(lines [][]rune, query string, startLine, startCol int, dir Direction, caseSensitive bool) (int, int, bool) {
	if query == "" {
		return -1, -1, false
	}

	if dir == SearchForward {
		return searchForward(lines, query, startLine, startCol, caseSensitive)
	} else {
		return searchBackward(lines, query, startLine, startCol, caseSensitive)
	}
}

func searchForward(lines [][]rune, query string, startLine, startCol int, caseSensitive bool) (int, int, bool) {
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

		// Search in the substring
		var matchIdx int
		if caseSensitive {
			matchIdx = strings.Index(lineStr[startIdx:], query)
		} else {
			matchIdx = indexCaseInsensitive(lineStr[startIdx:], query)
		}

		if matchIdx != -1 {
			return i, startIdx + matchIdx, true
		}
	}

	return -1, -1, false
}

func searchBackward(lines [][]rune, query string, startLine, startCol int, caseSensitive bool) (int, int, bool) {
	// Scan lines backwards
	for i := startLine; i >= 0; i-- {
		lineStr := string(lines[i])

		endIdx := len(lineStr)
		if i == startLine {
			endIdx = startCol
		}

		searchSpace := lineStr
		if i == startLine {
			if endIdx > len(lineStr) {
				endIdx = len(lineStr)
			}
			searchSpace = lineStr[:endIdx]
		}

		var matchIdx int
		if caseSensitive {
			matchIdx = strings.LastIndex(searchSpace, query)
		} else {
			matchIdx = lastIndexCaseInsensitive(searchSpace, query)
		}

		if matchIdx != -1 {
			return i, matchIdx, true
		}
	}
	return -1, -1, false
}

// indexCaseInsensitive finds the first occurrence of substr in s, case-insensitively.
// Returns -1 if not found.
func indexCaseInsensitive(s, substr string) int {
	if substr == "" {
		return 0
	}
	lowerS := strings.ToLower(s)
	lowerSubstr := strings.ToLower(substr)
	return strings.Index(lowerS, lowerSubstr)
}

// lastIndexCaseInsensitive finds the last occurrence of substr in s, case-insensitively.
// Returns -1 if not found.
func lastIndexCaseInsensitive(s, substr string) int {
	if substr == "" {
		if s == "" {
			return 0
		}
		return len(s)
	}
	lowerS := strings.ToLower(s)
	lowerSubstr := strings.ToLower(substr)
	return strings.LastIndex(lowerS, lowerSubstr)
}

// Match represents a single search match location.
type Match struct {
	Line   int
	Col    int
	Length int
}

// FindAllMatches finds all occurrences of query in the given lines.
// Returns a slice of Match structs containing the position and length of each match.
// The search respects the caseSensitive parameter.
// For performance, limits results to maxMatches (use 0 for unlimited).
func FindAllMatches(lines [][]rune, query string, caseSensitive bool, maxMatches int) []Match {
	if query == "" {
		return nil
	}

	matches := make([]Match, 0)
	queryLen := len(query)

	for lineNum, line := range lines {
		lineStr := string(line)
		offset := 0

		for offset < len(lineStr) {
			var matchIdx int
			if caseSensitive {
				matchIdx = strings.Index(lineStr[offset:], query)
			} else {
				matchIdx = indexCaseInsensitive(lineStr[offset:], query)
			}

			if matchIdx == -1 {
				break
			}

			// Found a match
			actualCol := offset + matchIdx
			matches = append(matches, Match{
				Line:   lineNum,
				Col:    actualCol,
				Length: queryLen,
			})

			// Check if we've hit the limit
			if maxMatches > 0 && len(matches) >= maxMatches {
				return matches
			}

			// Move past this match to find next one (avoid overlapping)
			offset = actualCol + queryLen
		}
	}

	return matches
}

// NewSearchSession creates a new search session with the given query and options.
func NewSearchSession(query string, caseSensitive bool, wholeWord bool) *SearchSession {
	return &SearchSession{
		Query:         query,
		CaseSensitive: caseSensitive,
		WholeWord:     wholeWord,
		Matches:       nil,
		CurrentIndex:  -1,
	}
}

// UpdateMatches updates the matches in the search session.
// This should be called when the search query changes or when the buffer changes.
func (s *SearchSession) UpdateMatches(lines [][]rune, maxMatches int) {
	s.Matches = FindAllMatches(lines, s.Query, s.CaseSensitive, maxMatches)
	// Reset to first match if we have any matches
	if len(s.Matches) > 0 {
		s.CurrentIndex = 0
	} else {
		s.CurrentIndex = -1
	}
}

// HasMatches returns true if there are any matches.
func (s *SearchSession) HasMatches() bool {
	return len(s.Matches) > 0
}

// GetCurrentMatch returns the current match, or nil if no matches.
func (s *SearchSession) GetCurrentMatch() *Match {
	if s.CurrentIndex >= 0 && s.CurrentIndex < len(s.Matches) {
		return &s.Matches[s.CurrentIndex]
	}
	return nil
}

// NextMatch moves to the next match. Wraps around to the first match.
func (s *SearchSession) NextMatch() {
	if len(s.Matches) == 0 {
		return
	}
	s.CurrentIndex = (s.CurrentIndex + 1) % len(s.Matches)
}

// PrevMatch moves to the previous match. Wraps around to the last match.
func (s *SearchSession) PrevMatch() {
	if len(s.Matches) == 0 {
		return
	}
	s.CurrentIndex--
	if s.CurrentIndex < 0 {
		s.CurrentIndex = len(s.Matches) - 1
	}
}

// MatchCount returns the total number of matches.
func (s *SearchSession) MatchCount() int {
	return len(s.Matches)
}
