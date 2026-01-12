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
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
)

// Common errors.
var (
	ErrNotFound = errors.New("linter command not found")
	ErrTimeout  = errors.New("linter timed out")
)

// OutputFormat specifies how to parse linter output.
type OutputFormat int

const (
	FormatDefault OutputFormat = iota // file:line:col: message
	FormatGCC                         // file:line:col: severity: message
	FormatGoVet                       // go vet specific format
	FormatJSON                        // JSON array of diagnostics
)

// linterDef holds a linter definition with format info.
type linterDef struct {
	Config Config
	Format OutputFormat
}

// builtinLinters maps language names to their default linter configurations.
var builtinLinters = map[string]linterDef{
	"go": {
		Config: Config{Command: "go", Args: []string{"vet"}},
		Format: FormatGoVet,
	},
	"python": {
		Config: Config{Command: "ruff", Args: []string{"check", "--output-format=concise", "--exit-zero"}},
		Format: FormatDefault,
	},
	"javascript": {
		Config: Config{Command: "eslint", Args: []string{"--format=unix", "--no-error-on-unmatched-pattern"}},
		Format: FormatDefault,
	},
	"typescript": {
		Config: Config{Command: "eslint", Args: []string{"--format=unix", "--no-error-on-unmatched-pattern"}},
		Format: FormatDefault,
	},
	"rust": {
		Config: Config{Command: "cargo", Args: []string{"check", "--message-format=short"}},
		Format: FormatDefault,
	},
	"c": {
		Config: Config{Command: "gcc", Args: []string{"-fsyntax-only", "-Wall", "-Wextra"}},
		Format: FormatGCC,
	},
	"c++": {
		Config: Config{Command: "g++", Args: []string{"-fsyntax-only", "-Wall", "-Wextra"}},
		Format: FormatGCC,
	},
	"cpp": {
		Config: Config{Command: "g++", Args: []string{"-fsyntax-only", "-Wall", "-Wextra"}},
		Format: FormatGCC,
	},
	"shell": {
		Config: Config{Command: "shellcheck", Args: []string{"-f", "gcc"}},
		Format: FormatGCC,
	},
	"bash": {
		Config: Config{Command: "shellcheck", Args: []string{"-f", "gcc"}},
		Format: FormatGCC,
	},
	"yaml": {
		Config: Config{Command: "yamllint", Args: []string{"-f", "parsable"}},
		Format: FormatDefault,
	},
	"json": {
		Config: Config{Command: "jsonlint", Args: []string{"--quiet"}},
		Format: FormatDefault,
	},
}

// GetLinter returns the linter configuration for the given language.
// It first checks user config, then falls back to built-in defaults.
// Returns nil if no linter is configured for the language.
func GetLinter(language string, userConfig map[string]Config) *Config {
	lang := strings.ToLower(language)

	// Check user config first
	if userConfig != nil {
		if cfg, ok := userConfig[lang]; ok {
			return &cfg
		}
	}

	// Fall back to built-in
	if def, ok := builtinLinters[lang]; ok {
		return &def.Config
	}

	return nil
}

// getOutputFormat returns the output format for the given language.
func getOutputFormat(language string) OutputFormat {
	lang := strings.ToLower(language)
	if def, ok := builtinLinters[lang]; ok {
		return def.Format
	}
	return FormatDefault
}

// IsSupported returns true if a linter is available for the given language.
func IsSupported(language string, userConfig map[string]Config) bool {
	return GetLinter(language, userConfig) != nil
}

// SupportedLanguages returns a sorted list of languages with built-in linter support.
func SupportedLanguages() []string {
	seen := make(map[string]bool)
	var languages []string
	for lang := range builtinLinters {
		// Skip aliases
		if lang == "cpp" {
			continue
		}
		if !seen[lang] {
			seen[lang] = true
			languages = append(languages, lang)
		}
	}
	sort.Strings(languages)
	return languages
}

// Lint runs the linter on the given file content and returns diagnostics.
// If the file is unsaved (content differs from disk), it creates a temp file.
func Lint(cfg *Config, filename string, content string, language string) (*LintResult, error) {
	return LintWithTimeout(cfg, filename, content, language, DefaultTimeout)
}

// LintWithTimeout runs the linter with a custom timeout.
func LintWithTimeout(cfg *Config, filename string, content string, language string, timeout time.Duration) (*LintResult, error) {
	if cfg == nil {
		return nil, errors.New("linter config is nil")
	}

	// Create temp file for linting
	tempFile, err := createTempFile(filename, content)
	if err != nil {
		return nil, err
	}
	defer os.Remove(tempFile)

	// Build command arguments
	args := make([]string, len(cfg.Args))
	copy(args, cfg.Args)

	// Append filename to args (most linters expect file as last arg)
	args = append(args, tempFile)

	// Execute linter
	result, err := execute(cfg.Command, args, timeout)
	if err != nil {
		return result, err
	}

	// Parse output based on language format
	format := getOutputFormat(language)

	// For most linters, output is on stderr, but some use stdout
	combinedOutput := result.Stderr
	if combinedOutput == "" && len(result.Diagnostics) > 0 && result.Diagnostics[0].Message != "" {
		combinedOutput = result.Diagnostics[0].Message
	}

	var diagnostics []Diagnostic
	switch format {
	case FormatGoVet:
		diagnostics = ParseGoVet(combinedOutput)
	case FormatGCC:
		diagnostics = ParseGCC(combinedOutput, cfg.Command)
	case FormatJSON:
		diagnostics = ParseJSON([]byte(combinedOutput), cfg.Command)
	default:
		diagnostics = ParseDefault(combinedOutput, cfg.Command)
	}

	result.Diagnostics = diagnostics
	return result, nil
}

// createTempFile creates a temporary file with the given content.
// The file extension is preserved from the original filename.
func createTempFile(filename string, content string) (string, error) {
	ext := filepath.Ext(filename)
	if ext == "" {
		ext = ".txt"
	}

	// Create temp file in system temp directory
	tempDir := os.TempDir()
	tempFile, err := os.CreateTemp(tempDir, "cooledit-lint-*"+ext)
	if err != nil {
		return "", err
	}

	_, err = tempFile.WriteString(content)
	if err != nil {
		tempFile.Close()
		os.Remove(tempFile.Name())
		return "", err
	}

	err = tempFile.Close()
	if err != nil {
		os.Remove(tempFile.Name())
		return "", err
	}

	return tempFile.Name(), nil
}

// execute runs a command with timeout and returns the result.
func execute(command string, args []string, timeout time.Duration) (*LintResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, command, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	result := &LintResult{
		Stderr: stderr.String(),
	}

	// Store stdout in a temporary diagnostic for parsing
	if stdout.Len() > 0 {
		result.Diagnostics = []Diagnostic{{Message: stdout.String()}}
	}

	if ctx.Err() == context.DeadlineExceeded {
		return result, ErrTimeout
	}

	if err != nil {
		// Check if command was not found
		if execErr, ok := err.(*exec.Error); ok {
			if execErr.Err == exec.ErrNotFound {
				return result, ErrNotFound
			}
		}

		// For linters, non-zero exit code usually means issues found, not an error
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
			// This is normal for linters - they return non-zero when issues are found
			return result, nil
		}

		// Check for "not found" on Windows
		if runtime.GOOS == "windows" && strings.Contains(err.Error(), "executable file not found") {
			return result, ErrNotFound
		}

		return result, err
	}

	return result, nil
}
