package ui

import (
	"fmt"

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
		w, h := u.screen.Size()
		viewH := h - 1
		if viewH < 1 {
			viewH = 1
		}

		u.draw(w, h, viewH)

		ev := u.screen.PollEvent()
		if ev == nil {
			continue
		}

		switch e := ev.(type) {
		case term.ResizeEvent:
			continue

		case term.KeyEvent:
			cmd := u.translateKey(e)
			res := u.editor.Apply(cmd, viewH)
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

	case e.Key == term.KeyHome && e.Modifiers&term.ModCtrl != 0:
		return core.CmdFileStart{}

	case e.Key == term.KeyEnd && e.Modifiers&term.ModCtrl != 0:
		return core.CmdFileEnd{}

	case e.Key == term.KeyHome:
		return core.CmdMoveHome{}

	case e.Key == term.KeyEnd:
		return core.CmdMoveEnd{}

	case e.Key == term.KeyPageUp:
		return core.CmdPageUp{}

	case e.Key == term.KeyPageDown:
		return core.CmdPageDown{}

	case e.Key == term.KeyRune &&
		e.Rune == 'c' &&
		e.Modifiers&term.ModCtrl != 0:
		return core.CmdQuit{}
	}

	return core.CmdNoOp{}
}

func (u *UI) draw(w, h, viewH int) {
	u.screen.HideCursor()
	u.clear(w, h)

	viewW := w
	if viewW < 1 {
		viewW = 1
	}

	u.editor.EnsureVisible(viewW, viewH)
	vp := u.editor.Viewport()

	lines := u.editor.Lines()

	for sy := 0; sy < viewH; sy++ {
		docY := vp.TopLine + sy
		if docY < 0 || docY >= len(lines) {
			continue
		}
		line := lines[docY]
		start := vp.LeftCol
		if start > len(line) {
			start = len(line)
		}
		for sx := 0; sx < viewW; sx++ {
			docX := start + sx
			if docX >= len(line) {
				break
			}
			u.screen.SetCell(sx, sy, line[docX])
		}
	}

	cy, cx := u.editor.Cursor()
	sx := cx - vp.LeftCol
	sy := cy - vp.TopLine
	if sy >= 0 && sy < viewH && sx >= 0 && sx < viewW {
		u.screen.ShowCursor(sx, sy)
	}

	u.drawStatusBar(w, h, vp)
	u.screen.Show()
}

func (u *UI) drawStatusBar(w, h int, vp core.Viewport) {
	if h < 1 {
		return
	}
	row := h - 1

	cy, cx := u.editor.Cursor()
	mod := ""
	if u.editor.Modified() {
		mod = "*"
	}

	status := fmt.Sprintf(
		"[No Name]%s  Ln %d, Col %d   Ctrl+Home End  PgUp PgDn",
		mod, cy+1, cx+1,
	)

	for i := 0; i < w; i++ {
		ch := ' '
		if i < len(status) {
			ch = rune(status[i])
		}
		u.screen.SetCell(i, row, ch)
	}
}

func (u *UI) clear(w, h int) {
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			u.screen.SetCell(x, y, ' ')
		}
	}
}
