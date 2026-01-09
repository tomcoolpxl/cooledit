package app

import (
	"cooledit/internal/config"
	"cooledit/internal/core"
	"cooledit/internal/fileio"
	"cooledit/internal/term"
	"cooledit/internal/term/tcell"
	"cooledit/internal/ui"
)

func Run(path string, lineNumbers bool, cfg *config.Config) error {
	return RunWithScreen(path, lineNumbers, cfg, tcell.New())
}

// RunWithScreen is exported for testing or custom backends
func RunWithScreen(path string, lineNumbers bool, cfg *config.Config, screen term.Screen) error {
	if err := screen.Init(); err != nil {
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

	u := ui.New(screen, editor, cfg)
	u.SetOptions(lineNumbers, cfg.Editor.SoftWrap)
	return u.Run()
}
