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
	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

// ChromaHighlighter wraps a Chroma lexer to implement the Highlighter interface.
type ChromaHighlighter struct {
	lexer    chroma.Lexer
	language string
}

// NewChromaHighlighter creates a new Chroma-based highlighter for the given language.
// Returns nil if no suitable lexer is found.
func NewChromaHighlighter(languageName string) *ChromaHighlighter {
	chromaName := GetChromaName(languageName)
	if chromaName == "" {
		return nil
	}

	lexer := lexers.Get(chromaName)
	if lexer == nil {
		// Try with the language name directly
		lexer = lexers.Get(languageName)
		if lexer == nil {
			return nil
		}
	}

	return &ChromaHighlighter{
		lexer:    lexer,
		language: languageName,
	}
}

// Language returns the name of the language this highlighter handles.
func (h *ChromaHighlighter) Language() string {
	return h.language
}

// Tokenize converts a line of text into syntax tokens.
func (h *ChromaHighlighter) Tokenize(line []rune) []Token {
	if len(line) == 0 {
		return nil
	}

	lineStr := string(line)
	iterator, err := h.lexer.Tokenise(nil, lineStr)
	if err != nil {
		return nil
	}

	var tokens []Token
	runePos := 0

	for _, tok := range iterator.Tokens() {
		tokType := mapChromaToken(tok.Type)
		tokRunes := []rune(tok.Value)
		tokLen := len(tokRunes)

		if tokType != TokenNone && tokLen > 0 {
			tokens = append(tokens, Token{
				Type:  tokType,
				Start: runePos,
				End:   runePos + tokLen,
			})
		}

		runePos += tokLen
	}

	return tokens
}

// mapChromaToken maps Chroma token types to our TokenType categories.
func mapChromaToken(ct chroma.TokenType) TokenType {
	switch {
	// Keywords
	case ct == chroma.Keyword,
		ct == chroma.KeywordConstant,
		ct == chroma.KeywordDeclaration,
		ct == chroma.KeywordNamespace,
		ct == chroma.KeywordPseudo,
		ct == chroma.KeywordReserved,
		ct == chroma.KeywordType:
		return TokenKeyword

	// Strings
	case ct == chroma.String,
		ct == chroma.StringAffix,
		ct == chroma.StringBacktick,
		ct == chroma.StringChar,
		ct == chroma.StringDelimiter,
		ct == chroma.StringDoc,
		ct == chroma.StringDouble,
		ct == chroma.StringEscape,
		ct == chroma.StringHeredoc,
		ct == chroma.StringInterpol,
		ct == chroma.StringOther,
		ct == chroma.StringRegex,
		ct == chroma.StringSingle,
		ct == chroma.StringSymbol:
		return TokenString

	// Comments
	case ct == chroma.Comment,
		ct == chroma.CommentHashbang,
		ct == chroma.CommentMultiline,
		ct == chroma.CommentPreproc,
		ct == chroma.CommentPreprocFile,
		ct == chroma.CommentSingle,
		ct == chroma.CommentSpecial:
		return TokenComment

	// Numbers
	case ct == chroma.Number,
		ct == chroma.NumberBin,
		ct == chroma.NumberFloat,
		ct == chroma.NumberHex,
		ct == chroma.NumberInteger,
		ct == chroma.NumberIntegerLong,
		ct == chroma.NumberOct:
		return TokenNumber

	// Operators
	case ct == chroma.Operator,
		ct == chroma.OperatorWord:
		return TokenOperator

	// Functions
	case ct == chroma.NameFunction,
		ct == chroma.NameFunctionMagic:
		return TokenFunction

	// Types
	case ct == chroma.NameClass,
		ct == chroma.NameException,
		ct == chroma.NameNamespace:
		return TokenType_

	// Variables
	case ct == chroma.NameVariable,
		ct == chroma.NameVariableAnonymous,
		ct == chroma.NameVariableClass,
		ct == chroma.NameVariableGlobal,
		ct == chroma.NameVariableInstance,
		ct == chroma.NameVariableMagic:
		return TokenVariable

	// Constants
	case ct == chroma.NameConstant,
		ct == chroma.LiteralDate:
		return TokenConstant

	// Punctuation
	case ct == chroma.Punctuation:
		return TokenPunctuation

	// Preprocessor
	case ct == chroma.CommentPreproc,
		ct == chroma.CommentPreprocFile:
		return TokenPreproc

	// Builtins
	case ct == chroma.NameBuiltin,
		ct == chroma.NameBuiltinPseudo:
		return TokenBuiltin

	default:
		return TokenNone
	}
}
