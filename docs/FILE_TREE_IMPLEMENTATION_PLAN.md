# File Tree View - Implementation Plan

## Overview

Implement a toggleable file browser side panel for cooledit, allowing users to navigate directories and open files without leaving the editor.

**Key binding**: `Ctrl+B` (toggle file browser)

---

## Phase 1: Rebind Bracket Matching

**Goal**: Free up `Ctrl+B` for the file browser

**New bracket matching shortcut**: `Ctrl+]`

### Files to modify:

1. `internal/ui/keymap/keymap.go` - Update binding constant
2. `internal/ui/editor.go` (or equivalent) - Update key handler
3. `internal/ui/help.go` - Update F1 help screen text
4. `internal/ui/statusbar.go` - Update mini-help if bracket matching is shown
5. `CLAUDE.md` - Update keyboard shortcuts section

### Changes:
- Remove `Ctrl+B` → bracket matching
- Add `Ctrl+]` → bracket matching
- All references to "Ctrl+B" for brackets become "Ctrl+]"

---

## Phase 2: Theme System Extension

**Goal**: Add fileview color elements to all 14 themes

### New color elements in theme structure:

```go
type FileviewColors struct {
    Fg          string  // default text color
    Bg          string  // panel background
    HeaderFg    string  // root directory header text
    HeaderBg    string  // root directory header background
    SelectionFg string  // selected item foreground
    SelectionBg string  // selected item background
    DirFg       string  // directory name color
    SymlinkFg   string  // symlink indicator color
    ExpandFg    string  // > and v expand/collapse indicators
}
```

### Files to modify:

1. `internal/theme/theme.go` - Add FileviewColors to Theme struct
2. `internal/theme/builtin.go` (or equivalent) - Add fileview colors to all 14 themes

### Theme color guidelines:

| Theme | Bg | Fg | SelectionBg | DirFg | ExpandFg |
|-------|----|----|-------------|-------|----------|
| default | terminal default | terminal default | inverse | default | default |
| dark | #1e1e1e | #d4d4d4 | #264f78 | #569cd6 | #808080 |
| light | #ffffff | #000000 | #add6ff | #0000ff | #808080 |
| monokai | #272822 | #f8f8f2 | #49483e | #66d9ef | #75715e |
| solarized-dark | #002b36 | #839496 | #073642 | #268bd2 | #586e75 |
| solarized-light | #fdf6e3 | #657b83 | #eee8d5 | #268bd2 | #93a1a1 |
| gruvbox-dark | #282828 | #ebdbb2 | #3c3836 | #83a598 | #928374 |
| gruvbox-light | #fbf1c7 | #3c3836 | #ebdbb2 | #076678 | #928374 |
| dracula | #282a36 | #f8f8f2 | #44475a | #8be9fd | #6272a4 |
| nord | #2e3440 | #d8dee9 | #3b4252 | #88c0d0 | #4c566a |
| dos | #0000aa | #ffffff | #00aaaa | #ffff55 | #aaaaaa |
| ibm-green | #000000 | #33ff33 | #005500 | #66ff66 | #009900 |
| ibm-amber | #000000 | #ffb000 | #553300 | #ffc000 | #996600 |
| cyberpunk | #0d0d0d | #0ff0fc | #1a1a2e | #ff2a6d | #05d9e8 |

---

## Phase 3: Core File Tree Component

**Goal**: Create self-contained file tree component

### New package: `internal/ui/filetree/`

#### File: `types.go`

```go
package filetree

// TreeNode represents a filesystem entry
type TreeNode struct {
    Path           string      // absolute path (identity key)
    Name           string      // base name for display
    IsDir          bool
    IsSymlink      bool
    Readable       bool        // false if permission denied
    Expanded       bool
    Children       []*TreeNode
    ChildrenLoaded bool
}

// VisibleItem represents a flattened tree entry for rendering
type VisibleItem struct {
    Node  *TreeNode
    Depth int
}

// Action represents tree component output actions
type Action int

const (
    ActionNone Action = iota
    ActionClosePanel
    ActionOpenFile
)

type ActionResult struct {
    Action Action
    Path   string // for ActionOpenFile
}
```

#### File: `tree.go`

```go
package filetree

type FileTree struct {
    rootPath     string
    headerLabel  string
    rootNode     *TreeNode
    visibleItems []VisibleItem
    selectedPath string        // path-based selection (persists across rebuilds)
    selectedIdx  int           // computed index in visibleItems
    openFilePath string        // currently open file (for underline)
    visible      bool
    width        int           // panel width
}

// Public methods:
func New(width int) *FileTree
func (t *FileTree) SetRoot(rootPath, headerLabel string)
func (t *FileTree) SetOpenFile(path string)
func (t *FileTree) SetVisible(visible bool)
func (t *FileTree) IsVisible() bool
func (t *FileTree) Refresh()
func (t *FileTree) Width() int

// Internal methods:
func (t *FileTree) buildVisibleItems()
func (t *FileTree) findSelectedIndex() int
func (t *FileTree) expandNode(node *TreeNode)
func (t *FileTree) collapseNode(node *TreeNode)
```

#### File: `fs.go`

```go
package filetree

import (
    "os"
    "path/filepath"
    "sort"
)

func (t *FileTree) loadChildren(node *TreeNode) error
func (t *FileTree) createNode(path string, info os.FileInfo) *TreeNode
func sortNodes(nodes []*TreeNode) // dirs first, then files, alphabetical
```

**Sorting rules**:
1. Directories before files
2. Alphabetical within each group (case-insensitive recommended)
3. Hidden files/dirs included (no filtering)

**Symlink handling**:
- Use `os.Lstat()` to detect symlinks
- If symlink points to directory, allow expansion
- Mark with `IsSymlink = true` for display

#### File: `input.go`

```go
package filetree

import "github.com/gdamore/tcell/v2"

func (t *FileTree) HandleKey(ev *tcell.EventKey) ActionResult

// Key bindings when tree is focused:
// - Up/Down: move selection
// - Left: collapse if expanded directory
// - Right: expand if collapsed directory
// - Enter: open file OR toggle directory expand/collapse
// - Ctrl+B: return ActionClosePanel
// - (Esc is NOT handled here - goes to menu system)
```

#### File: `render.go`

```go
package filetree

import "cooledit/internal/term"

func (t *FileTree) Render(screen term.Screen, x, y, height int, theme Theme)

// Rendering details:
// - Line 0: Header (root directory name, non-selectable, HeaderFg/HeaderBg)
// - Lines 1+: Visible items
// - Indentation: depth * 2 spaces
// - Prefix: "> " (collapsed dir), "v " (expanded dir), "  " (file)
// - Symlink suffix: " @"
// - Selection: SelectionFg/SelectionBg on selected item
// - Open file: underline style
// - Combine selection + underline when both apply
```

---

## Phase 4: App Integration

**Goal**: Wire file tree into the application

### Focus model extension

Add to focus enum (likely in `internal/ui/` or `internal/app/`):

```go
const (
    FocusEditor FocusTarget = iota
    FocusMenu
    FocusFileTree  // NEW
    // ... other focus targets
)
```

### Files to modify:

1. `internal/app/app.go` - Add FileTree instance, toggle logic
2. `internal/ui/editor.go` - Layout adjustment when tree visible
3. `internal/ui/keymap/keymap.go` - Add Ctrl+B binding for file tree toggle

### Integration logic:

```go
// In app/editor key handling:
case Ctrl+B:
    if fileTree.IsVisible() {
        fileTree.SetVisible(false)
        focus = FocusEditor
    } else {
        fileTree.SetVisible(true)
        focus = FocusFileTree
    }

// When tree is focused and returns ActionOpenFile:
case ActionOpenFile:
    if buffer.IsModified() {
        // prompt to save
    }
    closeCurrentFile()
    openFile(action.Path)
    fileTree.SetOpenFile(action.Path)

// When tree returns ActionClosePanel:
case ActionClosePanel:
    fileTree.SetVisible(false)
    focus = FocusEditor
```

### Layout changes:

```go
func calculateLayout(termWidth, termHeight int, treeVisible bool, treeWidth int) {
    if treeVisible {
        treeRect = Rect{0, 0, treeWidth, termHeight - statusBarHeight}
        editorRect = Rect{treeWidth, 0, termWidth - treeWidth, termHeight - statusBarHeight}
    } else {
        editorRect = Rect{0, 0, termWidth, termHeight - statusBarHeight}
    }
    // Enforce minimum editor width (e.g., 40 columns)
}
```

### Root context initialization:

```go
// At startup, determine file tree root:
func determineTreeRoot(args []string) (rootPath, headerLabel string) {
    if len(args) > 0 {
        path := args[0]
        info, err := os.Stat(path)
        if err == nil && info.IsDir() {
            // Launched with directory
            return filepath.Abs(path), filepath.Base(path)
        } else {
            // Launched with file
            dir := filepath.Dir(path)
            return filepath.Abs(dir), filepath.Base(dir)
        }
    }
    // No args - use CWD
    cwd, _ := os.Getwd()
    return cwd, filepath.Base(cwd)
}
```

### Save As integration:

```go
// After Save As completes with new path:
func onSaveAs(oldPath, newPath string) {
    newDir := filepath.Dir(newPath)
    oldDir := filepath.Dir(oldPath)

    if newDir != oldDir {
        // Directory changed - update tree root
        fileTree.SetRoot(newDir, filepath.Base(newDir))
    }
    fileTree.SetOpenFile(newPath)
    fileTree.Refresh()
}
```

### Menu interaction:

- Esc when tree focused → opens menu (handled by app, not tree)
- Menu close → return focus to tree if it was focused before Esc
- Menu must render correctly over/beside tree panel

---

## Phase 5: UI Updates

**Goal**: Update all UI elements to show Ctrl+B for file browser

### 1. F1 Help Screen

Add to Navigation or new "File Browser" section:
```
Ctrl+B     Toggle file browser
```

Update bracket matching:
```
Ctrl+]     Jump to matching bracket
```

### 2. Status Bar Mini-Help

Add to the rotation of help items:
```go
miniHelpItems := []string{
    "F1 Help",
    "Esc/F10 Menu",
    "Ctrl+Q Quit",
    "Ctrl+S Save",
    "Ctrl+F Find/Replace",
    "Ctrl+B Browser",  // NEW
}
```

### 3. View Menu

Add menu item:
```
View
├── Toggle Line Numbers
├── ...
├── ─────────────────
├── File Browser        Ctrl+B   // NEW - with checkmark when visible
├── ─────────────────
├── ...
```

### 4. CLAUDE.md Updates

- Add File Tree View to "Implemented" features
- Update keyboard shortcuts section
- Add fileview to theme color elements
- Document file tree behavior

---

## Phase 6: Testing

**Goal**: Comprehensive test coverage

### New test file: `internal/ui/filetree/filetree_test.go`

#### Sorting tests:
- Directories listed before files
- Alphabetical order within groups
- Hidden files included
- Case handling consistency

#### Expansion tests:
- Expanding loads children once (lazy)
- Collapsing hides children
- Nested expansion works
- Re-expanding preserves previous children or reloads

#### Selection tests:
- Up/Down navigation clamps at bounds
- Selection by path persists across toggle off/on
- Selection stable after refresh when item exists
- Selection fallback when item removed

#### Open file indicator tests:
- Underline applied only to matching path
- Updating open file updates underline
- No underline when no file open

#### Symlink tests:
- Symlinks detected via Lstat
- Display marker applied
- Symlink-to-dir is expandable

#### Integration tests (may require fake screen):
- Focus transitions: editor → tree → menu → tree → editor
- Ctrl+B toggles visibility
- Enter on file triggers open action
- Save As updates root when directory changes

---

## Implementation Order

1. **Phase 1**: Rebind bracket matching (Ctrl+B → Ctrl+]) - ~30 min
2. **Phase 2**: Theme extension (add fileview colors) - ~1 hour
3. **Phase 3**: Core component (types, tree, fs, input, render) - ~3-4 hours
4. **Phase 4**: App integration (focus, layout, handlers) - ~2-3 hours
5. **Phase 5**: UI updates (help, menu, statusbar) - ~1 hour
6. **Phase 6**: Testing - ~2 hours

---

## Open Questions / Decisions

1. **Panel width**: Fixed 30 chars? Configurable? Percentage of terminal?
   - Recommendation: Fixed 30 chars initially, can add config later

2. **Symlink display**: `name @` or `name -> target`?
   - Recommendation: `name @` (shorter, cleaner)

3. **Unreadable directories**: Show as non-expandable or show with error on expand?
   - Recommendation: Show with `> ` but expansion fails silently

4. **Case sensitivity in sorting**: Match OS behavior or always case-insensitive?
   - Recommendation: Case-insensitive for consistency

5. **Tree width in narrow terminals**: Fail to open or truncate?
   - Recommendation: Don't open if terminal < 60 cols wide

---

## Files Summary

### New files:
- `internal/ui/filetree/types.go`
- `internal/ui/filetree/tree.go`
- `internal/ui/filetree/fs.go`
- `internal/ui/filetree/input.go`
- `internal/ui/filetree/render.go`
- `internal/ui/filetree/filetree_test.go`

### Modified files:
- `internal/theme/theme.go` - add FileviewColors
- `internal/theme/builtin.go` - add colors to 14 themes
- `internal/ui/keymap/keymap.go` - Ctrl+B → file tree, Ctrl+] → brackets
- `internal/ui/editor.go` - layout, key handling
- `internal/ui/help.go` - F1 screen updates
- `internal/ui/statusbar.go` - mini-help update
- `internal/ui/menu.go` - add View → File Browser item
- `internal/app/app.go` - file tree instance, integration
- `CLAUDE.md` - documentation
