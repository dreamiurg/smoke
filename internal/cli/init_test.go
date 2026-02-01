package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
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

	if err != nil {
		t.Errorf("runInit() error = %v", err)
		return
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Verify dry-run output
	if !strings.Contains(output, "[dry-run]") {
		t.Error("runInit() dry-run output should contain [dry-run] prefix")
	}
	if !strings.Contains(output, "Would") {
		t.Error("runInit() dry-run output should contain 'Would' actions")
	}

	// Verify nothing was actually created
	configDir := filepath.Join(tempDir, ".config", "smoke")
	if _, err := os.Stat(configDir); err == nil {
		t.Error("runInit() dry-run should not create config directory")
	}
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

	if err != nil {
		t.Errorf("runInit() error = %v", err)
		return
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Verify output
	if !strings.Contains(output, "Initialized smoke") {
		t.Error("runInit() output should contain 'Initialized smoke'")
	}

	// Verify files were created
	configDir := filepath.Join(tempDir, ".config", "smoke")
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		t.Error("runInit() should create config directory")
	}

	feedPath := filepath.Join(configDir, "feed.jsonl")
	if _, err := os.Stat(feedPath); os.IsNotExist(err) {
		t.Error("runInit() should create feed file")
	}

	configPath := filepath.Join(configDir, "config.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("runInit() should create config file")
	}
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

	if err != nil {
		t.Errorf("runInit() error = %v", err)
		return
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Verify already initialized message
	if !strings.Contains(output, "already initialized") {
		t.Error("runInit() should indicate already initialized")
	}
	if !strings.Contains(output, "--force") {
		t.Error("runInit() should mention --force option")
	}
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

	if err != nil {
		t.Errorf("runInit() error = %v", err)
		return
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Verify reinitialized
	if !strings.Contains(output, "Initialized smoke") {
		t.Error("runInit() with --force should reinitialize")
	}
}

func TestInitCommandRegistered(t *testing.T) {
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "init" {
			found = true
			break
		}
	}
	if !found {
		t.Error("init command not registered with root")
	}
}

func TestInitFlagsRegistered(t *testing.T) {
	forceFlag := initCmd.Flags().Lookup("force")
	if forceFlag == nil {
		t.Error("--force flag not registered")
	}

	dryRunFlag := initCmd.Flags().Lookup("dry-run")
	if dryRunFlag == nil {
		t.Error("--dry-run flag not registered")
	}
}
