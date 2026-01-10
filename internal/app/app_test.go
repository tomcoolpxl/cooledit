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

package app

import (
	"os"
	"path/filepath"
	"testing"

	"cooledit/internal/config"
	"cooledit/internal/core"
	"cooledit/internal/fileio"
	"cooledit/internal/term"
)

type mockScreen struct {
	initCalled bool
	finiCalled bool
	pollCount  int
}

func (m *mockScreen) Init() error {
	m.initCalled = true
	return nil
}

func (m *mockScreen) Fini() {
	m.finiCalled = true
}

func (m *mockScreen) Size() (int, int) {
	return 80, 24
}

func (m *mockScreen) PollEvent() term.Event {
	m.pollCount++
	if m.pollCount == 1 {
		// Ctrl+Q to quit
		return term.KeyEvent{Key: term.KeyRune, Rune: 'q', Modifiers: term.ModCtrl}
	}
	// After first event, just return nil (though it should have exited)
	return nil
}

func (m *mockScreen) PushEvent(ev term.Event) {}

func (m *mockScreen) SetCell(x, y int, ch rune, style term.Style)             {}
func (m *mockScreen) Show()                                                   {}
func (m *mockScreen) SetCursorShape(shape term.CursorShape, color term.Color) {}
func (m *mockScreen) ShowCursor(x, y int)                                     {}
func (m *mockScreen) HideCursor()                                             {}

func TestRunWithScreenBasic(t *testing.T) {
	m := &mockScreen{}
	cfg := config.Default()
	// RunWithScreen will call Run which loops.
	// Our mock returns Ctrl+Q which sets quitNow=true.
	err := RunWithScreen("", false, cfg, m)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if !m.initCalled {
		t.Errorf("Init was not called")
	}
	if !m.finiCalled {
		t.Errorf("Fini was not called")
	}
}

func TestRunWithScreenNonExistentFile(t *testing.T) {
	m := &mockScreen{}
	cfg := config.Default()
	// Test opening a non-existent file - should not error
	err := RunWithScreen("nonexistent_file_that_does_not_exist.txt", false, cfg, m)
	if err != nil {
		t.Fatalf("Run with non-existent file failed: %v", err)
	}
	if !m.initCalled {
		t.Errorf("Init was not called")
	}
	if !m.finiCalled {
		t.Errorf("Fini was not called")
	}
}

// Test the core logic without UI
func TestNewFileSetup(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		shouldExist  bool
		wantBaseName string
		wantModified bool
	}{
		{
			name:         "non-existent file",
			path:         filepath.Join(t.TempDir(), "newfile.txt"),
			shouldExist:  false,
			wantBaseName: "newfile.txt",
			wantModified: false,
		},
		{
			name:         "non-existent file with path",
			path:         filepath.Join(t.TempDir(), "subdir", "another.go"),
			shouldExist:  false,
			wantBaseName: "another.go",
			wantModified: false,
		},
		{
			name:         "existing file",
			path:         "",
			shouldExist:  true,
			wantBaseName: "test.txt",
			wantModified: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			editor := core.NewEditor(nil)
			
			var path string
			if tt.shouldExist {
				// Create a temp file
				tmpfile, err := os.CreateTemp(t.TempDir(), "test*.txt")
				if err != nil {
					t.Fatal(err)
				}
				path = tmpfile.Name()
				tmpfile.WriteString("test content\n")
				tmpfile.Close()
				tt.wantBaseName = filepath.Base(path)
			} else {
				path = tt.path
			}

			// Simulate what RunWithScreen does
			fd, err := fileio.Open(path)
			if err != nil {
				// File doesn't exist - set up as new file
				editor.SetNewFile(path)
			} else {
				editor.LoadFile(fd)
			}

			// Verify editor state
			fileState := editor.File()
			if fileState.Path != path {
				t.Errorf("Path = %q, want %q", fileState.Path, path)
			}
			if fileState.BaseName != tt.wantBaseName {
				t.Errorf("BaseName = %q, want %q", fileState.BaseName, tt.wantBaseName)
			}
			if editor.Modified() != tt.wantModified {
				t.Errorf("Modified = %v, want %v", editor.Modified(), tt.wantModified)
			}

			// For non-existent files, verify it's empty
			if !tt.shouldExist {
				lines := editor.Lines()
				if len(lines) != 1 {
					t.Errorf("Expected 1 empty line, got %d lines", len(lines))
				}
				if len(lines) > 0 && len(lines[0]) != 0 {
					t.Errorf("Expected empty line, got %q", string(lines[0]))
				}
			}
		})
	}
}
