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

func (s *Screen) Init() error {
	ts, err := tcell.NewScreen()
	if err != nil {
		return err
	}
	if err := ts.Init(); err != nil {
		return err
	}
	ts.Clear()
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
	}

	return nil
}

func (s *Screen) SetCell(x, y int, ch rune) {
	s.screen.SetContent(x, y, ch, nil, tcell.StyleDefault)
}

func (s *Screen) Show() {
	s.screen.Show()
}

func (s *Screen) ShowCursor(x, y int) {
	s.screen.ShowCursor(x, y)
}

func (s *Screen) HideCursor() {
	s.screen.HideCursor()
}
