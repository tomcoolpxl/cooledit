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

package autosave

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestPathToFilename(t *testing.T) {
	// Basic test - should produce consistent hash
	path1 := "/path/to/file.txt"
	filename1 := PathToFilename(path1)
	filename2 := PathToFilename(path1)

	if filename1 != filename2 {
		t.Errorf("PathToFilename should be deterministic: got %s and %s", filename1, filename2)
	}

	// Different paths should produce different filenames
	path2 := "/path/to/other.txt"
	filename3 := PathToFilename(path2)

	if filename1 == filename3 {
		t.Errorf("Different paths should produce different filenames")
	}

	// Filename should be 16 hex characters
	if len(filename1) != 16 {
		t.Errorf("Filename should be 16 characters, got %d", len(filename1))
	}
}

func TestAutosaveDirCreation(t *testing.T) {
	dir, err := AutosaveDir()
	if err != nil {
		t.Fatalf("AutosaveDir failed: %v", err)
	}

	// Directory should exist
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("Autosave directory should exist: %v", err)
	}

	if !info.IsDir() {
		t.Error("Autosave path should be a directory")
	}
}

func TestWriteAndReadAutosave(t *testing.T) {
	// Create a temporary test file path
	tmpDir := t.TempDir()
	testPath := filepath.Join(tmpDir, "testfile.txt")

	// Write some content
	lines := [][]rune{
		[]rune("Hello, World!"),
		[]rune("Line 2"),
		[]rune("Line 3"),
	}

	err := WriteAutosave(testPath, lines, "\n", "UTF-8")
	if err != nil {
		t.Fatalf("WriteAutosave failed: %v", err)
	}

	// Check that autosave exists
	if !AutosaveExists(testPath) {
		t.Error("Autosave file should exist")
	}

	// Read it back
	readLines, err := ReadAutosaveContent(testPath)
	if err != nil {
		t.Fatalf("ReadAutosaveContent failed: %v", err)
	}

	// Verify content
	if len(readLines) != len(lines) {
		t.Errorf("Expected %d lines, got %d", len(lines), len(readLines))
	}

	for i, line := range lines {
		if string(readLines[i]) != string(line) {
			t.Errorf("Line %d mismatch: expected %q, got %q", i, string(line), string(readLines[i]))
		}
	}

	// Clean up
	_ = DeleteAutosave(testPath)

	// Verify cleanup
	if AutosaveExists(testPath) {
		t.Error("Autosave file should be deleted")
	}
}

func TestWriteAndReadMeta(t *testing.T) {
	tmpDir := t.TempDir()
	testPath := filepath.Join(tmpDir, "testfile.txt")

	// Write autosave (which also writes meta)
	lines := [][]rune{[]rune("test")}
	err := WriteAutosave(testPath, lines, "\r\n", "ISO-8859-1")
	if err != nil {
		t.Fatalf("WriteAutosave failed: %v", err)
	}

	// Read metadata
	meta, err := ReadMeta(testPath)
	if err != nil {
		t.Fatalf("ReadMeta failed: %v", err)
	}

	if meta.OriginalPath != testPath {
		t.Errorf("OriginalPath mismatch: expected %q, got %q", testPath, meta.OriginalPath)
	}

	if meta.EOL != "\r\n" {
		t.Errorf("EOL mismatch: expected CRLF, got %q", meta.EOL)
	}

	if meta.Encoding != "ISO-8859-1" {
		t.Errorf("Encoding mismatch: expected ISO-8859-1, got %q", meta.Encoding)
	}

	// Cleanup
	_ = DeleteAutosave(testPath)
}

func TestManagerNotifyEdit(t *testing.T) {
	var autosaveCalled bool
	var lastState AutosaveState

	manager := NewManager(true, 50*time.Millisecond, 0)
	manager.SetStateProvider(func() AutosaveState {
		return AutosaveState{
			Lines:    [][]rune{[]rune("test content")},
			Path:     "/test/path.txt",
			EOL:      "\n",
			Encoding: "UTF-8",
			Modified: true,
		}
	})
	manager.SetErrorCallback(func(err error) {
		autosaveCalled = true
		lastState = manager.getState()
	})

	// Notify edit
	manager.NotifyEdit()

	// Wait for timer to fire (a bit longer than idle timeout)
	time.Sleep(100 * time.Millisecond)

	// The callback should have been called
	if !autosaveCalled {
		t.Error("Autosave callback should have been called after idle timeout")
	}

	if lastState.Path != "/test/path.txt" {
		t.Errorf("State path mismatch: expected /test/path.txt, got %s", lastState.Path)
	}

	manager.Stop()
}

func TestManagerDisabled(t *testing.T) {
	autosaveCalled := false

	manager := NewManager(false, 10*time.Millisecond, 0)
	manager.SetStateProvider(func() AutosaveState {
		return AutosaveState{Modified: true, Path: "/test"}
	})
	manager.SetErrorCallback(func(err error) {
		autosaveCalled = true
	})

	manager.NotifyEdit()
	time.Sleep(50 * time.Millisecond)

	if autosaveCalled {
		t.Error("Autosave should not be called when disabled")
	}
}

func TestManagerNotModified(t *testing.T) {
	autosaveCalled := false

	manager := NewManager(true, 10*time.Millisecond, 0)
	manager.SetStateProvider(func() AutosaveState {
		return AutosaveState{Modified: false, Path: "/test"}
	})
	manager.SetErrorCallback(func(err error) {
		autosaveCalled = true
	})

	manager.NotifyEdit()
	time.Sleep(50 * time.Millisecond)

	if autosaveCalled {
		t.Error("Autosave should not be called when buffer is not modified")
	}

	manager.Stop()
}

func TestManagerEmptyPath(t *testing.T) {
	autosaveCalled := false

	manager := NewManager(true, 10*time.Millisecond, 0)
	manager.SetStateProvider(func() AutosaveState {
		return AutosaveState{Modified: true, Path: ""} // Unnamed buffer
	})
	manager.SetErrorCallback(func(err error) {
		autosaveCalled = true
	})

	manager.NotifyEdit()
	time.Sleep(50 * time.Millisecond)

	if autosaveCalled {
		t.Error("Autosave should not be called for unnamed buffer")
	}

	manager.Stop()
}

func TestManagerToggleEnabled(t *testing.T) {
	manager := NewManager(true, time.Second, time.Second)

	if !manager.IsEnabled() {
		t.Error("Manager should be enabled initially")
	}

	manager.SetEnabled(false)
	if manager.IsEnabled() {
		t.Error("Manager should be disabled after SetEnabled(false)")
	}

	manager.SetEnabled(true)
	if !manager.IsEnabled() {
		t.Error("Manager should be enabled after SetEnabled(true)")
	}
}

func TestDeleteNonExistent(t *testing.T) {
	// Deleting a non-existent autosave should not error
	err := DeleteAutosave("/nonexistent/path/file.txt")
	if err != nil {
		t.Errorf("DeleteAutosave should not error for non-existent file: %v", err)
	}
}

func TestManagerForceAutosave(t *testing.T) {
	tmpDir := t.TempDir()
	testPath := filepath.Join(tmpDir, "forcesave.txt")

	manager := NewManager(true, time.Hour, time.Hour) // Long timeouts
	manager.SetStateProvider(func() AutosaveState {
		return AutosaveState{
			Lines:    [][]rune{[]rune("force saved content")},
			Path:     testPath,
			EOL:      "\n",
			Encoding: "UTF-8",
			Modified: true,
		}
	})

	// Force autosave should work immediately
	err := manager.ForceAutosave()
	if err != nil {
		t.Fatalf("ForceAutosave failed: %v", err)
	}

	// Verify autosave exists
	if !AutosaveExists(testPath) {
		t.Error("Autosave should exist after ForceAutosave")
	}

	// Cleanup
	_ = DeleteAutosave(testPath)
	manager.Stop()
}

func TestManagerForceAutosaveNotModified(t *testing.T) {
	manager := NewManager(true, time.Hour, time.Hour)
	manager.SetStateProvider(func() AutosaveState {
		return AutosaveState{
			Modified: false,
			Path:     "/test/path",
		}
	})

	// Should return nil and do nothing for unmodified buffer
	err := manager.ForceAutosave()
	if err != nil {
		t.Errorf("ForceAutosave should not error for unmodified buffer: %v", err)
	}
}

func TestManagerForceAutosaveDisabled(t *testing.T) {
	manager := NewManager(false, time.Hour, time.Hour)
	manager.SetStateProvider(func() AutosaveState {
		return AutosaveState{Modified: true, Path: "/test"}
	})

	err := manager.ForceAutosave()
	if err != nil {
		t.Errorf("ForceAutosave should not error when disabled: %v", err)
	}
}

func TestManagerClearAutosave(t *testing.T) {
	tmpDir := t.TempDir()
	testPath := filepath.Join(tmpDir, "toclear.txt")

	// Create an autosave
	_ = WriteAutosave(testPath, [][]rune{[]rune("test")}, "\n", "UTF-8")
	if !AutosaveExists(testPath) {
		t.Fatal("Autosave should exist before clear")
	}

	manager := NewManager(true, time.Hour, time.Hour)

	// Clear it
	err := manager.ClearAutosave(testPath)
	if err != nil {
		t.Fatalf("ClearAutosave failed: %v", err)
	}

	if AutosaveExists(testPath) {
		t.Error("Autosave should be deleted after ClearAutosave")
	}
}

func TestManagerClearAutosaveEmptyPath(t *testing.T) {
	manager := NewManager(true, time.Hour, time.Hour)
	err := manager.ClearAutosave("")
	if err != nil {
		t.Errorf("ClearAutosave should not error for empty path: %v", err)
	}
}

func TestManagerClearCurrentAutosave(t *testing.T) {
	tmpDir := t.TempDir()
	testPath := filepath.Join(tmpDir, "current.txt")

	manager := NewManager(true, time.Hour, time.Hour)
	manager.SetStateProvider(func() AutosaveState {
		return AutosaveState{
			Lines:    [][]rune{[]rune("test")},
			Path:     testPath,
			EOL:      "\n",
			Encoding: "UTF-8",
			Modified: true,
		}
	})

	// Force autosave to set currentPath
	_ = manager.ForceAutosave()
	if !AutosaveExists(testPath) {
		t.Fatal("Autosave should exist")
	}

	// Clear current
	err := manager.ClearCurrentAutosave()
	if err != nil {
		t.Fatalf("ClearCurrentAutosave failed: %v", err)
	}

	if AutosaveExists(testPath) {
		t.Error("Current autosave should be deleted")
	}
}

func TestManagerUpdatePath(t *testing.T) {
	manager := NewManager(true, time.Hour, time.Hour)

	manager.UpdatePath("/new/path.txt")

	// Since currentPath is private, we can test indirectly
	// by verifying the manager doesn't panic
}

func TestManagerReset(t *testing.T) {
	manager := NewManager(true, 10*time.Millisecond, 0)
	manager.SetStateProvider(func() AutosaveState {
		return AutosaveState{Modified: true, Path: "/test"}
	})

	// Start a timer
	manager.NotifyEdit()

	// Reset should cancel the timer
	manager.Reset()

	// Wait to ensure timer would have fired
	time.Sleep(50 * time.Millisecond)

	// No way to directly test timer is cancelled, but at least verify no panic
}

func TestManagerSetEnabledCancelsTimer(t *testing.T) {
	autosaveCalled := false

	manager := NewManager(true, 50*time.Millisecond, 0)
	manager.SetStateProvider(func() AutosaveState {
		return AutosaveState{Modified: true, Path: "/test"}
	})
	manager.SetErrorCallback(func(err error) {
		autosaveCalled = true
	})

	// Start timer
	manager.NotifyEdit()

	// Disable immediately (should cancel timer)
	manager.SetEnabled(false)

	// Wait for timer to have fired (if it wasn't cancelled)
	time.Sleep(100 * time.Millisecond)

	if autosaveCalled {
		t.Error("Autosave should not be called when disabled before timer fires")
	}
}

func TestWriteAutosaveEmptyPath(t *testing.T) {
	err := WriteAutosave("", [][]rune{[]rune("test")}, "\n", "UTF-8")
	if err == nil {
		t.Error("WriteAutosave should error for empty path")
	}
}

func TestAutosavePathEmptyPath(t *testing.T) {
	path, err := AutosavePath("")
	if err != nil {
		t.Errorf("AutosavePath should not error for empty path: %v", err)
	}
	if path != "" {
		t.Error("AutosavePath should return empty for empty input")
	}
}

func TestMetaPathEmptyPath(t *testing.T) {
	path, err := MetaPath("")
	if err != nil {
		t.Errorf("MetaPath should not error for empty path: %v", err)
	}
	if path != "" {
		t.Error("MetaPath should return empty for empty input")
	}
}

func TestWriteAutosaveCRLF(t *testing.T) {
	tmpDir := t.TempDir()
	testPath := filepath.Join(tmpDir, "crlf.txt")

	lines := [][]rune{
		[]rune("Line 1"),
		[]rune("Line 2"),
	}

	err := WriteAutosave(testPath, lines, "\r\n", "UTF-8")
	if err != nil {
		t.Fatalf("WriteAutosave failed: %v", err)
	}

	// Read the raw content to verify CRLF
	autosavePath, _ := AutosavePath(testPath)
	content, err := os.ReadFile(autosavePath)
	if err != nil {
		t.Fatalf("Failed to read autosave: %v", err)
	}

	// Should contain CRLF
	if !contains(content, []byte("\r\n")) {
		t.Error("Autosave should contain CRLF line endings")
	}

	_ = DeleteAutosave(testPath)
}

func contains(haystack, needle []byte) bool {
	for i := 0; i <= len(haystack)-len(needle); i++ {
		match := true
		for j := 0; j < len(needle); j++ {
			if haystack[i+j] != needle[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

func TestReadMetaNoFile(t *testing.T) {
	tmpDir := t.TempDir()
	testPath := filepath.Join(tmpDir, "nometa.txt")

	meta, err := ReadMeta(testPath)
	if err == nil {
		t.Error("ReadMeta should error when meta file doesn't exist")
	}
	if meta != nil {
		t.Error("Meta should be nil when file doesn't exist")
	}
}
