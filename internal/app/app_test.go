package app

import (
	"testing"

	"cooledit/internal/config"
	"cooledit/internal/term"
)

type mockScreen struct {
	initCalled bool
	finiCalled bool
	pollCount  int
}

func (m *mockScreen) Init() error {
	m.initCalled = true
	return nil
}

func (m *mockScreen) Fini() {
	m.finiCalled = true
}

func (m *mockScreen) Size() (int, int) {
	return 80, 24
}

func (m *mockScreen) PollEvent() term.Event {
	m.pollCount++
	if m.pollCount == 1 {
		// Ctrl+Q to quit
		return term.KeyEvent{Key: term.KeyRune, Rune: 'q', Modifiers: term.ModCtrl}
	}
	// After first event, just return nil (though it should have exited)
	return nil
}

func (m *mockScreen) PushEvent(ev term.Event) {}

func (m *mockScreen) SetCell(x, y int, ch rune, style term.Style)             {}
func (m *mockScreen) Show()                                                   {}
func (m *mockScreen) SetCursorShape(shape term.CursorShape, color term.Color) {}
func (m *mockScreen) ShowCursor(x, y int)                                     {}
func (m *mockScreen) HideCursor()                                             {}

func TestRunWithScreenBasic(t *testing.T) {
	m := &mockScreen{}
	cfg := config.Default()
	// RunWithScreen will call Run which loops.
	// Our mock returns Ctrl+Q which sets quitNow=true.
	err := RunWithScreen("", false, cfg, m)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if !m.initCalled {
		t.Errorf("Init was not called")
	}
	if !m.finiCalled {
		t.Errorf("Fini was not called")
	}
}
