package hooks

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestDir creates a temporary directory for testing
func setupTestDir(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()

	// Override path functions for testing
	originalGetHooksDir := GetHooksDir
	originalGetSettingsPath := GetSettingsPath

	t.Cleanup(func() {
		GetHooksDir = originalGetHooksDir
		GetSettingsPath = originalGetSettingsPath
	})

	return tmpDir
}

// setTestPaths overrides GetHooksDir and GetSettingsPath for testing
func setTestPaths(hooksDir, settingsPath string) {
	oldGetHooksDir := GetHooksDir
	oldGetSettingsPath := GetSettingsPath

	GetHooksDir = func() string { return hooksDir }
	GetSettingsPath = func() string { return settingsPath }

	// Store originals for cleanup (not used in these tests as setupTestDir handles it)
	_, _ = oldGetHooksDir, oldGetSettingsPath
}

func TestInstall_FreshSystem(t *testing.T) {
	tmpDir := setupTestDir(t)
	hooksDir := filepath.Join(tmpDir, "hooks")
	settingsPath := filepath.Join(tmpDir, "settings.json")
	setTestPaths(hooksDir, settingsPath)

	// Install hooks
	err := Install(InstallOptions{Force: false})
	require.NoError(t, err)

	// Verify scripts exist
	for _, script := range ListScripts() {
		scriptPath := filepath.Join(hooksDir, script.Name)
		assert.FileExists(t, scriptPath)

		// Verify permissions
		fileInfo, statErr := os.Stat(scriptPath)
		require.NoError(t, statErr)
		assert.Equal(t, os.FileMode(0755), fileInfo.Mode().Perm())
	}

	// Verify settings.json exists and contains hooks
	assert.FileExists(t, settingsPath)
	data, err := os.ReadFile(settingsPath)
	require.NoError(t, err)

	var settings map[string]interface{}
	err = json.Unmarshal(data, &settings)
	require.NoError(t, err)

	hooks, ok := settings["hooks"].(map[string]interface{})
	require.True(t, ok)

	// Check Stop hook
	stopHooks, ok := hooks["Stop"].([]interface{})
	require.True(t, ok)
	assert.NotEmpty(t, stopHooks)

	// Check PostToolUse hook
	postToolUseHooks, ok := hooks["PostToolUse"].([]interface{})
	require.True(t, ok)
	assert.NotEmpty(t, postToolUseHooks)
}

func TestInstall_AlreadyInstalled(t *testing.T) {
	tmpDir := setupTestDir(t)
	hooksDir := filepath.Join(tmpDir, "hooks")
	settingsPath := filepath.Join(tmpDir, "settings.json")
	setTestPaths(hooksDir, settingsPath)

	// First install
	err := Install(InstallOptions{Force: false})
	require.NoError(t, err)

	// Second install (should be idempotent)
	err = Install(InstallOptions{Force: false})
	require.NoError(t, err)

	// Verify status is installed
	status, err := GetStatus()
	require.NoError(t, err)
	assert.Equal(t, StateInstalled, status.State)
}

func TestInstall_ModifiedScripts(t *testing.T) {
	tmpDir := setupTestDir(t)
	hooksDir := filepath.Join(tmpDir, "hooks")
	settingsPath := filepath.Join(tmpDir, "settings.json")
	setTestPaths(hooksDir, settingsPath)

	// First install
	err := Install(InstallOptions{Force: false})
	require.NoError(t, err)

	// Modify a script
	scriptPath := filepath.Join(hooksDir, "smoke-break.sh")
	err = os.WriteFile(scriptPath, []byte("#!/bin/bash\n# Modified\n"), 0755)
	require.NoError(t, err)

	// Attempt reinstall without --force
	err = Install(InstallOptions{Force: false})
	assert.ErrorIs(t, err, ErrScriptsModified)
}

func TestInstall_ForceOverwrite(t *testing.T) {
	tmpDir := setupTestDir(t)
	hooksDir := filepath.Join(tmpDir, "hooks")
	settingsPath := filepath.Join(tmpDir, "settings.json")
	setTestPaths(hooksDir, settingsPath)

	// First install
	err := Install(InstallOptions{Force: false})
	require.NoError(t, err)

	// Modify a script
	scriptPath := filepath.Join(hooksDir, "smoke-break.sh")
	err = os.WriteFile(scriptPath, []byte("#!/bin/bash\n# Modified\n"), 0755)
	require.NoError(t, err)

	// Reinstall with --force
	err = Install(InstallOptions{Force: true})
	require.NoError(t, err)

	// Verify script is restored
	content, err := os.ReadFile(scriptPath)
	require.NoError(t, err)
	embeddedContent, err := GetScriptContent("smoke-break.sh")
	require.NoError(t, err)
	assert.Equal(t, embeddedContent, content)
}

func TestUninstall_HooksPresent(t *testing.T) {
	tmpDir := setupTestDir(t)
	hooksDir := filepath.Join(tmpDir, "hooks")
	settingsPath := filepath.Join(tmpDir, "settings.json")
	setTestPaths(hooksDir, settingsPath)

	// Install hooks
	err := Install(InstallOptions{Force: false})
	require.NoError(t, err)

	// Uninstall
	err = Uninstall()
	require.NoError(t, err)

	// Verify scripts removed
	for _, script := range ListScripts() {
		scriptPath := filepath.Join(hooksDir, script.Name)
		assert.NoFileExists(t, scriptPath)
	}

	// Verify settings.json no longer has smoke hooks
	data, err := os.ReadFile(settingsPath)
	require.NoError(t, err)

	var settings map[string]interface{}
	err = json.Unmarshal(data, &settings)
	require.NoError(t, err)

	hooks, ok := settings["hooks"].(map[string]interface{})
	if ok {
		// If hooks section exists, verify no smoke hooks
		stopHooks, ok := hooks["Stop"].([]interface{})
		if ok {
			for _, entry := range stopHooks {
				entryMap, ok := entry.(map[string]interface{})
				require.True(t, ok, "entry should be map")
				entryHooks, ok := entryMap["hooks"].([]interface{})
				require.True(t, ok, "hooks should be array")
				for _, hook := range entryHooks {
					hookMap, ok := hook.(map[string]interface{})
					require.True(t, ok, "hook should be map")
					command, ok := hookMap["command"].(string)
					require.True(t, ok, "command should be string")
					assert.NotContains(t, command, "smoke-break.sh")
					assert.NotContains(t, command, "smoke-nudge.sh")
				}
			}
		}
	}
}

func TestUninstall_PreservesOtherHooks(t *testing.T) {
	tmpDir := setupTestDir(t)
	hooksDir := filepath.Join(tmpDir, "hooks")
	settingsPath := filepath.Join(tmpDir, "settings.json")
	setTestPaths(hooksDir, settingsPath)

	// Create settings with other hooks
	otherHook := map[string]interface{}{
		"hooks": map[string]interface{}{
			"Stop": []interface{}{
				map[string]interface{}{
					"matcher": "",
					"hooks": []interface{}{
						map[string]interface{}{
							"type":    "command",
							"command": "/other/hook.sh",
						},
					},
				},
			},
		},
	}
	data, err := json.Marshal(otherHook)
	require.NoError(t, err)
	err = os.MkdirAll(filepath.Dir(settingsPath), 0755)
	require.NoError(t, err)
	err = os.WriteFile(settingsPath, data, 0644)
	require.NoError(t, err)

	// Install smoke hooks
	err = Install(InstallOptions{Force: false})
	require.NoError(t, err)

	// Uninstall smoke hooks
	err = Uninstall()
	require.NoError(t, err)

	// Verify other hook still exists
	data, err = os.ReadFile(settingsPath)
	require.NoError(t, err)

	var settings map[string]interface{}
	err = json.Unmarshal(data, &settings)
	require.NoError(t, err)

	hooks := settings["hooks"].(map[string]interface{})
	stopHooks := hooks["Stop"].([]interface{})
	assert.Len(t, stopHooks, 1)

	entry := stopHooks[0].(map[string]interface{})
	entryHooks := entry["hooks"].([]interface{})
	hook := entryHooks[0].(map[string]interface{})
	assert.Equal(t, "/other/hook.sh", hook["command"])
}

func TestGetStatus_NotInstalled(t *testing.T) {
	tmpDir := setupTestDir(t)
	hooksDir := filepath.Join(tmpDir, "hooks")
	settingsPath := filepath.Join(tmpDir, "settings.json")
	setTestPaths(hooksDir, settingsPath)

	status, err := GetStatus()
	require.NoError(t, err)
	assert.Equal(t, StateNotInstalled, status.State)

	for _, script := range ListScripts() {
		info := status.Scripts[script.Name]
		assert.False(t, info.Exists)
		assert.Equal(t, StatusMissing, info.Status)
	}

	assert.False(t, status.Settings.Stop)
	assert.False(t, status.Settings.PostToolUse)
}

func TestGetStatus_Installed(t *testing.T) {
	tmpDir := setupTestDir(t)
	hooksDir := filepath.Join(tmpDir, "hooks")
	settingsPath := filepath.Join(tmpDir, "settings.json")
	setTestPaths(hooksDir, settingsPath)

	// Install hooks
	err := Install(InstallOptions{Force: false})
	require.NoError(t, err)

	status, err := GetStatus()
	require.NoError(t, err)
	assert.Equal(t, StateInstalled, status.State)

	for _, script := range ListScripts() {
		info := status.Scripts[script.Name]
		assert.True(t, info.Exists)
		assert.False(t, info.Modified)
		assert.Equal(t, StatusOK, info.Status)
	}

	assert.True(t, status.Settings.Stop)
	assert.True(t, status.Settings.PostToolUse)
}

func TestGetStatus_Modified(t *testing.T) {
	tmpDir := setupTestDir(t)
	hooksDir := filepath.Join(tmpDir, "hooks")
	settingsPath := filepath.Join(tmpDir, "settings.json")
	setTestPaths(hooksDir, settingsPath)

	// Install hooks
	err := Install(InstallOptions{Force: false})
	require.NoError(t, err)

	// Modify a script
	scriptPath := filepath.Join(hooksDir, "smoke-break.sh")
	err = os.WriteFile(scriptPath, []byte("#!/bin/bash\n# Modified\n"), 0755)
	require.NoError(t, err)

	status, err := GetStatus()
	require.NoError(t, err)
	assert.Equal(t, StateModified, status.State)

	info := status.Scripts["smoke-break.sh"]
	assert.True(t, info.Exists)
	assert.True(t, info.Modified)
	assert.Equal(t, StatusModified, info.Status)
}

func TestGetStatus_PartiallyInstalled(t *testing.T) {
	tmpDir := setupTestDir(t)
	hooksDir := filepath.Join(tmpDir, "hooks")
	settingsPath := filepath.Join(tmpDir, "settings.json")
	setTestPaths(hooksDir, settingsPath)

	// Install hooks
	err := Install(InstallOptions{Force: false})
	require.NoError(t, err)

	// Remove one script
	scriptPath := filepath.Join(hooksDir, "smoke-break.sh")
	err = os.Remove(scriptPath)
	require.NoError(t, err)

	status, err := GetStatus()
	require.NoError(t, err)
	assert.Equal(t, StatePartiallyInstalled, status.State)

	info := status.Scripts["smoke-break.sh"]
	assert.False(t, info.Exists)
	assert.Equal(t, StatusMissing, info.Status)
}
