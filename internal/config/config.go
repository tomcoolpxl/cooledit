package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// ConfigPath returns the path to the configuration file (can be overridden in tests)
var ConfigPath = func() (string, error) {
	// Get user's config directory
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	// Create cooledit config directory if it doesn't exist
	coolEditDir := filepath.Join(configDir, "cooledit")
	if err := os.MkdirAll(coolEditDir, 0755); err != nil {
		return "", err
	}

	return filepath.Join(coolEditDir, "config.toml"), nil
}

// Load reads the configuration file. If it doesn't exist, returns default config.
func Load() (*Config, error) {
	path, err := ConfigPath()
	if err != nil {
		return Default(), nil // Fallback to defaults if can't get config path
	}

	return LoadFrom(path)
}

// LoadFrom reads configuration from a specific file path
func LoadFrom(path string) (*Config, error) {
	// If config file doesn't exist, return defaults
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return Default(), nil
	}

	// Read and parse config file
	var cfg Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return nil, err
	}

	// Apply defaults for any missing values (in case file is partial)
	applyDefaults(&cfg)

	return &cfg, nil
}

// Save writes the configuration to the config file
func Save(cfg *Config) error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	// Open file for writing
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	// Write header comment
	f.WriteString("# cooledit Configuration\n")
	f.WriteString("# This file is automatically updated when you change settings\n\n")

	// Encode config to TOML
	encoder := toml.NewEncoder(f)
	return encoder.Encode(cfg)
}

// applyDefaults fills in any zero values with defaults
func applyDefaults(cfg *Config) {
	defaults := Default()

	// Apply Editor defaults
	if cfg.Editor.TabWidth == 0 {
		cfg.Editor.TabWidth = defaults.Editor.TabWidth
	}
}
