package app

import (
	"testing"

	"cooledit/internal/term"
)

type mockScreen struct {
	initCalled      bool
	mouseEnabled    bool
	finiCalled      bool
	pollCount       int
}

func (m *mockScreen) Init(enableMouse bool) error {
	m.initCalled = true
	m.mouseEnabled = enableMouse
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

func (m *mockScreen) SetCell(x, y int, ch rune, style term.Style) {}
func (m *mockScreen) Show()                                      {}
func (m *mockScreen) ShowCursor(x, y int)                        {}
func (m *mockScreen) HideCursor()                                {}

func TestRunWithScreenMouseSetting(t *testing.T) {
	t.Run("MouseEnabled", func(t *testing.T) {
		m := &mockScreen{}
		// RunWithScreen will call Run which loops.
		// Our mock returns Ctrl+Q which sets quitNow=true.
		err := RunWithScreen("", true, false, false, m)
		if err != nil {
			t.Fatalf("Run failed: %v", err)
		}
		if !m.initCalled {
			t.Errorf("Init was not called")
		}
		if !m.mouseEnabled {
			t.Errorf("Expected mouse to be enabled")
		}
		if !m.finiCalled {
			t.Errorf("Fini was not called")
		}
	})

	t.Run("MouseDisabled", func(t *testing.T) {
		m := &mockScreen{}
		err := RunWithScreen("", false, false, false, m)
		if err != nil {
			t.Fatalf("Run failed: %v", err)
		}
		if !m.initCalled {
			t.Errorf("Init was not called")
		}
		if m.mouseEnabled {
			t.Errorf("Expected mouse to be disabled")
		}
		if !m.finiCalled {
			t.Errorf("Fini was not called")
		}
	})
}