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

import "testing"

func TestSearchForward(t *testing.T) {
	lines := [][]rune{
		[]rune("hello world"),
		[]rune("foo bar"),
		[]rune("another hello here"),
	}

	// Search "hello" from start
	l, c, found := Search(lines, "hello", 0, 0, SearchForward, true, false)
	if !found || l != 0 || c != 0 {
		t.Fatalf("failed to find first hello: %d, %d", l, c)
	}

	// Search "hello" from (0, 1) -> should find same if not skipped
	// The core Search function just scans from (startLine, startCol).
	// If startCol is 1, substring is "ello world". "hello" not found in line 0.
	// Moves to next lines. Finds in line 2.
	l, c, found = Search(lines, "hello", 0, 1, SearchForward, true, false)
	if !found || l != 2 || c != 8 {
		t.Fatalf("failed to find second hello: %d, %d", l, c)
	}

	// Search "foo"
	l, c, found = Search(lines, "foo", 0, 0, SearchForward, true, false)
	if !found || l != 1 || c != 0 {
		t.Fatalf("failed to find foo: %d, %d", l, c)
	}
}

func TestSearchBackward(t *testing.T) {
	lines := [][]rune{
		[]rune("hello world"),
		[]rune("foo bar"),
		[]rune("another hello here"),
	}

	// Search "hello" from end (2, 20)
	l, c, found := Search(lines, "hello", 2, 20, SearchBackward, true, false)
	if !found || l != 2 || c != 8 {
		t.Fatalf("failed to find last hello: %d, %d", l, c)
	}

	// Search "hello" from (2, 8) -> excludes start at 8?
	l, c, found = Search(lines, "hello", 2, 8, SearchBackward, true, false)
	if !found || l != 0 || c != 0 {
		t.Fatalf("failed to find first hello going back: %d, %d", l, c)
	}
}

func TestSearchCaseSensitivity(t *testing.T) {
	lines := [][]rune{
		[]rune("Hello World"),
	}

	// Case-sensitive search should not find "hello" in "Hello"
	_, _, found := Search(lines, "hello", 0, 0, SearchForward, true, false)
	if found {
		t.Fatalf("case-sensitive search should not find 'hello' in 'Hello'")
	}

	// Case-sensitive search should find exact match
	_, _, found = Search(lines, "Hello", 0, 0, SearchForward, true, false)
	if !found {
		t.Fatalf("case-sensitive search failed to find exact case match")
	}

	// Case-insensitive search should find "hello" in "Hello"
	_, _, found = Search(lines, "hello", 0, 0, SearchForward, false, false)
	if !found {
		t.Fatalf("case-insensitive search should find 'hello' in 'Hello'")
	}

	// Case-insensitive search should find "HELLO" in "Hello"
	_, _, found = Search(lines, "HELLO", 0, 0, SearchForward, false, false)
	if !found {
		t.Fatalf("case-insensitive search should find 'HELLO' in 'Hello'")
	}
}

func TestSearchNotFound(t *testing.T) {
	lines := [][]rune{
		[]rune("abc"),
		[]rune("def"),
	}

	l, c, found := Search(lines, "xyz", 0, 0, SearchForward, true, false)
	if found {
		t.Fatalf("found non-existent string at (%d, %d)", l, c)
	}
}

func TestSearchFromCol(t *testing.T) {
	lines := [][]rune{
		[]rune("aaa"),
	}

	// Search 'a' from col 1
	_, c, found := Search(lines, "a", 0, 1, SearchForward, true, false)
	if !found || c != 1 {
		t.Fatalf("expected to find at col 1, got %d", c)
	}

	// Search 'a' from col 2
	_, c, found = Search(lines, "a", 0, 2, SearchForward, true, false)
	if !found || c != 2 {
		t.Fatalf("expected to find at col 2, got %d", c)
	}

	// Search 'a' from col 3
	_, _, found = Search(lines, "a", 0, 3, SearchForward, true, false)
	if found {
		t.Fatalf("should not find 'a' starting at col 3")
	}
}

func TestFindAllMatches(t *testing.T) {
	lines := [][]rune{
		[]rune("hello world hello"),
		[]rune("HELLO there"),
		[]rune("goodbye hello"),
	}

	// Case-sensitive: should find 3 matches of "hello"
	matches := FindAllMatches(lines, "hello", true, false, 0)
	if len(matches) != 3 {
		t.Fatalf("expected 3 case-sensitive matches, got %d", len(matches))
	}

	// Verify positions
	expected := []Match{
		{Line: 0, Col: 0, Length: 5},
		{Line: 0, Col: 12, Length: 5},
		{Line: 2, Col: 8, Length: 5},
	}

	for i, match := range matches {
		if match != expected[i] {
			t.Errorf("match %d: expected %+v, got %+v", i, expected[i], match)
		}
	}

	// Case-insensitive: should find 4 matches
	matches = FindAllMatches(lines, "hello", false, false, 0)
	if len(matches) != 4 {
		t.Fatalf("expected 4 case-insensitive matches, got %d", len(matches))
	}

	expectedInsensitive := []Match{
		{Line: 0, Col: 0, Length: 5},
		{Line: 0, Col: 12, Length: 5},
		{Line: 1, Col: 0, Length: 5},
		{Line: 2, Col: 8, Length: 5},
	}

	for i, match := range matches {
		if match != expectedInsensitive[i] {
			t.Errorf("case-insensitive match %d: expected %+v, got %+v", i, expectedInsensitive[i], match)
		}
	}
}

func TestFindAllMatchesWithLimit(t *testing.T) {
	lines := [][]rune{
		[]rune("test test test test test"),
	}

	// With limit of 3
	matches := FindAllMatches(lines, "test", true, false, 3)
	if len(matches) != 3 {
		t.Fatalf("expected 3 matches (limit), got %d", len(matches))
	}

	// Unlimited (0 means no limit)
	matches = FindAllMatches(lines, "test", true, false, 0)
	if len(matches) != 5 {
		t.Fatalf("expected 5 matches (unlimited), got %d", len(matches))
	}
}

func TestFindAllMatchesNoOverlap(t *testing.T) {
	lines := [][]rune{
		[]rune("aaa"),
	}

	// Should find 3 non-overlapping matches
	matches := FindAllMatches(lines, "a", true, false, 0)
	if len(matches) != 3 {
		t.Fatalf("expected 3 non-overlapping matches, got %d", len(matches))
	}

	for i, match := range matches {
		if match.Col != i {
			t.Errorf("match %d at wrong position: expected col %d, got %d", i, i, match.Col)
		}
	}
}

func TestFindAllMatchesEmpty(t *testing.T) {
	lines := [][]rune{
		[]rune("hello world"),
	}

	// Empty query
	matches := FindAllMatches(lines, "", true, false, 0)
	if matches != nil {
		t.Fatalf("expected nil for empty query, got %d matches", len(matches))
	}

	// No matches
	matches = FindAllMatches(lines, "xyz", true, false, 0)
	if len(matches) != 0 {
		t.Fatalf("expected 0 matches for non-existent pattern, got %d", len(matches))
	}
}
