# cooledit

A terminal-based text editor for Linux, macOS and Windows. Similar to nano but with modern keyboard shortcuts and better file handling.

## Features

- Cross-platform terminal UI
- UTF-8 and ISO-8859-1 encoding support with auto-detection
- LF and CRLF line ending detection and preservation
- Undo/redo with full history
- Find and replace with non-overlapping matches
- System clipboard integration (Ctrl+C/X/V)
- Line numbers (toggle with Ctrl+L)
- Word wrap (toggle with Ctrl+W)
- Configurable cursor shapes (block, underline, bar)
- 13 built-in color themes including retro DOS and IBM phosphor styles
- Auto-indentation (preserves leading whitespace on Enter)
- Zen mode (F11 to hide status bar)

## Installation

### Pre-built Binaries

Download from releases page.

### Build from Source

Requires Go 1.25.5 or higher:

```bash
git clone https://github.com/tomcoolpxl/cooledit
cd cooledit
go build ./cmd/cooledit
```

Binary will be in current directory. Move it to your PATH.

## Usage

```bash
# Open a file
cooledit filename.txt

# Create new file
cooledit

# Show line numbers
cooledit -l filename.txt
cooledit --line-numbers filename.txt

# Use custom config file
cooledit -c /path/to/config.toml filename.txt
cooledit --config /path/to/config.toml filename.txt

# Show version
cooledit -v
cooledit --version

# Show help
cooledit -h
cooledit --help
```

## Keyboard Shortcuts

### File Operations
- `Ctrl+S` - Save
- `Ctrl+Shift+S` - Save as
- `Ctrl+Q` - Quit

### Editing
- `Ctrl+Z` - Undo
- `Ctrl+Y` - Redo
- `Ctrl+X` - Cut (current line if no selection)
- `Ctrl+C` - Copy (current line if no selection)
- `Ctrl+V` - Paste
- `Ctrl+A` - Select all
- `Insert` - Toggle insert/replace mode
- `Tab` - Insert spaces to next tab stop
- `Ctrl+I` - Insert literal tab character
- `Backspace` - Delete one character

### Search
- `Ctrl+F` - Find/Replace (unified mode)
- `F3` - Find next
- `Shift+F3` - Find previous
- `Ctrl+G` - Go to line

### Navigation
- `Arrow keys` - Move cursor
- `Shift+Arrow keys` - Select text
- `Home` / `End` - Line start/end
- `Ctrl+Home` - File start
- `Ctrl+End` - File end
- `Page Up` / `Page Down` - Scroll by page

### View
- `Ctrl+L` - Toggle line numbers
- `Ctrl+W` - Toggle word wrap
- `F11` - Toggle status bar (Zen mode)
- `F10` or `Esc` - Toggle menu
- `F1` - Show keyboard shortcuts help

### Menu Shortcuts (when menu is open)
Press the underlined letter to activate menu items:
- File: **S**ave, Save **A**s, **Q**uit
- Edit: **U**ndo, **R**edo, Cu**t**, **C**opy, **P**aste, Select All (**G**rab All)
- Search: **F**ind/Replace, Find **N**ext, Find **P**rev
- Help: **K**eyboard Shortcuts

## Configuration

Config file is auto-created at:
- Linux/macOS: `~/.config/cooledit/config.toml`
- Windows: `%APPDATA%\cooledit\config.toml`

Example configuration:

```toml
[editor]
line_numbers = false
soft_wrap = false
tab_width = 4

[ui]
show_menubar = false
show_statusbar = true
theme = "default"
cursor_shape = "block"  # Options: block, underline, bar

[search]
case_sensitive = true
```

## Themes

13 built-in themes available via View → Themes menu:

1. **default** - Terminal defaults with green cursor
2. **dark** - Dark background, light text
3. **light** - Light background, dark text
4. **monokai** - Popular dark theme with vibrant colors
5. **solarized-dark** - Precision dark color scheme
6. **solarized-light** - Precision light color scheme
7. **gruvbox-dark** - Retro warm dark colors
8. **gruvbox-light** - Retro warm light colors
9. **dracula** - Dark with purple accents
10. **nord** - Arctic bluish theme
11. **dos** - Classic DOS Edit blue background
12. **ibm-green** - Classic green phosphor monitor
13. **ibm-amber** - Classic amber phosphor monitor

Custom themes can be defined in config file using `[themes.custom_name]` sections.

## Find/Replace Mode

Press `Ctrl+F` to enter find/replace mode. After entering search term:

- `N` - Find next match
- `P` - Find previous match
- `R` - Replace current match (prompts for replacement)
- `A` - Replace all matches
- `Q` or `Esc` - Exit find/replace mode

## File Handling

- Encoding auto-detection (UTF-8, ISO-8859-1)
- Line ending auto-detection (LF, CRLF)
- Original encoding and line endings preserved on save
- Files are loaded entirely into memory

## Building

```bash
# Build
go build ./cmd/cooledit

# Run tests
go test ./...

# Run without building
go run ./cmd/cooledit filename.txt
```

## Dependencies

- `github.com/gdamore/tcell/v2` - Terminal handling
- `github.com/atotto/clipboard` - System clipboard integration
- `github.com/BurntSushi/toml` - Configuration file parsing

## Non-Features

- No syntax highlighting (by design)
- No multiple file buffers
- No plugin system
- Single file editing only

## Author

Tom Cool

## License

Copyright (C) 2026 Tom Cool

This program is free software: you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program. If not, see <https://www.gnu.org/licenses/>.

See the [LICENSE](LICENSE) file for the complete license text.

## Source Code

The complete source code for this software is available at:
<https://github.com/tomcoolpxl/cooledit>

## Bug Reports

To report bugs or request features, please open an issue at:
<https://github.com/tomcoolpxl/cooledit/issues>
