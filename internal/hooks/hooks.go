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

// checkForModifiedScripts checks if any installed scripts have been modified
// from their embedded versions. Returns ErrScriptsModified if any differ.
func checkForModifiedScripts(hooksDir string) error {
	for _, script := range ListScripts() {
		content, err := GetScriptContent(script.Name)
		if err != nil {
			return fmt.Errorf("get embedded script %s: %w", script.Name, err)
		}

		installedPath := filepath.Join(hooksDir, script.Name)
		if scriptExists(installedPath) && isScriptModified(installedPath, content) {
			return ErrScriptsModified
		}
	}
	return nil
}

// installScriptFiles creates the hooks directory and writes all embedded scripts.
func installScriptFiles(hooksDir string) error {
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		return fmt.Errorf("%w: cannot create hooks directory", ErrPermissionDenied)
	}

	for _, script := range ListScripts() {
		content, err := GetScriptContent(script.Name)
		if err != nil {
			return fmt.Errorf("get embedded script %s: %w", script.Name, err)
		}

		installedPath := filepath.Join(hooksDir, script.Name)
		if err := writeScript(installedPath, content); err != nil {
			return err
		}
	}
	return nil
}

// loadOrResetSettings reads settings, handling invalid JSON by creating a backup
// and returning fresh settings. It also creates a backup before any modification.
func loadOrResetSettings(result *InstallResult) (map[string]interface{}, error) {
	settings, err := readSettings()
	if err != nil {
		if err == ErrInvalidSettings {
			backupPath, backupErr := BackupSettings()
			if backupErr != nil {
				return nil, fmt.Errorf("backup invalid settings: %w", backupErr)
			}
			result.BackupPath = backupPath
			return make(map[string]interface{}), nil
		}
		return nil, err
	}

	// Create backup before modifying settings
	if result.BackupPath == "" {
		backupPath, backupErr := BackupSettings()
		if backupErr != nil {
			return nil, fmt.Errorf("backup settings: %w", backupErr)
		}
		result.BackupPath = backupPath
	}

	return settings, nil
}

// Install installs smoke hooks to Claude Code
func Install(opts InstallOptions) (*InstallResult, error) {
	result := &InstallResult{}
	hooksDir := GetHooksDir()

	// Check for modified scripts first (unless --force)
	if !opts.Force {
		if err := checkForModifiedScripts(hooksDir); err != nil {
			return nil, err
		}
	}

	if err := installScriptFiles(hooksDir); err != nil {
		return nil, err
	}

	// Load settings (with backup and reset if invalid)
	settings, err := loadOrResetSettings(result)
	if err != nil {
		return nil, err
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

// removeSettingsHooks removes smoke hook entries from settings.json.
// If settings are invalid, it skips settings modification silently.
func removeSettingsHooks(result *UninstallResult) error {
	settings, err := readSettings()
	if err != nil && err != ErrInvalidSettings {
		return err
	}
	if err != nil {
		// Settings are invalid; skip settings modification
		return nil
	}

	// Create backup before modifying settings
	backupPath, backupErr := BackupSettings()
	if backupErr != nil {
		return fmt.Errorf("backup settings: %w", backupErr)
	}
	result.BackupPath = backupPath

	for _, script := range ListScripts() {
		if err := removeHookFromSettings(settings, script.Event); err != nil {
			return err
		}
	}

	return writeSettings(settings)
}

// removeScriptFiles deletes installed hook scripts and the state directory.
func removeScriptFiles() error {
	hooksDir := GetHooksDir()
	for _, script := range ListScripts() {
		scriptPath := filepath.Join(hooksDir, script.Name)
		if !scriptExists(scriptPath) {
			continue
		}
		if err := os.Remove(scriptPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("%w: cannot remove script", ErrPermissionDenied)
		}
	}

	// Clean up state directory (best effort)
	stateDir := filepath.Join(hooksDir, "smoke-nudge-state")
	if scriptExists(stateDir) {
		_ = os.RemoveAll(stateDir) // Best effort, ignore errors
	}

	return nil
}

// Uninstall removes smoke hooks from Claude Code
func Uninstall() (*UninstallResult, error) {
	result := &UninstallResult{}

	if err := removeSettingsHooks(result); err != nil {
		return nil, err
	}

	if err := removeScriptFiles(); err != nil {
		return nil, err
	}

	return result, nil
}

// getScriptStatuses checks each hook script and returns the script info map
// along with flags indicating whether any are installed, missing, or modified.
func getScriptStatuses(hooksDir string) (map[string]ScriptInfo, bool, bool, bool, error) {
	scripts := make(map[string]ScriptInfo)
	anyInstalled, anyMissing, anyModified := false, false, false

	for _, script := range ListScripts() {
		content, err := GetScriptContent(script.Name)
		if err != nil {
			return nil, false, false, false, fmt.Errorf("get embedded script %s: %w", script.Name, err)
		}

		installedPath := filepath.Join(hooksDir, script.Name)
		status := getScriptStatus(installedPath, content)
		scripts[script.Name] = ScriptInfo{
			Path:     installedPath,
			Exists:   scriptExists(installedPath),
			Modified: isScriptModified(installedPath, content),
			Status:   status,
		}

		switch status {
		case StatusOK:
			anyInstalled = true
		case StatusMissing:
			anyMissing = true
		case StatusModified:
			anyModified = true
		}
	}
	return scripts, anyInstalled, anyMissing, anyModified, nil
}

// determineInstallState maps per-script flags to an overall InstallState.
func determineInstallState(anyInstalled, anyMissing, anyModified bool) InstallState {
	switch {
	case !anyInstalled && !anyModified:
		return StateNotInstalled
	case anyModified:
		return StateModified
	case anyMissing:
		return StatePartiallyInstalled
	default:
		return StateInstalled
	}
}

// GetStatus returns the current hook installation status
func GetStatus() (*Status, error) {
	scripts, anyInstalled, anyMissing, anyModified, err := getScriptStatuses(GetHooksDir())
	if err != nil {
		return nil, err
	}

	settingsInfo := SettingsInfo{}
	if settings, sErr := readSettings(); sErr == nil {
		settingsInfo.Stop = checkHookInSettings(settings, EventStop)
		settingsInfo.PostToolUse = checkHookInSettings(settings, EventPostToolUse)
	}

	return &Status{
		State:    determineInstallState(anyInstalled, anyMissing, anyModified),
		Scripts:  scripts,
		Settings: settingsInfo,
	}, nil
}
