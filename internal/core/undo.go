package core

type Action interface {
	Apply(e *Editor)
	Undo(e *Editor)
}

type UndoStack struct {
	actions  []Action
	ptr      int // points to the next slot to write (len(actions) if at end)
	savedPtr int // the value of ptr when the file was last saved
}

func NewUndoStack() *UndoStack {
	return &UndoStack{
		actions:  make([]Action, 0),
		ptr:      0,
		savedPtr: 0, // initially saved (empty buffer matches empty file)??
		// Actually new buffers are usually empty.
		// If loaded from file, savedPtr should be 0 (if no edits yet).
	}
}

func (u *UndoStack) Push(a Action) {
	if u.ptr < len(u.actions) {
		u.actions = u.actions[:u.ptr]
		// If we truncated the history and savedPtr was in the truncated part,
		// we effectively lost the saved state (it's in an unreachable future).
		// But if savedPtr < ptr, it remains valid.
		if u.savedPtr > u.ptr {
			u.savedPtr = -1 // No longer matchable
		}
	}
	u.actions = append(u.actions, a)
	u.ptr++
}

func (u *UndoStack) Undo(e *Editor) bool {
	if u.ptr == 0 {
		return false
	}
	u.ptr--
	action := u.actions[u.ptr]
	action.Undo(e)
	return true
}

func (u *UndoStack) Redo(e *Editor) bool {
	if u.ptr >= len(u.actions) {
		return false
	}
	action := u.actions[u.ptr]
	action.Apply(e)
	u.ptr++
	return true
}

func (u *UndoStack) MarkSaved() {
	u.savedPtr = u.ptr
}

func (u *UndoStack) IsSaved() bool {
	return u.ptr == u.savedPtr
}

func (u *UndoStack) CanUndo() bool {
	return u.ptr > 0
}

func (u *UndoStack) CanRedo() bool {
	return u.ptr < len(u.actions)
}

// Action definitions

type InsertRuneAction struct {
	Rune rune
	Line int
	Col  int
}

func (a *InsertRuneAction) Apply(e *Editor) {
	e.buf.SetCursor(a.Line, a.Col)
	e.buf.InsertRune(a.Rune)
}

func (a *InsertRuneAction) Undo(e *Editor) {
	e.buf.SetCursor(a.Line, a.Col+1)
	e.buf.Backspace()
}

type InsertNewlineAction struct {
	Line   int
	Col    int
	Indent string // Leading whitespace to insert on new line
}

func (a *InsertNewlineAction) Apply(e *Editor) {
	e.buf.SetCursor(a.Line, a.Col)
	e.buf.InsertNewline()
	// Insert indentation if specified
	if a.Indent != "" {
		for _, r := range a.Indent {
			e.buf.InsertRune(r)
		}
	}
}

func (a *InsertNewlineAction) Undo(e *Editor) {
	// Delete indentation first if it was added
	if a.Indent != "" {
		e.buf.SetCursor(a.Line+1, len(a.Indent))
		for range a.Indent {
			e.buf.Backspace()
		}
	}
	// Then delete the newline
	e.buf.SetCursor(a.Line+1, 0)
	e.buf.Backspace()
}

type BackspaceAction struct {
	DeletedRune rune
	Line        int // Position BEFORE backspace
	Col         int // Position BEFORE backspace
	IsMerge     bool
	MergeCol    int // Length of previous line before merge (if IsMerge)
}

func (a *BackspaceAction) Apply(e *Editor) {
	e.buf.SetCursor(a.Line, a.Col)
	e.buf.Backspace()
}

func (a *BackspaceAction) Undo(e *Editor) {
	if a.IsMerge {
		// Undo merge = Insert Newline at the join point
		// The join point is at (Line-1, MergeCol)
		e.buf.SetCursor(a.Line-1, a.MergeCol)
		e.buf.InsertNewline()
	} else {
		// Undo char deletion = InsertRune
		// Deleted char was at Col-1
		e.buf.SetCursor(a.Line, a.Col-1)
		e.buf.InsertRune(a.DeletedRune)
	}
}

type CutLineAction struct {
	Line       int
	Runes      []rune
	CursorLine int
	CursorCol  int
}

func (a *CutLineAction) Apply(e *Editor) {
	e.buf.DeleteLine(a.Line)
}

func (a *CutLineAction) Undo(e *Editor) {
	e.buf.InsertLine(a.Line, a.Runes)
	e.buf.SetCursor(a.CursorLine, a.CursorCol)
}

type ReplaceLinesAction struct {
	StartLine  int
	OldLines   [][]rune
	NewLines   [][]rune
	BeforeLine int
	BeforeCol  int
	AfterLine  int
	AfterCol   int
}

func (a *ReplaceLinesAction) Apply(e *Editor) {
	// Remove old
	for i := 0; i < len(a.OldLines); i++ {
		e.buf.DeleteLine(a.StartLine)
	}
	// Insert new
	for i := len(a.NewLines) - 1; i >= 0; i-- {
		e.buf.InsertLine(a.StartLine, a.NewLines[i])
	}
	e.buf.SetCursor(a.AfterLine, a.AfterCol)
}

func (a *ReplaceLinesAction) Undo(e *Editor) {
	// Remove new
	for i := 0; i < len(a.NewLines); i++ {
		e.buf.DeleteLine(a.StartLine)
	}
	// Insert old
	for i := len(a.OldLines) - 1; i >= 0; i-- {
		e.buf.InsertLine(a.StartLine, a.OldLines[i])
	}
	e.buf.SetCursor(a.BeforeLine, a.BeforeCol)
}

type DeleteSelectionAction struct {
	StartLine, StartCol int
	EndLine, EndCol     int
	DeletedText         string
}

func (a *DeleteSelectionAction) Apply(e *Editor) {
	e.buf.DeleteRange(a.StartLine, a.StartCol, a.EndLine, a.EndCol)
}

func (a *DeleteSelectionAction) Undo(e *Editor) {
	// Restore logic is tricky with ranges.
	// For simplicity, we can use ReplaceLines logic or re-implement insert logic.
	// Actually, DeleteRange behaves like ReplaceLines where NewLines is empty/merged.
	// Let's reuse Paste/Insert logic but localized?
	// Or we can assume we just insert the DeletedText at Start.

	// We need to parse DeletedText back into lines.
	var lines [][]rune
	var current []rune
	for _, r := range a.DeletedText {
		if r == '\n' {
			lines = append(lines, current)
			current = nil
		} else {
			current = append(current, r)
		}
	}
	lines = append(lines, current)

	// Now insert 'lines' at StartLine, StartCol.
	// This mirrors the Paste logic but as an Undo step.
	// But Buffer interface doesn't have InsertRange.
	// We can use ReplaceLinesAction logic here manually or add InsertRange to Buffer.
	// Or we can abuse InsertLine / DeleteLine.

	// Simplest: We know where we are. We can simulate insertion.
	// But we need to update e.buf directly.

	// Let's implement a manual "InsertText" here for Undo.
	// 1. Split line at StartCol
	prefix := e.buf.Lines()[a.StartLine][:a.StartCol]
	suffix := e.buf.Lines()[a.StartLine][a.StartCol:]

	var newBlock [][]rune
	newBlock = append(newBlock, append(append([]rune{}, prefix...), lines[0]...))
	for i := 1; i < len(lines)-1; i++ {
		newBlock = append(newBlock, lines[i])
	}
	lastIdx := len(lines) - 1
	if lastIdx > 0 {
		finalLine := append(lines[lastIdx], suffix...)
		newBlock = append(newBlock, finalLine)
	} else {
		// Single line restoration
		newBlock[0] = append(newBlock[0], suffix...)
	}

	// Replace StartLine with newBlock
	e.buf.DeleteLine(a.StartLine)
	for i := len(newBlock) - 1; i >= 0; i-- {
		e.buf.InsertLine(a.StartLine, newBlock[i])
	}

	// Restore cursor selection if we want? Usually Undo restores cursor but not selection state unless tracked.
	e.buf.SetCursor(a.EndLine, a.EndCol) // Or Start? Usually Start.
}

type CompositeAction struct {
	Actions []Action
}

func (a *CompositeAction) Apply(e *Editor) {
	for _, action := range a.Actions {
		action.Apply(e)
	}
}

func (a *CompositeAction) Undo(e *Editor) {
	for i := len(a.Actions) - 1; i >= 0; i-- {
		a.Actions[i].Undo(e)
	}
}
