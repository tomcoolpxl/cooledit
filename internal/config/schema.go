package config

// Config represents the application configuration
type Config struct {
	Editor Editor               `toml:"editor"`
	UI     UI                   `toml:"ui"`
	Search Search               `toml:"search"`
	Themes map[string]ThemeSpec `toml:"themes"`
}

// Editor contains editor-specific settings
type Editor struct {
	LineNumbers bool `toml:"line_numbers"`
	SoftWrap    bool `toml:"soft_wrap"`
	TabWidth    int  `toml:"tab_width"`
}

// UI contains user interface settings
type UI struct {
	ShowMenubar   bool   `toml:"show_menubar"`
	ShowStatusBar bool   `toml:"show_statusbar"`

	Theme         string `toml:"theme"`
	CursorShape   string `toml:"cursor_shape"`
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
}

type EditorThemeSpec struct {
	Fg            string `toml:"fg"`
	Bg            string `toml:"bg"`
	SelectionFg   string `toml:"selection_fg"`
	SelectionBg   string `toml:"selection_bg"`
	LineNumbersFg string `toml:"line_numbers_fg"`
	LineNumbersBg string `toml:"line_numbers_bg"`
	CursorColor   string `toml:"cursor_color"`
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

// Default returns a Config with default values
func Default() *Config {
	return &Config{
		Editor: Editor{
			LineNumbers: false,
			SoftWrap:    false,
			TabWidth:    4,
		},
		UI: UI{
			ShowMenubar:   false,
			ShowStatusBar: true,
			Theme:         "default",
			CursorShape:   "block",
		},
		Search: Search{
			CaseSensitive: true,
		},
		Themes: make(map[string]ThemeSpec),
	}
}
