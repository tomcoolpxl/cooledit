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
	"fmt"
	"hash/fnv"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

// AutosaveMeta contains metadata about an autosave file
type AutosaveMeta struct {
	OriginalPath    string    `toml:"original_path"`
	Encoding        string    `toml:"encoding"`
	EOL             string    `toml:"eol"`
	Timestamp       time.Time `toml:"timestamp"`
	CoolEditVersion string    `toml:"cooledit_version"`
}

// Version is set by the build process
var Version = "dev"

// AutosaveDir returns the path to the autosave directory.
// Creates the directory if it doesn't exist.
func AutosaveDir() (string, error) {
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

	dir := filepath.Join(base, "cooledit", "autosave")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", fmt.Errorf("cannot create autosave directory: %w", err)
	}
	return dir, nil
}

// PathToFilename converts a file path to a safe autosave filename using FNV-1a hash.
// This ensures cross-platform compatibility and avoids issues with special characters.
func PathToFilename(path string) string {
	// Normalize path separators and case (Windows is case-insensitive)
	normalized := filepath.Clean(path)
	if runtime.GOOS == "windows" {
		normalized = strings.ToLower(normalized)
	}

	// FNV-1a hash for speed and simplicity
	h := fnv.New64a()
	h.Write([]byte(normalized))
	return fmt.Sprintf("%016x", h.Sum64())
}

// AutosavePath returns the full path to the autosave file for a given original path.
// For unnamed buffers (empty path), returns empty string.
func AutosavePath(originalPath string) (string, error) {
	if originalPath == "" {
		return "", nil
	}

	dir, err := AutosaveDir()
	if err != nil {
		return "", err
	}

	filename := PathToFilename(originalPath)
	return filepath.Join(dir, filename+".autosave"), nil
}

// MetaPath returns the full path to the metadata file for a given original path.
func MetaPath(originalPath string) (string, error) {
	if originalPath == "" {
		return "", nil
	}

	dir, err := AutosaveDir()
	if err != nil {
		return "", err
	}

	filename := PathToFilename(originalPath)
	return filepath.Join(dir, filename+".meta"), nil
}

// WriteAutosave writes the buffer content to the autosave file.
// Uses atomic write (write to temp, then rename) for safety.
func WriteAutosave(originalPath string, lines [][]rune, eol string, encoding string) error {
	autosavePath, err := AutosavePath(originalPath)
	if err != nil {
		return err
	}
	if autosavePath == "" {
		return fmt.Errorf("cannot autosave unnamed buffer")
	}

	// Build file content
	var data []byte
	for i, line := range lines {
		for _, r := range line {
			if encoding == "ISO-8859-1" && r > 255 {
				r = '?'
			}
			data = append(data, byte(r))
		}
		if i < len(lines)-1 {
			if eol == "\r\n" {
				data = append(data, '\r', '\n')
			} else {
				data = append(data, '\n')
			}
		}
	}

	// Atomic write: write to temp file, then rename
	tmp := autosavePath + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return fmt.Errorf("failed to write autosave temp file: %w", err)
	}

	if err := os.Rename(tmp, autosavePath); err != nil {
		// Fallback for Windows where rename may fail if target exists
		if writeErr := os.WriteFile(autosavePath, data, 0600); writeErr != nil {
			_ = os.Remove(tmp)
			return fmt.Errorf("failed to write autosave file: %w", writeErr)
		}
		_ = os.Remove(tmp)
	}

	// Write metadata
	return WriteMeta(originalPath, encoding, eol)
}

// WriteMeta writes the metadata file for an autosave.
func WriteMeta(originalPath string, encoding string, eol string) error {
	metaPath, err := MetaPath(originalPath)
	if err != nil {
		return err
	}
	if metaPath == "" {
		return nil
	}

	meta := AutosaveMeta{
		OriginalPath:    originalPath,
		Encoding:        encoding,
		EOL:             eol,
		Timestamp:       time.Now(),
		CoolEditVersion: Version,
	}

	// Write to temp file first
	tmp := metaPath + ".tmp"
	f, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to create meta temp file: %w", err)
	}

	encoder := toml.NewEncoder(f)
	if err := encoder.Encode(meta); err != nil {
		f.Close()
		_ = os.Remove(tmp)
		return fmt.Errorf("failed to encode meta file: %w", err)
	}
	f.Close()

	// Rename to final path
	if err := os.Rename(tmp, metaPath); err != nil {
		// Fallback for Windows
		content, _ := os.ReadFile(tmp)
		if writeErr := os.WriteFile(metaPath, content, 0600); writeErr != nil {
			_ = os.Remove(tmp)
			return fmt.Errorf("failed to write meta file: %w", writeErr)
		}
		_ = os.Remove(tmp)
	}

	return nil
}

// ReadMeta reads the metadata file for an autosave.
func ReadMeta(originalPath string) (*AutosaveMeta, error) {
	metaPath, err := MetaPath(originalPath)
	if err != nil {
		return nil, err
	}
	if metaPath == "" {
		return nil, nil
	}

	var meta AutosaveMeta
	if _, err := toml.DecodeFile(metaPath, &meta); err != nil {
		return nil, fmt.Errorf("failed to read meta file: %w", err)
	}

	return &meta, nil
}

// DeleteAutosave removes both the autosave file and its metadata.
func DeleteAutosave(originalPath string) error {
	autosavePath, err := AutosavePath(originalPath)
	if err != nil {
		return err
	}
	metaPath, err := MetaPath(originalPath)
	if err != nil {
		return err
	}

	// Remove both files, ignoring "not exist" errors
	var lastErr error
	if autosavePath != "" {
		if err := os.Remove(autosavePath); err != nil && !os.IsNotExist(err) {
			lastErr = err
		}
	}
	if metaPath != "" {
		if err := os.Remove(metaPath); err != nil && !os.IsNotExist(err) {
			lastErr = err
		}
	}

	return lastErr
}

// AutosaveExists checks if an autosave file exists for the given original path.
func AutosaveExists(originalPath string) bool {
	autosavePath, err := AutosavePath(originalPath)
	if err != nil || autosavePath == "" {
		return false
	}

	_, err = os.Stat(autosavePath)
	return err == nil
}

// ReadAutosaveContent reads the content of an autosave file and returns it as lines.
func ReadAutosaveContent(originalPath string) ([][]rune, error) {
	autosavePath, err := AutosavePath(originalPath)
	if err != nil {
		return nil, err
	}
	if autosavePath == "" {
		return nil, fmt.Errorf("no autosave path for empty original path")
	}

	data, err := os.ReadFile(autosavePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read autosave file: %w", err)
	}

	// Parse into lines (handle both LF and CRLF)
	var lines [][]rune
	var currentLine []rune

	for i := 0; i < len(data); i++ {
		if data[i] == '\r' && i+1 < len(data) && data[i+1] == '\n' {
			// CRLF - skip the \r, \n will be handled next
			continue
		} else if data[i] == '\n' {
			lines = append(lines, currentLine)
			currentLine = nil
		} else {
			currentLine = append(currentLine, rune(data[i]))
		}
	}
	// Don't forget the last line
	lines = append(lines, currentLine)

	return lines, nil
}
