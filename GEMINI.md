# CoolEdit

CoolEdit is a terminal text editor written in Go, designed to be similar to *nano* but with modern, predictable keyboard shortcuts and correct handling of encodings and newlines.

## Project Overview

*   **Goal:** A simplified, non-modal terminal text editor with modern features
*   **Key Features:**
    *   Cross-platform (Linux/Windows/macOS)
    *   Encoding aware (UTF-8, ISO-8859-1, auto-detection)
    *   Newline aware (LF, CRLF, preserves original format)
    *   No syntax highlighting (by design)
    *   **Mouse support:** Disabled by default, enabled via `-mouse` flag
    *   **Clipboard Integration:** System clipboard support for Copy/Cut/Paste
    *   **Text Selection:** Shift+Arrow selection support
    *   **Configurable cursor shapes:** Block, underline, or bar with smart alternation for replace mode
    *   **13 built-in themes:** Including classic DOS and retro IBM phosphor themes
    *   **DOS-style menu shortcuts:** Underlined letters for quick access (e.g., **S**ave, **Q**uit)
    *   **Automatic menu scrolling:** For small terminal screens with visual indicators
    *   **Secret vim command mode:** Type `:w`, `:q`, `:wq` from menu mode (easter egg)
*   **Tech Stack:**
    *   Language: Go (1.25.5+)
    *   Terminal Backend: `tcell/v2`
    *   Clipboard: `atotto/clipboard`
    *   Config: TOML format

## Building and Running

### Prerequisites

*   Go 1.25.5 or higher.

### Build

```bash
go build ./cmd/cooledit
```

### Run

To run directly from source:

```bash
go run ./cmd/cooledit [filename]
```

To enable mouse support:

```bash
go run ./cmd/cooledit -mouse [filename]
```

To show line numbers by default:

```bash
go run ./cmd/cooledit -line-numbers [filename]
```

## Key Keyboard Shortcuts

**File Operations:**
- `Ctrl+S` - Save
- `Ctrl+Q` - Quit
- `Ctrl+N` - New file

**Editing:**
- `Ctrl+Z` - Undo
- `Ctrl+Y` - Redo
- `Ctrl+X` - Cut
- `Ctrl+C` - Copy
- `Ctrl+V` - Paste
- `Insert` - Toggle Insert/Replace mode

**Search:**
- `Ctrl+F` - Find/Replace (unified nano-style mode)
- `Ctrl+G` - Go to line

**View:**
- `Ctrl+L` - Toggle line numbers
- `Ctrl+W` - Toggle word wrap
- `F11` - Toggle Zen mode (hide status bar)

**Navigation:**
- `F10` or `Esc` - Toggle menu bar
- `F1` - Show keyboard shortcuts help
- `:` (in menu mode) - Secret vim command mode (`:w`, `:q`, `:wq`, `:q!`, `:w!`)

**Menu Shortcuts (when menubar is open):**
- Press underlined letter to activate (e.g., `S` for Save, `Q` for Quit)

## Themes

13 built-in themes available via View → Themes menu:
1. **default** - Green cursor, terminal defaults
2. **dark** - Dark background, light text
3. **light** - Light background, dark text
4. **monokai** - Popular dark theme
5. **solarized-dark** / **solarized-light** - Precision color schemes
6. **gruvbox-dark** / **gruvbox-light** - Retro groove colors
7. **dracula** - Dark purple accents
8. **nord** - Arctic bluish theme
9. **dos** - Classic DOS Edit (blue background)
10. **ibm-green** - Classic green phosphor monitor
11. **ibm-amber** - Classic amber phosphor monitor

## Architecture

The project follows a strict layered architecture to ensure testability and separation of concerns.

### Directory Structure

*   `cmd/cooledit/`: Entry point (`main.go`).
*   `internal/`: Private application code.
    *   `app/`: Application wiring and lifecycle.
    *   `core/`: Pure editor logic (buffer, cursor, undo/redo, search, selection). **No terminal dependencies.**
    *   `ui/`: User interface model (rendering, dialogs, clipboard integration).
    *   `term/`: Terminal backend abstraction (wraps `tcell`).
    *   `fileio/`: File reading/writing, encoding detection, EOL handling.
    *   `config/`: User configuration settings.
*   `docs/`: Documentation (Architecture, Requirements).

### Core Concepts

*   **Buffer:** Line-based gap buffer supporting efficient text manipulation and selection ranges
*   **View:** The UI layer translates input events into `core.Command`s and renders the state via `term.Screen`
*   **Encodings/EOL:** Handled explicitly in `fileio`. The internal buffer always uses UTF-8 and `\n`. Conversion happens at the IO boundary
*   **Clipboard:** Interface-based system clipboard integration
*   **Themes:** 13 hardcoded built-in themes with custom theme support via config
*   **Cursor:** Configurable shapes (block/underline/bar) with theme-based colors and smart alternation for replace mode

## Configuration

Config file location:
- Linux/macOS: `~/.config/cooledit/config.toml`
- Windows: `%APPDATA%\cooledit\config.toml`

Auto-created on first use. Example settings:

```toml
[editor]
line_numbers = false
soft_wrap = false
tab_width = 4

[ui]
show_menubar = false
show_statusbar = true
mouse_enabled = false
theme = "default"
cursor_shape = "block"  # Options: "block", "underline", "bar"

[search]
case_sensitive = true
```

## Development Conventions

*   **Style:** Standard Go formatting (`gofmt`).
*   **Testing:**
    *   `core` logic should be unit-testable without a terminal.
    *   Run tests with `go test ./...`
*   **Architecture adherence:** Do not import `term` or `ui` packages into `core`. Keep the core logic pure.

## Completed Features

*   ✅ Basic Editing (Insert, Delete, Navigation)
*   ✅ File IO (Load, Save, Save As, Encoding/EOL auto-detection and preservation)
*   ✅ Undo/Redo (atomic operations with full history)
*   ✅ Search & Replace (Unified nano-style mode, non-overlapping matches)
*   ✅ Menubar (Auto-hide, keyboard navigable with DOS-style shortcuts)
*   ✅ System Clipboard Support (Ctrl+C, Ctrl+X, Ctrl+V)
*   ✅ Text Selection (Shift+Arrows, Mouse selection if enabled)
*   ✅ Mouse Support (Optional `-mouse` flag)
*   ✅ Configuration System (TOML persistence, auto-save on toggles)
*   ✅ Go To Line (Ctrl+G)
*   ✅ Soft Wrap (Ctrl+W toggle)
*   ✅ Insert/Replace Mode (Insert key toggle with cursor shape indicators)
*   ✅ Cursor Customization (3 shapes: block/underline/bar, theme-based colors)
*   ✅ Theme System (13 built-in themes + custom theme support)
*   ✅ Help Screen (F1, adaptive two-column layout)
*   ✅ Zen Mode (F11 - hide status bar for distraction-free editing)
*   ✅ Menu Scrolling (Automatic scrolling with ↑/↓ indicators on small screens)
*   ✅ Secret Vim Command Mode (`:w`, `:q`, `:wq`, `:q!`, `:w!` from menu)
*   ✅ Terminal State Preservation (Cursor color restored on exit)

## Planned / To-Do

*   Additional features as requested

## Key Documentation

*   [Architecture](docs/architecture.md)
*   [Requirements](docs/requirements.md)
