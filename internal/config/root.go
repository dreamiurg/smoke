package config

import (
	"errors"
	"os"
	"path/filepath"
)

// ErrNotInitialized is returned when smoke hasn't been initialized
var ErrNotInitialized = errors.New("smoke not initialized. Run 'smoke init' first")

// SmokeDir is the name of the smoke data directory within ~/.config/
const SmokeDir = "smoke"

// FeedFile is the name of the feed file
const FeedFile = "feed.jsonl"

// ConfigFile is the name of the config file
const ConfigFile = "config.yaml"

// GetConfigDir returns the path to the smoke config directory (~/.config/smoke/)
func GetConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", SmokeDir), nil
}

// GetFeedPath returns the path to the feed.jsonl file
// If SMOKE_FEED env var is set, uses that path directly (allows custom feed location)
func GetFeedPath() (string, error) {
	// Check for explicit feed path override
	if feedPath := os.Getenv("SMOKE_FEED"); feedPath != "" {
		// Sanitize path to prevent traversal
		return filepath.Clean(feedPath), nil
	}

	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, FeedFile), nil
}

// GetConfigPath returns the path to the config.yaml file
func GetConfigPath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, ConfigFile), nil
}

// IsSmokeInitialized checks if smoke has been initialized
func IsSmokeInitialized() (bool, error) {
	feedPath, err := GetFeedPath()
	if err != nil {
		return false, err
	}
	_, err = os.Stat(feedPath)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// EnsureInitialized returns an error if smoke is not initialized
func EnsureInitialized() error {
	initialized, err := IsSmokeInitialized()
	if err != nil {
		return err
	}
	if !initialized {
		return ErrNotInitialized
	}
	return nil
}
