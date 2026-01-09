package ui

import (
	"cooledit/internal/term"
)

type FakeScreen struct {
	w, h int

	cells map[[2]int]rune

	cursorX, cursorY int
	cursorVisible    bool

	events []term.Event
}

func NewFakeScreen(w, h int) *FakeScreen {
	return &FakeScreen{
		w:     w,
		h:     h,
		cells: make(map[[2]int]rune),
	}
}

func (s *FakeScreen) Init(enableMouse bool) error { return nil }
func (s *FakeScreen) Fini()                       {}

func (s *FakeScreen) Size() (int, int) {
	return s.w, s.h
}

func (s *FakeScreen) PollEvent() term.Event {
	if len(s.events) == 0 {
		return nil
	}
	ev := s.events[0]
	s.events = s.events[1:]
	return ev
}

func (s *FakeScreen) PushEvent(ev term.Event) {
	s.events = append(s.events, ev)
}

func (s *FakeScreen) SetCell(x, y int, ch rune, _ term.Style) {
	if x < 0 || y < 0 || x >= s.w || y >= s.h {
		return
	}
	s.cells[[2]int{x, y}] = ch
}

func (s *FakeScreen) Show() {}

func (s *FakeScreen) SetCursorShape(shape term.CursorShape, color term.Color) {
	// Fake screen doesn't need to track cursor shape/color for tests
}

func (s *FakeScreen) ShowCursor(x, y int) {
	s.cursorX = x
	s.cursorY = y
	s.cursorVisible = true
}

func (s *FakeScreen) HideCursor() {
	s.cursorVisible = false
}

// helpers for tests

func (s *FakeScreen) Cell(x, y int) rune {
	return s.cells[[2]int{x, y}]
}
