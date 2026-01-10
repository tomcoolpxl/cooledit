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
	"fmt"
	"path/filepath"

	"cooledit/internal/config"
	"cooledit/internal/core/buffer"
	"cooledit/internal/fileio"
)

type Viewport struct {
	TopLine int
	LeftCol int
}

type FileState struct {
	Path     string
	BaseName string
	EOL      string
	Encoding string
}

type Clipboard interface {
	Get() (string, error)
	Set(text string) error
}

type Editor struct {
	buf       buffer.Buffer
	vp        Viewport
	file      FileState
	undo      *UndoStack
	search    SearchState
	clipboard Clipboard

	selectionActive bool
	selectionAnchor struct{ Line, Col int }

	TabWidth       int // Number of spaces per tab (default: 4)
	bracketMatcher *BracketMatcher
}

type Result struct {
	Quit    bool
	Message string
}

func (e *Editor) HasSelection() bool {
	return e.selectionActive
}

func (e *Editor) ClearSelection() {
	e.selectionActive = false
}

func (e *Editor) SetSelection(line, col, length int) {
	e.selectionActive = true
	e.selectionAnchor.Line = line
	e.selectionAnchor.Col = col + length
	e.buf.SetCursor(line, col)
}

func (e *Editor) GetSelectionRange() (sl, sc, el, ec int) {
	if !e.selectionActive {
		l, c := e.buf.Cursor()
		return l, c, l, c
	}
	cl, cc := e.buf.Cursor()
	al, ac := e.selectionAnchor.Line, e.selectionAnchor.Col

	if al < cl || (al == cl && ac < cc) {
		return al, ac, cl, cc
	}
	return cl, cc, al, ac
}

func NewEditor(cb Clipboard) *Editor {
	return &Editor{
		buf: buffer.NewLineBuffer(),
		file: FileState{
			BaseName: "[No Name]",
			EOL:      "\n",
			Encoding: "UTF-8",
		},
		undo:           NewUndoStack(),
		clipboard:      cb,
		bracketMatcher: NewBracketMatcher(),
	}
}

func (e *Editor) LoadFile(fd *fileio.FileData) {
	e.buf = buffer.NewLineBufferFromLines(fd.Lines)
	e.vp = Viewport{}
	e.file = FileState{
		Path:     fd.Path,
		BaseName: fd.BaseName,
		EOL:      fd.EOL,
		Encoding: fd.Encoding,
	}
	e.undo = NewUndoStack()
	e.undo.MarkSaved()

	// Clear search session when loading a new file
	// This prevents stale search results from the previous file
	e.EndSearchSession()
}

// SetNewFile sets up the editor for a new file with the given path.
// The file doesn't exist yet and will be created on first save.
func (e *Editor) SetNewFile(path string) {
	e.buf = buffer.NewLineBuffer()
	e.vp = Viewport{}
	e.file = FileState{
		Path:     path,
		BaseName: filepath.Base(path),
		EOL:      "\n",
		Encoding: "UTF-8",
	}
	e.undo = NewUndoStack()
	// Don't mark as saved since the file doesn't exist yet

	// Clear search session when creating a new file
	// This prevents stale search results from any previous file
	e.EndSearchSession()
}

func (e *Editor) deleteSelection() Action {
	if !e.selectionActive {
		return nil
	}
	sl, sc, el, ec := e.GetSelectionRange()
	text := e.buf.RangeText(sl, sc, el, ec)

	action := &DeleteSelectionAction{
		StartLine:   sl,
		StartCol:    sc,
		EndLine:     el,
		EndCol:      ec,
		DeletedText: text,
	}
	// Note: We don't apply here, caller applies or adds to composite.
	// Actually, standard Apply pattern is: create, Push, Apply.
	// But here we want to return it to be part of Composite if needed.
	return action
}

func (e *Editor) Apply(cmd Command, viewHeight int) Result {
	switch c := cmd.(type) {

	case CmdSave:
		if e.file.Path == "" {
			return Result{Message: "No file name. Use Save As."}
		}
		if !e.Modified() {
			return Result{Message: "No changes to save"}
		}
		if err := fileio.Save(e.file.Path, e.buf.Lines(), e.file.EOL, e.file.Encoding); err != nil {
			return Result{Message: "Save failed: " + err.Error()}
		}
		e.undo.MarkSaved()
		return Result{Message: "File saved"}

	case CmdSaveAs:
		if err := fileio.Save(c.Path, e.buf.Lines(), e.file.EOL, e.file.Encoding); err != nil {
			return Result{Message: "Save failed: " + err.Error()}
		}
		e.file.Path = c.Path
		e.file.BaseName = filepath.Base(c.Path)
		e.undo.MarkSaved()
		return Result{Message: "File saved"}

	case CmdUndo:
		if e.undo.Undo(e) {
			return Result{Message: "Undo"}
		}
		return Result{Message: "Already at oldest change"}

	case CmdRedo:
		if e.undo.Redo(e) {
			return Result{Message: "Redo"}
		}
		return Result{Message: "Already at newest change"}

	case CmdFind:
		e.search.SetQuery(c.Query)
		line, col := e.buf.Cursor()
		fl, fc, found := Search(e.buf.Lines(), c.Query, line, col, SearchForward, e.search.CaseSensitive, e.search.WholeWord)
		if found {
			e.SetSelection(fl, fc, len(c.Query))
			return Result{Message: "Found: " + c.Query}
		}
		return Result{Message: "Not found: " + c.Query}

	case CmdFindNext:
		if e.search.LastQuery == "" {
			return Result{Message: "No previous search"}
		}
		line, col := e.buf.Cursor()
		// Start search after current match to avoid overlapping matches
		fl, fc, found := Search(e.buf.Lines(), e.search.LastQuery, line, col+len(e.search.LastQuery), SearchForward, e.search.CaseSensitive, e.search.WholeWord)
		if found {
			e.SetSelection(fl, fc, len(e.search.LastQuery))
			return Result{Message: "Found next: " + e.search.LastQuery}
		}
		return Result{Message: "Not found (next): " + e.search.LastQuery}

	case CmdFindPrev:
		if e.search.LastQuery == "" {
			return Result{Message: "No previous search"}
		}
		line, col := e.buf.Cursor()
		fl, fc, found := Search(e.buf.Lines(), e.search.LastQuery, line, col, SearchBackward, e.search.CaseSensitive, e.search.WholeWord)
		if found {
			e.SetSelection(fl, fc, len(e.search.LastQuery))
			return Result{Message: "Found prev: " + e.search.LastQuery}
		}
		return Result{Message: "Not found (prev): " + e.search.LastQuery}

	case CmdCopy:
		var content string
		if e.selectionActive {
			sl, sc, el, ec := e.GetSelectionRange()
			content = e.buf.RangeText(sl, sc, el, ec)
		} else {
			line, _ := e.buf.Cursor()
			lines := e.buf.Lines()
			if line < 0 || line >= len(lines) {
				return Result{Message: "Copy failed: invalid cursor position"}
			}
			content = string(lines[line])
		}

		if e.clipboard != nil {
			if err := e.clipboard.Set(content); err != nil {
				return Result{Message: "Copy failed: " + err.Error()}
			}
		}
		if e.selectionActive {
			return Result{Message: "Selection copied"}
		}
		return Result{Message: "Line copied"}

	case CmdCut:
		if e.selectionActive {
			sl, sc, el, ec := e.GetSelectionRange()
			content := e.buf.RangeText(sl, sc, el, ec)

			if e.clipboard != nil {
				if err := e.clipboard.Set(content); err != nil {
					return Result{Message: "Cut failed: " + err.Error()}
				}
			}

			action := &DeleteSelectionAction{
				StartLine:   sl,
				StartCol:    sc,
				EndLine:     el,
				EndCol:      ec,
				DeletedText: content,
			}
			e.undo.Push(action)
			action.Apply(e)
			e.ClearSelection()
			return Result{Message: "Selection cut"}
		}

		line, col := e.buf.Cursor()
		lines := e.buf.Lines()
		if line < 0 || line >= len(lines) {
			return Result{Message: "Cut failed: invalid cursor position"}
		}
		content := lines[line]
		if e.clipboard != nil {
			if err := e.clipboard.Set(string(content)); err != nil {
				return Result{Message: "Cut failed: " + err.Error()}
			}
		}
		action := &CutLineAction{
			Line:       line,
			Runes:      content,
			CursorLine: line,
			CursorCol:  col,
		}
		e.undo.Push(action)
		action.Apply(e)
		return Result{Message: "Line cut"}

	case CmdPaste:
		text := c.Text
		if text == "" && e.clipboard != nil {
			var err error
			text, err = e.clipboard.Get()
			if err != nil {
				return Result{Message: "Paste failed: " + err.Error()}
			}
		}
		if text == "" {
			return Result{Message: "Clipboard empty"}
		}

		// Handle selection replacement
		var delAction Action
		if e.selectionActive {
			delAction = e.deleteSelection()
			e.ClearSelection()
			delAction.Apply(e)
		}

		// For now, simpler multi-line paste:
		// We use ReplaceLinesAction to make it one undo block
		line, col := e.buf.Cursor()
		lines := e.buf.Lines()

		// Validate cursor position before array access
		if line < 0 || line >= len(lines) {
			return Result{Message: "Paste failed: invalid cursor position"}
		}
		if col < 0 {
			col = 0
		}
		if col > len(lines[line]) {
			col = len(lines[line])
		}

		var newLines [][]rune
		// Split text into lines
		var current []rune
		for _, r := range text {
			if r == '\n' {
				newLines = append(newLines, current)
				current = nil
			} else if r == '\r' {
				continue
			} else {
				current = append(current, r)
			}
		}
		newLines = append(newLines, current)

		var pasteAction Action

		if len(newLines) == 1 {
			// Single line paste at cursor
			// We'll reuse the ReplaceLines logic below for single line too
		}

		// General logic using ReplaceLinesAction (works for single line too)
		prefix := append([]rune{}, lines[line][:col]...)
		suffix := append([]rune{}, lines[line][col:]...)

		var inserted [][]rune
		inserted = append(inserted, append(prefix, newLines[0]...))
		for i := 1; i < len(newLines)-1; i++ {
			inserted = append(inserted, newLines[i])
		}
		lastIdx := len(newLines) - 1
		finalLine := append(newLines[lastIdx], suffix...)
		inserted = append(inserted, finalLine)

		pasteAction = &ReplaceLinesAction{
			StartLine:  line,
			OldLines:   [][]rune{lines[line]},
			NewLines:   inserted,
			BeforeLine: line,
			BeforeCol:  col,
			AfterLine:  line + len(newLines) - 1,
			AfterCol:   len(newLines[lastIdx]), // Start of pasted last line + len
		}

		if len(newLines) == 1 {
			pasteAction.(*ReplaceLinesAction).AfterCol = col + len(newLines[0])
		} else {
			pasteAction.(*ReplaceLinesAction).AfterCol = len(newLines[lastIdx])
		}

		pasteAction.Apply(e)

		if delAction != nil {
			e.undo.Push(&CompositeAction{Actions: []Action{delAction, pasteAction}})
		} else {
			e.undo.Push(pasteAction)
		}

		return Result{Message: "Pasted"}

	case CmdGoToLine:
		target := c.Line - 1 // convert 1-based to 0-based
		e.ClearSelection()
		e.buf.SetCursor(target, 0)
		return Result{}

	case CmdReplace:
		// Replace the current match (assumes cursor is on a match)
		// and find the next match
		if c.Find == "" {
			return Result{Message: "Replace: empty search term"}
		}

		// Update search state to match find term
		e.search.SetQuery(c.Find)

		line, col := e.buf.Cursor()
		lines := e.buf.Lines()

		// Verify we're on a match
		if line >= len(lines) {
			return Result{Message: "No more matches"}
		}

		lineText := string(lines[line])
		if col >= len(lineText) || col+len(c.Find) > len(lineText) {
			// Not on a match, try to find next
			return e.Apply(CmdFindNext{}, 0)
		}

		if lineText[col:col+len(c.Find)] != c.Find {
			// Not on a match, try to find next
			return e.Apply(CmdFindNext{}, 0)
		}

		// Replace the match
		before := lineText[:col]
		after := lineText[col+len(c.Find):]
		newLine := []rune(before + c.Replace + after)

		action := &ReplaceLinesAction{
			StartLine:  line,
			OldLines:   [][]rune{lines[line]},
			NewLines:   [][]rune{newLine},
			BeforeLine: line,
			BeforeCol:  col,
			AfterLine:  line,
			AfterCol:   col + len(c.Replace),
		}

		e.undo.Push(action)
		action.Apply(e)

		// Find next match
		fl, fc, found := Search(e.buf.Lines(), c.Find, line, col+len(c.Replace), SearchForward, e.search.CaseSensitive, e.search.WholeWord)
		if found {
			e.buf.SetCursor(fl, fc)
			return Result{}
		}
		return Result{Message: "Replaced (no more matches)"}

	case CmdReplaceAll:
		// Replace all matches from the beginning of the file
		if c.Find == "" {
			return Result{Message: "Replace: empty search term"}
		}

		// Update search state to match find term
		e.search.SetQuery(c.Find)

		// Collect all replacement actions first
		var actions []Action
		count := 0
		// Start from beginning of file
		line, col := 0, 0

		// Keep finding matches and building actions (without applying yet)
		lines := e.buf.Lines()
		for {
			fl, fc, found := Search(lines, c.Find, line, col, SearchForward, e.search.CaseSensitive, e.search.WholeWord)
			if !found {
				break
			}

			// Create action for this replacement
			lineText := string(lines[fl])
			before := lineText[:fc]
			after := lineText[fc+len(c.Find):]
			newLine := []rune(before + c.Replace + after)

			action := &ReplaceLinesAction{
				StartLine:  fl,
				OldLines:   [][]rune{lines[fl]},
				NewLines:   [][]rune{newLine},
				BeforeLine: fl,
				BeforeCol:  fc,
				AfterLine:  fl,
				AfterCol:   fc + len(c.Replace),
			}

			actions = append(actions, action)
			count++

			// Update lines array with the replacement for next search
			lines[fl] = newLine
			line = fl
			col = fc + len(c.Replace)
		}

		if count == 0 {
			return Result{Message: "No matches found"}
		}

		// Now apply all actions as a single composite operation
		if count == 1 {
			// Single replacement
			e.undo.Push(actions[0])
			actions[0].Apply(e)
		} else {
			// Multiple replacements - create and apply composite
			composite := &CompositeAction{Actions: actions}
			e.undo.Push(composite)
			composite.Apply(e)
		}

		if count == 1 {
			return Result{Message: "Replaced 1 occurrence - Ctrl+Z to undo"}
		}
		return Result{Message: fmt.Sprintf("Replaced %d occurrences - Ctrl+Z to undo", count)}

	case CmdClick:
		e.buf.SetCursor(c.Line, c.Col)

	case CmdJumpToMatchingBracket:
		line, col := e.buf.Cursor()
		lines := e.buf.Lines()

		if line < 0 || line >= len(lines) || col < 0 || col >= len(lines[line]) {
			return Result{Message: "Not on a bracket"}
		}

		ch := lines[line][col]
		if !e.bracketMatcher.IsBracket(ch) {
			return Result{Message: "Not on a bracket"}
		}

		// Find matching bracket (using nil skipFunc for simplicity)
		matchLine, matchCol, found := e.bracketMatcher.FindMatch(lines, line, col, nil)
		if !found {
			return Result{Message: "No matching bracket found"}
		}

		e.ClearSelection()
		e.buf.SetCursor(matchLine, matchCol)
		return Result{}

	case CmdInsertRune:
		if e.selectionActive {
			delAction := e.deleteSelection()
			e.ClearSelection()

			// We need to Apply delete first to update cursor for insert
			// But we want atomic Undo.
			// So we apply delete, get new cursor, create insert action.
			delAction.Apply(e)

			line, col := e.buf.Cursor()
			insAction := &InsertRuneAction{
				Rune: c.Rune,
				Line: line,
				Col:  col,
			}
			insAction.Apply(e)

			e.undo.Push(&CompositeAction{Actions: []Action{delAction, insAction}})
			return Result{}
		}

		line, col := e.buf.Cursor()
		action := &InsertRuneAction{
			Rune: c.Rune,
			Line: line,
			Col:  col,
		}
		e.undo.Push(action)
		action.Apply(e)

	case CmdReplaceRune:
		// Replace mode: overwrite character at cursor
		if e.selectionActive {
			// If there's a selection, delete it first (like insert mode)
			delAction := e.deleteSelection()
			e.ClearSelection()
			delAction.Apply(e)

			line, col := e.buf.Cursor()
			insAction := &InsertRuneAction{
				Rune: c.Rune,
				Line: line,
				Col:  col,
			}
			insAction.Apply(e)

			e.undo.Push(&CompositeAction{Actions: []Action{delAction, insAction}})
			return Result{}
		}

		line, col := e.buf.Cursor()
		lines := e.buf.Lines()

		// If at end of line or on empty line, just insert
		if line >= len(lines) || col >= len(lines[line]) {
			action := &InsertRuneAction{
				Rune: c.Rune,
				Line: line,
				Col:  col,
			}
			e.undo.Push(action)
			action.Apply(e)
		} else {
			// Replace: capture old char, use backspace to delete, then insert
			oldRune := lines[line][col]

			// Create a backspace action to delete current char
			// First move cursor forward, then backspace
			e.buf.SetCursor(line, col+1)
			backAction := &BackspaceAction{
				DeletedRune: oldRune,
				Line:        line,
				Col:         col + 1,
				IsMerge:     false,
			}
			backAction.Apply(e)

			// Now insert the new character (cursor is at col after backspace)
			insAction := &InsertRuneAction{
				Rune: c.Rune,
				Line: line,
				Col:  col,
			}
			insAction.Apply(e)

			e.undo.Push(&CompositeAction{Actions: []Action{backAction, insAction}})
		}

	case CmdInsertNewline:
		// Detect leading whitespace on current line to copy to new line (nano-style)
		line, col := e.buf.Cursor()
		lines := e.buf.Lines()
		indent := ""
		if line < len(lines) {
			lineText := lines[line]
			// Extract leading whitespace (spaces and tabs)
			for _, r := range lineText {
				if r == ' ' || r == '\t' {
					indent += string(r)
				} else {
					break
				}
			}
		}

		if e.selectionActive {
			delAction := e.deleteSelection()
			e.ClearSelection()
			delAction.Apply(e)

			line, col = e.buf.Cursor()
			insAction := &InsertNewlineAction{
				Line:   line,
				Col:    col,
				Indent: indent,
			}
			insAction.Apply(e)

			e.undo.Push(&CompositeAction{Actions: []Action{delAction, insAction}})
			return Result{}
		}

		action := &InsertNewlineAction{
			Line:   line,
			Col:    col,
			Indent: indent,
		}
		e.undo.Push(action)
		action.Apply(e)

	case CmdBackspace:
		if e.selectionActive {
			delAction := e.deleteSelection()
			e.ClearSelection()
			delAction.Apply(e)
			e.undo.Push(delAction)
			return Result{}
		}

		line, col := e.buf.Cursor()

		// Simple backspace: always delete one character at a time
		// No smart indentation - just delete whatever is before the cursor
		var action *BackspaceAction

		if col > 0 {
			r := e.buf.RuneAt(line, col-1)
			action = &BackspaceAction{
				DeletedRune: r,
				Line:        line,
				Col:         col,
				IsMerge:     false,
			}
		} else if line > 0 {
			prevLen := e.buf.LineLen(line - 1)
			action = &BackspaceAction{
				Line:     line,
				Col:      0,
				IsMerge:  true,
				MergeCol: prevLen,
			}
		} else {
			return Result{}
		}

		e.undo.Push(action)
		action.Apply(e)

	case CmdTab:
		// Insert spaces to next tab stop
		if e.selectionActive {
			delAction := e.deleteSelection()
			e.ClearSelection()
			delAction.Apply(e)
		}

		line, col := e.buf.Cursor()
		tabWidth := e.TabWidth
		if tabWidth <= 0 {
			tabWidth = config.DefaultTabWidth
		}

		// Calculate spaces to next tab stop
		spacesToInsert := tabWidth - (col % tabWidth)

		// Create composite action for inserting multiple spaces
		actions := make([]Action, spacesToInsert)
		for i := 0; i < spacesToInsert; i++ {
			actions[i] = &InsertRuneAction{
				Rune: ' ',
				Line: line,
				Col:  col + i,
			}
		}

		composite := &CompositeAction{Actions: actions}
		e.undo.Push(composite)
		composite.Apply(e)

	case CmdInsertLiteralTab:
		// Insert a raw tab character
		if e.selectionActive {
			delAction := e.deleteSelection()
			e.ClearSelection()
			delAction.Apply(e)
		}

		line, col := e.buf.Cursor()
		action := &InsertRuneAction{
			Rune: '\t',
			Line: line,
			Col:  col,
		}
		e.undo.Push(action)
		action.Apply(e)

	case CmdDelete:
		if e.selectionActive {
			delAction := e.deleteSelection()
			e.ClearSelection()
			delAction.Apply(e)
			e.undo.Push(delAction)
			return Result{}
		}

		line, col := e.buf.Cursor()

		// Delete character at cursor (forward delete)
		if col < e.buf.LineLen(line) {
			// Delete char at current position
			r := e.buf.RuneAt(line, col)
			action := &BackspaceAction{
				DeletedRune: r,
				Line:        line,
				Col:         col + 1, // Pretend cursor was after the char
				IsMerge:     false,
			}
			e.undo.Push(action)
			action.Apply(e)
		} else if line < len(e.buf.Lines())-1 {
			// At end of line, merge with next line (like delete newline)
			action := &BackspaceAction{
				Line:     line + 1,
				Col:      0,
				IsMerge:  true,
				MergeCol: col,
			}
			e.undo.Push(action)
			action.Apply(e)
		}

	case CmdMoveLeft:
		e.handleMove(c.Select, func() { e.buf.MoveLeft() })
	case CmdMoveRight:
		e.handleMove(c.Select, func() { e.buf.MoveRight() })
	case CmdMoveWordLeft:
		e.handleMove(c.Select, func() { e.moveWordLeft() })
	case CmdMoveWordRight:
		e.handleMove(c.Select, func() { e.moveWordRight() })
	case CmdMoveUp:
		e.handleMove(c.Select, func() { e.buf.MoveUp() })
	case CmdMoveDown:
		e.handleMove(c.Select, func() { e.buf.MoveDown() })
	case CmdMoveHome:
		e.handleMove(c.Select, func() { e.buf.MoveHome() })
	case CmdMoveEnd:
		e.handleMove(c.Select, func() { e.buf.MoveEnd() })

	case CmdPageUp:
		e.handleMove(c.Select, func() {
			for i := 0; i < viewHeight; i++ {
				e.buf.MoveUp()
			}
		})
	case CmdPageDown:
		e.handleMove(c.Select, func() {
			for i := 0; i < viewHeight; i++ {
				e.buf.MoveDown()
			}
		})

	case CmdFileStart:
		e.handleMove(c.Select, func() {
			for {
				prev, _ := e.buf.Cursor()
				e.buf.MoveUp()
				cur, _ := e.buf.Cursor()
				if cur == prev {
					break
				}
			}
			e.buf.MoveHome()
		})

	case CmdFileEnd:
		e.handleMove(c.Select, func() {
			for {
				prev, _ := e.buf.Cursor()
				e.buf.MoveDown()
				cur, _ := e.buf.Cursor()
				if cur == prev {
					break
				}
			}
			e.buf.MoveEnd()
		})
	}

	return Result{}
}

func (e *Editor) handleMove(selectMode bool, moveFunc func()) {
	if selectMode {
		if !e.selectionActive {
			e.selectionActive = true
			e.selectionAnchor.Line, e.selectionAnchor.Col = e.buf.Cursor()
		}
	} else {
		e.selectionActive = false
	}
	moveFunc()
}

func (e *Editor) Lines() [][]rune    { return e.buf.Lines() }
func (e *Editor) Cursor() (int, int) { return e.buf.Cursor() }
func (e *Editor) Viewport() Viewport { return e.vp }
func (e *Editor) File() FileState    { return e.file }

func (e *Editor) Modified() bool {
	return !e.undo.IsSaved()
}

// SearchState returns a pointer to the editor's search state
func (e *Editor) SearchState() *SearchState {
	return &e.search
}

// ToggleCaseSensitivity toggles case-sensitive search mode
func (e *Editor) ToggleCaseSensitivity() {
	e.search.CaseSensitive = !e.search.CaseSensitive
	// Update active session if one exists and re-run search
	if e.search.Session != nil {
		e.search.Session.CaseSensitive = e.search.CaseSensitive
		e.search.Session.UpdateMatches(e.buf.Lines(), 1000)
	}
}

// ToggleWholeWord toggles whole word search mode
func (e *Editor) ToggleWholeWord() {
	e.search.WholeWord = !e.search.WholeWord
	// Update active session if one exists and re-run search
	if e.search.Session != nil {
		e.search.Session.WholeWord = e.search.WholeWord
		e.search.Session.UpdateMatches(e.buf.Lines(), 1000)
	}
}

func (e *Editor) EnsureVisible(w, h int) {
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	cy, cx := e.buf.Cursor()
	if cy < e.vp.TopLine {
		e.vp.TopLine = cy
	} else if cy >= e.vp.TopLine+h {
		e.vp.TopLine = cy - h + 1
	}
	if cx < e.vp.LeftCol {
		e.vp.LeftCol = cx
	} else if cx >= e.vp.LeftCol+w {
		e.vp.LeftCol = cx - w + 1
	}
}

// isWordChar returns true if the rune is considered part of a word
func isWordChar(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_'
}

// moveWordLeft moves cursor to the beginning of the previous word
func (e *Editor) moveWordLeft() {
	lines := e.buf.Lines()
	line, col := e.buf.Cursor()

	if len(lines) == 0 {
		return
	}

	// Move left one position first
	if col > 0 {
		col--
	} else if line > 0 {
		// Jump to end of previous line
		line--
		col = len(lines[line])
	} else {
		// Already at start of file
		return
	}

	// Skip whitespace
	for line >= 0 && col >= 0 {
		if col >= len(lines[line]) {
			col = len(lines[line]) - 1
		}
		if col < 0 {
			if line == 0 {
				col = 0
				break
			}
			line--
			col = len(lines[line]) - 1
			continue
		}
		if !isWordChar(lines[line][col]) && lines[line][col] != ' ' && lines[line][col] != '\t' {
			// Skip punctuation
			if col == 0 {
				break
			}
			col--
			continue
		}
		if lines[line][col] != ' ' && lines[line][col] != '\t' {
			break
		}
		col--
	}

	// Move to start of word
	for line >= 0 && col > 0 {
		if col-1 < 0 {
			break
		}
		if !isWordChar(lines[line][col-1]) {
			break
		}
		col--
	}

	if line < 0 {
		line = 0
	}
	if col < 0 {
		col = 0
	}

	e.buf.SetCursor(line, col)
}

// moveWordRight moves cursor to the beginning of the next word
func (e *Editor) moveWordRight() {
	lines := e.buf.Lines()
	line, col := e.buf.Cursor()

	if len(lines) == 0 {
		return
	}

	// Skip current word
	for line < len(lines) && col < len(lines[line]) {
		if !isWordChar(lines[line][col]) {
			break
		}
		col++
	}

	// Skip whitespace and punctuation
	for line < len(lines) {
		for col < len(lines[line]) {
			if isWordChar(lines[line][col]) {
				e.buf.SetCursor(line, col)
				return
			}
			col++
		}
		// Move to next line
		line++
		if line >= len(lines) {
			break
		}
		col = 0
	}

	// Reached end of file
	if line >= len(lines) {
		line = len(lines) - 1
	}
	if line >= 0 && col > len(lines[line]) {
		col = len(lines[line])
	}
	e.buf.SetCursor(line, col)
}

// StartSearchSession starts a new search session with the given query.
// This updates the matches and creates a new session.
func (e *Editor) StartSearchSession(query string) {
	if e.search.Session == nil {
		e.search.Session = NewSearchSession(query, e.search.CaseSensitive, e.search.WholeWord)
	} else {
		e.search.Session.Query = query
		e.search.Session.CaseSensitive = e.search.CaseSensitive
		e.search.Session.WholeWord = e.search.WholeWord
	}

	// Update matches (limit to 1000 for performance)
	e.search.Session.UpdateMatches(e.buf.Lines(), 1000)

	// Store the last query
	e.search.LastQuery = query
}

// EndSearchSession ends the current search session and cleans up.
// NOTE: This only clears the active search session (match data, etc.).
// Session-level preferences (CaseSensitive, WholeWord) are preserved
// and will be used for the next search.
func (e *Editor) EndSearchSession() {
	e.search.Session = nil
}

// HasSearchSession returns true if there's an active search session.
func (e *Editor) HasSearchSession() bool {
	return e.search.Session != nil
}

// GetSearchSession returns the current search session, or nil if none.
func (e *Editor) GetSearchSession() *SearchSession {
	return e.search.Session
}

// GetCaseSensitive returns the current case sensitivity setting.
func (e *Editor) GetCaseSensitive() bool {
	return e.search.CaseSensitive
}

// GetWholeWord returns the current whole word setting.
func (e *Editor) GetWholeWord() bool {
	return e.search.WholeWord
}
