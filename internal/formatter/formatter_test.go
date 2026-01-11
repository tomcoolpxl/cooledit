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

package formatter

import (
	"errors"
	"strings"
	"testing"
)

func TestGetFormatter_BuiltinDefaults(t *testing.T) {
	tests := []struct {
		language string
		wantCmd  string
	}{
		{"go", "gofmt"},
		{"Go", "gofmt"},
		{"GO", "gofmt"},
		{"python", "black"},
		{"Python", "black"},
		{"javascript", "prettier"},
		{"rust", "rustfmt"},
		{"json", "prettier"},
		{"yaml", "prettier"},
		{"html", "prettier"},
		{"css", "prettier"},
		{"c", "clang-format"},
		{"c++", "clang-format"},
		{"shell", "shfmt"},
		{"bash", "shfmt"},
	}

	for _, tt := range tests {
		t.Run(tt.language, func(t *testing.T) {
			cfg := GetFormatter(tt.language, nil)
			if cfg == nil {
				t.Fatalf("GetFormatter(%q) returned nil, expected formatter", tt.language)
			}
			if cfg.Command != tt.wantCmd {
				t.Errorf("GetFormatter(%q).Command = %q, want %q", tt.language, cfg.Command, tt.wantCmd)
			}
		})
	}
}

func TestGetFormatter_NoFormatter(t *testing.T) {
	cfg := GetFormatter("unknownlanguage", nil)
	if cfg != nil {
		t.Errorf("GetFormatter(unknownlanguage) = %v, want nil", cfg)
	}
}

func TestGetFormatter_UserConfigOverride(t *testing.T) {
	userConfig := map[string]Config{
		"go": {
			Command: "goimports",
			Args:    []string{"-local", "myproject"},
		},
		"python": {
			Command: "autopep8",
			Args:    []string{"-"},
		},
	}

	// Test user config override
	cfg := GetFormatter("go", userConfig)
	if cfg == nil {
		t.Fatal("GetFormatter(go) returned nil")
	}
	if cfg.Command != "goimports" {
		t.Errorf("GetFormatter(go).Command = %q, want %q", cfg.Command, "goimports")
	}
	if len(cfg.Args) != 2 || cfg.Args[0] != "-local" {
		t.Errorf("GetFormatter(go).Args = %v, want [-local myproject]", cfg.Args)
	}

	// Test user config for another language
	cfg = GetFormatter("python", userConfig)
	if cfg == nil {
		t.Fatal("GetFormatter(python) returned nil")
	}
	if cfg.Command != "autopep8" {
		t.Errorf("GetFormatter(python).Command = %q, want %q", cfg.Command, "autopep8")
	}

	// Test fallback to builtin when user config doesn't have the language
	cfg = GetFormatter("rust", userConfig)
	if cfg == nil {
		t.Fatal("GetFormatter(rust) returned nil")
	}
	if cfg.Command != "rustfmt" {
		t.Errorf("GetFormatter(rust).Command = %q, want %q (builtin)", cfg.Command, "rustfmt")
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
		"mylang": {Command: "myformatter"},
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

	expected := []string{"go", "python", "javascript", "rust", "json"}
	for _, lang := range expected {
		if !found[lang] {
			t.Errorf("SupportedLanguages() missing %q", lang)
		}
	}
}

func TestExecute_CommandNotFound(t *testing.T) {
	result, err := Execute("nonexistent_command_xyz", nil, "input", DefaultTimeout)
	if err == nil {
		t.Fatal("Execute with nonexistent command should return error")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Execute error = %v, want ErrNotFound", err)
	}
	if result == nil {
		t.Error("Execute should return result even on error")
	}
}

func TestExecute_Echo(t *testing.T) {
	// Use echo as a simple test command (available on most systems)
	// Note: This test may behave differently on Windows vs Unix
	result, err := Execute("echo", []string{"hello"}, "", DefaultTimeout)
	if err != nil {
		t.Skipf("Skipping echo test (command may not be available): %v", err)
	}
	if result == nil {
		t.Fatal("Execute returned nil result")
	}
	if !strings.Contains(result.Stdout, "hello") {
		t.Errorf("Execute stdout = %q, want to contain 'hello'", result.Stdout)
	}
}

func TestFormat_InvalidFormatter(t *testing.T) {
	cfg := &Config{
		Command: "nonexistent_formatter_xyz",
		Args:    nil,
	}

	_, err := Format(cfg, "test input", "test.go")
	if err == nil {
		t.Fatal("Format with invalid formatter should return error")
	}
}

func TestFormat_Gofmt(t *testing.T) {
	// Test actual gofmt formatting - gofmt should be available since we're building a Go project
	cfg := GetFormatter("go", nil)
	if cfg == nil {
		t.Fatal("GetFormatter(go) returned nil")
	}

	// Unformatted Go code (missing spaces, wrong indentation)
	unformatted := `package main

func main(){
x:=1
if x==1{
fmt.Println("hello")
}
}
`

	// Expected formatted output from gofmt
	expected := `package main

func main() {
	x := 1
	if x == 1 {
		fmt.Println("hello")
	}
}
`

	result, err := Format(cfg, unformatted, "test.go")
	if err != nil {
		t.Skipf("Skipping gofmt test (gofmt may not be available): %v", err)
	}

	if result != expected {
		t.Errorf("Format with gofmt:\ngot:\n%s\nwant:\n%s", result, expected)
	}
}

func TestFormat_GofmtPreservesCorrectCode(t *testing.T) {
	// Test that already-formatted code stays the same
	cfg := GetFormatter("go", nil)
	if cfg == nil {
		t.Fatal("GetFormatter(go) returned nil")
	}

	formatted := `package main

func main() {
	fmt.Println("hello")
}
`

	result, err := Format(cfg, formatted, "test.go")
	if err != nil {
		t.Skipf("Skipping gofmt test (gofmt may not be available): %v", err)
	}

	if result != formatted {
		t.Errorf("Format changed already-formatted code:\ngot:\n%s\nwant:\n%s", result, formatted)
	}
}

func TestFormat_GofmtSyntaxError(t *testing.T) {
	// Test that gofmt returns error for invalid Go code
	cfg := GetFormatter("go", nil)
	if cfg == nil {
		t.Fatal("GetFormatter(go) returned nil")
	}

	invalidCode := `package main

func main() {
	this is not valid go code
}
`

	_, err := Format(cfg, invalidCode, "test.go")
	if err == nil {
		t.Error("Format should return error for invalid Go code")
	}
}
