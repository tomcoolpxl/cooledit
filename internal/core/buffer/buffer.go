package buffer

type Buffer interface {
	InsertRune(r rune)
	InsertNewline()
	Backspace()

	MoveLeft()
	MoveRight()
	MoveUp()
	MoveDown()
	MoveHome()
	MoveEnd()

	Lines() [][]rune
	Cursor() (line, col int)
}
