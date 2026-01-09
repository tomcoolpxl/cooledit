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
	l, c, found := Search(lines, "hello", 0, 0, SearchForward)
	if !found || l != 0 || c != 0 {
		t.Fatalf("failed to find first hello: %d, %d", l, c)
	}

	// Search "hello" from (0, 1) -> should find same if not skipped
	// The core Search function just scans from (startLine, startCol).
	// If startCol is 1, substring is "ello world". "hello" not found in line 0.
	// Moves to next lines. Finds in line 2.
	l, c, found = Search(lines, "hello", 0, 1, SearchForward)
	if !found || l != 2 || c != 8 {
		t.Fatalf("failed to find second hello: %d, %d", l, c)
	}

	// Search "foo"
	l, c, found = Search(lines, "foo", 0, 0, SearchForward)
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
	l, c, found := Search(lines, "hello", 2, 20, SearchBackward)
	if !found || l != 2 || c != 8 {
		t.Fatalf("failed to find last hello: %d, %d", l, c)
	}

	// Search "hello" from (2, 8) -> excludes start at 8?
	l, c, found = Search(lines, "hello", 2, 8, SearchBackward)
	if !found || l != 0 || c != 0 {
		t.Fatalf("failed to find first hello going back: %d, %d", l, c)
	}
}

func TestSearchCaseSensitivity(t *testing.T) {
	lines := [][]rune{
		[]rune("Hello World"),
	}

	_, _, found := Search(lines, "hello", 0, 0, SearchForward)
	if found {
		t.Fatalf("search should be case-sensitive (found 'hello' in 'Hello')")
	}

	_, _, found = Search(lines, "Hello", 0, 0, SearchForward)
	if !found {
		t.Fatalf("search failed to find exact case match")
	}
}

func TestSearchNotFound(t *testing.T) {
	lines := [][]rune{
		[]rune("abc"),
		[]rune("def"),
	}

	l, c, found := Search(lines, "xyz", 0, 0, SearchForward)
	if found {
		t.Fatalf("found non-existent string at (%d, %d)", l, c)
	}
}

func TestSearchFromCol(t *testing.T) {
	lines := [][]rune{
		[]rune("aaa"),
	}

	// Search 'a' from col 1
	_, c, found := Search(lines, "a", 0, 1, SearchForward)
	if !found || c != 1 {
		t.Fatalf("expected to find at col 1, got %d", c)
	}

	// Search 'a' from col 2
	_, c, found = Search(lines, "a", 0, 2, SearchForward)
	if !found || c != 2 {
		t.Fatalf("expected to find at col 2, got %d", c)
	}

	// Search 'a' from col 3
	_, _, found = Search(lines, "a", 0, 3, SearchForward)
	if found {
		t.Fatalf("should not find 'a' starting at col 3")
	}
}
