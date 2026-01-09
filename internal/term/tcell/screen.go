package tcell

import (
	"github.com/gdamore/tcell/v2"

	"cooledit/internal/term"
)

type Screen struct {
	screen tcell.Screen
}

func New() *Screen {
	return &Screen{}
}

func (s *Screen) Init(enableMouse bool) error {
	ts, err := tcell.NewScreen()
	if err != nil {
		return err
	}
	if err := ts.Init(); err != nil {
		return err
	}
	ts.Clear()
	if enableMouse {
		ts.EnableMouse(tcell.MouseButtonEvents)
	}
	s.screen = ts
	return nil
}

func (s *Screen) Fini() {
	if s.screen != nil {
		s.screen.Fini()
	}
}

func (s *Screen) Size() (int, int) {
	return s.screen.Size()
}

func (s *Screen) PollEvent() term.Event {
	ev := s.screen.PollEvent()

	switch e := ev.(type) {
	case *tcell.EventResize:
		w, h := e.Size()
		return term.ResizeEvent{Width: w, Height: h}

	case *tcell.EventKey:
		return translateKeyEvent(e)

	case *tcell.EventMouse:
		x, y := e.Position()
		btn := e.Buttons()

		var button term.MouseButton
		switch {
		case btn&tcell.Button1 != 0:
			button = term.MouseLeft
		case btn&tcell.Button2 != 0:
			button = term.MouseRight
		case btn&tcell.Button3 != 0:
			button = term.MouseMiddle
		case btn&tcell.WheelUp != 0:
			button = term.MouseWheelUp
		case btn&tcell.WheelDown != 0:
			button = term.MouseWheelDown
		default:
			return nil // Ignore release/move
		}

		return term.MouseEvent{X: x, Y: y, Button: button}
	}

	return nil
}

func (s *Screen) PushEvent(ev term.Event) {
	switch ev.(type) {
	case term.RedrawEvent:
		// Post a custom event to wake up PollEvent
		s.screen.PostEventWait(tcell.NewEventInterrupt(nil))
	default:
		// Could handle other custom events here
	}
}

func (s *Screen) SetCell(x, y int, ch rune, st term.Style) {
	style := tcell.StyleDefault
	if st.Inverse {
		style = style.Reverse(true)
	}
	s.screen.SetContent(x, y, ch, nil, style)
}

func (s *Screen) Show() {
	s.screen.Show()
}

func (s *Screen) SetCursorShape(shape term.CursorShape) {
	var cursorStyle tcell.CursorStyle
	switch shape {
	case term.CursorBlock:
		cursorStyle = tcell.CursorStyleBlinkingBlock
	case term.CursorUnderline:
		cursorStyle = tcell.CursorStyleBlinkingUnderline
	case term.CursorBar:
		cursorStyle = tcell.CursorStyleBlinkingBar
	default:
		cursorStyle = tcell.CursorStyleDefault
	}
	s.screen.SetCursorStyle(cursorStyle)
}

func (s *Screen) ShowCursor(x, y int) {
	s.screen.ShowCursor(x, y)
}

func (s *Screen) HideCursor() {
	s.screen.HideCursor()
}
