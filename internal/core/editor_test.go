package core

import (
	"os"
	"path/filepath"
	"testing"
)

func newTestEditor() *Editor {
	return NewEditor()
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
	if row <= 0 {
		t.Fatalf("page down did not move cursor, row=%d", row)
	}

	e.Apply(CmdPageUp{}, 5)
	row, _ = e.Cursor()
	if row != 0 {
		t.Fatalf("page up failed, row=%d", row)
	}
}
