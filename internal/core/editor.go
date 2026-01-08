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
}

type Result struct {
	Quit    bool
	Message string
}

func NewEditor(cb Clipboard) *Editor {
	return &Editor{
		buf: buffer.NewLineBuffer(),
		file: FileState{
			BaseName: "[No Name]",
			EOL:      "\n",
			Encoding: "UTF-8",
		},
		undo:      NewUndoStack(),
		clipboard: cb,
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
	
	case CmdCopy:
		line, _ := e.buf.Cursor()
		content := string(e.buf.Lines()[line])
		if e.clipboard != nil {
			if err := e.clipboard.Set(content); err != nil {
				return Result{Message: "Copy failed: " + err.Error()}
			}
		}
		return Result{Message: "Line copied"}

	case CmdCut:
		line, col := e.buf.Cursor()
		content := e.buf.Lines()[line]
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

		// For now, simpler multi-line paste:
		// We use ReplaceLinesAction to make it one undo block
		line, col := e.buf.Cursor()
		lines := e.buf.Lines()
		
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

		if len(newLines) == 1 {
			// Single line paste at cursor
			for _, r := range newLines[0] {
				e.Apply(CmdInsertRune{Rune: r}, viewHeight)
			}
		} else {
			// Multi-line paste. 
			// We'll replace the current line with (prefix + first line) ... (last line + suffix)
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

			action := &ReplaceLinesAction{
				StartLine:  line,
				OldLines:   [][]rune{lines[line]},
				NewLines:   inserted,
				BeforeLine: line,
				BeforeCol:  col,
				AfterLine:  line + len(newLines) - 1,
				AfterCol:   len(newLines[lastIdx]),
			}
			e.undo.Push(action)
			action.Apply(e)
		}
		return Result{Message: "Pasted"}

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