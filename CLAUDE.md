# Claude Context - cooledit Project

## Project Overview

**cooledit** is a terminal-based text editor written in vanilla Go, inspired by nano but with better UI and keyboard shortcuts. The goal is to create a simple, beginner-friendly editor with optional syntax highlighting.

## Core Design Principles

- **Syntax Highlighting** - Optional, enabled by default, uses Chroma library
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
  autosave/            - Autosave and recovery system
  config/              - Configuration management
  core/                - Editor core logic
    buffer/            - Text buffer implementation
    text/              - Text processing utilities
  fileio/              - File operations, encoding, EOL handling
  positionlog/         - Cursor position persistence across sessions
  syntax/              - Syntax highlighting with Chroma
  term/                - Terminal backend abstraction
    tcell/             - Tcell implementation
  theme/               - Theme system and color management
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
- ✅ Go to Line (Ctrl+G) - Always available
- ✅ Adaptive help screen (F1) with two-column layout for wide terminals
- ✅ Configuration persistence with TOML file
- ✅ Toggle settings auto-save (line numbers, soft wrap)
- ✅ Soft wrap implementation with adaptive line wrapping
- ✅ Insert/Replace mode toggle with Insert key and cursor shape indicators
- ✅ Tab handling with smart indentation (Tab inserts spaces, Ctrl+I inserts literal \t)

### Implemented (Milestone 4 - Theme System)
- ✅ Term.Style extended with Foreground/Background color fields
- ✅ Theme package with comprehensive color element definitions
- ✅ 14 built-in themes: default, dark, light, monokai, solarized-dark/light, gruvbox-dark/light, dracula, nord, dos, ibm-green, ibm-amber, cyberpunk
- ✅ Custom theme support from config file
- ✅ All UI elements use theme colors (editor, menubar, status bar, prompt, help)
- ✅ View menu with theme selection (checkmarks for current theme)
- ✅ Interactive theme switching with auto-save to config
- ✅ Graceful terminal capability detection via tcell (true color → 256 → 16 → monochrome)
- ✅ Backward compatibility with "default" theme using inverse video
- ✅ Menu backgrounds fixed to be distinct from editor backgrounds in all themes

### Implemented (Milestone 5 - Syntax Highlighting)
- ✅ Chroma-based syntax highlighting with 50+ supported languages
- ✅ Language auto-detection via file extension and shebang
- ✅ Separate Language menu (no longer in View menu)
- ✅ Language menu structure: Off/Auto/separator/languages (only Off/Auto stored in config)
- ✅ Viewport-based highlighting (only tokenizes visible lines for performance)
- ✅ Line-based token caching with hash invalidation
- ✅ Theme-integrated syntax colors (all 14 themes have syntax color schemes)
- ✅ Toggle syntax highlighting with Ctrl+H or View menu
- ✅ Status bar shows current language
- ✅ Syntax highlighting enabled by default

### Implemented (Additional Features)
- ✅ Word navigation with Ctrl+Left/Right arrow keys
- ✅ Bracket matching and jumping with Ctrl+B
- ✅ Non-existent file creation (allows opening files that don't exist yet)
- ✅ Whitespace visualization toggle (Ctrl+Shift+W) - displays spaces (·), tabs (→), and line endings (↵/¶)
- ✅ Smart tab rendering (single arrow per tab character, not per expanded space)
- ✅ Smart Home key (cycles between first non-whitespace and column 0)
- ✅ Block indent/unindent (Tab/Shift+Tab with selection indents/unindents lines)
- ✅ Trim trailing whitespace on save (configurable via View menu)
- ✅ Comment/Uncomment toggle (Ctrl+/) - language-aware line commenting
- ✅ Position Log - remembers cursor position in recently edited files, restores on reopen
- ✅ Scrollbar/Indicator - visual scrollbar on right edge showing viewport position
- ✅ Verbatim Character Input - insert Unicode characters by code point (Ctrl+Shift+U for hex, Ctrl+Shift+D for decimal)

### Implemented (Milestone 6 - Autosave)
- ✅ Automatic backup after idle timeout (default: 2 seconds)
- ✅ Minimum interval between autosaves (default: 30 seconds)
- ✅ Cross-platform autosave directory (Windows: %APPDATA%, Linux: ~/.local/share, macOS: ~/Library/Application Support)
- ✅ Recovery prompt on startup when autosave exists: [R]ecover, [O]pen original, [D]iscard
- ✅ Autosave cleared on explicit save (Ctrl+S)
- ✅ Autosave kept when quitting without saving (for future recovery)
- ✅ View menu toggle for enabling/disabling autosave
- ✅ Configurable via config file (enabled, idle_timeout, min_interval)

### Implemented (Additional UI Features)
- ✅ Current line highlight (toggle via View menu, off by default)
- ✅ Per-theme hardcoded CurrentLineBg colors for all 14 themes
- ✅ Highlight spans full line width including gutter/line numbers

### Planned (Remaining)
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
- `Ctrl+Left` / `Ctrl+Right` - Jump by word
- `Ctrl+B` - Jump to matching bracket
- `Ctrl+L` - Toggle line numbers
- `Ctrl+W` - Toggle word wrap
- `Ctrl+Shift+W` - Toggle whitespace display
- `Ctrl+H` - Toggle syntax highlighting
- `Ctrl+/` - Toggle comment on line or selection
- `F11` - Toggle status bar (Zen mode)
- `Home` - Smart home (first non-whitespace, then column 0)
- `Tab` (with selection) - Indent selected lines
- `Shift+Tab` - Unindent selected lines (or current line)
- `Tab` - Insert spaces to next tab stop (configurable width, default: 4)
- `Ctrl+I` - Insert literal tab character (\t)
- `Ctrl+Shift+U` - Verbatim Unicode hex input (e.g., type "2665" for ♥)
- `Ctrl+Shift+D` - Verbatim Unicode decimal input (e.g., type "9829" for ♥)
- `Insert` - Toggle Insert/Replace mode
- `F1` - Help overlay (adaptive two-column/single-column layout)
- `F10` / `Esc` - Toggle menubar

## Key Design Decisions

### Status Bar (Zen Mode Support)
**Left** (priority 2): Filename with modified indicator (`*`)
**Center** (priority 3): Mini-help with adaptive display:
  - `F1 Help` → `Esc/F10 Menu` → `Ctrl+Q Quit` → `Ctrl+S Save` → `Ctrl+F Find/Replace`
  - Shows as many items as fit, removes from right to left when space is limited
**Right** (priority 1): `REPLACE  Ln X, Col Y  Encoding EOL` (replace mode indicator when active, cursor position and file status)

**Zen Mode**: Press `F11` to toggle status bar visibility for distraction-free editing. The status bar automatically reappears when needed during:
- Prompt/message dialogs
- Find/Replace mode
- Go to Line
- Save As
- Overwrite confirmation

### Menubar (Auto-hidden by Default)
- Toggle with F10 or Esc
- Menus: File, Edit, Search, View, Language, Themes, Help
- **DOS-style keyboard shortcuts**: Press underlined letter to activate menu item
  - File: **S**ave, Save **A**s, **Q**uit
  - Edit: **U**ndo, **R**edo, Cu**t**, **C**opy, **P**aste, Select All (**G**rab All)
  - Search: **F**ind/Replace, Find **N**ext, Find **P**rev
  - Help: **K**eyboard Shortcuts
  - View: No shortcuts (toggles use arrow keys/Enter)
- **Automatic scrolling**: On small screens, menus scroll automatically with ↑/↓ indicators
- View menu includes:
  - Toggle Line Numbers (checkmark when enabled)
  - Current Line Highlight toggle (checkmark when enabled)
  - Toggle Word Wrap (checkmark when enabled)
  - Show Whitespace (checkmark when enabled)
  - Toggle Status Bar (checkmark when enabled)
  - Syntax Highlighting toggle (checkmark when enabled)
  - Autosave toggle (checkmark when enabled)
  - Trim Trailing Whitespace on Save (checkmark when enabled)
  - Remember Cursor Position (checkmark when enabled)
  - Show Scrollbar (checkmark when enabled)
  - **Separator line**
  - Cursor Blink toggle (checkmark when enabled)
  - Cursor shapes (block, underline, bar with checkmark for active shape)
  - **Separator line**
  - EOL Format display (readonly: LF or CRLF)
  - Encoding display (readonly: UTF-8, etc.)
  - **Separator line (in Themes menu)**
  - Themes submenu (all 14 available themes with checkmark for active theme)
- Navigation via arrow keys
- Menu items support: checkmarks (for toggles only), separators (visual lines), and readonly items (informational display, no checkmarks)
- **Smart navigation**: Up/Down arrows automatically skip separator lines and readonly items

### Command-Line Flags
- `-l, --line-numbers` - Show line numbers column
- `-c, --config <path>` - Use alternate config file location
- `-v, --version` - Show version information
- `-h, --help` - Show help message

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
- **Dynamic column width**: Calculates optimal column width to prevent overlap
- **Smart truncation**: Shows "(scroll down for more)" when content doesn't fit
- **Organized sections**: Menu & Help, File, Edit, Search, Navigation
- **Clean design**: Simple indentation, highlighted section headers, no box-drawing characters

### Word Wrap
- **Off by default**
- Can be toggled with Ctrl+W
- Adaptive wrapping to viewport width
- Line numbers shown only on first wrapped segment

### Insert/Replace Mode
- **Insert mode by default** (uses configured cursor shape)
- Toggle with Insert key to replace/overwrite mode (uses alternate cursor shape)
- Replace mode shows **"REPLACE"** indicator in status bar
- Replace mode overwrites characters instead of inserting
- At end of line, behaves like insert mode
- State not saved - always starts in insert mode

### Cursor Shapes
- **Configurable cursor shapes**: Choose from block, underline, or bar (vertical line)
- **Default**: Block cursor for insert mode
- **Smart alternation**: Replace mode automatically uses an alternate shape:
  - If insert cursor is **block** → replace cursor is **underline**
  - If insert cursor is **underline** → replace cursor is **block**
  - If insert cursor is **bar** → replace cursor is **block**
- **Persistent**: Cursor shape preference saved to config file
- **Menu access**: View → Cursor submenu (block, underline, bar with checkmarks)
- **All three shapes supported**: Block (█), Underline (_), Bar (|)

### Cursor Colors (Theme-based)
- **Themed cursor colors**: Each theme defines a cursor color that contrasts well with the background
- **Default theme**: Green cursor (#00FF00) - classic terminal aesthetic
- **IBM themes**: Authentic phosphor colors (green for ibm-green, amber for ibm-amber)
- **Color support**: Terminal support varies - modern terminals (Windows Terminal, iTerm2, Alacritty, Kitty) generally support it
- **Configuration**: Cursor colors can be customized in theme config using hex (#RRGGBB) or named colors
- **Terminal state preservation**: Original terminal cursor color is saved on startup and restored on exit

### Secret Features
- **Vim command mode**: Press `:` while in menu mode (F10/Esc) to enter vim command mode
  - Supported commands: `:w` (save), `:w!` (force save), `:q` (quit), `:q!` (force quit), `:wq` (save and quit)
  - Command shown in status bar while typing
  - Press Enter to execute, Esc to cancel, Backspace to delete characters
  - Not documented in F1 help screen by design (easter egg for vim users)

### Tab Handling
- **Tab key inserts spaces** (not literal `\t` characters)
- **Configurable tab width** (default: 4 spaces, set via `tab_width` in config)
- **Smart indentation**: Tab moves cursor to next tab stop (e.g., column 4, 8, 12...)
- **Simple backspace**: Backspace always deletes one character at a time (space, tab, or any character) - no smart deletion
- **Literal tabs**: Press `Ctrl+I` to insert a raw `\t` character
- **Rendering**: Literal tab characters render with proper width via tcell
- **Undo/Redo**: Tab insertion is an atomic operation

### Enter Key Behavior
- **Enter creates new line**: Splits current line at cursor position
- **Auto-indent**: Copies leading whitespace (spaces and tabs) from current line to new line (nano-style)
- **Smart behavior**: If current line starts with "    code", new line also starts with "    "
- **Rendering**: Literal tab characters render with proper width via tcell
- **Undo/Redo**: Tab insertion and smart backspace are atomic operations

### Syntax Highlighting
- **Enabled by default**: Can be toggled with Ctrl+H or View menu
- **Chroma-based**: Uses the Chroma library for lexer support
- **50+ languages supported**: Programming, scripting, config files, markup
- **Auto-detection**: Language detected from file extension first, then shebang
- **Manual override**: View → Language menu allows forcing specific language
- **Viewport-based**: Only visible lines are tokenized for performance
- **Line caching**: Tokens cached per line with FNV hash-based invalidation
- **Theme integration**: Each theme defines syntax colors for 13 token types
- **Token types**: Keyword, String, Comment, Number, Operator, Function, Type, Variable, Constant, Punctuation, Preproc, Builtin
- **Status bar**: Shows current language (e.g., "Go", "Python", "Auto")

**Supported Languages Include:**
- **Programming**: Go, Python, JavaScript, TypeScript, Java, C, C++, C#, Rust, Ruby, PHP, Swift, Kotlin, Scala, Perl, Lua, R, Haskell, Erlang, Elixir, Clojure, OCaml, F#, Dart, Julia, Zig, Nim, Crystal, V
- **Web**: HTML, CSS, SCSS, SASS, Less
- **Shell/Sysadmin**: Bash, PowerShell, Batch, Fish, Zsh
- **Config**: YAML, JSON, TOML, INI, XML, Nginx, Apache, Properties, Registry
- **Data**: SQL, GraphQL, Protobuf
- **Build**: Makefile, CMake, Gradle, Dockerfile
- **Markup**: Markdown, reStructuredText, LaTeX, Diff
- **Cloud/DevOps**: Terraform, HCL

### Autosave System
- **Enabled by default**: Can be toggled via View → Autosave menu
- **Idle-based trigger**: Autosave occurs after 2 seconds of no edits (configurable)
- **Minimum interval**: At least 30 seconds between autosaves (configurable)
- **Storage location**:
  - Windows: `%APPDATA%\cooledit\autosave\`
  - Linux: `~/.local/share/cooledit/autosave/`
  - macOS: `~/Library/Application Support/cooledit/autosave/`
- **File naming**: Uses FNV-1a hash of original path for safe cross-platform filenames
- **Metadata**: Each autosave has a `.meta` file with original path, encoding, EOL, timestamp
- **Recovery prompt**: On startup, if autosave exists for target file:
  - `[R]ecover backup` - Load autosave content, mark as modified
  - `[O]pen original` - Load original file, keep autosave for later
  - `[D]iscard backup` - Delete autosave, load original file
- **Lifecycle rules**:
  - Created: After idle timeout when buffer is modified
  - Deleted: On explicit save (Ctrl+S) or clean quit
  - Kept: On quit without saving (for future recovery)
- **Unnamed buffers**: Not supported (autosave requires a file path)

## Testing Strategy

- Unit tests for core components (buffer, editor, search, undo)
- UI tests with fake screen implementation
- Coverage tracking in place
- **170+ tests covering**:
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
  - Menu navigation and rendering (including menu scrolling on small screens)
  - Syntax highlighting (20 tests: token types, language detection, caching, Chroma integration)
  - Autosave system (28 tests: storage, manager, recovery, lifecycle)

### Technical Details

**Style System**:
- All text rendering uses `term.Style` struct with foreground, background, and **underline** attributes
- Underline support added for menu shortcut keys (DOS-style visual cue)
- tcell backend implements underline via `.Underline(true)` method

**Menu System Implementation**:
- `MenuItem` struct includes `ShortcutKey rune` field for keyboard shortcuts
- `Menubar` struct includes `ScrollOffset int` for automatic menu scrolling
- Shortcut keys work case-insensitively (both lowercase and uppercase)
- Scroll offset resets to 0 when switching between menus
- `adjustMenuScroll()` ensures selected item is always visible on screen
- Visual indicators (↑/↓) show when menu content extends beyond visible area

**Terminal Backend**:
- `SetCursorShape()` accepts both shape (block/underline/bar) and color parameters
- Original terminal cursor style saved during initialization, restored on exit
- tcell v2 provides cursor style control via `SetCursorStyle(tcell.CursorStyleXX, color)`

### EOL Format
- Auto-detect (LF vs CRLF)
- Display in status bar
- Preserve original format on save

## Theme System (Implemented - Milestone 4)

**Built-in Themes:**
14 hardcoded themes that work out of the box without any configuration:
1. `default` - Uses terminal defaults with inverse video, green cursor (backward compatibility)
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
12. `ibm-green` - Classic IBM green phosphor monitor (black background, green shades)
13. `ibm-amber` - Classic IBM amber phosphor monitor (black background, amber/orange shades)
14. `cyberpunk` - Neon colors on dark background with pink/cyan/yellow accents

**Custom Themes:**
Users can define additional themes in config file using `[themes.custom_name]` sections.

**UI Support:**
- View → Themes menu shows all available themes (built-in + custom)
- Click to switch theme (saves selection to config)
- Current theme indicated with checkmark

**Color Elements:**
Each element has `fg` (foreground) and `bg` (background) properties.

- **editor**: `fg`, `bg`, `selection_fg`, `selection_bg`, `line_numbers_fg`, `line_numbers_bg`, `cursor_color`
- **search**: `match_fg`, `match_bg`, `current_match_fg`, `current_match_bg`
- **statusbar**: `fg`, `bg`, `filename_fg`, `modified_fg`, `position_fg`, `mode_fg`, `help_fg`
- **menubar**: `fg`, `bg`, `selected_fg`, `selected_bg`, `dropdown_fg`, `dropdown_bg`, `dropdown_selected_fg`, `dropdown_selected_bg`, `accelerator_fg`
- **prompt**: `fg`, `bg`, `label_fg`, `input_fg`
- **help**: `fg`, `bg`, `title_fg`, `title_bg`, `footer_fg`
- **message**: `info_fg`, `info_bg`, `warning_fg`, `warning_bg`, `error_fg`, `error_bg`
- **syntax**: `keyword_fg/bg`, `string_fg/bg`, `comment_fg/bg`, `number_fg/bg`, `operator_fg/bg`, `function_fg/bg`, `type_fg/bg`, `variable_fg/bg`, `constant_fg/bg`, `preproc_fg/bg`, `builtin_fg/bg`, `punctuation_fg/bg`

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
line_numbers = false                    # Show line numbers column
soft_wrap = false                       # Enable word wrap
tab_width = 4                           # Spaces per tab
syntax_highlighting = true              # Enable syntax highlighting (default: true)
show_whitespace = false                 # Show whitespace characters (default: false)
current_line_highlight = false          # Highlight current line (default: false)
trim_trailing_whitespace = false        # Trim trailing whitespace on save (default: false)
remember_position = true                # Remember cursor position across sessions (default: true)
show_scrollbar = false                  # Show scrollbar indicator (default: false)

[ui]
show_menubar = false     # Show menubar by default
show_statusbar = true    # Show status bar (F11 toggles Zen mode)
theme = "default"        # Active theme name
cursor_shape = "block"   # Cursor shape: "block", "underline", or "bar"
cursor_blink = true      # Enable cursor blinking
language = ""            # Manual language override (empty = auto-detect)

[search]
case_sensitive = true # Case-sensitive search by default

[autosave]
enabled = true        # Enable autosave (default: true)
idle_timeout = 2      # Seconds of idle before autosave (default: 2)
min_interval = 30     # Minimum seconds between autosaves (default: 30)

# Theme definitions (colors support: named, hex #RRGGBB, or "default")
[themes.default.editor]
fg = "default"       # Normal text foreground
bg = "default"       # Normal text background
selection_fg = "default"
selection_bg = "default"
line_numbers_fg = "default"
line_numbers_bg = "default"
cursor_color = "default"  # Cursor color (terminal support varies)

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
- 14 built-in themes (hardcoded, no external dependencies required):
  1. `default` - Terminal defaults with inverse video, green cursor (backward compatibility)
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
  12. `ibm-green` - Classic IBM green phosphor monitor (black background, green shades)
  13. `ibm-amber` - Classic IBM amber phosphor monitor (black background, amber/orange shades)
  14. `cyberpunk` - Neon colors with pink/cyan/yellow accents
- Custom themes can be defined in `[themes.custom_name]` sections of config file
- Built-in themes always available without config file
- View menu includes theme menu items with checkmarks showing current theme
- Theme selection automatically saved to config
- Each theme element has `fg` (foreground) and `bg` (background) colors
- Color formats: named colors (e.g., "red", "blue"), hex `#RRGGBB`, or `"default"` for terminal default
- Automatic graceful degradation for terminals with limited color support (tcell handles this automatically)
- Active theme selected via `ui.theme` config value
- Menu backgrounds fixed to be distinct from editor backgrounds for better visibility (6 themes updated)

**Behavior:**
- Config file created automatically on first toggle action or theme switch
- CLI flags override config values
- Toggle actions (Ctrl+L, Ctrl+W) and theme switches automatically save config
- Missing config file or fields use sensible defaults

## Non-Goals

- ❌ Tabbed interface
- ❌ Multiple simultaneous file buffers
- ❌ Markdown rendering/preview
- ❌ Plugin system (not initial scope)
- ❌ Modal editing (vim-like modes)

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
- ✅ Adaptive help screen with two-column layout and dynamic column widths
- ✅ Message bar persistence during find/replace operations
- ✅ Configuration system with TOML persistence
- ✅ Toggle settings auto-save to config file
- ✅ Soft wrap rendering with proper line wrapping and cursor positioning
- ✅ Insert/Replace mode with Insert key toggle and smart cursor shape alternation
- ✅ Configurable cursor shapes (block, underline, bar) with theme-based colors
- ✅ 14 built-in themes including retro IBM phosphor themes
- ✅ DOS-style menu shortcuts with underlined letters
- ✅ Automatic menu scrolling for small screens
- ✅ Secret vim command mode (`:w`, `:q`, `:wq`, etc.)
- ✅ Terminal cursor state preservation
- ✅ Syntax highlighting with Chroma library (50+ languages)
- ✅ Language auto-detection and manual selection
- ✅ Theme-integrated syntax colors
- ✅ Autosave with idle-based trigger and recovery prompt
- ✅ Cross-platform autosave storage with metadata
- ✅ Comment/Uncomment with Ctrl+/ (language-aware)
- ✅ Position log for cursor position persistence across sessions
- ✅ Scrollbar indicator for viewport position
- ✅ Verbatim Unicode character input (Ctrl+Shift+U hex, Ctrl+Shift+D decimal)
- ✅ Comprehensive test coverage (170+ tests, all passing)

Focus areas:
- Additional features as requested

## When Working on This Project

1. **Follow existing patterns** - Buffer management, command pattern for undo/redo
2. **Test thoroughly** - Especially buffer operations and UI interactions
3. **Maintain simplicity** - This is meant to be a simple, nano-like editor
4. **Cross-platform** - Consider Windows, Linux, and macOS compatibility
5. **Terminal constraints** - Remember this runs in a terminal, not a GUI
6. **Syntax highlighting uses Chroma** - Language support via internal/syntax package

---

Last Updated: January 2025
