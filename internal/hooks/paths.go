package hooks

import (
	"os"
	"path/filepath"
)

// getHooksDirImpl is the default implementation
func getHooksDirImpl() string {
	home, err := os.UserHomeDir()
	if err != nil {
		// Fallback to HOME env var, then temp dir
		if homeEnv := os.Getenv("HOME"); homeEnv != "" {
			return filepath.Join(homeEnv, ".claude", "hooks")
		}
		return filepath.Join(os.TempDir(), ".claude", "hooks")
	}
	return filepath.Join(home, ".claude", "hooks")
}

// getSettingsPathImpl is the default implementation
func getSettingsPathImpl() string {
	home, err := os.UserHomeDir()
	if err != nil {
		// Fallback to HOME env var, then temp dir
		if homeEnv := os.Getenv("HOME"); homeEnv != "" {
			return filepath.Join(homeEnv, ".claude", "settings.json")
		}
		return filepath.Join(os.TempDir(), ".claude", "settings.json")
	}
	return filepath.Join(home, ".claude", "settings.json")
}

// GetHooksDir returns the Claude Code hooks directory path (expanded)
// Can be overridden for testing
var GetHooksDir = getHooksDirImpl

// GetSettingsPath returns the Claude Code settings file path (expanded)
// Can be overridden for testing
var GetSettingsPath = getSettingsPathImpl
