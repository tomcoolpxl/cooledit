package buffer

type LineBuffer struct {
	data   []rune
	cursor int
}

func NewLineBuffer() *LineBuffer {
	return &LineBuffer{
		data:   make([]rune, 0),
		cursor: 0,
	}
}

func (b *LineBuffer) InsertRune(r rune) {
	if b.cursor < 0 {
		b.cursor = 0
	}
	if b.cursor > len(b.data) {
		b.cursor = len(b.data)
	}

	b.data = append(b.data, 0)
	copy(b.data[b.cursor+1:], b.data[b.cursor:])
	b.data[b.cursor] = r
	b.cursor++
}

func (b *LineBuffer) Backspace() {
	if b.cursor == 0 || len(b.data) == 0 {
		return
	}

	copy(b.data[b.cursor-1:], b.data[b.cursor:])
	b.data = b.data[:len(b.data)-1]
	b.cursor--
}

func (b *LineBuffer) MoveLeft() {
	if b.cursor > 0 {
		b.cursor--
	}
}

func (b *LineBuffer) MoveRight() {
	if b.cursor < len(b.data) {
		b.cursor++
	}
}

func (b *LineBuffer) MoveHome() {
	b.cursor = 0
}

func (b *LineBuffer) MoveEnd() {
	b.cursor = len(b.data)
}

func (b *LineBuffer) Content() []rune {
	return b.data
}

func (b *LineBuffer) CursorCol() int {
	return b.cursor
}
