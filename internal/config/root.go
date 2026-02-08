// Package config provides configuration and initialization management for smoke.
// It handles directory paths, feed storage, and smoke initialization state.
package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// ErrNotInitialized is returned when smoke hasn't been initialized
var ErrNotInitialized = errors.New("smoke not initialized. Run 'smoke init' first")

// GetConfigDir returns the path to the smoke config directory (~/.config/smoke/)
func GetConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", DefaultSmokeDir), nil
}

// ErrInvalidFeedPath is returned when SMOKE_FEED path is outside allowed directories
var ErrInvalidFeedPath = errors.New("SMOKE_FEED path must be within home directory")

// resolveHomePaths returns the home directory and its symlink-resolved form.
func resolveHomePaths() (home, resolvedHome string, err error) {
	home, err = os.UserHomeDir()
	if err != nil {
		return "", "", err
	}
	// Resolve symlinks to handle /var -> /private/var on macOS
	resolvedHome, err = filepath.EvalSymlinks(home)
	if err != nil {
		resolvedHome = home
	}
	return home, resolvedHome, nil
}

// isPathInHome checks whether cleanPath or resolvedPath falls within the home directory.
func isPathInHome(cleanPath, resolvedPath, home, resolvedHome string) bool {
	homePrefix := home + string(filepath.Separator)
	resolvedHomePrefix := resolvedHome + string(filepath.Separator)
	return strings.HasPrefix(cleanPath, homePrefix) || cleanPath == home ||
		strings.HasPrefix(resolvedPath, resolvedHomePrefix) || resolvedPath == resolvedHome
}

// isPathInTemp checks whether the path falls within allowed temporary directories.
func isPathInTemp(cleanPath string) bool {
	if strings.HasPrefix(cleanPath, "/tmp/") ||
		strings.HasPrefix(cleanPath, "/var/folders/") ||
		strings.HasPrefix(cleanPath, "/private/tmp/") ||
		strings.HasPrefix(cleanPath, "/private/var/folders/") {
		return true
	}
	if tmpDir := os.Getenv("TMPDIR"); tmpDir != "" {
		return strings.HasPrefix(cleanPath, tmpDir)
	}
	return false
}

// validateFeedPath ensures the path is safe (absolute, within allowed directories)
// Allowed: home directory, temp directories (/tmp, $TMPDIR, /var/folders)
func validateFeedPath(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	cleanPath := filepath.Clean(absPath)

	home, resolvedHome, err := resolveHomePaths()
	if err != nil {
		return "", err
	}

	// Also try to resolve the path's parent for comparison
	resolvedPath := cleanPath
	if parentDir := filepath.Dir(cleanPath); parentDir != cleanPath {
		if resolved, resolveErr := filepath.EvalSymlinks(parentDir); resolveErr == nil {
			resolvedPath = filepath.Join(resolved, filepath.Base(cleanPath))
		}
	}

	if !isPathInHome(cleanPath, resolvedPath, home, resolvedHome) && !isPathInTemp(cleanPath) {
		return "", ErrInvalidFeedPath
	}
	return cleanPath, nil
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
	return filepath.Join(configDir, DefaultFeedFile), nil
}

// GetConfigPath returns the path to the config.yaml file
func GetConfigPath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, DefaultConfigFile), nil
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

// GetLogPath returns the path to the smoke.log file
func GetLogPath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, DefaultLogFile), nil
}
