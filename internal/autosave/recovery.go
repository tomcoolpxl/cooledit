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
	"time"
)

// RecoveryCandidate represents a potential recovery file
type RecoveryCandidate struct {
	// AutosavePath is the full path to the autosave file
	AutosavePath string

	// Meta contains the autosave metadata (may have defaults if meta file missing)
	Meta AutosaveMeta

	// OriginalExists is true if the original file still exists
	OriginalExists bool

	// OriginalModTime is the modification time of the original file (if it exists)
	OriginalModTime time.Time

	// OriginalNewer is true if the original file was modified after the autosave
	OriginalNewer bool
}

// FindRecoveryCandidate checks if there's an autosave file for the given path.
// Returns nil if no recovery candidate exists.
// For unnamed buffers (empty path), returns nil (unnamed recovery not yet supported).
func FindRecoveryCandidate(targetPath string) (*RecoveryCandidate, error) {
	if targetPath == "" {
		// Unnamed buffer recovery not implemented yet
		return nil, nil
	}

	// Check if autosave exists
	if !AutosaveExists(targetPath) {
		return nil, nil
	}

	// Get autosave path
	autosavePath, err := AutosavePath(targetPath)
	if err != nil {
		return nil, err
	}

	// Try to read metadata
	meta, err := ReadMeta(targetPath)
	if err != nil {
		// Metadata missing or corrupt - use defaults
		meta = &AutosaveMeta{
			OriginalPath: targetPath,
			Encoding:     "UTF-8",
			EOL:          "\n",
		}

		// Try to get timestamp from autosave file itself
		if info, statErr := os.Stat(autosavePath); statErr == nil {
			meta.Timestamp = info.ModTime()
		}
	}

	candidate := &RecoveryCandidate{
		AutosavePath: autosavePath,
		Meta:         *meta,
	}

	// Check if original file exists and compare timestamps
	if info, err := os.Stat(targetPath); err == nil {
		candidate.OriginalExists = true
		candidate.OriginalModTime = info.ModTime()
		candidate.OriginalNewer = info.ModTime().After(meta.Timestamp)
	}

	return candidate, nil
}

// RecoveryAction represents the user's choice during recovery
type RecoveryAction int

const (
	// RecoveryRecover loads the autosave content, marks buffer as modified
	RecoveryRecover RecoveryAction = iota

	// RecoveryOpenOriginal loads the original file, keeps autosave for later
	RecoveryOpenOriginal

	// RecoveryDiscard deletes the autosave and loads the original file
	RecoveryDiscard
)

// RecoveredFile contains the result of recovering an autosave
type RecoveredFile struct {
	// Lines is the file content as lines of runes
	Lines [][]rune

	// Path is the original file path
	Path string

	// Encoding is the file encoding
	Encoding string

	// EOL is the line ending style
	EOL string

	// FromAutosave is true if content came from autosave (should be marked modified)
	FromAutosave bool
}

// PerformRecovery executes the recovery action chosen by the user.
// Returns the file content to load and whether it should be marked as modified.
func PerformRecovery(candidate *RecoveryCandidate, action RecoveryAction) (*RecoveredFile, error) {
	switch action {
	case RecoveryRecover:
		// Load autosave content
		lines, err := ReadAutosaveContent(candidate.Meta.OriginalPath)
		if err != nil {
			return nil, err
		}

		return &RecoveredFile{
			Lines:        lines,
			Path:         candidate.Meta.OriginalPath,
			Encoding:     candidate.Meta.Encoding,
			EOL:          candidate.Meta.EOL,
			FromAutosave: true, // Mark as modified since it's unsaved changes
		}, nil

	case RecoveryOpenOriginal:
		// Just return info - caller will load the file normally
		// Keep autosave file for potential future recovery
		return &RecoveredFile{
			Path:         candidate.Meta.OriginalPath,
			FromAutosave: false,
		}, nil

	case RecoveryDiscard:
		// Delete autosave files
		_ = DeleteAutosave(candidate.Meta.OriginalPath)

		// Return info - caller will load the file normally
		return &RecoveredFile{
			Path:         candidate.Meta.OriginalPath,
			FromAutosave: false,
		}, nil

	default:
		return nil, nil
	}
}

// HasRecoveryFor is a convenience function to check if recovery exists for a path.
func HasRecoveryFor(path string) bool {
	if path == "" {
		return false
	}
	return AutosaveExists(path)
}
