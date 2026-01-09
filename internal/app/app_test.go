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
