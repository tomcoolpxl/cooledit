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

* Cut/Cut line, copy, paste.
* Undo/Redo.

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
* Word wrap **off by default**.

**3.2 Status Bar (persistent at bottom)**
Must display:

* Current **filename** or `[No Name]`.
* **Modified** flag if unsaved changes exist.
* **Cursor position**: line and column.
* **Encoding** of file.
* **EOL type** (LF / CR+LF).
* Mode or messages (search prompt, dialogs).

The status bar concept is similar to *nano’s* status and messages region. ([gnu.ist.utl.pt][2])

**3.3 Optional Line Numbers Column**

* Toggleable by user (persisted in settings).
* Fixed width depending on file size.

**3.4 Optional Menubar**
A top menubar that can be always on or hidden.
Menus to include (with shortcuts accessible from anywhere):

* `File`: New, Open, Save, Save As…, Quit/Exit.
* `Edit`: Undo, Redo, Cut, Copy, Paste, Delete Line.
* `Search`: Find, Find Next/Prev, Replace.
* `View`: Toggle line numbers, Toggle menubar, Toggle word wrap.
  Keyboard access via Alt key + letter (e.g., `Alt+F` for File). Menus should work with a keyboard-only UI.

---

## 4. Keyboard Shortcuts

Design shortcuts that **improve on nano’s** cumbersome conventions: eliminate multi-step Meta sequences where possible and follow familiar patterns (similar to Notepad where practical), but remain effective in a terminal.

**4.1 File Operations**

* `Ctrl+O` Save
* `Ctrl+S` Save As
* `Ctrl+Q` Quit/Exit

**4.2 Navigation**

* `Arrow keys` Move cursor
* `Ctrl+Home` / `Ctrl+End` Beginning / End of file
* `Ctrl+G` Go to line dialog

**4.3 Edit**

* `Ctrl+X` Cut line
* `Ctrl+C` Copy line
* `Ctrl+V` Paste
* `Ctrl+Z` Undo
* `Ctrl+Y` Redo

**4.4 Search**

* `Ctrl+F` Find
* `Ctrl+R` Replace
* `F3` Find next
* `Shift+F3` Find previous

**4.5 Display**

* `Ctrl+L` Toggle line numbers
* `Ctrl+W` Toggle word wrap
* `Ctrl+M` Toggle menubar

These are suggested keys adapted from and improving upon common *nano* commands. ([nano-editor.org][3])

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
* If enabled during session, wrap visually without altering file content.

**5.4 Dialogs**

* Inline dialogs (search, replace, go to line) should appear integrated into the UI above status bar.

**5.5 Mouse Support (Optional)**

* Click to move cursor.
* Click menubar entries.

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

**7.2 Default Behavior**

* Word wrap off by default.
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

## 12. Release Milestones (Draft)

**Milestone 1**

* Basic buffer editing, file load/save, status bar.

**Milestone 2**

* Search/replace, encoding display, newline type.

**Milestone 3**

* Optional line numbers, menubar, dialogs.

**Milestone 4**

* Keyboard customization, settings persistence.
