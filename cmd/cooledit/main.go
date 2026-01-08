package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

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
				" cooledit | %s%s | %s | %s | Ln %d, Col %d | Ctrl+S save | Esc quit ",
				filename,
				mod,
				encoding,
				eol,
				row+1,
				col+1,
			),
		)
	}

	// ---------- Layout (DECLARE EARLY) ----------

	layout := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(editor, 0, 1, true).
		AddItem(status, 1, 0, false)

	// ---------- File I/O ----------

	loadFile := func(path string) error {
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		editor.SetText(string(data), false)
		filename = path
		modified = false
		updateStatus()
		return nil
	}

	saveFile := func(path string) error {
		data := editor.GetText()
		if err := os.WriteFile(path, []byte(data), 0644); err != nil {
			return err
		}
		filename = path
		modified = false
		updateStatus()
		return nil
	}

	// ---------- Save As dialog ----------

	showSaveAs := func() {
		input := tview.NewInputField().
			SetLabel("Save as: ").
			SetText(filename)

		modal := tview.NewFlex().
			SetDirection(tview.FlexRow).
			AddItem(input, 1, 0, true)

		input.SetDoneFunc(func(key tcell.Key) {
			if key == tcell.KeyEnter {
				path := input.GetText()
				if path != "" {
					if err := saveFile(path); err != nil {
						log.Println("Save failed:", err)
					}
				}
			}
			app.SetRoot(layout, true)
			app.SetFocus(editor)
		})

		app.SetRoot(modal, true)
		app.SetFocus(input)
	}

	// ---------- Callbacks ----------

	editor.SetMovedFunc(updateStatus)

	editor.SetChangedFunc(func() {
		modified = true
		updateStatus()
	})

	editor.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		switch ev.Key() {
		case tcell.KeyEsc:
			app.Stop()
			return nil

		case tcell.KeyCtrlS:
			if filename == "[No Name]" {
				showSaveAs()
			} else {
				if err := saveFile(filename); err != nil {
					log.Println("Save failed:", err)
				}
			}
			return nil
		}
		return ev
	})

	// ---------- Startup ----------

	app.SetRoot(layout, true)
	app.SetFocus(editor)

	if len(os.Args) > 1 {
		path, _ := filepath.Abs(os.Args[1])
		if err := loadFile(path); err != nil {
			log.Println("Open failed:", err)
		}
	}

	updateStatus()

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
