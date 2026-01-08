package buffer

import (
	"testing"
)

func TestInsertRune(t *testing.T) {
	b := NewLineBuffer()
	
	// Insert 'a' at 0,0
	b.InsertRune('a')
	if l, c := b.Cursor(); l != 0 || c != 1 {
		t.Errorf("cursor at (%d,%d), expected (0,1)", l, c)
	}
	if string(b.Lines()[0]) != "a" {
		t.Errorf("expected 'a', got %q", string(b.Lines()[0]))
	}

	// Insert 'b' at 0,1
	b.InsertRune('b')
	if string(b.Lines()[0]) != "ab" {
		t.Errorf("expected 'ab', got %q", string(b.Lines()[0]))
	}
	
	// Insert 'x' at 0,1 (between a and b)
	b.SetCursor(0, 1)
	b.InsertRune('x')
	if string(b.Lines()[0]) != "axb" {
		t.Errorf("expected 'axb', got %q", string(b.Lines()[0]))
	}
}

func TestInsertNewline(t *testing.T) {
	b := NewLineBuffer()
	b.InsertRune('a')
	b.InsertRune('b')
	
	// Split "ab" -> "a", "b"
	b.SetCursor(0, 1)
	b.InsertNewline()
	
	lines := b.Lines()
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}
	if string(lines[0]) != "a" {
		t.Errorf("line 0: expected 'a', got %q", string(lines[0]))
	}
	if string(lines[1]) != "b" {
		t.Errorf("line 1: expected 'b', got %q", string(lines[1]))
	}
	
	if l, c := b.Cursor(); l != 1 || c != 0 {
		t.Errorf("cursor at (%d,%d), expected (1,0)", l, c)
	}
}

func TestBackspace(t *testing.T) {
	b := NewLineBuffer()
	b.InsertRune('a')
	b.InsertRune('b')
	// "ab"
	
	// Delete 'b'
	b.Backspace()
	if string(b.Lines()[0]) != "a" {
		t.Errorf("expected 'a', got %q", string(b.Lines()[0]))
	}
	
	// Delete 'a'
	b.Backspace()
	if len(b.Lines()[0]) != 0 {
		t.Errorf("expected empty line")
	}
	
	// Backspace at 0,0 (no-op)
	b.Backspace()
	if len(b.Lines()) != 1 {
		t.Errorf("expected 1 line")
	}
	
	// Merge lines
	b.InsertRune('a')
	b.InsertNewline()
	b.InsertRune('b')
	// a
	// b|
	
	b.SetCursor(1, 0) // Move to start of line 2
	b.Backspace() // merge b up to a
	
	lines := b.Lines()
	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(lines))
	}
	if string(lines[0]) != "ab" {
		t.Errorf("expected 'ab', got %q", string(lines[0]))
	}
}

func TestMovementClamping(t *testing.T) {
	b := NewLineBuffer()
	// Line 0: "long"
	// Line 1: "s"
	for _, r := range "long" {
		b.InsertRune(r)
	}
	b.InsertNewline()
	b.InsertRune('s')
	
	b.SetCursor(0, 4) // End of "long"
	b.MoveDown()
	
	if l, c := b.Cursor(); l != 1 || c != 1 {
		t.Errorf("expected (1,1) [end of s], got (%d,%d)", l, c)
	}
	
	b.MoveUp()
	if l, c := b.Cursor(); l != 0 || c != 4 {
		t.Errorf("expected (0,4) [restored col], got (%d,%d)", l, c)
	}
}

func TestMovementWrapping(t *testing.T) {
	b := NewLineBuffer()
	b.InsertRune('a')
	b.InsertNewline()
	b.InsertRune('b')
	
	b.SetCursor(0, 1) // After 'a'
	b.MoveRight()
	if l, c := b.Cursor(); l != 1 || c != 0 {
		t.Errorf("MoveRight should wrap to (1,0), got (%d,%d)", l, c)
	}
	
	b.MoveLeft()
	if l, c := b.Cursor(); l != 0 || c != 1 {
		t.Errorf("MoveLeft should wrap to (0,1), got (%d,%d)", l, c)
	}
}

func TestNewBufferFromLines(t *testing.T) {
	// Nil
	b := NewLineBufferFromLines(nil)
	if len(b.Lines()) != 1 {
		t.Errorf("expected 1 line for nil input")
	}
	
	// Empty
	b = NewLineBufferFromLines([][]rune{})
	if len(b.Lines()) != 1 {
		t.Errorf("expected 1 line for empty input")
	}
}

func TestMovementBoundaries(t *testing.T) {
	b := NewLineBuffer()
	
	// MoveDown on empty buffer
	b.MoveDown()
	if l, _ := b.Cursor(); l != 0 {
		t.Errorf("MoveDown on empty should stay at 0")
	}
	
	// MoveRight at EOF
	b.MoveRight()
	if l, c := b.Cursor(); l != 0 || c != 0 {
		t.Errorf("MoveRight at EOF should stay at 0,0")
	}
	
	b.InsertRune('a')
	// "a|"
	b.MoveRight()
	// "a" cursor at 1 (past char)
	if _, c := b.Cursor(); c != 1 {
		t.Errorf("expected cursor at 1")
	}
	
	// MoveRight again (should stay)
	b.MoveRight()
	if _, c := b.Cursor(); c != 1 {
		t.Errorf("expected cursor to stay at 1")
	}
}

func TestSetCursorBounds(t *testing.T) {
	b := NewLineBuffer()
	b.InsertRune('a')
	
	b.SetCursor(-1, -1)
	if l, c := b.Cursor(); l != 0 || c != 0 {
		t.Errorf("clamped low: (%d,%d)", l, c)
	}
	
	b.SetCursor(100, 100)
	if l, c := b.Cursor(); l != 0 || c != 1 {
		t.Errorf("clamped high: (%d,%d)", l, c)
	}
}
