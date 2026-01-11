// Copyright (C) 2026 Tom Cool
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package app

import (
	"cooledit/internal/autosave"
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

	// Check for autosave recovery before loading the file
	hasRecovery := false
	if path != "" && cfg.Autosave.Enabled {
		hasRecovery = autosave.HasRecoveryFor(path)
	}

	if path != "" && !hasRecovery {
		// Normal file loading - no recovery needed
		fd, err := fileio.Open(path)
		if err != nil {
			// If file doesn't exist, set it up as a new file to be created
			editor.SetNewFile(path)
		} else {
			editor.LoadFile(fd)
		}
	} else if path != "" && hasRecovery {
		// Load the file first (we'll offer recovery after)
		fd, err := fileio.Open(path)
		if err != nil {
			editor.SetNewFile(path)
		} else {
			editor.LoadFile(fd)
		}
	}

	u := ui.New(screen, editor, cfg)
	u.SetOptions(lineNumbers, cfg.Editor.SoftWrap)

	// If recovery is available, enter recovery mode before running
	if hasRecovery {
		u.CheckForRecovery(path)
	}

	return u.Run()
}
