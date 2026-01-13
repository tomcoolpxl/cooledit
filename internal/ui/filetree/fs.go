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
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// loadChildren loads the children of a directory node
func (t *FileTree) loadChildren(node *TreeNode) {
	if !node.IsDir {
		return
	}

	entries, err := os.ReadDir(node.Path)
	if err != nil {
		node.Readable = false
		node.ChildrenLoaded = true
		node.Children = nil
		return
	}

	node.Children = nil
	for _, entry := range entries {
		childPath := filepath.Join(node.Path, entry.Name())
		child := t.createNode(childPath, entry)
		if child != nil {
			node.Children = append(node.Children, child)
		}
	}

	// Sort: directories first, then files, alphabetically within each group
	sortNodes(node.Children)
	node.ChildrenLoaded = true
	node.Readable = true
}

// createNode creates a TreeNode from a directory entry
func (t *FileTree) createNode(path string, entry os.DirEntry) *TreeNode {
	info, err := entry.Info()
	if err != nil {
		// Can't stat, create minimal node
		return &TreeNode{
			Path:     path,
			Name:     entry.Name(),
			IsDir:    entry.IsDir(),
			Readable: false,
		}
	}

	node := &TreeNode{
		Path:     path,
		Name:     entry.Name(),
		IsDir:    entry.IsDir(),
		Readable: true,
	}

	// Check for symlink
	if info.Mode()&os.ModeSymlink != 0 {
		node.IsSymlink = true
		// Check if symlink points to directory
		targetInfo, err := os.Stat(path) // follows symlink
		if err == nil && targetInfo.IsDir() {
			node.IsDir = true
		}
	}

	return node
}

// sortNodes sorts nodes: directories first, then files, alphabetically (case-insensitive)
func sortNodes(nodes []*TreeNode) {
	sort.Slice(nodes, func(i, j int) bool {
		// Directories before files
		if nodes[i].IsDir != nodes[j].IsDir {
			return nodes[i].IsDir
		}
		// Alphabetical within each group (case-insensitive)
		return strings.ToLower(nodes[i].Name) < strings.ToLower(nodes[j].Name)
	})
}
