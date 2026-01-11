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

// Package positionlog provides functionality to remember and restore
// cursor positions in recently edited files.
package positionlog

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
)

// MaxEntries is the maximum number of file positions to remember
const MaxEntries = 100

// Position represents a saved cursor position in a file
type Position struct {
	Line      int       `toml:"line"`
	Col       int       `toml:"col"`
	Timestamp time.Time `toml:"timestamp"`
}

// PositionLog holds all saved positions
type PositionLog struct {
	Positions map[string]Position `toml:"positions"`
	mu        sync.RWMutex
}

var (
	instance *PositionLog
	once     sync.Once
)

// Get returns the singleton PositionLog instance
func Get() *PositionLog {
	once.Do(func() {
		instance = &PositionLog{
			Positions: make(map[string]Position),
		}
		instance.load()
	})
	return instance
}

// positionLogDir returns the directory for storing the position log
func positionLogDir() (string, error) {
	var base string
	switch runtime.GOOS {
	case "windows":
		base = os.Getenv("APPDATA")
		if base == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", fmt.Errorf("cannot determine home directory: %w", err)
			}
			base = filepath.Join(home, "AppData", "Roaming")
		}
	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("cannot determine home directory: %w", err)
		}
		base = filepath.Join(home, "Library", "Application Support")
	default:
		// Linux and other Unix-like systems
		base = os.Getenv("XDG_DATA_HOME")
		if base == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", fmt.Errorf("cannot determine home directory: %w", err)
			}
			base = filepath.Join(home, ".local", "share")
		}
	}

	dir := filepath.Join(base, "cooledit")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", fmt.Errorf("cannot create position log directory: %w", err)
	}
	return dir, nil
}

// positionLogPath returns the full path to the position log file
func positionLogPath() (string, error) {
	dir, err := positionLogDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "positions.toml"), nil
}

// normalizeFilePath normalizes a file path for consistent storage
func normalizeFilePath(path string) string {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	return filepath.Clean(absPath)
}

// load reads the position log from disk
func (pl *PositionLog) load() {
	path, err := positionLogPath()
	if err != nil {
		return
	}

	pl.mu.Lock()
	defer pl.mu.Unlock()

	if _, err := toml.DecodeFile(path, pl); err != nil {
		// File doesn't exist or is invalid, start fresh
		pl.Positions = make(map[string]Position)
	}
}

// save writes the position log to disk
func (pl *PositionLog) save() error {
	path, err := positionLogPath()
	if err != nil {
		return err
	}

	pl.mu.RLock()
	defer pl.mu.RUnlock()

	// Write to temp file first for atomic update
	tmp := path + ".tmp"
	f, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}

	encoder := toml.NewEncoder(f)
	if err := encoder.Encode(pl); err != nil {
		f.Close()
		_ = os.Remove(tmp)
		return fmt.Errorf("failed to encode position log: %w", err)
	}
	f.Close()

	// Rename to final path
	if err := os.Rename(tmp, path); err != nil {
		// Fallback for Windows
		content, _ := os.ReadFile(tmp)
		if writeErr := os.WriteFile(path, content, 0600); writeErr != nil {
			_ = os.Remove(tmp)
			return fmt.Errorf("failed to write position log: %w", writeErr)
		}
		_ = os.Remove(tmp)
	}

	return nil
}

// SavePosition saves the cursor position for a file
func (pl *PositionLog) SavePosition(filePath string, line, col int) error {
	if filePath == "" {
		return nil // Don't save positions for unnamed buffers
	}

	normalizedPath := normalizeFilePath(filePath)

	pl.mu.Lock()
	pl.Positions[normalizedPath] = Position{
		Line:      line,
		Col:       col,
		Timestamp: time.Now(),
	}

	// Prune old entries if we exceed the limit
	if len(pl.Positions) > MaxEntries {
		pl.pruneOldest()
	}
	pl.mu.Unlock()

	return pl.save()
}

// GetPosition retrieves the saved cursor position for a file
// Returns (0, 0, false) if no position is saved
func (pl *PositionLog) GetPosition(filePath string) (line, col int, found bool) {
	if filePath == "" {
		return 0, 0, false
	}

	normalizedPath := normalizeFilePath(filePath)

	pl.mu.RLock()
	defer pl.mu.RUnlock()

	pos, ok := pl.Positions[normalizedPath]
	if !ok {
		return 0, 0, false
	}

	return pos.Line, pos.Col, true
}

// RemovePosition removes the saved position for a file
func (pl *PositionLog) RemovePosition(filePath string) error {
	if filePath == "" {
		return nil
	}

	normalizedPath := normalizeFilePath(filePath)

	pl.mu.Lock()
	delete(pl.Positions, normalizedPath)
	pl.mu.Unlock()

	return pl.save()
}

// pruneOldest removes the oldest entries to stay within MaxEntries limit
// Must be called with pl.mu already locked
func (pl *PositionLog) pruneOldest() {
	if len(pl.Positions) <= MaxEntries {
		return
	}

	// Create a slice of path-timestamp pairs for sorting
	type entry struct {
		path string
		time time.Time
	}
	entries := make([]entry, 0, len(pl.Positions))
	for path, pos := range pl.Positions {
		entries = append(entries, entry{path, pos.Timestamp})
	}

	// Sort by timestamp (oldest first)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].time.Before(entries[j].time)
	})

	// Remove oldest entries until we're at MaxEntries
	toRemove := len(pl.Positions) - MaxEntries
	for i := 0; i < toRemove; i++ {
		delete(pl.Positions, entries[i].path)
	}
}

// Clear removes all saved positions
func (pl *PositionLog) Clear() error {
	pl.mu.Lock()
	pl.Positions = make(map[string]Position)
	pl.mu.Unlock()

	return pl.save()
}
