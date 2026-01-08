package core

import (
	"time"

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
	ReadOnly bool
}

type Editor struct {
	buf         buffer.Buffer
	vp          Viewport
	file        FileState
	modified    bool
	quitPending bool
	quitAt      time.Time
}

func NewEditor() *Editor {
	return &Editor{
		buf: buffer.NewLineBuffer(),
		vp:  Viewport{TopLine: 0, LeftCol: 0},
		file: FileState{
			Path:     "",
			BaseName: "[No Name]",
			EOL:      "\n",
			Encoding: "UTF-8",
			ReadOnly: false,
		},
	}
}

type Result struct {
	Quit    bool
	Message string
}

func (e *Editor) LoadFile(fd *fileio.FileData) {
	e.buf = buffer.NewLineBufferFromLines(fd.Lines)
	e.vp = Viewport{TopLine: 0, LeftCol: 0}
	e.modified = false

	e.file = FileState{
		Path:     fd.Path,
		BaseName: fd.BaseName,
		EOL:      fd.EOL,
		Encoding: fd.Encoding,
		ReadOnly: true, // important
	}
}

func (e *Editor) Apply(cmd Command, viewHeight int) Result {
	switch cmd.(type) {
	case CmdQuit:
		now := time.Now()
		if e.quitPending && now.Sub(e.quitAt) < 2*time.Second {
			return Result{Quit: true}
		}
		e.quitPending = true
		e.quitAt = now
		return Result{}
	}

	e.quitPending = false

	// Editing commands blocked in read-only mode
	if e.file.ReadOnly {
		switch cmd.(type) {
		case CmdInsertRune, CmdInsertNewline, CmdBackspace:
			return Result{Message: "Read-only: use Save As to enable editing"}
		}
	}

	switch c := cmd.(type) {
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

func (e *Editor) Lines() [][]rune {
	return e.buf.Lines()
}

func (e *Editor) Cursor() (int, int) {
	return e.buf.Cursor()
}

func (e *Editor) Viewport() Viewport {
	return e.vp
}

func (e *Editor) Modified() bool {
	return e.modified
}

func (e *Editor) File() FileState {
	return e.file
}

func (e *Editor) EnsureVisible(viewWidth, viewHeight int) {
	if viewWidth < 1 {
		viewWidth = 1
	}
	if viewHeight < 1 {
		viewHeight = 1
	}

	cy, cx := e.buf.Cursor()

	if cy < e.vp.TopLine {
		e.vp.TopLine = cy
	} else if cy >= e.vp.TopLine+viewHeight {
		e.vp.TopLine = cy - viewHeight + 1
	}
	if e.vp.TopLine < 0 {
		e.vp.TopLine = 0
	}

	if cx < e.vp.LeftCol {
		e.vp.LeftCol = cx
	} else if cx >= e.vp.LeftCol+viewWidth {
		e.vp.LeftCol = cx - viewWidth + 1
	}
	if e.vp.LeftCol < 0 {
		e.vp.LeftCol = 0
	}
}
