package core

import (
	"path/filepath"

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

type Editor struct {
	buf    buffer.Buffer
	vp     Viewport
	file   FileState
	undo   *UndoStack
	search SearchState
}

type Result struct {
	Quit    bool
	Message string
}

func NewEditor() *Editor {
	return &Editor{
		buf: buffer.NewLineBuffer(),
		file: FileState{
			BaseName: "[No Name]",
			EOL:      "\n",
			Encoding: "UTF-8",
		},
		undo: NewUndoStack(),
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
		fl, fc, found := Search(e.buf.Lines(), c.Query, line, col, SearchForward)
		if found {
			e.buf.SetCursor(fl, fc)
			return Result{Message: "Found: " + c.Query}
		}
		return Result{Message: "Not found: " + c.Query}

	case CmdFindNext:
		if e.search.LastQuery == "" {
			return Result{Message: "No previous search"}
		}
		line, col := e.buf.Cursor()
		// Start search after current position
		fl, fc, found := Search(e.buf.Lines(), e.search.LastQuery, line, col+1, SearchForward)
		if found {
			e.buf.SetCursor(fl, fc)
			return Result{Message: "Found next: " + e.search.LastQuery}
		}
		return Result{Message: "Not found (next): " + e.search.LastQuery}

	case CmdFindPrev:
		if e.search.LastQuery == "" {
			return Result{Message: "No previous search"}
		}
		line, col := e.buf.Cursor()
		fl, fc, found := Search(e.buf.Lines(), e.search.LastQuery, line, col, SearchBackward)
		if found {
			e.buf.SetCursor(fl, fc)
			return Result{Message: "Found prev: " + e.search.LastQuery}
		}
		return Result{Message: "Not found (prev): " + e.search.LastQuery}
	
	case CmdClick:
		e.buf.SetCursor(c.Line, c.Col)

	case CmdInsertRune:
		line, col := e.buf.Cursor()
		action := &InsertRuneAction{
			Rune: c.Rune,
			Line: line,
			Col:  col,
		}
		e.undo.Push(action)
		action.Apply(e)

	case CmdInsertNewline:
		line, col := e.buf.Cursor()
		action := &InsertNewlineAction{
			Line: line,
			Col:  col,
		}
		e.undo.Push(action)
		action.Apply(e)

	case CmdBackspace:
		line, col := e.buf.Cursor()
		
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

	case CmdMoveLeft:
		e.buf.MoveLeft()
	case CmdMoveRight:
		e.buf.MoveRight()
	case CmdMoveUp:
		e.buf.MoveUp()
	case CmdMoveDown:
		e.buf.MoveDown()
	case CmdMoveHome:
		e.buf.MoveHome()
	case CmdMoveEnd:
		e.buf.MoveEnd()

	case CmdPageUp:
		for i := 0; i < viewHeight; i++ {
			e.buf.MoveUp()
		}
	case CmdPageDown:
		for i := 0; i < viewHeight; i++ {
			e.buf.MoveDown()
		}

	case CmdFileStart:
		for {
			prev, _ := e.buf.Cursor()
			e.buf.MoveUp()
			cur, _ := e.buf.Cursor()
			if cur == prev {
				break
			}
		}
		e.buf.MoveHome()

	case CmdFileEnd:
		for {
			prev, _ := e.buf.Cursor()
			e.buf.MoveDown()
			cur, _ := e.buf.Cursor()
			if cur == prev {
				break
			}
		}
		e.buf.MoveEnd()
	}

	return Result{}
}

func (e *Editor) Lines() [][]rune    { return e.buf.Lines() }
func (e *Editor) Cursor() (int, int) { return e.buf.Cursor() }
func (e *Editor) Viewport() Viewport { return e.vp }
func (e *Editor) File() FileState    { return e.file }

func (e *Editor) Modified() bool { 
	return !e.undo.IsSaved()
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