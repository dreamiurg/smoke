package hooks

import (
	"fmt"
	"os"
	"path/filepath"
)

// InstallResult contains the result of a hook installation
type InstallResult struct {
	BackupPath string // Path to settings backup, empty if no backup created
}

// Install installs smoke hooks to Claude Code
func Install(opts InstallOptions) (*InstallResult, error) {
	result := &InstallResult{}
	hooksDir := GetHooksDir()

	// Check for modified scripts first (unless --force)
	if !opts.Force {
		for _, script := range ListScripts() {
			content, err := GetScriptContent(script.Name)
			if err != nil {
				return nil, fmt.Errorf("get embedded script %s: %w", script.Name, err)
			}

			installedPath := filepath.Join(hooksDir, script.Name)
			if scriptExists(installedPath) && isScriptModified(installedPath, content) {
				return nil, ErrScriptsModified
			}
		}
	}

	// Create hooks directory
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		return nil, fmt.Errorf("%w: cannot create hooks directory", ErrPermissionDenied)
	}

	// Install each script
	for _, script := range ListScripts() {
		content, err := GetScriptContent(script.Name)
		if err != nil {
			return nil, fmt.Errorf("get embedded script %s: %w", script.Name, err)
		}

		installedPath := filepath.Join(hooksDir, script.Name)
		if err := writeScript(installedPath, content); err != nil {
			return nil, err
		}
	}

	// Update settings.json
	settings, err := readSettings()
	if err != nil {
		// If settings are invalid, backup and create fresh
		if err == ErrInvalidSettings {
			backupPath, backupErr := BackupSettings()
			if backupErr != nil {
				return nil, fmt.Errorf("backup invalid settings: %w", backupErr)
			}
			result.BackupPath = backupPath
			settings = make(map[string]interface{})
		} else {
			return nil, err
		}
	}

	// Create backup before modifying settings
	if result.BackupPath == "" {
		backupPath, backupErr := BackupSettings()
		if backupErr != nil {
			return nil, fmt.Errorf("backup settings: %w", backupErr)
		}
		result.BackupPath = backupPath
	}

	// Add hook entries to settings
	for _, script := range ListScripts() {
		scriptPath := filepath.Join(hooksDir, script.Name)
		if err := addHookToSettings(settings, script.Event, scriptPath); err != nil {
			return nil, err
		}
	}

	// Write updated settings
	if err := writeSettings(settings); err != nil {
		return nil, err
	}

	return result, nil
}

// UninstallResult contains the result of a hook uninstallation
type UninstallResult struct {
	BackupPath string // Path to settings backup, empty if no backup created
}

// Uninstall removes smoke hooks from Claude Code
func Uninstall() (*UninstallResult, error) {
	result := &UninstallResult{}

	// Update settings.json first
	settings, err := readSettings()
	if err != nil && err != ErrInvalidSettings {
		return nil, err
	}

	// Remove hook entries from settings
	if err == nil { // Only if settings are valid
		// Create backup before modifying settings
		backupPath, backupErr := BackupSettings()
		if backupErr != nil {
			return nil, fmt.Errorf("backup settings: %w", backupErr)
		}
		result.BackupPath = backupPath

		for _, script := range ListScripts() {
			if err := removeHookFromSettings(settings, script.Event); err != nil {
				return nil, err
			}
		}

		// Write updated settings
		if err := writeSettings(settings); err != nil {
			return nil, err
		}
	}

	// Remove script files
	hooksDir := GetHooksDir()
	for _, script := range ListScripts() {
		scriptPath := filepath.Join(hooksDir, script.Name)
		if scriptExists(scriptPath) {
			if err := os.Remove(scriptPath); err != nil && !os.IsNotExist(err) {
				return nil, fmt.Errorf("%w: cannot remove script", ErrPermissionDenied)
			}
		}
	}

	// Clean up state directory (best effort)
	stateDir := filepath.Join(hooksDir, "smoke-nudge-state")
	if scriptExists(stateDir) {
		_ = os.RemoveAll(stateDir) // Best effort, ignore errors
	}

	return result, nil
}

// GetStatus returns the current hook installation status
func GetStatus() (*Status, error) {
	hooksDir := GetHooksDir()
	scripts := make(map[string]ScriptInfo)

	// Check each script
	anyInstalled := false
	anyMissing := false
	anyModified := false

	for _, script := range ListScripts() {
		content, err := GetScriptContent(script.Name)
		if err != nil {
			return nil, fmt.Errorf("get embedded script %s: %w", script.Name, err)
		}

		installedPath := filepath.Join(hooksDir, script.Name)
		status := getScriptStatus(installedPath, content)

		info := ScriptInfo{
			Path:     installedPath,
			Exists:   scriptExists(installedPath),
			Modified: isScriptModified(installedPath, content),
			Status:   status,
		}

		scripts[script.Name] = info

		// Track overall state
		switch status {
		case StatusOK:
			anyInstalled = true
		case StatusMissing:
			anyMissing = true
		case StatusModified:
			anyModified = true
		}
	}

	// Read settings
	settings, err := readSettings()
	settingsInfo := SettingsInfo{}

	if err == nil {
		settingsInfo.Stop = checkHookInSettings(settings, EventStop)
		settingsInfo.PostToolUse = checkHookInSettings(settings, EventPostToolUse)
	}

	// Determine overall state
	var state InstallState
	switch {
	case !anyInstalled && !anyModified:
		state = StateNotInstalled
	case anyModified:
		state = StateModified
	case anyMissing:
		state = StatePartiallyInstalled
	default:
		state = StateInstalled
	}

	return &Status{
		State:    state,
		Scripts:  scripts,
		Settings: settingsInfo,
	}, nil
}
