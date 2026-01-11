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

package core

type Command interface {
	isCommand()
}

type CmdQuit struct{}

func (CmdQuit) isCommand() {}

type CmdInsertRune struct{ Rune rune }

func (CmdInsertRune) isCommand() {}

type CmdReplaceRune struct{ Rune rune }

func (CmdReplaceRune) isCommand() {}

type CmdInsertNewline struct{}

func (CmdInsertNewline) isCommand() {}

type CmdBackspace struct{}

func (CmdBackspace) isCommand() {}

type CmdTab struct{}

func (CmdTab) isCommand() {}

type CmdInsertLiteralTab struct{}

func (CmdInsertLiteralTab) isCommand() {}

type CmdIndentBlock struct{}

func (CmdIndentBlock) isCommand() {}

type CmdUnindentBlock struct{}

func (CmdUnindentBlock) isCommand() {}

type CmdDelete struct{}

func (CmdDelete) isCommand() {}

type CmdMoveLeft struct{ Select bool }

func (CmdMoveLeft) isCommand() {}

type CmdMoveRight struct{ Select bool }

func (CmdMoveRight) isCommand() {}

type CmdMoveWordLeft struct{ Select bool }

func (CmdMoveWordLeft) isCommand() {}

type CmdMoveWordRight struct{ Select bool }

func (CmdMoveWordRight) isCommand() {}

type CmdMoveUp struct{ Select bool }

func (CmdMoveUp) isCommand() {}

type CmdMoveDown struct{ Select bool }

func (CmdMoveDown) isCommand() {}

type CmdMoveHome struct{ Select bool }

func (CmdMoveHome) isCommand() {}

type CmdMoveEnd struct{ Select bool }

func (CmdMoveEnd) isCommand() {}

type CmdPageUp struct{ Select bool }

func (CmdPageUp) isCommand() {}

type CmdPageDown struct{ Select bool }

func (CmdPageDown) isCommand() {}

type CmdFileStart struct{ Select bool }

func (CmdFileStart) isCommand() {}

type CmdFileEnd struct{ Select bool }

func (CmdFileEnd) isCommand() {}

type CmdSave struct{}

func (CmdSave) isCommand() {}

type CmdSaveAs struct {
	Path string
}

func (CmdSaveAs) isCommand() {}

type CmdUndo struct{}

func (CmdUndo) isCommand() {}

type CmdRedo struct{}

func (CmdRedo) isCommand() {}

type CmdFind struct{ Query string }

func (CmdFind) isCommand() {}

type CmdFindNext struct{}

func (CmdFindNext) isCommand() {}

type CmdFindPrev struct{}

func (CmdFindPrev) isCommand() {}

type CmdCopy struct{}

func (CmdCopy) isCommand() {}

type CmdCut struct{}

func (CmdCut) isCommand() {}

type CmdPaste struct {
	Text string
}

func (CmdPaste) isCommand() {}

type CmdGoToLine struct {
	Line int
}

func (CmdGoToLine) isCommand() {}

type CmdToggleLineNumbers struct{}

func (CmdToggleLineNumbers) isCommand() {}

type CmdToggleSoftWrap struct{}

func (CmdToggleSoftWrap) isCommand() {}

type CmdClick struct {
	Line, Col int
}

func (CmdClick) isCommand() {}

type CmdReplace struct {
	Find    string
	Replace string
}

func (CmdReplace) isCommand() {}

type CmdReplaceAll struct {
	Find    string
	Replace string
}

func (CmdReplaceAll) isCommand() {}

type CmdJumpToMatchingBracket struct{}

func (CmdJumpToMatchingBracket) isCommand() {}

type CmdSelectAll struct{}

func (CmdSelectAll) isCommand() {}

type CmdToggleComment struct {
	CommentPrefix string // The comment prefix (e.g., "//", "#")
}

func (CmdToggleComment) isCommand() {}

type CmdFormat struct {
	FormattedText string // The formatted text to replace buffer with
}

func (CmdFormat) isCommand() {}
