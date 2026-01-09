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

	DeleteLine(line int) []rune
	InsertLine(line int, runes []rune)
	
	DeleteRange(startLine, startCol, endLine, endCol int)
	RangeText(startLine, startCol, endLine, endCol int) string

	// Helper for Undo/Redo to inspect state
	LineLen(line int) int
	RuneAt(line, col int) rune
}