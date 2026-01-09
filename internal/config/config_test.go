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

	"github.com/BurntSushi/toml"
)

func TestDefaultConfig(t *testing.T) {
	cfg := Default()

	if cfg.Editor.LineNumbers {
		t.Error("Expected LineNumbers to be false by default")
	}
	if cfg.Editor.SoftWrap {
		t.Error("Expected SoftWrap to be false by default")
	}
	if cfg.Editor.TabWidth != 4 {
		t.Errorf("Expected TabWidth to be 4, got %d", cfg.Editor.TabWidth)
	}
	if cfg.UI.ShowMenubar {
		t.Error("Expected ShowMenubar to be false by default")
	}
	if !cfg.Search.CaseSensitive {
		t.Error("Expected CaseSensitive to be true by default")
	}
}

func TestSaveAndLoad(t *testing.T) {
	// Create temp directory for test
	tempDir := t.TempDir()

	// Save original ConfigPath and restore after test
	origConfigPath := ConfigPath
	defer func() { ConfigPath = origConfigPath }()

	// Override ConfigPath to use temp directory
	configFile := filepath.Join(tempDir, "config.toml")
	ConfigPath = func() (string, error) {
		return configFile, nil
	}

	// Create custom config
	cfg := Config{
		Editor: Editor{
			LineNumbers: false,
			SoftWrap:    false,
			TabWidth:    8,
		},
		UI: UI{
			ShowMenubar: false,
		},
		Search: Search{
			CaseSensitive: true,
		},
	}

	// Save config
	if err := Save(&cfg); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		t.Fatal("Config file was not created")
	}

	// Load config
	loaded, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify values
	if loaded.Editor.LineNumbers != false {
		t.Error("LineNumbers not loaded correctly")
	}
	if loaded.Editor.SoftWrap != false {
		t.Error("SoftWrap not loaded correctly")
	}
	if loaded.Editor.TabWidth != 8 {
		t.Errorf("TabWidth not loaded correctly, got %d", loaded.Editor.TabWidth)
	}
	if loaded.UI.ShowMenubar != false {
		t.Error("ShowMenubar not loaded correctly")
	}
	if loaded.Search.CaseSensitive != true {
		t.Error("CaseSensitive not loaded correctly")
	}
}

func TestLoadNonExistentFile(t *testing.T) {
	// Save original and restore
	origConfigPath := ConfigPath
	defer func() { ConfigPath = origConfigPath }()

	// Point to non-existent location
	ConfigPath = func() (string, error) {
		return "/nonexistent/path/config.toml", nil
	}

	// Load from non-existent path should return defaults
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load should not error on missing file: %v", err)
	}

	// Should have default values
	defaults := Default()
	if cfg.Editor.LineNumbers != defaults.Editor.LineNumbers {
		t.Error("Should return default LineNumbers when file missing")
	}
	if cfg.Editor.TabWidth != defaults.Editor.TabWidth {
		t.Error("Should return default TabWidth when file missing")
	}
}

func TestPartialConfig(t *testing.T) {
	// Create temp directory
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.toml")

	// Write partial config (only Editor section)
	content := `[editor]
line_numbers = false
tab_width = 2
`
	if err := os.WriteFile(configFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Parse config manually and apply defaults
	var cfg Config
	if _, err := toml.DecodeFile(configFile, &cfg); err != nil {
		t.Fatalf("Failed to decode config: %v", err)
	}
	applyDefaults(&cfg)

	// Check loaded values
	if cfg.Editor.LineNumbers != false {
		t.Error("LineNumbers should be false from config")
	}
	if cfg.Editor.TabWidth != 2 {
		t.Error("TabWidth should be 2 from config")
	}

	// Check default values for missing fields
	defaults := Default()
	if cfg.Editor.SoftWrap != defaults.Editor.SoftWrap {
		t.Error("SoftWrap should use default when not in config")
	}
}
