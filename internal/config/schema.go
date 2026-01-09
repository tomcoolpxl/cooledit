package config

// Config represents the application configuration
type Config struct {
	Editor Editor `toml:"editor"`
	UI     UI     `toml:"ui"`
	Search Search `toml:"search"`
}

// Editor contains editor-specific settings
type Editor struct {
	LineNumbers bool `toml:"line_numbers"`
	SoftWrap    bool `toml:"soft_wrap"`
	TabWidth    int  `toml:"tab_width"`
}

// UI contains user interface settings
type UI struct {
	ShowMenubar  bool `toml:"show_menubar"`
	MouseEnabled bool `toml:"mouse_enabled"`
}

// Search contains search-related settings
type Search struct {
	CaseSensitive bool `toml:"case_sensitive"`
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
			ShowMenubar:  false,
			MouseEnabled: false,
		},
		Search: Search{
			CaseSensitive: true,
		},
	}
}
