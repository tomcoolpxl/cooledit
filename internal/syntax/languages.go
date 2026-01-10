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
	"path/filepath"
	"sort"
	"strings"
)

// Language defines a programming language for syntax highlighting.
type Language struct {
	Name       string   // Display name (e.g., "Go", "Python")
	Extensions []string // File extensions (e.g., ".go", ".py")
	Shebangs   []string // Shebang interpreters (e.g., "python", "bash")
	ChromaName string   // Chroma lexer name
}

// Languages is the curated list of supported languages for syntax highlighting.
// This covers ~40 languages commonly used in software development and system administration.
var Languages = []Language{
	// Programming Languages
	{"Go", []string{".go"}, nil, "go"},
	{"Python", []string{".py", ".pyw", ".pyi"}, []string{"python", "python3", "python2"}, "python"},
	{"JavaScript", []string{".js", ".mjs", ".cjs"}, []string{"node", "nodejs"}, "javascript"},
	{"TypeScript", []string{".ts", ".tsx", ".mts", ".cts"}, nil, "typescript"},
	{"Java", []string{".java"}, nil, "java"},
	{"C", []string{".c", ".h"}, nil, "c"},
	{"C++", []string{".cpp", ".cc", ".cxx", ".hpp", ".hxx", ".hh"}, nil, "cpp"},
	{"C#", []string{".cs"}, nil, "csharp"},
	{"Rust", []string{".rs"}, nil, "rust"},
	{"PHP", []string{".php", ".php3", ".php4", ".php5"}, []string{"php"}, "php"},
	{"Ruby", []string{".rb", ".rake", ".gemspec"}, []string{"ruby"}, "ruby"},
	{"Swift", []string{".swift"}, nil, "swift"},
	{"Kotlin", []string{".kt", ".kts"}, nil, "kotlin"},
	{"Scala", []string{".scala", ".sc"}, nil, "scala"},
	{"Perl", []string{".pl", ".pm", ".t"}, []string{"perl"}, "perl"},
	{"Lua", []string{".lua"}, []string{"lua"}, "lua"},
	{"R", []string{".r", ".R"}, []string{"Rscript"}, "r"},
	{"Dart", []string{".dart"}, nil, "dart"},
	{"Elixir", []string{".ex", ".exs"}, []string{"elixir"}, "elixir"},
	{"Haskell", []string{".hs", ".lhs"}, nil, "haskell"},

	// Web Technologies
	{"HTML", []string{".html", ".htm", ".xhtml"}, nil, "html"},
	{"CSS", []string{".css"}, nil, "css"},
	{"SCSS", []string{".scss"}, nil, "scss"},
	{"LESS", []string{".less"}, nil, "less"},
	{"SQL", []string{".sql"}, nil, "sql"},
	{"GraphQL", []string{".graphql", ".gql"}, nil, "graphql"},

	// Shell and Sysadmin
	{"Bash", []string{".sh", ".bash", ".zsh"}, []string{"bash", "sh", "zsh", "ash", "dash"}, "bash"},
	{"PowerShell", []string{".ps1", ".psm1", ".psd1"}, []string{"pwsh", "powershell"}, "powershell"},
	{"Batch", []string{".bat", ".cmd"}, nil, "batchfile"},
	{"Fish", []string{".fish"}, []string{"fish"}, "fish"},

	// Configuration Files
	{"YAML", []string{".yaml", ".yml"}, nil, "yaml"},
	{"JSON", []string{".json", ".jsonc"}, nil, "json"},
	{"TOML", []string{".toml"}, nil, "toml"},
	{"INI", []string{".ini", ".cfg", ".conf"}, nil, "ini"},
	{"XML", []string{".xml", ".xsl", ".xslt", ".xsd", ".svg"}, nil, "xml"},
	{"Properties", []string{".properties"}, nil, "properties"},
	{"Registry", []string{".reg"}, nil, "registry"},
	{"Nginx", []string{".nginx"}, nil, "nginx"},
	{"Apache", []string{".htaccess"}, nil, "apacheconf"},

	// Cloud and DevOps
	{"Terraform", []string{".tf", ".tfvars"}, nil, "terraform"},
	{"HCL", []string{".hcl"}, nil, "hcl"},
	{"Dockerfile", []string{"Dockerfile", ".dockerfile"}, nil, "docker"},
	{"Docker Compose", []string{"docker-compose.yml", "docker-compose.yaml", "compose.yml", "compose.yaml"}, nil, "yaml"},

	// Build and Project Files
	{"Makefile", []string{"Makefile", "makefile", "GNUmakefile"}, nil, "makefile"},
	{"CMake", []string{"CMakeLists.txt", ".cmake"}, nil, "cmake"},
	{"Gradle", []string{".gradle"}, nil, "groovy"},

	// Documentation
	{"Markdown", []string{".md", ".markdown", ".mdown"}, nil, "markdown"},
	{"reStructuredText", []string{".rst"}, nil, "rst"},
	{"LaTeX", []string{".tex", ".latex"}, nil, "latex"},

	// Data Formats
	{"CSV", []string{".csv", ".tsv"}, nil, "text"},
	{"Diff", []string{".diff", ".patch"}, nil, "diff"},
	{"Protocol Buffers", []string{".proto"}, nil, "protobuf"},
}

// languageByExt maps file extensions to language names for fast lookup.
var languageByExt map[string]string

// languageByName maps language names (lowercase) to Language structs.
var languageByName map[string]*Language

// init builds the lookup maps from the Languages list.
func init() {
	languageByExt = make(map[string]string)
	languageByName = make(map[string]*Language)

	for i := range Languages {
		lang := &Languages[i]
		languageByName[strings.ToLower(lang.Name)] = lang

		for _, ext := range lang.Extensions {
			// Store with lowercase extension for case-insensitive matching
			languageByExt[strings.ToLower(ext)] = lang.Name
		}
	}
}

// DetectLanguage determines the programming language from a file path and content.
// It first checks the file extension, then looks for shebang lines in the content.
// Returns an empty string if the language cannot be determined.
func DetectLanguage(path string, firstLine []rune) string {
	// Try file extension first
	if path != "" {
		// Check exact filename match (for Makefile, Dockerfile, etc.)
		base := filepath.Base(path)
		if lang, ok := languageByExt[base]; ok {
			return lang
		}
		if lang, ok := languageByExt[strings.ToLower(base)]; ok {
			return lang
		}

		// Check file extension
		ext := strings.ToLower(filepath.Ext(path))
		if lang, ok := languageByExt[ext]; ok {
			return lang
		}
	}

	// Try shebang detection
	if len(firstLine) > 2 && firstLine[0] == '#' && firstLine[1] == '!' {
		shebang := string(firstLine)
		return detectFromShebang(shebang)
	}

	return ""
}

// detectFromShebang extracts the language from a shebang line.
// Handles both direct interpreters (#!/bin/bash) and env (#!/usr/bin/env python).
func detectFromShebang(shebang string) string {
	// Remove #! prefix
	shebang = strings.TrimPrefix(shebang, "#!")
	shebang = strings.TrimSpace(shebang)

	// Split into parts
	parts := strings.Fields(shebang)
	if len(parts) == 0 {
		return ""
	}

	// Get the interpreter name
	interpreter := filepath.Base(parts[0])

	// Handle /usr/bin/env
	if interpreter == "env" && len(parts) > 1 {
		interpreter = parts[1]
	}

	// Remove version suffixes (python3.9 -> python3)
	for i := 0; i < len(interpreter); i++ {
		if interpreter[i] == '.' {
			interpreter = interpreter[:i]
			break
		}
	}

	// Look up interpreter in language shebangs
	interpreterLower := strings.ToLower(interpreter)
	for i := range Languages {
		for _, s := range Languages[i].Shebangs {
			if strings.ToLower(s) == interpreterLower {
				return Languages[i].Name
			}
		}
	}

	return ""
}

// GetLanguageList returns a sorted list of available language names.
// This is used to populate the language selection menu.
func GetLanguageList() []string {
	names := make([]string, len(Languages))
	for i, lang := range Languages {
		names[i] = lang.Name
	}
	sort.Strings(names)
	return names
}

// GetLanguage returns the Language struct for a given language name.
// Returns nil if the language is not found.
func GetLanguage(name string) *Language {
	return languageByName[strings.ToLower(name)]
}

// GetChromaName returns the Chroma lexer name for a given language name.
// Returns empty string if the language is not found.
func GetChromaName(languageName string) string {
	lang := GetLanguage(languageName)
	if lang != nil {
		return lang.ChromaName
	}
	return ""
}
