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
	"testing"
)

// TestNewFileTree tests FileTree creation
func TestNewFileTree(t *testing.T) {
	ft := New(30)
	if ft == nil {
		t.Fatal("New returned nil")
	}
	if ft.Width() != 30 {
		t.Errorf("Expected width 30, got %d", ft.Width())
	}
	if ft.IsVisible() {
		t.Error("Expected new tree to be invisible")
	}
}

// TestSetRoot tests setting the root directory
func TestSetRoot(t *testing.T) {
	ft := New(30)

	// Create temp directory for testing
	tmpDir := t.TempDir()

	// Create some test files and directories
	os.MkdirAll(filepath.Join(tmpDir, "dir1"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "dir2"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "file2.txt"), []byte("test"), 0644)

	ft.SetRoot(tmpDir, "TestRoot")

	if ft.RootPath() != tmpDir {
		t.Errorf("Expected root path %s, got %s", tmpDir, ft.RootPath())
	}

	// Should have visible items (2 dirs + 2 files)
	if len(ft.visibleItems) != 4 {
		t.Errorf("Expected 4 visible items, got %d", len(ft.visibleItems))
	}
}

// TestSortingDirsFirst tests that directories are sorted before files
func TestSortingDirsFirst(t *testing.T) {
	ft := New(30)

	tmpDir := t.TempDir()

	// Create files and dirs - files alphabetically before dirs
	os.WriteFile(filepath.Join(tmpDir, "aaa.txt"), []byte("test"), 0644)
	os.MkdirAll(filepath.Join(tmpDir, "zzz"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "bbb.txt"), []byte("test"), 0644)

	ft.SetRoot(tmpDir, "Test")

	// First item should be the directory
	if len(ft.visibleItems) < 3 {
		t.Fatalf("Expected at least 3 items, got %d", len(ft.visibleItems))
	}

	if !ft.visibleItems[0].Node.IsDir {
		t.Error("Expected first item to be a directory")
	}
	if ft.visibleItems[0].Node.Name != "zzz" {
		t.Errorf("Expected first item to be 'zzz', got %s", ft.visibleItems[0].Node.Name)
	}
}

// TestSortingAlphabetical tests alphabetical sorting within dirs/files
func TestSortingAlphabetical(t *testing.T) {
	ft := New(30)

	tmpDir := t.TempDir()

	// Create directories in non-alphabetical order
	os.MkdirAll(filepath.Join(tmpDir, "charlie"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "alpha"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "bravo"), 0755)

	ft.SetRoot(tmpDir, "Test")

	// Check alphabetical order
	expected := []string{"alpha", "bravo", "charlie"}
	for i, name := range expected {
		if ft.visibleItems[i].Node.Name != name {
			t.Errorf("Expected item %d to be %s, got %s", i, name, ft.visibleItems[i].Node.Name)
		}
	}
}

// TestSortingCaseInsensitive tests case-insensitive alphabetical sorting
func TestSortingCaseInsensitive(t *testing.T) {
	ft := New(30)

	tmpDir := t.TempDir()

	// Create files with mixed case
	os.WriteFile(filepath.Join(tmpDir, "Zebra.txt"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "apple.txt"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "Banana.txt"), []byte("test"), 0644)

	ft.SetRoot(tmpDir, "Test")

	// Check case-insensitive alphabetical order
	expected := []string{"apple.txt", "Banana.txt", "Zebra.txt"}
	for i, name := range expected {
		if ft.visibleItems[i].Node.Name != name {
			t.Errorf("Expected item %d to be %s, got %s", i, name, ft.visibleItems[i].Node.Name)
		}
	}
}

// TestHiddenFilesIncluded tests that hidden files are included
func TestHiddenFilesIncluded(t *testing.T) {
	ft := New(30)

	tmpDir := t.TempDir()

	// Create hidden and visible files
	os.WriteFile(filepath.Join(tmpDir, ".hidden"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "visible.txt"), []byte("test"), 0644)

	ft.SetRoot(tmpDir, "Test")

	// Should have 2 items (hidden files included)
	if len(ft.visibleItems) != 2 {
		t.Errorf("Expected 2 visible items (including hidden), got %d", len(ft.visibleItems))
	}

	// Check hidden file is included
	hasHidden := false
	for _, item := range ft.visibleItems {
		if item.Node.Name == ".hidden" {
			hasHidden = true
			break
		}
	}
	if !hasHidden {
		t.Error("Expected hidden file to be included")
	}
}

// TestExpandCollapse tests expanding and collapsing directories
func TestExpandCollapse(t *testing.T) {
	ft := New(30)

	tmpDir := t.TempDir()

	// Create nested directory structure
	os.MkdirAll(filepath.Join(tmpDir, "parent", "child"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "parent", "file.txt"), []byte("test"), 0644)

	ft.SetRoot(tmpDir, "Test")

	// Initially, parent is collapsed, so only parent visible
	initialCount := len(ft.visibleItems)
	if initialCount != 1 {
		t.Errorf("Expected 1 item (collapsed parent), got %d", initialCount)
	}

	// Expand parent
	ft.selectedIdx = 0
	ft.selectedPath = ft.visibleItems[0].Node.Path
	ft.expandSelected()

	// Should now see parent's children (child dir + file.txt)
	if len(ft.visibleItems) != 3 {
		t.Errorf("Expected 3 items after expand (parent + child + file), got %d", len(ft.visibleItems))
	}

	// Collapse parent
	ft.selectedIdx = 0
	ft.selectedPath = ft.visibleItems[0].Node.Path
	ft.collapseSelected()

	// Should be back to just parent
	if len(ft.visibleItems) != 1 {
		t.Errorf("Expected 1 item after collapse, got %d", len(ft.visibleItems))
	}
}

// TestLazyLoading tests that children are loaded lazily
func TestLazyLoading(t *testing.T) {
	ft := New(30)

	tmpDir := t.TempDir()

	// Create nested structure
	os.MkdirAll(filepath.Join(tmpDir, "parent", "child"), 0755)

	ft.SetRoot(tmpDir, "Test")

	// Root children are loaded
	if ft.rootNode.ChildrenLoaded != true {
		t.Error("Expected root children to be loaded")
	}

	// But nested children should not be loaded until expanded
	parentNode := ft.visibleItems[0].Node
	if parentNode.ChildrenLoaded {
		t.Error("Expected nested children to NOT be loaded before expansion")
	}

	// Expand parent
	ft.selectedIdx = 0
	ft.expandSelected()

	// Now children should be loaded
	if !parentNode.ChildrenLoaded {
		t.Error("Expected children to be loaded after expansion")
	}
}

// TestSelectionPersistence tests that selection is preserved by path
func TestSelectionPersistence(t *testing.T) {
	ft := New(30)

	tmpDir := t.TempDir()

	// Create files
	os.WriteFile(filepath.Join(tmpDir, "aaa.txt"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "bbb.txt"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "ccc.txt"), []byte("test"), 0644)

	ft.SetRoot(tmpDir, "Test")

	// Select bbb.txt
	ft.moveSelection(1) // Move to second item (bbb.txt)
	selectedPath := ft.SelectedPath()

	if !containsName(selectedPath, "bbb.txt") {
		t.Errorf("Expected selection to be bbb.txt, got %s", selectedPath)
	}

	// Refresh tree
	ft.Refresh()

	// Selection should still be bbb.txt
	if ft.SelectedPath() != selectedPath {
		t.Errorf("Expected selection to be preserved as %s, got %s", selectedPath, ft.SelectedPath())
	}
}

// TestSelectionBoundsClamping tests that selection is clamped to valid range
func TestSelectionBoundsClamping(t *testing.T) {
	ft := New(30)

	tmpDir := t.TempDir()

	// Create files
	os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "file2.txt"), []byte("test"), 0644)

	ft.SetRoot(tmpDir, "Test")

	// Try to move selection beyond end
	ft.moveSelection(100)

	// Should be clamped to last item
	if ft.selectedIdx != len(ft.visibleItems)-1 {
		t.Errorf("Expected selection clamped to %d, got %d", len(ft.visibleItems)-1, ft.selectedIdx)
	}

	// Try to move before beginning
	ft.moveSelection(-100)

	// Should be clamped to 0
	if ft.selectedIdx != 0 {
		t.Errorf("Expected selection clamped to 0, got %d", ft.selectedIdx)
	}
}

// TestOpenFilePath tests setting and checking open file path
func TestOpenFilePath(t *testing.T) {
	ft := New(30)

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("test"), 0644)

	ft.SetRoot(tmpDir, "Test")
	ft.SetOpenFile(testFile)

	// The open file path should match the test file
	if ft.openFilePath != testFile {
		t.Errorf("Expected open file path %s, got %s", testFile, ft.openFilePath)
	}
}

// TestVisibility tests toggle visibility
func TestVisibility(t *testing.T) {
	ft := New(30)

	if ft.IsVisible() {
		t.Error("Expected new tree to be invisible")
	}

	ft.SetVisible(true)
	if !ft.IsVisible() {
		t.Error("Expected tree to be visible after SetVisible(true)")
	}

	ft.SetVisible(false)
	if ft.IsVisible() {
		t.Error("Expected tree to be invisible after SetVisible(false)")
	}
}

// TestToggleDirectory tests toggling a directory via toggleSelected
func TestToggleDirectory(t *testing.T) {
	ft := New(30)

	tmpDir := t.TempDir()

	// Create nested structure
	os.MkdirAll(filepath.Join(tmpDir, "dir"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "dir", "file.txt"), []byte("test"), 0644)

	ft.SetRoot(tmpDir, "Test")

	// Select directory
	ft.selectedIdx = 0
	ft.selectedPath = ft.visibleItems[0].Node.Path

	initialCount := len(ft.visibleItems)

	// Toggle (expand)
	result := ft.toggleSelected()

	// Should return ActionNone for directory toggle
	if result.Action != ActionNone {
		t.Errorf("Expected ActionNone for directory toggle, got %d", result.Action)
	}

	// Should now have more items
	if len(ft.visibleItems) <= initialCount {
		t.Error("Expected more items after expanding directory")
	}
}

// TestToggleFile tests toggling a file via toggleSelected
func TestToggleFile(t *testing.T) {
	ft := New(30)

	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("test"), 0644)

	ft.SetRoot(tmpDir, "Test")

	// Select file
	ft.selectedIdx = 0
	ft.selectedPath = ft.visibleItems[0].Node.Path

	// Toggle should return ActionOpenFile
	result := ft.toggleSelected()

	if result.Action != ActionOpenFile {
		t.Errorf("Expected ActionOpenFile for file toggle, got %d", result.Action)
	}

	if result.Path != testFile {
		t.Errorf("Expected path %s, got %s", testFile, result.Path)
	}
}

// TestMoveSelection tests moveSelection with various deltas
func TestMoveSelection(t *testing.T) {
	ft := New(30)

	tmpDir := t.TempDir()

	// Create multiple files
	for i := 0; i < 5; i++ {
		os.WriteFile(filepath.Join(tmpDir, string('a'+byte(i))+".txt"), []byte("test"), 0644)
	}

	ft.SetRoot(tmpDir, "Test")

	// Start at 0
	if ft.selectedIdx != 0 {
		t.Errorf("Expected initial selection at 0, got %d", ft.selectedIdx)
	}

	// Move down 2
	ft.moveSelection(2)
	if ft.selectedIdx != 2 {
		t.Errorf("Expected selection at 2, got %d", ft.selectedIdx)
	}

	// Move up 1
	ft.moveSelection(-1)
	if ft.selectedIdx != 1 {
		t.Errorf("Expected selection at 1, got %d", ft.selectedIdx)
	}
}

// TestEmptyDirectory tests handling of empty directories
func TestEmptyDirectory(t *testing.T) {
	ft := New(30)

	tmpDir := t.TempDir()

	// Create empty subdirectory
	emptyDir := filepath.Join(tmpDir, "empty")
	os.MkdirAll(emptyDir, 0755)

	ft.SetRoot(tmpDir, "Test")

	// Expand empty directory
	ft.selectedIdx = 0
	ft.selectedPath = ft.visibleItems[0].Node.Path
	ft.expandSelected()

	// Should still have just 1 item (empty dir has no children)
	if len(ft.visibleItems) != 1 {
		t.Errorf("Expected 1 item for empty dir, got %d", len(ft.visibleItems))
	}

	// Directory should be marked as expanded
	if !ft.visibleItems[0].Node.Expanded {
		t.Error("Expected empty directory to be marked as expanded")
	}
}

// TestDepthTracking tests that depth is correctly tracked for nested items
func TestDepthTracking(t *testing.T) {
	ft := New(30)

	tmpDir := t.TempDir()

	// Create nested structure
	os.MkdirAll(filepath.Join(tmpDir, "level1", "level2"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "level1", "level2", "file.txt"), []byte("test"), 0644)

	ft.SetRoot(tmpDir, "Test")

	// Expand level1
	ft.expandSelected()

	// Expand level2
	ft.moveSelection(1) // Move to level2
	ft.expandSelected()

	// Check depths
	for _, item := range ft.visibleItems {
		switch {
		case item.Node.Name == "level1":
			if item.Depth != 0 {
				t.Errorf("Expected level1 depth 0, got %d", item.Depth)
			}
		case item.Node.Name == "level2":
			if item.Depth != 1 {
				t.Errorf("Expected level2 depth 1, got %d", item.Depth)
			}
		case item.Node.Name == "file.txt":
			if item.Depth != 2 {
				t.Errorf("Expected file.txt depth 2, got %d", item.Depth)
			}
		}
	}
}

// Helper function to check if a path contains a filename
func containsName(path, name string) bool {
	return filepath.Base(path) == name
}
