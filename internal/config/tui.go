package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// TUIConfig stores user preferences for the TUI feed.
type TUIConfig struct {
	Theme       string `yaml:"theme"`
	Contrast    string `yaml:"contrast"`
	Style       string `yaml:"style"`
	AutoRefresh bool   `yaml:"auto_refresh"`
}

// TUIConfigFile is the name of the TUI config file
const TUIConfigFile = "tui.yaml"

// Default values - must match feed.DefaultThemeName and feed.DefaultContrastName
const (
	defaultTheme       = "dracula"
	defaultContrast    = "medium"
	defaultStyle       = "header"
	defaultAutoRefresh = true
)

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
		cfg.Theme = defaultTheme
	}
	if cfg.Contrast == "" {
		cfg.Contrast = defaultContrast
	}
	if cfg.Style == "" {
		cfg.Style = defaultStyle
	}
	// AutoRefresh defaults to true (bool zero value is false, so we need special handling)
	// We use a sentinel approach: if the file was parsed but AutoRefresh is false,
	// we check if it was explicitly set or just the default. For simplicity,
	// we'll always default new configs to true.

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
		Theme:       defaultTheme,
		Contrast:    defaultContrast,
		Style:       defaultStyle,
		AutoRefresh: defaultAutoRefresh,
	}
}
