# cooledit architecture and source tree

## Goals and constraints

* Terminal text editor similar to nano
* Non-modal editing
* No syntax highlighting
* Cross-platform: Linux and Windows terminals
* Vanilla Go plus minimal terminal layer (no tview)
* Correct handling of:
  * Encodings (UTF-8, ISO-8859-1 at minimum)
  * Newline formats (LF, CRLF) preserved on save
  * Large files without UI lag (reasonable for terminal editor)
* Predictable shortcuts with fallback behavior where terminal limitations exist

This document assumes a stack of:
* Go standard library
* A terminal screen/input backend (recommended: tcell) to implement “ncurses-like” behavior in a portable way

---

## High-level architecture

The project is split into three layers:

### Editor core (pure logic, no terminal dependencies)

Responsible for:
* Text buffer and editing operations
* Cursor, selection state
* Undo/redo (Implemented via `UndoStack` and `Action` pattern)
* Search (Implemented via `SearchState` and linear scan)
* File model (path, modified flag, encoding, EOL type)
* Settings and keybinding resolution (as data, not tied to terminal events)

This layer is unit-tested without a terminal.

### UI model (application state and commands)

Responsible for:
* Mapping key events to editor commands
* Managing dialogs and prompts (Search, Save As, Quit)
* Managing Menubar state (`Menubar` struct)
* Computing Layout (`Menubar`, `Viewport`, `Prompt`, `StatusBar`)
* Rendering to the `term.Screen` interface
* **Clipboard Integration:** Interfacing with system clipboard via `ui.SystemClipboard`

This layer depends on the terminal backend only via a small interface (events in, draw calls out).

### Terminal backend (screen, input, main loop)

Responsible for:
* Entering alternate screen, raw input mode
* Reading key/mouse/resize events
* Rendering a grid of cells (runes + style)
* **Mouse Handling:** Optional capture via `Init(enableMouse bool)`

---

## Core data model

### File model

`FileState`
* `Path string` (empty means [No Name])
* `Encoding Encoding` (enum + optional label)
* `EOL EOLType` (LF or CRLF)
* `Modified bool` (derived from UndoStack state)

### Buffer model

Current Implementation: **Line-based slice buffer**
* `[]Line`, where each `Line` is `[]rune`.
* Simple and effective for typical file sizes.
* Cursor tracking via `(line, col)`.
* **Selection:** Logic for `DeleteRange` and `RangeText`.

### Cursor and viewport

`Cursor`
* `Line int` (0-based)
* `Col int` (0-based in display columns)
* `PreferredCol int` (for up/down movement)

`Viewport`
* `TopLine int`
* `LeftCol int` (horizontal scroll)
* `Height int`
* `Width int`

`Selection`
* `Active bool`
* `Anchor struct { Line, Col int }`

---

## Editor operations and command system

Commands are defined as structs implementing the `Command` interface.
Examples: `CmdInsertRune`, `CmdMoveDown`, `CmdSave`, `CmdUndo`, `CmdFind`, `CmdCopy`, `CmdPaste`.

### Undo/Redo

* Implemented using the Command pattern.
* `UndoStack` stores a history of `Action` objects (e.g., `InsertRuneAction`, `BackspaceAction`, `ReplaceLinesAction`, `DeleteSelectionAction`).
* Each `Action` knows how to `Apply` and `Undo` itself.
* The "Saved" state is tracked by a pointer in the UndoStack to correctly report `Modified` status.

### Search

* `SearchState` stores the last query.
* `Search` function performs linear scan (forward/backward) over the buffer lines.
* UI provides `PromptFind` to capture input.

### Clipboard

* `Clipboard` interface in `core` package.
* `SystemClipboard` implementation in `ui` package using `atotto/clipboard`.
* `CmdCopy`, `CmdCut`, `CmdPaste` utilize this interface.

---

## UI composition

The UI uses a `Layout` struct to partition the screen:
1.  **Menubar** (Top, auto-hide)
2.  **Viewport** (Middle, flexible height)
3.  **Prompt** (Above Status Bar, transient)
4.  **Status Bar** (Bottom, fixed height 1)

Rendering is done layer by layer: Menubar -> Viewport -> Status Bar -> Prompt -> Menu Dropdowns.

---

## Source tree setup

```
cooledit/
  go.mod
  cmd/
    cooledit/
      main.go

  internal/
    app/
      app.go              // wires everything
      lifecycle.go        

    core/
      editor.go           // Editor struct, Apply() logic
      commands.go         // Command definitions
      undo.go             // UndoStack, Actions
      search.go           // Search logic, SearchState
      
      buffer/
        buffer.go         // Buffer interface
        linebuffer.go     // Implementation

    fileio/
      open.go             
      save.go             
      encoding.go         
      eol.go              

    ui/
      ui.go               // UI controller, main loop
      render.go           // Drawing logic
      layout.go           // Layout computation
      menubar.go          // Menubar data model
      clipboard.go        // Clipboard implementation
      prompt.go           // Prompt handling logic
      keymap/             // (Planned)

    term/
      backend.go          // Screen interface
      tcell/              // tcell implementation

    config/
      config.go           // Config load/save, path management
      schema.go           // Config data structures
```

---

## Milestone Status

### Milestone 1 (Complete)
* Basic buffer editing, file load/save.
* Status bar.
* Tcell backend.

### Milestone 2 (Complete)
* Search (Find/Next/Prev).
* Undo/Redo.
* UI Prompts (Save As, Quit, Find).
* Layout Engine.

### Milestone 3 (Complete)
* Menubar (Auto-hide).
* Mouse support (Optional via flag/config).
* Text Selection.
* Clipboard integration.
* Configuration persistence with TOML.
* Toggle settings auto-save.

### Milestone 4 (Planned)
* Keybinding customization.
* Complete soft wrap rendering.
