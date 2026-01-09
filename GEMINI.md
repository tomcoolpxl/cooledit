# CoolEdit

CoolEdit is a terminal text editor written in Go, designed to be similar to *nano* but with modern, predictable keyboard shortcuts and correct handling of encodings and newlines.

## Project Overview

*   **Goal:** A simplified, non-modal terminal text editor.
*   **Key Features:**
    *   Cross-platform (Linux/Windows).
    *   Encoding aware (UTF-8, ISO-8859-1).
    *   Newline aware (LF, CRLF).
    *   No syntax highlighting.
    *   **Mouse support:** Disabled by default, enabled via `-mouse` flag.
    *   **Clipboard Integration:** System clipboard support for Copy/Cut/Paste.
    *   **Text Selection:** Shift+Arrow selection support.
*   **Tech Stack:**
    *   Language: Go (1.25.5+)
    *   Terminal Backend: `tcell/v2`
    *   Clipboard: `atotto/clipboard`

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

*   **Buffer:** Line-based gap buffer supporting efficient text manipulation and selection ranges.
*   **View:** The UI layer translates input events into `core.Command`s and renders the state via `term.Screen`.
*   **Encodings/EOL:** Handled explicitly in `fileio`. The internal buffer always uses UTF-8 and `\n`. Conversion happens at the IO boundary.
*   **Clipboard:** Interface-based system clipboard integration.

## Development Conventions

*   **Style:** Standard Go formatting (`gofmt`).
*   **Testing:**
    *   `core` logic should be unit-testable without a terminal.
    *   Run tests with `go test ./...`
*   **Architecture adherence:** Do not import `term` or `ui` packages into `core`. Keep the core logic pure.

## Completed Features

*   Basic Editing (Insert, Delete, Navigation)
*   File IO (Load, Save, Save As, Encoding/EOL support)
*   Undo/Redo
*   Search (Find, Next, Previous)
*   Menubar (Auto-hide, Keyboard navigable)
*   System Clipboard Support (Ctrl+C, Ctrl+X, Ctrl+V)
*   Text Selection (Shift+Arrows, Mouse selection if enabled)
*   Mouse Support (Optional flag)

## Planned / To-Do

*   Configuration (Persistence)
*   Keybinding Customization
*   Go To Line
*   Soft Wrap toggling

## Key Documentation

*   [Architecture](docs/architecture.md)
*   [Requirements](docs/requirements.md)
