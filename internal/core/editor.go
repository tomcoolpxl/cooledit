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
	buf      buffer.Buffer
	vp       Viewport
	file     FileState
	modified bool
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
	}
}

func (e *Editor) LoadFile(fd *fileio.FileData) {
	e.buf = buffer.NewLineBufferFromLines(fd.Lines)
	e.vp = Viewport{}
	e.modified = false
	e.file = FileState{
		Path:     fd.Path,
		BaseName: fd.BaseName,
		EOL:      fd.EOL,
		Encoding: fd.Encoding,
	}
}

func (e *Editor) Apply(cmd Command, viewHeight int) Result {
	switch c := cmd.(type) {

	case CmdSave:
		if e.file.Path == "" {
			return Result{Message: "No file name. Use Save As."}
		}
		if !e.modified {
			return Result{Message: "No changes to save"}
		}
		if err := fileio.Save(e.file.Path, e.buf.Lines(), e.file.EOL, e.file.Encoding); err != nil {
			return Result{Message: "Save failed: " + err.Error()}
		}
		e.modified = false
		return Result{Message: "File saved"}

	case CmdSaveAs:
		// overwrite confirmation handled in UI
		if err := fileio.Save(c.Path, e.buf.Lines(), e.file.EOL, e.file.Encoding); err != nil {
			return Result{Message: "Save failed: " + err.Error()}
		}
		e.file.Path = c.Path
		e.file.BaseName = filepath.Base(c.Path)
		e.modified = false
		return Result{Message: "File saved"}

	case CmdInsertRune:
		e.modified = true
		e.buf.InsertRune(c.Rune)

	case CmdInsertNewline:
		e.modified = true
		e.buf.InsertNewline()

	case CmdBackspace:
		e.modified = true
		e.buf.Backspace()

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
func (e *Editor) Modified() bool     { return e.modified }
func (e *Editor) File() FileState    { return e.file }

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
