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

import (
	"testing"

	"cooledit/internal/term"
)

// TestParseColorNamed tests parsing of named colors
func TestParseColorNamed(t *testing.T) {
	tests := []struct {
		input    string
		expected term.Color
	}{
		{"black", term.ColorBlack},
		{"red", term.ColorRed},
		{"green", term.ColorGreen},
		{"yellow", term.ColorYellow},
		{"blue", term.ColorBlue},
		{"magenta", term.ColorMagenta},
		{"cyan", term.ColorCyan},
		{"white", term.ColorWhite},
	}

	for _, tc := range tests {
		result := ParseColor(tc.input)
		if result != tc.expected {
			t.Errorf("ParseColor(%q) = %q, expected %q", tc.input, result, tc.expected)
		}
	}
}

// TestParseColorHex tests parsing of hex colors
func TestParseColorHex(t *testing.T) {
	tests := []struct {
		input    string
		expected term.Color
	}{
		{"#FF0000", term.Color("#FF0000")},
		{"#00FF00", term.Color("#00FF00")},
		{"#0000FF", term.Color("#0000FF")},
		{"#FFFFFF", term.Color("#FFFFFF")},
		{"#000000", term.Color("#000000")},
		{"#282828", term.Color("#282828")},
		{"#EBDBB2", term.Color("#EBDBB2")},
	}

	for _, tc := range tests {
		result := ParseColor(tc.input)
		if result != tc.expected {
			t.Errorf("ParseColor(%q) = %q, expected %q", tc.input, result, tc.expected)
		}
	}
}

// TestParseColorDefault tests parsing of default and empty colors
func TestParseColorDefault(t *testing.T) {
	if ParseColor("") != term.ColorDefault {
		t.Error("Empty string should return ColorDefault")
	}
	if ParseColor("default") != term.ColorDefault {
		t.Error("\"default\" should return ColorDefault")
	}
}

// TestParseColorInvalid tests parsing of invalid/unknown colors
func TestParseColorInvalid(t *testing.T) {
	// Invalid colors should be passed through as-is
	result := ParseColor("invalid")
	if result != term.Color("invalid") {
		t.Errorf("Invalid color should be passed through, got %q", result)
	}

	// Short hex (invalid format) should be passed through
	result = ParseColor("#FFF")
	if result != term.Color("#FFF") {
		t.Errorf("Short hex should be passed through, got %q", result)
	}
}

// TestBuiltinThemesExist verifies all 13 built-in themes exist
func TestBuiltinThemesExist(t *testing.T) {
	expectedThemes := []string{
		"default", "dark", "light", "monokai",
		"solarized-dark", "solarized-light",
		"gruvbox-dark", "gruvbox-light",
		"dracula", "nord", "dos",
		"ibm-green", "ibm-amber", "cyberpunk",
	}

	if len(BuiltinThemes) != len(expectedThemes) {
		t.Errorf("Expected %d built-in themes, got %d", len(expectedThemes), len(BuiltinThemes))
	}

	for _, name := range expectedThemes {
		if _, ok := BuiltinThemes[name]; !ok {
			t.Errorf("Missing built-in theme: %s", name)
		}
	}
}

// TestBuiltinThemesNotNil verifies all themes are non-nil
func TestBuiltinThemesNotNil(t *testing.T) {
	for name, theme := range BuiltinThemes {
		if theme == nil {
			t.Errorf("Theme %q is nil", name)
		}
	}
}

// TestBuiltinThemeNameMatch verifies theme.Name matches the key
func TestBuiltinThemeNameMatch(t *testing.T) {
	for name, theme := range BuiltinThemes {
		if theme.Name != name {
			t.Errorf("Theme key %q has mismatched Name field: %q", name, theme.Name)
		}
	}
}

// TestListBuiltinThemes verifies the list function
func TestListBuiltinThemes(t *testing.T) {
	list := ListBuiltinThemes()

	if len(list) != 14 {
		t.Errorf("Expected 14 themes in list, got %d", len(list))
	}

	// First should be "default"
	if list[0] != "default" {
		t.Errorf("First theme should be 'default', got %s", list[0])
	}

	// Verify all themes in list exist
	for _, name := range list {
		if _, ok := BuiltinThemes[name]; !ok {
			t.Errorf("ListBuiltinThemes contains unknown theme: %s", name)
		}
	}
}

// TestGetBuiltinTheme tests theme retrieval
func TestGetBuiltinTheme(t *testing.T) {
	// Test retrieving existing theme
	theme := GetBuiltinTheme("dark")
	if theme == nil {
		t.Fatal("GetBuiltinTheme(\"dark\") returned nil")
	}
	if theme.Name != "dark" {
		t.Errorf("Expected 'dark' theme, got %s", theme.Name)
	}

	// Test retrieving another theme
	theme = GetBuiltinTheme("monokai")
	if theme == nil {
		t.Fatal("GetBuiltinTheme(\"monokai\") returned nil")
	}
	if theme.Name != "monokai" {
		t.Errorf("Expected 'monokai' theme, got %s", theme.Name)
	}
}

// TestGetBuiltinThemeFallback tests fallback to default theme
func TestGetBuiltinThemeFallback(t *testing.T) {
	// Unknown theme should return default
	theme := GetBuiltinTheme("nonexistent")
	if theme == nil {
		t.Fatal("GetBuiltinTheme(\"nonexistent\") returned nil")
	}
	if theme.Name != "default" {
		t.Errorf("Unknown theme should return 'default', got %s", theme.Name)
	}

	// Empty string should return default
	theme = GetBuiltinTheme("")
	if theme == nil {
		t.Fatal("GetBuiltinTheme(\"\") returned nil")
	}
	if theme.Name != "default" {
		t.Errorf("Empty theme name should return 'default', got %s", theme.Name)
	}
}

// TestGetStyle tests style creation
func TestGetStyle(t *testing.T) {
	style := GetStyle(term.ColorRed, term.ColorBlue)

	if style.Foreground != term.ColorRed {
		t.Errorf("Foreground = %q, expected %q", style.Foreground, term.ColorRed)
	}
	if style.Background != term.ColorBlue {
		t.Errorf("Background = %q, expected %q", style.Background, term.ColorBlue)
	}
	if style.Inverse {
		t.Error("Inverse should be false by default")
	}
}

// TestGetInverseStyle tests inverse style creation
func TestGetInverseStyle(t *testing.T) {
	style := GetInverseStyle()

	if !style.Inverse {
		t.Error("Inverse should be true")
	}
}

// TestLoadThemeBuiltin tests loading built-in themes
func TestLoadThemeBuiltin(t *testing.T) {
	theme := LoadTheme("dark", nil)
	if theme == nil {
		t.Fatal("LoadTheme(\"dark\", nil) returned nil")
	}
	if theme.Name != "dark" {
		t.Errorf("Expected 'dark' theme, got %s", theme.Name)
	}
}

// TestLoadThemeFallback tests fallback behavior
func TestLoadThemeFallback(t *testing.T) {
	theme := LoadTheme("nonexistent", nil)
	if theme == nil {
		t.Fatal("LoadTheme(\"nonexistent\", nil) returned nil")
	}
	if theme.Name != "default" {
		t.Errorf("Unknown theme should return 'default', got %s", theme.Name)
	}
}

// TestLoadThemeEmptyCustom tests with empty custom themes
func TestLoadThemeEmptyCustom(t *testing.T) {
	customThemes := map[string]ConfigThemeSpec{}
	theme := LoadTheme("monokai", customThemes)
	if theme == nil {
		t.Fatal("LoadTheme returned nil")
	}
	if theme.Name != "monokai" {
		t.Errorf("Expected 'monokai' theme, got %s", theme.Name)
	}
}

// TestGetAvailableThemes tests available themes list
func TestGetAvailableThemes(t *testing.T) {
	themes := GetAvailableThemes(nil)
	if len(themes) != 14 {
		t.Errorf("Expected 14 themes, got %d", len(themes))
	}
}

// TestDefaultThemeStructure verifies default theme has all fields
func TestDefaultThemeStructure(t *testing.T) {
	theme := BuiltinThemes["default"]
	if theme == nil {
		t.Fatal("Default theme is nil")
	}

	// Default theme uses ColorDefault for most colors
	if theme.Editor.Fg != term.ColorDefault {
		t.Errorf("Default theme Editor.Fg should be ColorDefault, got %q", theme.Editor.Fg)
	}
	if theme.Editor.Bg != term.ColorDefault {
		t.Errorf("Default theme Editor.Bg should be ColorDefault, got %q", theme.Editor.Bg)
	}

	// But cursor should have a specific color
	if theme.Editor.CursorColor == "" || theme.Editor.CursorColor == term.ColorDefault {
		t.Error("Default theme should have a specific cursor color")
	}
}

// TestDarkThemeHasColors verifies dark theme has specific colors
func TestDarkThemeHasColors(t *testing.T) {
	theme := BuiltinThemes["dark"]
	if theme == nil {
		t.Fatal("Dark theme is nil")
	}

	// Dark theme should have non-default colors
	if theme.Editor.Fg == term.ColorDefault {
		t.Error("Dark theme should have specific Editor.Fg color")
	}
	if theme.Editor.Bg == term.ColorDefault {
		t.Error("Dark theme should have specific Editor.Bg color")
	}
	if theme.Status.Bg == term.ColorDefault {
		t.Error("Dark theme should have specific Status.Bg color")
	}
}

// TestAllThemesHaveRequiredFields verifies all themes are complete
func TestAllThemesHaveRequiredFields(t *testing.T) {
	for name, theme := range BuiltinThemes {
		if theme.Name == "" {
			t.Errorf("Theme %q has empty Name", name)
		}

		// Skip default theme which uses ColorDefault
		if name == "default" {
			continue
		}

		// Non-default themes should have specific colors
		if theme.Editor.Fg == "" {
			t.Errorf("Theme %q has empty Editor.Fg", name)
		}
		if theme.Editor.Bg == "" {
			t.Errorf("Theme %q has empty Editor.Bg", name)
		}
		if theme.Status.Fg == "" {
			t.Errorf("Theme %q has empty Status.Fg", name)
		}
		if theme.Status.Bg == "" {
			t.Errorf("Theme %q has empty Status.Bg", name)
		}
		if theme.Menu.Fg == "" {
			t.Errorf("Theme %q has empty Menu.Fg", name)
		}
		if theme.Menu.Bg == "" {
			t.Errorf("Theme %q has empty Menu.Bg", name)
		}
	}
}

// TestRetroThemesAuthentic verifies IBM themes have authentic phosphor colors
func TestRetroThemesAuthentic(t *testing.T) {
	// IBM Green should use green shades
	green := BuiltinThemes["ibm-green"]
	if green == nil {
		t.Fatal("ibm-green theme is nil")
	}
	// Should have green foreground on black background
	if green.Editor.Bg != "#000000" {
		t.Errorf("IBM green should have black background, got %q", green.Editor.Bg)
	}

	// IBM Amber should use amber/orange shades
	amber := BuiltinThemes["ibm-amber"]
	if amber == nil {
		t.Fatal("ibm-amber theme is nil")
	}
	// Should have amber foreground on black background
	if amber.Editor.Bg != "#000000" {
		t.Errorf("IBM amber should have black background, got %q", amber.Editor.Bg)
	}
}

// TestDosThemeClassicColors verifies DOS theme has classic blue background
func TestDosThemeClassicColors(t *testing.T) {
	dos := BuiltinThemes["dos"]
	if dos == nil {
		t.Fatal("dos theme is nil")
	}

	// DOS Edit had blue background
	if dos.Editor.Bg != "#0000AA" {
		t.Errorf("DOS theme should have classic blue background (#0000AA), got %q", dos.Editor.Bg)
	}

	// And white text
	if dos.Editor.Fg != "#FFFFFF" {
		t.Errorf("DOS theme should have white foreground, got %q", dos.Editor.Fg)
	}
}
