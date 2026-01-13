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

// TreeNode represents a filesystem entry in the file tree
type TreeNode struct {
	Path           string      // Absolute path (identity key)
	Name           string      // Base name for display
	IsDir          bool        // True if directory
	IsSymlink      bool        // True if symbolic link
	Readable       bool        // False if permission denied
	Expanded       bool        // True if directory is expanded
	Children       []*TreeNode // Child nodes (nil until loaded)
	ChildrenLoaded bool        // True if children have been loaded
}

// VisibleItem represents a flattened tree entry for rendering
type VisibleItem struct {
	Node  *TreeNode
	Depth int // Indentation level
}

// Action represents the result of handling a key event
type Action int

const (
	ActionNone       Action = iota // No action needed
	ActionClosePanel               // Close the file tree panel
	ActionOpenFile                 // Open a file
)

// ActionResult contains the result of handling a key event
type ActionResult struct {
	Action Action
	Path   string // File path for ActionOpenFile
}
