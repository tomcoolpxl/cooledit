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
	"testing"
)

func TestSelectAll(t *testing.T) {
	e := NewEditor(nil)
	
	// Test with empty buffer
	res := e.Apply(CmdSelectAll{}, 20)
	if res.Message != "Buffer is empty" {
		t.Errorf("expected 'Buffer is empty' message for empty buffer, got '%s'", res.Message)
	}
	
	// Add some text
	e.Apply(CmdInsertRune{Rune: 'H'}, 20)
	e.Apply(CmdInsertRune{Rune: 'e'}, 20)
	e.Apply(CmdInsertRune{Rune: 'l'}, 20)
	e.Apply(CmdInsertRune{Rune: 'l'}, 20)
	e.Apply(CmdInsertRune{Rune: 'o'}, 20)
	e.Apply(CmdInsertNewline{}, 20)
	e.Apply(CmdInsertRune{Rune: 'W'}, 20)
	e.Apply(CmdInsertRune{Rune: 'o'}, 20)
	e.Apply(CmdInsertRune{Rune: 'r'}, 20)
	e.Apply(CmdInsertRune{Rune: 'l'}, 20)
	e.Apply(CmdInsertRune{Rune: 'd'}, 20)
	
	// Now test select all
	res = e.Apply(CmdSelectAll{}, 20)
	if res.Message != "All text selected" {
		t.Errorf("expected 'All text selected' message, got '%s'", res.Message)
	}
	
	// Verify selection is active
	if !e.HasSelection() {
		t.Error("expected selection to be active after select all")
	}
	
	// Verify selection range
	sl, sc, el, ec := e.GetSelectionRange()
	if sl != 0 || sc != 0 {
		t.Errorf("expected selection to start at (0,0), got (%d,%d)", sl, sc)
	}
	
	// End should be at last line, last column
	lines := e.Lines()
	expectedLine := len(lines) - 1
	expectedCol := len(lines[expectedLine])
	if el != expectedLine || ec != expectedCol {
		t.Errorf("expected selection to end at (%d,%d), got (%d,%d)", expectedLine, expectedCol, el, ec)
	}
	
	// Verify cursor is at the end
	cl, cc := e.Cursor()
	if cl != expectedLine || cc != expectedCol {
		t.Errorf("expected cursor at (%d,%d), got (%d,%d)", expectedLine, expectedCol, cl, cc)
	}
}

func TestSelectAllThenCopy(t *testing.T) {
	clip := &mockClipboard{}
	e := NewEditor(clip)
	
	// Add some text
	e.Apply(CmdInsertRune{Rune: 'T'}, 20)
	e.Apply(CmdInsertRune{Rune: 'e'}, 20)
	e.Apply(CmdInsertRune{Rune: 's'}, 20)
	e.Apply(CmdInsertRune{Rune: 't'}, 20)
	
	// Select all
	e.Apply(CmdSelectAll{}, 20)
	
	// Copy
	res := e.Apply(CmdCopy{}, 20)
	if res.Message != "Selection copied" {
		t.Errorf("expected 'Selection copied' message, got '%s'", res.Message)
	}
	
	// Verify clipboard content
	if clip.text != "Test" {
		t.Errorf("expected clipboard to contain 'Test', got '%s'", clip.text)
	}
}

func TestSelectAllThenDelete(t *testing.T) {
	e := NewEditor(nil)
	
	// Add some text
	e.Apply(CmdInsertRune{Rune: 'A'}, 20)
	e.Apply(CmdInsertRune{Rune: 'B'}, 20)
	e.Apply(CmdInsertRune{Rune: 'C'}, 20)
	
	// Select all
	e.Apply(CmdSelectAll{}, 20)
	
	// Delete by typing a character (should replace selection)
	e.Apply(CmdInsertRune{Rune: 'X'}, 20)
	
	// Verify only 'X' remains
	lines := e.Lines()
	if len(lines) != 1 {
		t.Errorf("expected 1 line after delete, got %d", len(lines))
	}
	
	text := string(lines[0])
	if text != "X" {
		t.Errorf("expected text to be 'X', got '%s'", text)
	}
}
