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

package theme

import "cooledit/internal/term"

// Theme defines all color elements for the editor UI.
// Each theme contains color definitions for the editor viewport, search highlights,
// status bar, menus, prompts, help screens, messages, and syntax highlighting.
type Theme struct {
	Name   string
	Editor EditorColors
	Search SearchColors
	Status StatusColors
	Menu   MenuColors
	Prompt PromptColors
	Help   HelpColors
	Msg    MessageColors
	Syntax SyntaxColors
}

// EditorColors defines colors for the text editing area
type EditorColors struct {
	Fg               term.Color // Normal text foreground
	Bg               term.Color // Normal text background
	SelectionFg      term.Color // Selected text foreground
	SelectionBg      term.Color // Selected text background
	LineNumbersFg    term.Color // Line numbers foreground
	LineNumbersBg    term.Color // Line numbers background
	CursorColor      term.Color // Cursor color (terminal support varies)
	BracketMatchBg   term.Color // Background for matched bracket pair
	BracketUnmatchBg term.Color // Background for unmatched bracket (error)
}

// SearchColors defines colors for search/find/replace
type SearchColors struct {
	MatchFg        term.Color // Search match foreground
	MatchBg        term.Color // Search match background
	CurrentMatchFg term.Color // Current match foreground
	CurrentMatchBg term.Color // Current match background
	ErrorBg        term.Color // Status bar background when no matches found
}

// StatusColors defines colors for the status bar
type StatusColors struct {
	Fg         term.Color // Status bar text
	Bg         term.Color // Status bar background
	FilenameFg term.Color // Filename display
	ModifiedFg term.Color // Modified indicator (*)
	PositionFg term.Color // Cursor position (Ln X, Col Y)
	ModeFg     term.Color // Mode indicator (REPLACE)
	HelpFg     term.Color // Mini-help text
}

// MenuColors defines colors for the menubar
type MenuColors struct {
	Fg            term.Color // Menu text
	Bg            term.Color // Menu background
	SelectedFg    term.Color // Selected menu foreground
	SelectedBg    term.Color // Selected menu background
	DropdownFg    term.Color // Dropdown item text
	DropdownBg    term.Color // Dropdown background
	DropdownSelFg term.Color // Dropdown selected item foreground
	DropdownSelBg term.Color // Dropdown selected item background
	AcceleratorFg term.Color // Keyboard shortcut hints
}

// PromptColors defines colors for the prompt/message area
type PromptColors struct {
	Fg      term.Color // Prompt text
	Bg      term.Color // Prompt background
	LabelFg term.Color // Prompt label (e.g., "Find: ")
	InputFg term.Color // User input text
}

// HelpColors defines colors for the help screen
type HelpColors struct {
	Fg       term.Color // Help text
	Bg       term.Color // Help background
	TitleFg  term.Color // Section titles
	TitleBg  term.Color // Section title background
	FooterFg term.Color // Footer message
}

// MessageColors defines colors for info/warning/error messages
type MessageColors struct {
	InfoFg    term.Color // Info message text
	InfoBg    term.Color // Info message background
	WarningFg term.Color // Warning message text
	WarningBg term.Color // Warning message background
	ErrorFg   term.Color // Error message text
	ErrorBg   term.Color // Error message background
}

// SyntaxColors defines colors for syntax highlighting tokens
type SyntaxColors struct {
	KeywordFg     term.Color // Keywords (if, for, func, etc.)
	KeywordBg     term.Color
	StringFg      term.Color // String literals
	StringBg      term.Color
	CommentFg     term.Color // Comments
	CommentBg     term.Color
	NumberFg      term.Color // Numeric literals
	NumberBg      term.Color
	OperatorFg    term.Color // Operators (+, -, *, etc.)
	OperatorBg    term.Color
	FunctionFg    term.Color // Function names
	FunctionBg    term.Color
	TypeFg        term.Color // Type names
	TypeBg        term.Color
	VariableFg    term.Color // Variable names
	VariableBg    term.Color
	ConstantFg    term.Color // Constants
	ConstantBg    term.Color
	PreprocFg     term.Color // Preprocessor directives
	PreprocBg     term.Color
	BuiltinFg     term.Color // Built-in functions
	BuiltinBg     term.Color
	PunctuationFg term.Color // Punctuation (brackets, braces)
	PunctuationBg term.Color
}

// GetStyle returns a term.Style with the given foreground and background colors
func GetStyle(fg, bg term.Color) term.Style {
	return term.Style{
		Foreground: fg,
		Background: bg,
		Inverse:    false,
	}
}

// GetInverseStyle returns a term.Style with inverse video enabled (for backwards compatibility)
func GetInverseStyle() term.Style {
	return term.Style{
		Inverse: true,
	}
}
