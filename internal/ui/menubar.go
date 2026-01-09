package ui

import "cooledit/internal/core"

type MenuItem struct {
	Label       string
	Accelerator string // e.g. "Ctrl+S"
	ShortcutKey rune   // Single letter shortcut when menu is open (e.g. 's' for Save)
	Command     core.Command
	Action      func(*UI)        // Special actions like "Quit" or "Toggle"
	Submenu     []MenuItem       // Submenu items (e.g., for Themes)
	IsCheckable bool             // If true, item can be checked
	IsChecked   func(*UI) bool   // Function to determine if checked
	IsSeparator bool             // If true, renders as separator line
	IsReadOnly  bool             // If true, item is not clickable (display only)
	GetValue    func(*UI) string // For readonly items, returns current value
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
	ScrollOffset      int  // Vertical scroll offset for current menu
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
	themeItems := m.buildThemeItems()
	cursorItems := m.buildCursorShapeItems()

	m.Menus = []Menu{
		{
			Title: "File",
			Items: []MenuItem{
				{Label: "Save", Accelerator: "Ctrl+S", ShortcutKey: 's', Command: core.CmdSave{}},
				{Label: "Save As", Accelerator: "Ctrl+Shift+S", ShortcutKey: 'a', Action: func(u *UI) { u.enterSaveAs(false) }},
				{Label: "Quit", Accelerator: "Ctrl+Q", ShortcutKey: 'q', Action: func(u *UI) { u.startQuitFlow() }},
			},
		},
		{
			Title: "Edit",
			Items: []MenuItem{
				{Label: "Undo", Accelerator: "Ctrl+Z", ShortcutKey: 'u', Command: core.CmdUndo{}},
				{Label: "Redo", Accelerator: "Ctrl+Y", ShortcutKey: 'r', Command: core.CmdRedo{}},
				{Label: "Cut", Accelerator: "Ctrl+X", ShortcutKey: 't', Command: core.CmdCut{}},
				{Label: "Copy", Accelerator: "Ctrl+C", ShortcutKey: 'c', Command: core.CmdCopy{}},
				{Label: "Paste", Accelerator: "Ctrl+V", ShortcutKey: 'p', Command: core.CmdPaste{}},
				{Label: "Go to Line", Accelerator: "Ctrl+G", ShortcutKey: 'g', Action: func(u *UI) { u.enterGoToLine() }},
			},
		},
		{
			Title: "Search",
			Items: []MenuItem{
				{Label: "Find / Replace", Accelerator: "Ctrl+F", ShortcutKey: 'f', Action: func(u *UI) { u.enterFind() }},
				{Label: "Find Next", Accelerator: "F3", ShortcutKey: 'n', Command: core.CmdFindNext{}},
				{Label: "Find Previous", Accelerator: "Shift+F3", ShortcutKey: 'p', Command: core.CmdFindPrev{}},
			},
		},
		{
			Title: "View",
			Items: append([]MenuItem{
				{Label: "Toggle Line Numbers", Accelerator: "Ctrl+L", IsCheckable: true, IsChecked: func(u *UI) bool {
					return u.showLineNumbers
				}, Action: func(u *UI) {
					u.showLineNumbers = !u.showLineNumbers
					u.saveConfig()
				}},
				{Label: "Toggle Word Wrap", Accelerator: "Ctrl+W", IsCheckable: true, IsChecked: func(u *UI) bool {
					return u.softWrap
				}, Action: func(u *UI) {
					u.softWrap = !u.softWrap
					u.saveConfig()
				}},
				{Label: "Toggle Status Bar", Accelerator: "F11", IsCheckable: true, IsChecked: func(u *UI) bool {
					return u.showStatusBar
				}, Action: func(u *UI) {
					u.showStatusBar = !u.showStatusBar
					u.saveConfig()
				}},
				{IsSeparator: true},
				{Label: "EOL Format", IsReadOnly: true, GetValue: func(u *UI) string {
					return u.editor.File().EOL
				}},
				{Label: "Encoding", IsReadOnly: true, GetValue: func(u *UI) string {
					return u.editor.File().Encoding
				}},
				{IsSeparator: true},
			}, append(cursorItems, append([]MenuItem{{IsSeparator: true}}, themeItems...)...)...),
		},
		{
			Title: "Help",
			Items: []MenuItem{
				{Label: "Keyboard Shortcuts", Accelerator: "F1", ShortcutKey: 'k', Action: func(u *UI) { u.mode = ModeHelp }},
			},
		},
	}
}

// buildThemeItems creates menu items for all available themes
func (m *Menubar) buildThemeItems() []MenuItem {
	themes := []string{
		"default",
		"dark",
		"light",
		"monokai",
		"solarized-dark",
		"solarized-light",
		"gruvbox-dark",
		"gruvbox-light",
		"dracula",
		"nord",
		"dos",
		"ibm-green",
		"ibm-amber",
	}

	items := make([]MenuItem, len(themes))
	for i, themeName := range themes {
		// Capture themeName in closure
		name := themeName
		items[i] = MenuItem{
			Label:       "Theme: " + name,
			Accelerator: "",
			IsCheckable: true,
			IsChecked: func(u *UI) bool {
				return u.GetCurrentThemeName() == name
			},
			Action: func(u *UI) {
				u.SwitchTheme(name)
			},
		}
	}
	return items
}

// buildCursorShapeItems creates menu items for cursor shapes
func (m *Menubar) buildCursorShapeItems() []MenuItem {
	cursorShapes := []string{"block", "underline", "bar"}
	items := make([]MenuItem, len(cursorShapes))

	for i, shapeName := range cursorShapes {
		// Capture shapeName in closure
		name := shapeName
		items[i] = MenuItem{
			Label:       "Cursor: " + name,
			Accelerator: "",
			IsCheckable: true,
			IsChecked: func(u *UI) bool {
				return u.config.UI.CursorShape == name
			},
			Action: func(u *UI) {
				u.config.UI.CursorShape = name
				u.saveConfig()
			},
		}
	}
	return items
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
	startIndex := m.SelectedItemIndex

	// Move to next item
	m.SelectedItemIndex = (m.SelectedItemIndex + 1) % len(items)

	// Skip separators and readonly items
	for items[m.SelectedItemIndex].IsSeparator || items[m.SelectedItemIndex].IsReadOnly {
		m.SelectedItemIndex = (m.SelectedItemIndex + 1) % len(items)
		// Prevent infinite loop if all items are separators/readonly
		if m.SelectedItemIndex == startIndex {
			break
		}
	}
}

func (m *Menubar) PrevItem() {
	items := m.Menus[m.SelectedMenuIndex].Items
	startIndex := m.SelectedItemIndex

	// Move to previous item
	m.SelectedItemIndex--
	if m.SelectedItemIndex < 0 {
		m.SelectedItemIndex = len(items) - 1
	}

	// Skip separators and readonly items
	for items[m.SelectedItemIndex].IsSeparator || items[m.SelectedItemIndex].IsReadOnly {
		m.SelectedItemIndex--
		if m.SelectedItemIndex < 0 {
			m.SelectedItemIndex = len(items) - 1
		}
		// Prevent infinite loop if all items are separators/readonly
		if m.SelectedItemIndex == startIndex {
			break
		}
	}
}
