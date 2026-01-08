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
	SetCursor(line, col int)
	
	// Helper for Undo/Redo to inspect state
	LineLen(line int) int
	RuneAt(line, col int) rune
}