# cooledit – Initial Architecture Design

## 1. Architectural Goals

* Simple, predictable structure.
* Clear separation between:

  * UI (tview)
  * Editor state
  * Text buffer and file I/O
* No modal editor complexity.
* Easy to reason about key handling and rendering.
* Extensible without committing to plugins or tabs.

Non-goals at this stage:

* Plugin system
* Multi-buffer editing
* Syntax highlighting

---

## 2. High-Level Component Overview

```
cmd/
  cooledit/
    main.go

internal/
  app/
    app.go
    lifecycle.go

  ui/
    layout.go
    menubar.go
    statusbar.go
    editorview.go
    dialogs.go

  editor/
    buffer.go
    cursor.go
    selection.go
    undo.go

  fileio/
    open.go
    save.go
    encoding.go
    newline.go

  input/
    keymap.go
    dispatcher.go
    commands.go

  config/
    config.go
    defaults.go

  model/
    state.go
    viewstate.go

pkg/
  util/
    text.go
    errors.go
```

---

## 3. Core Runtime Flow

1. `main.go`

   * Parses CLI arguments.
   * Initializes config.
   * Creates App instance.
   * Calls `Run()`.

2. `app.App`

   * Owns the tview Application.
   * Owns the global EditorState.
   * Wires UI components together.
   * Routes input events.

3. UI renders state, never mutates it directly.

4. Commands mutate state, then trigger redraw.

---

## 4. Application Layer (`internal/app`)

### Responsibilities

* Application lifecycle
* Startup and shutdown
* High-level error handling
* Coordinating UI and editor state

### Key Types

```go
type App struct {
    TVApp     *tview.Application
    State     *model.EditorState
    UI        *ui.RootLayout
    Keymap    *input.Keymap
}
```

### Lifecycle

* Initialize editor state
* Load file (if provided)
* Detect encoding and newline format
* Start event loop

---

## 5. Editor State Model (`internal/model`)

This is the single source of truth.

### EditorState

```go
type EditorState struct {
    Buffer        *editor.Buffer
    Cursor        *editor.Cursor
    UndoStack     *editor.UndoStack

    Filename      string
    Modified      bool

    Encoding      fileio.Encoding
    Newline       fileio.NewlineType

    View          ViewState
}
```

### ViewState

UI-only flags, never persisted to disk.

```go
type ViewState struct {
    ShowLineNumbers bool
    ShowMenuBar     bool
    WordWrap        bool

    StatusMessage   string
}
```

---

## 6. Text Buffer (`internal/editor`)

### Design Principles

* Line-based storage (slice of strings or rope later).
* Cursor is separate from buffer.
* No rendering logic inside buffer.

### Buffer

```go
type Buffer struct {
    Lines []string
}
```

Responsibilities:

* Insert/delete text
* Insert/delete lines
* Return slices for rendering

### Cursor

```go
type Cursor struct {
    Line int
    Col  int
}
```

Cursor movement rules:

* Always clamped to valid buffer positions.
* Column snapping when moving across lines.

### Undo / Redo

Command-based undo, not diff-based.

```go
type EditAction interface {
    Undo(*Buffer)
    Redo(*Buffer)
}
```

---

## 7. File I/O Layer (`internal/fileio`)

### Responsibilities

* Open files
* Detect encoding
* Detect newline format
* Save files preserving metadata

### Encoding Detection

* Default UTF-8.
* Fallback detection (simple heuristics).
* Encoding stored in EditorState and shown in status bar.

### Newline Detection

* Scan on load:

  * `\r\n` => CR+LF
  * `\n` only => LF
* Preserve on save unless explicitly changed.

---

## 8. UI Layer (`internal/ui`)

### Root Layout

Uses tview Flex:

```
+----------------------+
| Menubar (optional)   |
+----------------------+
| Editor View          |
|                      |
+----------------------+
| Status Bar           |
+----------------------+
```

### Editor View

* Custom wrapper around `tview.TextView`
* Responsible for:

  * Rendering visible lines
  * Applying word wrap (visual only)
  * Drawing line numbers if enabled
* Never edits buffer directly.

### Status Bar

Single-line TextView.
Shows:

* Filename
* Modified flag
* Line:Col
* Encoding
* LF / CR+LF
* Temporary messages

### Menubar

* tview List or custom primitive
* Activated via Alt shortcuts
* Dispatches commands, not logic

### Dialogs

* Implemented using tview Modal or Form
* Search, replace, go-to-line, save-as
* Push/pop from layout stack

---

## 9. Input Handling (`internal/input`)

### Design Goals

* No hardcoded behavior in UI widgets.
* All keys map to commands.
* Commands operate on EditorState.

### Keymap

```go
type Keymap struct {
    bindings map[tcell.Key]Command
}
```

Supports:

* Ctrl combinations
* Alt combinations
* Function keys

### Commands

```go
type Command interface {
    Execute(*model.EditorState) error
}
```

Examples:

* SaveFileCommand
* MoveCursorCommand
* ToggleLineNumbersCommand
* FindCommand

Dispatcher:

* Receives key event
* Looks up command
* Executes command
* Triggers redraw

---

## 10. Configuration (`internal/config`)

### Stored Settings

* Show line numbers
* Show menubar
* Default word wrap
* Key bindings

### Design

* Simple config file (TOML or JSON).
* Load at startup.
* Save on explicit user action or exit.

---

## 11. Error Handling Strategy

* Non-fatal errors:

  * Display in status bar.
* Fatal errors:

  * Modal dialog, then exit.
* No panics beyond main.

---

## 12. Rendering and Performance Notes

* Render only visible lines.
* Avoid full redraw on cursor move.
* Keep buffer operations O(1) per line where possible.
* Large-file support prioritized over UI polish.

---

## 13. Suggested Implementation Order

1. EditorState + Buffer + Cursor
2. File open/save with encoding and newline detection
3. Basic editor view rendering
4. Status bar
5. Keymap and command dispatcher
6. Search and replace
7. Menubar
8. Config persistence

---

## 14. Known Risks

* tview TextView limitations for large files.
* Undo stack memory growth.
* Terminal key inconsistencies across platforms.

Mitigation: keep abstractions thin and replaceable.
