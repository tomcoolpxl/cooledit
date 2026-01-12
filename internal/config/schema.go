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

package config

// DefaultTabWidth is the default number of spaces per tab
const DefaultTabWidth = 4

// Config represents the application configuration
type Config struct {
	Editor     Editor                         `toml:"editor"`
	UI         UI                             `toml:"ui"`
	Search     Search                         `toml:"search"`
	Autosave   Autosave                       `toml:"autosave"`
	Themes     map[string]ThemeSpec           `toml:"themes"`
	Formatters map[string]FormatterConfigSpec `toml:"formatters"`
	Linters    map[string]LinterConfigSpec    `toml:"linters"`
}

// FormatterConfigSpec defines a formatter configuration in the config file
type FormatterConfigSpec struct {
	Command string   `toml:"command"`
	Args    []string `toml:"args"`
}

// LinterConfigSpec defines a linter configuration in the config file
type LinterConfigSpec struct {
	Command string   `toml:"command"`
	Args    []string `toml:"args"`
}

// Autosave contains autosave settings
type Autosave struct {
	Enabled     bool `toml:"enabled"`      // Enable autosave (default: true)
	IdleTimeout int  `toml:"idle_timeout"` // Seconds of idle before autosave (default: 2)
	MinInterval int  `toml:"min_interval"` // Minimum seconds between autosaves (default: 30)
}

// Editor contains editor-specific settings
type Editor struct {
	LineNumbers                  bool `toml:"line_numbers"`
	SoftWrap                     bool `toml:"soft_wrap"`
	TabWidth                     int  `toml:"tab_width"`
	SyntaxHighlighting           bool `toml:"syntax_highlighting"`
	ShowWhitespace               bool `toml:"show_whitespace"`
	CurrentLineHighlight         bool `toml:"current_line_highlight"`
	TrimTrailingWhitespaceOnSave bool `toml:"trim_trailing_whitespace"` // Trim trailing whitespace on save
	RememberPosition             bool `toml:"remember_position"`        // Remember cursor position in files
	ShowScrollbar                bool `toml:"show_scrollbar"`           // Show scrollbar on right edge
	ShowDiagnostics              bool `toml:"show_diagnostics"`         // Show linter diagnostics in gutter
}

// UI contains user interface settings
type UI struct {
	ShowMenubar   bool `toml:"show_menubar"`
	ShowStatusBar bool `toml:"show_statusbar"`

	Theme       string `toml:"theme"`
	CursorShape string `toml:"cursor_shape"`
	CursorBlink bool   `toml:"cursor_blink"`
	Language    string `toml:"language"` // Manual language override, empty = auto-detect
}

// Search contains search-related settings
type Search struct {
	CaseSensitive bool `toml:"case_sensitive"`
}

// ThemeSpec defines a custom theme in the config file
type ThemeSpec struct {
	Editor EditorThemeSpec `toml:"editor"`
	Search SearchThemeSpec `toml:"search"`
	Status StatusThemeSpec `toml:"statusbar"`
	Menu   MenuThemeSpec   `toml:"menubar"`
	Prompt PromptThemeSpec `toml:"prompt"`
	Help   HelpThemeSpec   `toml:"help"`
	Msg    MsgThemeSpec    `toml:"message"`
	Syntax SyntaxThemeSpec `toml:"syntax"`
}

type EditorThemeSpec struct {
	Fg               string `toml:"fg"`
	Bg               string `toml:"bg"`
	SelectionFg      string `toml:"selection_fg"`
	SelectionBg      string `toml:"selection_bg"`
	LineNumbersFg    string `toml:"line_numbers_fg"`
	LineNumbersBg    string `toml:"line_numbers_bg"`
	CursorColor      string `toml:"cursor_color"`
	BracketMatchBg   string `toml:"bracket_match_bg"`
	BracketUnmatchBg string `toml:"bracket_unmatch_bg"`
	CurrentLineBg    string `toml:"current_line_bg"`
}

type SearchThemeSpec struct {
	MatchFg        string `toml:"match_fg"`
	MatchBg        string `toml:"match_bg"`
	CurrentMatchFg string `toml:"current_match_fg"`
	CurrentMatchBg string `toml:"current_match_bg"`
}

type StatusThemeSpec struct {
	Fg         string `toml:"fg"`
	Bg         string `toml:"bg"`
	FilenameFg string `toml:"filename_fg"`
	ModifiedFg string `toml:"modified_fg"`
	PositionFg string `toml:"position_fg"`
	ModeFg     string `toml:"mode_fg"`
	HelpFg     string `toml:"help_fg"`
}

type MenuThemeSpec struct {
	Fg            string `toml:"fg"`
	Bg            string `toml:"bg"`
	SelectedFg    string `toml:"selected_fg"`
	SelectedBg    string `toml:"selected_bg"`
	DropdownFg    string `toml:"dropdown_fg"`
	DropdownBg    string `toml:"dropdown_bg"`
	DropdownSelFg string `toml:"dropdown_selected_fg"`
	DropdownSelBg string `toml:"dropdown_selected_bg"`
	AcceleratorFg string `toml:"accelerator_fg"`
}

type PromptThemeSpec struct {
	Fg      string `toml:"fg"`
	Bg      string `toml:"bg"`
	LabelFg string `toml:"label_fg"`
	InputFg string `toml:"input_fg"`
}

type HelpThemeSpec struct {
	Fg       string `toml:"fg"`
	Bg       string `toml:"bg"`
	TitleFg  string `toml:"title_fg"`
	TitleBg  string `toml:"title_bg"`
	FooterFg string `toml:"footer_fg"`
}

type MsgThemeSpec struct {
	InfoFg    string `toml:"info_fg"`
	InfoBg    string `toml:"info_bg"`
	WarningFg string `toml:"warning_fg"`
	WarningBg string `toml:"warning_bg"`
	ErrorFg   string `toml:"error_fg"`
	ErrorBg   string `toml:"error_bg"`
}

// SyntaxThemeSpec defines syntax highlighting colors in a custom theme
type SyntaxThemeSpec struct {
	KeywordFg     string `toml:"keyword_fg"`
	KeywordBg     string `toml:"keyword_bg"`
	StringFg      string `toml:"string_fg"`
	StringBg      string `toml:"string_bg"`
	CommentFg     string `toml:"comment_fg"`
	CommentBg     string `toml:"comment_bg"`
	NumberFg      string `toml:"number_fg"`
	NumberBg      string `toml:"number_bg"`
	OperatorFg    string `toml:"operator_fg"`
	OperatorBg    string `toml:"operator_bg"`
	FunctionFg    string `toml:"function_fg"`
	FunctionBg    string `toml:"function_bg"`
	TypeFg        string `toml:"type_fg"`
	TypeBg        string `toml:"type_bg"`
	VariableFg    string `toml:"variable_fg"`
	VariableBg    string `toml:"variable_bg"`
	ConstantFg    string `toml:"constant_fg"`
	ConstantBg    string `toml:"constant_bg"`
	PreprocFg     string `toml:"preproc_fg"`
	PreprocBg     string `toml:"preproc_bg"`
	BuiltinFg     string `toml:"builtin_fg"`
	BuiltinBg     string `toml:"builtin_bg"`
	PunctuationFg string `toml:"punctuation_fg"`
	PunctuationBg string `toml:"punctuation_bg"`
}

// Default autosave timing values
const (
	DefaultAutosaveIdleTimeout = 2  // seconds
	DefaultAutosaveMinInterval = 30 // seconds
)

// Default returns a Config with default values
func Default() *Config {
	return &Config{
		Editor: Editor{
			LineNumbers:        false,
			SoftWrap:           false,
			TabWidth:           DefaultTabWidth,
			SyntaxHighlighting: true,
			RememberPosition:   true,
			ShowScrollbar:      true,
			ShowDiagnostics:    true,
		},
		UI: UI{
			ShowMenubar:   false,
			ShowStatusBar: true,
			Theme:         "default",
			CursorShape:   "block",
			CursorBlink:   true,
		},
		Search: Search{
			CaseSensitive: true,
		},
		Autosave: Autosave{
			Enabled:     true,
			IdleTimeout: DefaultAutosaveIdleTimeout,
			MinInterval: DefaultAutosaveMinInterval,
		},
		Themes:     make(map[string]ThemeSpec),
		Formatters: make(map[string]FormatterConfigSpec),
		Linters:    make(map[string]LinterConfigSpec),
	}
}
