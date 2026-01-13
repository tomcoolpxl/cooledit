// Copyright (C) 2026 Tom Cool
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package filetree

import (
	"path/filepath"
)

// FileTree represents the file browser panel
type FileTree struct {
	rootPath     string        // Absolute path to root directory
	headerLabel  string        // Display name for header
	rootNode     *TreeNode     // Root node of the tree
	visibleItems []VisibleItem // Flattened visible items
	selectedPath string        // Path of selected item (for persistence)
	selectedIdx  int           // Index in visibleItems
	openFilePath string        // Path of currently open file (for underline)
	visible      bool          // Whether tree is visible
	width        int           // Panel width in characters
}

// New creates a new FileTree with the given width
func New(width int) *FileTree {
	return &FileTree{
		width:   width,
		visible: false,
	}
}

// SetRoot sets the root directory of the tree
func (t *FileTree) SetRoot(rootPath, headerLabel string) {
	absPath, err := filepath.Abs(rootPath)
	if err != nil {
		absPath = rootPath
	}
	t.rootPath = absPath
	t.headerLabel = headerLabel
	t.rootNode = &TreeNode{
		Path:     absPath,
		Name:     headerLabel,
		IsDir:    true,
		Readable: true,
		Expanded: true, // Root is always expanded
	}
	// Load root children immediately
	t.loadChildren(t.rootNode)
	t.buildVisibleItems()

	// Set initial selection if none
	if t.selectedPath == "" && len(t.visibleItems) > 0 {
		t.selectedPath = t.visibleItems[0].Node.Path
		t.selectedIdx = 0
	}
}

// SetOpenFile sets the path of the currently open file (for underline display)
func (t *FileTree) SetOpenFile(path string) {
	if path == "" {
		t.openFilePath = ""
		return
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		t.openFilePath = path
	} else {
		t.openFilePath = absPath
	}
}

// SetVisible sets the visibility of the file tree
func (t *FileTree) SetVisible(visible bool) {
	t.visible = visible
}

// IsVisible returns whether the file tree is visible
func (t *FileTree) IsVisible() bool {
	return t.visible
}

// SelectedPath returns the path of the currently selected item
func (t *FileTree) SelectedPath() string {
	return t.selectedPath
}

// Width returns the panel width
func (t *FileTree) Width() int {
	return t.width
}

// RootPath returns the root directory path
func (t *FileTree) RootPath() string {
	return t.rootPath
}

// Refresh reloads the tree from the filesystem
func (t *FileTree) Refresh() {
	if t.rootNode == nil {
		return
	}
	// Reload root children
	t.loadChildren(t.rootNode)
	t.buildVisibleItems()
	t.ensureSelectedVisible()
}

// buildVisibleItems flattens the tree into a list of visible items
func (t *FileTree) buildVisibleItems() {
	t.visibleItems = nil
	if t.rootNode == nil {
		return
	}
	// Don't include root node itself in visible items, just its children
	t.flattenNode(t.rootNode, -1)
	t.findSelectedIndex()
}

// flattenNode recursively adds nodes to visibleItems
func (t *FileTree) flattenNode(node *TreeNode, depth int) {
	// Only add non-root nodes to visible items
	if depth >= 0 {
		t.visibleItems = append(t.visibleItems, VisibleItem{
			Node:  node,
			Depth: depth,
		})
	}

	// If expanded directory, add children
	if node.IsDir && node.Expanded && node.ChildrenLoaded {
		for _, child := range node.Children {
			t.flattenNode(child, depth+1)
		}
	}
}

// findSelectedIndex finds the index of the selected path in visibleItems
func (t *FileTree) findSelectedIndex() {
	t.selectedIdx = -1
	for i, item := range t.visibleItems {
		if item.Node.Path == t.selectedPath {
			t.selectedIdx = i
			return
		}
	}
	// If not found, select first item
	if len(t.visibleItems) > 0 {
		t.selectedIdx = 0
		t.selectedPath = t.visibleItems[0].Node.Path
	}
}

// ensureSelectedVisible ensures the selected item is visible and valid
func (t *FileTree) ensureSelectedVisible() {
	if len(t.visibleItems) == 0 {
		t.selectedIdx = -1
		t.selectedPath = ""
		return
	}

	// Try to find current selection
	for i, item := range t.visibleItems {
		if item.Node.Path == t.selectedPath {
			t.selectedIdx = i
			return
		}
	}

	// Selection no longer valid, clamp to valid range
	if t.selectedIdx >= len(t.visibleItems) {
		t.selectedIdx = len(t.visibleItems) - 1
	}
	if t.selectedIdx < 0 {
		t.selectedIdx = 0
	}
	t.selectedPath = t.visibleItems[t.selectedIdx].Node.Path
}

// moveSelection moves the selection by delta items
func (t *FileTree) moveSelection(delta int) {
	if len(t.visibleItems) == 0 {
		return
	}

	newIdx := t.selectedIdx + delta
	if newIdx < 0 {
		newIdx = 0
	}
	if newIdx >= len(t.visibleItems) {
		newIdx = len(t.visibleItems) - 1
	}

	t.selectedIdx = newIdx
	t.selectedPath = t.visibleItems[newIdx].Node.Path
}

// expandSelected expands the selected directory
func (t *FileTree) expandSelected() {
	if t.selectedIdx < 0 || t.selectedIdx >= len(t.visibleItems) {
		return
	}

	node := t.visibleItems[t.selectedIdx].Node
	if !node.IsDir || !node.Readable {
		return
	}

	if !node.Expanded {
		node.Expanded = true
		if !node.ChildrenLoaded {
			t.loadChildren(node)
		}
		t.buildVisibleItems()
	}
}

// collapseSelected collapses the selected directory
func (t *FileTree) collapseSelected() {
	if t.selectedIdx < 0 || t.selectedIdx >= len(t.visibleItems) {
		return
	}

	node := t.visibleItems[t.selectedIdx].Node
	if !node.IsDir {
		return
	}

	if node.Expanded {
		node.Expanded = false
		t.buildVisibleItems()
	}
}

// toggleSelected toggles expand/collapse for directories or returns file path
func (t *FileTree) toggleSelected() ActionResult {
	if t.selectedIdx < 0 || t.selectedIdx >= len(t.visibleItems) {
		return ActionResult{Action: ActionNone}
	}

	node := t.visibleItems[t.selectedIdx].Node

	if node.IsDir {
		if node.Expanded {
			t.collapseSelected()
		} else {
			t.expandSelected()
		}
		return ActionResult{Action: ActionNone}
	}

	// It's a file - return open action
	return ActionResult{
		Action: ActionOpenFile,
		Path:   node.Path,
	}
}

// VisibleItems returns the current visible items for rendering
func (t *FileTree) VisibleItems() []VisibleItem {
	return t.visibleItems
}

// SelectedIndex returns the current selection index
func (t *FileTree) SelectedIndex() int {
	return t.selectedIdx
}

// OpenFilePath returns the path of the currently open file
func (t *FileTree) OpenFilePath() string {
	return t.openFilePath
}

// HeaderLabel returns the header label for display
func (t *FileTree) HeaderLabel() string {
	return t.headerLabel
}
