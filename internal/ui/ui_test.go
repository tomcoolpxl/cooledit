package ui

import (
	"testing"
	"time"

	"cooledit/internal/core"
	"cooledit/internal/term"
)

func newTestUI(w, h int) (*UI, *FakeScreen) {
	screen := NewFakeScreen(w, h)
	editor := core.NewEditor()
	ui := New(screen, editor)
	return ui, screen
}

func TestTypingShowsCursorAndText(t *testing.T) {
	ui, screen := newTestUI(20, 5)

	// Simulate typing 'a'
	cmd := ui.translateKey(term.KeyEvent{
		Key:       term.KeyRune,
		Rune:      'a',
		Modifiers: 0,
	})
	if cmd == nil {
		t.Fatalf("expected insert command")
	}

	ui.editor.Apply(cmd, 4)
	ui.draw(20, 5, 4)

	if !screen.cursorVisible {
		t.Fatalf("cursor should be visible")
	}

	if got := screen.Cell(0, 0); got != 'a' {
		t.Fatalf("expected 'a' at (0,0), got %q", got)
	}

	if screen.cursorX != 1 || screen.cursorY != 0 {
		t.Fatalf("expected cursor at (1,0), got (%d,%d)", screen.cursorX, screen.cursorY)
	}
}

func TestCtrlSSaveAsPrompt(t *testing.T) {
	ui, screen := newTestUI(40, 5)

	// Ctrl+S with no filename → Save As prompt
	cmd := ui.translateKey(term.KeyEvent{
		Key:       term.KeyRune,
		Rune:      's',
		Modifiers: term.ModCtrl,
	})
	if cmd != nil {
		t.Fatalf("expected no command for Ctrl+S when prompting")
	}

	ui.draw(40, 5, 4)

	row := 4
	if got := screen.Cell(0, row); got != 'S' {
		t.Fatalf("expected Save as prompt, got %q", got)
	}
}

func TestMessageExpiresToNormalStatus(t *testing.T) {
	ui, screen := newTestUI(40, 5)

	ui.enterMessage("File saved")
	ui.draw(40, 5, 4)

	row := 4
	if screen.Cell(0, row) != 'F' {
		t.Fatalf("expected message visible")
	}

	// force expiration
	ui.messageUntil = time.Now().Add(-1 * time.Second)

	ui.draw(40, 5, 4)

	if screen.Cell(0, row) == 'F' {
		t.Fatalf("message should have expired")
	}
}

func TestCtrlQCleanQuitSetsFlag(t *testing.T) {
	ui, _ := newTestUI(40, 5)

	cmd := ui.translateKey(term.KeyEvent{
		Key:       term.KeyRune,
		Rune:      'q',
		Modifiers: term.ModCtrl,
	})
	if cmd != nil {
		t.Fatalf("Ctrl+Q should not return a command")
	}

	if !ui.quitNow {
		t.Fatalf("Ctrl+Q on clean editor should set quitNow")
	}
}
