package main

import (
	"fmt"
	"log"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func main() {
	app := tview.NewApplication()

	editor := tview.NewTextArea()
	editor.SetWrap(false)

	status := tview.NewTextView()
	status.SetDynamicColors(false)

	filename := "[No Name]"
	encoding := "UTF-8"
	eol := "LF"
	modified := false

	updateStatus := func() {
		_, _, row, col := editor.GetCursor()

		mod := ""
		if modified {
			mod = " [Modified]"
		}

		status.SetText(
			fmt.Sprintf(
				" cooledit | %s%s | %s | %s | Ln %d, Col %d | Esc quit | F2 mark saved ",
				filename,
				mod,
				encoding,
				eol,
				row+1,
				col+1,
			),
		)
	}

	editor.SetMovedFunc(func() {
		updateStatus()
	})

	editor.SetChangedFunc(func() {
		modified = true
		updateStatus()
	})

	editor.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		switch ev.Key() {
		case tcell.KeyEsc:
			app.Stop()
			return nil
		case tcell.KeyF2:
			// Temporary: simulate a successful save.
			modified = false
			updateStatus()
			return nil
		}
		return ev
	})

	layout := tview.NewFlex()
	layout.SetDirection(tview.FlexRow)
	layout.AddItem(editor, 0, 1, true)
	layout.AddItem(status, 1, 0, false)

	updateStatus()

	app.SetRoot(layout, true)
	app.SetFocus(editor)

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
