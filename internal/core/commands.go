package core

type Command interface {
	isCommand()
}

type CmdQuit struct{}

func (CmdQuit) isCommand() {}

type CmdInsertRune struct{ Rune rune }

func (CmdInsertRune) isCommand() {}

type CmdInsertNewline struct{}

func (CmdInsertNewline) isCommand() {}

type CmdBackspace struct{}

func (CmdBackspace) isCommand() {}

type CmdDelete struct{}

func (CmdDelete) isCommand() {}

type CmdMoveLeft struct{ Select bool }

func (CmdMoveLeft) isCommand() {}

type CmdMoveRight struct{ Select bool }

func (CmdMoveRight) isCommand() {}

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
