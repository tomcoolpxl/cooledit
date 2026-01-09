package theme

import "cooledit/internal/term"

// BuiltinThemes contains all 11 hardcoded themes
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
}

// GetBuiltinTheme returns a built-in theme by name, or the default theme if not found
func GetBuiltinTheme(name string) *Theme {
	if theme, ok := BuiltinThemes[name]; ok {
		return theme
	}
	return BuiltinThemes["default"]
}

// ListBuiltinThemes returns a sorted list of built-in theme names
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
	}
}

// defaultTheme uses terminal defaults with inverse video (current behavior)
func defaultTheme() *Theme {
	return &Theme{
		Name: "default",
		Editor: EditorColors{
			Fg:            term.ColorDefault,
			Bg:            term.ColorDefault,
			SelectionFg:   term.ColorDefault,
			SelectionBg:   term.ColorDefault,
			LineNumbersFg: term.ColorDefault,
			LineNumbersBg: term.ColorDefault,
		},
		Search: SearchColors{
			MatchFg:        term.ColorDefault,
			MatchBg:        term.ColorDefault,
			CurrentMatchFg: term.ColorDefault,
			CurrentMatchBg: term.ColorDefault,
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
	}
}

// darkTheme - classic dark background with light text
func darkTheme() *Theme {
	return &Theme{
		Name: "dark",
		Editor: EditorColors{
			Fg:            "#D0D0D0",
			Bg:            "#1E1E1E",
			SelectionFg:   "#FFFFFF",
			SelectionBg:   "#264F78",
			LineNumbersFg: "#858585",
			LineNumbersBg: "#1E1E1E",
		},
		Search: SearchColors{
			MatchFg:        "#000000",
			MatchBg:        "#A8FF60",
			CurrentMatchFg: "#000000",
			CurrentMatchBg: "#FFD700",
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
	}
}

// lightTheme - classic light background with dark text
func lightTheme() *Theme {
	return &Theme{
		Name: "light",
		Editor: EditorColors{
			Fg:            "#000000",
			Bg:            "#FFFFFF",
			SelectionFg:   "#000000",
			SelectionBg:   "#ADD6FF",
			LineNumbersFg: "#6E6E6E",
			LineNumbersBg: "#F5F5F5",
		},
		Search: SearchColors{
			MatchFg:        "#000000",
			MatchBg:        "#FFFF00",
			CurrentMatchFg: "#000000",
			CurrentMatchBg: "#FFA500",
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
	}
}

// monokaiTheme - popular dark theme with vibrant colors
func monokaiTheme() *Theme {
	return &Theme{
		Name: "monokai",
		Editor: EditorColors{
			Fg:            "#F8F8F2",
			Bg:            "#272822",
			SelectionFg:   "#F8F8F2",
			SelectionBg:   "#49483E",
			LineNumbersFg: "#90908A",
			LineNumbersBg: "#272822",
		},
		Search: SearchColors{
			MatchFg:        "#272822",
			MatchBg:        "#E6DB74",
			CurrentMatchFg: "#272822",
			CurrentMatchBg: "#FD971F",
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
	}
}

// solarizedDarkTheme - Ethan Schoonover's precision dark color scheme
func solarizedDarkTheme() *Theme {
	return &Theme{
		Name: "solarized-dark",
		Editor: EditorColors{
			Fg:            "#839496",
			Bg:            "#002B36",
			SelectionFg:   "#FDF6E3",
			SelectionBg:   "#073642",
			LineNumbersFg: "#586E75",
			LineNumbersBg: "#002B36",
		},
		Search: SearchColors{
			MatchFg:        "#002B36",
			MatchBg:        "#B58900",
			CurrentMatchFg: "#002B36",
			CurrentMatchBg: "#CB4B16",
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
			DropdownBg:    "#002B36",
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
	}
}

// solarizedLightTheme - Ethan Schoonover's precision light color scheme
func solarizedLightTheme() *Theme {
	return &Theme{
		Name: "solarized-light",
		Editor: EditorColors{
			Fg:            "#657B83",
			Bg:            "#FDF6E3",
			SelectionFg:   "#002B36",
			SelectionBg:   "#EEE8D5",
			LineNumbersFg: "#93A1A1",
			LineNumbersBg: "#FDF6E3",
		},
		Search: SearchColors{
			MatchFg:        "#FDF6E3",
			MatchBg:        "#B58900",
			CurrentMatchFg: "#FDF6E3",
			CurrentMatchBg: "#CB4B16",
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
			DropdownBg:    "#FDF6E3",
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
	}
}

// gruvboxDarkTheme - retro groove warm dark colors
func gruvboxDarkTheme() *Theme {
	return &Theme{
		Name: "gruvbox-dark",
		Editor: EditorColors{
			Fg:            "#EBDBB2",
			Bg:            "#282828",
			SelectionFg:   "#EBDBB2",
			SelectionBg:   "#504945",
			LineNumbersFg: "#928374",
			LineNumbersBg: "#282828",
		},
		Search: SearchColors{
			MatchFg:        "#282828",
			MatchBg:        "#FABD2F",
			CurrentMatchFg: "#282828",
			CurrentMatchBg: "#FE8019",
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
			DropdownBg:    "#282828",
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
	}
}

// gruvboxLightTheme - retro groove warm light colors
func gruvboxLightTheme() *Theme {
	return &Theme{
		Name: "gruvbox-light",
		Editor: EditorColors{
			Fg:            "#3C3836",
			Bg:            "#FBF1C7",
			SelectionFg:   "#3C3836",
			SelectionBg:   "#EBDBB2",
			LineNumbersFg: "#928374",
			LineNumbersBg: "#FBF1C7",
		},
		Search: SearchColors{
			MatchFg:        "#FBF1C7",
			MatchBg:        "#D79921",
			CurrentMatchFg: "#FBF1C7",
			CurrentMatchBg: "#D65D0E",
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
			DropdownBg:    "#FBF1C7",
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
	}
}

// draculaTheme - dark purple/pink theme, easy on the eyes
func draculaTheme() *Theme {
	return &Theme{
		Name: "dracula",
		Editor: EditorColors{
			Fg:            "#F8F8F2",
			Bg:            "#282A36",
			SelectionFg:   "#F8F8F2",
			SelectionBg:   "#44475A",
			LineNumbersFg: "#6272A4",
			LineNumbersBg: "#282A36",
		},
		Search: SearchColors{
			MatchFg:        "#282A36",
			MatchBg:        "#F1FA8C",
			CurrentMatchFg: "#282A36",
			CurrentMatchBg: "#FFB86C",
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
			DropdownBg:    "#282A36",
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
	}
}

// nordTheme - arctic bluish theme inspired by northern lights
func nordTheme() *Theme {
	return &Theme{
		Name: "nord",
		Editor: EditorColors{
			Fg:            "#D8DEE9",
			Bg:            "#2E3440",
			SelectionFg:   "#ECEFF4",
			SelectionBg:   "#434C5E",
			LineNumbersFg: "#4C566A",
			LineNumbersBg: "#2E3440",
		},
		Search: SearchColors{
			MatchFg:        "#2E3440",
			MatchBg:        "#EBCB8B",
			CurrentMatchFg: "#2E3440",
			CurrentMatchBg: "#D08770",
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
			DropdownBg:    "#2E3440",
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
	}
}

// dosTheme - classic DOS Edit colors (blue background, white/cyan text)
func dosTheme() *Theme {
	return &Theme{
		Name: "dos",
		Editor: EditorColors{
			Fg:            "#FFFFFF",
			Bg:            "#0000AA",
			SelectionFg:   "#000000",
			SelectionBg:   "#00AAAA",
			LineNumbersFg: "#AAAAAA",
			LineNumbersBg: "#000055",
		},
		Search: SearchColors{
			MatchFg:        "#000000",
			MatchBg:        "#FFFF00",
			CurrentMatchFg: "#FFFFFF",
			CurrentMatchBg: "#00AA00",
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
	}
}
