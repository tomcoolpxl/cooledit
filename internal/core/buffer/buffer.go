package buffer

type Buffer interface {
	InsertRune(r rune)
	Backspace()

	MoveLeft()
	MoveRight()
	MoveHome()
	MoveEnd()

	Content() []rune
	CursorCol() int
}
