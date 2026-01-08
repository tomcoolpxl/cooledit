package ui

import (
	"fmt"
	"time"

	"cooledit/internal/core"
	"cooledit/internal/term"
)

type UIMode int

const (
	ModeNormal UIMode = iota
	ModeMessage
	ModePrompt
)

type PromptKind int

const (
	PromptSaveAs PromptKind = iota
	PromptQuitConfirm
)

type UI struct {
	screen term.Screen
	editor *core.Editor

	mode UIMode

	// message mode
	message      string
	messageUntil time.Time

	// prompt mode
	promptKind  PromptKind
	promptLabel string
	promptText  []rune

	// ctrl-c quit
	quitPending bool
	quitUntil   time.Time
}

func New(screen term.Screen, editor *core.Editor) *UI {
	return &UI{
		screen: screen,
		editor: editor,
		mode:   ModeNormal,
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
		case term.KeyEvent:
			if u.mode == ModePrompt {
				if u.handlePromptKey(e) {
					continue
				}
			}

			if u.handleCtrlC(e) {
				return nil
			}

			cmd := u.translateKey(e)
			if cmd == nil {
				continue
			}

			res := u.editor.Apply(cmd, viewH)
			if res.Quit {
				return nil
			}
			if res.Message != "" {
				u.enterMessage(res.Message)
			}
		}
	}
}

func (u *UI) handleCtrlC(e term.KeyEvent) bool {
	if e.Key != term.KeyRune || e.Rune != 'c' || e.Modifiers&term.ModCtrl == 0 {
		return false
	}

	now := time.Now()

	if u.quitPending && now.Before(u.quitUntil) {
		return true
	}

	u.quitPending = true
	u.quitUntil = now.Add(2 * time.Second)

	if u.editor.Modified() {
		u.enterMessage("UNSAVED changes — press Ctrl+C again to quit")
	} else {
		u.enterMessage("Press Ctrl+C again to quit")
	}

	return false
}

func (u *UI) translateKey(e term.KeyEvent) core.Command {
	// Esc cancels pending Ctrl+C
	if e.Key == term.KeyEscape {
		u.quitPending = false
		if u.mode == ModeMessage {
			u.mode = ModeNormal
		}
		return nil
	}

	switch {
	case e.Key == term.KeyRune && e.Rune == 'q' && e.Modifiers&term.ModCtrl != 0:
		u.startQuitFlow()
		return nil

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

	case e.Key == term.KeyPageUp:
		return core.CmdPageUp{}

	case e.Key == term.KeyPageDown:
		return core.CmdPageDown{}

	case e.Key == term.KeyHome && e.Modifiers&term.ModCtrl != 0:
		return core.CmdFileStart{}

	case e.Key == term.KeyEnd && e.Modifiers&term.ModCtrl != 0:
		return core.CmdFileEnd{}

	case e.Key == term.KeyHome:
		return core.CmdMoveHome{}

	case e.Key == term.KeyEnd:
		return core.CmdMoveEnd{}
	}

	return nil
}

func (u *UI) startQuitFlow() {
	if !u.editor.Modified() {
		// safe quit
		u.enterMessage("Quit")
		u.mode = ModeNormal
		u.editor.Apply(core.CmdQuit{}, 0)
		return
	}

	u.mode = ModePrompt
	u.promptKind = PromptQuitConfirm
	u.promptLabel = "Unsaved changes. Save before quitting? (y/n) "
	u.promptText = nil
}

func (u *UI) handlePromptKey(e term.KeyEvent) bool {
	switch u.promptKind {
	case PromptSaveAs:
		return u.handleSaveAsPrompt(e)
	case PromptQuitConfirm:
		return u.handleQuitConfirmPrompt(e)
	}
	return false
}

func (u *UI) handleQuitConfirmPrompt(e term.KeyEvent) bool {
	switch e.Key {
	case term.KeyRune:
		switch e.Rune {
		case 'y', 'Y':
			u.exitPrompt()
			if u.editor.File().Path == "" {
				u.enterSaveAs()
				return true
			}
			u.editor.Apply(core.CmdSaveAs{Path: u.editor.File().Path}, 0)
			return true
		case 'n', 'N':
			return true
		}

	case term.KeyEscape:
		u.exitPrompt()
		return true
	}
	return true
}

func (u *UI) handleSaveAsPrompt(e term.KeyEvent) bool {
	switch e.Key {
	case term.KeyEnter:
		path := string(u.promptText)
		u.exitPrompt()
		u.editor.Apply(core.CmdSaveAs{Path: path}, 0)
		return true

	case term.KeyEscape:
		u.exitPrompt()
		return true

	case term.KeyBackspace:
		if len(u.promptText) > 0 {
			u.promptText = u.promptText[:len(u.promptText)-1]
		}
		return true

	case term.KeyRune:
		u.promptText = append(u.promptText, e.Rune)
		return true
	}
	return true
}

func (u *UI) enterSaveAs() {
	u.mode = ModePrompt
	u.promptKind = PromptSaveAs
	u.promptLabel = "Save as: "
	u.promptText = nil
}

func (u *UI) enterMessage(msg string) {
	u.mode = ModeMessage
	u.message = msg
	u.messageUntil = time.Now().Add(2 * time.Second)
}

func (u *UI) exitPrompt() {
	u.mode = ModeNormal
	u.promptText = nil
	u.promptLabel = ""
}

func (u *UI) draw(w, h, viewH int) {
	u.screen.HideCursor()
	u.clear(w, h)

	viewW := w
	u.editor.EnsureVisible(viewW, viewH)
	vp := u.editor.Viewport()
	lines := u.editor.Lines()

	for sy := 0; sy < viewH; sy++ {
		docY := vp.TopLine + sy
		if docY < 0 || docY >= len(lines) {
			continue
		}
		line := lines[docY]
		for sx, r := range line {
			if sx >= viewW {
				break
			}
			u.screen.SetCell(sx, sy, r, term.Style{})
		}
	}

	if u.mode == ModeNormal || u.mode == ModeMessage {
		cy, cx := u.editor.Cursor()
		sx := cx - vp.LeftCol
		sy := cy - vp.TopLine
		if sx >= 0 && sx < viewW && sy >= 0 && sy < viewH {
			u.screen.ShowCursor(sx, sy)
		}
	}

	u.drawStatusBar(w, h)
	u.screen.Show()
}

func (u *UI) drawStatusBar(w, h int) {
	row := h - 1
	style := term.Style{Inverse: true}

	for x := 0; x < w; x++ {
		u.screen.SetCell(x, row, ' ', style)
	}

	switch u.mode {
	case ModePrompt:
		text := u.promptLabel + string(u.promptText)
		for i, r := range text {
			if i >= w {
				break
			}
			u.screen.SetCell(i, row, r, style)
		}
		cx := len(u.promptLabel) + len(u.promptText)
		if cx < w {
			u.screen.ShowCursor(cx, row)
		}

	case ModeMessage:
		if time.Now().After(u.messageUntil) {
			u.mode = ModeNormal
			return
		}
		for i, r := range u.message {
			if i >= w {
				break
			}
			u.screen.SetCell(i, row, r, style)
		}

	case ModeNormal:
		fs := u.editor.File()
		mod := ""
		if u.editor.Modified() {
			mod = "*"
		}
		left := fs.BaseName + mod
		for i, r := range left {
			if i >= w {
				break
			}
			u.screen.SetCell(i, row, r, style)
		}

		cy, cx := u.editor.Cursor()
		right := fmt.Sprintf("Ln %d, Col %d", cy+1, cx+1)
		start := w - len(right)
		if start < 0 {
			start = 0
		}
		for i, r := range right {
			u.screen.SetCell(start+i, row, r, style)
		}
	}
}

func (u *UI) clear(w, h int) {
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			u.screen.SetCell(x, y, ' ', term.Style{})
		}
	}
}
