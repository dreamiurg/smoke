package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func setupSmokeEnv(t *testing.T) (cleanup func()) {
	t.Helper()

	tempDir := t.TempDir()
	origHome := os.Getenv("HOME")
	origSmokeName := os.Getenv("SMOKE_NAME")

	os.Setenv("HOME", tempDir)
	os.Setenv("SMOKE_NAME", "testbot@testproject")

	// Create smoke config
	configDir := filepath.Join(tempDir, ".config", "smoke")
	os.MkdirAll(configDir, 0755)
	feedPath := filepath.Join(configDir, "feed.jsonl")
	os.WriteFile(feedPath, []byte{}, 0644)

	return func() {
		os.Setenv("HOME", origHome)
		os.Setenv("SMOKE_NAME", origSmokeName)
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

	assert.NoError(t, err)

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Verify output contains confirmation
	assert.Contains(t, output, "smk-")
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

	assert.NoError(t, err)

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Output shows "Posted smk-xxx" confirmation
	assert.Contains(t, output, "Posted smk-")
}

func TestRunPostNotInitialized(t *testing.T) {
	tempDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", origHome)

	// Don't create smoke config

	err := runPost(nil, []string{"test message"})

	assert.Error(t, err)
}

func TestRunPostMessageTooLong(t *testing.T) {
	cleanup := setupSmokeEnv(t)
	defer cleanup()

	// Reset flag
	postAuthor = ""

	// Create a message longer than 280 chars
	longMessage := strings.Repeat("a", 300)

	err := runPost(nil, []string{longMessage})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "280")
}

func TestPostCommandRegistered(t *testing.T) {
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "post <message>" {
			found = true
			break
		}
	}
	assert.True(t, found)
}

func TestPostFlagsRegistered(t *testing.T) {
	authorFlag := postCmd.Flags().Lookup("author")
	assert.NotNil(t, authorFlag)
}
