## File Tree View

### Overview

The File Tree View provides a navigable representation of the filesystem associated with the current cooledit session. It is inspired by Visual Studio Code's Explorer, adapted for terminal constraints and cooledit's single-file editing model.

The file tree is a toggleable side panel that allows browsing directories and opening files. In the initial implementation, cooledit supports only one open file at a time.

Multiple open files will be introduced later as a future feature.

---

### Scope and Constraints (Initial Implementation)

* Only one file can be open at any given time
* Opening a file via the tree replaces the currently open file
* No background buffers are kept
* No tabs or buffer switching UI
* The file tree is bolted onto the existing editor layout
* The tree remains open until explicitly closed

---

### Terminology

* File Tree View: The navigable directory and file list panel
* Editor Pane: The main text editing area
* Open File: The single file currently loaded into the editor
* OPEN EDITORS: Reserved label for future multi-file support (not functional yet)

---

### Visibility and Toggling

* The File Tree View is toggled with `Ctrl+B`
* When opened, it remains visible until toggled off
* Toggling the tree does not close the currently open file
* Editor focus shifts to the file tree when it opens
* Focus returns to the editor when the tree is closed
* `Esc` opens the menubar/menu (it does not close the File Tree View)

  * The menu system must be robust enough to open and operate correctly while the File Tree View is visible
  * Closing the menu returns focus to the File Tree View if it was the active panel before `Esc`

---

### Root Display (Top Header)

The top line of the File Tree View shows a single header:

* If cooledit was launched with a file path:

  * Display the parent directory name of that file
* If cooledit was launched with a directory path:

  * Display the directory name itself
* If cooledit was launched without a file or directory argument:

  * Display the name of the current working directory where cooledit was launched

This header is informational and not selectable.

---

### Directory and File Listing

#### Ordering Rules

* Directories are listed first
* Files are listed after directories
* Both directories and files are sorted alphabetically
* Hidden files and directories are shown (no default hiding)

#### Symbolic Links

* Symbolic links are displayed distinctly from regular files/directories
* The UI must provide a clear visual differentiation for symlinks (exact glyph/styling is implementation-defined)
* If a symlink points to a directory, it is treated as expandable (subject to normal expansion rules and filesystem permissions)

---

#### Expandable Items

* Expandable directories are prefixed with:

  ```
  > dirname
  ```

* When expanded, the indicator changes to a downward arrow equivalent:

  ```
  v dirname
  ```

* Expanded directories reveal their children immediately below them

* Nested directories follow the same rules recursively

---

### Navigation and Interaction

#### Keyboard Controls

* `Up` / `Down`

  * Move selection up and down the visible tree
* `Left`

  * Collapse an expanded directory
* `Right`

  * Expand a collapsed directory
* `Enter`

  * On directory:

    * Toggle expand/collapse
  * On file:

    * Open the file in the editor pane
* `Ctrl+B`

  * Close the File Tree View and return focus to the editor
* `Esc`

  * Opens the menubar/menu (see Visibility and Toggling)

Mouse interaction is not supported.

---

### File Opening Behavior

* Pressing `Enter` on a file:

  * Closes the currently open file (if any)
  * Loads the selected file into the editor
  * Updates the editor pane content immediately
* If the file does not exist:

  * Existing non-existent file creation rules apply (as with direct open)
* Once a file is opened from the File Tree View, the File Tree View stays open until closed with `Ctrl+B`

---

### Open File Indicator

* The currently open file is visually indicated in the tree
* The indicator uses underline styling
* Only one file can be underlined at any time

---

### Editor Pane State When No File Is Open

* If the current file is closed and no file is open:

  * The editor pane is empty
  * A distinct background or visual state may be shown to indicate no file open
* No placeholder text or message is required in the initial implementation

---

### Layout Behavior

* The File Tree View occupies a fixed-width vertical panel
* The editor pane resizes horizontally to accommodate the tree
* Minimum editor width must be preserved to avoid unusable layouts
* If terminal width is too small:

  * The File Tree View may fail to open or be truncated (implementation-defined)

---

### Persistence

* Expanded/collapsed directory state is not persisted across sessions
* Selection state is preserved while toggling the File Tree View:

  * If the user closes the File Tree View (Ctrl+B) and reopens it, the last selected item is restored
* The tree root and content reflect the current effective root directory rules (see Root Display)
* No filesystem caching across sessions

---

### Refresh and Runtime Updates

* The File Tree View must refresh to reflect directory changes caused by editor actions
* Specifically, if the opened file changes directories via Save As:

  * The File Tree View updates its root/header and listing to reflect the new directory context
  * The open file underline indicator updates accordingly
* Filesystem changes during runtime (external changes) do not require automatic live refresh unless they coincide with the above editor-driven refresh triggers

---

### Error Handling

* Unreadable directories:

  * Are displayed but cannot be expanded
* Permission errors:

  * Do not crash the editor
  * Expansion simply fails silently or with a message (implementation-defined)
* Filesystem changes during runtime:

  * No automatic refresh required beyond the explicit refresh rules in Refresh and Runtime Updates

---

### Non-Goals (Initial Implementation)

* No multiple open files
* No tabs
* No drag-and-drop
* No file creation, deletion, or renaming
* No filesystem watching or live refresh (beyond explicit refresh triggers)
* No search or filtering in the tree

---

### Future Additions

#### Multiple Open Files

* Support multiple simultaneously open files
* Internally referred to as buffers
* User-facing terminology:

  * Open files
  * OPEN EDITORS (displayed at the top of the File Tree View)
* The File Tree View will list open files separately, similar to VS Code
* The currently active file remains visually distinct

This feature is explicitly out of scope for the initial implementation.
