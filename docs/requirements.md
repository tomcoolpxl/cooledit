# Text Editor Requirements Document

## 1. Overview

Projectname: cooledit

The goal is to develop a **terminal text editor** similar in functionality and simplicity to *nano*, addressing *nano*’s UI and keyboard shortcut issues. It will be written in vanilla **Go**. The editor should be easy to use both for beginners and experienced users, with a clean status bar, optional menubar, and predictable shortcuts.

Optional features include a line numbers column, an always-visible menubar, and syntax highlighting.

---

## 2. Core Use Cases

**2.1 Open, View, and Edit a Text File**

* Open files by passing the filename as an argument.
* Edit text immediately, no modal states.
* Save and exit using common shortcuts.

**2.2 Efficient Navigation**

* Standard cursor movement (left, right, up, down).
* Page-up / page-down, beginning/end of line.
* Go-to-line functionality.

**2.3 Search and Replace**

* Unified search mode with real-time results (incremental search)
* Type to search - matches appear immediately as you type
* Navigate through matches with keyboard shortcuts
* Case-sensitive and whole word search options (toggleable)
* Search history navigation with up/down arrows
* Pre-fill search from current selection
* Replace current match or replace all with confirmation
* Visual highlighting of all matches in viewport
* Match count display showing current position (e.g., "Match 3 of 15")
* All operations work while typing (no need to "commit" search)
* Persistent search preferences across searches

**2.4 Text Manipulation**

* Cut, copy, paste (selection or current line).
* System Clipboard integration.
* Undo/Redo.
* Text Selection via Shift+Arrow keys.

**2.5 Encoding Support**

* Load and save files with different encodings (e.g., UTF-8, ISO-8859-1).
* Detect and display encoding in status bar.

**2.6 Newline Format Awareness**

* Recognize and display newline types (LF vs CR+LF).
* Preserve original newline format on save.

---

## 3. UI Structure

**3.1 Main Editing Pane**

* Displays file contents.
* Cursor visibly indicated.
* Selection highlighting.
* Word wrap **off by default**.

**3.2 Status Bar (persistent at bottom)**
Must display:

* Current **filename** or `[No Name]`.
* **Modified** flag if unsaved changes exist.
* **Cursor position**: line and column.
* **Encoding** of file.
* **EOL type** (LF / CR+LF).
* Mode or messages (search prompt, dialogs).

**3.3 Optional Line Numbers Column**

* Toggleable by user (persisted in settings).
* Fixed width depending on file size.

**3.4 Optional Menubar**
A top menubar that is **auto-hidden by default**.
Toggled via `F10` or `Esc` (in Normal mode).
Menus to include:

* `File`: Save, Save As…, Quit.
* `Edit`: Undo, Redo, Cut, Copy, Paste, Go to Line.
* `Search`: Find, Find Next/Prev.
* `View`: Toggle Line Numbers, Toggle Word Wrap, Themes (submenu).
* `Help`: About.

---

## 4. Keyboard Shortcuts

**4.1 File Operations**

* `Ctrl+S` Save
* `Ctrl+Shift+S` Save As
* `Ctrl+Q` Quit/Exit

**4.2 Navigation & Selection**

* `Arrow keys` Move cursor
* `Shift+Arrow keys` Select text
* `Ctrl+G` Go to Line
* `Ctrl+Home` / `Ctrl+End` Beginning / End of file
* `PageUp` / `PageDown` Scroll pages

**4.3 Edit (Clipboard)**

* `Ctrl+X` / `Shift+Del`: Cut selection (or current line if empty)
* `Ctrl+C`: Copy selection (or current line if empty)
* `Ctrl+V` / `Shift+Ins`: Paste from system clipboard
* `Ctrl+Z` Undo
* `Ctrl+Y` Redo

**4.4 Search (Unified Mode)**

* `Ctrl+F` Enter unified search mode (incremental search with real-time results)
* `Alt+C` Toggle case sensitivity (in search mode)
* `Alt+W` Toggle whole word matching (in search mode)
* `N` / `P` Navigate to next/previous match (in search mode)
* `F3` / `Shift+F3` Find next/previous (works in and out of search mode)
* `Up` / `Down` Navigate search history (in search mode)
* `Enter` Move to next match (same as N in search mode)
* `R` Replace current match (in search mode, prompts for replacement text)
* `A` Replace all matches (in search mode, shows confirmation dialog)
* `Esc` / `Q` Exit search mode
* `Backspace` Delete character from query (or exit search if query is empty)
* Any character: Add to search query (immediate real-time search)

**4.5 Edit Mode**

* `Insert` Toggle Insert/Replace mode

**4.6 Display**

* `F1` Help overlay
* `F10` Toggle Menubar focus

---

## 5. Behavior

**5.1 Saving**

* Prompt if file already exists (with preview of filename).
* If the buffer has no filename, prompt for Save As.

**5.2 Untitled Buffer**

* When there’s no file name, status bar should show `[No Name]`.
* Exit should prompt to save.

**5.3 Word Wrap**

* Off by default.

**5.4 Dialogs**

* Inline dialogs (go to line) appear integrated into the UI above status bar.
* Search uses a unified search mode with real-time feedback in the status bar.
* Replace prompts appear as inline dialogs.
* Replace-all shows a confirmation dialog with match count before proceeding.

**5.5 Configuration Persistence**

* Settings stored in TOML format.
* Location:
  * Linux/macOS: `~/.config/cooledit/config.toml`
  * Windows: `%APPDATA%\cooledit\config.toml`
* Supported settings:
  * `editor.line_numbers` - Show line numbers column
  * `editor.soft_wrap` - Enable word wrap
  * `editor.tab_width` - Spaces per tab
  * `ui.show_menubar` - Show menubar by default
  * `ui.theme` - Active theme name (default: "default")
  * `search.case_sensitive` - Case-sensitive search preference (persists across editor sessions)
  * `search.whole_word` - Whole word search preference (persists across editor sessions)
  * `themes.*` - Theme definitions (see Section 9)
* Behavior:
  * Config file created automatically on first toggle action
  * CLI flags override config values
  * Toggle actions (Ctrl+L, Ctrl+W) automatically save config
  * Missing config file or fields use sensible defaults

**5.7 Theme System (Planned)**

* 10 built-in themes (hardcoded, no external dependencies):
  1. default, 2. dark, 3. light, 4. monokai, 5. solarized-dark,
  6. solarized-light, 7. gruvbox-dark, 8. gruvbox-light, 9. dracula, 10. nord
* Custom themes can be defined in config file
* Built-in themes always available without config file
* UI menu support: View → Themes submenu to switch themes
* Theme selection automatically saved to config
* Each UI element has configurable foreground and background colors
* Color formats supported:
  * Named colors (e.g., "red", "blue")
  * Hex colors (e.g., "#282828", "#EBDBB2")
  * "default" for terminal's default colors
* Automatic graceful degradation:
  * True color (16M colors) on modern terminals
  * 256-color fallback
  * 16-color (ANSI) fallback
  * Monochrome fallback using text attributes (inverse, bold)
* Works correctly over SSH and in limited terminals
* CLI flag: `--config <path>` to override config file location

---

## 6. Command-Line Interface

**6.1 Command-Line Flags**

* `cooledit [OPTIONS] [filename]` - Open file for editing
* `-l, --line-numbers` - Show line numbers column
* `-c, --config <path>` - Use alternate config file location
* `-v, --version` - Show version information
* `-h, --help` - Show help message

**6.2 Environment Variables**

* `TERM` - Terminal type (automatically detected)
* Standard XDG and Windows environment variables for config location

---

## 7. Encoding and Newline Support

**6.1 Encoding Support**

* Detect file encoding on open.
* Allow user to re-open/convert with specific encoding.

**7.2 Newline Format**

* Detect type on load (LF or CR+LF).
* Display type in status bar.
* Preserve type on save unless user chooses to change it.

---

## 8. Configuration

**8.1 Settings File**

* Store user preferences (line numbers, menubar visibility, theme selection).
* Load on startup.
* **Keybinding customization**: Config-file-only (no UI) - Future/Optional feature.
* Users can manually edit config file to customize keyboard shortcuts (when implemented).

**7.2 Default Behavior**

* Word wrap off by default.

---

## 9. Theme System (Planned - Milestone 4 - Next Phase)

**9.1 Theme Elements**

Each element has `fg` (foreground) and `bg` (background) properties:

* **editor**: Normal text, selection, line numbers
* **search**: Search matches, current match highlight
* **statusbar**: Background, text, filename, modified indicator, position, mode indicator, mini-help
* **menubar**: Background, text, selected item, dropdown, accelerators
* **prompt**: Background, text, label, input
* **help**: Background, text, titles, footer
* **message**: Info, warning, error messages

**9.2 Built-in Themes**

10 hardcoded themes (no external dependencies required):

1. `default` - Terminal defaults with inverse video (current behavior)
2. `dark` - Classic dark background with light text
3. `light` - Classic light background with dark text  
4. `monokai` - Popular dark theme with vibrant purple, pink, yellow, green
5. `solarized-dark` - Ethan Schoonover's precision dark color scheme
6. `solarized-light` - Ethan Schoonover's precision light color scheme
7. `gruvbox-dark` - Retro groove colors, warm dark background
8. `gruvbox-light` - Retro groove colors, warm light background
9. `dracula` - Dark purple/pink theme, easy on the eyes
10. `nord` - Arctic bluish theme inspired by northern lights

**Theme Selection:**
* View → Themes menu shows all available themes
* Current theme indicated with checkmark
* Selection automatically saved to config
* Keyboard shortcut: `Ctrl+T` or customizable

**8.3 Custom Themes**

* Users can define themes in `[themes.name]` sections of config file
* Each theme element requires `fg` and `bg` properties
* Select active theme via `ui.theme` setting

**9.4 Color Format**

```toml
[themes.custom.editor]
fg = "#EBDBB2"      # Hex RGB
bg = "#282828"      # Hex RGB
selection_fg = "black"  # Named color
selection_bg = "white"  # Named color
line_numbers_fg = "default"  # Terminal default
line_numbers_bg = "default"
```

**9.5 Terminal Compatibility**

* Automatic detection of terminal color capabilities via tcell
* Graceful degradation: true color → 256 color → 16 color → monochrome
* Monochrome terminals use text attributes (inverse, bold, underline)
* Works over SSH with proper TERM environment variable
* Line numbers off by default.

---

## 8. Error and Prompt Handling

**8.1 Status Messages**

* Show messages (e.g., “File saved”, “Search string not found”).
* Errors in status bar until acknowledged.

**8.2 Blocking Prompts**

* Use modal dialog overlays for critical confirmations (overwrite, unsaved changes).

---

## 9. Extensibility and Future Features (Notes)

Not in initial scope but candidates:

* Plugin support (future).
* Macro recording.
* Multiple file buffers (later).

---

## 10. Non-Goals

* No tabbed interface.
* No markdown rendering.

---

## 11. Platform & Technical

**11.1 Use Golang**

* Efficient text buffer implementation.

**11.2 Performance**

* Should handle large files without UI lag.

---

## 12. Release Milestones

**Milestone 1 (Complete)**

* Basic buffer editing, file load/save, status bar.
* Tcell backend.

**Milestone 2 (Complete)**

* Search (Find/Next/Prev).
* Undo/Redo.
* UI Prompts (Save As, Quit, Find).
* Layout Engine.

**Milestone 3 (Complete)**

* Menubar (Auto-hide).
* Text Selection and System Clipboard.
* Configuration persistence with TOML.
* Toggle settings auto-save.
* Soft wrap rendering with adaptive line wrapping.
* Insert/Replace mode with Insert key and cursor shapes.

**Milestone 4 (Planned)**

* Keybinding customization.



Saving

Ctrl+S

If file has a path → Save (overwrite)

If modified → write file

If not modified → no-op + brief message

If file has no path → Save As

Ctrl+Shift+S

Always Save As

Overwrite confirmation:

Only when overwriting an existing file that is not the current file

Same rule as VS Code:

Normal Save never asks

Save As asks only if path exists and is different

Help

F1

Opens a simple full-screen help overlay

Any key exits help

Does not modify editor state