package theme

import (
	"cooledit/internal/term"
)

// ConfigThemeSpec is an interface to avoid circular dependencies
// It matches config.ThemeSpec structure
type ConfigThemeSpec interface{}

// LoadTheme loads a theme by name, checking built-in themes first, then custom themes from config
func LoadTheme(name string, customThemes map[string]ConfigThemeSpec) *Theme {
	// Check built-in themes first
	if theme, ok := BuiltinThemes[name]; ok {
		return theme
	}
	
	// Check custom themes from config
	if spec, ok := customThemes[name]; ok {
		if customTheme := convertConfigToTheme(name, spec); customTheme != nil {
			return customTheme
		}
	}
	
	// Fallback to default
	return BuiltinThemes["default"]
}

// convertConfigToTheme converts a config theme spec to a Theme
// This uses type assertion to access the config.ThemeSpec fields
func convertConfigToTheme(name string, spec ConfigThemeSpec) *Theme {
	// For now, return nil to avoid circular dependency
	// This will be properly implemented when integrating with config
	return nil
}

// ParseColor converts a color string to term.Color
// Supports: named colors, hex (#RRGGBB), or "default"
func ParseColor(s string) term.Color {
	if s == "" || s == "default" {
		return term.ColorDefault
	}
	
	// Validate hex color format
	if len(s) == 7 && s[0] == '#' {
		return term.Color(s)
	}
	
	// Named colors
	switch s {
	case "black":
		return term.ColorBlack
	case "red":
		return term.ColorRed
	case "green":
		return term.ColorGreen
	case "yellow":
		return term.ColorYellow
	case "blue":
		return term.ColorBlue
	case "magenta":
		return term.ColorMagenta
	case "cyan":
		return term.ColorCyan
	case "white":
		return term.ColorWhite
	default:
		// If not a standard color, treat as hex or custom
		return term.Color(s)
	}
}

// GetAvailableThemes returns all available theme names (built-in + custom)
func GetAvailableThemes(customThemes map[string]ConfigThemeSpec) []string {
	themes := ListBuiltinThemes()
	
	// Add custom theme names
	for name := range customThemes {
		themes = append(themes, name)
	}
	
	return themes
}
