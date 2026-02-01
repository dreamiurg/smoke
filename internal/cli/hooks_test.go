package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dreamiurg/smoke/internal/hooks"
)

func setupHooksTest(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	t.Cleanup(func() {
		os.Setenv("HOME", origHome)
		hooksForce = false
		hooksStatusJSON = false
	})

	// Create .claude directory
	claudeDir := filepath.Join(tmpDir, ".claude")
	require.NoError(t, os.MkdirAll(claudeDir, 0755))

	return tmpDir
}

func TestHooksInstall_Fresh(t *testing.T) {
	setupHooksTest(t)

	// Run install
	err := runHooksInstall(nil, nil)
	require.NoError(t, err)

	// Verify hooks installed
	status, err := hooks.GetStatus()
	require.NoError(t, err)
	assert.Equal(t, hooks.StateInstalled, status.State)
}

func TestHooksInstall_Modified(t *testing.T) {
	tmpDir := setupHooksTest(t)

	// Install first
	err := runHooksInstall(nil, nil)
	require.NoError(t, err)

	// Modify a script
	scriptPath := filepath.Join(tmpDir, ".claude", "hooks", "smoke-break.sh")
	err = os.WriteFile(scriptPath, []byte("#!/bin/bash\n# Modified\n"), 0755)
	require.NoError(t, err)

	// Reset force flag
	hooksForce = false

	// Capture stderr
	oldStderr := os.Stderr
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stderr = w

	// Attempt reinstall without --force
	err = runHooksInstall(nil, nil)

	w.Close()
	os.Stderr = oldStderr

	assert.NoError(t, err) // Function returns nil after printing error

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Verify error message
	assert.Contains(t, output, "modified")
	assert.Contains(t, output, "--force")
}

func TestHooksInstall_Force(t *testing.T) {
	tmpDir := setupHooksTest(t)

	// Install first
	err := runHooksInstall(nil, nil)
	require.NoError(t, err)

	// Modify a script
	scriptPath := filepath.Join(tmpDir, ".claude", "hooks", "smoke-break.sh")
	err = os.WriteFile(scriptPath, []byte("#!/bin/bash\n# Modified\n"), 0755)
	require.NoError(t, err)

	// Set force flag
	hooksForce = true

	// Reinstall with --force
	err = runHooksInstall(nil, nil)
	require.NoError(t, err)

	// Verify script restored
	content, err := os.ReadFile(scriptPath)
	require.NoError(t, err)
	embeddedContent, err := hooks.GetScriptContent("smoke-break.sh")
	require.NoError(t, err)
	assert.Equal(t, embeddedContent, content)
}

func TestHooksUninstall(t *testing.T) {
	tmpDir := setupHooksTest(t)

	// Install hooks
	err := hooks.Install(hooks.InstallOptions{Force: false})
	require.NoError(t, err)

	// Uninstall
	err = runHooksUninstall(nil, nil)
	require.NoError(t, err)

	// Verify hooks removed
	status, err := hooks.GetStatus()
	require.NoError(t, err)
	assert.Equal(t, hooks.StateNotInstalled, status.State)

	// Verify scripts removed
	hooksDir := filepath.Join(tmpDir, ".claude", "hooks")
	for _, script := range hooks.ListScripts() {
		scriptPath := filepath.Join(hooksDir, script.Name)
		assert.NoFileExists(t, scriptPath)
	}
}

func TestHooksUninstall_NotInstalled(t *testing.T) {
	setupHooksTest(t)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	// Uninstall when not installed
	err = runHooksUninstall(nil, nil)

	w.Close()
	os.Stdout = oldStdout

	require.NoError(t, err)

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	assert.Contains(t, output, "not installed")
}

func TestHooksStatus_NotInstalled(t *testing.T) {
	setupHooksTest(t)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	// Run status
	err = runHooksStatus(nil, nil)

	w.Close()
	os.Stdout = oldStdout

	require.NoError(t, err)

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	assert.Contains(t, output, "Status: not_installed")
	assert.Contains(t, output, "smoke hooks install")
}

func TestHooksStatus_Installed(t *testing.T) {
	setupHooksTest(t)

	// Install hooks
	err := hooks.Install(hooks.InstallOptions{Force: false})
	require.NoError(t, err)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	// Run status
	err = runHooksStatus(nil, nil)

	w.Close()
	os.Stdout = oldStdout

	require.NoError(t, err)

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	assert.Contains(t, output, "Status: installed")
	assert.Contains(t, output, "ok")
}

func TestHooksStatus_JSON(t *testing.T) {
	setupHooksTest(t)

	// Install hooks
	err := hooks.Install(hooks.InstallOptions{Force: false})
	require.NoError(t, err)

	// Set JSON flag
	hooksStatusJSON = true

	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	// Run status
	err = runHooksStatus(nil, nil)

	w.Close()
	os.Stdout = oldStdout

	require.NoError(t, err)

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Verify valid JSON
	var result map[string]interface{}
	err = json.Unmarshal([]byte(output), &result)
	require.NoError(t, err)

	// Verify structure
	assert.Equal(t, "installed", result["status"])
	assert.NotNil(t, result["scripts"])
	assert.NotNil(t, result["settings"])
}

func TestHooksCommandRegistered(t *testing.T) {
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "hooks" {
			found = true
			break
		}
	}
	assert.True(t, found)
}

func TestHooksSubcommandsRegistered(t *testing.T) {
	var hooksCmdRef *cobra.Command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "hooks" {
			hooksCmdRef = cmd
			break
		}
	}
	require.NotNil(t, hooksCmdRef)

	subcommands := []string{"install", "uninstall", "status"}
	for _, subcmd := range subcommands {
		found := false
		for _, cmd := range hooksCmdRef.Commands() {
			if cmd.Use == subcmd {
				found = true
				break
			}
		}
		assert.True(t, found, "Subcommand %s not found", subcmd)
	}
}
