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

// Package syntax provides syntax highlighting support using the Chroma library.
// It offers language detection, tokenization, and caching for efficient
// viewport-based syntax highlighting in the terminal editor.
package syntax

// TokenType represents syntax token categories for highlighting.
// These categories map to theme colors for rendering.
type TokenType int

const (
	// TokenNone represents plain text with no special highlighting
	TokenNone TokenType = iota
	// TokenKeyword represents language keywords (if, for, func, etc.)
	TokenKeyword
	// TokenString represents string literals
	TokenString
	// TokenComment represents comments (single-line and multi-line)
	TokenComment
	// TokenNumber represents numeric literals
	TokenNumber
	// TokenOperator represents operators (+, -, *, /, etc.)
	TokenOperator
	// TokenFunction represents function names
	TokenFunction
	// TokenType represents type names
	TokenType_
	// TokenVariable represents variable names
	TokenVariable
	// TokenConstant represents constants
	TokenConstant
	// TokenPunctuation represents punctuation (brackets, braces, etc.)
	TokenPunctuation
	// TokenPreproc represents preprocessor directives
	TokenPreproc
	// TokenBuiltin represents built-in functions and types
	TokenBuiltin
)

// String returns a human-readable name for the token type
func (t TokenType) String() string {
	switch t {
	case TokenNone:
		return "None"
	case TokenKeyword:
		return "Keyword"
	case TokenString:
		return "String"
	case TokenComment:
		return "Comment"
	case TokenNumber:
		return "Number"
	case TokenOperator:
		return "Operator"
	case TokenFunction:
		return "Function"
	case TokenType_:
		return "Type"
	case TokenVariable:
		return "Variable"
	case TokenConstant:
		return "Constant"
	case TokenPunctuation:
		return "Punctuation"
	case TokenPreproc:
		return "Preproc"
	case TokenBuiltin:
		return "Builtin"
	default:
		return "Unknown"
	}
}

// Token represents a syntax token with its type and position within a line.
// Start and End are rune indices (not byte indices).
type Token struct {
	Type  TokenType
	Start int // Rune index in line (inclusive)
	End   int // Rune index in line (exclusive)
}

// Highlighter interface for language-specific highlighting.
// Implementations tokenize lines of text for syntax highlighting.
type Highlighter interface {
	// Tokenize returns syntax tokens for a single line of text.
	// The tokens are sorted by Start position and should not overlap.
	Tokenize(line []rune) []Token

	// Language returns the name of the language this highlighter handles.
	Language() string
}

// GetTokenAt returns the token type at a specific rune index within a line.
// If no token covers that position, returns TokenNone.
func GetTokenAt(tokens []Token, runeIdx int) TokenType {
	for _, t := range tokens {
		if runeIdx >= t.Start && runeIdx < t.End {
			return t.Type
		}
		// Tokens are sorted, so if we've passed the index, stop searching
		if t.Start > runeIdx {
			break
		}
	}
	return TokenNone
}
