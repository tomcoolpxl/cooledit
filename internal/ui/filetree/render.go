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
	"cooledit/internal/theme"
	"strings"
)

const (
	indentWidth   = 2  // Spaces per indent level
	expandPrefix  = "> "
	collapsePrefix = "v "
	filePrefix    = "  "
	symlinkSuffix = " @"
)

// Render draws the file tree to the screen
func (t *FileTree) Render(screen term.Screen, x, y, height int, colors *theme.FileviewColors, isDefaultTheme bool) {
	if height <= 0 || t.width <= 0 {
		return
	}

	// Calculate scroll offset to keep selection visible
	scrollOffset := t.calculateScrollOffset(height - 1) // -1 for header

	// Draw header
	t.drawHeader(screen, x, y, colors, isDefaultTheme)

	// Draw items
	for i := 0; i < height-1; i++ {
		itemIdx := scrollOffset + i
		screenY := y + 1 + i

		if itemIdx < len(t.visibleItems) {
			item := t.visibleItems[itemIdx]
			t.drawItem(screen, x, screenY, item, itemIdx == t.selectedIdx, colors, isDefaultTheme)
		} else {
			// Clear empty lines
			t.clearLine(screen, x, screenY, colors, isDefaultTheme)
		}
	}
}

// calculateScrollOffset calculates the scroll offset to keep selection visible
func (t *FileTree) calculateScrollOffset(visibleLines int) int {
	if len(t.visibleItems) <= visibleLines {
		return 0
	}

	// Keep selection in middle third of visible area when possible
	scrollOffset := t.selectedIdx - visibleLines/2
	if scrollOffset < 0 {
		scrollOffset = 0
	}
	maxOffset := len(t.visibleItems) - visibleLines
	if scrollOffset > maxOffset {
		scrollOffset = maxOffset
	}

	return scrollOffset
}

// drawHeader draws the header line
func (t *FileTree) drawHeader(screen term.Screen, x, y int, colors *theme.FileviewColors, isDefaultTheme bool) {
	style := t.getHeaderStyle(colors, isDefaultTheme)

	// Clear line with header background
	for i := 0; i < t.width; i++ {
		screen.SetCell(x+i, y, ' ', style)
	}

	// Draw header text (truncate if needed)
	label := t.headerLabel
	if len(label) > t.width-2 {
		label = label[:t.width-5] + "..."
	}

	// Center or left-align header
	startX := x + 1
	for i, r := range label {
		if startX+i >= x+t.width-1 {
			break
		}
		screen.SetCell(startX+i, y, r, style)
	}
}

// drawItem draws a single tree item
func (t *FileTree) drawItem(screen term.Screen, x, y int, item VisibleItem, isSelected bool, colors *theme.FileviewColors, isDefaultTheme bool) {
	node := item.Node

	// Determine style
	style := t.getItemStyle(node, isSelected, colors, isDefaultTheme)

	// Check if this is the open file (for underline)
	isOpenFile := node.Path == t.openFilePath

	// Clear line with background
	bgStyle := t.getBackgroundStyle(colors, isDefaultTheme)
	if isSelected {
		bgStyle = t.getSelectionStyle(colors, isDefaultTheme)
	}
	for i := 0; i < t.width; i++ {
		screen.SetCell(x+i, y, ' ', bgStyle)
	}

	// Build the line content
	var line strings.Builder

	// Indentation
	indent := strings.Repeat(" ", item.Depth*indentWidth)
	line.WriteString(indent)

	// Prefix (expand/collapse indicator)
	if node.IsDir {
		if node.Readable {
			if node.Expanded {
				line.WriteString(collapsePrefix)
			} else {
				line.WriteString(expandPrefix)
			}
		} else {
			line.WriteString(filePrefix) // Unreadable dir shown as file-like
		}
	} else {
		line.WriteString(filePrefix)
	}

	// Name
	line.WriteString(node.Name)

	// Symlink indicator
	if node.IsSymlink {
		line.WriteString(symlinkSuffix)
	}

	// Draw the line
	lineStr := line.String()
	col := 0
	for _, r := range lineStr {
		if col >= t.width {
			break
		}

		charStyle := style
		if isOpenFile {
			charStyle.Underline = true
		}

		screen.SetCell(x+col, y, r, charStyle)
		col++
	}
}

// clearLine clears a line with background color
func (t *FileTree) clearLine(screen term.Screen, x, y int, colors *theme.FileviewColors, isDefaultTheme bool) {
	style := t.getBackgroundStyle(colors, isDefaultTheme)
	for i := 0; i < t.width; i++ {
		screen.SetCell(x+i, y, ' ', style)
	}
}

// Style helper methods

func (t *FileTree) getHeaderStyle(colors *theme.FileviewColors, isDefaultTheme bool) term.Style {
	if isDefaultTheme {
		return term.Style{Inverse: true}
	}
	return term.Style{
		Foreground: colors.HeaderFg,
		Background: colors.HeaderBg,
	}
}

func (t *FileTree) getBackgroundStyle(colors *theme.FileviewColors, isDefaultTheme bool) term.Style {
	if isDefaultTheme {
		return term.Style{}
	}
	return term.Style{
		Foreground: colors.Fg,
		Background: colors.Bg,
	}
}

func (t *FileTree) getSelectionStyle(colors *theme.FileviewColors, isDefaultTheme bool) term.Style {
	if isDefaultTheme {
		return term.Style{Inverse: true}
	}
	return term.Style{
		Foreground: colors.SelectionFg,
		Background: colors.SelectionBg,
	}
}

func (t *FileTree) getItemStyle(node *TreeNode, isSelected bool, colors *theme.FileviewColors, isDefaultTheme bool) term.Style {
	if isDefaultTheme {
		if isSelected {
			return term.Style{Inverse: true}
		}
		return term.Style{}
	}

	var fg term.Color
	if isSelected {
		fg = colors.SelectionFg
	} else if node.IsDir {
		fg = colors.DirFg
	} else if node.IsSymlink {
		fg = colors.SymlinkFg
	} else {
		fg = colors.Fg
	}

	var bg term.Color
	if isSelected {
		bg = colors.SelectionBg
	} else {
		bg = colors.Bg
	}

	return term.Style{
		Foreground: fg,
		Background: bg,
	}
}
