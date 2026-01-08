package core

type Command interface {
	isCommand()
}

type CmdQuit struct{}

func (CmdQuit) isCommand() {}

type CmdInsertRune struct {
	Rune rune
}

func (CmdInsertRune) isCommand() {}

type CmdBackspace struct{}

func (CmdBackspace) isCommand() {}

type CmdMoveLeft struct{}

func (CmdMoveLeft) isCommand() {}

type CmdMoveRight struct{}

func (CmdMoveRight) isCommand() {}

type CmdMoveHome struct{}

func (CmdMoveHome) isCommand() {}

type CmdMoveEnd struct{}

func (CmdMoveEnd) isCommand() {}

type CmdNoOp struct{}

func (CmdNoOp) isCommand() {}
