package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// TUIConfig stores user preferences for the TUI feed.
type TUIConfig struct {
	Theme    string `yaml:"theme"`
	Contrast string `yaml:"contrast"`
}

// TUIConfigFile is the name of the TUI config file
const TUIConfigFile = "tui.yaml"

// GetTUIConfigPath returns the path to the tui.yaml file
func GetTUIConfigPath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, TUIConfigFile), nil
}

// LoadTUIConfig loads TUI configuration from disk.
// Returns default config if file doesn't exist, is empty, or is invalid.
// Never returns an error - gracefully handles all failure cases with defaults.
func LoadTUIConfig() *TUIConfig {
	path, err := GetTUIConfigPath()
	if err != nil {
		return defaultTUIConfig()
	}

	data, err := os.ReadFile(path)
	if err != nil {
		// File doesn't exist or can't be read - return defaults
		return defaultTUIConfig()
	}

	// Handle empty file
	if len(data) == 0 {
		return defaultTUIConfig()
	}

	var cfg TUIConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		// JSON/YAML is invalid - return defaults
		return defaultTUIConfig()
	}

	// Apply defaults for empty fields
	if cfg.Theme == "" {
		cfg.Theme = "tomorrow-night"
	}
	if cfg.Contrast == "" {
		cfg.Contrast = "medium"
	}

	return &cfg
}

// SaveTUIConfig saves TUI configuration to disk.
func SaveTUIConfig(cfg *TUIConfig) error {
	path, err := GetTUIConfigPath()
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// defaultTUIConfig returns the default TUI configuration.
func defaultTUIConfig() *TUIConfig {
	return &TUIConfig{
		Theme:    "tomorrow-night",
		Contrast: "medium",
	}
}
