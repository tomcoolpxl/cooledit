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
	"cooledit/internal/term"
	"cooledit/internal/theme"
)

// ConvertThemeSpec converts a config ThemeSpec to a theme.Theme
func ConvertThemeSpec(name string, spec ThemeSpec) *theme.Theme {
	return &theme.Theme{
		Name: name,
		Editor: theme.EditorColors{
			Fg:            parseColorField(spec.Editor.Fg),
			Bg:            parseColorField(spec.Editor.Bg),
			SelectionFg:   parseColorField(spec.Editor.SelectionFg),
			SelectionBg:   parseColorField(spec.Editor.SelectionBg),
			LineNumbersFg: parseColorField(spec.Editor.LineNumbersFg),
			LineNumbersBg: parseColorField(spec.Editor.LineNumbersBg),
			CursorColor:   parseColorField(spec.Editor.CursorColor),
		},
		Search: theme.SearchColors{
			MatchFg:        parseColorField(spec.Search.MatchFg),
			MatchBg:        parseColorField(spec.Search.MatchBg),
			CurrentMatchFg: parseColorField(spec.Search.CurrentMatchFg),
			CurrentMatchBg: parseColorField(spec.Search.CurrentMatchBg),
		},
		Status: theme.StatusColors{
			Fg:         parseColorField(spec.Status.Fg),
			Bg:         parseColorField(spec.Status.Bg),
			FilenameFg: parseColorField(spec.Status.FilenameFg),
			ModifiedFg: parseColorField(spec.Status.ModifiedFg),
			PositionFg: parseColorField(spec.Status.PositionFg),
			ModeFg:     parseColorField(spec.Status.ModeFg),
			HelpFg:     parseColorField(spec.Status.HelpFg),
		},
		Menu: theme.MenuColors{
			Fg:            parseColorField(spec.Menu.Fg),
			Bg:            parseColorField(spec.Menu.Bg),
			SelectedFg:    parseColorField(spec.Menu.SelectedFg),
			SelectedBg:    parseColorField(spec.Menu.SelectedBg),
			DropdownFg:    parseColorField(spec.Menu.DropdownFg),
			DropdownBg:    parseColorField(spec.Menu.DropdownBg),
			DropdownSelFg: parseColorField(spec.Menu.DropdownSelFg),
			DropdownSelBg: parseColorField(spec.Menu.DropdownSelBg),
			AcceleratorFg: parseColorField(spec.Menu.AcceleratorFg),
		},
		Prompt: theme.PromptColors{
			Fg:      parseColorField(spec.Prompt.Fg),
			Bg:      parseColorField(spec.Prompt.Bg),
			LabelFg: parseColorField(spec.Prompt.LabelFg),
			InputFg: parseColorField(spec.Prompt.InputFg),
		},
		Help: theme.HelpColors{
			Fg:       parseColorField(spec.Help.Fg),
			Bg:       parseColorField(spec.Help.Bg),
			TitleFg:  parseColorField(spec.Help.TitleFg),
			TitleBg:  parseColorField(spec.Help.TitleBg),
			FooterFg: parseColorField(spec.Help.FooterFg),
		},
		Msg: theme.MessageColors{
			InfoFg:    parseColorField(spec.Msg.InfoFg),
			InfoBg:    parseColorField(spec.Msg.InfoBg),
			WarningFg: parseColorField(spec.Msg.WarningFg),
			WarningBg: parseColorField(spec.Msg.WarningBg),
			ErrorFg:   parseColorField(spec.Msg.ErrorFg),
			ErrorBg:   parseColorField(spec.Msg.ErrorBg),
		},
	}
}

// parseColorField converts a string color field to term.Color
func parseColorField(s string) term.Color {
	return theme.ParseColor(s)
}

// GetTheme loads a theme by name from config (built-in or custom)
func (c *Config) GetTheme(name string) *theme.Theme {
	// Try built-in themes first
	if t := theme.GetBuiltinTheme(name); t.Name == name {
		return t
	}

	// Try custom themes from config
	if spec, ok := c.Themes[name]; ok {
		return ConvertThemeSpec(name, spec)
	}

	// Fallback to default
	return theme.GetBuiltinTheme("default")
}

// GetCurrentTheme returns the currently active theme
func (c *Config) GetCurrentTheme() *theme.Theme {
	themeName := c.UI.Theme
	if themeName == "" {
		themeName = "default"
	}
	return c.GetTheme(themeName)
}

// GetAvailableThemes returns all available theme names
func (c *Config) GetAvailableThemes() []string {
	themes := theme.ListBuiltinThemes()

	// Add custom theme names
	for name := range c.Themes {
		themes = append(themes, name)
	}

	return themes
}
