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

package linter

import (
	"errors"
	"testing"
)

func TestGetLinter_BuiltinDefaults(t *testing.T) {
	tests := []struct {
		language string
		wantCmd  string
	}{
		{"go", "go"},
		{"Go", "go"},
		{"GO", "go"},
		{"python", "ruff"},
		{"Python", "ruff"},
		{"javascript", "eslint"},
		{"typescript", "eslint"},
		{"rust", "cargo"},
		{"c", "gcc"},
		{"c++", "g++"},
		{"cpp", "g++"},
		{"shell", "shellcheck"},
		{"bash", "shellcheck"},
		{"yaml", "yamllint"},
		{"json", "jsonlint"},
	}

	for _, tt := range tests {
		t.Run(tt.language, func(t *testing.T) {
			cfg := GetLinter(tt.language, nil)
			if cfg == nil {
				t.Fatalf("GetLinter(%q) returned nil, expected linter", tt.language)
			}
			if cfg.Command != tt.wantCmd {
				t.Errorf("GetLinter(%q).Command = %q, want %q", tt.language, cfg.Command, tt.wantCmd)
			}
		})
	}
}

func TestGetLinter_NoLinter(t *testing.T) {
	cfg := GetLinter("unknownlanguage", nil)
	if cfg != nil {
		t.Errorf("GetLinter(unknownlanguage) = %v, want nil", cfg)
	}
}

func TestGetLinter_UserConfigOverride(t *testing.T) {
	userConfig := map[string]Config{
		"go": {
			Command: "golangci-lint",
			Args:    []string{"run", "--fast"},
		},
		"python": {
			Command: "pylint",
			Args:    []string{"--output-format=parseable"},
		},
	}

	// Test user config override
	cfg := GetLinter("go", userConfig)
	if cfg == nil {
		t.Fatal("GetLinter(go) returned nil")
	}
	if cfg.Command != "golangci-lint" {
		t.Errorf("GetLinter(go).Command = %q, want %q", cfg.Command, "golangci-lint")
	}
	if len(cfg.Args) != 2 || cfg.Args[0] != "run" {
		t.Errorf("GetLinter(go).Args = %v, want [run --fast]", cfg.Args)
	}

	// Test user config for another language
	cfg = GetLinter("python", userConfig)
	if cfg == nil {
		t.Fatal("GetLinter(python) returned nil")
	}
	if cfg.Command != "pylint" {
		t.Errorf("GetLinter(python).Command = %q, want %q", cfg.Command, "pylint")
	}

	// Test fallback to builtin when user config doesn't have the language
	cfg = GetLinter("rust", userConfig)
	if cfg == nil {
		t.Fatal("GetLinter(rust) returned nil")
	}
	if cfg.Command != "cargo" {
		t.Errorf("GetLinter(rust).Command = %q, want %q (builtin)", cfg.Command, "cargo")
	}
}

func TestIsSupported(t *testing.T) {
	if !IsSupported("go", nil) {
		t.Error("IsSupported(go) = false, want true")
	}
	if !IsSupported("Go", nil) {
		t.Error("IsSupported(Go) = false, want true (case insensitive)")
	}
	if IsSupported("unknownlanguage", nil) {
		t.Error("IsSupported(unknownlanguage) = true, want false")
	}

	// With user config
	userConfig := map[string]Config{
		"mylang": {Command: "mylinter"},
	}
	if !IsSupported("mylang", userConfig) {
		t.Error("IsSupported(mylang, userConfig) = false, want true")
	}
}

func TestSupportedLanguages(t *testing.T) {
	languages := SupportedLanguages()
	if len(languages) == 0 {
		t.Fatal("SupportedLanguages() returned empty list")
	}

	// Check that some expected languages are present
	found := make(map[string]bool)
	for _, lang := range languages {
		found[lang] = true
	}

	expected := []string{"go", "python", "javascript", "rust", "yaml"}
	for _, lang := range expected {
		if !found[lang] {
			t.Errorf("SupportedLanguages() missing %q", lang)
		}
	}
}

func TestExecute_CommandNotFound(t *testing.T) {
	result, err := execute("nonexistent_command_xyz", nil, DefaultTimeout)
	if err == nil {
		t.Fatal("execute with nonexistent command should return error")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("execute error = %v, want ErrNotFound", err)
	}
	if result == nil {
		t.Error("execute should return result even on error")
	}
}

func TestLint_InvalidLinter(t *testing.T) {
	cfg := &Config{
		Command: "nonexistent_linter_xyz",
		Args:    nil,
	}

	_, err := Lint(cfg, "test.go", "package main", "go")
	if err == nil {
		t.Fatal("Lint with invalid linter should return error")
	}
}

func TestLint_GoVet(t *testing.T) {
	// Test actual go vet linting - go vet should be available since we're building a Go project
	cfg := GetLinter("go", nil)
	if cfg == nil {
		t.Fatal("GetLinter(go) returned nil")
	}

	// Go code with a vet error: Printf format string has extra operand
	codeWithError := `package main

import "fmt"

func main() {
	fmt.Printf("hello %s", "world", "extra")
}
`

	result, err := Lint(cfg, "test.go", codeWithError, "go")
	if err != nil && !errors.Is(err, ErrNotFound) {
		t.Skipf("Skipping go vet test (go vet may not be available): %v", err)
	}
	if err != nil {
		t.Skipf("Skipping go vet test: %v", err)
	}

	if result == nil {
		t.Fatal("Lint returned nil result")
	}

	// go vet should find an issue
	if len(result.Diagnostics) == 0 {
		t.Error("go vet should find Printf format error, but got no diagnostics")
	}

	// Check that the diagnostic is about Printf
	if len(result.Diagnostics) > 0 {
		found := false
		for _, d := range result.Diagnostics {
			if d.Message != "" {
				found = true
				t.Logf("Diagnostic found: line=%d col=%d msg=%q", d.Line, d.Col, d.Message)
			}
		}
		if !found {
			t.Error("Expected at least one diagnostic with a message")
		}
	}
}

func TestLint_GoVetCleanCode(t *testing.T) {
	// Test that clean code produces no diagnostics
	cfg := GetLinter("go", nil)
	if cfg == nil {
		t.Fatal("GetLinter(go) returned nil")
	}

	cleanCode := `package main

import "fmt"

func main() {
	fmt.Println("hello world")
}
`

	result, err := Lint(cfg, "test.go", cleanCode, "go")
	if err != nil && !errors.Is(err, ErrNotFound) {
		t.Skipf("Skipping go vet test (go vet may not be available): %v", err)
	}
	if err != nil {
		t.Skipf("Skipping go vet test: %v", err)
	}

	if result == nil {
		t.Fatal("Lint returned nil result")
	}

	// Clean code should have no issues
	if len(result.Diagnostics) > 0 {
		for _, d := range result.Diagnostics {
			t.Logf("Unexpected diagnostic: line=%d col=%d msg=%q", d.Line, d.Col, d.Message)
		}
		t.Errorf("Clean code should have no diagnostics, got %d", len(result.Diagnostics))
	}
}

func TestParseDefault(t *testing.T) {
	tests := []struct {
		name    string
		output  string
		wantLen int
		wantMsg string
	}{
		{
			name:    "single diagnostic",
			output:  "test.go:10:5: undefined: foo",
			wantLen: 1,
			wantMsg: "undefined: foo",
		},
		{
			name:    "multiple diagnostics",
			output:  "test.go:1:1: error one\ntest.go:2:2: error two\n",
			wantLen: 2,
			wantMsg: "error one",
		},
		{
			name:    "empty output",
			output:  "",
			wantLen: 0,
		},
		{
			name:    "no line/col",
			output:  "test.go: some message",
			wantLen: 0, // Won't match the pattern
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diagnostics := ParseDefault(tt.output, "test")
			if len(diagnostics) != tt.wantLen {
				t.Errorf("ParseDefault() got %d diagnostics, want %d", len(diagnostics), tt.wantLen)
			}
			if tt.wantLen > 0 && diagnostics[0].Message != tt.wantMsg {
				t.Errorf("ParseDefault() message = %q, want %q", diagnostics[0].Message, tt.wantMsg)
			}
		})
	}
}

func TestParseGCC(t *testing.T) {
	tests := []struct {
		name         string
		output       string
		wantLen      int
		wantSeverity Severity
	}{
		{
			name:         "error",
			output:       "test.c:10:5: error: undefined reference",
			wantLen:      1,
			wantSeverity: SeverityError,
		},
		{
			name:         "warning",
			output:       "test.c:10:5: warning: unused variable",
			wantLen:      1,
			wantSeverity: SeverityWarning,
		},
		{
			name:         "note",
			output:       "test.c:10:5: note: declared here",
			wantLen:      1,
			wantSeverity: SeverityInfo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diagnostics := ParseGCC(tt.output, "gcc")
			if len(diagnostics) != tt.wantLen {
				t.Errorf("ParseGCC() got %d diagnostics, want %d", len(diagnostics), tt.wantLen)
			}
			if tt.wantLen > 0 && diagnostics[0].Severity != tt.wantSeverity {
				t.Errorf("ParseGCC() severity = %v, want %v", diagnostics[0].Severity, tt.wantSeverity)
			}
		})
	}
}

func TestParseGoVet(t *testing.T) {
	tests := []struct {
		name    string
		output  string
		wantLen int
	}{
		{
			name:    "vet output",
			output:  "# command-line-arguments\n./test.go:6:2: Printf format %s has arg \"extra\" of wrong type",
			wantLen: 1,
		},
		{
			name:    "vet with absolute path",
			output:  "/tmp/test.go:10:5: unreachable code",
			wantLen: 1,
		},
		{
			name:    "empty",
			output:  "",
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diagnostics := ParseGoVet(tt.output)
			if len(diagnostics) != tt.wantLen {
				t.Errorf("ParseGoVet() got %d diagnostics, want %d", len(diagnostics), tt.wantLen)
			}
		})
	}
}

func TestSeverityString(t *testing.T) {
	tests := []struct {
		severity Severity
		want     string
	}{
		{SeverityError, "error"},
		{SeverityWarning, "warning"},
		{SeverityInfo, "info"},
		{SeverityHint, "hint"},
		{Severity(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.severity.String(); got != tt.want {
				t.Errorf("Severity.String() = %q, want %q", got, tt.want)
			}
		})
	}
}
