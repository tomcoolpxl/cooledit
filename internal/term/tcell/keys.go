package tcell

import (
	"github.com/gdamore/tcell/v2"

	"cooledit/internal/term"
)

func translateKeyEvent(ev *tcell.EventKey) term.KeyEvent {
	// Handle Ctrl+A..Ctrl+Z which tcell reports as distinct keys (not Rune+Ctrl).
	if ev.Key() >= tcell.KeyCtrlA && ev.Key() <= tcell.KeyCtrlZ {
		r := rune('a' + (ev.Key() - tcell.KeyCtrlA))

		mods := term.ModCtrl
		// Keep Alt/Shift if present (some terminals may set them)
		if ev.Modifiers()&tcell.ModAlt != 0 {
			mods |= term.ModAlt
		}
		if ev.Modifiers()&tcell.ModShift != 0 {
			mods |= term.ModShift
		}

		return term.KeyEvent{
			Key:       term.KeyRune,
			Rune:      r,
			Modifiers: mods,
		}
	}

	var k term.Key
	var r rune

	switch ev.Key() {
	case tcell.KeyRune:
		k = term.KeyRune
		r = ev.Rune()

	case tcell.KeyEnter:
		k = term.KeyEnter
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		k = term.KeyBackspace
	case tcell.KeyDelete:
		k = term.KeyDelete
	case tcell.KeyEscape:
		k = term.KeyEscape
	case tcell.KeyTab:
		k = term.KeyTab

	case tcell.KeyUp:
		k = term.KeyUp
	case tcell.KeyDown:
		k = term.KeyDown
	case tcell.KeyLeft:
		k = term.KeyLeft
	case tcell.KeyRight:
		k = term.KeyRight

	case tcell.KeyHome:
		k = term.KeyHome
	case tcell.KeyEnd:
		k = term.KeyEnd
	case tcell.KeyPgUp:
		k = term.KeyPageUp
	case tcell.KeyPgDn:
		k = term.KeyPageDown

	case tcell.KeyF1:
		k = term.KeyF1
	case tcell.KeyF2:
		k = term.KeyF2
	case tcell.KeyF3:
		k = term.KeyF3
	case tcell.KeyF4:
		k = term.KeyF4
	case tcell.KeyF5:
		k = term.KeyF5
	case tcell.KeyF6:
		k = term.KeyF6
	case tcell.KeyF7:
		k = term.KeyF7
	case tcell.KeyF8:
		k = term.KeyF8
	case tcell.KeyF9:
		k = term.KeyF9
	case tcell.KeyF10:
		k = term.KeyF10
	case tcell.KeyF11:
		k = term.KeyF11
	case tcell.KeyF12:
		k = term.KeyF12

	default:
		k = term.KeyUnknown
	}

	var mods term.ModMask
	if ev.Modifiers()&tcell.ModCtrl != 0 {
		mods |= term.ModCtrl
	}
	if ev.Modifiers()&tcell.ModAlt != 0 {
		mods |= term.ModAlt
	}
	if ev.Modifiers()&tcell.ModShift != 0 {
		mods |= term.ModShift
	}

	return term.KeyEvent{
		Key:       k,
		Rune:      r,
		Modifiers: mods,
	}
}
