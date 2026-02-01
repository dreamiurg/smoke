package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dreamiurg/smoke/internal/hooks"
)

func TestRunInitDryRun(t *testing.T) {
	// Set up temp directory for config
	tempDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", origHome)

	// Create .claude dir for hint check
	claudeDir := filepath.Join(tempDir, ".claude")
	os.MkdirAll(claudeDir, 0755)

	// Reset flags
	initForce = false
	initDryRun = true
	defer func() {
		initDryRun = false
	}()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runInit(nil, nil)

	w.Close()
	os.Stdout = oldStdout

	assert.NoError(t, err)

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Verify dry-run output
	assert.Contains(t, output, "[dry-run]")
	assert.Contains(t, output, "Would")

	// Verify nothing was actually created
	configDir := filepath.Join(tempDir, ".config", "smoke")
	_, err = os.Stat(configDir)
	assert.True(t, os.IsNotExist(err))
}

func TestRunInitFresh(t *testing.T) {
	// Set up temp directory for config
	tempDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", origHome)

	// Create .claude dir for hint check
	claudeDir := filepath.Join(tempDir, ".claude")
	os.MkdirAll(claudeDir, 0755)

	// Reset flags
	initForce = false
	initDryRun = false

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runInit(nil, nil)

	w.Close()
	os.Stdout = oldStdout

	assert.NoError(t, err)

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Verify output
	assert.Contains(t, output, "Initialized smoke")

	// Verify files were created
	configDir := filepath.Join(tempDir, ".config", "smoke")
	_, err = os.Stat(configDir)
	assert.False(t, os.IsNotExist(err))

	feedPath := filepath.Join(configDir, "feed.jsonl")
	_, err = os.Stat(feedPath)
	assert.False(t, os.IsNotExist(err))

	configPath := filepath.Join(configDir, "config.yaml")
	_, err = os.Stat(configPath)
	assert.False(t, os.IsNotExist(err))
}

func TestRunInitAlreadyInitialized(t *testing.T) {
	// Set up temp directory with existing smoke config
	tempDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", origHome)

	// Create existing config
	configDir := filepath.Join(tempDir, ".config", "smoke")
	os.MkdirAll(configDir, 0755)
	feedPath := filepath.Join(configDir, "feed.jsonl")
	os.WriteFile(feedPath, []byte{}, 0644)

	// Reset flags
	initForce = false
	initDryRun = false

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runInit(nil, nil)

	w.Close()
	os.Stdout = oldStdout

	assert.NoError(t, err)

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Verify already initialized message
	assert.Contains(t, output, "already initialized")
	assert.Contains(t, output, "--force")
}

func TestRunInitForce(t *testing.T) {
	// Set up temp directory with existing smoke config
	tempDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", origHome)

	// Create existing config
	configDir := filepath.Join(tempDir, ".config", "smoke")
	os.MkdirAll(configDir, 0755)
	feedPath := filepath.Join(configDir, "feed.jsonl")
	os.WriteFile(feedPath, []byte("old content"), 0644)

	// Create .claude dir for hint check
	claudeDir := filepath.Join(tempDir, ".claude")
	os.MkdirAll(claudeDir, 0755)

	// Set force flag
	initForce = true
	initDryRun = false
	defer func() {
		initForce = false
	}()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runInit(nil, nil)

	w.Close()
	os.Stdout = oldStdout

	assert.NoError(t, err)

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Verify reinitialized
	assert.Contains(t, output, "Initialized smoke")
}

func TestInitCommandRegistered(t *testing.T) {
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "init" {
			found = true
			break
		}
	}
	assert.True(t, found)
}

func TestInitFlagsRegistered(t *testing.T) {
	forceFlag := initCmd.Flags().Lookup("force")
	assert.NotNil(t, forceFlag)

	dryRunFlag := initCmd.Flags().Lookup("dry-run")
	assert.NotNil(t, dryRunFlag)
}

// Hook integration tests

func TestRunInit_InstallsHooks(t *testing.T) {
	// Set up temp directory
	tempDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", origHome)

	// Create .claude dir
	claudeDir := filepath.Join(tempDir, ".claude")
	os.MkdirAll(claudeDir, 0755)

	// Reset flags
	initForce = false
	initDryRun = false

	// Run init
	err := runInit(nil, nil)
	assert.NoError(t, err)

	// Verify hooks were installed
	status, err := hooks.GetStatus()
	require.NoError(t, err)
	assert.Equal(t, hooks.StateInstalled, status.State)

	// Verify hook scripts exist
	hooksDir := filepath.Join(tempDir, ".claude", "hooks")
	assert.DirExists(t, hooksDir)
	assert.FileExists(t, filepath.Join(hooksDir, "smoke-break.sh"))
	assert.FileExists(t, filepath.Join(hooksDir, "smoke-nudge.sh"))

	// Verify settings.json was created
	settingsPath := filepath.Join(tempDir, ".claude", "settings.json")
	assert.FileExists(t, settingsPath)
}

func TestRunInit_HookErrorsDoNotFailInit(t *testing.T) {
	// Set up temp directory with permission issue
	tempDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", origHome)

	// Create .claude dir and make hooks dir unwritable
	claudeDir := filepath.Join(tempDir, ".claude")
	os.MkdirAll(claudeDir, 0755)
	hooksDir := filepath.Join(claudeDir, "hooks")
	os.MkdirAll(hooksDir, 0000) // No permissions
	defer os.Chmod(hooksDir, 0755)

	// Reset flags
	initForce = false
	initDryRun = false

	// Run init - should succeed despite hook failure
	err := runInit(nil, nil)
	assert.NoError(t, err)

	// Verify smoke was initialized
	configDir := filepath.Join(tempDir, ".config", "smoke")
	assert.DirExists(t, configDir)
}

func TestRunInit_AlreadyInitializedSuggestsHooksInstall(t *testing.T) {
	// Set up temp directory with smoke initialized but no hooks
	tempDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", origHome)

	// Create existing config (smoke initialized)
	configDir := filepath.Join(tempDir, ".config", "smoke")
	os.MkdirAll(configDir, 0755)
	feedPath := filepath.Join(configDir, "feed.jsonl")
	os.WriteFile(feedPath, []byte{}, 0644)

	// Ensure hooks are NOT installed
	claudeDir := filepath.Join(tempDir, ".claude")
	os.MkdirAll(claudeDir, 0755)

	// Reset flags
	initForce = false
	initDryRun = false

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runInit(nil, nil)

	w.Close()
	os.Stdout = oldStdout

	assert.NoError(t, err)

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Verify suggests hooks install
	assert.Contains(t, output, "already initialized")
	assert.Contains(t, output, "smoke hooks install")
}
