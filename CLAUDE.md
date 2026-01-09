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
- ✅ Search (Find/Next/Previous)
- ✅ Undo/Redo with command pattern
- ✅ Text selection via Shift+Arrow keys
- ✅ System clipboard integration (Cut/Copy/Paste)
- ✅ Status bar (filename, modified flag, cursor position, encoding, EOL)
- ✅ Auto-hiding menubar (toggle with F10)
- ✅ Mouse support (optional via `-mouse` flag)

### Planned (Milestone 4)
- ⏳ Configuration persistence
- ⏳ Keybinding customization
- ✅ Go to Line (Ctrl+G) - **Implemented and always available**
- 🔧 Soft wrap - **Partially implemented** (toggle exists, rendering may be incomplete)

## Important Keyboard Shortcuts

- `Ctrl+S` - Save
- `Ctrl+Shift+S` - Save As
- `Ctrl+Q` - Quit
- `Ctrl+X` - Cut
- `Ctrl+C` - Copy
- `Ctrl+V` - Paste
- `Ctrl+Z` - Undo
- `Ctrl+Y` - Redo
- `Ctrl+F` - Find
- `F3` / `Shift+F3` - Find Next/Previous
- `F1` - Help overlay
- `F10` - Toggle menubar

## Key Design Decisions

### Status Bar (Always Visible)
Displays: filename, modified indicator, line:col position, encoding, EOL type

### Menubar (Auto-hidden by Default)
- Toggle with F10
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

### Word Wrap
- **Off by default**
- Can be toggled (to be implemented)

## Testing Strategy

- Unit tests for core components (buffer, editor, search, undo)
- UI tests with fake screen implementation
- Coverage tracking in place

## File Handling

### Encoding Detection
- Auto-detect on file open
- Display in status bar
- Preserve on save

### EOL Format
- Auto-detect (LF vs CRLF)
- Display in status bar
- Preserve original format on save

## Configuration (Future)

Will support:
- Custom keybindings
- Line numbers toggle persistence
- Menubar visibility preference
- Default settings

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

## When Working on This Project

1. **No syntax highlighting** - If asked about it, refer to design decision
2. **Follow existing patterns** - Buffer management, command pattern for undo/redo
3. **Test thoroughly** - Especially buffer operations and UI interactions
4. **Maintain simplicity** - This is meant to be a simple, nano-like editor
5. **Cross-platform** - Consider Windows, Linux, and macOS compatibility
6. **Terminal constraints** - Remember this runs in a terminal, not a GUI

## Current State (as of milestone 3+)

The editor is functional with core features complete, including Go to Line (behind flag). Focus is now on:
- Configuration persistence (to remove need for command-line flags)
- Keybinding customization
- Completing soft wrap rendering logic

---

Last Updated: January 9, 2026
