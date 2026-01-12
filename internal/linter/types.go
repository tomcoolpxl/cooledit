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

import "time"

// DefaultTimeout is the default timeout for linter execution.
const DefaultTimeout = 10 * time.Second

// Severity represents the diagnostic severity level.
type Severity int

const (
	SeverityError Severity = iota
	SeverityWarning
	SeverityInfo
	SeverityHint
)

// String returns a string representation of the severity.
func (s Severity) String() string {
	switch s {
	case SeverityError:
		return "error"
	case SeverityWarning:
		return "warning"
	case SeverityInfo:
		return "info"
	case SeverityHint:
		return "hint"
	default:
		return "unknown"
	}
}

// Symbol returns a single-character symbol for the severity.
func (s Severity) Symbol() rune {
	switch s {
	case SeverityError:
		return '✗'
	case SeverityWarning:
		return '⚠'
	case SeverityInfo:
		return 'ℹ'
	case SeverityHint:
		return '•'
	default:
		return '?'
	}
}

// ASCIISymbol returns an ASCII fallback symbol for limited terminals.
func (s Severity) ASCIISymbol() rune {
	switch s {
	case SeverityError:
		return 'E'
	case SeverityWarning:
		return 'W'
	case SeverityInfo:
		return 'I'
	case SeverityHint:
		return 'H'
	default:
		return '?'
	}
}

// Diagnostic represents a single linter diagnostic.
type Diagnostic struct {
	Line     int      // 0-based line number
	Col      int      // 0-based column number (0 = unknown)
	EndLine  int      // End line for range (0 = same as Line)
	EndCol   int      // End column for range (0 = unknown)
	Severity Severity // Error, Warning, Info, Hint
	Message  string   // The diagnostic message
	Source   string   // Linter name (e.g., "go vet", "eslint")
	Code     string   // Error code (e.g., "E501", "unused-variable")
}

// Config holds the configuration for a linter.
type Config struct {
	Command string   `toml:"command"`
	Args    []string `toml:"args"`
}

// LintResult holds the result of running a linter.
type LintResult struct {
	Diagnostics []Diagnostic
	Stderr      string
	ExitCode    int
}

// DiagnosticsByLine returns a map of line numbers to diagnostics.
func (r *LintResult) DiagnosticsByLine() map[int][]Diagnostic {
	result := make(map[int][]Diagnostic)
	for _, d := range r.Diagnostics {
		result[d.Line] = append(result[d.Line], d)
	}
	return result
}

// HighestSeverityForLine returns the highest severity diagnostic for a given line.
func (r *LintResult) HighestSeverityForLine(line int) *Diagnostic {
	var highest *Diagnostic
	for i := range r.Diagnostics {
		d := &r.Diagnostics[i]
		if d.Line == line {
			if highest == nil || d.Severity < highest.Severity {
				highest = d
			}
		}
	}
	return highest
}
