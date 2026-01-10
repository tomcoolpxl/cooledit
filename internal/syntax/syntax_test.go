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

package syntax

import (
	"testing"
)

// TestTokenTypeString tests the String method of TokenType
func TestTokenTypeString(t *testing.T) {
	tests := []struct {
		token    TokenType
		expected string
	}{
		{TokenNone, "None"},
		{TokenKeyword, "Keyword"},
		{TokenString, "String"},
		{TokenComment, "Comment"},
		{TokenNumber, "Number"},
		{TokenOperator, "Operator"},
		{TokenFunction, "Function"},
		{TokenType_, "Type"},
		{TokenVariable, "Variable"},
		{TokenConstant, "Constant"},
		{TokenPunctuation, "Punctuation"},
		{TokenPreproc, "Preproc"},
		{TokenBuiltin, "Builtin"},
		{TokenType(100), "Unknown"},
	}

	for _, tc := range tests {
		result := tc.token.String()
		if result != tc.expected {
			t.Errorf("TokenType(%d).String() = %q, expected %q", tc.token, result, tc.expected)
		}
	}
}

// TestGetTokenAt tests the GetTokenAt function
func TestGetTokenAt(t *testing.T) {
	tokens := []Token{
		{Type: TokenKeyword, Start: 0, End: 4},   // "func"
		{Type: TokenFunction, Start: 5, End: 9},  // "main"
		{Type: TokenPunctuation, Start: 9, End: 11}, // "()"
	}

	tests := []struct {
		runeIdx  int
		expected TokenType
	}{
		{0, TokenKeyword},
		{1, TokenKeyword},
		{3, TokenKeyword},
		{4, TokenNone}, // Between tokens
		{5, TokenFunction},
		{8, TokenFunction},
		{9, TokenPunctuation},
		{10, TokenPunctuation},
		{11, TokenNone}, // After all tokens
		{100, TokenNone},
	}

	for _, tc := range tests {
		result := GetTokenAt(tokens, tc.runeIdx)
		if result != tc.expected {
			t.Errorf("GetTokenAt(tokens, %d) = %v, expected %v", tc.runeIdx, result, tc.expected)
		}
	}
}

// TestGetTokenAtEmptyTokens tests GetTokenAt with empty tokens
func TestGetTokenAtEmptyTokens(t *testing.T) {
	result := GetTokenAt(nil, 0)
	if result != TokenNone {
		t.Errorf("GetTokenAt(nil, 0) = %v, expected TokenNone", result)
	}

	result = GetTokenAt([]Token{}, 5)
	if result != TokenNone {
		t.Errorf("GetTokenAt([], 5) = %v, expected TokenNone", result)
	}
}

// TestDetectLanguageByExtension tests language detection by file extension
func TestDetectLanguageByExtension(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"main.go", "Go"},
		{"script.py", "Python"},
		{"app.js", "JavaScript"},
		{"index.html", "HTML"},
		{"style.css", "CSS"},
		{"config.yaml", "YAML"},
		{"data.json", "JSON"},
		{"Makefile", "Makefile"},
		{"Dockerfile", "Dockerfile"},
		{"script.sh", "Bash"},
		{"script.ps1", "PowerShell"},
		{"file.bat", "Batch"},
		{"config.toml", "TOML"},
		{"settings.ini", "INI"},
		{"main.rs", "Rust"},
		{"App.java", "Java"},
		{"program.c", "C"},
		{"module.cpp", "C++"},
		{"main.ts", "TypeScript"},
		{"unknown.xyz", ""},
	}

	for _, tc := range tests {
		result := DetectLanguage(tc.path, nil)
		if result != tc.expected {
			t.Errorf("DetectLanguage(%q, nil) = %q, expected %q", tc.path, result, tc.expected)
		}
	}
}

// TestDetectLanguageByShebang tests language detection by shebang line
func TestDetectLanguageByShebang(t *testing.T) {
	tests := []struct {
		firstLine []rune
		expected  string
	}{
		{[]rune("#!/bin/bash"), "Bash"},
		{[]rune("#!/usr/bin/env python"), "Python"},
		{[]rune("#!/usr/bin/env python3"), "Python"},
		{[]rune("#!/usr/bin/perl"), "Perl"},
		{[]rune("#!/usr/bin/env ruby"), "Ruby"},
		{[]rune("#!/bin/sh"), "Bash"},
		{[]rune("#!/usr/bin/env node"), "JavaScript"},
		{[]rune("# Not a shebang"), ""},
		{[]rune(""), ""},
		{nil, ""},
	}

	for _, tc := range tests {
		result := DetectLanguage("", tc.firstLine)
		if result != tc.expected {
			t.Errorf("DetectLanguage(\"\", %q) = %q, expected %q", string(tc.firstLine), result, tc.expected)
		}
	}
}

// TestDetectLanguageExtensionPrecedence tests that extension takes precedence over shebang
func TestDetectLanguageExtensionPrecedence(t *testing.T) {
	// Even with bash shebang, .py extension should detect as Python
	result := DetectLanguage("script.py", []rune("#!/bin/bash"))
	if result != "Python" {
		t.Errorf("DetectLanguage(\"script.py\", shebang) = %q, expected \"Python\"", result)
	}
}

// TestGetLanguageList tests that GetLanguageList returns a sorted list
func TestGetLanguageList(t *testing.T) {
	list := GetLanguageList()
	if len(list) == 0 {
		t.Error("GetLanguageList() returned empty list")
	}

	// Check that list is sorted
	for i := 1; i < len(list); i++ {
		if list[i] < list[i-1] {
			t.Errorf("GetLanguageList() is not sorted: %q < %q", list[i], list[i-1])
		}
	}
}

// TestGetLanguage tests GetLanguage function
func TestGetLanguage(t *testing.T) {
	lang := GetLanguage("Go")
	if lang == nil {
		t.Fatal("GetLanguage(\"Go\") returned nil")
	}
	if lang.Name != "Go" {
		t.Errorf("GetLanguage(\"Go\").Name = %q, expected \"Go\"", lang.Name)
	}
	if lang.ChromaName != "go" {
		t.Errorf("GetLanguage(\"Go\").ChromaName = %q, expected \"go\"", lang.ChromaName)
	}

	// Case insensitive
	lang = GetLanguage("go")
	if lang == nil {
		t.Fatal("GetLanguage(\"go\") returned nil (case insensitive)")
	}

	// Unknown language
	lang = GetLanguage("NonexistentLanguage")
	if lang != nil {
		t.Error("GetLanguage(\"NonexistentLanguage\") should return nil")
	}
}

// TestGetChromaName tests GetChromaName function
func TestGetChromaName(t *testing.T) {
	tests := []struct {
		langName string
		expected string
	}{
		{"Go", "go"},
		{"Python", "python"},
		{"JavaScript", "javascript"},
		{"Unknown", ""},
	}

	for _, tc := range tests {
		result := GetChromaName(tc.langName)
		if result != tc.expected {
			t.Errorf("GetChromaName(%q) = %q, expected %q", tc.langName, result, tc.expected)
		}
	}
}

// TestNewChromaHighlighter tests creating a new Chroma highlighter
func TestNewChromaHighlighter(t *testing.T) {
	// Valid language
	h := NewChromaHighlighter("Go")
	if h == nil {
		t.Fatal("NewChromaHighlighter(\"Go\") returned nil")
	}
	if h.Language() != "Go" {
		t.Errorf("NewChromaHighlighter(\"Go\").Language() = %q, expected \"Go\"", h.Language())
	}

	// Unknown language
	h = NewChromaHighlighter("NonexistentLanguage")
	if h != nil {
		t.Error("NewChromaHighlighter(\"NonexistentLanguage\") should return nil")
	}
}

// TestChromaHighlighterTokenize tests tokenization of Go code
func TestChromaHighlighterTokenize(t *testing.T) {
	h := NewChromaHighlighter("Go")
	if h == nil {
		t.Fatal("NewChromaHighlighter(\"Go\") returned nil")
	}

	// Test tokenizing a simple Go line
	line := []rune("func main() {")
	tokens := h.Tokenize(line)
	if len(tokens) == 0 {
		t.Error("Tokenize returned no tokens for Go code")
	}

	// Check that "func" is a keyword
	funcToken := GetTokenAt(tokens, 0)
	if funcToken != TokenKeyword {
		t.Errorf("Expected \"func\" to be TokenKeyword, got %v", funcToken)
	}
}

// TestChromaHighlighterTokenizeEmpty tests tokenization of empty line
func TestChromaHighlighterTokenizeEmpty(t *testing.T) {
	h := NewChromaHighlighter("Go")
	if h == nil {
		t.Fatal("NewChromaHighlighter(\"Go\") returned nil")
	}

	tokens := h.Tokenize(nil)
	if tokens != nil {
		t.Error("Tokenize(nil) should return nil")
	}

	tokens = h.Tokenize([]rune{})
	if tokens != nil {
		t.Error("Tokenize([]) should return nil")
	}
}

// TestNewLineCache tests creating a line cache
func TestNewLineCache(t *testing.T) {
	// Valid language
	cache := NewLineCache("Go")
	if cache == nil {
		t.Fatal("NewLineCache(\"Go\") returned nil")
	}
	if cache.Language() != "Go" {
		t.Errorf("NewLineCache(\"Go\").Language() = %q, expected \"Go\"", cache.Language())
	}

	// Unknown language
	cache = NewLineCache("NonexistentLanguage")
	if cache != nil {
		t.Error("NewLineCache(\"NonexistentLanguage\") should return nil")
	}
}

// TestLineCacheGetTokens tests the caching behavior
func TestLineCacheGetTokens(t *testing.T) {
	cache := NewLineCache("Go")
	if cache == nil {
		t.Fatal("NewLineCache(\"Go\") returned nil")
	}

	line := []rune("func main() {")

	// First call should tokenize
	tokens1 := cache.GetTokens(0, line)
	if len(tokens1) == 0 {
		t.Error("First GetTokens call returned no tokens")
	}

	// Second call with same content should return cached
	tokens2 := cache.GetTokens(0, line)
	if len(tokens2) != len(tokens1) {
		t.Error("Cached GetTokens returned different number of tokens")
	}

	// Different line should get different tokens
	line2 := []rune("var x int")
	tokens3 := cache.GetTokens(1, line2)
	if len(tokens3) == 0 {
		t.Error("GetTokens for different line returned no tokens")
	}
}

// TestLineCacheInvalidate tests cache invalidation
func TestLineCacheInvalidate(t *testing.T) {
	cache := NewLineCache("Go")
	if cache == nil {
		t.Fatal("NewLineCache(\"Go\") returned nil")
	}

	line := []rune("func main() {")
	cache.GetTokens(0, line)

	// Invalidate the line
	cache.InvalidateLine(0)

	// Get tokens again - should retokenize
	tokens := cache.GetTokens(0, line)
	if len(tokens) == 0 {
		t.Error("GetTokens after invalidation returned no tokens")
	}
}

// TestLineCacheInvalidateAll tests InvalidateAll
func TestLineCacheInvalidateAll(t *testing.T) {
	cache := NewLineCache("Go")
	if cache == nil {
		t.Fatal("NewLineCache(\"Go\") returned nil")
	}

	cache.GetTokens(0, []rune("func main() {"))
	cache.GetTokens(1, []rune("var x int"))
	cache.GetTokens(2, []rune("// comment"))

	cache.InvalidateAll()

	// Language should still be set
	if cache.Language() != "Go" {
		t.Errorf("Language changed after InvalidateAll: %q", cache.Language())
	}
}

// TestLineCacheSetLanguage tests changing the language
func TestLineCacheSetLanguage(t *testing.T) {
	cache := NewLineCache("Go")
	if cache == nil {
		t.Fatal("NewLineCache(\"Go\") returned nil")
	}

	// Cache some tokens
	cache.GetTokens(0, []rune("func main() {"))

	// Change language
	ok := cache.SetLanguage("Python")
	if !ok {
		t.Error("SetLanguage(\"Python\") returned false")
	}
	if cache.Language() != "Python" {
		t.Errorf("Language after SetLanguage = %q, expected \"Python\"", cache.Language())
	}

	// Try invalid language
	ok = cache.SetLanguage("NonexistentLanguage")
	if ok {
		t.Error("SetLanguage(\"NonexistentLanguage\") should return false")
	}
	// Language should remain Python
	if cache.Language() != "Python" {
		t.Errorf("Language should remain \"Python\" after failed SetLanguage")
	}

	// Same language should return true without change
	ok = cache.SetLanguage("Python")
	if !ok {
		t.Error("SetLanguage with same language should return true")
	}
}

// TestLineCacheNil tests that nil cache methods don't panic
func TestLineCacheNil(t *testing.T) {
	var cache *LineCache

	// These should not panic
	tokens := cache.GetTokens(0, []rune("test"))
	if tokens != nil {
		t.Error("nil cache GetTokens should return nil")
	}

	cache.InvalidateLine(0)
	cache.InvalidateRange(0, 10)
	cache.InvalidateFromLine(0)
	cache.InvalidateAll()

	if cache.Language() != "" {
		t.Error("nil cache Language should return empty string")
	}

	if cache.SetLanguage("Go") {
		t.Error("nil cache SetLanguage should return false")
	}
}

// TestLanguagesExist tests that commonly expected languages are defined
func TestLanguagesExist(t *testing.T) {
	expectedLanguages := []string{
		"Go", "Python", "JavaScript", "TypeScript", "Java",
		"C", "C++", "Rust", "Ruby", "PHP",
		"Bash", "PowerShell", "Batch",
		"YAML", "JSON", "TOML", "INI", "XML",
		"HTML", "CSS", "SQL",
		"Makefile", "Dockerfile",
		"Markdown",
	}

	for _, name := range expectedLanguages {
		lang := GetLanguage(name)
		if lang == nil {
			t.Errorf("Expected language %q not found", name)
		}
	}
}

// TestLanguageExtensions tests that key extensions are mapped correctly
func TestLanguageExtensions(t *testing.T) {
	tests := []struct {
		ext      string
		expected string
	}{
		{".go", "Go"},
		{".py", "Python"},
		{".js", "JavaScript"},
		{".ts", "TypeScript"},
		{".java", "Java"},
		{".c", "C"},
		{".cpp", "C++"},
		{".rs", "Rust"},
		{".rb", "Ruby"},
		{".php", "PHP"},
		{".sh", "Bash"},
		{".ps1", "PowerShell"},
		{".bat", "Batch"},
		{".yaml", "YAML"},
		{".json", "JSON"},
		{".toml", "TOML"},
		{".html", "HTML"},
		{".css", "CSS"},
		{".sql", "SQL"},
		{".md", "Markdown"},
	}

	for _, tc := range tests {
		result := DetectLanguage("file"+tc.ext, nil)
		if result != tc.expected {
			t.Errorf("DetectLanguage(\"file%s\", nil) = %q, expected %q", tc.ext, result, tc.expected)
		}
	}
}
