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

// BuiltinThemes contains all 14 hardcoded themes that ship with cooledit.
// These themes are always available without any configuration file.
var BuiltinThemes = map[string]*Theme{
	"default":         defaultTheme(),
	"dark":            darkTheme(),
	"light":           lightTheme(),
	"monokai":         monokaiTheme(),
	"solarized-dark":  solarizedDarkTheme(),
	"solarized-light": solarizedLightTheme(),
	"gruvbox-dark":    gruvboxDarkTheme(),
	"gruvbox-light":   gruvboxLightTheme(),
	"dracula":         draculaTheme(),
	"nord":            nordTheme(),
	"dos":             dosTheme(),
	"ibm-green":       ibmGreenTheme(),
	"ibm-amber":       ibmAmberTheme(),
	"cyberpunk":       cyberpunkTheme(),
}

// GetBuiltinTheme returns a built-in theme by name, or the default theme if not found
func GetBuiltinTheme(name string) *Theme {
	if theme, ok := BuiltinThemes[name]; ok {
		return theme
	}
	return BuiltinThemes["default"]
}

// ListBuiltinThemes returns a sorted list of all built-in theme names.
// The list is ordered with "default" first, followed by other themes.
func ListBuiltinThemes() []string {
	return []string{
		"default",
		"dark",
		"light",
		"monokai",
		"solarized-dark",
		"solarized-light",
		"gruvbox-dark",
		"gruvbox-light",
		"dracula",
		"nord",
		"dos",
		"ibm-green",
		"ibm-amber",
		"cyberpunk",
	}
}

// defaultTheme uses terminal defaults with inverse video (current behavior)
func defaultTheme() *Theme {
	return &Theme{
		Name: "default",
		Editor: EditorColors{
			Fg:               term.ColorDefault,
			Bg:               term.ColorDefault,
			SelectionFg:      term.ColorDefault,
			SelectionBg:      term.ColorDefault,
			LineNumbersFg:    "#585858",
			LineNumbersBg:    term.ColorDefault,
			CursorColor:      "#00FF00",
			BracketMatchBg:   term.ColorCyan,
			BracketUnmatchBg: term.ColorRed,
			CurrentLineBg:    term.ColorDefault, // Cannot derive - feature disabled on default theme
		},
		Search: SearchColors{
			MatchFg:        term.ColorDefault,
			MatchBg:        term.ColorDefault,
			CurrentMatchFg: term.ColorDefault,
			CurrentMatchBg: term.ColorDefault,
			ErrorBg:        term.ColorRed,
		},
		Status: StatusColors{
			Fg:         term.ColorDefault,
			Bg:         term.ColorDefault,
			FilenameFg: term.ColorDefault,
			ModifiedFg: term.ColorDefault,
			PositionFg: term.ColorDefault,
			ModeFg:     term.ColorDefault,
			HelpFg:     term.ColorDefault,
		},
		Menu: MenuColors{
			Fg:            term.ColorDefault,
			Bg:            term.ColorDefault,
			SelectedFg:    term.ColorDefault,
			SelectedBg:    term.ColorDefault,
			DropdownFg:    term.ColorDefault,
			DropdownBg:    term.ColorDefault,
			DropdownSelFg: term.ColorDefault,
			DropdownSelBg: term.ColorDefault,
			AcceleratorFg: term.ColorDefault,
		},
		Prompt: PromptColors{
			Fg:      term.ColorDefault,
			Bg:      term.ColorDefault,
			LabelFg: term.ColorDefault,
			InputFg: term.ColorDefault,
		},
		Help: HelpColors{
			Fg:       term.ColorDefault,
			Bg:       term.ColorDefault,
			TitleFg:  term.ColorDefault,
			TitleBg:  term.ColorDefault,
			FooterFg: term.ColorDefault,
		},
		Msg: MessageColors{
			InfoFg:    term.ColorDefault,
			InfoBg:    term.ColorDefault,
			WarningFg: term.ColorDefault,
			WarningBg: term.ColorDefault,
			ErrorFg:   term.ColorDefault,
			ErrorBg:   term.ColorDefault,
		},
		Syntax: SyntaxColors{
			KeywordFg:     term.ColorBlue,
			StringFg:      term.ColorGreen,
			CommentFg:     term.ColorCyan,
			NumberFg:      term.ColorMagenta,
			OperatorFg:    term.ColorDefault,
			FunctionFg:    term.ColorYellow,
			TypeFg:        term.ColorCyan,
			VariableFg:    term.ColorDefault,
			ConstantFg:    term.ColorMagenta,
			PreprocFg:     term.ColorRed,
			BuiltinFg:     term.ColorCyan,
			PunctuationFg: term.ColorDefault,
		},
		Diagnostic: DiagnosticColors{
			ErrorFg:   term.ColorWhite,
			ErrorBg:   term.ColorRed,
			WarningFg: term.ColorBlack,
			WarningBg: term.ColorYellow,
			InfoFg:    term.ColorWhite,
			InfoBg:    term.ColorBlue,
			HintFg:    term.ColorWhite,
			HintBg:    term.ColorCyan,
		},
		Fileview: FileviewColors{
			Fg:          term.ColorDefault,
			Bg:          term.ColorDefault,
			HeaderFg:    term.ColorDefault,
			HeaderBg:    term.ColorDefault,
			SelectionFg: term.ColorDefault,
			SelectionBg: term.ColorDefault,
			DirFg:       term.ColorDefault,
			SymlinkFg:   term.ColorCyan,
			ExpandFg:    term.ColorDefault,
		},
	}
}

// darkTheme - classic dark background with light text
func darkTheme() *Theme {
	return &Theme{
		Name: "dark",
		Editor: EditorColors{
			Fg:               "#D0D0D0",
			Bg:               "#1E1E1E",
			SelectionFg:      "#FFFFFF",
			SelectionBg:      "#264F78",
			LineNumbersFg:    "#858585",
			LineNumbersBg:    "#1E1E1E",
			CursorColor:      "#FFD700",
			BracketMatchBg:   "#3A3A3A",
			BracketUnmatchBg: "#5C2020",
			CurrentLineBg:    "#2A2A2A",
		},
		Search: SearchColors{
			MatchFg:        "#000000",
			MatchBg:        "#A8FF60",
			CurrentMatchFg: "#000000",
			CurrentMatchBg: "#FFD700",
			ErrorBg:        "#5C2020",
		},
		Status: StatusColors{
			Fg:         "#FFFFFF",
			Bg:         "#007ACC",
			FilenameFg: "#FFFFFF",
			ModifiedFg: "#FFD700",
			PositionFg: "#FFFFFF",
			ModeFg:     "#FFD700",
			HelpFg:     "#D0D0D0",
		},
		Menu: MenuColors{
			Fg:            "#FFFFFF",
			Bg:            "#2D2D30",
			SelectedFg:    "#FFFFFF",
			SelectedBg:    "#094771",
			DropdownFg:    "#CCCCCC",
			DropdownBg:    "#252526",
			DropdownSelFg: "#FFFFFF",
			DropdownSelBg: "#094771",
			AcceleratorFg: "#858585",
		},
		Prompt: PromptColors{
			Fg:      "#FFFFFF",
			Bg:      "#007ACC",
			LabelFg: "#FFD700",
			InputFg: "#FFFFFF",
		},
		Help: HelpColors{
			Fg:       "#D0D0D0",
			Bg:       "#1E1E1E",
			TitleFg:  "#000000",
			TitleBg:  "#007ACC",
			FooterFg: "#858585",
		},
		Msg: MessageColors{
			InfoFg:    "#FFFFFF",
			InfoBg:    "#007ACC",
			WarningFg: "#000000",
			WarningBg: "#FFD700",
			ErrorFg:   "#FFFFFF",
			ErrorBg:   "#FF0000",
		},
		Syntax: SyntaxColors{
			KeywordFg:     "#569CD6", // Blue
			StringFg:      "#CE9178", // Orange/brown
			CommentFg:     "#6A9955", // Green
			NumberFg:      "#B5CEA8", // Light green
			OperatorFg:    "#D4D4D4", // Light gray
			FunctionFg:    "#DCDCAA", // Yellow
			TypeFg:        "#4EC9B0", // Teal
			VariableFg:    "#9CDCFE", // Light blue
			ConstantFg:    "#4FC1FF", // Bright blue
			PreprocFg:     "#C586C0", // Purple
			BuiltinFg:     "#4EC9B0", // Teal
			PunctuationFg: "#D4D4D4", // Light gray
		},
		Diagnostic: DiagnosticColors{
			ErrorFg:   "#FFFFFF",
			ErrorBg:   "#F44747",
			WarningFg: "#000000",
			WarningBg: "#CCA700",
			InfoFg:    "#FFFFFF",
			InfoBg:    "#3794FF",
			HintFg:    "#FFFFFF",
			HintBg:    "#89D185",
		},
		Fileview: FileviewColors{
			Fg:          "#D4D4D4",
			Bg:          "#1E1E1E",
			HeaderFg:    "#FFFFFF",
			HeaderBg:    "#2D2D2D",
			SelectionFg: "#FFFFFF",
			SelectionBg: "#264F78",
			DirFg:       "#569CD6",
			SymlinkFg:   "#CE9178",
			ExpandFg:    "#808080",
		},
	}
}

// lightTheme - classic light background with dark text
func lightTheme() *Theme {
	return &Theme{
		Name: "light",
		Editor: EditorColors{
			Fg:               "#000000",
			Bg:               "#FFFFFF",
			SelectionFg:      "#000000",
			SelectionBg:      "#ADD6FF",
			LineNumbersFg:    "#6E6E6E",
			LineNumbersBg:    "#F5F5F5",
			CursorColor:      "#FFFFFF",
			BracketMatchBg:   "#E0E0E0",
			BracketUnmatchBg: "#FFCCCC",
			CurrentLineBg:    "#F0F0F0",
		},
		Search: SearchColors{
			MatchFg:        "#000000",
			MatchBg:        "#FFFF00",
			CurrentMatchFg: "#000000",
			CurrentMatchBg: "#FFA500",
			ErrorBg:        "#E51400",
		},
		Status: StatusColors{
			Fg:         "#FFFFFF",
			Bg:         "#007ACC",
			FilenameFg: "#FFFFFF",
			ModifiedFg: "#FFD700",
			PositionFg: "#FFFFFF",
			ModeFg:     "#FFD700",
			HelpFg:     "#E0E0E0",
		},
		Menu: MenuColors{
			Fg:            "#000000",
			Bg:            "#F3F3F3",
			SelectedFg:    "#FFFFFF",
			SelectedBg:    "#0078D7",
			DropdownFg:    "#000000",
			DropdownBg:    "#F0F0F0",
			DropdownSelFg: "#FFFFFF",
			DropdownSelBg: "#0078D7",
			AcceleratorFg: "#6E6E6E",
		},
		Prompt: PromptColors{
			Fg:      "#FFFFFF",
			Bg:      "#007ACC",
			LabelFg: "#FFD700",
			InputFg: "#FFFFFF",
		},
		Help: HelpColors{
			Fg:       "#000000",
			Bg:       "#FFFFFF",
			TitleFg:  "#FFFFFF",
			TitleBg:  "#007ACC",
			FooterFg: "#6E6E6E",
		},
		Msg: MessageColors{
			InfoFg:    "#FFFFFF",
			InfoBg:    "#007ACC",
			WarningFg: "#000000",
			WarningBg: "#FFD700",
			ErrorFg:   "#FFFFFF",
			ErrorBg:   "#E51400",
		},
		Syntax: SyntaxColors{
			KeywordFg:     "#0000FF", // Blue
			StringFg:      "#A31515", // Red/brown
			CommentFg:     "#008000", // Green
			NumberFg:      "#098658", // Dark green
			OperatorFg:    "#000000", // Black
			FunctionFg:    "#795E26", // Brown
			TypeFg:        "#267F99", // Teal
			VariableFg:    "#001080", // Dark blue
			ConstantFg:    "#0070C1", // Blue
			PreprocFg:     "#AF00DB", // Purple
			BuiltinFg:     "#267F99", // Teal
			PunctuationFg: "#000000", // Black
		},
		Diagnostic: DiagnosticColors{
			ErrorFg:   "#FFFFFF",
			ErrorBg:   "#E51400",
			WarningFg: "#000000",
			WarningBg:  "#F9A825",
			InfoFg:    "#FFFFFF",
			InfoBg:    "#1976D2",
			HintFg:    "#FFFFFF",
			HintBg:    "#388E3C",
		},
		Fileview: FileviewColors{
			Fg:          "#000000",
			Bg:          "#F3F3F3",
			HeaderFg:    "#000000",
			HeaderBg:    "#E0E0E0",
			SelectionFg: "#000000",
			SelectionBg: "#ADD6FF",
			DirFg:       "#0000FF",
			SymlinkFg:   "#795E26",
			ExpandFg:    "#808080",
		},
	}
}

// monokaiTheme - popular dark theme with vibrant colors
func monokaiTheme() *Theme {
	return &Theme{
		Name: "monokai",
		Editor: EditorColors{
			Fg:               "#F8F8F2",
			Bg:               "#272822",
			SelectionFg:      "#F8F8F2",
			SelectionBg:      "#49483E",
			LineNumbersFg:    "#90908A",
			LineNumbersBg:    "#272822",
			CursorColor:      "#F92672",
			BracketMatchBg:   "#49483E",
			BracketUnmatchBg: "#6E2020",
			CurrentLineBg:    "#3E3D32",
		},
		Search: SearchColors{
			MatchFg:        "#272822",
			MatchBg:        "#E6DB74",
			CurrentMatchFg: "#272822",
			CurrentMatchBg: "#FD971F",
			ErrorBg:        "#6E2020",
		},
		Status: StatusColors{
			Fg:         "#F8F8F2",
			Bg:         "#75715E",
			FilenameFg: "#F8F8F2",
			ModifiedFg: "#F92672",
			PositionFg: "#F8F8F2",
			ModeFg:     "#F92672",
			HelpFg:     "#F8F8F2",
		},
		Menu: MenuColors{
			Fg:            "#F8F8F2",
			Bg:            "#3E3D32",
			SelectedFg:    "#272822",
			SelectedBg:    "#F92672",
			DropdownFg:    "#F8F8F2",
			DropdownBg:    "#3E3D32",
			DropdownSelFg: "#272822",
			DropdownSelBg: "#F92672",
			AcceleratorFg: "#90908A",
		},
		Prompt: PromptColors{
			Fg:      "#F8F8F2",
			Bg:      "#75715E",
			LabelFg: "#F92672",
			InputFg: "#F8F8F2",
		},
		Help: HelpColors{
			Fg:       "#F8F8F2",
			Bg:       "#272822",
			TitleFg:  "#272822",
			TitleBg:  "#A6E22E",
			FooterFg: "#75715E",
		},
		Msg: MessageColors{
			InfoFg:    "#F8F8F2",
			InfoBg:    "#66D9EF",
			WarningFg: "#272822",
			WarningBg: "#E6DB74",
			ErrorFg:   "#F8F8F2",
			ErrorBg:   "#F92672",
		},
		Syntax: SyntaxColors{
			KeywordFg:     "#F92672", // Pink
			StringFg:      "#E6DB74", // Yellow
			CommentFg:     "#75715E", // Gray
			NumberFg:      "#AE81FF", // Purple
			OperatorFg:    "#F92672", // Pink
			FunctionFg:    "#A6E22E", // Green
			TypeFg:        "#66D9EF", // Cyan
			VariableFg:    "#F8F8F2", // White
			ConstantFg:    "#AE81FF", // Purple
			PreprocFg:     "#F92672", // Pink
			BuiltinFg:     "#66D9EF", // Cyan
			PunctuationFg: "#F8F8F2", // White
		},
		Diagnostic: DiagnosticColors{
			ErrorFg:   "#272822",
			ErrorBg:   "#F92672",
			WarningFg: "#272822",
			WarningBg: "#E6DB74",
			InfoFg:    "#272822",
			InfoBg:    "#66D9EF",
			HintFg:    "#272822",
			HintBg:    "#A6E22E",
		},
		Fileview: FileviewColors{
			Fg:          "#F8F8F2",
			Bg:          "#272822",
			HeaderFg:    "#F8F8F2",
			HeaderBg:    "#3E3D32",
			SelectionFg: "#F8F8F2",
			SelectionBg: "#49483E",
			DirFg:       "#66D9EF",
			SymlinkFg:   "#AE81FF",
			ExpandFg:    "#75715E",
		},
	}
}

// solarizedDarkTheme - Ethan Schoonover's precision dark color scheme
func solarizedDarkTheme() *Theme {
	return &Theme{
		Name: "solarized-dark",
		Editor: EditorColors{
			Fg:               "#839496",
			Bg:               "#002B36",
			SelectionFg:      "#FDF6E3",
			SelectionBg:      "#073642",
			LineNumbersFg:    "#586E75",
			LineNumbersBg:    "#002B36",
			CursorColor:      "#268BD2",
			BracketMatchBg:   "#073642",
			BracketUnmatchBg: "#6E2020",
			CurrentLineBg:    "#073642",
		},
		Search: SearchColors{
			MatchFg:        "#002B36",
			MatchBg:        "#B58900",
			CurrentMatchFg: "#002B36",
			CurrentMatchBg: "#CB4B16",
			ErrorBg:        "#6E2020",
		},
		Status: StatusColors{
			Fg:         "#93A1A1",
			Bg:         "#073642",
			FilenameFg: "#93A1A1",
			ModifiedFg: "#DC322F",
			PositionFg: "#93A1A1",
			ModeFg:     "#DC322F",
			HelpFg:     "#657B83",
		},
		Menu: MenuColors{
			Fg:            "#93A1A1",
			Bg:            "#073642",
			SelectedFg:    "#FDF6E3",
			SelectedBg:    "#268BD2",
			DropdownFg:    "#93A1A1",
			DropdownBg:    "#073642",
			DropdownSelFg: "#FDF6E3",
			DropdownSelBg: "#268BD2",
			AcceleratorFg: "#586E75",
		},
		Prompt: PromptColors{
			Fg:      "#93A1A1",
			Bg:      "#073642",
			LabelFg: "#268BD2",
			InputFg: "#93A1A1",
		},
		Help: HelpColors{
			Fg:       "#839496",
			Bg:       "#002B36",
			TitleFg:  "#002B36",
			TitleBg:  "#268BD2",
			FooterFg: "#586E75",
		},
		Msg: MessageColors{
			InfoFg:    "#FDF6E3",
			InfoBg:    "#268BD2",
			WarningFg: "#002B36",
			WarningBg: "#B58900",
			ErrorFg:   "#FDF6E3",
			ErrorBg:   "#DC322F",
		},
		Syntax: SyntaxColors{
			KeywordFg:     "#859900", // Green
			StringFg:      "#2AA198", // Cyan
			CommentFg:     "#586E75", // Base01
			NumberFg:      "#D33682", // Magenta
			OperatorFg:    "#839496", // Base0
			FunctionFg:    "#268BD2", // Blue
			TypeFg:        "#B58900", // Yellow
			VariableFg:    "#839496", // Base0
			ConstantFg:    "#CB4B16", // Orange
			PreprocFg:     "#CB4B16", // Orange
			BuiltinFg:     "#268BD2", // Blue
			PunctuationFg: "#839496", // Base0
		},
		Diagnostic: DiagnosticColors{
			ErrorFg:   "#002B36",
			ErrorBg:   "#DC322F",
			WarningFg: "#002B36",
			WarningBg: "#B58900",
			InfoFg:    "#002B36",
			InfoBg:    "#268BD2",
			HintFg:    "#002B36",
			HintBg:    "#2AA198",
		},
		Fileview: FileviewColors{
			Fg:          "#839496",
			Bg:          "#002B36",
			HeaderFg:    "#93A1A1",
			HeaderBg:    "#073642",
			SelectionFg: "#FDF6E3",
			SelectionBg: "#073642",
			DirFg:       "#268BD2",
			SymlinkFg:   "#2AA198",
			ExpandFg:    "#586E75",
		},
	}
}

// solarizedLightTheme - Ethan Schoonover's precision light color scheme
func solarizedLightTheme() *Theme {
	return &Theme{
		Name: "solarized-light",
		Editor: EditorColors{
			Fg:               "#657B83",
			Bg:               "#FDF6E3",
			SelectionFg:      "#002B36",
			SelectionBg:      "#EEE8D5",
			LineNumbersFg:    "#93A1A1",
			LineNumbersBg:    "#FDF6E3",
			CursorColor:      "#FDF6E3",
			BracketMatchBg:   "#EEE8D5",
			BracketUnmatchBg: "#FFCCCC",
			CurrentLineBg:    "#EEE8D5",
		},
		Search: SearchColors{
			MatchFg:        "#FDF6E3",
			MatchBg:        "#B58900",
			CurrentMatchFg: "#FDF6E3",
			CurrentMatchBg: "#CB4B16",
			ErrorBg:        "#DC322F",
		},
		Status: StatusColors{
			Fg:         "#586E75",
			Bg:         "#EEE8D5",
			FilenameFg: "#586E75",
			ModifiedFg: "#DC322F",
			PositionFg: "#586E75",
			ModeFg:     "#DC322F",
			HelpFg:     "#93A1A1",
		},
		Menu: MenuColors{
			Fg:            "#586E75",
			Bg:            "#EEE8D5",
			SelectedFg:    "#FDF6E3",
			SelectedBg:    "#268BD2",
			DropdownFg:    "#586E75",
			DropdownBg:    "#EEE8D5",
			DropdownSelFg: "#FDF6E3",
			DropdownSelBg: "#268BD2",
			AcceleratorFg: "#93A1A1",
		},
		Prompt: PromptColors{
			Fg:      "#586E75",
			Bg:      "#EEE8D5",
			LabelFg: "#268BD2",
			InputFg: "#586E75",
		},
		Help: HelpColors{
			Fg:       "#657B83",
			Bg:       "#FDF6E3",
			TitleFg:  "#FDF6E3",
			TitleBg:  "#268BD2",
			FooterFg: "#93A1A1",
		},
		Msg: MessageColors{
			InfoFg:    "#FDF6E3",
			InfoBg:    "#268BD2",
			WarningFg: "#002B36",
			WarningBg: "#B58900",
			ErrorFg:   "#FDF6E3",
			ErrorBg:   "#DC322F",
		},
		Syntax: SyntaxColors{
			KeywordFg:     "#859900", // Green
			StringFg:      "#2AA198", // Cyan
			CommentFg:     "#93A1A1", // Base1
			NumberFg:      "#D33682", // Magenta
			OperatorFg:    "#657B83", // Base00
			FunctionFg:    "#268BD2", // Blue
			TypeFg:        "#B58900", // Yellow
			VariableFg:    "#657B83", // Base00
			ConstantFg:    "#CB4B16", // Orange
			PreprocFg:     "#CB4B16", // Orange
			BuiltinFg:     "#268BD2", // Blue
			PunctuationFg: "#657B83", // Base00
		},
		Diagnostic: DiagnosticColors{
			ErrorFg:   "#FDF6E3",
			ErrorBg:   "#DC322F",
			WarningFg: "#FDF6E3",
			WarningBg: "#B58900",
			InfoFg:    "#FDF6E3",
			InfoBg:    "#268BD2",
			HintFg:    "#FDF6E3",
			HintBg:    "#2AA198",
		},
		Fileview: FileviewColors{
			Fg:          "#657B83",
			Bg:          "#FDF6E3",
			HeaderFg:    "#586E75",
			HeaderBg:    "#EEE8D5",
			SelectionFg: "#002B36",
			SelectionBg: "#EEE8D5",
			DirFg:       "#268BD2",
			SymlinkFg:   "#2AA198",
			ExpandFg:    "#93A1A1",
		},
	}
}

// gruvboxDarkTheme - retro groove warm dark colors
func gruvboxDarkTheme() *Theme {
	return &Theme{
		Name: "gruvbox-dark",
		Editor: EditorColors{
			Fg:               "#EBDBB2",
			Bg:               "#282828",
			SelectionFg:      "#EBDBB2",
			SelectionBg:      "#504945",
			LineNumbersFg:    "#928374",
			LineNumbersBg:    "#282828",
			CursorColor:      "#FE8019",
			BracketMatchBg:   "#3C3836",
			BracketUnmatchBg: "#9D0006",
			CurrentLineBg:    "#3C3836",
		},
		Search: SearchColors{
			MatchFg:        "#282828",
			MatchBg:        "#FABD2F",
			CurrentMatchFg: "#282828",
			CurrentMatchBg: "#FE8019",
			ErrorBg:        "#9D0006",
		},
		Status: StatusColors{
			Fg:         "#EBDBB2",
			Bg:         "#3C3836",
			FilenameFg: "#EBDBB2",
			ModifiedFg: "#FB4934",
			PositionFg: "#EBDBB2",
			ModeFg:     "#FB4934",
			HelpFg:     "#A89984",
		},
		Menu: MenuColors{
			Fg:            "#EBDBB2",
			Bg:            "#3C3836",
			SelectedFg:    "#282828",
			SelectedBg:    "#83A598",
			DropdownFg:    "#EBDBB2",
			DropdownBg:    "#3C3836",
			DropdownSelFg: "#282828",
			DropdownSelBg: "#83A598",
			AcceleratorFg: "#928374",
		},
		Prompt: PromptColors{
			Fg:      "#EBDBB2",
			Bg:      "#3C3836",
			LabelFg: "#83A598",
			InputFg: "#EBDBB2",
		},
		Help: HelpColors{
			Fg:       "#EBDBB2",
			Bg:       "#282828",
			TitleFg:  "#282828",
			TitleBg:  "#B8BB26",
			FooterFg: "#928374",
		},
		Msg: MessageColors{
			InfoFg:    "#EBDBB2",
			InfoBg:    "#458588",
			WarningFg: "#282828",
			WarningBg: "#FABD2F",
			ErrorFg:   "#EBDBB2",
			ErrorBg:   "#FB4934",
		},
		Syntax: SyntaxColors{
			KeywordFg:     "#FB4934", // Red
			StringFg:      "#B8BB26", // Green
			CommentFg:     "#928374", // Gray
			NumberFg:      "#D3869B", // Purple
			OperatorFg:    "#EBDBB2", // Cream
			FunctionFg:    "#FABD2F", // Yellow
			TypeFg:        "#83A598", // Aqua
			VariableFg:    "#EBDBB2", // Cream
			ConstantFg:    "#D3869B", // Purple
			PreprocFg:     "#FE8019", // Orange
			BuiltinFg:     "#83A598", // Aqua
			PunctuationFg: "#EBDBB2", // Cream
		},
		Diagnostic: DiagnosticColors{
			ErrorFg:   "#282828",
			ErrorBg:   "#FB4934",
			WarningFg: "#282828",
			WarningBg: "#FABD2F",
			InfoFg:    "#282828",
			InfoBg:    "#83A598",
			HintFg:    "#282828",
			HintBg:    "#8EC07C",
		},
		Fileview: FileviewColors{
			Fg:          "#EBDBB2",
			Bg:          "#282828",
			HeaderFg:    "#FBF1C7",
			HeaderBg:    "#3C3836",
			SelectionFg: "#EBDBB2",
			SelectionBg: "#3C3836",
			DirFg:       "#83A598",
			SymlinkFg:   "#D3869B",
			ExpandFg:    "#928374",
		},
	}
}

// gruvboxLightTheme - retro groove warm light colors
func gruvboxLightTheme() *Theme {
	return &Theme{
		Name: "gruvbox-light",
		Editor: EditorColors{
			Fg:               "#3C3836",
			Bg:               "#FBF1C7",
			SelectionFg:      "#3C3836",
			SelectionBg:      "#EBDBB2",
			LineNumbersFg:    "#928374",
			LineNumbersBg:    "#FBF1C7",
			CursorColor:      "#458588",
			BracketMatchBg:   "#EBDBB2",
			BracketUnmatchBg: "#9D0006",
			CurrentLineBg:    "#EBDBB2",
		},
		Search: SearchColors{
			MatchFg:        "#FBF1C7",
			MatchBg:        "#D79921",
			CurrentMatchFg: "#FBF1C7",
			CurrentMatchBg: "#D65D0E",
			ErrorBg:        "#9D0006",
		},
		Status: StatusColors{
			Fg:         "#3C3836",
			Bg:         "#D5C4A1",
			FilenameFg: "#3C3836",
			ModifiedFg: "#CC241D",
			PositionFg: "#3C3836",
			ModeFg:     "#CC241D",
			HelpFg:     "#7C6F64",
		},
		Menu: MenuColors{
			Fg:            "#3C3836",
			Bg:            "#D5C4A1",
			SelectedFg:    "#FBF1C7",
			SelectedBg:    "#458588",
			DropdownFg:    "#3C3836",
			DropdownBg:    "#D5C4A1",
			DropdownSelFg: "#FBF1C7",
			DropdownSelBg: "#458588",
			AcceleratorFg: "#928374",
		},
		Prompt: PromptColors{
			Fg:      "#3C3836",
			Bg:      "#D5C4A1",
			LabelFg: "#458588",
			InputFg: "#3C3836",
		},
		Help: HelpColors{
			Fg:       "#3C3836",
			Bg:       "#FBF1C7",
			TitleFg:  "#FBF1C7",
			TitleBg:  "#98971A",
			FooterFg: "#928374",
		},
		Msg: MessageColors{
			InfoFg:    "#FBF1C7",
			InfoBg:    "#458588",
			WarningFg: "#3C3836",
			WarningBg: "#D79921",
			ErrorFg:   "#FBF1C7",
			ErrorBg:   "#CC241D",
		},
		Syntax: SyntaxColors{
			KeywordFg:     "#9D0006", // Red
			StringFg:      "#79740E", // Green
			CommentFg:     "#928374", // Gray
			NumberFg:      "#8F3F71", // Purple
			OperatorFg:    "#3C3836", // Dark
			FunctionFg:    "#B57614", // Yellow
			TypeFg:        "#076678", // Aqua
			VariableFg:    "#3C3836", // Dark
			ConstantFg:    "#8F3F71", // Purple
			PreprocFg:     "#AF3A03", // Orange
			BuiltinFg:     "#076678", // Aqua
			PunctuationFg: "#3C3836", // Dark
		},
		Diagnostic: DiagnosticColors{
			ErrorFg:   "#FBF1C7",
			ErrorBg:   "#9D0006",
			WarningFg: "#FBF1C7",
			WarningBg: "#B57614",
			InfoFg:    "#FBF1C7",
			InfoBg:    "#076678",
			HintFg:    "#FBF1C7",
			HintBg:    "#427B58",
		},
		Fileview: FileviewColors{
			Fg:          "#3C3836",
			Bg:          "#FBF1C7",
			HeaderFg:    "#282828",
			HeaderBg:    "#EBDBB2",
			SelectionFg: "#3C3836",
			SelectionBg: "#EBDBB2",
			DirFg:       "#076678",
			SymlinkFg:   "#8F3F71",
			ExpandFg:    "#928374",
		},
	}
}

// draculaTheme - dark purple/pink theme, easy on the eyes
func draculaTheme() *Theme {
	return &Theme{
		Name: "dracula",
		Editor: EditorColors{
			Fg:               "#F8F8F2",
			Bg:               "#282A36",
			SelectionFg:      "#F8F8F2",
			SelectionBg:      "#44475A",
			LineNumbersFg:    "#6272A4",
			LineNumbersBg:    "#282A36",
			CursorColor:      "#FF79C6",
			BracketMatchBg:   "#44475A",
			BracketUnmatchBg: "#FF5555",
			CurrentLineBg:    "#44475A",
		},
		Search: SearchColors{
			MatchFg:        "#282A36",
			MatchBg:        "#F1FA8C",
			CurrentMatchFg: "#282A36",
			CurrentMatchBg: "#FFB86C",
			ErrorBg:        "#FF5555",
		},
		Status: StatusColors{
			Fg:         "#F8F8F2",
			Bg:         "#44475A",
			FilenameFg: "#F8F8F2",
			ModifiedFg: "#FF79C6",
			PositionFg: "#F8F8F2",
			ModeFg:     "#FF79C6",
			HelpFg:     "#F8F8F2",
		},
		Menu: MenuColors{
			Fg:            "#F8F8F2",
			Bg:            "#44475A",
			SelectedFg:    "#282A36",
			SelectedBg:    "#BD93F9",
			DropdownFg:    "#F8F8F2",
			DropdownBg:    "#44475A",
			DropdownSelFg: "#282A36",
			DropdownSelBg: "#BD93F9",
			AcceleratorFg: "#6272A4",
		},
		Prompt: PromptColors{
			Fg:      "#F8F8F2",
			Bg:      "#44475A",
			LabelFg: "#BD93F9",
			InputFg: "#F8F8F2",
		},
		Help: HelpColors{
			Fg:       "#F8F8F2",
			Bg:       "#282A36",
			TitleFg:  "#282A36",
			TitleBg:  "#50FA7B",
			FooterFg: "#6272A4",
		},
		Msg: MessageColors{
			InfoFg:    "#F8F8F2",
			InfoBg:    "#8BE9FD",
			WarningFg: "#282A36",
			WarningBg: "#F1FA8C",
			ErrorFg:   "#F8F8F2",
			ErrorBg:   "#FF5555",
		},
		Syntax: SyntaxColors{
			KeywordFg:     "#FF79C6", // Pink
			StringFg:      "#F1FA8C", // Yellow
			CommentFg:     "#6272A4", // Purple/gray
			NumberFg:      "#BD93F9", // Purple
			OperatorFg:    "#FF79C6", // Pink
			FunctionFg:    "#50FA7B", // Green
			TypeFg:        "#8BE9FD", // Cyan
			VariableFg:    "#F8F8F2", // White
			ConstantFg:    "#BD93F9", // Purple
			PreprocFg:     "#FFB86C", // Orange
			BuiltinFg:     "#8BE9FD", // Cyan
			PunctuationFg: "#F8F8F2", // White
		},
		Diagnostic: DiagnosticColors{
			ErrorFg:   "#282A36",
			ErrorBg:   "#FF5555",
			WarningFg: "#282A36",
			WarningBg: "#F1FA8C",
			InfoFg:    "#282A36",
			InfoBg:    "#8BE9FD",
			HintFg:    "#282A36",
			HintBg:    "#50FA7B",
		},
		Fileview: FileviewColors{
			Fg:          "#F8F8F2",
			Bg:          "#282A36",
			HeaderFg:    "#F8F8F2",
			HeaderBg:    "#44475A",
			SelectionFg: "#F8F8F2",
			SelectionBg: "#44475A",
			DirFg:       "#8BE9FD",
			SymlinkFg:   "#BD93F9",
			ExpandFg:    "#6272A4",
		},
	}
}

// nordTheme - arctic bluish theme inspired by northern lights
func nordTheme() *Theme {
	return &Theme{
		Name: "nord",
		Editor: EditorColors{
			Fg:               "#D8DEE9",
			Bg:               "#2E3440",
			SelectionFg:      "#ECEFF4",
			SelectionBg:      "#434C5E",
			LineNumbersFg:    "#4C566A",
			LineNumbersBg:    "#2E3440",
			CursorColor:      "#88C0D0",
			BracketMatchBg:   "#3B4252",
			BracketUnmatchBg: "#BF616A",
			CurrentLineBg:    "#3B4252",
		},
		Search: SearchColors{
			MatchFg:        "#2E3440",
			MatchBg:        "#EBCB8B",
			CurrentMatchFg: "#2E3440",
			CurrentMatchBg: "#D08770",
			ErrorBg:        "#BF616A",
		},
		Status: StatusColors{
			Fg:         "#ECEFF4",
			Bg:         "#3B4252",
			FilenameFg: "#ECEFF4",
			ModifiedFg: "#BF616A",
			PositionFg: "#ECEFF4",
			ModeFg:     "#BF616A",
			HelpFg:     "#D8DEE9",
		},
		Menu: MenuColors{
			Fg:            "#ECEFF4",
			Bg:            "#3B4252",
			SelectedFg:    "#2E3440",
			SelectedBg:    "#88C0D0",
			DropdownFg:    "#ECEFF4",
			DropdownBg:    "#3B4252",
			DropdownSelFg: "#2E3440",
			DropdownSelBg: "#88C0D0",
			AcceleratorFg: "#4C566A",
		},
		Prompt: PromptColors{
			Fg:      "#ECEFF4",
			Bg:      "#3B4252",
			LabelFg: "#88C0D0",
			InputFg: "#ECEFF4",
		},
		Help: HelpColors{
			Fg:       "#D8DEE9",
			Bg:       "#2E3440",
			TitleFg:  "#2E3440",
			TitleBg:  "#A3BE8C",
			FooterFg: "#4C566A",
		},
		Msg: MessageColors{
			InfoFg:    "#ECEFF4",
			InfoBg:    "#5E81AC",
			WarningFg: "#2E3440",
			WarningBg: "#EBCB8B",
			ErrorFg:   "#ECEFF4",
			ErrorBg:   "#BF616A",
		},
		Syntax: SyntaxColors{
			KeywordFg:     "#81A1C1", // Blue
			StringFg:      "#A3BE8C", // Green
			CommentFg:     "#616E88", // Gray
			NumberFg:      "#B48EAD", // Purple
			OperatorFg:    "#81A1C1", // Blue
			FunctionFg:    "#88C0D0", // Cyan
			TypeFg:        "#8FBCBB", // Teal
			VariableFg:    "#D8DEE9", // White
			ConstantFg:    "#B48EAD", // Purple
			PreprocFg:     "#D08770", // Orange
			BuiltinFg:     "#8FBCBB", // Teal
			PunctuationFg: "#ECEFF4", // White
		},
		Diagnostic: DiagnosticColors{
			ErrorFg:   "#2E3440",
			ErrorBg:   "#BF616A",
			WarningFg: "#2E3440",
			WarningBg: "#EBCB8B",
			InfoFg:    "#2E3440",
			InfoBg:    "#81A1C1",
			HintFg:    "#2E3440",
			HintBg:    "#A3BE8C",
		},
		Fileview: FileviewColors{
			Fg:          "#D8DEE9",
			Bg:          "#2E3440",
			HeaderFg:    "#ECEFF4",
			HeaderBg:    "#3B4252",
			SelectionFg: "#ECEFF4",
			SelectionBg: "#3B4252",
			DirFg:       "#88C0D0",
			SymlinkFg:   "#B48EAD",
			ExpandFg:    "#4C566A",
		},
	}
}

// dosTheme - classic DOS Edit colors (blue background, white/cyan text)
func dosTheme() *Theme {
	return &Theme{
		Name: "dos",
		Editor: EditorColors{
			Fg:               "#FFFFFF",
			Bg:               "#0000AA",
			SelectionFg:      "#000000",
			SelectionBg:      "#00AAAA",
			LineNumbersFg:    "#AAAAAA",
			LineNumbersBg:    "#000055",
			CursorColor:      "#FFFF00",
			BracketMatchBg:   "#0000AA",
			BracketUnmatchBg: "#AA0000",
			CurrentLineBg:    "#0000CC",
		},
		Search: SearchColors{
			MatchFg:        "#000000",
			MatchBg:        "#FFFF00",
			CurrentMatchFg: "#FFFFFF",
			CurrentMatchBg: "#00AA00",
			ErrorBg:        "#AA0000",
		},
		Status: StatusColors{
			Fg:         "#000000",
			Bg:         "#00AAAA",
			FilenameFg: "#000000",
			ModifiedFg: "#FF0000",
			PositionFg: "#000000",
			ModeFg:     "#FFFF00",
			HelpFg:     "#000000",
		},
		Menu: MenuColors{
			Fg:            "#000000",
			Bg:            "#00AAAA",
			SelectedFg:    "#FFFFFF",
			SelectedBg:    "#AA00AA",
			DropdownFg:    "#FFFFFF",
			DropdownBg:    "#000055",
			DropdownSelFg: "#FFFF00",
			DropdownSelBg: "#AA00AA",
			AcceleratorFg: "#AAAAAA",
		},
		Prompt: PromptColors{
			Fg:      "#000000",
			Bg:      "#00AAAA",
			LabelFg: "#FFFF00",
			InputFg: "#FFFFFF",
		},
		Help: HelpColors{
			Fg:       "#FFFFFF",
			Bg:       "#0000AA",
			TitleFg:  "#FFFF00",
			TitleBg:  "#00AAAA",
			FooterFg: "#AAAAAA",
		},
		Msg: MessageColors{
			InfoFg:    "#FFFFFF",
			InfoBg:    "#0000AA",
			WarningFg: "#000000",
			WarningBg: "#FFFF00",
			ErrorFg:   "#FFFFFF",
			ErrorBg:   "#AA0000",
		},
		Syntax: SyntaxColors{
			KeywordFg:     "#00AAAA", // Cyan
			StringFg:      "#FFFF00", // Yellow
			CommentFg:     "#AAAAAA", // Gray
			NumberFg:      "#AA00AA", // Magenta
			OperatorFg:    "#FFFFFF", // White
			FunctionFg:    "#00AA00", // Green
			TypeFg:        "#00AAAA", // Cyan
			VariableFg:    "#FFFFFF", // White
			ConstantFg:    "#AA00AA", // Magenta
			PreprocFg:     "#AA5500", // Brown/Orange
			BuiltinFg:     "#00AAAA", // Cyan
			PunctuationFg: "#FFFFFF", // White
		},
		Diagnostic: DiagnosticColors{
			ErrorFg:   "#FFFFFF",
			ErrorBg:   "#AA0000",
			WarningFg: "#000000",
			WarningBg: "#FFFF00",
			InfoFg:    "#FFFFFF",
			InfoBg:    "#0000AA",
			HintFg:    "#000000",
			HintBg:    "#00FFFF",
		},
		Fileview: FileviewColors{
			Fg:          "#FFFFFF",
			Bg:          "#0000AA",
			HeaderFg:    "#FFFFFF",
			HeaderBg:    "#000080",
			SelectionFg: "#000000",
			SelectionBg: "#00AAAA",
			DirFg:       "#FFFF55",
			SymlinkFg:   "#AA00AA",
			ExpandFg:    "#AAAAAA",
		},
	}
}

// ibmGreenTheme - classic IBM green phosphor monitor colors
func ibmGreenTheme() *Theme {
	return &Theme{
		Name: "ibm-green",
		Editor: EditorColors{
			Fg:               "#00FF00",
			Bg:               "#000000",
			SelectionFg:      "#000000",
			SelectionBg:      "#00AA00",
			LineNumbersFg:    "#008800",
			LineNumbersBg:    "#000000",
			CursorColor:      "#00FF00",
			BracketMatchBg:   "#003300",
			BracketUnmatchBg: "#330000",
			CurrentLineBg:    "#0A1A0A",
		},
		Search: SearchColors{
			MatchFg:        "#000000",
			MatchBg:        "#00FF00",
			CurrentMatchFg: "#000000",
			CurrentMatchBg: "#00DD00",
			ErrorBg:        "#330000",
		},
		Status: StatusColors{
			Fg:         "#00FF00",
			Bg:         "#003300",
			FilenameFg: "#00FF00",
			ModifiedFg: "#00FF00",
			PositionFg: "#00FF00",
			ModeFg:     "#00DD00",
			HelpFg:     "#00CC00",
		},
		Menu: MenuColors{
			Fg:            "#00FF00",
			Bg:            "#002200",
			SelectedFg:    "#000000",
			SelectedBg:    "#00AA00",
			DropdownFg:    "#00FF00",
			DropdownBg:    "#001100",
			DropdownSelFg: "#000000",
			DropdownSelBg: "#00AA00",
			AcceleratorFg: "#008800",
		},
		Prompt: PromptColors{
			Fg:      "#00FF00",
			Bg:      "#003300",
			LabelFg: "#00DD00",
			InputFg: "#00FF00",
		},
		Help: HelpColors{
			Fg:       "#00FF00",
			Bg:       "#000000",
			TitleFg:  "#000000",
			TitleBg:  "#00AA00",
			FooterFg: "#008800",
		},
		Msg: MessageColors{
			InfoFg:    "#00FF00",
			InfoBg:    "#003300",
			WarningFg: "#00FF00",
			WarningBg: "#005500",
			ErrorFg:   "#00FF00",
			ErrorBg:   "#006600",
		},
		Syntax: SyntaxColors{
			KeywordFg:     "#00FF00", // Bright green
			StringFg:      "#00DD00", // Medium green
			CommentFg:     "#008800", // Dark green
			NumberFg:      "#00EE00", // Light green
			OperatorFg:    "#00FF00", // Bright green
			FunctionFg:    "#00CC00", // Medium-bright green
			TypeFg:        "#00BB00", // Medium green
			VariableFg:    "#00FF00", // Bright green
			ConstantFg:    "#00EE00", // Light green
			PreprocFg:     "#00AA00", // Darker green
			BuiltinFg:     "#00BB00", // Medium green
			PunctuationFg: "#00FF00", // Bright green
		},
		Diagnostic: DiagnosticColors{
			ErrorFg:   "#000000",
			ErrorBg:   "#FF0000",
			WarningFg: "#000000",
			WarningBg: "#FFFF00",
			InfoFg:    "#000000",
			InfoBg:    "#00FF00",
			HintFg:    "#000000",
			HintBg:    "#00AA00",
		},
		Fileview: FileviewColors{
			Fg:          "#33FF33",
			Bg:          "#000000",
			HeaderFg:    "#66FF66",
			HeaderBg:    "#003300",
			SelectionFg: "#000000",
			SelectionBg: "#005500",
			DirFg:       "#66FF66",
			SymlinkFg:   "#00CC00",
			ExpandFg:    "#009900",
		},
	}
}

// ibmAmberTheme - classic IBM amber phosphor monitor colors
func ibmAmberTheme() *Theme {
	return &Theme{
		Name: "ibm-amber",
		Editor: EditorColors{
			Fg:               "#FFAA00",
			Bg:               "#000000",
			SelectionFg:      "#000000",
			SelectionBg:      "#CC8800",
			LineNumbersFg:    "#AA7700",
			LineNumbersBg:    "#000000",
			CursorColor:      "#FFAA00",
			BracketMatchBg:   "#332200",
			BracketUnmatchBg: "#330000",
			CurrentLineBg:    "#1A1000",
		},
		Search: SearchColors{
			MatchFg:        "#000000",
			MatchBg:        "#FFAA00",
			CurrentMatchFg: "#000000",
			CurrentMatchBg: "#DD9900",
			ErrorBg:        "#330000",
		},
		Status: StatusColors{
			Fg:         "#FFAA00",
			Bg:         "#332200",
			FilenameFg: "#FFAA00",
			ModifiedFg: "#FFAA00",
			PositionFg: "#FFAA00",
			ModeFg:     "#FFCC00",
			HelpFg:     "#DD9900",
		},
		Menu: MenuColors{
			Fg:            "#FFAA00",
			Bg:            "#221800",
			SelectedFg:    "#000000",
			SelectedBg:    "#CC8800",
			DropdownFg:    "#FFAA00",
			DropdownBg:    "#110C00",
			DropdownSelFg: "#000000",
			DropdownSelBg: "#CC8800",
			AcceleratorFg: "#AA7700",
		},
		Prompt: PromptColors{
			Fg:      "#FFAA00",
			Bg:      "#332200",
			LabelFg: "#FFCC00",
			InputFg: "#FFAA00",
		},
		Help: HelpColors{
			Fg:       "#FFAA00",
			Bg:       "#000000",
			TitleFg:  "#000000",
			TitleBg:  "#CC8800",
			FooterFg: "#AA7700",
		},
		Msg: MessageColors{
			InfoFg:    "#FFAA00",
			InfoBg:    "#332200",
			WarningFg: "#FFAA00",
			WarningBg: "#443300",
			ErrorFg:   "#FFAA00",
			ErrorBg:   "#554400",
		},
		Syntax: SyntaxColors{
			KeywordFg:     "#FFAA00", // Bright amber
			StringFg:      "#DD9900", // Medium amber
			CommentFg:     "#AA7700", // Dark amber
			NumberFg:      "#FFCC00", // Light amber
			OperatorFg:    "#FFAA00", // Bright amber
			FunctionFg:    "#FFBB00", // Bright amber
			TypeFg:        "#CC8800", // Medium amber
			VariableFg:    "#FFAA00", // Bright amber
			ConstantFg:    "#FFCC00", // Light amber
			PreprocFg:     "#BB7700", // Darker amber
			BuiltinFg:     "#CC8800", // Medium amber
			PunctuationFg: "#FFAA00", // Bright amber
		},
		Diagnostic: DiagnosticColors{
			ErrorFg:   "#000000",
			ErrorBg:   "#FF0000",
			WarningFg: "#000000",
			WarningBg: "#FFFF00",
			InfoFg:    "#000000",
			InfoBg:    "#FFAA00",
			HintFg:    "#000000",
			HintBg:    "#AA7700",
		},
		Fileview: FileviewColors{
			Fg:          "#FFB000",
			Bg:          "#000000",
			HeaderFg:    "#FFC000",
			HeaderBg:    "#331A00",
			SelectionFg: "#000000",
			SelectionBg: "#553300",
			DirFg:       "#FFC000",
			SymlinkFg:   "#CC8800",
			ExpandFg:    "#996600",
		},
	}
}

// cyberpunkTheme is a neon/gradient theme with vibrant colors
// Features: hot pink, cyan, purple gradients on dark background
func cyberpunkTheme() *Theme {
	return &Theme{
		Name: "cyberpunk",
		Editor: EditorColors{
			Fg:               "#00FFFF", // Cyan text
			Bg:               "#0A0E27", // Deep dark blue
			SelectionFg:      "#000000",
			SelectionBg:      "#FF2A6D", // Hot pink selection
			LineNumbersFg:    "#8B5CF6", // Purple line numbers
			LineNumbersBg:    "#1a1f3a",
			CursorColor:      "#00FF41", // Matrix green cursor
			BracketMatchBg:   "#1A1A2E",
			BracketUnmatchBg: "#FF0055",
			CurrentLineBg:    "#151934",
		},
		Search: SearchColors{
			MatchFg:        "#000000",
			MatchBg:        "#FFD700", // Gold
			CurrentMatchFg: "#000000",
			CurrentMatchBg: "#FF2A6D", // Hot pink
		},
		Status: StatusColors{
			Fg:         "#00FFFF",
			Bg:         "#1a1f3a",
			FilenameFg: "#FF2A6D", // Hot pink filename
			ModifiedFg: "#FFD700", // Gold for modified
			PositionFg: "#8B5CF6", // Purple position
			ModeFg:     "#00FF41", // Matrix green mode
			HelpFg:     "#888888",
		},
		Menu: MenuColors{
			Fg:            "#00FFFF",
			Bg:            "#1a1f3a",
			SelectedFg:    "#000000",
			SelectedBg:    "#FF2A6D",
			DropdownFg:    "#00FFFF",
			DropdownBg:    "#0A0E27",
			DropdownSelFg: "#000000",
			DropdownSelBg: "#8B5CF6",
			AcceleratorFg: "#FF2A6D",
		},
		Prompt: PromptColors{
			Fg:      "#00FFFF",
			Bg:      "#1a1f3a",
			LabelFg: "#FF2A6D",
			InputFg: "#00FF41",
		},
		Help: HelpColors{
			Fg:       "#00FFFF",
			Bg:       "#0A0E27",
			TitleFg:  "#000000",
			TitleBg:  "#FF2A6D",
			FooterFg: "#8B5CF6",
		},
		Msg: MessageColors{
			InfoFg:    "#00FFFF",
			InfoBg:    "#1a1f3a",
			WarningFg: "#FFD700",
			WarningBg: "#332200",
			ErrorFg:   "#FF2A6D",
			ErrorBg:   "#330011",
		},
		Syntax: SyntaxColors{
			KeywordFg:     "#FF2A6D", // Hot pink
			StringFg:      "#00FF41", // Matrix green
			CommentFg:     "#666699", // Muted purple
			NumberFg:      "#8B5CF6", // Purple
			OperatorFg:    "#FF2A6D", // Hot pink
			FunctionFg:    "#FFD700", // Gold
			TypeFg:        "#00FFFF", // Cyan
			VariableFg:    "#00FFFF", // Cyan
			ConstantFg:    "#8B5CF6", // Purple
			PreprocFg:     "#FFD700", // Gold
			BuiltinFg:     "#00FFFF", // Cyan
			PunctuationFg: "#00FFFF", // Cyan
		},
		Diagnostic: DiagnosticColors{
			ErrorFg:   "#0D0D0D",
			ErrorBg:   "#FF0055",
			WarningFg: "#0D0D0D",
			WarningBg: "#FFFF00",
			InfoFg:    "#0D0D0D",
			InfoBg:    "#00FFFF",
			HintFg:    "#0D0D0D",
			HintBg:    "#BD00FF",
		},
		Fileview: FileviewColors{
			Fg:          "#0FF0FC",
			Bg:          "#0D0D0D",
			HeaderFg:    "#00FFFF",
			HeaderBg:    "#1A1A2E",
			SelectionFg: "#0D0D0D",
			SelectionBg: "#1A1A2E",
			DirFg:       "#FF2A6D",
			SymlinkFg:   "#BD00FF",
			ExpandFg:    "#05D9E8",
		},
	}
}
