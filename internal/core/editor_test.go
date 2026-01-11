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
	"os"
	"path/filepath"
	"strings"
	"testing"

	"cooledit/internal/fileio"
)

type mockClipboard struct {
	text string
}

func (m *mockClipboard) Get() (string, error)  { return m.text, nil }
func (m *mockClipboard) Set(text string) error { m.text = text; return nil }

func newTestEditor() *Editor {
	return NewEditor(&mockClipboard{})
}

func TestInsertMarksModified(t *testing.T) {
	e := newTestEditor()

	if e.Modified() {
		t.Fatalf("new editor should not be modified")
	}

	e.Apply(CmdInsertRune{Rune: 'a'}, 10)

	if !e.Modified() {
		t.Fatalf("insert rune should mark editor modified")
	}
}

func TestNavigationDoesNotMarkModified(t *testing.T) {
	e := newTestEditor()

	e.Apply(CmdMoveRight{}, 10)
	e.Apply(CmdMoveLeft{}, 10)
	e.Apply(CmdMoveUp{}, 10)
	e.Apply(CmdMoveDown{}, 10)

	if e.Modified() {
		t.Fatalf("navigation must not mark editor modified")
	}
}

func TestSaveClearsModified(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")

	e := newTestEditor()
	e.Apply(CmdInsertRune{Rune: 'x'}, 10)

	if !e.Modified() {
		t.Fatalf("editor should be modified before save")
	}

	res := e.Apply(CmdSaveAs{Path: path}, 10)
	if res.Message == "" {
		t.Fatalf("expected save message")
	}

	if e.Modified() {
		t.Fatalf("save should clear modified flag")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read saved file: %v", err)
	}
	if string(data) != "x" {
		t.Fatalf("unexpected file contents: %q", string(data))
	}
}

func TestSaveWithoutPathFails(t *testing.T) {
	e := newTestEditor()
	e.Apply(CmdInsertRune{Rune: 'a'}, 10)

	res := e.Apply(CmdSave{}, 10)
	if res.Message == "" {
		t.Fatalf("expected error message when saving without path")
	}

	if !e.Modified() {
		t.Fatalf("failed save must not clear modified flag")
	}
}

func TestSaveWhenUnmodifiedIsNoOp(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")

	e := newTestEditor()
	e.Apply(CmdInsertRune{Rune: 'a'}, 10)
	e.Apply(CmdSaveAs{Path: path}, 10)

	if e.Modified() {
		t.Fatalf("editor should not be modified after save")
	}

	res := e.Apply(CmdSave{}, 10)
	if res.Message == "" {
		t.Fatalf("expected informational message on save with no changes")
	}

	if e.Modified() {
		t.Fatalf("save with no changes must not mark modified")
	}
}

func TestCursorMovement(t *testing.T) {
	e := newTestEditor()

	e.Apply(CmdInsertRune{Rune: 'a'}, 10)
	e.Apply(CmdInsertRune{Rune: 'b'}, 10)
	e.Apply(CmdInsertNewline{}, 10)
	e.Apply(CmdInsertRune{Rune: 'c'}, 10)

	row, col := e.Cursor()
	if row != 1 || col != 1 {
		t.Fatalf("expected cursor at (1,1), got (%d,%d)", row, col)
	}

	e.Apply(CmdMoveUp{}, 10)
	row, col = e.Cursor()
	if row != 0 || col != 1 {
		t.Fatalf("move up failed, got (%d,%d)", row, col)
	}

	e.Apply(CmdMoveHome{}, 10)
	row, col = e.Cursor()
	if col != 0 {
		t.Fatalf("move home failed, col=%d", col)
	}

	e.Apply(CmdMoveEnd{}, 10)
	_, col = e.Cursor()
	if col != 2 {
		t.Fatalf("move end failed, col=%d", col)
	}
}

func TestPageDownAndUp(t *testing.T) {
	e := newTestEditor()

	for i := 0; i < 50; i++ {
		e.Apply(CmdInsertRune{Rune: 'x'}, 10)
		e.Apply(CmdInsertNewline{}, 10)
	}

	e.Apply(CmdFileStart{}, 10)
	row, _ := e.Cursor()
	if row != 0 {
		t.Fatalf("file start failed, row=%d", row)
	}

	e.Apply(CmdPageDown{}, 5)
	row, _ = e.Cursor()
	if row != 5 {
		t.Fatalf("page down did not move cursor correctly, row=%d", row)
	}

	e.Apply(CmdPageUp{}, 5)
	row, _ = e.Cursor()
	if row != 0 {
		t.Fatalf("page up failed, row=%d", row)
	}
}

func TestNavigationEdgeCases(t *testing.T) {
	e := newTestEditor()

	// Test on empty buffer
	e.Apply(CmdMoveLeft{}, 10)
	e.Apply(CmdMoveUp{}, 10)
	row, col := e.Cursor()
	if row != 0 || col != 0 {
		t.Fatalf("expected (0,0) on empty buffer, got (%d,%d)", row, col)
	}

	e.Apply(CmdMoveRight{}, 10)
	e.Apply(CmdMoveDown{}, 10)
	row, col = e.Cursor()
	if row != 0 || col != 0 {
		t.Fatalf("expected (0,0) on empty buffer after right/down, got (%d,%d)", row, col)
	}

	// Test with content
	e.Apply(CmdInsertRune{Rune: 'a'}, 10)
	e.Apply(CmdInsertNewline{}, 10)
	e.Apply(CmdInsertRune{Rune: 'b'}, 10)
	// a
	// b|

	e.Apply(CmdMoveRight{}, 10) // No-op at EOF
	row, col = e.Cursor()
	if row != 1 || col != 1 {
		t.Fatalf("expected EOF at (1,1), got (%d,%d)", row, col)
	}

	e.Apply(CmdMoveLeft{}, 10) // At (1,0)
	e.Apply(CmdMoveLeft{}, 10) // Should wrap to (0,1)
	row, col = e.Cursor()
	if row != 0 || col != 1 {
		t.Fatalf("expected wrap to (0,1), got (%d,%d)", row, col)
	}
}

func TestInsertInMiddleOfBuffer(t *testing.T) {
	e := newTestEditor()
	e.Apply(CmdInsertRune{Rune: 'a'}, 10)
	e.Apply(CmdInsertRune{Rune: 'c'}, 10)
	e.Apply(CmdMoveLeft{}, 10)
	e.Apply(CmdInsertRune{Rune: 'b'}, 10)

	lines := e.Lines()
	if string(lines[0]) != "abc" {
		t.Fatalf("expected 'abc', got %q", string(lines[0]))
	}

	e.Apply(CmdMoveHome{}, 10)
	e.Apply(CmdInsertNewline{}, 10)
	// \n
	// abc
	lines = e.Lines()
	if len(lines) != 2 || len(lines[0]) != 0 || string(lines[1]) != "abc" {
		t.Fatalf("unexpected lines: %v", lines)
	}
}

func TestBackspaceMergesLines(t *testing.T) {
	e := newTestEditor()
	e.Apply(CmdInsertRune{Rune: 'a'}, 10)
	e.Apply(CmdInsertNewline{}, 10)
	e.Apply(CmdInsertRune{Rune: 'b'}, 10)
	// a
	// b

	e.Apply(CmdMoveHome{}, 10) // At (1,0)
	e.Apply(CmdBackspace{}, 10)

	lines := e.Lines()
	if len(lines) != 1 || string(lines[0]) != "ab" {
		t.Fatalf("expected merge to 'ab', got %v", lines)
	}

	row, col := e.Cursor()
	if row != 0 || col != 1 {
		t.Fatalf("expected cursor at (0,1) after merge, got (%d,%d)", row, col)
	}
}

func TestSearchEmptyQuery(t *testing.T) {
	e := newTestEditor()
	e.Apply(CmdInsertRune{Rune: 'a'}, 10)

	// CmdFind with empty
	res := e.Apply(CmdFind{Query: ""}, 10)
	if res.Message != "Not found: " {
		t.Fatalf("expected Not Found for empty query, got %q", res.Message)
	}

	// CmdFindNext with no previous
	res = e.Apply(CmdFindNext{}, 10)
	if res.Message != "No previous search" {
		t.Fatalf("expected No previous search, got %q", res.Message)
	}

	// CmdFindPrev with no previous
	res = e.Apply(CmdFindPrev{}, 10)
	if res.Message != "No previous search" {
		t.Fatalf("expected No previous search, got %q", res.Message)
	}
}

func TestEnsureVisibleSmall(t *testing.T) {
	e := newTestEditor()
	e.Apply(CmdInsertRune{Rune: 'a'}, 10)

	// Check EnsureVisible with 0/0 dims (should not crash)
	e.EnsureVisible(0, 0)
	vp := e.Viewport()
	if vp.TopLine != 0 {
		t.Errorf("EnsureVisible(0,0) moved TopLine")
	}
}

func TestEnsureVisibleScrolling(t *testing.T) {
	e := newTestEditor()

	// Create 20 lines
	for i := 0; i < 20; i++ {
		e.Apply(CmdInsertRune{Rune: 'x'}, 10)
		e.Apply(CmdInsertNewline{}, 10)
	}

	// Viewport size 5
	e.EnsureVisible(10, 5)
	vp := e.Viewport()
	// Cursor is at row 20. Viewport height 5.
	// TopLine should be 20 - 5 + 1 = 16.
	if vp.TopLine != 16 {
		t.Fatalf("expected TopLine 16, got %d", vp.TopLine)
	}

	e.Apply(CmdFileStart{}, 10)
	e.EnsureVisible(10, 5)
	vp = e.Viewport()
	if vp.TopLine != 0 {
		t.Fatalf("expected TopLine 0 after FileStart, got %d", vp.TopLine)
	}
}

func TestLoadFile(t *testing.T) {
	e := newTestEditor()
	fd := &fileio.FileData{
		Path:     "test.txt",
		BaseName: "test.txt",
		Lines:    [][]rune{{'h', 'i'}},
		EOL:      "\n",
		Encoding: "UTF-8",
	}

	e.LoadFile(fd)

	if e.File().Path != "test.txt" {
		t.Errorf("expected path test.txt, got %s", e.File().Path)
	}
	if string(e.Lines()[0]) != "hi" {
		t.Errorf("expected lines 'hi', got %q", string(e.Lines()[0]))
	}
	if e.Modified() {
		t.Errorf("loaded file should not be modified")
	}
}

func TestDeleteCharacter(t *testing.T) {
	e := newTestEditor()
	e.Apply(CmdInsertRune{Rune: 'a'}, 10)
	e.Apply(CmdInsertRune{Rune: 'b'}, 10)
	e.Apply(CmdInsertRune{Rune: 'c'}, 10)
	// "abc" cursor at (0,3)

	e.Apply(CmdMoveHome{}, 10) // cursor at (0,0)
	e.Apply(CmdDelete{}, 10)   // delete 'a'

	lines := e.Lines()
	if string(lines[0]) != "bc" {
		t.Fatalf("expected 'bc' after delete, got %q", string(lines[0]))
	}

	row, col := e.Cursor()
	if row != 0 || col != 0 {
		t.Fatalf("expected cursor at (0,0) after delete, got (%d,%d)", row, col)
	}
}

func TestDeleteMiddleCharacter(t *testing.T) {
	e := newTestEditor()
	e.Apply(CmdInsertRune{Rune: 'a'}, 10)
	e.Apply(CmdInsertRune{Rune: 'b'}, 10)
	e.Apply(CmdInsertRune{Rune: 'c'}, 10)
	// "abc" cursor at (0,3)

	e.Apply(CmdMoveLeft{}, 10) // cursor at (0,2)
	e.Apply(CmdMoveLeft{}, 10) // cursor at (0,1)
	e.Apply(CmdDelete{}, 10)   // delete 'b'

	lines := e.Lines()
	if string(lines[0]) != "ac" {
		t.Fatalf("expected 'ac' after delete, got %q", string(lines[0]))
	}

	row, col := e.Cursor()
	if row != 0 || col != 1 {
		t.Fatalf("expected cursor at (0,1) after delete, got (%d,%d)", row, col)
	}
}

func TestDeleteMergesLines(t *testing.T) {
	e := newTestEditor()
	e.Apply(CmdInsertRune{Rune: 'a'}, 10)
	e.Apply(CmdInsertNewline{}, 10)
	e.Apply(CmdInsertRune{Rune: 'b'}, 10)
	// a
	// b

	e.Apply(CmdMoveUp{}, 10) // cursor at (0,1)
	e.Apply(CmdDelete{}, 10) // delete newline, merge lines

	lines := e.Lines()
	if len(lines) != 1 || string(lines[0]) != "ab" {
		t.Fatalf("expected merge to 'ab', got %v", lines)
	}

	row, col := e.Cursor()
	if row != 0 || col != 1 {
		t.Fatalf("expected cursor at (0,1) after merge, got (%d,%d)", row, col)
	}
}

func TestDeleteOnEmptyLine(t *testing.T) {
	e := newTestEditor()
	e.Apply(CmdInsertNewline{}, 10)
	e.Apply(CmdInsertRune{Rune: 'a'}, 10)
	// (empty line)
	// a

	e.Apply(CmdMoveUp{}, 10) // cursor at (0,0) on empty line
	e.Apply(CmdDelete{}, 10) // should merge empty line with next

	lines := e.Lines()
	if len(lines) != 1 || string(lines[0]) != "a" {
		t.Fatalf("expected 'a' on single line, got %v", lines)
	}
}

func TestDeleteAtEndOfLastLine(t *testing.T) {
	e := newTestEditor()
	e.Apply(CmdInsertRune{Rune: 'a'}, 10)
	// "a" cursor at (0,1)

	e.Apply(CmdDelete{}, 10) // at end of last line, should be no-op

	lines := e.Lines()
	if string(lines[0]) != "a" {
		t.Fatalf("expected 'a' unchanged, got %q", string(lines[0]))
	}
}

func TestDeleteWithSelection(t *testing.T) {
	e := newTestEditor()
	e.Apply(CmdInsertRune{Rune: 'a'}, 10)
	e.Apply(CmdInsertRune{Rune: 'b'}, 10)
	e.Apply(CmdInsertRune{Rune: 'c'}, 10)
	// "abc"

	e.Apply(CmdMoveHome{}, 10)
	e.Apply(CmdMoveRight{Select: true}, 10) // select 'a'
	e.Apply(CmdMoveRight{Select: true}, 10) // select 'ab'

	e.Apply(CmdDelete{}, 10) // should delete selection

	lines := e.Lines()
	if string(lines[0]) != "c" {
		t.Fatalf("expected 'c' after delete selection, got %q", string(lines[0]))
	}

	if e.HasSelection() {
		t.Fatalf("selection should be cleared after delete")
	}
}

func TestSearchHighlightsText(t *testing.T) {
	e := newTestEditor()
	e.Apply(CmdInsertRune{Rune: 'h'}, 10)
	e.Apply(CmdInsertRune{Rune: 'e'}, 10)
	e.Apply(CmdInsertRune{Rune: 'l'}, 10)
	e.Apply(CmdInsertRune{Rune: 'l'}, 10)
	e.Apply(CmdInsertRune{Rune: 'o'}, 10)
	e.Apply(CmdMoveHome{}, 10)

	// Search for "ell"
	res := e.Apply(CmdFind{Query: "ell"}, 10)
	if res.Message != "Found: ell" {
		t.Fatalf("expected to find 'ell', got message: %s", res.Message)
	}

	// Should have selection
	if !e.HasSelection() {
		t.Fatalf("search should create selection")
	}

	sl, sc, el, ec := e.GetSelectionRange()
	if sl != 0 || sc != 1 || el != 0 || ec != 4 {
		t.Fatalf("expected selection (0,1)-(0,4), got (%d,%d)-(%d,%d)", sl, sc, el, ec)
	}
}

func TestReplaceOne(t *testing.T) {
	e := newTestEditor()
	// "hello world hello"
	for _, r := range "hello world hello" {
		e.Apply(CmdInsertRune{Rune: r}, 10)
	}
	e.Apply(CmdMoveHome{}, 10)

	// Find "hello"
	e.Apply(CmdFind{Query: "hello"}, 10)

	// Replace with "hi"
	e.Apply(CmdReplace{Find: "hello", Replace: "hi"}, 10)

	lines := e.Lines()
	text := string(lines[0])
	expected := "hi world hello"
	if text != expected {
		t.Fatalf("expected %q after replace, got %q", expected, text)
	}

	// Should be at next match
	if !e.HasSelection() {
		t.Fatalf("should have selection on next match")
	}
}

func TestReplaceAll(t *testing.T) {
	e := newTestEditor()
	// "hello world hello"
	for _, r := range "hello world hello" {
		e.Apply(CmdInsertRune{Rune: r}, 10)
	}
	e.Apply(CmdMoveHome{}, 10)

	// Replace all "hello" with "hi"
	res := e.Apply(CmdReplaceAll{Find: "hello", Replace: "hi"}, 10)

	if res.Message != "Replaced 2 occurrences - Ctrl+Z to undo" {
		t.Fatalf("expected 'Replaced 2 occurrences - Ctrl+Z to undo', got: %s", res.Message)
	}

	lines := e.Lines()
	text := string(lines[0])
	expected := "hi world hi"
	if text != expected {
		t.Fatalf("expected %q after replace all, got %q", expected, text)
	}
}

func TestReplaceNotFound(t *testing.T) {
	e := newTestEditor()
	for _, r := range "hello" {
		e.Apply(CmdInsertRune{Rune: r}, 10)
	}
	e.Apply(CmdMoveHome{}, 10)

	res := e.Apply(CmdReplaceAll{Find: "xyz", Replace: "abc"}, 10)

	if res.Message != "No matches found" {
		t.Fatalf("expected 'No matches found', got: %s", res.Message)
	}

	lines := e.Lines()
	text := string(lines[0])
	if text != "hello" {
		t.Fatalf("text should be unchanged, got %q", text)
	}
}

func TestReplaceUndoable(t *testing.T) {
	e := newTestEditor()
	for _, r := range "hello world" {
		e.Apply(CmdInsertRune{Rune: r}, 10)
	}
	e.Apply(CmdMoveHome{}, 10)

	// Replace all
	e.Apply(CmdReplaceAll{Find: "hello", Replace: "hi"}, 10)

	lines := e.Lines()
	if string(lines[0]) != "hi world" {
		t.Fatalf("expected 'hi world' after replace, got %q", string(lines[0]))
	}

	// Undo
	e.Apply(CmdUndo{}, 10)

	lines = e.Lines()
	if string(lines[0]) != "hello world" {
		t.Fatalf("expected 'hello world' after undo, got %q", string(lines[0]))
	}
}

func TestFindNextNoOverlapping(t *testing.T) {
	e := newTestEditor()
	// "ttttt" should find "ttt" only once at position 0
	for _, r := range "ttttt" {
		e.Apply(CmdInsertRune{Rune: r}, 10)
	}
	e.Apply(CmdMoveHome{}, 10)

	// First find
	res := e.Apply(CmdFind{Query: "ttt"}, 10)
	if res.Message != "Found: ttt" {
		t.Fatalf("expected to find 'ttt', got: %s", res.Message)
	}

	line, col := e.Cursor()
	if line != 0 || col != 0 {
		t.Fatalf("expected cursor at (0,0), got (%d,%d)", line, col)
	}

	// Next find should not overlap
	res = e.Apply(CmdFindNext{}, 10)
	if res.Message != "Not found (next): ttt" {
		t.Fatalf("expected no more matches, got: %s", res.Message)
	}
}

func TestFindNextTwoNonOverlapping(t *testing.T) {
	e := newTestEditor()
	// "ttttttt" should find "ttt" twice: at positions 0 and 3
	for _, r := range "ttttttt" {
		e.Apply(CmdInsertRune{Rune: r}, 10)
	}
	e.Apply(CmdMoveHome{}, 10)

	// First find at position 0
	res := e.Apply(CmdFind{Query: "ttt"}, 10)
	if res.Message != "Found: ttt" {
		t.Fatalf("expected to find 'ttt', got: %s", res.Message)
	}

	line, col := e.Cursor()
	if line != 0 || col != 0 {
		t.Fatalf("expected cursor at (0,0), got (%d,%d)", line, col)
	}

	// Second find at position 3
	res = e.Apply(CmdFindNext{}, 10)
	if res.Message != "Found next: ttt" {
		t.Fatalf("expected to find next 'ttt', got: %s", res.Message)
	}

	line, col = e.Cursor()
	if line != 0 || col != 3 {
		t.Fatalf("expected cursor at (0,3), got (%d,%d)", line, col)
	}

	// No more matches
	res = e.Apply(CmdFindNext{}, 10)
	if res.Message != "Not found (next): ttt" {
		t.Fatalf("expected no more matches, got: %s", res.Message)
	}
}

func TestReplaceAllFromBeginning(t *testing.T) {
	e := newTestEditor()
	// "foo bar foo baz foo"
	for _, r := range "foo bar foo baz foo" {
		e.Apply(CmdInsertRune{Rune: r}, 10)
	}

	// Move cursor to middle of text (position 10)
	e.Apply(CmdMoveHome{}, 10)
	for i := 0; i < 10; i++ {
		e.Apply(CmdMoveRight{}, 10)
	}

	_, col := e.Cursor()
	if col != 10 {
		t.Fatalf("cursor should be at col 10, got %d", col)
	}

	// Replace all should start from beginning, not cursor position
	res := e.Apply(CmdReplaceAll{Find: "foo", Replace: "XXX"}, 10)
	if res.Message != "Replaced 3 occurrences - Ctrl+Z to undo" {
		t.Fatalf("expected 'Replaced 3 occurrences - Ctrl+Z to undo', got: %s", res.Message)
	}

	lines := e.Lines()
	text := string(lines[0])
	expected := "XXX bar XXX baz XXX"
	if text != expected {
		t.Fatalf("expected %q, got %q", expected, text)
	}
}

func TestReplaceAllUndoable(t *testing.T) {
	e := newTestEditor()
	for _, r := range "hello world hello again hello" {
		e.Apply(CmdInsertRune{Rune: r}, 10)
	}
	e.Apply(CmdMoveHome{}, 10)

	// Save original text
	originalText := string(e.Lines()[0])

	// Replace all "hello" with "HELLO"
	res := e.Apply(CmdReplaceAll{Find: "hello", Replace: "HELLO"}, 10)
	if res.Message != "Replaced 3 occurrences - Ctrl+Z to undo" {
		t.Fatalf("expected 'Replaced 3 occurrences - Ctrl+Z to undo', got: %s", res.Message)
	}

	// Verify replacements happened
	lines := e.Lines()
	text := string(lines[0])
	expected := "HELLO world HELLO again HELLO"
	if text != expected {
		t.Fatalf("expected %q after replace all, got %q", expected, text)
	}

	// Undo should revert ALL replacements in a single undo
	e.Apply(CmdUndo{}, 10)
	lines = e.Lines()
	text = string(lines[0])
	if text != originalText {
		t.Fatalf("expected %q after single undo, got %q", originalText, text)
	}

	// Redo should reapply ALL replacements
	e.Apply(CmdRedo{}, 10)
	lines = e.Lines()
	text = string(lines[0])
	if text != expected {
		t.Fatalf("expected %q after redo, got %q", expected, text)
	}
}

func TestEditorIncrementalSearch(t *testing.T) {
	e := newTestEditor()

	// Create test content with newlines
	for _, r := range "hello world" {
		e.Apply(CmdInsertRune{Rune: r}, 10)
	}
	e.Apply(CmdInsertNewline{}, 10)
	for _, r := range "hello again" {
		e.Apply(CmdInsertRune{Rune: r}, 10)
	}
	e.Apply(CmdInsertNewline{}, 10)
	for _, r := range "goodbye" {
		e.Apply(CmdInsertRune{Rune: r}, 10)
	}
	e.Apply(CmdMoveHome{}, 10)

	// Start search session
	e.SearchState().CaseSensitive = false
	matches := FindAllMatches(e.Lines(), "hello", false, false, 0)

	if len(matches) != 2 {
		t.Fatalf("expected 2 matches for 'hello', got %d", len(matches))
	}

	// Create search session
	e.SearchState().Session = &SearchSession{
		Query:         "hello",
		CaseSensitive: false,
		WholeWord:     false,
		Matches:       matches,
		CurrentIndex:  0,
	}

	// Verify session state
	if e.SearchState().Session.Query != "hello" {
		t.Errorf("expected query 'hello', got '%s'", e.SearchState().Session.Query)
	}
	if len(e.SearchState().Session.Matches) != 2 {
		t.Errorf("expected 2 matches in session, got %d", len(e.SearchState().Session.Matches))
	}

	// Verify match positions
	if matches[0].Line != 0 || matches[0].Col != 0 {
		t.Errorf("expected first match at (0, 0), got (%d, %d)", matches[0].Line, matches[0].Col)
	}
	if matches[1].Line != 1 || matches[1].Col != 0 {
		t.Errorf("expected second match at (1, 0), got (%d, %d)", matches[1].Line, matches[1].Col)
	}
}

func TestToggleCaseSensitivity(t *testing.T) {
	e := newTestEditor()

	// Default should be case-insensitive (false)
	if e.SearchState().CaseSensitive {
		t.Error("expected default case sensitivity to be false")
	}

	// Toggle to case-sensitive
	e.ToggleCaseSensitivity()
	if !e.SearchState().CaseSensitive {
		t.Error("expected case sensitivity to be true after first toggle")
	}

	// Toggle back to case-insensitive
	e.ToggleCaseSensitivity()
	if e.SearchState().CaseSensitive {
		t.Error("expected case sensitivity to be false after second toggle")
	}
}

func TestToggleWholeWord(t *testing.T) {
	e := newTestEditor()

	// Default should be false
	if e.SearchState().WholeWord {
		t.Error("expected default whole word to be false")
	}

	// Toggle to true
	e.ToggleWholeWord()
	if !e.SearchState().WholeWord {
		t.Error("expected whole word to be true after first toggle")
	}

	// Toggle back to false
	e.ToggleWholeWord()
	if e.SearchState().WholeWord {
		t.Error("expected whole word to be false after second toggle")
	}
}

func TestSearchSessionPersistence(t *testing.T) {
	e := newTestEditor()

	// Create test content
	for _, r := range "test testing tested" {
		e.Apply(CmdInsertRune{Rune: r}, 10)
	}

	// Perform first search
	e.SearchState().LastQuery = "test"
	e.SearchState().CaseSensitive = true
	e.SearchState().WholeWord = false

	// Verify state persists
	if e.SearchState().LastQuery != "test" {
		t.Errorf("expected last query 'test', got '%s'", e.SearchState().LastQuery)
	}
	if !e.SearchState().CaseSensitive {
		t.Error("expected case sensitivity to persist")
	}

	// Perform second search with different settings
	e.SearchState().LastQuery = "testing"
	e.SearchState().CaseSensitive = false

	// Verify new state
	if e.SearchState().LastQuery != "testing" {
		t.Errorf("expected last query 'testing', got '%s'", e.SearchState().LastQuery)
	}
	if e.SearchState().CaseSensitive {
		t.Error("expected case sensitivity to be false")
	}

	// End session and verify state is cleared properly
	e.EndSearchSession()
	if e.SearchState().Session != nil {
		t.Error("expected session to be nil after ending")
	}

	// But preferences should persist
	if e.SearchState().LastQuery != "testing" {
		t.Error("expected last query to persist after ending session")
	}
}

func TestSearchStateAfterFileChange(t *testing.T) {
	e := newTestEditor()

	// Create initial content and search
	for _, r := range "hello world" {
		e.Apply(CmdInsertRune{Rune: r}, 10)
	}

	e.SearchState().LastQuery = "hello"
	e.SearchState().Session = &SearchSession{
		Query:        "hello",
		Matches:      []Match{{Line: 0, Col: 0, Length: 5}},
		CurrentIndex: 0,
	}

	// Load a new file (simulated by SetNewFile)
	e.SetNewFile("")

	// Session should be cleared
	if e.SearchState().Session != nil {
		t.Error("expected search session to be cleared after file change")
	}

	// But preferences should persist
	if e.SearchState().LastQuery != "hello" {
		t.Error("expected last query to persist after file change")
	}
}

// Tests for Smart Home Key feature

func TestSmartHomeFromMiddle(t *testing.T) {
	e := newTestEditor()

	// Insert "    hello" (4 spaces then hello)
	for _, r := range "    hello" {
		e.Apply(CmdInsertRune{Rune: r}, 10)
	}

	// Cursor is at end (col 9)
	line, col := e.Cursor()
	if col != 9 {
		t.Fatalf("expected cursor at col 9, got %d", col)
	}

	// Press Home - should go to first non-whitespace (col 4)
	e.Apply(CmdMoveHome{}, 10)
	line, col = e.Cursor()
	if line != 0 || col != 4 {
		t.Fatalf("expected cursor at (0, 4), got (%d, %d)", line, col)
	}
}

func TestSmartHomeFromFirstNonWS(t *testing.T) {
	e := newTestEditor()

	// Insert "    hello" (4 spaces then hello)
	for _, r := range "    hello" {
		e.Apply(CmdInsertRune{Rune: r}, 10)
	}

	// Move to first non-whitespace (col 4)
	e.Apply(CmdMoveHome{}, 10)
	line, col := e.Cursor()
	if col != 4 {
		t.Fatalf("expected cursor at col 4, got %d", col)
	}

	// Press Home again - should go to column 0
	e.Apply(CmdMoveHome{}, 10)
	line, col = e.Cursor()
	if line != 0 || col != 0 {
		t.Fatalf("expected cursor at (0, 0), got (%d, %d)", line, col)
	}
}

func TestSmartHomeFromCol0(t *testing.T) {
	e := newTestEditor()

	// Insert "    hello" (4 spaces then hello)
	for _, r := range "    hello" {
		e.Apply(CmdInsertRune{Rune: r}, 10)
	}

	// Move to column 0
	e.buf.SetCursor(0, 0)
	line, col := e.Cursor()
	if col != 0 {
		t.Fatalf("expected cursor at col 0, got %d", col)
	}

	// Press Home - should go to first non-whitespace (col 4)
	e.Apply(CmdMoveHome{}, 10)
	line, col = e.Cursor()
	if line != 0 || col != 4 {
		t.Fatalf("expected cursor at (0, 4), got (%d, %d)", line, col)
	}
}

func TestSmartHomeNoIndent(t *testing.T) {
	e := newTestEditor()

	// Insert "hello" (no leading whitespace)
	for _, r := range "hello" {
		e.Apply(CmdInsertRune{Rune: r}, 10)
	}

	// Move to middle
	e.buf.SetCursor(0, 2)

	// Press Home - should go to column 0 (first non-WS is at 0)
	e.Apply(CmdMoveHome{}, 10)
	line, col := e.Cursor()
	if line != 0 || col != 0 {
		t.Fatalf("expected cursor at (0, 0), got (%d, %d)", line, col)
	}

	// Press Home again - still at column 0
	e.Apply(CmdMoveHome{}, 10)
	line, col = e.Cursor()
	if line != 0 || col != 0 {
		t.Fatalf("expected cursor at (0, 0), got (%d, %d)", line, col)
	}
}

func TestSmartHomeAllWhitespace(t *testing.T) {
	e := newTestEditor()

	// Insert "    " (all spaces)
	for _, r := range "    " {
		e.Apply(CmdInsertRune{Rune: r}, 10)
	}

	// Move to column 2
	e.buf.SetCursor(0, 2)

	// Press Home - should go to column 0 (no non-whitespace)
	e.Apply(CmdMoveHome{}, 10)
	line, col := e.Cursor()
	if line != 0 || col != 0 {
		t.Fatalf("expected cursor at (0, 0), got (%d, %d)", line, col)
	}
}

func TestSmartHomeWithTabs(t *testing.T) {
	e := newTestEditor()

	// Insert "\t\thello" (2 tabs then hello)
	for _, r := range "\t\thello" {
		e.Apply(CmdInsertRune{Rune: r}, 10)
	}

	// Cursor is at end (col 7)
	line, col := e.Cursor()
	if col != 7 {
		t.Fatalf("expected cursor at col 7, got %d", col)
	}

	// Press Home - should go to first non-whitespace (col 2)
	e.Apply(CmdMoveHome{}, 10)
	line, col = e.Cursor()
	if line != 0 || col != 2 {
		t.Fatalf("expected cursor at (0, 2), got (%d, %d)", line, col)
	}

	// Press Home again - should go to column 0
	e.Apply(CmdMoveHome{}, 10)
	line, col = e.Cursor()
	if line != 0 || col != 0 {
		t.Fatalf("expected cursor at (0, 0), got (%d, %d)", line, col)
	}
}

func TestSmartHomeWithSelection(t *testing.T) {
	e := newTestEditor()

	// Insert "    hello"
	for _, r := range "    hello" {
		e.Apply(CmdInsertRune{Rune: r}, 10)
	}

	// Press Shift+Home - should select to first non-whitespace
	e.Apply(CmdMoveHome{Select: true}, 10)
	line, col := e.Cursor()
	if line != 0 || col != 4 {
		t.Fatalf("expected cursor at (0, 4), got (%d, %d)", line, col)
	}

	// Should have selection
	if !e.HasSelection() {
		t.Fatal("expected selection to be active")
	}
}

// Tests for Indent/Unindent Block feature

func TestIndentBlockSingleLine(t *testing.T) {
	e := newTestEditor()

	// Insert "hello"
	for _, r := range "hello" {
		e.Apply(CmdInsertRune{Rune: r}, 10)
	}

	// No selection, cursor at end
	// Indent current line (single line, no selection)
	e.Apply(CmdIndentBlock{}, 10)

	lines := e.Lines()
	// Line should now have 4 spaces at the beginning
	expected := "    hello"
	if string(lines[0]) != expected {
		t.Fatalf("expected %q, got %q", expected, string(lines[0]))
	}
}

func TestIndentBlockWithSelection(t *testing.T) {
	e := newTestEditor()

	// Insert multiple lines
	for _, r := range "line1" {
		e.Apply(CmdInsertRune{Rune: r}, 10)
	}
	e.Apply(CmdInsertNewline{}, 10)
	for _, r := range "line2" {
		e.Apply(CmdInsertRune{Rune: r}, 10)
	}
	e.Apply(CmdInsertNewline{}, 10)
	for _, r := range "line3" {
		e.Apply(CmdInsertRune{Rune: r}, 10)
	}

	// Select lines 0-1 (first two lines)
	e.buf.SetCursor(0, 0)
	e.SetSelection(0, 0, 5) // Select from (0,0) with length 5

	// Create a proper selection spanning multiple lines
	e.selectionActive = true
	e.selectionAnchor.Line = 0
	e.selectionAnchor.Col = 0
	e.buf.SetCursor(1, 5)

	// Indent selected lines
	e.Apply(CmdIndentBlock{}, 10)

	lines := e.Lines()
	// Lines 0 and 1 should be indented
	if !hasPrefix(lines[0], "    ") {
		t.Fatalf("line 0 should be indented, got %q", string(lines[0]))
	}
	if !hasPrefix(lines[1], "    ") {
		t.Fatalf("line 1 should be indented, got %q", string(lines[1]))
	}
	// Line 2 should NOT be indented
	if hasPrefix(lines[2], "    ") {
		t.Fatalf("line 2 should NOT be indented, got %q", string(lines[2]))
	}
}

func TestUnindentBlockSingleLine(t *testing.T) {
	e := newTestEditor()

	// Insert "    hello" (4 spaces then hello)
	for _, r := range "    hello" {
		e.Apply(CmdInsertRune{Rune: r}, 10)
	}

	// No selection
	e.ClearSelection()

	// Unindent current line
	e.Apply(CmdUnindentBlock{}, 10)

	lines := e.Lines()
	// Line should now have no leading spaces
	expected := "hello"
	if string(lines[0]) != expected {
		t.Fatalf("expected %q, got %q", expected, string(lines[0]))
	}
}

func TestUnindentBlockPartialSpaces(t *testing.T) {
	e := newTestEditor()

	// Insert "  hello" (2 spaces then hello)
	for _, r := range "  hello" {
		e.Apply(CmdInsertRune{Rune: r}, 10)
	}

	// No selection
	e.ClearSelection()

	// Unindent current line - should only remove 2 spaces (not 4)
	e.Apply(CmdUnindentBlock{}, 10)

	lines := e.Lines()
	// Line should now have no leading spaces
	expected := "hello"
	if string(lines[0]) != expected {
		t.Fatalf("expected %q, got %q", expected, string(lines[0]))
	}
}

func TestUnindentBlockNoSpaces(t *testing.T) {
	e := newTestEditor()

	// Insert "hello" (no leading spaces)
	for _, r := range "hello" {
		e.Apply(CmdInsertRune{Rune: r}, 10)
	}

	// No selection
	e.ClearSelection()

	// Unindent current line - should do nothing
	result := e.Apply(CmdUnindentBlock{}, 10)

	lines := e.Lines()
	expected := "hello"
	if string(lines[0]) != expected {
		t.Fatalf("expected %q, got %q", expected, string(lines[0]))
	}
	if result.Message != "No indentation to remove" {
		t.Fatalf("expected 'No indentation to remove' message, got %q", result.Message)
	}
}

func TestIndentUnindentUndo(t *testing.T) {
	e := newTestEditor()

	// Insert "hello"
	for _, r := range "hello" {
		e.Apply(CmdInsertRune{Rune: r}, 10)
	}

	// Indent
	e.Apply(CmdIndentBlock{}, 10)

	lines := e.Lines()
	if string(lines[0]) != "    hello" {
		t.Fatalf("expected '    hello', got %q", string(lines[0]))
	}

	// Undo
	e.Apply(CmdUndo{}, 10)

	lines = e.Lines()
	if string(lines[0]) != "hello" {
		t.Fatalf("expected 'hello' after undo, got %q", string(lines[0]))
	}
}

func TestTabWithSelectionIndents(t *testing.T) {
	e := newTestEditor()

	// Insert "hello"
	for _, r := range "hello" {
		e.Apply(CmdInsertRune{Rune: r}, 10)
	}

	// Select all text
	e.selectionActive = true
	e.selectionAnchor.Line = 0
	e.selectionAnchor.Col = 0
	e.buf.SetCursor(0, 5)

	// Tab should indent the line (not delete and insert spaces)
	e.Apply(CmdTab{}, 10)

	lines := e.Lines()
	// Line should be indented, text preserved
	if string(lines[0]) != "    hello" {
		t.Fatalf("expected '    hello', got %q", string(lines[0]))
	}
}

// Helper function
func hasPrefix(line []rune, prefix string) bool {
	if len(line) < len(prefix) {
		return false
	}
	return string(line[:len(prefix)]) == prefix
}

// Tests for Trim Trailing Whitespace feature

func TestTrimTrailingWhitespaceOnSave(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")

	e := newTestEditor()
	e.TrimTrailingWhitespaceOnSave = true

	// Insert "hello   " (with trailing spaces)
	for _, r := range "hello   " {
		e.Apply(CmdInsertRune{Rune: r}, 10)
	}

	// Save the file
	e.Apply(CmdSaveAs{Path: path}, 10)

	// Read the saved file
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read saved file: %v", err)
	}

	// Should have trailing whitespace trimmed
	if string(data) != "hello" {
		t.Fatalf("expected 'hello', got %q", string(data))
	}
}

func TestTrimTrailingWhitespaceDisabled(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")

	e := newTestEditor()
	e.TrimTrailingWhitespaceOnSave = false // Disabled

	// Insert "hello   " (with trailing spaces)
	for _, r := range "hello   " {
		e.Apply(CmdInsertRune{Rune: r}, 10)
	}

	// Save the file
	e.Apply(CmdSaveAs{Path: path}, 10)

	// Read the saved file
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read saved file: %v", err)
	}

	// Should NOT have trailing whitespace trimmed
	if string(data) != "hello   " {
		t.Fatalf("expected 'hello   ', got %q", string(data))
	}
}

func TestTrimTrailingWhitespaceMultipleLines(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")

	e := newTestEditor()
	e.TrimTrailingWhitespaceOnSave = true

	// Insert multiple lines with trailing whitespace
	for _, r := range "line1   " {
		e.Apply(CmdInsertRune{Rune: r}, 10)
	}
	e.Apply(CmdInsertNewline{}, 10)
	for _, r := range "line2\t" {
		e.Apply(CmdInsertRune{Rune: r}, 10)
	}
	e.Apply(CmdInsertNewline{}, 10)
	for _, r := range "line3" {
		e.Apply(CmdInsertRune{Rune: r}, 10)
	}

	// Save the file
	e.Apply(CmdSaveAs{Path: path}, 10)

	// Read the saved file - expected result depends on EOL format
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read saved file: %v", err)
	}

	// Check that trailing whitespace was trimmed from all lines
	content := string(data)
	if strings.Contains(content, "line1   ") {
		t.Fatalf("line1 should have trailing spaces trimmed")
	}
	if strings.Contains(content, "line2\t") {
		t.Fatalf("line2 should have trailing tab trimmed")
	}
}

// Tests for Toggle Comment feature

func TestToggleCommentSingleLine(t *testing.T) {
	e := newTestEditor()

	// Insert "hello"
	for _, r := range "hello" {
		e.Apply(CmdInsertRune{Rune: r}, 10)
	}

	// Toggle comment with "//"
	e.Apply(CmdToggleComment{CommentPrefix: "//"}, 10)

	lines := e.Lines()
	if string(lines[0]) != "// hello" {
		t.Fatalf("expected '// hello', got %q", string(lines[0]))
	}
}

func TestToggleCommentUncomment(t *testing.T) {
	e := newTestEditor()

	// Insert "// hello" (already commented)
	for _, r := range "// hello" {
		e.Apply(CmdInsertRune{Rune: r}, 10)
	}

	// Toggle comment should uncomment
	e.Apply(CmdToggleComment{CommentPrefix: "//"}, 10)

	lines := e.Lines()
	if string(lines[0]) != "hello" {
		t.Fatalf("expected 'hello', got %q", string(lines[0]))
	}
}

func TestToggleCommentWithIndent(t *testing.T) {
	e := newTestEditor()

	// Insert "    hello" (indented)
	for _, r := range "    hello" {
		e.Apply(CmdInsertRune{Rune: r}, 10)
	}

	// Toggle comment
	e.Apply(CmdToggleComment{CommentPrefix: "//"}, 10)

	lines := e.Lines()
	// Comment should be added after the indent
	if string(lines[0]) != "    // hello" {
		t.Fatalf("expected '    // hello', got %q", string(lines[0]))
	}
}

func TestToggleCommentMultipleLines(t *testing.T) {
	e := newTestEditor()

	// Insert multiple lines
	for _, r := range "line1" {
		e.Apply(CmdInsertRune{Rune: r}, 10)
	}
	e.Apply(CmdInsertNewline{}, 10)
	for _, r := range "line2" {
		e.Apply(CmdInsertRune{Rune: r}, 10)
	}
	e.Apply(CmdInsertNewline{}, 10)
	for _, r := range "line3" {
		e.Apply(CmdInsertRune{Rune: r}, 10)
	}

	// Select lines 0-1
	e.selectionActive = true
	e.selectionAnchor.Line = 0
	e.selectionAnchor.Col = 0
	e.buf.SetCursor(1, 5)

	// Toggle comment
	e.Apply(CmdToggleComment{CommentPrefix: "#"}, 10)

	lines := e.Lines()
	if string(lines[0]) != "# line1" {
		t.Fatalf("line 0: expected '# line1', got %q", string(lines[0]))
	}
	if string(lines[1]) != "# line2" {
		t.Fatalf("line 1: expected '# line2', got %q", string(lines[1]))
	}
	// Line 2 should NOT be commented (not in selection)
	if string(lines[2]) != "line3" {
		t.Fatalf("line 2: expected 'line3', got %q", string(lines[2]))
	}
}

func TestToggleCommentPython(t *testing.T) {
	e := newTestEditor()

	// Insert Python-style comment
	for _, r := range "# hello" {
		e.Apply(CmdInsertRune{Rune: r}, 10)
	}

	// Toggle should uncomment
	e.Apply(CmdToggleComment{CommentPrefix: "#"}, 10)

	lines := e.Lines()
	if string(lines[0]) != "hello" {
		t.Fatalf("expected 'hello', got %q", string(lines[0]))
	}
}

func TestToggleCommentNoLanguage(t *testing.T) {
	e := newTestEditor()

	// Insert text
	for _, r := range "hello" {
		e.Apply(CmdInsertRune{Rune: r}, 10)
	}

	// Toggle comment with empty prefix (no language support)
	result := e.Apply(CmdToggleComment{CommentPrefix: ""}, 10)

	// Should return message about no comment syntax
	if result.Message != "No comment syntax for this language" {
		t.Fatalf("expected 'No comment syntax for this language', got %q", result.Message)
	}

	// Text should be unchanged
	lines := e.Lines()
	if string(lines[0]) != "hello" {
		t.Fatalf("expected 'hello', got %q", string(lines[0]))
	}
}

func TestToggleCommentUndo(t *testing.T) {
	e := newTestEditor()

	// Insert "hello"
	for _, r := range "hello" {
		e.Apply(CmdInsertRune{Rune: r}, 10)
	}

	// Toggle comment
	e.Apply(CmdToggleComment{CommentPrefix: "//"}, 10)

	lines := e.Lines()
	if string(lines[0]) != "// hello" {
		t.Fatalf("expected '// hello', got %q", string(lines[0]))
	}

	// Undo
	e.Apply(CmdUndo{}, 10)

	lines = e.Lines()
	if string(lines[0]) != "hello" {
		t.Fatalf("expected 'hello' after undo, got %q", string(lines[0]))
	}
}
