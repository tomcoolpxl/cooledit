package buffer

type LineBuffer struct {
	lines        [][]rune
	line         int
	col          int
	preferredCol int
}

func NewLineBuffer() *LineBuffer {
	return &LineBuffer{
		lines:        [][]rune{make([]rune, 0)},
		line:         0,
		col:          0,
		preferredCol: 0,
	}
}

func (b *LineBuffer) clampCol() {
	if b.col > len(b.lines[b.line]) {
		b.col = len(b.lines[b.line])
	}
}

func (b *LineBuffer) InsertRune(r rune) {
	line := b.lines[b.line]

	line = append(line, 0)
	copy(line[b.col+1:], line[b.col:])
	line[b.col] = r

	b.lines[b.line] = line
	b.col++
	b.preferredCol = b.col
}

func (b *LineBuffer) InsertNewline() {
	line := b.lines[b.line]

	newLine := append([]rune{}, line[b.col:]...)
	b.lines[b.line] = line[:b.col]

	b.line++
	b.col = 0
	b.preferredCol = 0

	b.lines = append(b.lines, nil)
	copy(b.lines[b.line+1:], b.lines[b.line:])
	b.lines[b.line] = newLine
}

func (b *LineBuffer) Backspace() {
	if b.col > 0 {
		line := b.lines[b.line]
		copy(line[b.col-1:], line[b.col:])
		b.lines[b.line] = line[:len(line)-1]
		b.col--
		b.preferredCol = b.col
		return
	}

	if b.line == 0 {
		return
	}

	prev := b.lines[b.line-1]
	curr := b.lines[b.line]

	b.col = len(prev)
	b.preferredCol = b.col
	b.lines[b.line-1] = append(prev, curr...)

	copy(b.lines[b.line:], b.lines[b.line+1:])
	b.lines = b.lines[:len(b.lines)-1]
	b.line--
}

func (b *LineBuffer) MoveLeft() {
	if b.col > 0 {
		b.col--
		b.preferredCol = b.col
		return
	}
	if b.line > 0 {
		b.line--
		b.col = len(b.lines[b.line])
		b.preferredCol = b.col
	}
}

func (b *LineBuffer) MoveRight() {
	if b.col < len(b.lines[b.line]) {
		b.col++
		b.preferredCol = b.col
		return
	}
	if b.line+1 < len(b.lines) {
		b.line++
		b.col = 0
		b.preferredCol = 0
	}
}

func (b *LineBuffer) MoveUp() {
	if b.line == 0 {
		return
	}
	b.line--
	b.col = b.preferredCol
	b.clampCol()
}

func (b *LineBuffer) MoveDown() {
	if b.line+1 >= len(b.lines) {
		return
	}
	b.line++
	b.col = b.preferredCol
	b.clampCol()
}

func (b *LineBuffer) MoveHome() {
	b.col = 0
	b.preferredCol = 0
}

func (b *LineBuffer) MoveEnd() {
	b.col = len(b.lines[b.line])
	b.preferredCol = b.col
}

func (b *LineBuffer) Lines() [][]rune {
	return b.lines
}

func (b *LineBuffer) Cursor() (int, int) {
	return b.line, b.col
}
