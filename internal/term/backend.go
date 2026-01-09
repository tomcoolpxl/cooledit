// Copyright (C) 2026 Tom Cool
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package term

type Key int

const (
	KeyUnknown Key = iota

	KeyRune
	KeyEnter
	KeyBackspace
	KeyDelete
	KeyInsert
	KeyEscape
	KeyTab

	KeyUp
	KeyDown
	KeyLeft
	KeyRight

	KeyHome
	KeyEnd
	KeyPageUp
	KeyPageDown

	KeyF1
	KeyF2
	KeyF3
	KeyF4
	KeyF5
	KeyF6
	KeyF7
	KeyF8
	KeyF9
	KeyF10
	KeyF11
	KeyF12
)

type ModMask int

const (
	ModNone ModMask = 0
	ModCtrl ModMask = 1 << iota
	ModAlt
	ModShift
)

type Color string

const (
	ColorDefault Color = "default"
	ColorBlack   Color = "black"
	ColorRed     Color = "red"
	ColorGreen   Color = "green"
	ColorYellow  Color = "yellow"
	ColorBlue    Color = "blue"
	ColorMagenta Color = "magenta"
	ColorCyan    Color = "cyan"
	ColorWhite   Color = "white"
)

type Style struct {
	Foreground Color
	Background Color
	Inverse    bool // Legacy support, overrides colors when true
	Underline  bool // Underline text
}

type CursorShape int

const (
	CursorBlock CursorShape = iota
	CursorUnderline
	CursorBar
)

type Event interface {
	isEvent()
}

type KeyEvent struct {
	Key       Key
	Rune      rune
	Modifiers ModMask
}

func (KeyEvent) isEvent() {}

type ResizeEvent struct {
	Width  int
	Height int
}

func (ResizeEvent) isEvent() {}

type RedrawEvent struct{}

func (RedrawEvent) isEvent() {}

type Screen interface {
	Init() error
	Fini()

	Size() (width, height int)
	PollEvent() Event
	PushEvent(Event)

	SetCell(x, y int, ch rune, style Style)
	Show()

	SetCursorShape(shape CursorShape, color Color)
	ShowCursor(x, y int)
	HideCursor()
}
