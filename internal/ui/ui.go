package ui

import (
	"time"

	"cooledit/internal/core"
	"cooledit/internal/term"
)

type UIMode int

const (
	ModeNormal UIMode = iota
	ModeMessage
	ModePrompt
	ModeHelp
	ModeMenu
)

type UI struct {
	screen term.Screen
	editor *core.Editor
	menubar *Menubar

	mode   UIMode
	layout Layout
	showMenubar bool

	// message mode
	message      string
	messageUntil time.Time

	// prompt mode
	promptKind  PromptKind
	promptLabel string
	promptText  []rune

	// used by overwrite prompt
	pendingPath string

	// used by quit flow
	quitAfterSave bool
	quitNow       bool
}

func New(screen term.Screen, editor *core.Editor) *UI {
	return &UI{
		screen:      screen,
		editor:      editor,
		menubar:     NewMenubar(),
		mode:        ModeNormal,
		showMenubar: false,
	}
}

func (u *UI) Run() error {
	for {
		if u.quitNow {
			return nil
		}

		w, h := u.screen.Size()
		
		u.layout = ComputeLayout(w, h, u.mode, u.showMenubar)
		
		u.draw()

		ev := u.screen.PollEvent()
		if ev == nil {
			continue
		}

		switch e := ev.(type) {
		case term.KeyEvent:
			if u.mode == ModeHelp {
				u.mode = ModeNormal
				continue
			}

			if u.mode == ModePrompt {
				if u.handlePromptKey(e) {
					continue
				}
			}
			
			if u.mode == ModeMenu {
				if u.handleMenuKey(e) {
					continue
				}
			}

			if e.Key == term.KeyF10 {
				u.toggleMenuFocus()
				continue
			}

			if e.Key == term.KeyEscape {
				if u.mode == ModeMessage {
					u.mode = ModeNormal
					continue
				}
				if u.mode == ModeNormal {
					u.toggleMenuFocus()
					continue
				}
				continue
			}

			cmd := u.translateKey(e)
			if cmd != nil {
				res := u.editor.Apply(cmd, u.layout.Viewport.H)
				if res.Message != "" {
					u.enterMessage(res.Message)
				}
			}
			
		case term.MouseEvent:
			u.handleMouseEvent(e)
		}
	}
}

func (u *UI) toggleMenuFocus() {
	if !u.showMenubar {
		u.showMenubar = true
		u.mode = ModeMenu
		u.menubar.Active = true
		return
	}
	
	// If visible, hide it
	u.showMenubar = false
	u.mode = ModeNormal
	u.menubar.Active = false
}

func (u *UI) handleMenuKey(e term.KeyEvent) bool {
	switch e.Key {
	case term.KeyEscape:
		u.mode = ModeNormal
		u.menubar.Active = false
		u.showMenubar = false
		return true
	case term.KeyLeft:
		u.menubar.PrevMenu()
		return true
	case term.KeyRight:
		u.menubar.NextMenu()
		return true
	case term.KeyUp:
		u.menubar.PrevItem()
		return true
	case term.KeyDown:
		u.menubar.NextItem()
		return true
	case term.KeyEnter:
		u.executeMenuItem()
		return true
	}
	return false
}

func (u *UI) executeMenuItem() {
	menu := u.menubar.Menus[u.menubar.SelectedMenuIndex]
	item := menu.Items[u.menubar.SelectedItemIndex]
	
	// Exit menu mode
	u.mode = ModeNormal
	u.menubar.Active = false
	u.showMenubar = false
	
	if item.Action != nil {
		item.Action(u)
	} else if item.Command != nil {
		res := u.editor.Apply(item.Command, u.layout.Viewport.H)
		if res.Message != "" {
			u.enterMessage(res.Message)
		}
	}
}

func (u *UI) handleMouseEvent(e term.MouseEvent) {
	// 1. Check Menubar
	if u.showMenubar && e.Y == u.layout.Menubar.Y {
		if e.Button == term.MouseLeft {
			// Find which menu was clicked
			x := 0
			for i, menu := range u.menubar.Menus {
				width := len(menu.Title) + 2 // " Title "
				if e.X >= x && e.X < x+width {
					u.mode = ModeMenu
					u.menubar.Active = true
					u.menubar.SelectedMenuIndex = i
					u.menubar.SelectedItemIndex = 0
					return
				}
				x += width
			}
		}
		return
	}
	
	// 2. Check Menu Dropdown (if active)
	if u.mode == ModeMenu {
		menuIdx := u.menubar.SelectedMenuIndex
		menuX := 0
		for i := 0; i < menuIdx; i++ {
			menuX += len(u.menubar.Menus[i].Title) + 2
		}
		
		menu := u.menubar.Menus[menuIdx]
		width := 0
		for _, item := range menu.Items {
			w := len(item.Label) + 4 + len(item.Accelerator)
			if w > width { width = w }
		}
		if width < 10 { width = 10 }
		
		startX := menuX
		startY := 1
		if startX+width > u.layout.Width { startX = u.layout.Width - width }
		
		// Check bounds
		numItems := len(menu.Items)
		if e.X >= startX && e.X < startX+width && e.Y >= startY && e.Y < startY+numItems {
			if e.Button == term.MouseLeft {
				idx := e.Y - startY
				u.menubar.SelectedItemIndex = idx
				u.executeMenuItem()
				return
			}
		} else {
			// Click outside menu -> close
			if e.Button == term.MouseLeft {
				u.mode = ModeNormal
				u.menubar.Active = false
				u.showMenubar = false
			}
		}
		return
	}

	// 3. Check Viewport
	vp := u.layout.Viewport
	if e.X >= vp.X && e.X < vp.X+vp.W && e.Y >= vp.Y && e.Y < vp.Y+vp.H {
		if e.Button == term.MouseLeft {
			viewX := e.X - vp.X
			viewY := e.Y - vp.Y
			
			docLine := u.editor.Viewport().TopLine + viewY
			docCol := u.editor.Viewport().LeftCol + viewX
			
			u.editor.Apply(core.CmdClick{Line: docLine, Col: docCol}, u.layout.Viewport.H)
			
		} else if e.Button == term.MouseWheelUp {
			u.editor.Apply(core.CmdMoveUp{}, u.layout.Viewport.H)
			u.editor.Apply(core.CmdMoveUp{}, u.layout.Viewport.H)
			u.editor.Apply(core.CmdMoveUp{}, u.layout.Viewport.H)
		} else if e.Button == term.MouseWheelDown {
			u.editor.Apply(core.CmdMoveDown{}, u.layout.Viewport.H)
			u.editor.Apply(core.CmdMoveDown{}, u.layout.Viewport.H)
			u.editor.Apply(core.CmdMoveDown{}, u.layout.Viewport.H)
		}
	}
}

func (u *UI) translateKey(e term.KeyEvent) core.Command {
	switch {
	case e.Key == term.KeyF1:
		u.mode = ModeHelp
		return nil

	case e.Key == term.KeyRune && e.Rune == 'q' && (e.Modifiers&term.ModCtrl) != 0:
		u.startQuitFlow()
		return nil

	case e.Key == term.KeyRune && e.Rune == 's' && e.Modifiers == term.ModCtrl:
		if u.editor.File().Path == "" {
			u.enterSaveAs(false)
			return nil
		}
		return core.CmdSave{}

	case e.Key == term.KeyRune && e.Rune == 's' &&
		(e.Modifiers&(term.ModCtrl|term.ModShift)) == (term.ModCtrl|term.ModShift):
		u.enterSaveAs(false)
		return nil

	case e.Key == term.KeyRune && e.Rune == 'c' && e.Modifiers == term.ModCtrl:
		return core.CmdCopy{}

	case e.Key == term.KeyRune && e.Rune == 'x' && e.Modifiers == term.ModCtrl:
		return core.CmdCut{}

	case (e.Key == term.KeyDelete && e.Modifiers == term.ModShift):
		return core.CmdCut{}

	case e.Key == term.KeyRune && e.Rune == 'v' && e.Modifiers == term.ModCtrl:
		return core.CmdPaste{}

	case (e.Key == term.KeyInsert && e.Modifiers == term.ModShift):
		return core.CmdPaste{}

	case e.Key == term.KeyRune && e.Rune == 'z' && (e.Modifiers&term.ModCtrl) != 0:
		return core.CmdUndo{}

	case e.Key == term.KeyRune && e.Rune == 'y' && (e.Modifiers&term.ModCtrl) != 0:
		return core.CmdRedo{}
	
	case e.Key == term.KeyRune && e.Rune == 'z' && (e.Modifiers&(term.ModCtrl|term.ModShift)) == (term.ModCtrl|term.ModShift):
		return core.CmdRedo{}
		
	case e.Key == term.KeyRune && e.Rune == 'f' && (e.Modifiers&term.ModCtrl) != 0:
		u.enterFind()
		return nil

	case e.Key == term.KeyF3:
		if e.Modifiers == term.ModShift {
			return core.CmdFindPrev{}
		}
		return core.CmdFindNext{}

	case e.Key == term.KeyRune && e.Modifiers == 0:
		return core.CmdInsertRune{Rune: e.Rune}

	case e.Key == term.KeyEnter:
		return core.CmdInsertNewline{}

	case e.Key == term.KeyBackspace:
		return core.CmdBackspace{}

	case e.Key == term.KeyLeft:
		return core.CmdMoveLeft{Select: e.Modifiers&term.ModShift != 0}
	case e.Key == term.KeyRight:
		return core.CmdMoveRight{Select: e.Modifiers&term.ModShift != 0}
	case e.Key == term.KeyUp:
		return core.CmdMoveUp{Select: e.Modifiers&term.ModShift != 0}
	case e.Key == term.KeyDown:
		return core.CmdMoveDown{Select: e.Modifiers&term.ModShift != 0}

	case e.Key == term.KeyPageUp:
		return core.CmdPageUp{Select: e.Modifiers&term.ModShift != 0}
	case e.Key == term.KeyPageDown:
		return core.CmdPageDown{Select: e.Modifiers&term.ModShift != 0}

	case e.Key == term.KeyHome && (e.Modifiers&term.ModCtrl) != 0:
		return core.CmdFileStart{Select: e.Modifiers&term.ModShift != 0}
	case e.Key == term.KeyEnd && (e.Modifiers&term.ModCtrl) != 0:
		return core.CmdFileEnd{Select: e.Modifiers&term.ModShift != 0}

	case e.Key == term.KeyHome:
		return core.CmdMoveHome{Select: e.Modifiers&term.ModShift != 0}
	case e.Key == term.KeyEnd:
		return core.CmdMoveEnd{Select: e.Modifiers&term.ModShift != 0}
	}

	return nil
}

func (u *UI) enterMessage(msg string) {
	u.mode = ModeMessage
	u.message = msg
	u.messageUntil = time.Now().Add(2 * time.Second)
}