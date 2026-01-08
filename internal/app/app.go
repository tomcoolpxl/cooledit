package app

import (
	"cooledit/internal/core"
	"cooledit/internal/term/tcell"
	"cooledit/internal/ui"
)

func Run() error {
	screen := tcell.New()
	if err := screen.Init(); err != nil {
		return err
	}
	defer screen.Fini()

	editor := core.NewEditor()
	u := ui.New(screen, editor)

	return u.Run()
}
