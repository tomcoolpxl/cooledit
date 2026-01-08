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

type CmdMoveLeft struct{}
func (CmdMoveLeft) isCommand() {}

type CmdMoveRight struct{}
func (CmdMoveRight) isCommand() {}

type CmdMoveUp struct{}
func (CmdMoveUp) isCommand() {}

type CmdMoveDown struct{}
func (CmdMoveDown) isCommand() {}

type CmdMoveHome struct{}
func (CmdMoveHome) isCommand() {}

type CmdMoveEnd struct{}
func (CmdMoveEnd) isCommand() {}

type CmdPageUp struct{}
func (CmdPageUp) isCommand() {}

type CmdPageDown struct{}
func (CmdPageDown) isCommand() {}

type CmdFileStart struct{}
func (CmdFileStart) isCommand() {}

type CmdFileEnd struct{}
func (CmdFileEnd) isCommand() {}

type CmdNoOp struct{}
func (CmdNoOp) isCommand() {}
