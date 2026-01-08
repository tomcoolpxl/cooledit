package ui

import (
	"fmt"
	"time"

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

	for sy := 0; sy < vpRect.H; sy++ {
		docY := vp.TopLine + sy
		if docY < 0 || docY >= len(lines) {
			continue
		}

		line := lines[docY]
		start := vp.LeftCol
		if start > len(line) {
			start = len(line)
		}

		for sx := 0; sx < vpRect.W; sx++ {
			docX := start + sx
			if docX >= len(line) {
				break
			}
			u.screen.SetCell(vpRect.X+sx, vpRect.Y+sy, line[docX], term.Style{})
		}
	}

	if u.mode == ModeNormal || u.mode == ModeMessage {
		cy, cx := u.editor.Cursor()
		sx := cx - vp.LeftCol
		sy := cy - vp.TopLine
		if sx >= 0 && sx < vpRect.W && sy >= 0 && sy < vpRect.H {
			u.screen.ShowCursor(vpRect.X+sx, vpRect.Y+sy)
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

	fs := u.editor.File()
	mod := ""
	if u.editor.Modified() {
		mod = "*"
	}

	left := fmt.Sprintf("%s%s  Ctrl+S Save  Ctrl+Q Quit  F1 Help  F10 Menu", fs.BaseName, mod)

	cy, cx := u.editor.Cursor()
	eol := "LF"
	if fs.EOL == "\r\n" {
		eol = "CRLF"
	}
	right := fmt.Sprintf("Ln %d, Col %d  %s %s", cy+1, cx+1, fs.Encoding, eol)

	// Draw Right
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

	// Draw Left
	maxLeft := startRight - 1
	if maxLeft < 0 {
		maxLeft = 0
	}
	for i, r := range left {
		if i >= maxLeft {
			break
		}
		u.screen.SetCell(rect.X+i, rect.Y, r, style)
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
		if time.Now().After(u.messageUntil) {
			u.mode = ModeNormal
			// Next frame will correct layout
			return
		}
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
	lines := []string{
		"cooledit - help",
		"",
		"Ctrl+S        Save",
		"Ctrl+Shift+S  Save As",
		"Ctrl+Q        Quit",
		"Ctrl+C        Force quit",
		"Ctrl+Z        Undo",
		"Ctrl+Y        Redo",
		"Arrows        Move cursor",
		"PgUp/PgDn     Scroll",
		"Ctrl+Home/End File start/end",
		"F10           Menu",
		"F1            Help",
		"",
		"Press any key to return",
	}

	for y := 0; y < len(lines) && y < h; y++ {
		for x, r := range lines[y] {
			if x >= w {
				break
			}
			u.screen.SetCell(x, y, r, term.Style{})
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