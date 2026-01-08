package app

import (
	"errors"
	"time"

	"cooledit/internal/term"
	tcellbackend "cooledit/internal/term/tcell"
)

var errQuit = errors.New("quit")

func Run() error {
	screen := tcellbackend.New()
	if err := screen.Init(); err != nil {
		return err
	}
	defer screen.Fini()

	var (
		quitPending bool
		quitAt      time.Time
	)

	for {
		draw(screen, quitPending)

		ev := screen.PollEvent()
		if ev == nil {
			continue
		}

		switch e := ev.(type) {
		case term.ResizeEvent:
			// redraw on next loop iteration
			continue

		case term.KeyEvent:
			if isQuit(e, &quitPending, &quitAt) {
				return nil
			}
		}
	}
}

func draw(s term.Screen, quitPending bool) {
	w, h := s.Size()
	s.HideCursor()

	clearScreen(s, w, h)

	msg := "cooledit - scaffolding OK"
	if quitPending {
		msg = "Press Ctrl+C again to quit"
	}

	for i, r := range msg {
		if i >= w {
			break
		}
		s.SetCell(i, 0, r)
	}

	s.Show()
}

func clearScreen(s term.Screen, w, h int) {
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			s.SetCell(x, y, ' ')
		}
	}
}

func isQuit(e term.KeyEvent, pending *bool, at *time.Time) bool {
	if e.Key == term.KeyRune && e.Rune == 'c' && e.Modifiers&term.ModCtrl != 0 {
		now := time.Now()
		if *pending && now.Sub(*at) < 2*time.Second {
			return true
		}
		*pending = true
		*at = now
		return false
	}

	*pending = false
	return false
}
