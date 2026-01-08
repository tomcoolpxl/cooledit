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

If you truly mean zero third-party deps, you can do raw terminal control with ANSI + termios on Unix and Console APIs on Windows, but that becomes a platform project. The rest of the architecture remains the same.

---

## High-level architecture

The project is split into three layers:

### Editor core (pure logic, no terminal dependencies)

Responsible for:

* Text buffer and editing operations
* Cursor, selection (even if you do not expose selection yet), viewport state
* Undo/redo
* Search/replace
* File model (path, modified flag, encoding, EOL type)
* Settings and keybinding resolution (as data, not tied to terminal events)

This layer should be unit-testable without a terminal.

### UI model (application state and commands)

Responsible for:

* Mapping key events to editor commands
* Managing dialogs and mode-like prompts (search input, go-to-line input, replace prompt)
* Status bar messages and transient errors
* Menu model (optional) and toggles (line numbers, wrap, menubar)
* Clipboard integration policy (internal clipboard vs system)

This layer depends on the terminal backend only via a small interface (events in, draw calls out).

### Terminal backend (screen, input, main loop)

Responsible for:

* Entering alternate screen, raw input mode
* Reading key/mouse/resize events
* Rendering a grid of cells (runes + style)
* Cursor show/hide and position
* Frame scheduling (event-driven redraw, optional throttling)

---

## Core data model

### File model

`FileState`

* `Path string` (empty means [No Name])
* `Encoding Encoding` (enum + optional label)
* `EOL EOLType` (LF or CRLF)
* `Modified bool`
* `ReadOnly bool` (optional; detect permission errors on save)

### Buffer model

The buffer is the most important decision.

For “nano-like” scope, you can choose:

Option A (simpler, acceptable for moderate files): line-based gap buffer per line

* Store `[]Line`, where each `Line` holds a `[]rune` (or `[]byte` with UTF-8 indexing helpers)
* Edits are localized, but inserts at start of long lines cost more
* Easier to implement column calculations

Option B (better for large files and frequent edits): piece table

* Keep original file bytes and add buffer bytes, represent document as pieces
* Maintain a line index for fast row/col mapping
* More complex but scales better

Given your “should handle large files without UI lag”, implement:

* Piece table OR line-based with incremental line index and cautious redraw

Recommended compromise:

* Line-oriented storage with a separate “line index” and per-line rope/gap buffer later if needed
* Start with line-based, but keep interfaces so you can swap implementation

### Cursor and viewport

`Cursor`

* `Line int` (0-based)
* `Col int` (0-based in display columns, not bytes)
* `PreferredCol int` (for up/down movement retaining visual column)

`Viewport`

* `TopLine int`
* `LeftCol int` (horizontal scroll when wrap off)
* `Height int`
* `Width int`

Word wrap behavior:

* Wrap OFF by default
* When wrap ON: viewport `LeftCol` should be 0 and horizontal scrolling disabled; use visual line wrapping in renderer only

Line numbers:

* When enabled: compute gutter width = digits(lineCount) + 1 padding
* Renderer uses gutter width to offset text region

---

## Editor operations and command system

Define a command interface so UI can map keys to commands cleanly.

`Command` (data or interface) examples:

* `CmdInsertRune(rune)`
* `CmdBackspace`
* `CmdDelete`
* `CmdNewline`
* `CmdMoveLeft/Right/Up/Down`
* `CmdPageUp/PageDown`
* `CmdHome/End` (line start/end)
* `CmdFileStart/FileEnd`
* `CmdGoToLine(n)`
* `CmdCutLine`
* `CmdCopyLine`
* `CmdPaste`
* `CmdUndo/CmdRedo`
* `CmdFind(query, direction)`
* `CmdReplace(query, repl, mode)`
* `CmdToggleLineNumbers`
* `CmdToggleWordWrap`
* `CmdToggleMenubar`
* `CmdSave`
* `CmdSaveAs(path)`
* `CmdQuit`

The UI should never mutate buffer directly; it issues commands to the core.

Undo/redo

* Use an operation log of “edits”:

  * Insert range
  * Delete range
  * Replace range
* Group operations into transactions for:

  * continuous typing
  * paste
  * replace-all
* Store positions in a stable coordinate system (line/col) plus internal offsets depending on buffer design

---

## Encoding and EOL handling

### On open

Pipeline:

1. Read file bytes
2. Detect EOL:

  * Use a bounded scan (e.g., first 64 KiB) rather than the whole file; if any `\r\n` is found in that window, treat as CRLF; otherwise default to LF
  * Normalize to internal representation (store as `\n` in buffer)
3. Detect encoding:

   * If UTF-8 with BOM, treat as UTF-8 and strip BOM
  * If valid UTF-8, treat as UTF-8 (this is the default assumption, matching modern tools)
  * Otherwise fall back to user-configured default (initial default is UTF-8; legacy fallback like ISO-8859-1 only if explicitly configured)
4. Decode to internal representation:

   * Internal should be Unicode text in runes or UTF-8 bytes with helpers

Status bar shows:

* Encoding label
* EOL type

### On save

1. Serialize internal buffer to bytes using chosen encoding
2. Convert internal `\n` to desired EOL (preserve original unless user changed)
3. Write to temp file then atomic rename where possible
4. Update modified flag and status message

Implement “Save As” which:

* Sets `Path`
* Leaves encoding/EOL unchanged unless user chooses

---

## UI composition

Everything is drawn manually with a renderer that knows screen dimensions.

Regions:

* Optional menubar at top (height 1)
* Main editor viewport in the middle (height = screenHeight - menubar - statusbar - promptbar)
* Prompt/dialog line(s) above status bar (height 1 or more, but keep minimal)
* Status bar at bottom (height 1)

Dialogs:

* For critical confirmations (overwrite, unsaved changes), use a centered modal box:

  * Draw a rectangle
  * Draw text
  * Draw buttons or key hints
  * Capture input until dismissed

Search/go-to-line:

* Use inline prompt bar above status bar:

  * Example: `Find: <input>` with live editing
  * Enter confirms, Esc cancels
  * F3 / Shift+F3 trigger next/prev using last query

Menu:

* Menubar visibility is optional; default startup state is hidden (no menubar drawn).
* When menubar enabled:

  * Alt+F, Alt+E, Alt+S, Alt+V to open menu
  * Arrow keys navigate
  * Enter selects
  * Esc closes
* When hidden:

  * Alt shortcuts still open it as overlay OR ignore (your choice). Requirement suggests it should be accessible.

---

## Input handling and keybinding strategy

Terminal reality:

* Some combinations (Shift+F3, Ctrl+Home) are not consistently distinguishable on all terminals.
* Define fallback keys for every “non-portable” binding and expose them in help.

Keybinding system should support:

* Primary binding (requested)
* Fallback binding(s)
* User overrides in config

Example fallback suggestions:

* File start/end: Ctrl+Home/Ctrl+End fallback to Alt+< / Alt+>
* Find previous: Shift+F3 fallback to Alt+F3 or Ctrl+Shift+F (if available) or Ctrl+P (if not conflicting)
* Menubar: Ctrl+M fallback to F10

Quit / Esc policy:

* Esc is only for dismissing dialogs, menus, and prompts; it must never exit the application.
* The quit gesture is pressing Ctrl+C twice (a single press should not terminate the app without a second confirmation press).

Implementation approach:

* Normalize raw terminal key events to internal `KeyEvent`:

  * `Key` (enum: Rune, Enter, Backspace, Delete, Up, Down, F1..F12, etc.)
  * `Rune` (if Key==Rune)
  * `Modifiers` (Ctrl, Alt, Shift)
* Then match against keybinding table to yield a `Command`

---

## Performance model

To avoid UI lag:

* Rendering should be incremental:

  * Only redraw dirty regions: viewport lines that changed, status bar, prompt bar
* Keep a line cache for the viewport:

  * Rendered slices for each visible line
  * Invalidate on edits affecting those lines or on horizontal scroll changes
* Avoid converting entire file to lines on every change
* Maintain a line index:

  * Map line number to internal offsets
  * Update incrementally on newline insert/delete
* Search:

  * For “find next/prev”, avoid scanning entire file every keypress
  * Implement forward scan from cursor position with reasonable chunking

For milestone 1, it is acceptable to redraw the viewport each event for smaller files, but keep the renderer structured to optimize later.

---

## Error handling and status messaging

Status bar should show:

* `[No Name]` or filename
* Modified marker, e.g. `*`
* `Ln X, Col Y`
* `UTF-8` / `ISO-8859-1`
* `LF` / `CRLF`
* A hint that F1 opens Help (pressing F1 pops a minimal help dialog)
* Right-aligned message area (transient)

Message rules:

* Success messages time out after N seconds or next keypress
* Errors persist until acknowledged (Esc or Enter)

---

## Configuration

Settings file:

* Location:

  * Linux: `~/.config/cooledit/config.json`
  * Windows: `%APPDATA%\cooledit\config.json`
* Store:

  * `lineNumbers bool`
  * `wordWrap bool`
  * `menuBar bool` (default false; no menubar shown unless enabled)
  * `theme string` (named color theme selection; default "default"; controls editor/status/menu/dialog colors)
  * `keymap map[string][]Binding` (command name -> list of bindings)
  * `defaultEncoding string` (default UTF-8, matching common editor defaults)
* Load on startup, apply to UI state
* Save on explicit action or on exit

---

## Source tree setup

This structure keeps the core testable and UI/backend replaceable.

```
cooledit/
  go.mod
  cmd/
    cooledit/
      main.go

  internal/
    app/
      app.go              // wires everything: backend + ui + core, main loop orchestration
      lifecycle.go        // open file from args, quit flow, save prompts

    core/
      editor.go           // Editor: buffer + cursor + viewport, Apply(Command)
      commands.go         // Command definitions and constructors
      selection.go        // (optional now) selection model for future
      undo.go             // undo/redo stack, transactions
      search.go           // find, find next/prev, replace logic

      buffer/
        buffer.go         // Buffer interface
        linebuffer.go     // initial implementation (line-based)
        piece_table.go    // optional later
        position.go       // Position types, range types

      text/
        width.go          // rune width and grapheme helpers (as needed)
        normalize.go      // normalize newlines internally

    fileio/
      open.go             // read file bytes, detect encoding and eol, decode
      save.go             // encode, eol conversion, atomic write, overwrite checks
      encoding.go         // Encoding enum, UTF-8 validation, ISO-8859-1 codec
      eol.go              // EOLType and helpers
      detect.go           // detection routines

    ui/
      ui.go               // UI controller: dialogs, prompt state, status messages
      render.go           // renderer: draws to screen interface
      layout.go           // compute regions (menubar, viewport, prompt, status)
      menubar.go          // menu model and interactions
      prompt.go           // find/replace/go-to-line input handling
      dialogs.go          // confirmation dialogs
      clipboard.go        // internal clipboard; optional system integration later

      keymap/
        keymap.go         // keybinding resolution, default bindings
        bindings.go       // Binding struct, parsing/formatting
        defaults.go       // default keymap per requirements

    term/
      backend.go          // Screen interface abstraction
      tcell/
        tcell.go          // tcell implementation of backend
        keys.go           // map tcell events to internal KeyEvent
        screen.go         // init/teardown, alternate screen, cursor

    config/
      config.go           // load/save config, paths per OS
      schema.go           // config structs and defaults

  assets/
    README.md             // optional: keybinding table, manual

  docs/
    requirements.md       // your requirements (source of truth)
    architecture.md       // this document
    keybindings.md        // generated or maintained docs
```

Notes on why this tree works:

* `internal/core` has no terminal dependency.
* `internal/term/backend.go` defines a minimal interface so you can swap tcell or implement raw ANSI later.
* `internal/ui` is responsible for translating input events into `core.Command` and drawing via `term.Screen`.
* `internal/fileio` isolates encoding and newline behavior so the editor core remains encoding-agnostic.

---

## Minimal interfaces (contract between layers)

### Terminal backend interface

`internal/term/backend.go` should define something like:

* `Screen.Init() error`
* `Screen.Fini()`
* `Screen.Size() (w, h int)`
* `Screen.PollEvent() Event` (KeyEvent, MouseEvent, ResizeEvent)
* `Screen.SetCell(x,y, rune, style)`
* `Screen.ShowCursor(x,y)`
* `Screen.HideCursor()`
* `Screen.Show()`

Keep it small. The UI renderer should be able to draw without caring about backend details.

### UI <-> core

UI calls:

* `editor.Apply(cmd Command) Result`
* `editor.GetViewModel() ViewModel` (or individual getters)

`Result` should include:

* `DirtyLines []int` or `DirtyRange`
* `StatusMessage string` or error
* `CursorChanged bool`
* `ViewportChanged bool`

This allows efficient redraw decisions.

---

## Milestone mapping to modules

### Milestone 1

* `core/editor.go`, `buffer/linebuffer.go`
* `fileio/open.go`, `fileio/save.go` basic UTF-8 + preserve EOL
* `ui/render.go` viewport + status bar
* `term/tcell` backend
* Basic keymap: arrows, typing, backspace/delete, save, quit via double Ctrl+C (Esc reserved for cancel/close only)

### Milestone 2

* `core/search.go`, `ui/prompt.go` find/replace/go-to-line
* Encoding detection/display + reopen/convert hooks
* Status messaging behavior

### Milestone 3

* `ui/menubar.go` + accelerators
* `ui/layout.go` + toggles
* Optional line numbers column

### Milestone 4

* `config/config.go` persistence
* User keymap overrides with validation

---

## Implementation notes that prevent common bugs

* Store cursor column in “visual columns” and translate to buffer index per line using rune width.
* Never treat `len(string)` as screen columns.
* Internally normalize newlines to `\n`. Track original EOL type separately.
* Encoding conversion must happen only at file boundaries (open/save), not during editing.
* For undo/redo, store operations in terms of positions and inserted/deleted text in internal Unicode form, not encoded bytes.
* Add an explicit fallback keybinding table now, even if you do not expose config yet, because some bindings will not work everywhere.

---

## Default file layout and naming conventions

* Package names match folder names (`core`, `fileio`, `ui`, `term`, `config`)
* Keep exported identifiers minimal inside `internal/`
* Prefer small files per responsibility; terminal editors become unmaintainable when everything is in one file

---
