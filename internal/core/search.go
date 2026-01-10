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

// Direction specifies the direction of a search operation.
type Direction int

const (
	// SearchForward searches from the start position towards the end of the document.
	SearchForward Direction = iota
	// SearchBackward searches from the start position towards the beginning of the document.
	SearchBackward
)

// SearchState maintains search preferences and the active search session.
// This struct contains both session-level preferences (CaseSensitive, WholeWord) which persist
// across multiple searches, and the active search session (Session) which contains the current
// search results and query.
//
// The session-level preferences are preserved even after EndSearchSession() is called,
// ensuring that user preferences like case sensitivity persist throughout the editing session.
//
// Example usage:
//
//	e := NewEditor(clipboard)
//	e.search.CaseSensitive = true // Set preference
//	e.StartSearchSession("query") // Start session with case-sensitive search
//	e.EndSearchSession()          // Clear session but preserve CaseSensitive setting
//	e.StartSearchSession("other") // Next search still uses CaseSensitive=true
type SearchState struct {
	LastQuery     string         // Last search query used
	CaseSensitive bool           // Session-level case sensitivity preference (persists across searches)
	WholeWord     bool           // Session-level whole word preference (persists across searches)
	Session       *SearchSession // Active search session (nil when not searching)
}

// SearchSession represents an active search session with real-time results.
// A search session is created when the user enters search mode and contains all the
// match positions, current match index, and search options for that particular search.
//
// The session lifecycle:
//  1. Created via NewSearchSession() or StartSearchSession()
//  2. Updated via UpdateMatches() when query changes or buffer is modified
//  3. Navigation via NextMatch()/PrevMatch()
//  4. Destroyed via EndSearchSession()
//
// The session supports:
//   - Real-time match finding (all matches are pre-computed)
//   - Current match tracking (CurrentIndex points to active match)
//   - Performance limits (LimitReached indicates if maxMatches was hit)
//   - Replace operations (LastReplaceStr tracks replacement text)
type SearchSession struct {
	Query          string  // Current search term
	CaseSensitive  bool    // Case sensitivity for this search
	WholeWord      bool    // Whole word matching for this search
	Matches        []Match // All match positions in current buffer
	CurrentIndex   int     // Index of currently selected match (-1 if none)
	LastReplaceStr string  // Last replacement string used
	LimitReached   bool    // True if search hit the maxMatches limit
}

// SetQuery sets the last query used for search. This is typically called when
// a search is performed to remember the query for future searches.
func (s *SearchState) SetQuery(q string) {
	s.LastQuery = q
}

// Search finds the next occurrence of query starting from (line, col).
// Returns (line, col, true) if found, otherwise (-1, -1, false).
// caseSensitive controls whether the search is case-sensitive.
// wholeWord controls whether to match whole words only.
func Search(lines [][]rune, query string, startLine, startCol int, dir Direction, caseSensitive, wholeWord bool) (int, int, bool) {
	if query == "" {
		return -1, -1, false
	}

	if dir == SearchForward {
		return searchForward(lines, query, startLine, startCol, caseSensitive, wholeWord)
	} else {
		return searchBackward(lines, query, startLine, startCol, caseSensitive, wholeWord)
	}
}

// searchForward searches for the query string starting from (startLine, startCol) towards the end of the document.
// Returns (line, col, true) if found, (-1, -1, false) otherwise.
// Supports case-sensitive/insensitive matching and whole word matching.
func searchForward(lines [][]rune, query string, startLine, startCol int, caseSensitive, wholeWord bool) (int, int, bool) {
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
		searchIn := lineStr[startIdx:]
		offset := 0

		for {
			var matchIdx int
			if caseSensitive {
				matchIdx = strings.Index(searchIn, query)
			} else {
				matchIdx = indexCaseInsensitive(searchIn, query)
			}

			if matchIdx == -1 {
				break
			}

			actualPos := startIdx + offset + matchIdx

			// Check whole word boundary if needed
			if !wholeWord || isWholeWordMatch(lineStr, actualPos, len(query)) {
				return i, actualPos, true
			}

			// Move past this match and continue searching
			offset += matchIdx + 1
			searchIn = lineStr[startIdx+offset:]
		}
	}

	return -1, -1, false
}

// searchBackward searches for the query string starting from (startLine, startCol) towards the beginning of the document.
// Returns (line, col, true) if found, (-1, -1, false) otherwise.
// Supports case-sensitive/insensitive matching and whole word matching.
func searchBackward(lines [][]rune, query string, startLine, startCol int, caseSensitive, wholeWord bool) (int, int, bool) {
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

		// For backward search, find all matches and take the last one that satisfies whole word
		var lastValidMatch int = -1

		offset := 0
		for {
			var matchIdx int
			if caseSensitive {
				matchIdx = strings.Index(searchSpace[offset:], query)
			} else {
				matchIdx = indexCaseInsensitive(searchSpace[offset:], query)
			}

			if matchIdx == -1 {
				break
			}

			actualPos := offset + matchIdx

			// Check whole word boundary if needed
			if !wholeWord || isWholeWordMatch(lineStr, actualPos, len(query)) {
				lastValidMatch = actualPos
			}

			// Move past this match
			offset = actualPos + 1
		}

		if lastValidMatch != -1 {
			return i, lastValidMatch, true
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

// isWordBoundary checks if a position is at a word boundary.
// A word boundary is before/after a word character (letter, digit, underscore).
func isWordBoundary(s string, pos int) bool {
	if pos < 0 || pos > len(s) {
		return true
	}
	if pos == 0 || pos == len(s) {
		return true
	}
	// Check if characters on either side are different types (word vs non-word)
	before := pos > 0 && isWordRune(rune(s[pos-1]))
	at := pos < len(s) && isWordRune(rune(s[pos]))
	return before != at
}

// isWordRune returns true if the rune is a word character (letter, digit, underscore).
func isWordRune(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_'
}

// isWholeWordMatch checks if a match at the given position is a whole word match.
// A whole word match requires that the characters immediately before and after the match
// are non-word characters (or the match is at the start/end of the line).
// Word characters are: a-z, A-Z, 0-9, and underscore.
func isWholeWordMatch(s string, matchPos, matchLen int) bool {
	// Check start boundary
	if matchPos > 0 && isWordRune(rune(s[matchPos-1])) {
		return false
	}
	// Check end boundary
	endPos := matchPos + matchLen
	if endPos < len(s) && isWordRune(rune(s[endPos])) {
		return false
	}
	return true
}

// Match represents a single search match location in the document.
// Line and Col are 0-based indices into the document's line array.
// Length is the number of runes (not bytes) in the matched text.
//
// Example:
//
//	match := Match{Line: 5, Col: 10, Length: 4}
//	// This represents a 4-character match at line 6 (1-based), column 11 (1-based)
type Match struct {
	Line   int
	Col    int
	Length int
}

// FindAllMatches finds all occurrences of query in the given lines.
// Returns a slice of Match structs containing the position and length of each match.
// The search respects the caseSensitive and wholeWord parameters.
// For performance, limits results to maxMatches (use 0 for unlimited).
func FindAllMatches(lines [][]rune, query string, caseSensitive, wholeWord bool, maxMatches int) []Match {
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

			// Check whole word boundary if needed
			if !wholeWord || isWholeWordMatch(lineStr, actualCol, queryLen) {
				matches = append(matches, Match{
					Line:   lineNum,
					Col:    actualCol,
					Length: queryLen,
				})

				// Check if we've hit the limit
				if maxMatches > 0 && len(matches) >= maxMatches {
					return matches
				}
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
	s.Matches = FindAllMatches(lines, s.Query, s.CaseSensitive, s.WholeWord, maxMatches)
	// Check if we hit the limit
	s.LimitReached = maxMatches > 0 && len(s.Matches) >= maxMatches
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
