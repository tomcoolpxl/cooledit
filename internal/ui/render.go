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

package ui

import (
	"fmt"

	"cooledit/internal/config"
	"cooledit/internal/core"
	"cooledit/internal/syntax"
	"cooledit/internal/term"
)

// expandTabsToColumn converts a line with tabs into display columns
// Returns the display rune for each screen column
func expandTabsToColumn(line []rune, maxCols int, tabWidth int) []rune {
	if tabWidth <= 0 {
		tabWidth = config.DefaultTabWidth
	}

	result := make([]rune, 0, len(line)*2)
	col := 0

	for _, r := range line {
		if r == '\t' {
			// Tab stops at multiples of tabWidth
			spacesToAdd := tabWidth - (col % tabWidth)
			for i := 0; i < spacesToAdd && col < maxCols; i++ {
				result = append(result, ' ')
				col++
			}
		} else {
			result = append(result, r)
			col++
		}

		if col >= maxCols {
			break
		}
	}

	return result
}

// runeToDisplayCol converts a rune index to display column (accounting for tabs)
func runeToDisplayCol(line []rune, runeIdx int, tabWidth int) int {
	if tabWidth <= 0 {
		tabWidth = config.DefaultTabWidth
	}

	col := 0
	for i := 0; i < runeIdx && i < len(line); i++ {
		if line[i] == '\t' {
			col += tabWidth - (col % tabWidth)
		} else {
			col++
		}
	}
	return col
}

func (u *UI) draw() {
	u.screen.HideCursor()
	u.clear()

	// Set cursor shape and color based on insert/replace mode
	insertShape := ParseCursorShapeWithBlink(u.config.UI.CursorShape, u.config.UI.CursorBlink)
	cursorColor := u.theme.Editor.CursorColor
	if u.insertMode {
		u.screen.SetCursorShape(insertShape, cursorColor)
	} else {
		u.screen.SetCursorShape(GetAlternateCursorShape(insertShape), cursorColor)
	}

	w, h := u.layout.Width, u.layout.Height

	if w < 16 || h < 4 {
		u.drawSmallScreenWarning(w, h)
		u.screen.Show()
		return
	}

	if u.mode == ModeHelp {
		u.drawHelp(w, h)
		u.screen.Show()
		return
	}

	if u.mode == ModeAbout {
		u.drawAbout(w, h)
		u.screen.Show()
		return
	}

	if u.mode == ModeRecovery {
		u.drawRecoveryPrompt(w, h)
		u.screen.Show()
		return
	}

	u.drawMenubar()
	u.drawViewport()
	u.drawStatusBar()
	u.drawPrompt() // Draws prompt or message if active

	if u.menubar.Active {
		u.drawMenuDropdown()
	}

	u.screen.Show()
}

func (u *UI) drawSmallScreenWarning(w, h int) {
	msg := "Screen too small"
	style := u.getStatusStyle()
	for x, r := range msg {
		if x >= w {
			break
		}
		u.screen.SetCell(x, 0, r, style)
	}
}

func (u *UI) drawMenubar() {
	rect := u.layout.Menubar
	if rect.H < 1 {
		return
	}

	style := u.getMenuStyle()
	styleSelected := u.getMenuSelectedStyle()

	// Fill background
	for x := 0; x < rect.W; x++ {
		u.screen.SetCell(rect.X+x, rect.Y, ' ', style)
	}

	x := 0
	for i, menu := range u.menubar.Menus {
		s := style
		if u.menubar.Active && i == u.menubar.SelectedMenuIndex {
			s = styleSelected
		}

		// Draw space before
		u.screen.SetCell(rect.X+x, rect.Y, ' ', s)
		x++

		// Draw menu title with underlined shortcut key
		for j, r := range menu.Title {
			if x >= rect.W {
				break
			}
			cellStyle := s
			// Underline the first character (shortcut key)
			if j == 0 && menu.ShortcutKey != 0 {
				cellStyle.Underline = true
			}
			u.screen.SetCell(rect.X+x, rect.Y, r, cellStyle)
			x++
		}

		// Draw space after
		if x < rect.W {
			u.screen.SetCell(rect.X+x, rect.Y, ' ', s)
			x++
		}
	}
}

func (u *UI) drawMenuDropdown() {
	// Re-calculate X of selected menu
	menuIdx := u.menubar.SelectedMenuIndex
	if menuIdx < 0 || menuIdx >= len(u.menubar.Menus) {
		return
	}

	menuX := 0
	for i := 0; i < menuIdx; i++ {
		menuX += len(u.menubar.Menus[i].Title) + 2 // " Title "
	}

	menu := u.menubar.Menus[menuIdx]
	items := menu.Items

	// Calculate width
	width := 0
	for _, item := range items {
		w := len(item.Label) + 4 + len(item.Accelerator) // "Label    Accel"
		if w > width {
			width = w
		}
	}
	if width < 10 {
		width = 10
	}

	// Draw at (menuX, 1)
	startX := menuX
	startY := 1

	// Ensure fits on screen
	if startX+width > u.layout.Width {
		startX = u.layout.Width - width
	}

	// Calculate available height for menu
	availableHeight := u.layout.Height - startY
	if availableHeight < 1 {
		availableHeight = 1
	}

	// Determine visible range based on scroll offset
	scrollOffset := u.menubar.ScrollOffset
	visibleEnd := scrollOffset + availableHeight
	if visibleEnd > len(items) {
		visibleEnd = len(items)
	}

	canScrollUp := scrollOffset > 0
	canScrollDown := visibleEnd < len(items)

	style := u.getDropdownStyle()
	styleSelected := u.getDropdownSelectedStyle()

	// Draw visible items
	for i := scrollOffset; i < visibleEnd; i++ {
		item := items[i]
		y := startY + (i - scrollOffset)
		if y >= u.layout.Height {
			break
		}

		// Check if separator
		if item.IsSeparator {
			// Draw separator line
			for x := 0; x < width; x++ {
				u.screen.SetCell(startX+x, y, '─', style)
			}
			continue
		}

		s := style
		if i == u.menubar.SelectedItemIndex {
			// Don't highlight readonly items or separators
			if !item.IsReadOnly {
				s = styleSelected
			}
		}

		// Fill line
		for x := 0; x < width; x++ {
			u.screen.SetCell(startX+x, y, ' ', s)
		}

		// Show scroll indicator on first/last line if needed
		if (i == scrollOffset && canScrollUp) || (i == visibleEnd-1 && canScrollDown) {
			indicator := '↑'
			if i == visibleEnd-1 && canScrollDown {
				indicator = '↓'
			}
			u.screen.SetCell(startX+width-1, y, indicator, s)
		}

		// Draw checkmark if item is checkable and checked (but not readonly)
		checkmark := ' '
		labelOffset := 1
		if item.IsCheckable && !item.IsReadOnly && item.IsChecked != nil && item.IsChecked(u) {
			checkmark = '✓'
			u.screen.SetCell(startX, y, checkmark, s)
		}

		// Draw Label
		label := item.Label
		if item.IsReadOnly && item.GetValue != nil {
			label = item.Label + ": " + item.GetValue(u)
		}

		// Find position of shortcut key to underline it
		shortcutPos := -1
		if item.ShortcutKey != 0 {
			for j, r := range label {
				if r == item.ShortcutKey || r == item.ShortcutKey-32 { // case insensitive
					shortcutPos = j
					break
				}
			}
		}

		for j, r := range label {
			if labelOffset+j < width {
				styleToUse := s
				// Underline the shortcut key
				if j == shortcutPos {
					styleToUse = term.Style{
						Foreground: s.Foreground,
						Background: s.Background,
						Inverse:    s.Inverse,
						Underline:  true,
					}
				}
				u.screen.SetCell(startX+labelOffset+j, y, r, styleToUse)
			}
		}

		// Draw Accelerator (Right aligned)
		accelLen := len(item.Accelerator)
		accelStart := width - 1 - accelLen
		for j, r := range item.Accelerator {
			if accelStart+j < width {
				u.screen.SetCell(startX+accelStart+j, y, r, s)
			}
		}
	}
}

func (u *UI) drawViewport() {
	vpRect := u.layout.Viewport
	if vpRect.H < 1 {
		return
	}

	u.editor.EnsureVisible(vpRect.W, vpRect.H)
	vp := u.editor.Viewport()
	lines := u.editor.Lines()

	sl, sc, el, ec := u.editor.GetSelectionRange()
	hasSelection := u.editor.HasSelection()

	gutterWidth := 0
	if u.showLineNumbers {
		totalLines := len(lines)
		if totalLines == 0 {
			totalLines = 1
		}
		gutterWidth = len(fmt.Sprintf("%d", totalLines)) + 1 // +1 for padding
	}

	availW := vpRect.W - gutterWidth
	if availW < 0 {
		availW = 0
	}

	if u.softWrap {
		u.drawViewportWrapped(vpRect, gutterWidth, availW, lines, vp, sl, sc, el, ec, hasSelection)
	} else {
		u.drawViewportNoWrap(vpRect, gutterWidth, availW, lines, vp, sl, sc, el, ec, hasSelection)
	}
}

func (u *UI) drawViewportNoWrap(vpRect Rect, gutterWidth, availW int, lines [][]rune, vp core.Viewport, sl, sc, el, ec int, hasSelection bool) {
	cursorLine, _ := u.editor.Cursor()
	// Check if current line highlight is enabled and has a valid color (not ColorDefault)
	currentLineBgEnabled := u.currentLineHighlight && u.theme.Editor.CurrentLineBg != term.ColorDefault

	for sy := 0; sy < vpRect.H; sy++ {
		docY := vp.TopLine + sy
		isCurrentLine := (docY == cursorLine) && currentLineBgEnabled

		// Draw Gutter
		gutterStyle := u.getLineNumberStyle()
		if isCurrentLine {
			// Apply current line background to gutter
			gutterStyle.Background = u.theme.Editor.CurrentLineBg
		}

		if u.showLineNumbers {
			if docY < len(lines) {
				numStr := fmt.Sprintf("%d", docY+1) // 1-based
				// Right align
				padding := gutterWidth - len(numStr) - 1
				for i := 0; i < padding; i++ {
					u.screen.SetCell(vpRect.X+i, vpRect.Y+sy, ' ', gutterStyle)
				}
				for i, r := range numStr {
					u.screen.SetCell(vpRect.X+padding+i, vpRect.Y+sy, r, gutterStyle)
				}
				u.screen.SetCell(vpRect.X+gutterWidth-1, vpRect.Y+sy, ' ', gutterStyle)
			} else {
				// Empty gutter
				for i := 0; i < gutterWidth; i++ {
					u.screen.SetCell(vpRect.X+i, vpRect.Y+sy, ' ', gutterStyle)
				}
			}
		}

		if docY < 0 || docY >= len(lines) {
			continue
		}

		line := lines[docY]

		// Expand tabs to display columns using TabWidth setting
		tabWidth := u.editor.TabWidth
		if tabWidth <= 0 {
			tabWidth = 4
		}
		expanded := expandTabsToColumn(line, vp.LeftCol+availW, tabWidth)

		drawX := vpRect.X + gutterWidth

		editorStyle := u.getEditorStyle()
		if isCurrentLine {
			// Override background for current line
			editorStyle.Background = u.theme.Editor.CurrentLineBg
		}
		selectionStyle := u.getSelectionStyle()

		// Draw from LeftCol in display space
		for sx := 0; sx < availW; sx++ {
			displayCol := vp.LeftCol + sx

			if displayCol >= len(expanded) {
				// Past end of line
				break
			}

			// Find which rune index this display column corresponds to
			runeIdx := 0
			col := 0
			isFirstColOfTab := false
			tabOffsetInExpansion := -1 // Position within tab expansion (0=first, 1=second, etc.)
			for runeIdx < len(line) {
				if line[runeIdx] == '\t' {
					nextStop := ((col / tabWidth) + 1) * tabWidth
					if displayCol < nextStop {
						// We're inside this tab expansion
						isFirstColOfTab = (displayCol == col)
						tabOffsetInExpansion = displayCol - col
						break
					}
					col = nextStop
				} else {
					if col == displayCol {
						break
					}
					col++
				}
				runeIdx++
			}

			isSelected := false
			if hasSelection {
				if docY > sl && docY < el {
					isSelected = true
				} else if docY == sl && docY == el {
					if runeIdx >= sc && runeIdx < ec {
						isSelected = true
					}
				} else if docY == sl {
					if runeIdx >= sc {
						isSelected = true
					}
				} else if docY == el {
					if runeIdx < ec {
						isSelected = true
					}
				}
			}

			style := editorStyle

			// Apply syntax highlighting
			if syntaxStyle := u.getSyntaxStyle(docY, runeIdx, line); syntaxStyle != nil {
				style = *syntaxStyle
				// Preserve current line background if enabled
				if isCurrentLine {
					style.Background = u.theme.Editor.CurrentLineBg
				}
			}

			// Apply search match highlighting (overrides syntax)
			if searchStyle := u.getSearchMatchStyle(docY, runeIdx); searchStyle != nil {
				style = *searchStyle
			}

			// Apply bracket matching highlight (overrides search)
			if bracketStyle := u.getBracketStyle(docY, runeIdx); bracketStyle != nil {
				style = *bracketStyle
			}

			// Selection overrides all
			if isSelected {
				style = selectionStyle
			}

			// Draw the expanded character
			ch := expanded[displayCol]
			if u.showWhitespace {
				// Show visible representations of whitespace
				if runeIdx < len(line) && line[runeIdx] == '\t' {
					if isFirstColOfTab {
						ch = '→' // Tab: show arrow on first column
					} else if tabOffsetInExpansion == 1 {
						ch = ' ' // Second column: render as space (no dot)
					} else {
						ch = '·' // Remaining columns: show dots
					}
				} else if ch == ' ' {
					ch = '·' // Regular space character
				}
			}
			u.screen.SetCell(drawX+sx, vpRect.Y+sy, ch, style)
		}

		// Highlight newline at end of line if selected, or show line ending marker
		expandedLen := len(expanded)
		if expandedLen >= vp.LeftCol && expandedLen < vp.LeftCol+availW {
			sx := expandedLen - vp.LeftCol
			if hasSelection && docY >= sl && docY < el {
				u.screen.SetCell(drawX+sx, vpRect.Y+sy, ' ', selectionStyle)
			} else if u.showWhitespace {
				// Show line ending marker
				eol := u.editor.File().EOL
				marker := '↵' // LF (Unix/Linux/Mac)
				if eol == "\r\n" {
					marker = '⏎' // CRLF (Windows) - Return symbol
				}
				u.screen.SetCell(drawX+sx, vpRect.Y+sy, marker, editorStyle)
			}
		}

		// Fill remaining space with current line background if enabled
		if isCurrentLine {
			startCol := expandedLen - vp.LeftCol
			if startCol < 0 {
				startCol = 0
			}
			for sx := startCol; sx < availW; sx++ {
				// Don't overwrite already drawn content
				if expandedLen > 0 && sx < expandedLen-vp.LeftCol {
					continue
				}
				u.screen.SetCell(drawX+sx, vpRect.Y+sy, ' ', editorStyle)
			}
		}
	}

	// Draw cursor
	if u.mode == ModeNormal || u.mode == ModeMessage {
		cy, cx := u.editor.Cursor()

		// Convert cursor rune position to display column
		if cy >= 0 && cy < len(lines) {
			tabWidth := u.editor.TabWidth
			if tabWidth <= 0 {
				tabWidth = 4
			}
			displayCol := runeToDisplayCol(lines[cy], cx, tabWidth)
			sx := displayCol - vp.LeftCol
			sy := cy - vp.TopLine

			drawX := vpRect.X + gutterWidth

			if sx >= 0 && sx < availW && sy >= 0 && sy < vpRect.H {
				u.screen.ShowCursor(drawX+sx, vpRect.Y+sy)
			}
		}
	}
}

// drawViewportWrapped renders the viewport with soft wrap enabled
func (u *UI) drawViewportWrapped(vpRect Rect, gutterWidth, availW int, lines [][]rune, vp core.Viewport, sl, sc, el, ec int, hasSelection bool) {
	if availW <= 0 {
		return
	}

	tabWidth := u.editor.TabWidth
	if tabWidth <= 0 {
		tabWidth = config.DefaultTabWidth
	}

	// Build wrapped lines structure with tab expansion
	type wrappedLine struct {
		lineNum     int    // Original line number
		startRune   int    // Start rune index in original line
		startCol    int    // Start display column
		content     []rune // Wrapped segment (expanded)
		runeIndices []int  // Rune index for each display column
	}

	var wrapped []wrappedLine
	for lineNum, line := range lines {
		if len(line) == 0 {
			wrapped = append(wrapped, wrappedLine{
				lineNum:     lineNum,
				startRune:   0,
				startCol:    0,
				content:     []rune{},
				runeIndices: []int{},
			})
			continue
		}

		// Expand tabs and wrap by display columns
		expanded := expandTabsToColumn(line, len(line)*tabWidth, tabWidth)

		// Build rune index mapping (display col -> rune index)
		runeIndices := make([]int, len(expanded))
		runeIdx := 0
		col := 0
		for runeIdx < len(line) {
			if line[runeIdx] == '\t' {
				spacesToAdd := tabWidth - (col % tabWidth)
				for i := 0; i < spacesToAdd && col < len(expanded); i++ {
					runeIndices[col] = runeIdx
					col++
				}
			} else {
				if col < len(expanded) {
					runeIndices[col] = runeIdx
				}
				col++
			}
			runeIdx++
		}

		// Wrap expanded line into segments
		for startCol := 0; startCol < len(expanded); startCol += availW {
			endCol := startCol + availW
			if endCol > len(expanded) {
				endCol = len(expanded)
			}

			startRune := 0
			if startCol < len(runeIndices) {
				startRune = runeIndices[startCol]
			}

			wrapped = append(wrapped, wrappedLine{
				lineNum:     lineNum,
				startRune:   startRune,
				startCol:    startCol,
				content:     expanded[startCol:endCol],
				runeIndices: runeIndices[startCol:endCol],
			})
		}
	}

	// Draw wrapped lines starting from vp.TopLine
	cursorLine, _ := u.editor.Cursor()
	gutterStyle := u.getLineNumberStyle()
	editorStyle := u.getEditorStyle()
	selectionStyle := u.getSelectionStyle()

	for sy := 0; sy < vpRect.H; sy++ {
		wrappedIdx := vp.TopLine + sy

		// Determine if this is the current line for highlighting
		isCurrentLine := false
		if wrappedIdx >= 0 && wrappedIdx < len(wrapped) {
			isCurrentLine = (wrapped[wrappedIdx].lineNum == cursorLine) && u.currentLineHighlight
		}

		// Draw Gutter - show line number for first wrap of each line
		currentGutterStyle := gutterStyle
		if isCurrentLine {
			currentGutterStyle.Background = u.theme.Editor.CurrentLineBg
		}

		if u.showLineNumbers {
			shouldShowNum := false
			lineNum := 0
			if wrappedIdx >= 0 && wrappedIdx < len(wrapped) {
				lineNum = wrapped[wrappedIdx].lineNum
				// Check if this is the first wrapped segment of this line
				if wrappedIdx == 0 || wrapped[wrappedIdx-1].lineNum != lineNum {
					shouldShowNum = true
				}
			}

			if shouldShowNum {
				numStr := fmt.Sprintf("%d", lineNum+1) // 1-based
				padding := gutterWidth - len(numStr) - 1
				for i := 0; i < padding; i++ {
					u.screen.SetCell(vpRect.X+i, vpRect.Y+sy, ' ', currentGutterStyle)
				}
				for i, r := range numStr {
					u.screen.SetCell(vpRect.X+padding+i, vpRect.Y+sy, r, currentGutterStyle)
				}
				u.screen.SetCell(vpRect.X+gutterWidth-1, vpRect.Y+sy, ' ', currentGutterStyle)
			} else {
				// Empty gutter
				for i := 0; i < gutterWidth; i++ {
					u.screen.SetCell(vpRect.X+i, vpRect.Y+sy, ' ', currentGutterStyle)
				}
			}
		}

		if wrappedIdx < 0 || wrappedIdx >= len(wrapped) {
			continue
		}

		wLine := wrapped[wrappedIdx]
		drawX := vpRect.X + gutterWidth

		// Apply current line background if needed
		currentEditorStyle := editorStyle
		if isCurrentLine {
			currentEditorStyle.Background = u.theme.Editor.CurrentLineBg
		}

		// Draw the wrapped line segment (already expanded)
		for sx, r := range wLine.content {
			if sx >= availW {
				break
			}

			docY := wLine.lineNum

			// Get the rune index for this display position
			runeIdx := wLine.startRune
			if sx < len(wLine.runeIndices) {
				runeIdx = wLine.runeIndices[sx]
			}

			isSelected := false
			if hasSelection {
				if docY > sl && docY < el {
					isSelected = true
				} else if docY == sl && docY == el {
					if runeIdx >= sc && runeIdx < ec {
						isSelected = true
					}
				} else if docY == sl {
					if runeIdx >= sc {
						isSelected = true
					}
				} else if docY == el {
					if runeIdx < ec {
						isSelected = true
					}
				}
			}

			style := currentEditorStyle

			// Apply syntax highlighting
			if docY >= 0 && docY < len(lines) {
				if syntaxStyle := u.getSyntaxStyle(docY, runeIdx, lines[docY]); syntaxStyle != nil {
					style = *syntaxStyle
					// Preserve current line background if enabled
					if isCurrentLine {
						style.Background = u.theme.Editor.CurrentLineBg
					}
				}
			}

			// Apply search match highlighting (overrides syntax)
			if searchStyle := u.getSearchMatchStyle(docY, runeIdx); searchStyle != nil {
				style = *searchStyle
			}

			// Apply bracket matching highlight (overrides search)
			if bracketStyle := u.getBracketStyle(docY, runeIdx); bracketStyle != nil {
				style = *bracketStyle
			}

			// Selection overrides all
			if isSelected {
				style = selectionStyle
			}

			u.screen.SetCell(drawX+sx, vpRect.Y+sy, r, style)
		}

		// Highlight newline at end of last wrap segment for this line if selected
		if wrappedIdx+1 >= len(wrapped) || wrapped[wrappedIdx+1].lineNum != wLine.lineNum {
			// This is the last segment of this line
			if hasSelection && wLine.lineNum >= sl && wLine.lineNum < el {
				endX := len(wLine.content)
				if endX < availW {
					u.screen.SetCell(drawX+endX, vpRect.Y+sy, ' ', selectionStyle)
				}
			}
		}

		// Fill remaining space with current line background if enabled
		if isCurrentLine {
			contentLen := len(wLine.content)
			for sx := contentLen; sx < availW; sx++ {
				u.screen.SetCell(drawX+sx, vpRect.Y+sy, ' ', currentEditorStyle)
			}
		}
	}

	// Draw cursor
	if u.mode == ModeNormal || u.mode == ModeMessage {
		cy, cx := u.editor.Cursor()

		// Convert cursor rune position to display column
		displayCol := 0
		if cy >= 0 && cy < len(lines) {
			displayCol = runeToDisplayCol(lines[cy], cx, tabWidth)
		}

		// Find which wrapped line contains this cursor position
		for wrappedIdx, wLine := range wrapped {
			if wLine.lineNum == cy {
				// Check if cursor display column is in this segment
				segmentEndCol := wLine.startCol + len(wLine.content)

				if displayCol >= wLine.startCol && displayCol < segmentEndCol {
					sx := displayCol - wLine.startCol
					sy := wrappedIdx - vp.TopLine

					drawX := vpRect.X + gutterWidth

					if sy >= 0 && sy < vpRect.H && sx >= 0 && sx < availW {
						u.screen.ShowCursor(drawX+sx, vpRect.Y+sy)
						return
					}
				} else if displayCol == segmentEndCol {
					// Cursor at end of this segment
					isLastSegment := (wrappedIdx+1 >= len(wrapped) || wrapped[wrappedIdx+1].lineNum != cy)
					if isLastSegment || len(wLine.content) < availW {
						sx := len(wLine.content)
						sy := wrappedIdx - vp.TopLine

						drawX := vpRect.X + gutterWidth

						if sy >= 0 && sy < vpRect.H && sx >= 0 && sx < availW {
							u.screen.ShowCursor(drawX+sx, vpRect.Y+sy)
							return
						}
					}
				}
			}
		}
	}
}

// drawSearchStatus renders an enhanced status bar for ModeSearch (unified incremental search mode).
// Shows: search query, match count, current match index, case sensitivity, shortcuts
// Format: Find: query | Match 3 of 15 | Match Case | Alt+C Alt+W | F3 R Esc | Ln 42, Col 8
func (u *UI) drawSearchStatus(rect Rect, style term.Style) {
	// Determine if we're in an error state (no matches found)
	searchState := u.editor.SearchState()
	session := searchState.Session

	query := string(u.searchQuery)

	// Error state: query is non-empty but no matches found (and not currently searching)
	if query != "" && !u.searchIsSearching && session != nil && session.Query != "" && len(session.Matches) == 0 {
		// Use error background color
		style = term.Style{Foreground: style.Foreground, Background: u.theme.Search.ErrorBg}
	}

	// Background
	for x := 0; x < rect.W; x++ {
		u.screen.SetCell(rect.X+x, rect.Y, ' ', style)
	}

	if query == "" {
		query = "..."
	}

	// Build left section: "Find: query | Match X of Y | Match Case"
	left := fmt.Sprintf("Find: %s", query)

	// Show "Searching..." if a search is in progress
	if u.searchIsSearching {
		left += " | Searching..."
	} else if session != nil && len(session.Matches) > 0 {
		// Add match count if we have a session
		currentIdx := session.CurrentIndex + 1 // 1-based for display
		totalMatches := len(session.Matches)
		if session.LimitReached {
			left += fmt.Sprintf(" | Match %d of %d+", currentIdx, totalMatches)
		} else {
			left += fmt.Sprintf(" | Match %d of %d", currentIdx, totalMatches)
		}
	} else if session != nil && session.Query != "" && len(session.Matches) == 0 {
		left += " | No matches"
	}

	// Add case sensitivity indicator
	if searchState != nil && searchState.CaseSensitive {
		left += " | Match Case"
	} else {
		left += " | Ignore Case"
	}

	// Add whole word indicator if active
	if searchState != nil && searchState.WholeWord {
		left += " | Whole Word"
	}

	// Build right section: cursor position
	cy, cx := u.editor.Cursor()
	fs := u.editor.File()
	eol := "LF"
	if fs.EOL == "\r\n" {
		eol = "CRLF"
	}
	right := fmt.Sprintf("Ln %d, Col %d  %s %s", cy+1, cx+1, fs.Encoding, eol)

	// Priority 1: Draw right section
	startRight := rect.W - len(right)
	if startRight < 0 {
		startRight = 0
	}
	for i, r := range right {
		x := startRight + i
		if x >= 0 && x < rect.W {
			u.screen.SetCell(rect.X+x, rect.Y, r, style)
		}
	}

	// Priority 2: Draw left section (truncate if needed)
	maxLeft := startRight - 1
	if maxLeft < 0 {
		maxLeft = 0
	}
	leftEnd := len(left)
	if leftEnd > maxLeft {
		leftEnd = maxLeft
	}
	for i, r := range left {
		if i >= leftEnd {
			break
		}
		u.screen.SetCell(rect.X+i, rect.Y, r, style)
	}

	// Priority 3: Draw shortcuts in center if space allows
	shortcuts := "Alt+C:Case  Alt+W:Word  F3:Next  R:Replace  Esc:Exit"
	availStart := leftEnd + 2
	availEnd := startRight - 2
	availWidth := availEnd - availStart

	if availWidth > len(shortcuts) {
		// Center the shortcuts
		centerX := availStart + (availWidth-len(shortcuts))/2
		for i, r := range shortcuts {
			x := centerX + i
			if x >= availStart && x < availEnd {
				u.screen.SetCell(rect.X+x, rect.Y, r, style)
			}
		}
	} else if availWidth > 20 {
		// Abbreviated shortcuts for medium screens
		shortcutsAbbr := "Alt+C  Alt+W  F3  R  Esc"
		centerX := availStart + (availWidth-len(shortcutsAbbr))/2
		for i, r := range shortcutsAbbr {
			x := centerX + i
			if x >= availStart && x < availEnd {
				u.screen.SetCell(rect.X+x, rect.Y, r, style)
			}
		}
	}
}

func (u *UI) drawStatusBar() {
	rect := u.layout.StatusBar
	if rect.H < 1 {
		return
	}

	style := u.getStatusStyle()

	// Background
	for x := 0; x < rect.W; x++ {
		u.screen.SetCell(rect.X+x, rect.Y, ' ', style)
	}

	// Special status bar for vim command mode
	if u.mode == ModeVimCommand {
		vimCmd := ":" + string(u.vimCommand)
		for i, r := range vimCmd {
			if i >= rect.W {
				break
			}
			u.screen.SetCell(rect.X+i, rect.Y, r, style)
		}
		// Show cursor after the text
		if len(vimCmd) < rect.W {
			u.screen.ShowCursor(rect.X+len(vimCmd), rect.Y)
		}
		return
	}

	// Enhanced status bar for search mode (unified incremental search)
	if u.mode == ModeSearch {
		u.drawSearchStatus(rect, style)
		return
	}

	// Normal mode status bar
	var left string
	fs := u.editor.File()
	mod := ""
	if u.editor.Modified() {
		mod = "*"
	}
	left = fmt.Sprintf("%s%s", fs.BaseName, mod)

	cy, cx := u.editor.Cursor()
	eol := "LF"
	if fs.EOL == "\r\n" {
		eol = "CRLF"
	}

	// Add replace mode indicator to right section
	modeIndicator := ""
	if !u.insertMode {
		modeIndicator = "  REPLACE"
	}

	// Add language indicator
	lang := u.GetCurrentLanguage()

	right := fmt.Sprintf("%s%s  Ln %d, Col %d  %s %s", lang, modeIndicator, cy+1, cx+1, fs.Encoding, eol)

	// Priority 1: Draw Right (position and status)
	startRight := rect.W - len(right)
	if startRight < 0 {
		startRight = 0
	}
	for i, r := range right {
		x := startRight + i
		if x >= 0 && x < rect.W {
			u.screen.SetCell(rect.X+x, rect.Y, r, style)
		}
	}

	// Priority 2: Draw Left (filename)
	maxLeft := startRight - 1
	if maxLeft < 0 {
		maxLeft = 0
	}
	leftEnd := len(left)
	if leftEnd > maxLeft {
		leftEnd = maxLeft
	}
	for i, r := range left {
		if i >= leftEnd {
			break
		}
		u.screen.SetCell(rect.X+i, rect.Y, r, style)
	}

	// Priority 3: Draw centered mini-help
	// Build mini-help with priority from left to right
	miniHelp := []string{
		"F1 Help",
		"Esc/F10 Menu",
		"Ctrl+Q Quit",
		"Ctrl+S Save",
		"Ctrl+F Find/Replace",
	}

	// Calculate available space for center section
	availStart := leftEnd + 2
	availEnd := startRight - 2
	availWidth := availEnd - availStart

	if availWidth > 0 {
		// Build help string with available space
		var helpParts []string
		helpLen := 0
		for _, part := range miniHelp {
			newLen := helpLen
			if len(helpParts) > 0 {
				newLen += 3 // "  " separator
			}
			newLen += len(part)

			if newLen <= availWidth {
				helpParts = append(helpParts, part)
				helpLen = newLen
			} else {
				break
			}
		}

		if len(helpParts) > 0 {
			// Join with separators
			helpText := ""
			for i, part := range helpParts {
				if i > 0 {
					helpText += "  "
				}
				helpText += part
			}

			// Center the help text
			centerX := availStart + (availWidth-len(helpText))/2
			if centerX < availStart {
				centerX = availStart
			}

			// Draw centered help
			for i, r := range helpText {
				x := centerX + i
				if x >= availStart && x < availEnd {
					u.screen.SetCell(rect.X+x, rect.Y, r, style)
				}
			}
		}
	}
}

func (u *UI) drawPrompt() {
	rect := u.layout.Prompt
	if rect.H < 1 {
		return
	}

	style := u.getPromptStyle()

	// Clear prompt area
	for x := 0; x < rect.W; x++ {
		u.screen.SetCell(rect.X+x, rect.Y, ' ', style)
	}

	if u.mode == ModeMessage {
		for i, r := range u.message {
			if i >= rect.W {
				break
			}
			u.screen.SetCell(rect.X+i, rect.Y, r, style)
		}
		return
	}

	if u.mode == ModePrompt {
		text := u.promptLabel + string(u.promptText)
		for i, r := range text {
			if i >= rect.W {
				break
			}
			u.screen.SetCell(rect.X+i, rect.Y, r, style)
		}

		// Show cursor in prompt
		cx := len(u.promptLabel) + len(u.promptText)
		if cx < rect.W {
			u.screen.ShowCursor(rect.X+cx, rect.Y)
		}
	}
}

func (u *UI) drawHelp(w, h int) {
	leftCol := []string{
		"  cooledit - Quick Reference",
		"",
		"  MENU & HELP",
		"    F10, Esc      Menu bar",
		"    F1            This help",
		"    F11           Toggle status bar (Zen mode)",
		"",
		"  FILE",
		"    Ctrl+S        Save",
		"    Ctrl+Shift+S  Save as",
		"    Ctrl+Q        Quit",
		"",
		"  EDIT",
		"    Ctrl+Z        Undo",
		"    Ctrl+Y        Redo",
		"    Ctrl+X        Cut",
		"    Ctrl+C        Copy",
		"    Ctrl+V        Paste",
		"    Ctrl+A        Select all",
		"    Insert        Insert/Replace mode",
		"    Tab           Insert spaces to tab stop",
		"    Backspace     Delete one character",
		"    Ctrl+I        Insert literal tab (\\t)",
	}

	rightCol := []string{
		"",
		"",
		"  SEARCH",
		"    Ctrl+F        Find/Replace (unified search mode)",
		"    Alt+C         Toggle case sensitivity",
		"    Alt+W         Toggle whole word",
		"    F3, Shift+F3  Next/Previous match",
		"    Enter         Next match (in search mode)",
		"    Up/Down       Search history (in search mode)",
		"    Ctrl+R        Replace current match",
		"    Ctrl+H        Replace all matches",
		"    Escape        Exit search",
		"    Ctrl+G        Go to line",
		"",
		"  NAVIGATION",
		"    Arrows        Move cursor",
		"    Ctrl+Arrows   Jump word",
		"    Ctrl+B        Jump bracket",
		"    Home/End      Line start/end",
		"    Ctrl+Home     File start",
		"    Ctrl+End      File end",
		"    PgUp/PgDn     Page up/down",
		"",
		"  VIEW",
		"    Ctrl+L        Line numbers",
		"    Ctrl+W        Word wrap",
	}

	footer := "  Press any key to close"

	style := u.getHelpStyle()
	titleStyle := u.getHelpTitleStyle()

	// Determine if we can use two columns (need at least 80 width)
	useTwoColumns := w >= 80

	if useTwoColumns {
		// Two-column layout - calculate left column width
		colWidth := 0
		for _, line := range leftCol {
			if len(line) > colWidth {
				colWidth = len(line)
			}
		}
		colWidth += 2 // Add spacing between columns

		maxLines := len(leftCol)
		if len(rightCol) > maxLines {
			maxLines = len(rightCol)
		}

		for y := 0; y < maxLines && y < h-1; y++ {
			// Left column
			if y < len(leftCol) {
				line := leftCol[y]
				s := style
				if y == 0 || (len(line) > 2 && line[2] != ' ' && line != "") {
					s = titleStyle
				}
				for x, r := range line {
					if x >= colWidth {
						break
					}
					u.screen.SetCell(x, y, r, s)
				}
			}

			// Right column
			if y < len(rightCol) {
				line := rightCol[y]
				s := style
				if len(line) > 2 && line[2] != ' ' && line != "" {
					s = titleStyle
				}
				for x, r := range line {
					if colWidth+x >= w {
						break
					}
					u.screen.SetCell(colWidth+x, y, r, s)
				}
			}
		}

		// Footer
		if h > 0 {
			for x, r := range footer {
				if x >= w {
					break
				}
				u.screen.SetCell(x, h-1, r, style)
			}
		}
	} else {
		// Single column layout with pagination
		allLines := append([]string{}, leftCol...)
		allLines = append(allLines, "")
		allLines = append(allLines, rightCol...)
		allLines = append(allLines, "")
		allLines = append(allLines, footer)

		// Show as many lines as fit
		for y := 0; y < len(allLines) && y < h; y++ {
			line := allLines[y]
			s := style
			if y == 0 || (len(line) > 2 && line[2] != ' ' && line != "") {
				s = titleStyle
			}
			for x, r := range line {
				if x >= w {
					break
				}
				u.screen.SetCell(x, y, r, s)
			}
		}

		// If content doesn't fit, show indicator
		if len(allLines) > h && h > 0 {
			indicator := " (scroll down for more)"
			startX := w - len(indicator)
			if startX < 0 {
				startX = 0
			}
			for i, r := range indicator {
				if startX+i < w {
					u.screen.SetCell(startX+i, h-1, r, titleStyle)
				}
			}
		}
	}
}

func (u *UI) drawAbout(w, h int) {
	lines := []string{
		"",
		"  cooledit",
		"",
		"  A terminal text editor",
		"",
		"  Copyright (C) 2026 Tom Cool",
		"",
		"  This program is free software licensed under GPL-3.0.",
		"  You can redistribute it and/or modify it under the terms",
		"  of the GNU General Public License as published by the",
		"  Free Software Foundation, either version 3 of the License,",
		"  or (at your option) any later version.",
		"",
		"  See the LICENSE file for the full license text.",
		"",
		"  https://github.com/tomcoolpxl/cooledit",
		"",
		"",
		"  Press any key to close",
	}

	style := u.getHelpStyle()
	titleStyle := u.getHelpTitleStyle()

	// Center vertically if space allows
	startY := 0
	if h > len(lines) {
		startY = (h - len(lines)) / 2
	}

	for i, line := range lines {
		y := startY + i
		if y >= h {
			break
		}

		s := style
		// Title line uses title style
		if i == 1 {
			s = titleStyle
		}

		// Center horizontally
		startX := 0
		if w > len(line) {
			startX = (w - len(line)) / 2
		}

		for x, r := range line {
			if startX+x >= w {
				break
			}
			u.screen.SetCell(startX+x, y, r, s)
		}
	}
}

func (u *UI) clear() {
	w, h := u.layout.Width, u.layout.Height
	bgStyle := u.getEditorStyle()
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			u.screen.SetCell(x, y, ' ', bgStyle)
		}
	}
}

// Theme style helpers - use inverse mode for "default" theme, colors for others
func (u *UI) isDefaultTheme() bool {
	return u.theme.Name == "default"
}

func (u *UI) getEditorStyle() term.Style {
	if u.isDefaultTheme() {
		return term.Style{Foreground: term.ColorDefault, Background: term.ColorDefault}
	}
	return term.Style{Foreground: u.theme.Editor.Fg, Background: u.theme.Editor.Bg}
}

func (u *UI) getSelectionStyle() term.Style {
	if u.isDefaultTheme() {
		return term.Style{Inverse: true}
	}
	return term.Style{Foreground: u.theme.Editor.SelectionFg, Background: u.theme.Editor.SelectionBg}
}

func (u *UI) getLineNumberStyle() term.Style {
	return term.Style{Foreground: u.theme.Editor.LineNumbersFg, Background: u.theme.Editor.LineNumbersBg}
}

// getSyntaxStyle returns the style for a character at the given position.
// Returns nil if syntax highlighting is disabled or no token applies.
func (u *UI) getSyntaxStyle(docY, runeIdx int, line []rune) *term.Style {
	if !u.syntaxHighlighting || u.syntaxCache == nil {
		return nil
	}

	tokens := u.syntaxCache.GetTokens(docY, line)
	tokenType := syntax.GetTokenAt(tokens, runeIdx)

	if tokenType == syntax.TokenNone {
		return nil
	}

	return u.getStyleForToken(tokenType)
}

// getStyleForToken returns the style for a specific token type
func (u *UI) getStyleForToken(t syntax.TokenType) *term.Style {
	var fg, bg term.Color

	switch t {
	case syntax.TokenKeyword:
		fg, bg = u.theme.Syntax.KeywordFg, u.theme.Syntax.KeywordBg
	case syntax.TokenString:
		fg, bg = u.theme.Syntax.StringFg, u.theme.Syntax.StringBg
	case syntax.TokenComment:
		fg, bg = u.theme.Syntax.CommentFg, u.theme.Syntax.CommentBg
	case syntax.TokenNumber:
		fg, bg = u.theme.Syntax.NumberFg, u.theme.Syntax.NumberBg
	case syntax.TokenOperator:
		fg, bg = u.theme.Syntax.OperatorFg, u.theme.Syntax.OperatorBg
	case syntax.TokenFunction:
		fg, bg = u.theme.Syntax.FunctionFg, u.theme.Syntax.FunctionBg
	case syntax.TokenType_:
		fg, bg = u.theme.Syntax.TypeFg, u.theme.Syntax.TypeBg
	case syntax.TokenVariable:
		fg, bg = u.theme.Syntax.VariableFg, u.theme.Syntax.VariableBg
	case syntax.TokenConstant:
		fg, bg = u.theme.Syntax.ConstantFg, u.theme.Syntax.ConstantBg
	case syntax.TokenPreproc:
		fg, bg = u.theme.Syntax.PreprocFg, u.theme.Syntax.PreprocBg
	case syntax.TokenBuiltin:
		fg, bg = u.theme.Syntax.BuiltinFg, u.theme.Syntax.BuiltinBg
	case syntax.TokenPunctuation:
		fg, bg = u.theme.Syntax.PunctuationFg, u.theme.Syntax.PunctuationBg
	default:
		return nil
	}

	// Use editor background if syntax background is empty
	if bg == "" || bg == term.ColorDefault {
		bg = u.theme.Editor.Bg
	}

	return &term.Style{Foreground: fg, Background: bg}
}

func (u *UI) getStatusStyle() term.Style {
	if u.isDefaultTheme() {
		return term.Style{Inverse: true}
	}
	return term.Style{Foreground: u.theme.Status.Fg, Background: u.theme.Status.Bg}
}

func (u *UI) getMenuStyle() term.Style {
	if u.isDefaultTheme() {
		return term.Style{Inverse: true}
	}
	return term.Style{Foreground: u.theme.Menu.Fg, Background: u.theme.Menu.Bg}
}

func (u *UI) getMenuSelectedStyle() term.Style {
	if u.isDefaultTheme() {
		// Use subtle dark grey background for selected menu items to distinguish from editor background
		// Almost the same as editor background but with just enough contrast to be visible
		return term.Style{Foreground: term.ColorDefault, Background: "#3A3A3A"}
	}
	return term.Style{Foreground: u.theme.Menu.SelectedFg, Background: u.theme.Menu.SelectedBg}
}

func (u *UI) getDropdownStyle() term.Style {
	if u.isDefaultTheme() {
		return term.Style{Inverse: true}
	}
	return term.Style{Foreground: u.theme.Menu.DropdownFg, Background: u.theme.Menu.DropdownBg}
}

func (u *UI) getDropdownSelectedStyle() term.Style {
	if u.isDefaultTheme() {
		// Use subtle dark grey background for selected dropdown items to distinguish from editor background
		// Almost the same as editor background but with just enough contrast to be visible
		return term.Style{Foreground: term.ColorDefault, Background: "#3A3A3A"}
	}
	return term.Style{Foreground: u.theme.Menu.DropdownSelFg, Background: u.theme.Menu.DropdownSelBg}
}

func (u *UI) getPromptStyle() term.Style {
	if u.isDefaultTheme() {
		return term.Style{Inverse: true}
	}
	return term.Style{Foreground: u.theme.Prompt.Fg, Background: u.theme.Prompt.Bg}
}

func (u *UI) getHelpStyle() term.Style {
	if u.isDefaultTheme() {
		return term.Style{Foreground: term.ColorDefault, Background: term.ColorDefault}
	}
	return term.Style{Foreground: u.theme.Help.Fg, Background: u.theme.Help.Bg}
}

func (u *UI) getHelpTitleStyle() term.Style {
	if u.isDefaultTheme() {
		return term.Style{Inverse: true}
	}
	return term.Style{Foreground: u.theme.Help.TitleFg, Background: u.theme.Help.TitleBg}
}

// getBracketStyle returns the style for bracket highlighting at the given position.
// Returns nil if the position is not a bracket that should be highlighted.
func (u *UI) getBracketStyle(docY, runeIdx int) *term.Style {
	if u.bracketMatchState == nil || !u.bracketMatchState.IsOnBracket {
		return nil
	}

	state := u.bracketMatchState

	// Check if this position is the cursor bracket or match bracket
	isCursorBracket := docY == state.CursorLine && runeIdx == state.CursorCol
	isMatchBracket := state.HasMatch && docY == state.MatchLine && runeIdx == state.MatchCol

	if !isCursorBracket && !isMatchBracket {
		return nil
	}

	// Choose color based on match status
	var bg term.Color
	if state.HasMatch {
		bg = u.theme.Editor.BracketMatchBg
	} else {
		bg = u.theme.Editor.BracketUnmatchBg
	}

	// Use editor foreground with bracket background
	fg := u.theme.Editor.Fg
	if u.isDefaultTheme() {
		fg = term.ColorDefault
	}

	return &term.Style{
		Foreground: fg,
		Background: bg,
	}
}

// getSearchMatchStyle returns the style for search match highlighting at the given position.
// Returns nil if the position is not in a search match.
// Distinguishes between current match and other matches.
func (u *UI) getSearchMatchStyle(docY, runeIdx int) *term.Style {
	// Only apply search highlighting in search mode
	if u.mode != ModeSearch {
		return nil
	}

	// Get active search session from editor
	searchState := u.editor.SearchState()
	if searchState == nil || searchState.Session == nil {
		return nil
	}

	session := searchState.Session
	if len(session.Matches) == 0 {
		return nil
	}

	// Check if this position is in any search match
	for i, match := range session.Matches {
		if match.Line != docY {
			continue
		}

		// Check if runeIdx is within this match [Col, Col+Length)
		if runeIdx >= match.Col && runeIdx < match.Col+match.Length {
			// Determine if this is the current match
			isCurrentMatch := (session.CurrentIndex == i)

			var fg, bg term.Color
			if isCurrentMatch {
				fg = u.theme.Search.CurrentMatchFg
				bg = u.theme.Search.CurrentMatchBg
			} else {
				fg = u.theme.Search.MatchFg
				bg = u.theme.Search.MatchBg
			}

			return &term.Style{
				Foreground: fg,
				Background: bg,
			}
		}
	}

	return nil
}

// drawRecoveryPrompt draws the autosave recovery prompt
func (u *UI) drawRecoveryPrompt(w, h int) {
	candidate := u.recoveryCandidate
	if candidate == nil {
		return
	}

	// Build the prompt lines
	lines := []string{
		"",
		"  Recover unsaved changes?",
		"",
	}

	// Add file path (truncate if too long)
	path := candidate.Meta.OriginalPath
	maxPathLen := w - 6
	if len(path) > maxPathLen && maxPathLen > 3 {
		path = "..." + path[len(path)-maxPathLen+3:]
	}
	lines = append(lines, "  An autosave backup was found for:")
	lines = append(lines, "  "+path)
	lines = append(lines, "")

	// Add timestamps
	autosaveTime := candidate.Meta.Timestamp.Format("2006-01-02 15:04:05")
	lines = append(lines, "  Autosave from: "+autosaveTime)

	if candidate.OriginalExists {
		originalTime := candidate.OriginalModTime.Format("2006-01-02 15:04:05")
		lines = append(lines, "  Original file: "+originalTime)

		if candidate.OriginalNewer {
			lines = append(lines, "")
			lines = append(lines, "  Warning: Original file is newer than autosave!")
		}
	} else {
		lines = append(lines, "  Original file: (does not exist)")
	}

	lines = append(lines, "")
	lines = append(lines, "")
	lines = append(lines, "  [R]ecover backup  [O]pen original  [D]iscard backup")
	lines = append(lines, "")

	style := u.getHelpStyle()
	titleStyle := u.getHelpTitleStyle()

	// Center vertically if space allows
	startY := 0
	if h > len(lines) {
		startY = (h - len(lines)) / 2
	}

	for i, line := range lines {
		y := startY + i
		if y >= h {
			break
		}

		s := style
		// Title line uses title style
		if i == 1 {
			s = titleStyle
		}

		// Draw line
		for x, r := range line {
			if x >= w {
				break
			}
			u.screen.SetCell(x, y, r, s)
		}

		// Fill rest of line with spaces
		for x := len(line); x < w; x++ {
			u.screen.SetCell(x, y, ' ', style)
		}
	}

	// Fill remaining lines with background
	for y := startY + len(lines); y < h; y++ {
		for x := 0; x < w; x++ {
			u.screen.SetCell(x, y, ' ', style)
		}
	}
	for y := 0; y < startY; y++ {
		for x := 0; x < w; x++ {
			u.screen.SetCell(x, y, ' ', style)
		}
	}
}
