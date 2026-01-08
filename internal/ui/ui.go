package ui

import (
	"fmt"
	"os"
	"time"

	"cooledit/internal/core"
	"cooledit/internal/term"
)

type UIMode int

const (
	ModeNormal UIMode = iota
	ModeMessage
	ModePrompt
	ModeHelp
)

type PromptKind int

const (
	PromptSaveAs PromptKind = iota
	PromptOverwrite
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

	// used by overwrite prompt
	pendingPath string

	// used by quit flow
	quitAfterSave bool

	// ctrl-c force quit
	ctrlCArmed bool
	ctrlCUntil time.Time
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
			// Help: any key exits
			if u.mode == ModeHelp {
				u.mode = ModeNormal
				continue
			}

			// Prompt has priority over everything else
			if u.mode == ModePrompt {
				if u.handlePromptKey(e) {
					continue
				}
			}

			// Esc cancels Ctrl+C armed state and clears message
			if e.Key == term.KeyEscape {
				u.ctrlCArmed = false
				if u.mode == ModeMessage {
					u.mode = ModeNormal
				}
				continue
			}

			// Ctrl+C force quit path
			if u.handleCtrlC(e) {
				return nil
			}

			// Normal keys
			cmd := u.translateKey(e)
			if cmd == nil {
				continue
			}

			res := u.editor.Apply(cmd, viewH)
			if res.Message != "" {
				u.enterMessage(res.Message)
			}
		}
	}
}

func (u *UI) handleCtrlC(e term.KeyEvent) bool {
	if !(e.Key == term.KeyRune && e.Rune == 'c' && (e.Modifiers&term.ModCtrl) != 0) {
		return false
	}

	now := time.Now()
	if u.ctrlCArmed && now.Before(u.ctrlCUntil) {
		return true
	}

	u.ctrlCArmed = true
	u.ctrlCUntil = now.Add(2 * time.Second)

	if u.editor.Modified() {
		u.enterMessage("UNSAVED changes - press Ctrl+C again to quit")
	} else {
		u.enterMessage("Press Ctrl+C again to quit")
	}

	return false
}

func (u *UI) translateKey(e term.KeyEvent) core.Command {
	switch {
	case e.Key == term.KeyF1:
		u.mode = ModeHelp
		return nil

	// Ctrl+Q quit flow
	case e.Key == term.KeyRune && e.Rune == 'q' && (e.Modifiers&term.ModCtrl) != 0:
		u.startQuitFlow()
		return nil

	// Ctrl+S Save
	case e.Key == term.KeyRune && e.Rune == 's' && e.Modifiers == term.ModCtrl:
		if u.editor.File().Path == "" {
			u.enterSaveAs(false)
			return nil
		}
		return core.CmdSave{}

	// Ctrl+Shift+S Save As
	case e.Key == term.KeyRune && e.Rune == 's' && (e.Modifiers&(term.ModCtrl|term.ModShift)) == (term.ModCtrl|term.ModShift):
		u.enterSaveAs(false)
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

	case e.Key == term.KeyHome && (e.Modifiers&term.ModCtrl) != 0:
		return core.CmdFileStart{}
	case e.Key == term.KeyEnd && (e.Modifiers&term.ModCtrl) != 0:
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
		// Quit immediately when clean.
		u.mode = ModeNormal
		u.editor.Apply(core.CmdQuit{}, 0)
		return
	}

	u.mode = ModePrompt
	u.promptKind = PromptQuitConfirm
	u.promptLabel = "Unsaved changes. Save before quitting? (y/n) "
	u.promptText = nil
	u.quitAfterSave = false
}

func (u *UI) enterSaveAs(quitAfter bool) {
	u.mode = ModePrompt
	u.promptKind = PromptSaveAs
	u.promptLabel = "Save as: "
	u.promptText = nil
	u.pendingPath = ""
	u.quitAfterSave = quitAfter
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
	u.pendingPath = ""
	u.quitAfterSave = false
}

func (u *UI) handlePromptKey(e term.KeyEvent) bool {
	switch u.promptKind {

	case PromptQuitConfirm:
		switch e.Key {
		case term.KeyRune:
			switch e.Rune {
			case 'y', 'Y':
				// Save then quit
				if u.editor.File().Path == "" {
					// need Save As, and quit after successful save
					u.enterSaveAs(true)
					return true
				}
				u.exitPrompt()
				res := u.editor.Apply(core.CmdSave{}, 0)
				if res.Message != "" {
					u.enterMessage(res.Message)
				}
				// If save succeeded, modified should now be false.
				if !u.editor.Modified() {
					u.editor.Apply(core.CmdQuit{}, 0)
				}
				return true

			case 'n', 'N':
				u.exitPrompt()
				u.editor.Apply(core.CmdQuit{}, 0)
				return true
			}
		case term.KeyEscape:
			u.exitPrompt()
			return true
		}
		return true

	case PromptSaveAs:
		switch e.Key {
		case term.KeyEnter:
			path := string(u.promptText)
			if path == "" {
				u.enterMessage("Save As: empty path")
				u.exitPrompt()
				return true
			}

			// Overwrite confirm only if:
			// - target exists
			// - and target differs from current file path
			if _, err := os.Stat(path); err == nil && path != u.editor.File().Path {
				u.promptKind = PromptOverwrite
				u.promptLabel = "Overwrite existing file? (y/n) "
				u.pendingPath = path
				u.promptText = nil
				return true
			}

			u.exitPrompt()
			res := u.editor.Apply(core.CmdSaveAs{Path: path}, 0)
			if res.Message != "" {
				u.enterMessage(res.Message)
			}
			if u.quitAfterSave && !u.editor.Modified() {
				u.editor.Apply(core.CmdQuit{}, 0)
			}
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

	case PromptOverwrite:
		switch e.Key {
		case term.KeyRune:
			switch e.Rune {
			case 'y', 'Y':
				path := u.pendingPath
				quitAfter := u.quitAfterSave
				u.exitPrompt()
				res := u.editor.Apply(core.CmdSaveAs{Path: path}, 0)
				if res.Message != "" {
					u.enterMessage(res.Message)
				}
				if quitAfter && !u.editor.Modified() {
					u.editor.Apply(core.CmdQuit{}, 0)
				}
				return true

			case 'n', 'N':
				// Back to Save As prompt, keep quitAfterSave behavior
				quitAfter := u.quitAfterSave
				u.enterSaveAs(quitAfter)
				return true
			}
		case term.KeyEscape:
			u.exitPrompt()
			return true
		}
		return true
	}

	return false
}

func (u *UI) draw(w, h, viewH int) {
	u.screen.HideCursor()
	u.clear(w, h)

	if u.mode == ModeHelp {
		u.drawHelp(w, h)
		u.screen.Show()
		return
	}

	viewW := w
	if viewW < 1 {
		viewW = 1
	}

	u.editor.EnsureVisible(viewW, viewH)
	vp := u.editor.Viewport()
	lines := u.editor.Lines()

	// editor area (viewport aware: vertical + horizontal)
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
			u.screen.SetCell(sx, sy, line[docX], term.Style{})
		}
	}

	// cursor
	if u.mode == ModeNormal || u.mode == ModeMessage {
		cy, cx := u.editor.Cursor()
		sx := cx - vp.LeftCol
		sy := cy - vp.TopLine
		if sx >= 0 && sx < viewW && sy >= 0 && sy < viewH {
			u.screen.ShowCursor(sx, sy)
		}
	}

	u.drawStatusBar(w, h, vp)
	u.screen.Show()
}

func (u *UI) drawStatusBar(w, h int, vp core.Viewport) {
	row := h - 1
	style := term.Style{Inverse: true}

	// clear bar
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

		// Cursor for prompt input only for Save As (text entry)
		if u.promptKind == PromptSaveAs {
			cx := len(u.promptLabel) + len(u.promptText)
			if cx < w {
				u.screen.ShowCursor(cx, row)
			}
		}
		return

	case ModeMessage:
		if time.Now().After(u.messageUntil) {
			u.mode = ModeNormal
		} else {
			for i, r := range u.message {
				if i >= w {
					break
				}
				u.screen.SetCell(i, row, r, style)
			}
			return
		}
	}

	// ModeNormal layout: left + right (stable, right-aligned)
	fs := u.editor.File()
	mod := ""
	if u.editor.Modified() {
		mod = "*"
	}

	left := fmt.Sprintf("%s%s  Ctrl+S Save  Ctrl+Shift+S Save As  F1 Help", fs.BaseName, mod)

	cy, cx := u.editor.Cursor()
	eol := "LF"
	if fs.EOL == "\r\n" {
		eol = "CRLF"
	}
	right := fmt.Sprintf("Ln %d, Col %d  %s %s", cy+1, cx+1, fs.Encoding, eol)

	// draw right aligned first
	startRight := w - len(right)
	if startRight < 0 {
		startRight = 0
	}
	for i, r := range right {
		x := startRight + i
		if x >= 0 && x < w {
			u.screen.SetCell(x, row, r, style)
		}
	}

	// draw left, but do not overwrite right area
	maxLeft := startRight - 1
	if maxLeft < 0 {
		maxLeft = 0
	}
	for i, r := range left {
		if i >= w {
			break
		}
		if i >= maxLeft {
			break
		}
		u.screen.SetCell(i, row, r, style)
	}
}

func (u *UI) drawHelp(w, h int) {
	lines := []string{
		"cooledit - help",
		"",
		"Ctrl+S        Save",
		"Ctrl+Shift+S  Save As",
		"Ctrl+Q        Quit (prompts if unsaved)",
		"Ctrl+C        Force quit (press twice, Esc cancels)",
		"Arrows        Move cursor",
		"PgUp/PgDn     Scroll",
		"Ctrl+Home/End File start/end",
		"F1            Help",
		"",
		"Press any key to return",
	}

	for y := 0; y < len(lines) && y < h; y++ {
		for x, r := range lines[y] {
			if x >= w {
				break
			}
			u.screen.SetCell(x, y, r, term.Style{})
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
