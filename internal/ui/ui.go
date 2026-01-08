package ui

import (
	"cooledit/internal/core"
	"cooledit/internal/term"
)

type UI struct {
	screen term.Screen
	editor *core.Editor
}

func New(screen term.Screen, editor *core.Editor) *UI {
	return &UI{
		screen: screen,
		editor: editor,
	}
}

func (u *UI) Run() error {
	for {
		u.draw()

		ev := u.screen.PollEvent()
		if ev == nil {
			continue
		}

		switch e := ev.(type) {
		case term.ResizeEvent:
			continue

		case term.KeyEvent:
			cmd := u.translateKey(e)
			res := u.editor.Apply(cmd)
			if res.Quit {
				return nil
			}
		}
	}
}

func (u *UI) translateKey(e term.KeyEvent) core.Command {
	switch {
	case e.Key == term.KeyRune && e.Modifiers == 0:
		return core.CmdInsertRune{Rune: e.Rune}

	case e.Key == term.KeyEnter:
		return core.CmdInsertNewline{}

	case e.Key == term.KeyBackspace:
		return core.CmdBackspace{}

	case e.Key == term.KeyLeft:
		return core.CmdMoveLeft{}

	case e.Key == term.KeyRight:
		return core.CmdMoveRight{}

	case e.Key == term.KeyUp:
		return core.CmdMoveUp{}

	case e.Key == term.KeyDown:
		return core.CmdMoveDown{}

	case e.Key == term.KeyHome:
		return core.CmdMoveHome{}

	case e.Key == term.KeyEnd:
		return core.CmdMoveEnd{}

	case e.Key == term.KeyRune &&
		e.Rune == 'c' &&
		e.Modifiers&term.ModCtrl != 0:
		return core.CmdQuit{}
	}

	return core.CmdNoOp{}
}

func (u *UI) draw() {
	w, h := u.screen.Size()
	u.screen.HideCursor()
	u.clear(w, h)

	lines := u.editor.Lines()
	for y := 0; y < len(lines) && y < h; y++ {
		line := lines[y]
		for x, r := range line {
			if x >= w {
				break
			}
			u.screen.SetCell(x, y, r)
		}
	}

	cy, cx := u.editor.Cursor()
	if cy < h && cx < w {
		u.screen.ShowCursor(cx, cy)
	}

	u.screen.Show()
}

func (u *UI) clear(w, h int) {
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			u.screen.SetCell(x, y, ' ')
		}
	}
}
