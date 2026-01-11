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

package positionlog

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPositionLog_SaveAndGet(t *testing.T) {
	// Create a temp file for testing
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	// Create a fresh position log for testing
	pl := &PositionLog{
		Positions: make(map[string]Position),
	}

	// Save a position
	err := pl.SavePosition(testFile, 10, 5)
	if err != nil {
		t.Fatalf("SavePosition failed: %v", err)
	}

	// Get the position
	line, col, found := pl.GetPosition(testFile)
	if !found {
		t.Fatal("GetPosition should return found=true")
	}
	if line != 10 {
		t.Errorf("Expected line 10, got %d", line)
	}
	if col != 5 {
		t.Errorf("Expected col 5, got %d", col)
	}
}

func TestPositionLog_GetNotFound(t *testing.T) {
	pl := &PositionLog{
		Positions: make(map[string]Position),
	}

	_, _, found := pl.GetPosition("/nonexistent/file.txt")
	if found {
		t.Error("GetPosition should return found=false for unknown file")
	}
}

func TestPositionLog_EmptyPath(t *testing.T) {
	pl := &PositionLog{
		Positions: make(map[string]Position),
	}

	// Empty path should not save
	err := pl.SavePosition("", 10, 5)
	if err != nil {
		t.Fatalf("SavePosition with empty path should not error: %v", err)
	}

	// Should have no positions
	if len(pl.Positions) != 0 {
		t.Error("Expected no positions for empty path")
	}

	// Empty path should not find
	_, _, found := pl.GetPosition("")
	if found {
		t.Error("GetPosition should return found=false for empty path")
	}
}

func TestPositionLog_RemovePosition(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	pl := &PositionLog{
		Positions: make(map[string]Position),
	}

	// Save a position
	pl.SavePosition(testFile, 10, 5)

	// Remove the position
	err := pl.RemovePosition(testFile)
	if err != nil {
		t.Fatalf("RemovePosition failed: %v", err)
	}

	// Should not be found anymore
	_, _, found := pl.GetPosition(testFile)
	if found {
		t.Error("GetPosition should return found=false after remove")
	}
}

func TestPositionLog_Clear(t *testing.T) {
	tmpDir := t.TempDir()

	pl := &PositionLog{
		Positions: make(map[string]Position),
	}

	// Save multiple positions
	pl.SavePosition(filepath.Join(tmpDir, "file1.txt"), 1, 1)
	pl.SavePosition(filepath.Join(tmpDir, "file2.txt"), 2, 2)
	pl.SavePosition(filepath.Join(tmpDir, "file3.txt"), 3, 3)

	// Clear all
	err := pl.Clear()
	if err != nil {
		t.Fatalf("Clear failed: %v", err)
	}

	// Should have no positions
	if len(pl.Positions) != 0 {
		t.Error("Expected no positions after clear")
	}
}

func TestPositionLog_PruneOldest(t *testing.T) {
	pl := &PositionLog{
		Positions: make(map[string]Position),
	}

	// Add more than MaxEntries positions
	for i := 0; i <= MaxEntries+10; i++ {
		pl.Positions[filepath.Join("/tmp", "file"+string(rune('a'+i%26))+".txt")] = Position{
			Line: i,
			Col:  i,
		}
	}

	// Prune
	pl.pruneOldest()

	// Should be at MaxEntries
	if len(pl.Positions) > MaxEntries {
		t.Errorf("Expected at most %d positions, got %d", MaxEntries, len(pl.Positions))
	}
}

func TestPositionLog_NormalizePath(t *testing.T) {
	tmpDir := t.TempDir()

	pl := &PositionLog{
		Positions: make(map[string]Position),
	}

	// Save with relative-ish path
	testFile := filepath.Join(tmpDir, "test.txt")
	pl.SavePosition(testFile, 10, 5)

	// Should find with normalized path
	_, _, found := pl.GetPosition(testFile)
	if !found {
		t.Error("GetPosition should find position with same path")
	}
}

func TestPositionLogPath(t *testing.T) {
	path, err := positionLogPath()
	if err != nil {
		t.Fatalf("positionLogPath failed: %v", err)
	}

	// Should contain positions.toml
	if filepath.Base(path) != "positions.toml" {
		t.Errorf("Expected positions.toml, got %s", filepath.Base(path))
	}

	// Directory should exist or be creatable
	dir := filepath.Dir(path)
	_, err = os.Stat(dir)
	if err != nil && !os.IsNotExist(err) {
		t.Errorf("Directory check failed: %v", err)
	}
}
