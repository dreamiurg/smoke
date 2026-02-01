package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
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

// ErrInvalidFeedPath is returned when SMOKE_FEED path is outside allowed directories
var ErrInvalidFeedPath = errors.New("SMOKE_FEED path must be within home directory")

// validateFeedPath ensures the path is safe (absolute, within home directory)
func validateFeedPath(path string) (string, error) {
	// Resolve to absolute path and clean it
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	cleanPath := filepath.Clean(absPath)

	// Get home directory and resolve symlinks for consistent comparison
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	// Resolve symlinks on home to handle /var -> /private/var on macOS
	resolvedHome, err := filepath.EvalSymlinks(home)
	if err != nil {
		resolvedHome = home // Fall back if home doesn't exist or can't resolve
	}

	// Also try to resolve the path's parent for comparison
	resolvedPath := cleanPath
	if parentDir := filepath.Dir(cleanPath); parentDir != cleanPath {
		if resolved, resolveErr := filepath.EvalSymlinks(parentDir); resolveErr == nil {
			resolvedPath = filepath.Join(resolved, filepath.Base(cleanPath))
		}
	}

	// Check if path is within home directory (using both resolved and unresolved)
	homePrefix := home + string(filepath.Separator)
	resolvedHomePrefix := resolvedHome + string(filepath.Separator)

	inHome := strings.HasPrefix(cleanPath, homePrefix) || cleanPath == home ||
		strings.HasPrefix(resolvedPath, resolvedHomePrefix) || resolvedPath == resolvedHome

	if !inHome {
		return "", ErrInvalidFeedPath
	}
	return absPath, nil
}

// GetFeedPath returns the path to the feed.jsonl file
// If SMOKE_FEED env var is set, uses that path after validation (must be within home directory)
func GetFeedPath() (string, error) {
	// Check for explicit feed path override
	if feedPath := os.Getenv("SMOKE_FEED"); feedPath != "" {
		return validateFeedPath(feedPath)
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

// ensureInitialized returns an error if smoke is not initialized
func ensureInitialized() error {
	initialized, err := IsSmokeInitialized()
	if err != nil {
		return err
	}
	if !initialized {
		return ErrNotInitialized
	}
	return nil
}
