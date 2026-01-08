package app

import (
	"cooledit/internal/core"
	"cooledit/internal/fileio"
	"cooledit/internal/term"
	"cooledit/internal/term/tcell"
	"cooledit/internal/ui"
)

func Run(path string, enableMouse bool) error {
	return RunWithScreen(path, enableMouse, tcell.New())
}

// RunWithScreen is exported for testing or custom backends
func RunWithScreen(path string, enableMouse bool, screen term.Screen) error {
	if err := screen.Init(enableMouse); err != nil {
		return err
	}
	defer screen.Fini()

	editor := core.NewEditor(&ui.SystemClipboard{})

	if path != "" {
		fd, err := fileio.Open(path)
		if err != nil {
			return err
		}
		editor.LoadFile(fd)
	}

	u := ui.New(screen, editor)
	return u.Run()
}
