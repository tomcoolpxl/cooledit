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

import (
	"os"
	"path/filepath"
	"testing"

	"cooledit/internal/term"
)

// --- ConvertThemeSpec ---

// TestConvertThemeSpecName verifies the theme name is carried through.
func TestConvertThemeSpecName(t *testing.T) {
	spec := ThemeSpec{}
	got := ConvertThemeSpec("my-theme", spec)
	if got.Name != "my-theme" {
		t.Errorf("Name = %q, want %q", got.Name, "my-theme")
	}
}

// TestConvertThemeSpecEmpty verifies that an all-empty ThemeSpec produces
// ColorDefault for every field (empty string → ColorDefault via ParseColor).
func TestConvertThemeSpecEmpty(t *testing.T) {
	got := ConvertThemeSpec("empty", ThemeSpec{})

	if got.Editor.Fg != term.ColorDefault {
		t.Errorf("Editor.Fg = %q, want ColorDefault", got.Editor.Fg)
	}
	if got.Editor.Bg != term.ColorDefault {
		t.Errorf("Editor.Bg = %q, want ColorDefault", got.Editor.Bg)
	}
	if got.Status.Fg != term.ColorDefault {
		t.Errorf("Status.Fg = %q, want ColorDefault", got.Status.Fg)
	}
	if got.Menu.Bg != term.ColorDefault {
		t.Errorf("Menu.Bg = %q, want ColorDefault", got.Menu.Bg)
	}
}

// TestConvertThemeSpecHexColors verifies hex color strings are preserved.
func TestConvertThemeSpecHexColors(t *testing.T) {
	spec := ThemeSpec{
		Editor: EditorThemeSpec{
			Fg:          "#282828",
			Bg:          "#EBDBB2",
			SelectionFg: "#FF0000",
			SelectionBg: "#0000FF",
			CursorColor: "#00FF00",
		},
		Status: StatusThemeSpec{
			Fg: "#FFFFFF",
			Bg: "#1D1D1D",
		},
	}

	got := ConvertThemeSpec("hex-test", spec)

	if got.Editor.Fg != term.Color("#282828") {
		t.Errorf("Editor.Fg = %q, want #282828", got.Editor.Fg)
	}
	if got.Editor.Bg != term.Color("#EBDBB2") {
		t.Errorf("Editor.Bg = %q, want #EBDBB2", got.Editor.Bg)
	}
	if got.Editor.SelectionFg != term.Color("#FF0000") {
		t.Errorf("Editor.SelectionFg = %q, want #FF0000", got.Editor.SelectionFg)
	}
	if got.Editor.SelectionBg != term.Color("#0000FF") {
		t.Errorf("Editor.SelectionBg = %q, want #0000FF", got.Editor.SelectionBg)
	}
	if got.Editor.CursorColor != term.Color("#00FF00") {
		t.Errorf("Editor.CursorColor = %q, want #00FF00", got.Editor.CursorColor)
	}
	if got.Status.Fg != term.Color("#FFFFFF") {
		t.Errorf("Status.Fg = %q, want #FFFFFF", got.Status.Fg)
	}
	if got.Status.Bg != term.Color("#1D1D1D") {
		t.Errorf("Status.Bg = %q, want #1D1D1D", got.Status.Bg)
	}
}

// TestConvertThemeSpecNamedColors verifies named color strings are resolved.
func TestConvertThemeSpecNamedColors(t *testing.T) {
	spec := ThemeSpec{
		Editor: EditorThemeSpec{Fg: "white", Bg: "black"},
		Status: StatusThemeSpec{Fg: "cyan", Bg: "blue"},
		Menu:   MenuThemeSpec{Fg: "yellow", Bg: "magenta"},
	}

	got := ConvertThemeSpec("named-test", spec)

	if got.Editor.Fg != term.ColorWhite {
		t.Errorf("Editor.Fg = %q, want white", got.Editor.Fg)
	}
	if got.Editor.Bg != term.ColorBlack {
		t.Errorf("Editor.Bg = %q, want black", got.Editor.Bg)
	}
	if got.Status.Fg != term.ColorCyan {
		t.Errorf("Status.Fg = %q, want cyan", got.Status.Fg)
	}
	if got.Status.Bg != term.ColorBlue {
		t.Errorf("Status.Bg = %q, want blue", got.Status.Bg)
	}
	if got.Menu.Fg != term.ColorYellow {
		t.Errorf("Menu.Fg = %q, want yellow", got.Menu.Fg)
	}
	if got.Menu.Bg != term.ColorMagenta {
		t.Errorf("Menu.Bg = %q, want magenta", got.Menu.Bg)
	}
}

// TestConvertThemeSpecDefaultKeyword verifies "default" maps to ColorDefault.
func TestConvertThemeSpecDefaultKeyword(t *testing.T) {
	spec := ThemeSpec{
		Editor: EditorThemeSpec{Fg: "default", Bg: "default"},
	}
	got := ConvertThemeSpec("default-kw", spec)

	if got.Editor.Fg != term.ColorDefault {
		t.Errorf("Editor.Fg = %q, want ColorDefault", got.Editor.Fg)
	}
	if got.Editor.Bg != term.ColorDefault {
		t.Errorf("Editor.Bg = %q, want ColorDefault", got.Editor.Bg)
	}
}

// TestConvertThemeSpecMenuColors verifies all menu color fields are converted.
func TestConvertThemeSpecMenuColors(t *testing.T) {
	spec := ThemeSpec{
		Menu: MenuThemeSpec{
			Fg:            "#111111",
			Bg:            "#222222",
			SelectedFg:    "#333333",
			SelectedBg:    "#444444",
			DropdownFg:    "#555555",
			DropdownBg:    "#666666",
			DropdownSelFg: "#777777",
			DropdownSelBg: "#888888",
			AcceleratorFg: "#999999",
		},
	}

	got := ConvertThemeSpec("menu-test", spec)

	if got.Menu.Fg != term.Color("#111111") {
		t.Errorf("Menu.Fg = %q, want #111111", got.Menu.Fg)
	}
	if got.Menu.DropdownSelBg != term.Color("#888888") {
		t.Errorf("Menu.DropdownSelBg = %q, want #888888", got.Menu.DropdownSelBg)
	}
	if got.Menu.AcceleratorFg != term.Color("#999999") {
		t.Errorf("Menu.AcceleratorFg = %q, want #999999", got.Menu.AcceleratorFg)
	}
}

// TestConvertThemeSpecPromptAndHelp verifies prompt and help colors are converted.
func TestConvertThemeSpecPromptAndHelp(t *testing.T) {
	spec := ThemeSpec{
		Prompt: PromptThemeSpec{Fg: "green", Bg: "black", LabelFg: "cyan", InputFg: "white"},
		Help:   HelpThemeSpec{Fg: "white", Bg: "blue", TitleFg: "yellow", TitleBg: "blue", FooterFg: "cyan"},
	}

	got := ConvertThemeSpec("ph-test", spec)

	if got.Prompt.Fg != term.ColorGreen {
		t.Errorf("Prompt.Fg = %q, want green", got.Prompt.Fg)
	}
	if got.Prompt.LabelFg != term.ColorCyan {
		t.Errorf("Prompt.LabelFg = %q, want cyan", got.Prompt.LabelFg)
	}
	if got.Help.TitleFg != term.ColorYellow {
		t.Errorf("Help.TitleFg = %q, want yellow", got.Help.TitleFg)
	}
	if got.Help.FooterFg != term.ColorCyan {
		t.Errorf("Help.FooterFg = %q, want cyan", got.Help.FooterFg)
	}
}

// TestConvertThemeSpecMessageColors verifies message colors are converted.
func TestConvertThemeSpecMessageColors(t *testing.T) {
	spec := ThemeSpec{
		Msg: MsgThemeSpec{
			InfoFg:    "#AABBCC",
			InfoBg:    "#001122",
			WarningFg: "yellow",
			WarningBg: "#333333",
			ErrorFg:   "red",
			ErrorBg:   "black",
		},
	}

	got := ConvertThemeSpec("msg-test", spec)

	if got.Msg.InfoFg != term.Color("#AABBCC") {
		t.Errorf("Msg.InfoFg = %q, want #AABBCC", got.Msg.InfoFg)
	}
	if got.Msg.WarningFg != term.ColorYellow {
		t.Errorf("Msg.WarningFg = %q, want yellow", got.Msg.WarningFg)
	}
	if got.Msg.ErrorFg != term.ColorRed {
		t.Errorf("Msg.ErrorFg = %q, want red", got.Msg.ErrorFg)
	}
}

// --- Config.GetTheme ---

// TestGetThemeBuiltin verifies built-in themes are returned by name.
func TestGetThemeBuiltin(t *testing.T) {
	cfg := Default()

	for _, name := range []string{"default", "dark", "light", "monokai", "dracula", "nord", "dos"} {
		t.Run(name, func(t *testing.T) {
			got := cfg.GetTheme(name)
			if got == nil {
				t.Fatalf("GetTheme(%q) returned nil", name)
			}
			if got.Name != name {
				t.Errorf("GetTheme(%q).Name = %q", name, got.Name)
			}
		})
	}
}

// TestGetThemeFallback verifies unknown names return the default theme.
func TestGetThemeFallback(t *testing.T) {
	cfg := Default()
	got := cfg.GetTheme("nonexistent-theme")
	if got == nil {
		t.Fatal("GetTheme(nonexistent) returned nil")
	}
	if got.Name != "default" {
		t.Errorf("GetTheme(nonexistent).Name = %q, want default", got.Name)
	}
}

// TestGetThemeCustom verifies a custom theme in Config.Themes is loaded.
func TestGetThemeCustom(t *testing.T) {
	cfg := Default()
	cfg.Themes["my-custom"] = ThemeSpec{
		Editor: EditorThemeSpec{Fg: "#AABBCC", Bg: "#112233"},
		Status: StatusThemeSpec{Fg: "white", Bg: "blue"},
	}

	got := cfg.GetTheme("my-custom")
	if got == nil {
		t.Fatal("GetTheme(my-custom) returned nil")
	}
	if got.Name != "my-custom" {
		t.Errorf("Name = %q, want my-custom", got.Name)
	}
	if got.Editor.Fg != term.Color("#AABBCC") {
		t.Errorf("Editor.Fg = %q, want #AABBCC", got.Editor.Fg)
	}
	if got.Editor.Bg != term.Color("#112233") {
		t.Errorf("Editor.Bg = %q, want #112233", got.Editor.Bg)
	}
	if got.Status.Fg != term.ColorWhite {
		t.Errorf("Status.Fg = %q, want white", got.Status.Fg)
	}
}

// TestGetThemeCustomOverridesBuiltin verifies a custom theme with the same name
// as a built-in is NOT allowed to shadow the built-in (built-ins take priority).
func TestGetThemeCustomBuiltinPriority(t *testing.T) {
	cfg := Default()
	// Try to shadow the "dark" built-in with a custom theme
	cfg.Themes["dark"] = ThemeSpec{
		Editor: EditorThemeSpec{Fg: "#000001"},
	}

	got := cfg.GetTheme("dark")
	// Built-in should win: it won't have the overridden fg
	if got.Editor.Fg == term.Color("#000001") {
		t.Error("Custom theme should not shadow built-in theme")
	}
}

// --- Config.GetCurrentTheme ---

// TestGetCurrentTheme verifies the active theme is returned.
func TestGetCurrentTheme(t *testing.T) {
	cfg := Default()
	cfg.UI.Theme = "monokai"

	got := cfg.GetCurrentTheme()
	if got == nil {
		t.Fatal("GetCurrentTheme() returned nil")
	}
	if got.Name != "monokai" {
		t.Errorf("Name = %q, want monokai", got.Name)
	}
}

// TestGetCurrentThemeEmptyFallsBackToDefault verifies empty UI.Theme uses "default".
func TestGetCurrentThemeEmptyFallsBackToDefault(t *testing.T) {
	cfg := Default()
	cfg.UI.Theme = ""

	got := cfg.GetCurrentTheme()
	if got == nil {
		t.Fatal("GetCurrentTheme() returned nil")
	}
	if got.Name != "default" {
		t.Errorf("Name = %q, want default", got.Name)
	}
}

// TestGetCurrentThemeUnknownFallsBackToDefault verifies an unknown UI.Theme uses "default".
func TestGetCurrentThemeUnknownFallsBackToDefault(t *testing.T) {
	cfg := Default()
	cfg.UI.Theme = "does-not-exist"

	got := cfg.GetCurrentTheme()
	if got == nil {
		t.Fatal("GetCurrentTheme() returned nil")
	}
	if got.Name != "default" {
		t.Errorf("Name = %q, want default", got.Name)
	}
}

// --- Config.GetAvailableThemes ---

// TestGetAvailableThemesBuiltinOnly verifies 14 themes when no custom themes exist.
func TestGetAvailableThemesBuiltinOnly(t *testing.T) {
	cfg := Default()
	themes := cfg.GetAvailableThemes()
	if len(themes) != 14 {
		t.Errorf("len = %d, want 14", len(themes))
	}
}

// TestGetAvailableThemesWithCustom verifies custom themes are included.
func TestGetAvailableThemesWithCustom(t *testing.T) {
	cfg := Default()
	cfg.Themes["custom-a"] = ThemeSpec{}
	cfg.Themes["custom-b"] = ThemeSpec{}

	themes := cfg.GetAvailableThemes()
	if len(themes) != 16 {
		t.Errorf("len = %d, want 16", len(themes))
	}

	found := map[string]bool{}
	for _, name := range themes {
		found[name] = true
	}
	if !found["custom-a"] {
		t.Error("custom-a missing from available themes")
	}
	if !found["custom-b"] {
		t.Error("custom-b missing from available themes")
	}
}

// --- Theme switching (persistence round-trip) ---

// TestThemeSwitchPersistence verifies that changing UI.Theme, saving, and
// reloading the config file restores the chosen theme.
func TestThemeSwitchPersistence(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.toml")

	origConfigPath := ConfigPath
	defer func() { ConfigPath = origConfigPath }()
	ConfigPath = func() (string, error) { return configFile, nil }

	// Start with default config, switch to "dracula"
	cfg := Default()
	cfg.UI.Theme = "dracula"
	if err := Save(cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.UI.Theme != "dracula" {
		t.Errorf("UI.Theme = %q, want dracula", loaded.UI.Theme)
	}

	got := loaded.GetCurrentTheme()
	if got.Name != "dracula" {
		t.Errorf("GetCurrentTheme().Name = %q, want dracula", got.Name)
	}
}

// TestThemeSwitchMultipleTimes verifies repeated theme switches persist correctly.
func TestThemeSwitchMultipleTimes(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.toml")

	origConfigPath := ConfigPath
	defer func() { ConfigPath = origConfigPath }()
	ConfigPath = func() (string, error) { return configFile, nil }

	cfg := Default()
	for _, name := range []string{"nord", "monokai", "gruvbox-dark", "dos"} {
		cfg.UI.Theme = name
		if err := Save(cfg); err != nil {
			t.Fatalf("Save(%s): %v", name, err)
		}
		loaded, err := Load()
		if err != nil {
			t.Fatalf("Load after %s: %v", name, err)
		}
		if loaded.UI.Theme != name {
			t.Errorf("after switch to %s: UI.Theme = %q", name, loaded.UI.Theme)
		}
	}
}

// TestCustomThemeRoundTrip verifies a custom theme survives a config save/load cycle.
func TestCustomThemeRoundTrip(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.toml")

	origConfigPath := ConfigPath
	defer func() { ConfigPath = origConfigPath }()
	ConfigPath = func() (string, error) { return configFile, nil }

	cfg := Default()
	cfg.UI.Theme = "ocean"
	cfg.Themes["ocean"] = ThemeSpec{
		Editor: EditorThemeSpec{Fg: "#E0E0FF", Bg: "#001030"},
		Status: StatusThemeSpec{Fg: "#AADDFF", Bg: "#000820"},
		Menu:   MenuThemeSpec{Fg: "#FFFFFF", Bg: "#003060"},
	}

	if err := Save(cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	spec, ok := loaded.Themes["ocean"]
	if !ok {
		t.Fatal("custom theme 'ocean' missing after load")
	}
	if spec.Editor.Fg != "#E0E0FF" {
		t.Errorf("Editor.Fg = %q, want #E0E0FF", spec.Editor.Fg)
	}
	if spec.Editor.Bg != "#001030" {
		t.Errorf("Editor.Bg = %q, want #001030", spec.Editor.Bg)
	}

	got := loaded.GetTheme("ocean")
	if got.Name != "ocean" {
		t.Errorf("GetTheme(ocean).Name = %q", got.Name)
	}
	if got.Editor.Fg != term.Color("#E0E0FF") {
		t.Errorf("GetTheme(ocean).Editor.Fg = %q, want #E0E0FF", got.Editor.Fg)
	}
	if got.Status.Bg != term.Color("#000820") {
		t.Errorf("GetTheme(ocean).Status.Bg = %q, want #000820", got.Status.Bg)
	}
}

// TestThemeConfigFileContent verifies the saved TOML contains the theme name.
func TestThemeConfigFileContent(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.toml")

	origConfigPath := ConfigPath
	defer func() { ConfigPath = origConfigPath }()
	ConfigPath = func() (string, error) { return configFile, nil }

	cfg := Default()
	cfg.UI.Theme = "cyberpunk"
	if err := Save(cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	content := string(data)
	if !containsSubstring(content, "cyberpunk") {
		t.Error("saved config file does not contain theme name 'cyberpunk'")
	}
}

func containsSubstring(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && findSubstring(s, sub))
}

func findSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
