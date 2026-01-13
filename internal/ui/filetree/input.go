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

package filetree

import (
	"cooledit/internal/term"
)

// HandleKey processes a key event and returns an action result
// Returns ActionNone if the key was not handled by the file tree
func (t *FileTree) HandleKey(key term.Key, r rune, mods term.ModMask) ActionResult {
	// Ctrl+B closes the panel
	if key == term.KeyRune && r == 'b' && (mods&term.ModCtrl) != 0 {
		return ActionResult{Action: ActionClosePanel}
	}

	switch key {
	case term.KeyUp:
		t.moveSelection(-1)
		return ActionResult{Action: ActionNone}

	case term.KeyDown:
		t.moveSelection(1)
		return ActionResult{Action: ActionNone}

	case term.KeyLeft:
		t.collapseSelected()
		return ActionResult{Action: ActionNone}

	case term.KeyRight:
		t.expandSelected()
		return ActionResult{Action: ActionNone}

	case term.KeyEnter:
		return t.toggleSelected()

	case term.KeyPageUp:
		t.moveSelection(-10)
		return ActionResult{Action: ActionNone}

	case term.KeyPageDown:
		t.moveSelection(10)
		return ActionResult{Action: ActionNone}

	case term.KeyHome:
		if len(t.visibleItems) > 0 {
			t.selectedIdx = 0
			t.selectedPath = t.visibleItems[0].Node.Path
		}
		return ActionResult{Action: ActionNone}

	case term.KeyEnd:
		if len(t.visibleItems) > 0 {
			t.selectedIdx = len(t.visibleItems) - 1
			t.selectedPath = t.visibleItems[t.selectedIdx].Node.Path
		}
		return ActionResult{Action: ActionNone}
	}

	// Key not handled
	return ActionResult{Action: ActionNone}
}
