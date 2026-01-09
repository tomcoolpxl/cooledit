package tcell

import (
	"strings"

	"github.com/gdamore/tcell/v2"

	"cooledit/internal/term"
)

type Screen struct {
	screen              tcell.Screen
	originalCursorStyle tcell.CursorStyle
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

	// Save the original cursor style to restore on exit
	// Note: tcell doesn't provide a way to query current cursor style,
	// so we save the default which will be restored on Fini
	s.originalCursorStyle = tcell.CursorStyleDefault

	ts.Clear()
	s.screen = ts
	return nil
}

func (s *Screen) Fini() {
	if s.screen != nil {
		// Restore the original cursor style before exiting
		s.screen.SetCursorStyle(s.originalCursorStyle)
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

	// Apply colors if not using inverse mode
	if st.Inverse {
		// Legacy inverse mode - use reverse video
		style = style.Reverse(true)
	} else {
		// Apply foreground color
		if st.Foreground != "" && st.Foreground != term.ColorDefault {
			style = style.Foreground(parseColor(st.Foreground))
		}
		// Apply background color
		if st.Background != "" && st.Background != term.ColorDefault {
			style = style.Background(parseColor(st.Background))
		}
	}

	// Apply underline if requested
	if st.Underline {
		style = style.Underline(true)
	}

	s.screen.SetContent(x, y, ch, nil, style)
}

// parseColor converts our Color type to tcell.Color
func parseColor(c term.Color) tcell.Color {
	s := string(c)

	// Handle hex colors (#RRGGBB)
	if strings.HasPrefix(s, "#") && len(s) == 7 {
		return tcell.GetColor(s)
	}

	// Handle named colors
	switch c {
	case term.ColorDefault:
		return tcell.ColorDefault
	case term.ColorBlack:
		return tcell.ColorBlack
	case term.ColorRed:
		return tcell.ColorRed
	case term.ColorGreen:
		return tcell.ColorGreen
	case term.ColorYellow:
		return tcell.ColorYellow
	case term.ColorBlue:
		return tcell.ColorBlue
	case term.ColorMagenta:
		return tcell.ColorPurple
	case term.ColorCyan:
		return tcell.ColorTeal
	case term.ColorWhite:
		return tcell.ColorWhite
	default:
		// Try to parse as hex or tcell color name
		return tcell.GetColor(s)
	}
}

func (s *Screen) Show() {
	s.screen.Show()
}

func (s *Screen) SetCursorShape(shape term.CursorShape, color term.Color) {
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

	// Set cursor style with color (terminal support varies)
	tcellColor := parseColor(color)
	s.screen.SetCursorStyle(cursorStyle, tcellColor)
}

func (s *Screen) ShowCursor(x, y int) {
	s.screen.ShowCursor(x, y)
}

func (s *Screen) HideCursor() {
	s.screen.HideCursor()
}
