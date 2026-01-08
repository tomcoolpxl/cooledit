package core

import (
	"time"

	"cooledit/internal/core/buffer"
)

type Editor struct {
	buf         buffer.Buffer
	quitPending bool
	quitAt      time.Time
}

func NewEditor() *Editor {
	return &Editor{
		buf: buffer.NewLineBuffer(),
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
		e.buf.InsertRune(c.Rune)

	case CmdInsertNewline:
		e.quitPending = false
		e.buf.InsertNewline()

	case CmdBackspace:
		e.quitPending = false
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
