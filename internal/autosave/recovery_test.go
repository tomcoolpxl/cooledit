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

func TestFindRecoveryCandidateNoAutosave(t *testing.T) {
	tmpDir := t.TempDir()
	testPath := filepath.Join(tmpDir, "noautosave.txt")

	// No autosave exists
	candidate, err := FindRecoveryCandidate(testPath)
	if err != nil {
		t.Fatalf("FindRecoveryCandidate failed: %v", err)
	}

	if candidate != nil {
		t.Error("Should return nil when no autosave exists")
	}
}

func TestFindRecoveryCandidateEmptyPath(t *testing.T) {
	candidate, err := FindRecoveryCandidate("")
	if err != nil {
		t.Fatalf("FindRecoveryCandidate failed: %v", err)
	}

	if candidate != nil {
		t.Error("Should return nil for empty path")
	}
}

func TestFindRecoveryCandidateWithAutosave(t *testing.T) {
	tmpDir := t.TempDir()
	testPath := filepath.Join(tmpDir, "testfile.txt")

	// Create an autosave
	lines := [][]rune{[]rune("autosaved content")}
	err := WriteAutosave(testPath, lines, "\n", "UTF-8")
	if err != nil {
		t.Fatalf("WriteAutosave failed: %v", err)
	}

	// Find recovery candidate
	candidate, err := FindRecoveryCandidate(testPath)
	if err != nil {
		t.Fatalf("FindRecoveryCandidate failed: %v", err)
	}

	if candidate == nil {
		t.Fatal("Should find recovery candidate")
	}

	if candidate.Meta.OriginalPath != testPath {
		t.Errorf("OriginalPath mismatch: expected %q, got %q", testPath, candidate.Meta.OriginalPath)
	}

	// Original file doesn't exist
	if candidate.OriginalExists {
		t.Error("OriginalExists should be false")
	}

	// Cleanup
	_ = DeleteAutosave(testPath)
}

func TestFindRecoveryCandidateWithOriginalFile(t *testing.T) {
	tmpDir := t.TempDir()
	testPath := filepath.Join(tmpDir, "testfile.txt")

	// Create the original file
	if err := os.WriteFile(testPath, []byte("original content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Wait a moment to ensure timestamp difference
	time.Sleep(10 * time.Millisecond)

	// Create an autosave
	lines := [][]rune{[]rune("autosaved content")}
	err := WriteAutosave(testPath, lines, "\n", "UTF-8")
	if err != nil {
		t.Fatalf("WriteAutosave failed: %v", err)
	}

	// Find recovery candidate
	candidate, err := FindRecoveryCandidate(testPath)
	if err != nil {
		t.Fatalf("FindRecoveryCandidate failed: %v", err)
	}

	if candidate == nil {
		t.Fatal("Should find recovery candidate")
	}

	if !candidate.OriginalExists {
		t.Error("OriginalExists should be true")
	}

	if candidate.OriginalNewer {
		t.Error("OriginalNewer should be false (autosave is newer)")
	}

	// Cleanup
	_ = DeleteAutosave(testPath)
}

func TestPerformRecoveryRecover(t *testing.T) {
	tmpDir := t.TempDir()
	testPath := filepath.Join(tmpDir, "testfile.txt")

	// Create autosave
	autosaveContent := [][]rune{
		[]rune("recovered line 1"),
		[]rune("recovered line 2"),
	}
	err := WriteAutosave(testPath, autosaveContent, "\n", "UTF-8")
	if err != nil {
		t.Fatalf("WriteAutosave failed: %v", err)
	}

	// Find candidate
	candidate, err := FindRecoveryCandidate(testPath)
	if err != nil || candidate == nil {
		t.Fatalf("FindRecoveryCandidate failed")
	}

	// Perform recovery
	result, err := PerformRecovery(candidate, RecoveryRecover)
	if err != nil {
		t.Fatalf("PerformRecovery failed: %v", err)
	}

	if result == nil {
		t.Fatal("Result should not be nil")
	}

	if !result.FromAutosave {
		t.Error("FromAutosave should be true")
	}

	if len(result.Lines) != 2 {
		t.Errorf("Expected 2 lines, got %d", len(result.Lines))
	}

	if string(result.Lines[0]) != "recovered line 1" {
		t.Errorf("First line mismatch: got %q", string(result.Lines[0]))
	}

	// Autosave should still exist after recovery (kept until user saves)
	if !AutosaveExists(testPath) {
		t.Error("Autosave should still exist after recovery")
	}

	// Cleanup
	_ = DeleteAutosave(testPath)
}

func TestPerformRecoveryOpenOriginal(t *testing.T) {
	tmpDir := t.TempDir()
	testPath := filepath.Join(tmpDir, "testfile.txt")

	// Create autosave
	err := WriteAutosave(testPath, [][]rune{[]rune("autosave")}, "\n", "UTF-8")
	if err != nil {
		t.Fatalf("WriteAutosave failed: %v", err)
	}

	candidate, _ := FindRecoveryCandidate(testPath)

	// Open original
	result, err := PerformRecovery(candidate, RecoveryOpenOriginal)
	if err != nil {
		t.Fatalf("PerformRecovery failed: %v", err)
	}

	if result.FromAutosave {
		t.Error("FromAutosave should be false")
	}

	// Autosave should still exist
	if !AutosaveExists(testPath) {
		t.Error("Autosave should still exist after opening original")
	}

	// Cleanup
	_ = DeleteAutosave(testPath)
}

func TestPerformRecoveryDiscard(t *testing.T) {
	tmpDir := t.TempDir()
	testPath := filepath.Join(tmpDir, "testfile.txt")

	// Create autosave
	err := WriteAutosave(testPath, [][]rune{[]rune("autosave")}, "\n", "UTF-8")
	if err != nil {
		t.Fatalf("WriteAutosave failed: %v", err)
	}

	candidate, _ := FindRecoveryCandidate(testPath)

	// Discard
	result, err := PerformRecovery(candidate, RecoveryDiscard)
	if err != nil {
		t.Fatalf("PerformRecovery failed: %v", err)
	}

	if result.FromAutosave {
		t.Error("FromAutosave should be false")
	}

	// Autosave should be deleted
	if AutosaveExists(testPath) {
		t.Error("Autosave should be deleted after discard")
	}
}

func TestHasRecoveryFor(t *testing.T) {
	tmpDir := t.TempDir()
	testPath := filepath.Join(tmpDir, "testfile.txt")

	// No autosave yet
	if HasRecoveryFor(testPath) {
		t.Error("Should return false when no autosave exists")
	}

	// Empty path
	if HasRecoveryFor("") {
		t.Error("Should return false for empty path")
	}

	// Create autosave
	_ = WriteAutosave(testPath, [][]rune{[]rune("test")}, "\n", "UTF-8")

	if !HasRecoveryFor(testPath) {
		t.Error("Should return true when autosave exists")
	}

	// Cleanup
	_ = DeleteAutosave(testPath)
}
