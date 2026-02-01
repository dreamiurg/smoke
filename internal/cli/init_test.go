package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
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
