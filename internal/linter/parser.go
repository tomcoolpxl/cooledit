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
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
)

// Parser patterns for different linter output formats.
// The filename group uses (?:[a-zA-Z]:)?[^:]+ to handle Windows paths like C:\path\to\file.go
var (
	// defaultPattern matches: file:line:col: message or file:line: message
	defaultPattern = regexp.MustCompile(`^((?:[a-zA-Z]:)?[^:]+):(\d+):(?:(\d+):)?\s*(.+)$`)

	// gccPattern matches: file:line:col: severity: message
	gccPattern = regexp.MustCompile(`^((?:[a-zA-Z]:)?[^:]+):(\d+):(\d+):\s*(error|warning|note|info):\s*(.+)$`)

	// goVetPattern matches: file:line:col: message (go vet output)
	goVetPattern = regexp.MustCompile(`^((?:[a-zA-Z]:)?[^:]+):(\d+):(\d+):\s*(.+)$`)
)

// ParseDefault parses linter output in the default format: file:line:col: message
// Lines are converted to 0-based indexing.
func ParseDefault(output string, source string) []Diagnostic {
	var diagnostics []Diagnostic
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		matches := defaultPattern.FindStringSubmatch(line)
		if matches == nil {
			continue
		}

		lineNum, err := strconv.Atoi(matches[2])
		if err != nil {
			continue
		}

		col := 0
		if matches[3] != "" {
			col, _ = strconv.Atoi(matches[3])
			if col > 0 {
				col-- // Convert to 0-based
			}
		}

		diagnostics = append(diagnostics, Diagnostic{
			Line:     lineNum - 1, // Convert to 0-based
			Col:      col,
			Severity: SeverityError, // Default to error
			Message:  strings.TrimSpace(matches[4]),
			Source:   source,
		})
	}

	return diagnostics
}

// ParseGCC parses linter output in GCC format: file:line:col: severity: message
// Lines are converted to 0-based indexing.
func ParseGCC(output string, source string) []Diagnostic {
	var diagnostics []Diagnostic
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		matches := gccPattern.FindStringSubmatch(line)
		if matches == nil {
			// Try default pattern as fallback
			defMatches := defaultPattern.FindStringSubmatch(line)
			if defMatches != nil {
				lineNum, err := strconv.Atoi(defMatches[2])
				if err != nil {
					continue
				}
				col := 0
				if defMatches[3] != "" {
					col, _ = strconv.Atoi(defMatches[3])
					if col > 0 {
						col--
					}
				}
				diagnostics = append(diagnostics, Diagnostic{
					Line:     lineNum - 1,
					Col:      col,
					Severity: SeverityError,
					Message:  strings.TrimSpace(defMatches[4]),
					Source:   source,
				})
			}
			continue
		}

		lineNum, err := strconv.Atoi(matches[2])
		if err != nil {
			continue
		}

		col, _ := strconv.Atoi(matches[3])
		if col > 0 {
			col-- // Convert to 0-based
		}

		severity := parseSeverity(matches[4])
		message := strings.TrimSpace(matches[5])

		diagnostics = append(diagnostics, Diagnostic{
			Line:     lineNum - 1, // Convert to 0-based
			Col:      col,
			Severity: severity,
			Message:  message,
			Source:   source,
		})
	}

	return diagnostics
}

// ParseGoVet parses go vet output format.
func ParseGoVet(output string) []Diagnostic {
	var diagnostics []Diagnostic
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Skip lines that start with # (package names)
		if strings.HasPrefix(line, "#") {
			continue
		}

		// Skip "vet:" prefix lines
		if strings.HasPrefix(line, "vet:") {
			continue
		}

		matches := goVetPattern.FindStringSubmatch(line)
		if matches == nil {
			continue
		}

		lineNum, err := strconv.Atoi(matches[2])
		if err != nil {
			continue
		}

		col, _ := strconv.Atoi(matches[3])
		if col > 0 {
			col-- // Convert to 0-based
		}

		diagnostics = append(diagnostics, Diagnostic{
			Line:     lineNum - 1, // Convert to 0-based
			Col:      col,
			Severity: SeverityWarning, // go vet reports warnings
			Message:  strings.TrimSpace(matches[4]),
			Source:   "go vet",
		})
	}

	return diagnostics
}

// JSONDiagnostic represents a diagnostic in JSON format.
type JSONDiagnostic struct {
	File     string `json:"file"`
	Line     int    `json:"line"`
	Column   int    `json:"column"`
	EndLine  int    `json:"endLine"`
	EndCol   int    `json:"endColumn"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
	Code     string `json:"code"`
	Source   string `json:"source"`
}

// ParseJSON parses linter output in JSON format.
// Expects an array of diagnostic objects.
func ParseJSON(data []byte, source string) []Diagnostic {
	var jsonDiags []JSONDiagnostic
	if err := json.Unmarshal(data, &jsonDiags); err != nil {
		// Try single object
		var single JSONDiagnostic
		if err := json.Unmarshal(data, &single); err != nil {
			return nil
		}
		jsonDiags = []JSONDiagnostic{single}
	}

	var diagnostics []Diagnostic
	for _, jd := range jsonDiags {
		d := Diagnostic{
			Line:     jd.Line - 1, // Convert to 0-based
			Col:      jd.Column - 1,
			EndLine:  jd.EndLine - 1,
			EndCol:   jd.EndCol - 1,
			Severity: parseSeverity(jd.Severity),
			Message:  jd.Message,
			Code:     jd.Code,
			Source:   source,
		}
		if jd.Source != "" {
			d.Source = jd.Source
		}
		if d.Col < 0 {
			d.Col = 0
		}
		if d.Line < 0 {
			d.Line = 0
		}
		diagnostics = append(diagnostics, d)
	}

	return diagnostics
}

// parseSeverity converts a severity string to a Severity value.
func parseSeverity(s string) Severity {
	switch strings.ToLower(s) {
	case "error", "fatal", "err":
		return SeverityError
	case "warning", "warn":
		return SeverityWarning
	case "info", "information", "note":
		return SeverityInfo
	case "hint", "suggestion":
		return SeverityHint
	default:
		return SeverityError
	}
}
