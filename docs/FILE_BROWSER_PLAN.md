```md
# File Tree View Implementation Notes (No External Tree Library)

## Purpose

This document proposes a minimal, testable internal design for implementing the File Tree View in cooledit without adopting a third-party TUI widget library.

Goals:
- Fit into cooledit as it exists today (tcell-based, custom UI components, keyboard-driven).
- Avoid assumptions about internal architecture that are not known from the current conversation.
- Keep the design small enough to implement quickly and evolve later into "OPEN EDITORS".

Non-goals:
- Do not specify exact package names, file paths, or existing interfaces unless they are confirmed.
- Do not design multi-buffer support now. Only reserve extension points.

This is a design and implementation guidance document, not a requirements spec (the requirements live elsewhere).

## Why no third-party tree widget library

A tree widget library typically imposes at least one of these:
- It owns the input event loop or focus model.
- It assumes its own layout system and redraw orchestration.
- It has its own styling primitives instead of your existing underline/style system.
- It assumes mouse support or different key semantics.
- It makes "virtual sections" (future OPEN EDITORS) awkward.

In cooledit, the complexity is integration (focus, Esc menu interaction, underline open file, selection persistence, refresh rules), not the tree algorithm itself.

## Constraints to respect (from the requirements)

- Single open file at a time (for now).
- File tree toggled by Ctrl+B; stays open until Ctrl+B closes it.
- Esc opens the menu, not closing the tree. Menu must work while tree is visible.
- Root header shows:
  - parent directory name if launched with file,
  - directory name if launched with dir,
  - current working directory name if launched without args.
- Expandable items show "> " when collapsed and "v " when expanded.
- Enter opens files, toggles directories. Arrow keys navigate.
- Directories listed before files, both alphabetically.
- Hidden files shown.
- Symlinks shown distinctly.
- Selected item preserved when toggling tree off and on.
- Underline indicates the currently open file.
- Refresh must occur when Save As changes directory context.

## High-level approach

Implement the File Tree View as a self-contained component with:
- Its own state (root, expanded nodes, visible flattened list, selection).
- A small set of methods:
  - HandleKey(event) -> (consumed bool, actions)
  - Render(screen, bounds, styles)
  - SetVisible(bool)
  - Refresh(trigger)
  - SetOpenFile(path or none)
  - SetRoot(context)
- Zero external dependencies beyond Go standard library filesystem utilities.

The tree should use:
- Lazy directory loading: only read children when a node is expanded (or first displayed).
- A flattened "visible list" derived from the tree state for navigation and rendering.

## Data model (minimal)

The tree needs two representations:
- A structural model for expansion state.
- A flattened view for drawing and selection.

### Node identity

Use absolute path as the stable identity key.

Rationale:
- It is stable across redraws.
- It is unique within a session.
- It is easy to map open file path and expansion state.

Normalization:
- Use filepath.Clean and filepath.Abs when storing identity paths.
- Preserve original display name separately if needed.

### Node metadata

Store only what you need to render and decide behavior.

Suggested structs (names illustrative, you can rename):

type TreeNode struct {
    Path        string   // absolute identity
    Name        string   // base name for display
    IsDir       bool
    IsSymlink   bool
    LinkTarget  string   // optional, for display or dir detection (if resolved)
    Readable    bool     // false if permission or stat failed
    Expanded    bool
    Children    []*TreeNode
    ChildrenLoaded bool
}

Notes:
- ChildrenLoaded enables lazy loading.
- Readable indicates whether expand should be allowed.
- LinkTarget is optional, but you need a way to distinguish symlinks visually.
- If you want to treat symlink-to-dir as expandable, you must decide how to detect that:
  - Either by using os.Stat (follows symlink) in addition to os.Lstat (does not follow),
  - Or by attempting to read directory and seeing if it succeeds.
  Both approaches are valid; choose based on how you want to handle broken links.

### Flattened view items

Flattened items should carry:
- Pointer to node
- Depth for indentation
- Render prefix info
- A stable index for selection

type VisibleItem struct {
    Node  *TreeNode
    Depth int
    // Optional: computed flags like IsHeader, but header should be separate line
}

### Expansion state map (optional)

If you do not want to store Expanded in each node, keep a map:

expanded: map[string]bool  // keyed by absolute path

However, storing Expanded in TreeNode is simpler if you keep nodes around. If you rebuild nodes often, a map is safer.

Given the "selection persists when toggling", you likely will keep state in the component instance anyway.

## Root selection and header

Separate:
- Root directory path (absolute)
- Header label (display name)

The header label is not selectable and not part of the flattened list.

Root directory choice rules:
- Launched with file path: root = parent directory of that file.
- Launched with directory path: root = that directory.
- Launched without args: root = current working directory at launch time.

This implies you need the initial launch context passed into the tree component once at startup.

Save As refresh rule:
- If open file moved to a different directory context, the tree must update root and header accordingly.

This requires a call from the Save As workflow into the tree component, for example:
- tree.OnFilePathChanged(oldPath, newPath)
or
- tree.SetOpenFile(newPath) plus tree.SetRootFromOpenFile(newPath)
depending on where the logic belongs.

Do not bury Save As detection inside the tree. The file tree should be told when the open file path changes.

## Rendering rules

### Layout

The file tree is a vertical panel with fixed width. The editor pane uses the remaining width. This requires the layout manager to decide bounds and call tree.Render with the rectangle.

This document does not assume how layout is implemented. The file tree should accept bounds and not compute layout globally.

### Indentation and prefixes

Rendering for each visible item:

- indentation: Depth * indentWidth
  - indentWidth can be 2 spaces for terminal friendliness
- prefix:
  - if expandable and collapsed: "> "
  - if expandable and expanded: "v "
  - if not expandable: "  " (align with others)

Do not attempt fancy box drawing. Keep consistent spacing.

### Symlink display

Requirement: symlinks displayed distinctly.

Minimal terminal-safe options:
- Add a suffix like " @" or " -> target" (but target can be long).
- Change color/style if theming supports it.
- Add a prefix marker like "~ " (but conflicts with expand indicator).
- Append a small ASCII marker like "[L]" or "@".

Given you already use underline for open file, prefer a non-underline signal for symlink.

One robust option:
- Append " @" to the name for symlinks
- If LinkTarget is available and short enough, optionally " -> name"

This document does not mandate the exact glyph. It mandates distinctness.

### Open file underline

If item.Node.Path equals the current open file absolute path, render with underline style.

If no file open, underline none.

### Selection

Selection is independent from underline. Underline indicates open file. Selection indicates focus cursor in the tree.

If a selected item is also the open file, both attributes should be visible. If your style system cannot combine selection background with underline cleanly, you need a precedence rule. Suggested precedence:
- Apply selection background/foreground
- Preserve underline if possible
If not possible, selection wins while selected, underline visible when not selected.

This should be explicitly tested because it is easy to regress.

### Unreadable directories

If a directory is not readable:
- Still list it (Readable false)
- Show it as non-expandable (no "> " prefix), or show expandable but expansion fails
Pick one behavior and keep it consistent.

The requirements allow "displayed but cannot be expanded". The simplest is treat as non-expandable.

## Input handling and focus

The tree component should not decide global focus rules. It should:
- Expose a HandleKey method that consumes relevant keys.
- Not intercept Esc, because Esc is reserved for the menu.
- Only operate when it is the active focus target.

Keys it should handle when focused:
- Up, Down: move selection
- Left: collapse selected dir if expanded, otherwise move selection to parent (optional, see note)
- Right: expand selected dir if collapsible, otherwise no-op
- Enter: open file or toggle expand dir
- Ctrl+B: request close panel (it can return an action to the app)

Note on Left behavior:
- Requirements say Left collapses expanded directory. It does not say it must navigate to parent.
- Do not add "go to parent" unless you confirm that is desired.

Menu interaction:
- When Esc opens menu, focus moves to menu.
- When menu closes, focus should return to file tree if it was focused before Esc.
This is global focus management, not tree internals. The tree should not special-case Esc.

## Flattened view generation

Flatten visible list by DFS over expanded nodes:

BuildVisibleItems(node, depth):
- append (node, depth)
- if node.IsDir and node.Expanded:
  - ensure children loaded
  - for each child in sorted order:
    - recurse with depth+1

Sorting:
- child directories first, then files
- alphabetical within each group
- define case sensitivity rule (recommend stable and predictable; if you already sort elsewhere, reuse that behavior)
This document does not assume whether cooledit uses case-sensitive sorting today. Decide explicitly.

Selection stability:
- When you rebuild visible list, try to keep selection on the same path if it still exists.
- If it no longer exists, clamp selection to nearest valid index.

Selection persistence when toggling:
- Store SelectedPath string, not just SelectedIndex.
- On reopen, rebuild visible list and compute index of SelectedPath. If missing, fallback to first item.

This solves cases where the list changes due to refresh and avoids selecting the wrong item.

## Lazy loading and refresh strategy

### Lazy loading

When expanding a directory:
- If ChildrenLoaded false:
  - read directory entries
  - stat/lstat to determine isDir and isSymlink
  - build children nodes
  - sort children
  - set ChildrenLoaded true

If reading fails:
- mark node.Readable false
- set ChildrenLoaded false and Expanded false, or keep Expanded false
- optionally surface an error message via a returned action
This document does not define messaging mechanisms.

### Refresh triggers

Avoid filesystem watchers for now.

Perform refresh on explicit triggers only:
- When the file tree is opened (optional; depends on how fresh you want it)
- When Save As causes directory context change (required)
- When user expands a directory (always reads live contents then)

What refresh means:
- Recompute root if needed (on Save As)
- Rebuild visible list
- Preserve expansion state where possible

Preserving expansion across root changes:
- If root changes, you may want to reset expansion state entirely because paths changed context.
- Alternatively, keep expansion state for nodes under the new root if they match.
This is a design choice. The requirements do not specify.
Given simplicity, resetting expansions on root change is acceptable, but selection persistence requirement still applies. If you reset expansions, selection may become invalid, so your fallback logic matters.

## File open action flow (single-file mode)

When Enter on a file:
- The tree component should not directly perform file I/O.
- It should return an action: OpenFile(path)
- The application layer executes:
  - close current file (if any, including prompting for unsaved changes if that exists today)
  - open new file
  - update open file path in tree: tree.SetOpenFile(path)

This separation matters for testing and avoids hidden dependencies.

This document does not assume how unsaved changes prompts are done. It only states the tree should not bypass them.

## Integration points you will need (minimal)

To integrate cleanly, you likely need these cross-component signals:

From app/editor to tree:
- SetVisible(bool) triggered by Ctrl+B
- SetOpenFile(path or empty) whenever open file changes or closes
- SetRootFromLaunchContext(...) once
- NotifySaveAs(oldPath, newPath) or NotifyFilePathChanged

From tree to app/editor:
- RequestClosePanel (Ctrl+B when tree focused)
- RequestOpenFile(path)
- Optional: RequestShowMessage(error) on permission failures

This document does not assume your existing message bar API. Use whatever you already have.

## Future extension points (OPEN EDITORS)

Do not implement now, but do not block it.

Recommended extension strategy:
- Allow the visible list to include non-filesystem entries (section headers and open file entries).
- This implies VisibleItem should be able to represent:
  - filesystem node
  - section header
  - open editor entry

However, since you do not want to add this complexity yet, keep the current VisibleItem simple and leave a comment or TODO to evolve it.

The one important rule now:
- Keep selection and open-file tracking path-based. This will also work when multiple open files exist.

## Testing strategy

You should be able to test the file tree component without a real terminal screen.

If you already have a fake screen for UI tests, reuse it. If not, keep rendering logic pure enough to validate lines output.

Core tests (logic, not rendering):
- Sorting:
  - directories before files
  - alphabetical order within each group
  - includes hidden files
- Expansion:
  - expanding loads children once
  - collapsing hides children
  - nested expansion works
- Selection:
  - Up/Down clamp at bounds
  - selection preserved when toggling panel off/on (via SelectedPath)
  - selection stable across refresh when item still exists
- Open file underline mapping:
  - underline applied only to open file path
  - updating open file updates underline target
- Save As refresh:
  - root changes when open file changes directory (as per requirements)
  - selection and underline update accordingly
- Symlink distinctness:
  - symlink nodes are detected and marked
  - display marker decision is applied consistently

Rendering tests (if feasible):
- Prefix changes "> " vs "v "
- Indentation increments with depth
- Underline style on open file line
- Selection style applied to selected line

Do not over-test filesystem edge cases using real filesystem unless you already have patterns for temporary directories. Prefer temp dirs if you do.

## Pitfalls and how to avoid them

### Selection by index only
If you store selection only as an index, refresh or expansion can shift indices and selection will jump to the wrong item. Store SelectedPath.

### Rebuilding nodes too aggressively
If you rebuild the whole tree on every key press, performance will degrade in large directories and tests become flaky. Only rebuild visible list when needed.

### Symlink detection
Use os.Lstat to detect symlink. If you also need to detect symlink-to-dir, decide how to resolve safely without infinite loops:
- Avoid recursive traversal through symlinks automatically.
- Only expand when user requests and treat failures gracefully.

### Combining underline and selection styles
If your style system cannot combine attributes, you need an explicit precedence rule and tests.

### Menu and focus interactions
Esc must open menu and not close the tree. Ensure Ctrl+B toggling works whether focus is in editor or tree. Focus restoration after menu closes must be deterministic.

## Information needed from the codebase (do not guess)

To turn this document into an implementation plan that matches your project exactly, I would need to see:

- The current UI component architecture:
  - how components are rendered and how input is routed
- The focus model:
  - how editor vs menu vs overlays are selected
- The keymap layer:
  - where Ctrl+B is currently bound (it is currently bracket match per your earlier text, so this is a conflict that must be resolved)
- The "close current file" behavior:
  - how unsaved changes prompts are handled
- Save As workflow:
  - where the final path change is committed and how other components are notified
- Theme/style API:
  - how underline and selection are represented and combined

If you share the relevant files (or even just the key types and methods), this document can be tightened into concrete interfaces and exact method names without inventing anything.

## Review checklist for this document

- Does it avoid assuming package names, file names, or specific existing interfaces? Yes.
- Does it preserve all requirements already defined (Ctrl+B toggle, Esc menu, root rules including CWD, selection persistence, symlinks distinct, hidden files shown, Save As refresh)? Yes.
- Does it keep single-file mode and reserve future OPEN EDITORS without implementing it? Yes.
- Does it define a minimal state model that is testable and fits terminal constraints? Yes.
- Does it highlight a major likely conflict: Ctrl+B currently used for bracket matching in your earlier shortcut list? Yes, this is an integration risk to resolve explicitly.

```

Important integration warning (not a summary): your existing shortcut list already uses `Ctrl+B` for bracket matching/jump. That is a direct conflict with the file tree toggle requirement. You need to introduce a new binding for one of them, otherwise the implementation cannot be consistent.

EXTRA NOTES:
Ctrl+B should be file browser, brackets matching should be something else shortcut. Fileview colors (fg, bg, selection fg/bg, expand character fg, ...) should be included in theme and existing themes should get consistent colors. Menu, statusbar mini help section and F1 help screen should display Ctrl+B