# cooledit architecture and source tree

## Goals and constraints

* Terminal text editor similar to nano
* Non-modal editing
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
* Reading key/resize events
* Rendering a grid of cells (runes + style)

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

**Architecture:** Unified search mode with real-time incremental search

**Core Components:**

1. **SearchState** (`core/search.go`)
   - Maintains session-level preferences (CaseSensitive, WholeWord)
   - Stores last query for history and pre-filling
   - References active SearchSession (nil when not searching)
   - Preferences persist across multiple searches within editor session

2. **SearchSession** (`core/search.go`)
   - Represents an active search with pre-computed match positions
   - Contains: Query, Matches array, CurrentIndex, search options
   - Created when user enters search mode
   - Destroyed when user exits search mode
   - Supports navigation between matches via NextMatch()/PrevMatch()

3. **Match** (`core/search.go`)
   - Simple struct representing a single match: Line, Col, Length
   - All matches in buffer are pre-computed for performance

4. **SearchHistory** (`ui/ui.go`)
   - Maintains up to 20 recent search queries
   - Supports bidirectional navigation (up/down arrows)
   - Stores temporary query when navigating history

**Search Algorithm:**

* Linear scan through buffer lines using `strings.Index()` or case-insensitive equivalent
* Supports case-sensitive/insensitive matching
* Supports whole word matching (checks word boundaries)
* `FindAllMatches()` pre-computes all match positions (limited to 1000 for performance)
* Real-time search executes on every keystroke with 150ms debouncing

**Search Mode State Machine:**

```
ModeNormal ──[Ctrl+F]──> ModeSearch ──[Esc/Q]──> ModeNormal
                              │
                              │ [R] ──> ModePrompt (PromptReplaceWith)
                              │             │
                              │             └─[Enter]─> ModeSearch
                              │
                              │ [A] ──> ModeMessage (Confirm Replace All)
                                            │
                                            └─[Y]─> ModeSearch
```

**Key Features:**

* **Real-time search:** Matches appear as you type (debounced 150ms)
* **Visual feedback:** All matches highlighted in viewport, current match distinct color
* **Match navigation:** Works while typing (no mode transition needed)
* **Search options:** Alt+C (case), Alt+W (whole word), toggle with immediate re-search
* **Selection pre-fill:** Ctrl+F with selection pre-fills query
* **Search history:** Up/Down arrows navigate previous 20 queries
* **Replace support:** R (single), A (all with confirmation), undoable
* **Error state:** Red status bar when no matches, mode stays active for correction
* **Performance:** Debouncing, 1000 match limit, "Searching..." indicator

**UI Integration:**

* Status bar shows: query, match count (3 of 15), case/word indicators, shortcuts
* Viewport rendering applies search highlighting via `getSearchMatchStyle()`
* Key handling in `handleSearchKey()` consumes ALL keys to prevent leakage
* Clean state transitions documented with guards and side effects

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
      keymap/             // (Planned - config-file-only, no UI)

    term/
      backend.go          // Screen interface
      tcell/              // tcell implementation

    config/
      config.go           // Config load/save, path management
      schema.go           // Config data structures

    autosave/
      autosave.go         // Autosave manager and storage

    syntax/
      syntax.go           // Chroma-based syntax highlighting

    positionlog/
      positionlog.go      // Cursor position persistence

    formatter/
      formatter.go        // External formatter integration
      executor.go         // Command execution with timeout
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

### Milestone 4 (Complete)
* Theme/Color system with 14 built-in themes + custom themes from config.
* All UI elements have configurable foreground/background colors.
* UI menu support (View → Themes) to switch themes interactively.
* Automatic terminal capability detection and graceful degradation.
* CLI flag `--config <path>` to override config file location.
* Built-in themes: default, dark, light, monokai, solarized-dark/light, gruvbox-dark/light, dracula, nord, dos, ibm-green, ibm-amber, cyberpunk.
* Current line highlight (toggle via View menu, off by default).

### Milestone 5 (Complete)
* Syntax highlighting with Chroma library (50+ languages).
* Language auto-detection and manual selection.
* Theme-integrated syntax colors.

### Milestone 6 (Complete)
* Autosave with idle-based trigger and recovery prompt.
* Cross-platform autosave storage with metadata.

### Milestone 7 (Complete) - Nano-inspired Features
* Smart Home key (cycles between first non-whitespace and column 0).
* Block indent/unindent (Tab/Shift+Tab with selection).
* Comment/Uncomment toggle (Ctrl+/) - language-aware.
* Trim trailing whitespace on save (configurable).
* Position log - remembers cursor position across sessions.
* Scrollbar indicator showing viewport position.
* Verbatim Unicode character input (Ctrl+Shift+U hex, Ctrl+Shift+D decimal).
* External formatter integration (Ctrl+Shift+F) with 20+ built-in language defaults.

### Future (Optional)
* Keybinding customization (config-file-only).
* File browser for open/save operations.
* Linter integration.
