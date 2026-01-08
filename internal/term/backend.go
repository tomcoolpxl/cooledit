package term

type Key int

const (
	KeyUnknown Key = iota

	KeyRune
	KeyEnter
	KeyBackspace
	KeyDelete
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

type Style struct {
	Inverse bool
}

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

type MouseButton int

const (
	MouseNone MouseButton = iota
	MouseLeft
	MouseRight
	MouseMiddle
	MouseWheelUp
	MouseWheelDown
)

type MouseEvent struct {
	X, Y   int
	Button MouseButton
}

func (MouseEvent) isEvent() {}

type Screen interface {
	Init(enableMouse bool) error
	Fini()

	Size() (width, height int)
	PollEvent() Event

	SetCell(x, y int, ch rune, style Style)
	Show()

	ShowCursor(x, y int)
	HideCursor()
}