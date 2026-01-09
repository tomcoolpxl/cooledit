package ui

import (
	"fmt"

	"cooledit/internal/term"
)

func (u *UI) draw() {
	u.screen.HideCursor()
	u.clear()

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
	for x, r := range msg {
		if x >= w {
			break
		}
		u.screen.SetCell(x, 0, r, term.Style{Inverse: true})
	}
}

func (u *UI) drawMenubar() {
	rect := u.layout.Menubar
	if rect.H < 1 {
		return
	}

	style := term.Style{Inverse: true}
	styleSelected := term.Style{Inverse: false}

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

	style := term.Style{Inverse: true}
	styleSelected := term.Style{Inverse: false}

	for i, item := range items {
		y := startY + i
		if y >= u.layout.Height {
			break
		}

		s := style
		if i == u.menubar.SelectedItemIndex {
			s = styleSelected
		}

		// Fill line
		for x := 0; x < width; x++ {
			u.screen.SetCell(startX+x, y, ' ', s)
		}

		// Draw Label
		for j, r := range item.Label {
			if j < width {
				u.screen.SetCell(startX+1+j, y, r, s)
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

	for sy := 0; sy < vpRect.H; sy++ {
		docY := vp.TopLine + sy

		// Draw Gutter
		if u.showLineNumbers {
			if docY < len(lines) {
				numStr := fmt.Sprintf("%d", docY+1) // 1-based
				// Right align
				padding := gutterWidth - len(numStr) - 1
				for i := 0; i < padding; i++ {
					u.screen.SetCell(vpRect.X+i, vpRect.Y+sy, ' ', term.Style{})
				}
				for i, r := range numStr {
					u.screen.SetCell(vpRect.X+padding+i, vpRect.Y+sy, r, term.Style{}) // Maybe diff style?
				}
				u.screen.SetCell(vpRect.X+gutterWidth-1, vpRect.Y+sy, ' ', term.Style{})
			} else {
				// Empty gutter
				for i := 0; i < gutterWidth; i++ {
					u.screen.SetCell(vpRect.X+i, vpRect.Y+sy, ' ', term.Style{})
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
		availW := vpRect.W - gutterWidth
		if availW < 0 {
			availW = 0
		}

		for sx := 0; sx < availW; sx++ {
			docX := start + sx
			// We check if docX is in selection range [sl:sc, el:ec)
			// But wait, GetSelectionRange returns normalized range? Yes.
			// Is it inclusive of end?
			// RangeText logic: [start, end) usually?
			// Let's re-read RangeText: "for l := sl; l <= el".
			// "if l == el { end = ec }".
			// So it includes up to ec-1.
			// e.g. "abc", select "a". Range 0,0 -> 0,1.

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

			style := term.Style{}
			if isSelected {
				style.Inverse = true
			}

			if docX >= len(line) {
				// Past end of line, maybe show selection if it spans newline?
				// If selection goes to next line, we should highlight the "newline char" (space)
				// RangeText logic: includes newline if l < el.
				if hasSelection && docY >= sl && docY < el {
					// We are on a selected line, but not the last one.
					// The "newline" at the end should be selected.
					// Draw a space with inverse style.
					u.screen.SetCell(drawX+sx, vpRect.Y+sy, ' ', term.Style{Inverse: true})
				}
				break
			}
			u.screen.SetCell(drawX+sx, vpRect.Y+sy, line[docX], style)
		}

		// Edge case: Empty line selection or selection extending past end of line logic above
		// The loop above breaks if docX >= len(line).
		// If line is empty, len(line) is 0. start is 0. Loop runs once for sx=0?
		// No, if vpRect.W > 0, sx=0. docX=0. if 0 >= 0 -> break.
		// So it breaks immediately. We need to handle the "newline" highlighting outside the loop or ensure loop runs.
		// Actually, simpler to check if we need to draw the newline highlight after the loop.
		// But we need the correct sx.

		lineLen := len(line)
		if lineLen < start {
			lineLen = start
		} // Should not happen if start clamped

		// If the line end is visible
		if lineLen >= start && lineLen < start+availW {
			sx := lineLen - start
			if hasSelection && docY >= sl && docY < el {
				u.screen.SetCell(drawX+sx, vpRect.Y+sy, ' ', term.Style{Inverse: true})
			}
		}
	}

	if u.mode == ModeNormal || u.mode == ModeMessage {
		cy, cx := u.editor.Cursor()
		sx := cx - vp.LeftCol
		sy := cy - vp.TopLine

		drawX := vpRect.X + gutterWidth
		availW := vpRect.W - gutterWidth

		if sx >= 0 && sx < availW && sy >= 0 && sy < vpRect.H {
			u.screen.ShowCursor(drawX+sx, vpRect.Y+sy)
		}
	}
}

func (u *UI) drawStatusBar() {
	rect := u.layout.StatusBar
	if rect.H < 1 {
		return
	}

	style := term.Style{Inverse: true}

	// Background
	for x := 0; x < rect.W; x++ {
		u.screen.SetCell(rect.X+x, rect.Y, ' ', style)
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
	right := fmt.Sprintf("Ln %d, Col %d  %s %s", cy+1, cx+1, fs.Encoding, eol)

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

	style := term.Style{Inverse: true} // Maybe different style for prompt?

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
	}

	footer := "  Press any key to close"

	style := term.Style{}
	titleStyle := term.Style{Inverse: true}

	// Determine if we can use two columns (need at least 80 width)
	useTwoColumns := w >= 80

	if useTwoColumns {
		// Two-column layout
		colWidth := 40
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
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			u.screen.SetCell(x, y, ' ', term.Style{})
		}
	}
}
