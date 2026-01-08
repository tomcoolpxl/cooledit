package ui

import "cooledit/internal/core"

type MenuItem struct {
	Label       string
	Accelerator string // e.g. "Ctrl+S"
	Command     core.Command
	Action      func(*UI) // Special actions like "Quit" or "Toggle"
}

type Menu struct {
	Title string
	Items []MenuItem
}

type Menubar struct {
	Menus []Menu

	// State
	Active            bool // If true, we are navigating the menu
	SelectedMenuIndex int
	SelectedItemIndex int
}

func NewMenubar() *Menubar {
	m := &Menubar{
		Menus:             make([]Menu, 0),
		Active:            false,
		SelectedMenuIndex: 0,
		SelectedItemIndex: 0,
	}
	m.initDefaults()
	return m
}

func (m *Menubar) initDefaults() {
	m.Menus = []Menu{
		{
			Title: "File",
			Items: []MenuItem{
				{Label: "Save", Accelerator: "Ctrl+S", Command: core.CmdSave{}},
				{Label: "Save As", Accelerator: "Ctrl+Shift+S", Action: func(u *UI) { u.enterSaveAs(false) }},
				{Label: "Quit", Accelerator: "Ctrl+Q", Action: func(u *UI) { u.startQuitFlow() }},
			},
		},
		{
			Title: "Edit",
			Items: []MenuItem{
				{Label: "Undo", Accelerator: "Ctrl+Z", Command: core.CmdUndo{}},
				{Label: "Redo", Accelerator: "Ctrl+Y", Command: core.CmdRedo{}},
				// Cut/Copy/Paste not implemented yet
			},
		},
		{
			Title: "Search",
			Items: []MenuItem{
				{Label: "Find", Accelerator: "Ctrl+F", Action: func(u *UI) { u.enterFind() }},
				{Label: "Find Next", Accelerator: "F3", Command: core.CmdFindNext{}},
				{Label: "Find Previous", Accelerator: "Shift+F3", Command: core.CmdFindPrev{}},
			},
		},
		{
			Title: "View",
			Items: []MenuItem{
				// TODO: Toggle Line Numbers, Toggle Wrap
				{Label: "Toggle Menubar", Action: func(u *UI) { /* Toggle logic */ }},
			},
		},
		{
			Title: "Help",
			Items: []MenuItem{
				{Label: "About", Accelerator: "F1", Action: func(u *UI) { u.mode = ModeHelp }},
			},
		},
	}
}

// Navigation methods

func (m *Menubar) NextMenu() {
	m.SelectedMenuIndex = (m.SelectedMenuIndex + 1) % len(m.Menus)
	m.SelectedItemIndex = 0 // Reset item selection when switching menus
}

func (m *Menubar) PrevMenu() {
	m.SelectedMenuIndex--
	if m.SelectedMenuIndex < 0 {
		m.SelectedMenuIndex = len(m.Menus) - 1
	}
	m.SelectedItemIndex = 0
}

func (m *Menubar) NextItem() {
	items := m.Menus[m.SelectedMenuIndex].Items
	m.SelectedItemIndex = (m.SelectedItemIndex + 1) % len(items)
}

func (m *Menubar) PrevItem() {
	items := m.Menus[m.SelectedMenuIndex].Items
	m.SelectedItemIndex--
	if m.SelectedItemIndex < 0 {
		m.SelectedItemIndex = len(items) - 1
	}
}