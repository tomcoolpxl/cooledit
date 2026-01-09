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

package fileio

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOpenAndSaveUTF8(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "utf8.txt")
	content := []byte("hello\nworld")
	
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}
	
	fd, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	
	if fd.Encoding != "UTF-8" {
		t.Errorf("expected UTF-8, got %s", fd.Encoding)
	}
	if fd.EOL != "\n" {
		t.Errorf("expected LF, got %q", fd.EOL)
	}
	if len(fd.Lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(fd.Lines))
	}
	if string(fd.Lines[0]) != "hello" {
		t.Errorf("expected 'hello', got %q", string(fd.Lines[0]))
	}
	
	// Save back
	if err := Save(path, fd.Lines, fd.EOL, fd.Encoding); err != nil {
		t.Fatal(err)
	}
	
	readBack, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(readBack) != string(content) {
		t.Errorf("content mismatch after save: %q != %q", string(readBack), string(content))
	}
}

func TestOpenCRLF(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "crlf.txt")
	content := []byte("hello\r\nworld")
	
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}
	
	fd, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	
	if fd.EOL != "\r\n" {
		t.Errorf("expected CRLF, got %q", fd.EOL)
	}
	if string(fd.Lines[0]) != "hello" {
		t.Errorf("expected 'hello', got %q", string(fd.Lines[0]))
	}
	
	// Save should preserve CRLF
	if err := Save(path, fd.Lines, fd.EOL, fd.Encoding); err != nil {
		t.Fatal(err)
	}
	
	readBack, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(readBack) != string(content) {
		t.Errorf("content mismatch: %q", string(readBack))
	}
}

func TestOpenISO88591(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "latin1.txt")
	// 0xFF is valid ISO-8859-1 (ÿ) but invalid UTF-8
	content := []byte{0x68, 0x65, 0xFF, 0x6C, 0x6F} // heÿlo
	
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}
	
	fd, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	
	if fd.Encoding != "ISO-8859-1" {
		t.Errorf("expected ISO-8859-1, got %s", fd.Encoding)
	}
	
	// Check internal representation (should be rune 255)
	if len(fd.Lines[0]) != 5 {
		t.Fatalf("expected length 5, got %d", len(fd.Lines[0]))
	}
	if fd.Lines[0][2] != 0xFF {
		t.Errorf("expected rune 0xFF, got %x", fd.Lines[0][2])
	}
	
	// Save back
	if err := Save(path, fd.Lines, fd.EOL, fd.Encoding); err != nil {
		t.Fatal(err)
	}
	
	readBack, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(readBack) != string(content) {
		t.Errorf("content mismatch: %x != %x", readBack, content)
	}
}

func TestOpenError(t *testing.T) {
	_, err := Open("non_existent_file.txt")
	if err == nil {
		t.Fatal("expected error opening non-existent file")
	}
}

func TestSaveError(t *testing.T) {
	// Try to save to a directory path
	dir := t.TempDir()
	err := Save(dir, [][]rune{}, "\n", "UTF-8")
	if err == nil {
		t.Fatal("expected error saving to directory path")
	}
}
