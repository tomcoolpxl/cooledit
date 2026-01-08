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
		return Result{}

	case CmdBackspace:
		e.quitPending = false
		e.buf.Backspace()
		return Result{}

	case CmdMoveLeft:
		e.quitPending = false
		e.buf.MoveLeft()
		return Result{}

	case CmdMoveRight:
		e.quitPending = false
		e.buf.MoveRight()
		return Result{}

	case CmdMoveHome:
		e.quitPending = false
		e.buf.MoveHome()
		return Result{}

	case CmdMoveEnd:
		e.quitPending = false
		e.buf.MoveEnd()
		return Result{}

	default:
		e.quitPending = false
		return Result{}
	}
}

func (e *Editor) Content() []rune {
	return e.buf.Content()
}

func (e *Editor) CursorCol() int {
	return e.buf.CursorCol()
}
