package app

import (
	"cooledit/internal/core"
	"cooledit/internal/fileio"
	"cooledit/internal/term/tcell"
	"cooledit/internal/ui"
)

func Run(path string) error {
	screen := tcell.New()
	if err := screen.Init(); err != nil {
		return err
	}
	defer screen.Fini()

	editor := core.NewEditor()

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
