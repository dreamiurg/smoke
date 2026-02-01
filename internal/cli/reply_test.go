package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/dreamiurg/smoke/internal/feed"
)

func setupSmokeEnvWithPost(t *testing.T) (postID string, cleanup func()) {
	t.Helper()

	tempDir := t.TempDir()
	origHome := os.Getenv("HOME")
	origBDActor := os.Getenv("BD_ACTOR")
	origSmokeAuthor := os.Getenv("SMOKE_AUTHOR")

	os.Setenv("HOME", tempDir)
	os.Setenv("BD_ACTOR", "testbot@testproject")
	os.Setenv("SMOKE_AUTHOR", "")

	// Create smoke config with a post
	configDir := filepath.Join(tempDir, ".config", "smoke")
	os.MkdirAll(configDir, 0755)
	feedPath := filepath.Join(configDir, "feed.jsonl")

	// Create a test post
	post := feed.Post{
		ID:        "smk-abc123",
		Author:    "testbot@testproject",
		Project:   "testproject",
		Suffix:    "test-suffix",
		Content:   "test post",
		CreatedAt: "2026-01-31T12:00:00Z",
	}
	data, _ := json.Marshal(post)
	os.WriteFile(feedPath, append(data, '\n'), 0644)

	return post.ID, func() {
		os.Setenv("HOME", origHome)
		os.Setenv("BD_ACTOR", origBDActor)
		os.Setenv("SMOKE_AUTHOR", origSmokeAuthor)
	}
}

func TestRunReply(t *testing.T) {
	postID, cleanup := setupSmokeEnvWithPost(t)
	defer cleanup()

	// Reset flag
	replyAuthor = ""

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runReply(nil, []string{postID, "test reply"})

	w.Close()
	os.Stdout = oldStdout

	assert.NoError(t, err)

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Verify output contains reply confirmation
	assert.Contains(t, output, "smk-")
}

func TestRunReplyWithAuthor(t *testing.T) {
	postID, cleanup := setupSmokeEnvWithPost(t)
	defer cleanup()

	// Set custom author
	replyAuthor = "custom-replier"
	defer func() { replyAuthor = "" }()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runReply(nil, []string{postID, "reply with custom author"})

	w.Close()
	os.Stdout = oldStdout

	assert.NoError(t, err)

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Output shows "Replied smk-xxx -> smk-yyy" confirmation
	assert.Contains(t, output, "Replied smk-")
}

func TestRunReplyNotInitialized(t *testing.T) {
	tempDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", origHome)

	// Don't create smoke config

	err := runReply(nil, []string{"smk-abc123", "test reply"})

	assert.Error(t, err)
}

func TestRunReplyInvalidPostID(t *testing.T) {
	_, cleanup := setupSmokeEnvWithPost(t)
	defer cleanup()

	err := runReply(nil, []string{"invalid-id", "test reply"})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid post ID")
}

func TestRunReplyPostNotFound(t *testing.T) {
	_, cleanup := setupSmokeEnvWithPost(t)
	defer cleanup()

	err := runReply(nil, []string{"smk-notfnd", "test reply"})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestRunReplyMessageTooLong(t *testing.T) {
	postID, cleanup := setupSmokeEnvWithPost(t)
	defer cleanup()

	// Reset flag
	replyAuthor = ""

	// Create a message longer than 280 chars
	longMessage := strings.Repeat("a", 300)

	err := runReply(nil, []string{postID, longMessage})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "280")
}

func TestReplyCommandRegistered(t *testing.T) {
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "reply <post-id> <message>" {
			found = true
			break
		}
	}
	assert.True(t, found)
}

func TestReplyFlagsRegistered(t *testing.T) {
	authorFlag := replyCmd.Flags().Lookup("author")
	assert.NotNil(t, authorFlag)
}
