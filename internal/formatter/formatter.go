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
	"path/filepath"
	"strings"
	"time"
)

// Config holds the configuration for a formatter
type Config struct {
	Command string   `toml:"command"`
	Args    []string `toml:"args"`
}

// builtinFormatters contains default formatters for common languages.
// Keys are lowercase language names as returned by syntax.DetectLanguage().
var builtinFormatters = map[string]Config{
	"go": {
		Command: "gofmt",
		Args:    nil,
	},
	"python": {
		Command: "black",
		Args:    []string{"-"},
	},
	"javascript": {
		Command: "prettier",
		Args:    []string{"--stdin-filepath", "file.js"},
	},
	"typescript": {
		Command: "prettier",
		Args:    []string{"--stdin-filepath", "file.ts"},
	},
	"rust": {
		Command: "rustfmt",
		Args:    nil,
	},
	"json": {
		Command: "prettier",
		Args:    []string{"--stdin-filepath", "file.json"},
	},
	"yaml": {
		Command: "prettier",
		Args:    []string{"--stdin-filepath", "file.yaml"},
	},
	"html": {
		Command: "prettier",
		Args:    []string{"--stdin-filepath", "file.html"},
	},
	"css": {
		Command: "prettier",
		Args:    []string{"--stdin-filepath", "file.css"},
	},
	"scss": {
		Command: "prettier",
		Args:    []string{"--stdin-filepath", "file.scss"},
	},
	"markdown": {
		Command: "prettier",
		Args:    []string{"--stdin-filepath", "file.md"},
	},
	"xml": {
		Command: "prettier",
		Args:    []string{"--stdin-filepath", "file.xml", "--plugin", "@prettier/plugin-xml"},
	},
	"c": {
		Command: "clang-format",
		Args:    nil,
	},
	"c++": {
		Command: "clang-format",
		Args:    nil,
	},
	"java": {
		Command: "google-java-format",
		Args:    []string{"-"},
	},
	"ruby": {
		Command: "rubocop",
		Args:    []string{"-a", "--stdin", "-"},
	},
	"php": {
		Command: "php-cs-fixer",
		Args:    []string{"fix", "--using-cache=no", "-"},
	},
	"sql": {
		Command: "sql-formatter",
		Args:    nil,
	},
	"toml": {
		Command: "taplo",
		Args:    []string{"fmt", "-"},
	},
	"shell": {
		Command: "shfmt",
		Args:    nil,
	},
	"bash": {
		Command: "shfmt",
		Args:    nil,
	},
}

// GetFormatter returns the formatter configuration for the given language.
// It first checks user config, then falls back to built-in defaults.
// Returns nil if no formatter is configured for the language.
func GetFormatter(language string, userConfig map[string]Config) *Config {
	// Normalize language name to lowercase
	lang := strings.ToLower(language)

	// Check user config first
	if userConfig != nil {
		if cfg, ok := userConfig[lang]; ok {
			return &cfg
		}
	}

	// Fall back to built-in defaults
	if cfg, ok := builtinFormatters[lang]; ok {
		return &cfg
	}

	return nil
}

// Format formats the input text using the given formatter configuration.
// The filename is used for formatters that need it (like prettier).
// Returns the formatted text or an error.
func Format(cfg *Config, input string, filename string) (string, error) {
	// Build args, replacing placeholder filename if present
	args := make([]string, len(cfg.Args))
	for i, arg := range cfg.Args {
		if strings.Contains(arg, "file.") {
			// Replace placeholder with actual filename extension
			ext := filepath.Ext(filename)
			if ext == "" {
				ext = ".txt"
			}
			args[i] = strings.Replace(arg, "file.", "stdin"+ext[:1], 1)
			// Actually, prettier needs the full placeholder format
			args[i] = arg
			if filename != "" {
				// Use actual filename for better prettier detection
				args[i] = strings.Replace(arg, "file."+strings.TrimPrefix(filepath.Ext(arg), "."), filename, 1)
			}
		} else {
			args[i] = arg
		}
	}

	result, err := Execute(cfg.Command, args, input, DefaultTimeout)
	if err != nil {
		return "", err
	}

	return result.Stdout, nil
}

// IsSupported returns true if there is a formatter available for the language.
func IsSupported(language string, userConfig map[string]Config) bool {
	return GetFormatter(language, userConfig) != nil
}

// SupportedLanguages returns a list of languages with built-in formatter support.
func SupportedLanguages() []string {
	languages := make([]string, 0, len(builtinFormatters))
	for lang := range builtinFormatters {
		languages = append(languages, lang)
	}
	return languages
}

// FormatWithTimeout formats the input text with a custom timeout.
func FormatWithTimeout(cfg *Config, input string, filename string, timeout time.Duration) (string, error) {
	args := make([]string, len(cfg.Args))
	copy(args, cfg.Args)

	result, err := Execute(cfg.Command, args, input, timeout)
	if err != nil {
		return "", err
	}

	return result.Stdout, nil
}
