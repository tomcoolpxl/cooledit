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

// NewLineBufferFromLines creates a buffer from existing lines (read-only load).
// Cursor starts at (0,0).
func NewLineBufferFromLines(lines [][]rune) *LineBuffer {
	if len(lines) == 0 {
		lines = [][]rune{make([]rune, 0)}
	}
	return &LineBuffer{
		lines:        lines,
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

	b.lines = append(b.lines, nil)
	copy(b.lines[b.line+2:], b.lines[b.line+1:])
	b.lines[b.line+1] = newLine

	b.line++
	b.col = 0
	b.preferredCol = 0
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

func (b *LineBuffer) SetCursor(line, col int) {
	if line < 0 {
		line = 0
	}
	if line >= len(b.lines) {
		line = len(b.lines) - 1
	}
	b.line = line

	if col < 0 {
		col = 0
	}
	if col > len(b.lines[b.line]) {
		col = len(b.lines[b.line])
	}
	b.col = col
	b.preferredCol = col
}

func (b *LineBuffer) DeleteLine(line int) []rune {
	if line < 0 || line >= len(b.lines) {
		return nil
	}
	runes := b.lines[line]
	b.lines = append(b.lines[:line], b.lines[line+1:]...)
	if len(b.lines) == 0 {
		b.lines = [][]rune{make([]rune, 0)}
	}
	if b.line >= len(b.lines) {
		b.line = len(b.lines) - 1
		b.clampCol()
	} else if b.line == line {
		b.clampCol()
	}
	return runes
}

func (b *LineBuffer) InsertLine(line int, runes []rune) {
	if line < 0 {
		line = 0
	}
	if line > len(b.lines) {
		line = len(b.lines)
	}

	b.lines = append(b.lines, nil)
	copy(b.lines[line+1:], b.lines[line:])
	b.lines[line] = runes

	if b.line >= line {
		b.line++
	}
}

func (b *LineBuffer) LineLen(line int) int {
	if line < 0 || line >= len(b.lines) {
		return 0
	}
	return len(b.lines[line])
}

func (b *LineBuffer) RuneAt(line, col int) rune {
	if line < 0 || line >= len(b.lines) {
		return 0
	}
	l := b.lines[line]
	if col < 0 || col >= len(l) {
		return 0
	}
	return l[col]
}