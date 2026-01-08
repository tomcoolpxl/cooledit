package core

import (
	"os"
	"path/filepath"
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
