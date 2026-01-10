// Copyright (C) 2026 Tom Cool
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package ui

import (
	"fmt"
	"os"

	"cooledit/internal/core"
	"cooledit/internal/term"
)

type PromptKind int

const (
	PromptSaveAs PromptKind = iota
	PromptOverwrite
	PromptQuitConfirm
	PromptGoToLine
	PromptReplaceWith
	PromptReplaceAllConfirm
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

func (u *UI) enterGoToLine() {
	u.mode = ModePrompt
	u.promptKind = PromptGoToLine
	u.promptLabel = "Go to line: "
	u.promptText = nil
}

func (u *UI) enterReplaceAllConfirm() {
	// Get match count from search session
	matchCount := 0
	searchState := u.editor.SearchState()

	// If no session exists, create one from lastFindTerm
	if searchState != nil && searchState.Session == nil && u.lastFindTerm != "" {
		u.editor.StartSearchSession(u.lastFindTerm)
		searchState = u.editor.SearchState()
	}

	if searchState != nil && searchState.Session != nil {
		matchCount = len(searchState.Session.Matches)
	}

	u.mode = ModePrompt
	u.promptKind = PromptReplaceAllConfirm
	if matchCount > 0 {
		u.promptLabel = fmt.Sprintf("Replace all %d matches? (y/n) ", matchCount)
	} else {
		u.promptLabel = "Replace all matches? (y/n) "
	}
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

	case PromptGoToLine:
		switch e.Key {
		case term.KeyEnter:
			lineStr := string(u.promptText)
			u.exitPrompt()
			if lineStr != "" {
				var line int
				_, err := fmt.Sscanf(lineStr, "%d", &line)
				if err == nil {
					res := u.editor.Apply(core.CmdGoToLine{Line: line}, u.layout.Viewport.H)
					if res.Message != "" {
						u.enterMessage(res.Message)
					}
				} else {
					u.enterMessage("Invalid line number")
				}
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
			if e.Rune >= '0' && e.Rune <= '9' {
				u.promptText = append(u.promptText, e.Rune)
			}
			return true
		}
		return true

	case PromptReplaceWith:
		switch e.Key {
		case term.KeyEnter:
			replaceTerm := string(u.promptText)
			u.lastReplaceTerm = replaceTerm
			u.exitPrompt()

			if u.replacingAll {
				// Replace all occurrences
				u.replacingAll = false
				res := u.editor.Apply(core.CmdReplaceAll{
					Find:    u.lastFindTerm,
					Replace: replaceTerm,
				}, u.layout.Viewport.H)
				u.mode = ModeNormal
				u.enterMessage(res.Message)
			} else {
				// Replace current match and return to search mode
				res := u.editor.Apply(core.CmdReplace{
					Find:    u.lastFindTerm,
					Replace: replaceTerm,
				}, u.layout.Viewport.H)
				if res.Message != "" && (res.Message[:9] == "Not found" || res.Message == "No matches found") {
					u.mode = ModeNormal
					u.enterMessage(res.Message)
				} else {
					// Return to search mode with the search query
					u.searchQuery = []rune(u.lastFindTerm)
					u.mode = ModeSearch
					u.performSearch()
				}
			}
			return true

		case term.KeyEscape:
			u.exitPrompt()
			u.replacingAll = false
			// Go back to search mode if we have a search query
			if len(u.searchQuery) > 0 {
				u.mode = ModeSearch
			} else {
				u.mode = ModeNormal
			}
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

	case PromptReplaceAllConfirm:
		switch e.Key {
		case term.KeyRune:
			switch e.Rune {
			case 'y', 'Y':
				// User confirmed - prompt for replacement text
				u.exitPrompt()
				u.replacingAll = true
				u.mode = ModePrompt
				u.promptKind = PromptReplaceWith
				u.promptLabel = "Replace all with: "
				u.promptText = []rune(u.lastReplaceTerm)
				return true

			case 'n', 'N':
				// User cancelled - return to previous mode
				u.exitPrompt()
				if len(u.searchQuery) > 0 {
					u.mode = ModeSearch
				} else {
					u.mode = ModeNormal
				}
				return true
			}
		case term.KeyEscape:
			// Escape cancels - return to previous mode
			u.exitPrompt()
			if len(u.searchQuery) > 0 {
				u.mode = ModeSearch
			} else {
				u.mode = ModeNormal
			}
			return true
		}
		return true
	}

	return false
}
