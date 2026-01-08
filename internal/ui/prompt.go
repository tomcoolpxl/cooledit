package ui

import (
	"os"

	"cooledit/internal/core"
	"cooledit/internal/term"
)

type PromptKind int

const (
	PromptSaveAs PromptKind = iota
	PromptOverwrite
	PromptQuitConfirm
	PromptFind
)

func (u *UI) startQuitFlow() {
	if !u.editor.Modified() {
		u.quitNow = true
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

func (u *UI) enterFind() {
	u.mode = ModePrompt
	u.promptKind = PromptFind
	u.promptLabel = "Find: "
	u.promptText = nil
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
				if u.editor.File().Path == "" {
					u.enterSaveAs(true)
					return true
				}
				u.exitPrompt()
				res := u.editor.Apply(core.CmdSave{}, 0)
				if res.Message != "" {
					u.enterMessage(res.Message)
				}
				if !u.editor.Modified() {
					u.quitNow = true
				}
				return true

			case 'n', 'N':
				u.quitNow = true
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
				u.exitPrompt()
				u.enterMessage("Save As: empty path")
				return true
			}

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
				u.quitNow = true
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
					u.quitNow = true
				}
				return true

			case 'n', 'N':
				u.promptKind = PromptSaveAs
				u.promptLabel = "Save as: "
				u.promptText = []rune(u.pendingPath)
				u.pendingPath = ""
				return true
			}
		case term.KeyEscape:
			u.exitPrompt()
			return true
		}
		return true
		
	case PromptFind:
		switch e.Key {
		case term.KeyEnter:
			query := string(u.promptText)
			u.exitPrompt()
			res := u.editor.Apply(core.CmdFind{Query: query}, u.layout.Viewport.H)
			if res.Message != "" {
				u.enterMessage(res.Message)
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
	}

	return false
}