# Claude Context - Cooledit Project

## Project Overview

**Cooledit** is a terminal-based text editor written in vanilla Go, inspired by nano but with better UI and keyboard shortcuts. The goal is to create a simple, beginner-friendly editor without syntax highlighting.

## Core Design Principles

- **No Syntax Highlighting** - Intentional design decision
- **No Modal Editing** - Direct text editing, no vim-like modes
- **Familiar Shortcuts** - Modern keyboard shortcuts (Ctrl+S, Ctrl+C, etc.)
- **Clean UI** - Status bar always visible, menubar auto-hidden by default
- **Cross-platform** - Works on Windows, Linux, macOS

## Technology Stack

- **Language**: Go (vanilla, no frameworks)
- **Terminal Library**: tcell
- **Structure**: Clean architecture with internal packages

## Project Structure

```
cmd/cooledit/          - Entry point
internal/
  app/                 - Application lifecycle
  config/              - Configuration management
  core/                - Editor core logic
    buffer/            - Text buffer implementation
    text/              - Text processing utilities
  fileio/              - File operations, encoding, EOL handling
  term/                - Terminal backend abstraction
    tcell/             - Tcell implementation
  ui/                  - User interface components
    keymap/            - Keyboard bindings
```

## Key Features

### Implemented (Milestones 1-3)
- ✅ Basic text editing with buffer management
- ✅ File load/save with encoding detection (UTF-8, ISO-8859-1, etc.)
- ✅ EOL format detection and preservation (LF/CRLF)
- ✅ Search (Find/Next/Previous) with non-overlapping matches
- ✅ Replace (unified find/replace mode with R/N/P/A/Q options)
- ✅ Undo/Redo with command pattern
- ✅ Text selection via Shift+Arrow keys
- ✅ System clipboard integration (Cut/Copy/Paste)
- ✅ Status bar with priority-based layout and centered mini-help
- ✅ Auto-hiding menubar (toggle with F10 or Esc)
- ✅ Mouse support (optional via CLI flag or config)
- ✅ Go to Line (Ctrl+G) - Always available
- ✅ Adaptive help screen (F1) with two-column layout for wide terminals
- ✅ Configuration persistence with TOML file
- ✅ Toggle settings auto-save (line numbers, soft wrap)
- ✅ Soft wrap implementation with adaptive line wrapping
- ✅ Insert/Replace mode toggle with Insert key and cursor shape indicators

### Planned (Milestone 4)
- ⏳ Keybinding customization

## Important Keyboard Shortcuts

- `Ctrl+S` - Save
- `Ctrl+Shift+S` - Save As
- `Ctrl+Q` - Quit
- `Ctrl+X` - Cut
- `Ctrl+C` - Copy
- `Ctrl+V` - Paste
- `Ctrl+Z` - Undo
- `Ctrl+Y` - Redo
- `Ctrl+F` - Find/Replace (unified mode)
- `F3` / `Shift+F3` - Find Next/Previous
- `Ctrl+G` - Go to Line (always available)
- `Insert` - Toggle Insert/Replace mode
- `F1` - Help overlay (adaptive two-column/single-column layout)
- `F10` / `Esc` - Toggle menubar

## Key Design Decisions

### Status Bar (Always Visible, Priority-Based Layout)
**Left** (priority 2): Filename with modified indicator (`*`)
**Center** (priority 3): Mini-help with adaptive display:
  - `F1 Help` → `Esc/F10 Menu` → `Ctrl+Q Quit` → `Ctrl+S Save` → `Ctrl+F Find/Replace`
  - Shows as many items as fit, removes from right to left when space is limited
**Right** (priority 1): `Ln X, Col Y  Encoding EOL` (cursor position and file status)

### Menubar (Auto-hidden by Default)
- Toggle with F10 or Esc
- Menus: File, Edit, Search, View, Help
- Navigation via arrow keys or mouse (if enabled)

### Command-Line Flags
- `-mouse` - Enable mouse support (click to position cursor, scroll)
- `-line-numbers` - Show line numbers column

### Clipboard Behavior
- Cut/Copy with no selection operates on current line
- Integrates with system clipboard (not just internal buffer)

### Save Behavior
- `Ctrl+S` on unnamed buffer → triggers Save As
- Save As only prompts for overwrite if file exists and is different from current
- Normal save never prompts (VS Code behavior)

### Find/Replace Workflow (Unified Nano-Style Mode)
1. `Ctrl+F` opens "Find: " prompt (pre-filled with last search term)
2. Enter search term → `Enter`
3. Editor enters **Find/Replace Mode** with first match highlighted
4. Status bar shows: **"[R]eplace  [N]ext  [P]rev  [A]ll  [Q]uit"**
5. User can:
   - `N` - Find next match (non-overlapping: "ttt" in "ttttt" finds once)
   - `P` - Find previous match
   - `R` - Replace current match (prompts for replacement text, then finds next)
   - `A` - Replace all matches from beginning of file (prompts for replacement text)
   - `Q` or `Esc` - Quit find/replace mode
6. Message bar persists during find/replace operations

**Key Features:**
- Find and Replace share the same search term
- Previous search/replace values are remembered and pre-populated
- Found text is highlighted with selection during search
- Non-overlapping search (no repeated matches on same text)
- Replace All always starts from beginning of file regardless of cursor position
- Each replace operation is undoable
- Unified mode keeps user in find/replace context instead of separate dialogs

### Help Screen (F1)
- **Adaptive layout**: Two columns on wide terminals (≥80 chars), single column on narrow
- **Smart truncation**: Shows "(scroll down for more)" when content doesn't fit
- **Organized sections**: Menu & Help, File, Edit, Search, Navigation
- **Clean design**: Simple indentation, highlighted section headers, no box-drawing characters

### Word Wrap
- **Off by default**
- Can be toggled with Ctrl+W
- Adaptive wrapping to viewport width
- Line numbers shown only on first wrapped segment

### Insert/Replace Mode
- **Insert mode by default** (block cursor)
- Toggle with Insert key to replace/overwrite mode (underline cursor)
- Replace mode overwrites characters instead of inserting
- At end of line, behaves like insert mode
- State not saved - always starts in insert mode
- Cursor shapes customizable in future

## Testing Strategy

- Unit tests for core components (buffer, editor, search, undo)
- UI tests with fake screen implementation
- Coverage tracking in place
- **102+ tests covering**:
  - Non-overlapping search matches (TestFindNextNoOverlapping, TestFindNextTwoNonOverlapping)
  - Replace All starting from file beginning (TestReplaceAllFromBeginning)
  - Text highlighting during search (TestSearchHighlightsText)
  - Replace operations (TestReplaceOne, TestReplaceAll, TestReplaceNotFound, TestReplaceUndoable)
  - Search UI integration (TestSearchUIIntegration)
  - Status bar mini-help (TestStatusBarMiniHelp, TestStatusBarMiniHelpNarrowTerminal)
  - Status bar in find/replace mode (TestStatusBarFindReplaceMode)
  - Help screen adaptive layout (TestHelpScreenWideTerminal, TestHelpScreenNarrowTerminal)
  - Help mode (TestHelpMode)
  - Status bar rendering (TestStatusBarCursorPosition)
  - Configuration save/load (TestDefaultConfig, TestSaveAndLoad, TestPartialConfig)
  - Toggle actions save config (TestToggleLineNumbersSavesConfig, TestToggleSoftWrapSavesConfig)
  - Soft wrap rendering (TestSoftWrapRendering, TestSoftWrapVsNoWrap)
  - Insert/Replace mode (TestInsertMode, TestReplaceMode, TestReplaceModeAtEndOfLine, TestInsertKeyToggle)

### EOL Format
- Auto-detect (LF vs CRLF)
- Display in status bar
- Preserve original format on save

## Configuration System

**Location:**
- Linux/macOS: `~/.config/cooledit/config.toml`
- Windows: `%APPDATA%\cooledit\config.toml`

**Settings:**
```toml
[editor]
line_numbers = false  # Show line numbers column
soft_wrap = false     # Enable word wrap
tab_width = 4         # Spaces per tab

[ui]
show_menubar = false  # Show menubar by default
mouse_enabled = false # Enable mouse support

[search]
case_sensitive = true # Case-sensitive search by default
```

**Behavior:**
- Config file created automatically on first toggle action
- CLI flags override config values
- Toggle actions (Ctrl+L, Ctrl+W) automatically save config
- Missing config file or fields use sensible defaults

## Non-Goals

- ❌ Syntax highlighting
- ❌ Tabbed interface
- ❌ Multiple simultaneous file buffers
- ❌ Markdown rendering
- ❌ Plugin system (not initial scope)

## Development Context

- Written with modern Go practices
- Clean architecture with clear separation of concerns
- Terminal abstraction allows for different backend implementations
- Test coverage is important - maintain tests for core functionality

## Current Status

Project is fully functional with all core features complete:
- ✅ Go to Line is always available
- ✅ Unified Find/Replace mode with nano-style workflow
- ✅ Non-overlapping search (proper match advancement)
- ✅ Replace All starts from beginning of file
- ✅ Priority-based status bar with adaptive centered mini-help
- ✅ Adaptive help screen with two-column layout for wide terminals
- ✅ Message bar persistence during find/replace operations
- ✅ Configuration system with TOML persistence
- ✅ Toggle settings auto-save to config file
- ✅ Soft wrap rendering with proper line wrapping and cursor positioning
- ✅ Insert/Replace mode with Insert key toggle and cursor shape indicators
- ✅ Comprehensive test coverage (120+ tests, all passing)

Focus areas:
- Keybinding customization

## When Working on This Project

1. **No syntax highlighting** - If asked about it, refer to design decision
2. **Follow existing patterns** - Buffer management, command pattern for undo/redo
3. **Test thoroughly** - Especially buffer operations and UI interactions
4. **Maintain simplicity** - This is meant to be a simple, nano-like editor
5. **Cross-platform** - Consider Windows, Linux, and macOS compatibility
6. **Terminal constraints** - Remember this runs in a terminal, not a GUI

---

Last Updated: January 9, 2026
