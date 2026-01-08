package core

import (
	"time"

	"cooledit/internal/core/buffer"
)

type Viewport struct {
	TopLine int
	LeftCol int
}

type Editor struct {
	buf         buffer.Buffer
	vp          Viewport
	modified    bool
	quitPending bool
	quitAt      time.Time
}

func NewEditor() *Editor {
	return &Editor{
		buf: buffer.NewLineBuffer(),
		vp:  Viewport{TopLine: 0, LeftCol: 0},
	}
}

type Result struct {
	Quit bool
}

func (e *Editor) Apply(cmd Command) Result {
	switch c := cmd.(type) {
	case CmdQuit:
		now := time.Now()
		if e.quitPending && now.Sub(e.quitAt) < 2*time.Second {
			return Result{Quit: true}
		}
		e.quitPending = true
		e.quitAt = now
		return Result{}

	case CmdInsertRune:
		e.quitPending = false
		e.modified = true
		e.buf.InsertRune(c.Rune)

	case CmdInsertNewline:
		e.quitPending = false
		e.modified = true
		e.buf.InsertNewline()

	case CmdBackspace:
		e.quitPending = false
		e.modified = true
		e.buf.Backspace()

	case CmdMoveLeft:
		e.quitPending = false
		e.buf.MoveLeft()

	case CmdMoveRight:
		e.quitPending = false
		e.buf.MoveRight()

	case CmdMoveUp:
		e.quitPending = false
		e.buf.MoveUp()

	case CmdMoveDown:
		e.quitPending = false
		e.buf.MoveDown()

	case CmdMoveHome:
		e.quitPending = false
		e.buf.MoveHome()

	case CmdMoveEnd:
		e.quitPending = false
		e.buf.MoveEnd()

	case CmdPageUp:
		e.quitPending = false
		for i := 0; i < e.pageSize(); i++ {
			e.buf.MoveUp()
		}

	case CmdPageDown:
		e.quitPending = false
		for i := 0; i < e.pageSize(); i++ {
			e.buf.MoveDown()
		}

	case CmdFileStart:
		e.quitPending = false
		for {
			prevLine, _ := e.buf.Cursor()
			e.buf.MoveUp()
			line, _ := e.buf.Cursor()
			if line == prevLine {
				break
			}
		}
		e.buf.MoveHome()

	case CmdFileEnd:
		e.quitPending = false
		for {
			prevLine, _ := e.buf.Cursor()
			e.buf.MoveDown()
			line, _ := e.buf.Cursor()
			if line == prevLine {
				break
			}
		}
		e.buf.MoveEnd()

	default:
		e.quitPending = false
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

// pageSize returns a conservative default; UI will keep viewport in sync.
func (e *Editor) pageSize() int {
	return 10
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
