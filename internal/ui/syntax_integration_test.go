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

package ui

import (
	"strings"
	"testing"

	"cooledit/internal/config"
	"cooledit/internal/core"
	"cooledit/internal/fileio"
	"cooledit/internal/syntax"
)

// Helper to convert a string to lines for the editor
func stringToLines(s string) [][]rune {
	parts := strings.Split(s, "\n")
	lines := make([][]rune, len(parts))
	for i, p := range parts {
		lines[i] = []rune(p)
	}
	return lines
}

// Helper to create an editor with content and filename
func newTestEditorWithContent(content, filename string) *core.Editor {
	editor := core.NewEditor(nil)
	fd := &fileio.FileData{
		Path:     filename,
		BaseName: filename,
		Lines:    stringToLines(content),
		EOL:      "\n",
		Encoding: "UTF-8",
	}
	editor.LoadFile(fd)
	return editor
}

// TestSyntaxAutoDetectionGoFile tests that .go files are auto-detected as Go
func TestSyntaxAutoDetectionGoFile(t *testing.T) {
	screen := NewFakeScreen(80, 24)
	editor := newTestEditorWithContent("package main\n\nfunc main() {\n\tprintln(\"hello\")\n}\n", "test.go")

	cfg := config.Default()
	cfg.Editor.SyntaxHighlighting = true
	cfg.UI.Language = "" // Auto-detect

	ui := New(screen, editor, cfg)

	// Check that language was detected
	lang := ui.GetCurrentLanguage()
	if lang != "Go" {
		t.Errorf("Expected language 'Go' for test.go, got %q", lang)
	}

	// Check that syntax cache was created
	if ui.syntaxCache == nil {
		t.Error("Expected syntaxCache to be initialized for Go file")
	}
}

// TestSyntaxAutoDetectionPythonFile tests that .py files are auto-detected as Python
func TestSyntaxAutoDetectionPythonFile(t *testing.T) {
	screen := NewFakeScreen(80, 24)
	editor := newTestEditorWithContent("def hello():\n    print('hello')\n", "script.py")

	cfg := config.Default()
	cfg.Editor.SyntaxHighlighting = true
	cfg.UI.Language = "" // Auto-detect

	ui := New(screen, editor, cfg)

	lang := ui.GetCurrentLanguage()
	if lang != "Python" {
		t.Errorf("Expected language 'Python' for script.py, got %q", lang)
	}
}

// TestSyntaxAutoDetectionShebang tests shebang-based detection
func TestSyntaxAutoDetectionShebang(t *testing.T) {
	screen := NewFakeScreen(80, 24)
	editor := newTestEditorWithContent("#!/bin/bash\necho hello\n", "myscript") // No extension

	cfg := config.Default()
	cfg.Editor.SyntaxHighlighting = true
	cfg.UI.Language = "" // Auto-detect

	ui := New(screen, editor, cfg)

	lang := ui.GetCurrentLanguage()
	if lang != "Bash" {
		t.Errorf("Expected language 'Bash' for shebang script, got %q", lang)
	}
}

// TestSyntaxHighlightingTokenization tests that tokens are actually generated
func TestSyntaxHighlightingTokenization(t *testing.T) {
	screen := NewFakeScreen(80, 24)
	editor := newTestEditorWithContent("func main() {}\n", "test.go")

	cfg := config.Default()
	cfg.Editor.SyntaxHighlighting = true
	cfg.UI.Language = "" // Auto-detect

	ui := New(screen, editor, cfg)

	if ui.syntaxCache == nil {
		t.Fatal("syntaxCache is nil, cannot test tokenization")
	}

	// Get tokens for the first line
	line := editor.Lines()[0]
	tokens := ui.syntaxCache.GetTokens(0, line)

	if len(tokens) == 0 {
		t.Error("Expected tokens for Go code 'func main() {}', got none")
	}

	// Check that 'func' is recognized as a keyword
	funcToken := syntax.GetTokenAt(tokens, 0) // 'f' of 'func'
	if funcToken != syntax.TokenKeyword {
		t.Errorf("Expected 'func' to be TokenKeyword, got %v", funcToken)
	}
}

// TestSyntaxHighlightingWithDarkTheme tests that highlighting works with non-default theme
func TestSyntaxHighlightingWithDarkTheme(t *testing.T) {
	screen := NewFakeScreen(80, 24)
	editor := newTestEditorWithContent("func main() {}\n", "test.go")

	cfg := config.Default()
	cfg.Editor.SyntaxHighlighting = true
	cfg.UI.Theme = "dark" // Use dark theme instead of default

	ui := New(screen, editor, cfg)

	if ui.syntaxCache == nil {
		t.Fatal("syntaxCache is nil")
	}

	// Get style for 'func' keyword
	line := editor.Lines()[0]
	style := ui.getSyntaxStyle(0, 0, line) // 'f' of 'func'

	if style == nil {
		t.Error("Expected syntax style for 'func' keyword with dark theme, got nil")
	}
}

// TestSyntaxHighlightingWithDefaultTheme tests highlighting with default theme
// NOTE: This test documents a bug - syntax highlighting should work with default theme
func TestSyntaxHighlightingWithDefaultTheme(t *testing.T) {
	screen := NewFakeScreen(80, 24)
	editor := newTestEditorWithContent("func main() {}\n", "test.go")

	cfg := config.Default()
	cfg.Editor.SyntaxHighlighting = true
	cfg.UI.Theme = "default"

	ui := New(screen, editor, cfg)

	if ui.syntaxCache == nil {
		t.Fatal("syntaxCache is nil - auto-detection may have failed")
	}

	// With default theme, syntax highlighting SHOULD work
	// Currently it doesn't because of isDefaultTheme() check in getSyntaxStyle
	line := editor.Lines()[0]
	style := ui.getSyntaxStyle(0, 0, line)

	// This test will fail until we fix the isDefaultTheme check
	if style == nil {
		t.Error("Expected syntax style for 'func' keyword with default theme, got nil (BUG: syntax highlighting disabled for default theme)")
	}
}

// TestSyntaxCacheInvalidation tests that cache is invalidated on edit
func TestSyntaxCacheInvalidation(t *testing.T) {
	screen := NewFakeScreen(80, 24)
	editor := newTestEditorWithContent("func main() {}\n", "test.go")

	cfg := config.Default()
	cfg.Editor.SyntaxHighlighting = true

	ui := New(screen, editor, cfg)

	if ui.syntaxCache == nil {
		t.Fatal("syntaxCache is nil")
	}

	// Get initial tokens
	line := editor.Lines()[0]
	tokens1 := ui.syntaxCache.GetTokens(0, line)

	// Invalidate the line
	ui.InvalidateSyntaxLine(0)

	// Get tokens again (should re-tokenize)
	tokens2 := ui.syntaxCache.GetTokens(0, line)

	// Both should have tokens (cache should work after invalidation)
	if len(tokens1) == 0 || len(tokens2) == 0 {
		t.Error("Expected tokens before and after invalidation")
	}
}

// TestManualLanguageOverride tests manual language selection
func TestManualLanguageOverride(t *testing.T) {
	screen := NewFakeScreen(80, 24)
	editor := newTestEditorWithContent("print('hello')\n", "unknown.txt") // No known extension

	cfg := config.Default()
	cfg.Editor.SyntaxHighlighting = true
	cfg.UI.Language = "Python" // Manual override

	ui := New(screen, editor, cfg)

	lang := ui.GetCurrentLanguage()
	if lang != "Python" {
		t.Errorf("Expected manual language 'Python', got %q", lang)
	}
}

// TestSwitchLanguage tests switching language at runtime
func TestSwitchLanguage(t *testing.T) {
	screen := NewFakeScreen(80, 24)
	editor := newTestEditorWithContent("code\n", "test.go")

	cfg := config.Default()
	cfg.Editor.SyntaxHighlighting = true

	ui := New(screen, editor, cfg)

	// Initially should be Go
	if ui.GetCurrentLanguage() != "Go" {
		t.Errorf("Expected initial language 'Go', got %q", ui.GetCurrentLanguage())
	}

	// Switch to Python
	ui.SwitchLanguage("Python")
	if ui.GetCurrentLanguage() != "Python" {
		t.Errorf("Expected language 'Python' after switch, got %q", ui.GetCurrentLanguage())
	}

	// Switch back to Auto
	ui.SwitchLanguage("Auto")
	// Should auto-detect back to Go based on file extension
	if ui.GetCurrentLanguage() != "Go" {
		t.Errorf("Expected auto-detected 'Go' after switching to Auto, got %q", ui.GetCurrentLanguage())
	}
}

// TestDetectLanguageDirectly tests the DetectLanguage function in isolation
func TestDetectLanguageDirectly(t *testing.T) {
	tests := []struct {
		path      string
		firstLine string
		expected  string
	}{
		{"test.go", "", "Go"},
		{"script.py", "", "Python"},
		{"app.js", "", "JavaScript"},
		{"style.css", "", "CSS"},
		{"config.yaml", "", "YAML"},
		{"Makefile", "", "Makefile"},
		{"Dockerfile", "", "Dockerfile"},
		{"unknown", "#!/bin/bash", "Bash"},
		{"unknown", "#!/usr/bin/env python3", "Python"},
		{"unknown", "#!/usr/bin/env node", "JavaScript"},
		{"unknown.txt", "", ""}, // No detection possible
	}

	for _, tc := range tests {
		var firstLine []rune
		if tc.firstLine != "" {
			firstLine = []rune(tc.firstLine)
		}
		result := syntax.DetectLanguage(tc.path, firstLine)
		if result != tc.expected {
			t.Errorf("DetectLanguage(%q, %q) = %q, expected %q", tc.path, tc.firstLine, result, tc.expected)
		}
	}
}

// TestChromaHighlighterCreation tests that Chroma highlighters are created correctly
func TestChromaHighlighterCreation(t *testing.T) {
	tests := []struct {
		language string
		wantNil  bool
	}{
		{"Go", false},
		{"Python", false},
		{"JavaScript", false},
		{"NonexistentLanguage", true},
		{"", true},
	}

	for _, tc := range tests {
		h := syntax.NewChromaHighlighter(tc.language)
		if tc.wantNil && h != nil {
			t.Errorf("NewChromaHighlighter(%q) should return nil", tc.language)
		}
		if !tc.wantNil && h == nil {
			t.Errorf("NewChromaHighlighter(%q) should not return nil", tc.language)
		}
	}
}

// TestGoTokenization tests tokenization of Go code
func TestGoTokenization(t *testing.T) {
	h := syntax.NewChromaHighlighter("Go")
	if h == nil {
		t.Fatal("Failed to create Go highlighter")
	}

	line := []rune("func main() { fmt.Println(\"hello\") }")
	tokens := h.Tokenize(line)

	if len(tokens) == 0 {
		t.Fatal("Expected tokens for Go code, got none")
	}

	// Find the 'func' keyword
	funcToken := syntax.GetTokenAt(tokens, 0)
	if funcToken != syntax.TokenKeyword {
		t.Errorf("Expected 'func' to be TokenKeyword, got %v", funcToken)
	}

	// Find the string "hello"
	// The string starts around position 26 (after the opening quote)
	stringToken := syntax.GetTokenAt(tokens, 27)
	if stringToken != syntax.TokenString {
		t.Errorf("Expected string literal to be TokenString, got %v", stringToken)
	}
}

// TestPythonTokenization tests tokenization of Python code
func TestPythonTokenization(t *testing.T) {
	h := syntax.NewChromaHighlighter("Python")
	if h == nil {
		t.Fatal("Failed to create Python highlighter")
	}

	line := []rune("def hello(): # comment")
	tokens := h.Tokenize(line)

	if len(tokens) == 0 {
		t.Fatal("Expected tokens for Python code, got none")
	}

	// Find the 'def' keyword
	defToken := syntax.GetTokenAt(tokens, 0)
	if defToken != syntax.TokenKeyword {
		t.Errorf("Expected 'def' to be TokenKeyword, got %v", defToken)
	}

	// Find the comment
	commentToken := syntax.GetTokenAt(tokens, 14) // Position of '#'
	if commentToken != syntax.TokenComment {
		t.Errorf("Expected '# comment' to be TokenComment, got %v", commentToken)
	}
}
