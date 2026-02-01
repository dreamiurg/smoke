package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupSmokeEnv(t *testing.T) (cleanup func()) {
	t.Helper()

	tempDir := t.TempDir()
	origHome := os.Getenv("HOME")
	origBDActor := os.Getenv("BD_ACTOR")
	origSmokeAuthor := os.Getenv("SMOKE_AUTHOR")

	os.Setenv("HOME", tempDir)
	os.Setenv("BD_ACTOR", "testbot@testproject")
	os.Setenv("SMOKE_AUTHOR", "")

	// Create smoke config
	configDir := filepath.Join(tempDir, ".config", "smoke")
	os.MkdirAll(configDir, 0755)
	feedPath := filepath.Join(configDir, "feed.jsonl")
	os.WriteFile(feedPath, []byte{}, 0644)

	return func() {
		os.Setenv("HOME", origHome)
		os.Setenv("BD_ACTOR", origBDActor)
		os.Setenv("SMOKE_AUTHOR", origSmokeAuthor)
	}
}

func TestRunPost(t *testing.T) {
	cleanup := setupSmokeEnv(t)
	defer cleanup()

	// Reset flag
	postAuthor = ""

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runPost(nil, []string{"test message"})

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Errorf("runPost() error = %v", err)
		return
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Verify output contains confirmation
	if !strings.Contains(output, "smk-") {
		t.Error("runPost() output should contain post ID (smk-*)")
	}
}

func TestRunPostWithAuthor(t *testing.T) {
	cleanup := setupSmokeEnv(t)
	defer cleanup()

	// Set custom author
	postAuthor = "custom-author"
	defer func() { postAuthor = "" }()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runPost(nil, []string{"test with custom author"})

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Errorf("runPost() error = %v", err)
		return
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Output shows "Posted smk-xxx" confirmation
	if !strings.Contains(output, "Posted smk-") {
		t.Error("runPost() output should contain post confirmation")
	}
}

func TestRunPostNotInitialized(t *testing.T) {
	tempDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", origHome)

	// Don't create smoke config

	err := runPost(nil, []string{"test message"})

	if err == nil {
		t.Error("runPost() should return error when not initialized")
	}
}

func TestRunPostMessageTooLong(t *testing.T) {
	cleanup := setupSmokeEnv(t)
	defer cleanup()

	// Reset flag
	postAuthor = ""

	// Create a message longer than 280 chars
	longMessage := strings.Repeat("a", 300)

	err := runPost(nil, []string{longMessage})

	if err == nil {
		t.Error("runPost() should return error for message > 280 chars")
	}
	if !strings.Contains(err.Error(), "280") {
		t.Errorf("error should mention 280 char limit, got: %v", err)
	}
}

func TestPostCommandRegistered(t *testing.T) {
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "post <message>" {
			found = true
			break
		}
	}
	if !found {
		t.Error("post command not registered with root")
	}
}

func TestPostFlagsRegistered(t *testing.T) {
	authorFlag := postCmd.Flags().Lookup("author")
	if authorFlag == nil {
		t.Error("--author flag not registered")
	}
}
