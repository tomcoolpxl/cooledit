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
- ✅ Tab handling with smart indentation (Tab inserts spaces, Ctrl+Tab inserts literal \t)

### Implemented (Milestone 4 - Theme System)
- ✅ Term.Style extended with Foreground/Background color fields
- ✅ Theme package with comprehensive color element definitions
- ✅ 11 built-in themes: default, dark, light, monokai, solarized-dark/light, gruvbox-dark/light, dracula, nord, dos
- ✅ Custom theme support from config file
- ✅ All UI elements use theme colors (editor, menubar, status bar, prompt, help)
- ✅ View menu with theme selection (checkmarks for current theme)
- ✅ Interactive theme switching with auto-save to config
- ✅ Graceful terminal capability detection via tcell (true color → 256 → 16 → monochrome)
- ✅ Backward compatibility with "default" theme using inverse video
- ✅ Menu backgrounds fixed to be distinct from editor backgrounds in all themes

### Planned (Milestone 4 Remaining)
- ⏳ Add --config CLI flag for alternate config file location
- ⏳ Add tests for theme system (ParseColor, theme loading, switching)

### Future/Optional
- Keybinding customization (config-file-only, no UI)

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
- `Tab` - Insert spaces to next tab stop (configurable width, default: 4)
- `Ctrl+Tab` - Insert literal tab character (\t)
- `Insert` - Toggle Insert/Replace mode
- `F1` - Help overlay (adaptive two-column/single-column layout)
- `F10` / `Esc` - Toggle menubar

## Key Design Decisions

### Status Bar (Always Visible, Priority-Based Layout)
**Left** (priority 2): Filename with modified indicator (`*`)
**Center** (priority 3): Mini-help with adaptive display:
  - `F1 Help` → `Esc/F10 Menu` → `Ctrl+Q Quit` → `Ctrl+S Save` → `Ctrl+F Find/Replace`
  - Shows as many items as fit, removes from right to left when space is limited
**Right** (priority 1): `REPLACE  Ln X, Col Y  Encoding EOL` (replace mode indicator when active, cursor position and file status)

### Menubar (Auto-hidden by Default)
- Toggle with F10 or Esc
- Menus: File, Edit, Search, View, Help
- View menu includes:
  - Toggle Line Numbers
  - Toggle Word Wrap
  - Themes (submenu with all available themes)
- Navigation via arrow keys or mouse (if enabled)

### Command-Line Flags
- `-mouse` - Enable mouse support (click to position cursor, scroll)
- `-line-numbers` - Show line numbers column
- `--config <path>` - Use alternate config file location

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
- Replace mode shows **"REPLACE"** indicator in status bar
- Replace mode overwrites characters instead of inserting
- At end of line, behaves like insert mode
- State not saved - always starts in insert mode

### Tab Handling
- **Tab key inserts spaces** (not literal `\t` characters)
- **Configurable tab width** (default: 4 spaces, set via `tab_width` in config)
- **Smart indentation**: Tab moves cursor to next tab stop (e.g., column 4, 8, 12...)
- **Smart backspace**: When in leading whitespace, backspace removes one indentation unit
- **Literal tabs**: Press `Ctrl+Tab` to insert a raw `\t` character
- **Rendering**: Literal tab characters render with proper width via tcell
- **Undo/Redo**: Tab insertion and smart backspace are atomic operations
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
  - Status bar replace indicator (TestStatusBarReplaceModeIndicator)

### EOL Format
- Auto-detect (LF vs CRLF)
- Display in status bar
- Preserve original format on save

## Theme System (Implemented - Milestone 4)

**Built-in Themes:**
11 hardcoded themes that work out of the box without any configuration:
1. `default` - Uses terminal defaults with inverse video (backward compatibility)
2. `dark` - Dark background with light text (simple, high contrast)
3. `light` - Light background with dark text (simple, high contrast)
4. `monokai` - Popular dark theme with purple, pink, yellow, green accents
5. `solarized-dark` - Ethan Schoonover's precision dark color scheme
6. `solarized-light` - Ethan Schoonover's precision light color scheme
7. `gruvbox-dark` - Retro groove colors, warm dark background
8. `gruvbox-light` - Retro groove colors, warm light background
9. `dracula` - Dark theme with purple accents and pink highlights
10. `nord` - Arctic bluish dark theme inspired by northern lights
11. `dos` - Classic DOS Edit colors (blue background, white/cyan text)

**Custom Themes:**
Users can define additional themes in config file using `[themes.custom_name]` sections.

**UI Support:**
- View → Themes menu shows all available themes (built-in + custom)
- Click to switch theme (saves selection to config)
- Current theme indicated with checkmark

**Color Elements:**
Each element has `fg` (foreground) and `bg` (background) properties.

- **editor**: `fg`, `bg`, `selection_fg`, `selection_bg`, `line_numbers_fg`, `line_numbers_bg`
- **search**: `match_fg`, `match_bg`, `current_match_fg`, `current_match_bg`
- **statusbar**: `fg`, `bg`, `filename_fg`, `modified_fg`, `position_fg`, `mode_fg`, `help_fg`
- **menubar**: `fg`, `bg`, `selected_fg`, `selected_bg`, `dropdown_fg`, `dropdown_bg`, `dropdown_selected_fg`, `dropdown_selected_bg`, `accelerator_fg`
- **prompt**: `fg`, `bg`, `label_fg`, `input_fg`
- **help**: `fg`, `bg`, `title_fg`, `title_bg`, `footer_fg`
- **message**: `info_fg`, `info_bg`, `warning_fg`, `warning_bg`, `error_fg`, `error_bg`

**Color Format:**
- Named colors: `"black"`, `"red"`, `"green"`, `"blue"`, `"white"`, etc.
- Hex colors: `"#RRGGBB"` (e.g., `"#282828"`, `"#EBDBB2"`)
- Terminal default: `"default"` (uses terminal's default colors)

**Terminal Compatibility:**
- tcell automatically detects terminal color capabilities
- Gracefully degrades from true color → 256 colors → 16 colors → monochrome
- Works correctly over SSH sessions
- 2-color terminals use text attributes (inverse, bold) instead of colors

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
theme = "default"     # Active theme name

[search]
case_sensitive = true # Case-sensitive search by default

# Theme definitions (colors support: named, hex #RRGGBB, or "default")
[themes.default.editor]
fg = "default"       # Normal text foreground
bg = "default"       # Normal text background
selection_fg = "default"
selection_bg = "default"
line_numbers_fg = "default"
line_numbers_bg = "default"

[themes.default.statusbar]
fg = "default"
bg = "default"

[themes.default.menubar]
fg = "default"
bg = "default"
selected_fg = "default"
selected_bg = "default"

[themes.default.prompt]
fg = "default"
bg = "default"

[themes.default.help]
fg = "default"
bg = "default"
title_fg = "default"
title_bg = "default"
```

**Keybinding Customization (Future/Optional):**
- Custom keybindings defined in `[keybindings]` section
- No UI for editing keybindings - config file only
- Users edit config.toml manually to customize shortcuts
- Invalid or conflicting bindings fall back to defaults with warning

**Theme System (Implemented):**
- 11 built-in themes (hardcoded, no external dependencies required):
  1. `default` - Terminal defaults with inverse video (backward compatibility)
  2. `dark` - Classic dark background with light text
  3. `light` - Classic light background with dark text
  4. `monokai` - Popular dark theme with vibrant colors
  5. `solarized-dark` - Precision dark color scheme
  6. `solarized-light` - Precision light color scheme
  7. `gruvbox-dark` - Retro groove dark colors
  8. `gruvbox-light` - Retro groove light colors
  9. `dracula` - Dark theme with purple accents
  10. `nord` - Arctic, bluish dark theme
  11. `dos` - Classic DOS Edit colors (blue background, white/cyan text)
- Custom themes can be defined in `[themes.custom_name]` sections of config file
- Built-in themes always available without config file
- View menu includes theme menu items with checkmarks showing current theme
- Theme selection automatically saved to config
- Each theme element has `fg` (foreground) and `bg` (background) colors
- Color formats: named colors (e.g., "red", "blue"), hex `#RRGGBB`, or `"default"` for terminal default
- Automatic graceful degradation for terminals with limited color support (tcell handles this automatically)
- Active theme selected via `ui.theme` config value
- Menu backgrounds fixed to be distinct from editor backgrounds for better visibility
- TODO: CLI flag `--config <path>` to override config file location (planned)

**Behavior:**
- Config file created automatically on first toggle action or theme switch
- CLI flags override config values
- Toggle actions (Ctrl+L, Ctrl+W) and theme switches automatically save config
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
