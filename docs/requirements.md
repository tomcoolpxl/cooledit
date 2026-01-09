# Text Editor Requirements Document

## 1. Overview

Projectname: cooledit

The goal is to develop a **terminal text editor** similar in functionality and simplicity to *nano*, addressing *nano*’s UI and keyboard shortcut issues. It will be written in vanilla **Go**. The editor should be easy to use both for beginners and experienced users, with a clean status bar, optional menubar, and predictable shortcuts.

The editor will **not include syntax highlighting**. Optional features include a line numbers column and an always-visible menubar.

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

* Interactive search (forward and backward).
* Optional replace dialog.

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
* `View`: Toggle Line Numbers, Toggle Word Wrap.
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

**4.4 Search**

* `Ctrl+F` Find
* `F3` Find next
* `Shift+F3` Find previous

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

* Inline dialogs (search, replace, go to line) should appear integrated into the UI above status bar.

**5.5 Mouse Support**

* **Disabled by default**.
* Enabled via `-mouse` command line flag or config file.
* When enabled: Click to move cursor, scroll wheel support.

**5.6 Configuration Persistence**

* Settings stored in TOML format.
* Location:
  * Linux/macOS: `~/.config/cooledit/config.toml`
  * Windows: `%APPDATA%\cooledit\config.toml`
* Supported settings:
  * `editor.line_numbers` - Show line numbers column
  * `editor.soft_wrap` - Enable word wrap
  * `editor.tab_width` - Spaces per tab
  * `ui.show_menubar` - Show menubar by default
  * `ui.mouse_enabled` - Enable mouse support
  * `ui.theme` - Active theme name (default: "default")
  * `search.case_sensitive` - Case-sensitive search by default
  * `themes.*` - Theme definitions (see Section 8)
* Behavior:
  * Config file created automatically on first toggle action
  * CLI flags override config values
  * Toggle actions (Ctrl+L, Ctrl+W) automatically save config
  * Missing config file or fields use sensible defaults

**5.7 Theme System (Planned)**

* Multiple built-in themes provided
* Users can define custom themes in config file
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

---

## 6. Encoding and Newline Support

**6.1 Encoding Support**

* Detect file encoding on open.
* Allow user to re-open/convert with specific encoding.

**6.2 Newline Format**

* Detect type on load (LF or CR+LF).
* Display type in status bar.
* Preserve type on save unless user chooses to change it.

---

## 7. Configuration

**7.1 Settings File**

* Store user preferences (key mappings, line numbers, menubar visibility).
* Load on startup.
* **Keybinding customization**: Config-file-only (no UI for editing keybindings).
* Users manually edit config file to customize keyboard shortcuts.

**7.2 Default Behavior**

* Word wrap off by default.

---

## 8. Theme System (Planned - Milestone 5)

**8.1 Theme Elements**

Each element has `fg` (foreground) and `bg` (background) properties:

* **editor**: Normal text, selection, line numbers
* **search**: Search matches, current match highlight
* **statusbar**: Background, text, filename, modified indicator, position, mode indicator, mini-help
* **menubar**: Background, text, selected item, dropdown, accelerators
* **prompt**: Background, text, label, input
* **help**: Background, text, titles, footer
* **message**: Info, warning, error messages

**8.2 Built-in Themes**

* `default` - Terminal defaults with inverse video
* `dark` - Dark background theme
* `light` - Light background theme
* `monokai` - Monokai color scheme
* `solarized-dark` / `solarized-light` - Solarized themes
* `gruvbox-dark` / `gruvbox-light` - Gruvbox themes

**8.3 Custom Themes**

* Users can define themes in `[themes.name]` sections of config file
* Each theme element requires `fg` and `bg` properties
* Select active theme via `ui.theme` setting

**8.4 Color Format**

```toml
[themes.custom.editor]
fg = "#EBDBB2"      # Hex RGB
bg = "#282828"      # Hex RGB
selection_fg = "black"  # Named color
selection_bg = "white"  # Named color
line_numbers_fg = "default"  # Terminal default
line_numbers_bg = "default"
```

**8.5 Terminal Compatibility**

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

* No syntax highlighting.
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
* Mouse support (Optional via flag/config).
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