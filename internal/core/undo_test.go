package core

import (
	"testing"
)

func TestUndoInsertRune(t *testing.T) {
	e := NewEditor(nil)
	
e.Apply(CmdInsertRune{Rune: 'a'}, 10)
	if lines := e.Lines(); len(lines) != 1 || len(lines[0]) != 1 || lines[0][0] != 'a' {
		t.Fatalf("setup failed: expected 'a'")
	}
	
e.Apply(CmdUndo{}, 10)
	if lines := e.Lines(); len(lines) != 1 || len(lines[0]) != 0 {
		t.Fatalf("undo failed: expected empty line")
	}
	
e.Apply(CmdRedo{}, 10)
	if lines := e.Lines(); len(lines) != 1 || len(lines[0]) != 1 || lines[0][0] != 'a' {
		t.Fatalf("redo failed: expected 'a'")
	}
}

func TestUndoInsertNewline(t *testing.T) {
	e := NewEditor(nil)
	e.Apply(CmdInsertRune{Rune: 'a'}, 10)
	e.Apply(CmdInsertNewline{}, 10)
	e.Apply(CmdInsertRune{Rune: 'b'}, 10)
	
	// State: "a\nb"
	lines := e.Lines()
	if len(lines) != 2 {
		t.Fatalf("setup failed: expected 2 lines")
	}
	
e.Apply(CmdUndo{}, 10) // Undo 'b'
	e.Apply(CmdUndo{}, 10) // Undo Newline
	
	lines = e.Lines()
	if len(lines) != 1 {
		t.Fatalf("undo newline failed: expected 1 line")
	}
	if string(lines[0]) != "a" {
		t.Fatalf("undo newline failed: expected 'a', got %q", string(lines[0]))
	}
}

func TestUndoBackspaceChar(t *testing.T) {
	e := NewEditor(nil)
	e.Apply(CmdInsertRune{Rune: 'a'}, 10)
	e.Apply(CmdBackspace{}, 10)
	
	// State: ""
	if len(e.Lines()[0]) != 0 {
		t.Fatalf("backspace failed")
	}
	
e.Apply(CmdUndo{}, 10)
	// State: "a"
	if string(e.Lines()[0]) != "a" {
		t.Fatalf("undo backspace failed: expected 'a'")
	}
}

func TestUndoBackspaceMerge(t *testing.T) {
	e := NewEditor(nil)
	e.Apply(CmdInsertRune{Rune: 'a'}, 10)
	e.Apply(CmdInsertNewline{}, 10)
	e.Apply(CmdInsertRune{Rune: 'b'}, 10)
	// a
	// b
	
e.Apply(CmdMoveHome{}, 10) // Cursor at 'b' (start of line 2)
	e.Apply(CmdBackspace{}, 10) // Merge line 2 into 1
	
	// State: "ab"
	lines := e.Lines()
	if len(lines) != 1 {
		t.Fatalf("merge failed: expected 1 line")
	}
	if string(lines[0]) != "ab" {
		t.Fatalf("merge content mismatch: expected 'ab', got %q", string(lines[0]))
	}
	
e.Apply(CmdUndo{}, 10)
	
	// State: "a", "b"
	lines = e.Lines()
	if len(lines) != 2 {
		t.Fatalf("undo merge failed: expected 2 lines")
	}
	if string(lines[0]) != "a" || string(lines[1]) != "b" {
		t.Fatalf("undo merge content mismatch")
	}
}

func TestModifiedStateWithUndo(t *testing.T) {

	e := NewEditor(nil)

	// Initial: Modified=false (SavedPtr=0, Ptr=0)

	if e.Modified() {

		t.Fatalf("initially modified")

	}

	

	e.Apply(CmdInsertRune{Rune: 'a'}, 10)

	// Ptr=1

	if !e.Modified() {

		t.Fatalf("should be modified after insert")

	}

	

	e.Apply(CmdUndo{}, 10)

	// Ptr=0

	if e.Modified() {

		t.Fatalf("should not be modified after undo to start")

	}

	

	e.Apply(CmdRedo{}, 10)

	// Ptr=1

	if !e.Modified() {

		t.Fatalf("should be modified after redo")

	}

}



func TestUndoRedoMultiStep(t *testing.T) {

	e := NewEditor(nil)

	e.Apply(CmdInsertRune{Rune: 'a'}, 10)

	e.Apply(CmdInsertRune{Rune: 'b'}, 10)

	e.Apply(CmdInsertRune{Rune: 'c'}, 10)

	

	e.Apply(CmdUndo{}, 10)

	e.Apply(CmdUndo{}, 10)

	if string(e.Lines()[0]) != "a" {

		t.Fatalf("expected 'a', got %q", string(e.Lines()[0]))

	}

	

	e.Apply(CmdRedo{}, 10)

	e.Apply(CmdRedo{}, 10)

	if string(e.Lines()[0]) != "abc" {

		t.Fatalf("expected 'abc', got %q", string(e.Lines()[0]))

	}

}



func TestRedoTruncation(t *testing.T) {

	e := NewEditor(nil)

	e.Apply(CmdInsertRune{Rune: 'a'}, 10)

	e.Apply(CmdInsertRune{Rune: 'b'}, 10)

	

	e.Apply(CmdUndo{}, 10) // state: "a", ptr: 1, history: ["a", "b"]

	e.Apply(CmdInsertRune{Rune: 'c'}, 10) // state: "ac", ptr: 2, history: ["a", "c"]

	

	e.Apply(CmdRedo{}, 10) // should be no-op

	if string(e.Lines()[0]) != "ac" {

		t.Fatalf("redo after truncation should be no-op")

	}

	

	e.Apply(CmdUndo{}, 10)

	e.Apply(CmdUndo{}, 10)

	if len(e.Lines()[0]) != 0 {

		t.Fatalf("expected empty buffer")

	}

}



func TestUndoToSavedState(t *testing.T) {

	e := NewEditor(nil)

	e.Apply(CmdInsertRune{Rune: 'a'}, 10)

	

	// Mock save

	e.undo.MarkSaved()

	if e.Modified() {

		t.Fatalf("should not be modified after save")

	}

	

	e.Apply(CmdInsertRune{Rune: 'b'}, 10)

	if !e.Modified() {

		t.Fatalf("should be modified after second insert")

	}

	

	e.Apply(CmdUndo{}, 10)

	if e.Modified() {

		t.Fatalf("should not be modified after undo to saved state")

	}

	

	e.Apply(CmdUndo{}, 10)

	if !e.Modified() {

		t.Fatalf("should be modified after undo past saved state")

	}

}


