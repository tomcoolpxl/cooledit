# cooledit

A terminal-based text editor for Linux, macOS and Windows. Similar to nano but with modern keyboard shortcuts and better file handling. And maybe a tiny bit larger in size :)

## Features

- Cross-platform terminal UI
- UTF-8 and ISO-8859-1 encoding support with auto-detection
- LF and CRLF line ending detection and preservation
- Undo/redo with full history
- Find and replace with unified search mode (real-time incremental search)
- Search history navigation and case-sensitive/whole word options
- System clipboard integration (Ctrl+C/X/V)
- Line numbers (toggle with Ctrl+L)
- Word wrap (toggle with Ctrl+W)
- Configurable cursor shapes (block, underline, bar)
- 14 built-in color themes including retro DOS and IBM phosphor styles
- Syntax highlighting with 50+ language support (Chroma-based)
- Whitespace visualization (spaces, tabs, line endings)
- Auto-indentation (preserves leading whitespace on Enter)
- Zen mode (F11 to hide status bar)
- Bracket matching and navigation (Ctrl+B)
- Current line highlight (toggle via View menu)
- Autosave with automatic recovery on startup
- Smart Home key (cycles between first non-whitespace and column 0)
- Block indent/unindent (Tab/Shift+Tab with selection)
- Comment/Uncomment toggle (Ctrl+/) - language-aware
- Trim trailing whitespace on save (configurable)
- Position log - remembers cursor position across sessions
- Scrollbar indicator showing viewport position
- Verbatim Unicode character input (Ctrl+Shift+U hex, Ctrl+Shift+D decimal)

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

# Create new file (opens for editing even if doesn't exist)
cooledit newfile.txt

# Create file without name
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
- `Ctrl+/` - Toggle comment on line or selection
- `Insert` - Toggle insert/replace mode
- `Tab` - Insert spaces to next tab stop (or indent selection)
- `Shift+Tab` - Unindent line or selection
- `Ctrl+I` - Insert literal tab character
- `Ctrl+Shift+U` - Insert Unicode character by hex code point
- `Ctrl+Shift+D` - Insert Unicode character by decimal code point
- `Backspace` - Delete one character

### Search
- `Ctrl+F` - Enter unified search mode (incremental search with real-time results)
- `Alt+C` - Toggle case sensitivity (in search mode)
- `Alt+W` - Toggle whole word matching (in search mode)
- `F3` / `Shift+F3` - Find next/previous match
- `Enter` - Navigate to next match (in search mode)
- `Up` / `Down` - Navigate search history (in search mode)
- `Ctrl+R` - Replace current match (in search mode)
- `Ctrl+H` - Replace all matches (in search mode, with confirmation)
- `Escape` - Exit search mode
- `Ctrl+G` - Go to line

### Navigation
- `Arrow keys` - Move cursor
- `Shift+Arrow keys` - Select text
- `Ctrl+Left` / `Ctrl+Right` - Jump by word
- `Home` - Smart home (first non-whitespace, then column 0)
- `End` - Line end
- `Ctrl+Home` - File start
- `Ctrl+End` - File end
- `Page Up` / `Page Down` - Scroll by page
- `Ctrl+B` - Jump to matching bracket

### View
- `Ctrl+L` - Toggle line numbers
- `Ctrl+W` - Toggle word wrap
- `Ctrl+Shift+W` - Toggle whitespace display
- `Ctrl+H` - Toggle syntax highlighting
- `F11` - Toggle status bar (Zen mode)
- `F10` or `Esc` - Toggle menu
- `F1` - Show keyboard shortcuts help
- Additional toggles via View menu: current line highlight, scrollbar, trim whitespace on save, remember cursor position

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
syntax_highlighting = true
show_whitespace = false
current_line_highlight = false      # Highlight current line with different background
trim_trailing_whitespace = false    # Trim trailing whitespace on save
remember_position = true            # Remember cursor position across sessions
show_scrollbar = false              # Show scrollbar indicator

[ui]
show_menubar = false
show_statusbar = true
theme = "default"
cursor_shape = "block"  # Options: block, underline, bar
cursor_blink = true

[search]
case_sensitive = false
whole_word = false

[autosave]
enabled = true
idle_timeout = 2    # seconds before autosave
min_interval = 30   # minimum seconds between autosaves
```

## Themes

14 built-in themes available via View → Themes menu:

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
14. **cyberpunk** - Neon theme with hot pink and cyan

Custom themes can be defined in config file using `[themes.custom_name]` sections.

## Unified Search Mode

Press `Ctrl+F` to enter unified search mode with real-time incremental search:

### Searching
- Type to search - matches appear immediately as you type (all letters and characters are typed)
- Search highlights all matches in the viewport
- Status bar shows match count (e.g., "Match 3 of 15")
- `Enter` - Navigate to next match
- `F3` / `Shift+F3` - Navigate next/previous match
- `Up` / `Down` - Navigate search history
- `Backspace` - Delete character from query (or exit if query is empty)

### Options
- `Alt+C` - Toggle case sensitivity (indicator shows "Match Case" or "Ignore Case")
- `Alt+W` - Toggle whole word matching (indicator shows "Whole Word")
- Search preferences persist across searches within the session

### Replacing
- `Ctrl+R` - Replace current match (prompts for replacement text)
- `Ctrl+H` - Replace all matches (shows confirmation dialog with match count)
- Replace-all is undoable as a single operation

### Exiting
- `Escape` - Exit search mode and return to normal editing
- Query is saved to history for next search

### Features
- Pre-fills search from current selection (if text is selected)
- Real-time match highlighting (current match vs other matches)
- Visual error state when no matches found (red status bar)
- Debounced search (waits 150ms after last keystroke)
- Performance limit: up to 1000 matches (shows "1000+" if more exist)

## File Handling

- Encoding auto-detection (UTF-8, ISO-8859-1)
- Line ending auto-detection (LF, CRLF)
- Original encoding and line endings preserved on save
- Files are loaded entirely into memory

## Autosave

Cooledit automatically saves backup copies of your work to prevent data loss.

### How it Works
- Autosave triggers after 2 seconds of inactivity (configurable)
- Backups are stored in a dedicated directory:
  - Windows: `%APPDATA%\cooledit\autosave\`
  - Linux: `~/.local/share/cooledit/autosave/`
  - macOS: `~/Library/Application Support/cooledit/autosave/`

### Recovery
If cooledit finds an autosave backup when opening a file, it shows a recovery prompt:
- **[R]ecover** - Load the autosave content (marks file as modified)
- **[O]pen original** - Load the original file (keeps autosave for later)
- **[D]iscard** - Delete the autosave and load original file

### Lifecycle
- Autosave is **cleared** when you save with Ctrl+S
- Autosave is **kept** if you quit without saving (for future recovery)
- Toggle autosave via View → Autosave menu

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
