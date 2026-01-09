package ui

import (
	"fmt"

	"cooledit/internal/core"
	"cooledit/internal/term"
)

func (u *UI) draw() {
	u.screen.HideCursor()
	u.clear()

	// Set cursor shape and color based on insert/replace mode
	insertShape := ParseCursorShape(u.config.UI.CursorShape)
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
		label := fmt.Sprintf(" %s ", menu.Title)

		s := style
		if u.menubar.Active && i == u.menubar.SelectedMenuIndex {
			s = styleSelected
		}

		for _, r := range label {
			if x >= rect.W {
				break
			}
			u.screen.SetCell(rect.X+x, rect.Y, r, s)
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
	for sy := 0; sy < vpRect.H; sy++ {
		docY := vp.TopLine + sy

		// Draw Gutter
		gutterStyle := u.getLineNumberStyle()
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
		start := vp.LeftCol
		if start > len(line) {
			start = len(line)
		}

		drawX := vpRect.X + gutterWidth

		editorStyle := u.getEditorStyle()
		selectionStyle := u.getSelectionStyle()

		for sx := 0; sx < availW; sx++ {
			docX := start + sx

			isSelected := false
			if hasSelection {
				if docY > sl && docY < el {
					isSelected = true
				} else if docY == sl && docY == el {
					if docX >= sc && docX < ec {
						isSelected = true
					}
				} else if docY == sl {
					if docX >= sc {
						isSelected = true
					}
				} else if docY == el {
					if docX < ec {
						isSelected = true
					}
				}
			}

			style := editorStyle
			if isSelected {
				style = selectionStyle
			}

			if docX >= len(line) {
				// Past end of line - highlight newline if selected
				if hasSelection && docY >= sl && docY < el {
					u.screen.SetCell(drawX+sx, vpRect.Y+sy, ' ', selectionStyle)
				}
				break
			}
			u.screen.SetCell(drawX+sx, vpRect.Y+sy, line[docX], style)
		}

		// Highlight newline at end of line if selected
		lineLen := len(line)
		if lineLen < vp.LeftCol {
			lineLen = vp.LeftCol
		}
		if lineLen >= vp.LeftCol && lineLen < vp.LeftCol+availW {
			sx := lineLen - vp.LeftCol
			if hasSelection && docY >= sl && docY < el {
				u.screen.SetCell(drawX+sx, vpRect.Y+sy, ' ', selectionStyle)
			}
		}
	}

	// Draw cursor
	if u.mode == ModeNormal || u.mode == ModeMessage {
		cy, cx := u.editor.Cursor()
		sx := cx - vp.LeftCol
		sy := cy - vp.TopLine

		drawX := vpRect.X + gutterWidth

		if sx >= 0 && sx < availW && sy >= 0 && sy < vpRect.H {
			u.screen.ShowCursor(drawX+sx, vpRect.Y+sy)
		}
	}
}

// drawViewportWrapped renders the viewport with soft wrap enabled
func (u *UI) drawViewportWrapped(vpRect Rect, gutterWidth, availW int, lines [][]rune, vp core.Viewport, sl, sc, el, ec int, hasSelection bool) {
	if availW <= 0 {
		return
	}

	// Build wrapped lines structure
	type wrappedLine struct {
		lineNum int    // Original line number
		start   int    // Start column in original line
		content []rune // Wrapped segment
	}

	var wrapped []wrappedLine
	for lineNum, line := range lines {
		if len(line) == 0 {
			wrapped = append(wrapped, wrappedLine{lineNum: lineNum, start: 0, content: []rune{}})
			continue
		}

		// Wrap line into segments of availW width
		for start := 0; start < len(line); start += availW {
			end := start + availW
			if end > len(line) {
				end = len(line)
			}
			wrapped = append(wrapped, wrappedLine{
				lineNum: lineNum,
				start:   start,
				content: line[start:end],
			})
		}
	}

	// Draw wrapped lines starting from vp.TopLine
	gutterStyle := u.getLineNumberStyle()
	editorStyle := u.getEditorStyle()
	selectionStyle := u.getSelectionStyle()

	for sy := 0; sy < vpRect.H; sy++ {
		wrappedIdx := vp.TopLine + sy

		// Draw Gutter - show line number for first wrap of each line
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

		if wrappedIdx < 0 || wrappedIdx >= len(wrapped) {
			continue
		}

		wLine := wrapped[wrappedIdx]
		drawX := vpRect.X + gutterWidth

		// Draw the wrapped line segment
		for sx, r := range wLine.content {
			if sx >= availW {
				break
			}

			docY := wLine.lineNum
			docX := wLine.start + sx

			isSelected := false
			if hasSelection {
				if docY > sl && docY < el {
					isSelected = true
				} else if docY == sl && docY == el {
					if docX >= sc && docX < ec {
						isSelected = true
					}
				} else if docY == sl {
					if docX >= sc {
						isSelected = true
					}
				} else if docY == el {
					if docX < ec {
						isSelected = true
					}
				}
			}

			style := editorStyle
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
	}

	// Draw cursor
	if u.mode == ModeNormal || u.mode == ModeMessage {
		cy, cx := u.editor.Cursor()

		// Find which wrapped line contains this cursor position
		screenY := 0
		for wrappedIdx, wLine := range wrapped {
			if wLine.lineNum == cy {
				// Check if cursor is in this segment
				if cx >= wLine.start && cx < wLine.start+len(wLine.content) {
					sx := cx - wLine.start
					sy := wrappedIdx - vp.TopLine

					drawX := vpRect.X + gutterWidth

					if sy >= 0 && sy < vpRect.H && sx >= 0 && sx < availW {
						u.screen.ShowCursor(drawX+sx, vpRect.Y+sy)
						return
					}
				} else if cx == wLine.start+len(wLine.content) {
					// Cursor at end of this segment
					// Check if this is the last segment or if cursor should be on next segment
					isLastSegment := (wrappedIdx+1 >= len(wrapped) || wrapped[wrappedIdx+1].lineNum != cy)
					if isLastSegment || len(wLine.content) < availW {
						// Show cursor at end of this segment
						sx := len(wLine.content)
						sy := wrappedIdx - vp.TopLine

						drawX := vpRect.X + gutterWidth

						if sy >= 0 && sy < vpRect.H && sx >= 0 && sx < availW {
							u.screen.ShowCursor(drawX+sx, vpRect.Y+sy)
							return
						}
					}
				}
				screenY++
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

	var left string

	// Special status bar for find/replace mode
	if u.mode == ModeFindReplace {
		left = "[R]eplace  [N]ext  [P]rev  [A]ll  [Q]uit"
	} else {
		fs := u.editor.File()
		mod := ""
		if u.editor.Modified() {
			mod = "*"
		}
		left = fmt.Sprintf("%s%s", fs.BaseName, mod)
	}

	cy, cx := u.editor.Cursor()
	fs := u.editor.File()
	eol := "LF"
	if fs.EOL == "\r\n" {
		eol = "CRLF"
	}

	// Add replace mode indicator to right section
	modeIndicator := ""
	if !u.insertMode {
		modeIndicator = "REPLACE  "
	}
	right := fmt.Sprintf("%sLn %d, Col %d  %s %s", modeIndicator, cy+1, cx+1, fs.Encoding, eol)

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

	// Priority 3: Draw centered mini-help (if not in find/replace mode)
	if u.mode != ModeFindReplace {
		// Build mini-help with priority from left to right
		miniHelp := []string{
			"F1 Help",
			"Esc/F10 Menu",
			"Ctrl+Q Quit",
			"Ctrl+S Save",
			"Ctrl+F Find/Replace",
			"Ins Insert/Replace",
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
		"  CoolEdit - Quick Reference",
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
		"    Ctrl+Tab      Insert literal tab (\\t)",
	}

	rightCol := []string{
		"",
		"",
		"  SEARCH",
		"    Ctrl+F        Find/Replace",
		"    F3, Shift+F3  Next/Previous",
		"    Ctrl+G        Go to line",
		"",
		"  NAVIGATION",
		"    Arrows        Move cursor",
		"    Home/End      Line start/end",
		"    Ctrl+Home     File start",
		"    Ctrl+End      File end",
		"    PgUp/PgDn     Page up/down",
		"",
		"  VIEW",
		"    Ctrl+L        Line numbers",
		"    Ctrl+W        Word wrap",
		"    View menu     Cursor shapes",
		"",
		"  THEMES",
		"    View menu     13 built-in themes",
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
	if u.isDefaultTheme() {
		return term.Style{Foreground: term.ColorDefault, Background: term.ColorDefault}
	}
	return term.Style{Foreground: u.theme.Editor.LineNumbersFg, Background: u.theme.Editor.LineNumbersBg}
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
		return term.Style{Foreground: term.ColorDefault, Background: term.ColorDefault}
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
		return term.Style{Foreground: term.ColorDefault, Background: term.ColorDefault}
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
