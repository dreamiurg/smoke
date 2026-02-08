package cli

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/dreamiurg/smoke/internal/feed"
)

func TestRunSuggest_SkippedJSON(t *testing.T) {
	tmpDir := t.TempDir()
	feedPath := filepath.Join(tmpDir, "feed.jsonl")
	if err := os.WriteFile(feedPath, []byte(""), 0o600); err != nil {
		t.Fatalf("write feed file: %v", err)
	}

	oldFeed := os.Getenv("SMOKE_FEED")
	_ = os.Setenv("SMOKE_FEED", feedPath)
	defer func() {
		if oldFeed == "" {
			_ = os.Unsetenv("SMOKE_FEED")
		} else {
			_ = os.Setenv("SMOKE_FEED", oldFeed)
		}
	}()

	oldXDG := os.Getenv("XDG_CONFIG_HOME")
	_ = os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer func() {
		if oldXDG == "" {
			_ = os.Unsetenv("XDG_CONFIG_HOME")
		} else {
			_ = os.Setenv("XDG_CONFIG_HOME", oldXDG)
		}
	}()

	prevSince := suggestSince
	prevJSON := suggestJSON
	prevContext := suggestContext
	prevPressure := suggestPressure
	defer func() {
		suggestSince = prevSince
		suggestJSON = prevJSON
		suggestContext = prevContext
		suggestPressure = prevPressure
	}()

	suggestSince = 4 * time.Hour
	suggestJSON = true
	suggestContext = ""
	suggestPressure = 0

	output := captureSuggestStdout(t, func() {
		if err := runSuggest(nil, []string{}); err != nil {
			t.Fatalf("runSuggest error: %v", err)
		}
	})

	if !strings.Contains(output, "\"skipped\": true") {
		t.Fatalf("expected skipped JSON, got: %s", output)
	}
}

func TestRunSuggest_FiredText(t *testing.T) {
	tmpDir := t.TempDir()
	feedPath := filepath.Join(tmpDir, "feed.jsonl")
	if err := os.WriteFile(feedPath, []byte(""), 0o600); err != nil {
		t.Fatalf("write feed file: %v", err)
	}
	store := feed.NewStoreWithPath(feedPath)

	post, err := feed.NewPost("tester", "project", "sfx", "hello suggest")
	if err != nil {
		t.Fatal(err)
	}
	post.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	if err := store.Append(post); err != nil {
		t.Fatal(err)
	}

	oldFeed := os.Getenv("SMOKE_FEED")
	_ = os.Setenv("SMOKE_FEED", feedPath)
	defer func() {
		if oldFeed == "" {
			_ = os.Unsetenv("SMOKE_FEED")
		} else {
			_ = os.Setenv("SMOKE_FEED", oldFeed)
		}
	}()

	oldXDG := os.Getenv("XDG_CONFIG_HOME")
	_ = os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer func() {
		if oldXDG == "" {
			_ = os.Unsetenv("XDG_CONFIG_HOME")
		} else {
			_ = os.Setenv("XDG_CONFIG_HOME", oldXDG)
		}
	}()

	prevSince := suggestSince
	prevJSON := suggestJSON
	prevContext := suggestContext
	prevPressure := suggestPressure
	defer func() {
		suggestSince = prevSince
		suggestJSON = prevJSON
		suggestContext = prevContext
		suggestPressure = prevPressure
	}()

	suggestSince = 24 * time.Hour
	suggestJSON = false
	suggestContext = "deep-in-it"
	suggestPressure = 4

	output := captureSuggestStdout(t, func() {
		if err := runSuggest(nil, []string{}); err != nil {
			t.Fatalf("runSuggest error: %v", err)
		}
	})

	// Mode is probabilistic (30% reply when recent posts exist)
	// Accept either post mode or reply mode output
	hasPostMode := strings.Contains(output, "What's happening:") && strings.Contains(output, "Post ideas:")
	hasReplyMode := strings.Contains(output, "Recent activity (pick one and reply):")
	if !hasPostMode && !hasReplyMode {
		t.Fatalf("expected either post mode (What's happening + Post ideas) or reply mode (Recent activity), got: %s", output)
	}
}

func TestRunSuggest_ReplyContext(t *testing.T) {
	tmpDir := t.TempDir()
	feedPath := filepath.Join(tmpDir, "feed.jsonl")
	if err := os.WriteFile(feedPath, []byte(""), 0o600); err != nil {
		t.Fatalf("write feed file: %v", err)
	}
	store := feed.NewStoreWithPath(feedPath)

	post, err := feed.NewPost("tester", "project", "sfx", "reply context test")
	if err != nil {
		t.Fatal(err)
	}
	post.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	if err := store.Append(post); err != nil {
		t.Fatal(err)
	}

	oldFeed := os.Getenv("SMOKE_FEED")
	_ = os.Setenv("SMOKE_FEED", feedPath)
	defer func() {
		if oldFeed == "" {
			_ = os.Unsetenv("SMOKE_FEED")
		} else {
			_ = os.Setenv("SMOKE_FEED", oldFeed)
		}
	}()

	oldXDG := os.Getenv("XDG_CONFIG_HOME")
	_ = os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer func() {
		if oldXDG == "" {
			_ = os.Unsetenv("XDG_CONFIG_HOME")
		} else {
			_ = os.Setenv("XDG_CONFIG_HOME", oldXDG)
		}
	}()

	prevSince := suggestSince
	prevJSON := suggestJSON
	prevContext := suggestContext
	prevPressure := suggestPressure
	defer func() {
		suggestSince = prevSince
		suggestJSON = prevJSON
		suggestContext = prevContext
		suggestPressure = prevPressure
	}()

	suggestSince = 24 * time.Hour
	suggestJSON = false
	suggestContext = "reply"
	suggestPressure = 4

	output := captureSuggestStdout(t, func() {
		if err := runSuggest(nil, []string{}); err != nil {
			t.Fatalf("runSuggest error: %v", err)
		}
	})

	// --context=reply forces reply mode deterministically
	if !strings.Contains(output, "Recent activity (pick one and reply):") {
		t.Fatalf("expected reply mode output, got: %s", output)
	}
	if !strings.Contains(output, "smoke reply") {
		t.Error("missing reply command hint")
	}
}

func TestRunSuggest_InvalidContext(t *testing.T) {
	tmpDir := t.TempDir()
	feedPath := filepath.Join(tmpDir, "feed.jsonl")
	if err := os.WriteFile(feedPath, []byte(""), 0o600); err != nil {
		t.Fatalf("write feed file: %v", err)
	}

	oldFeed := os.Getenv("SMOKE_FEED")
	_ = os.Setenv("SMOKE_FEED", feedPath)
	defer func() {
		if oldFeed == "" {
			_ = os.Unsetenv("SMOKE_FEED")
		} else {
			_ = os.Setenv("SMOKE_FEED", oldFeed)
		}
	}()

	oldXDG := os.Getenv("XDG_CONFIG_HOME")
	_ = os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer func() {
		if oldXDG == "" {
			_ = os.Unsetenv("XDG_CONFIG_HOME")
		} else {
			_ = os.Setenv("XDG_CONFIG_HOME", oldXDG)
		}
	}()

	prevSince := suggestSince
	prevJSON := suggestJSON
	prevContext := suggestContext
	prevPressure := suggestPressure
	defer func() {
		suggestSince = prevSince
		suggestJSON = prevJSON
		suggestContext = prevContext
		suggestPressure = prevPressure
	}()

	suggestSince = 4 * time.Hour
	suggestJSON = false
	suggestContext = "nope"
	suggestPressure = 4

	if err := runSuggest(nil, []string{}); err == nil {
		t.Fatal("expected error for unknown context")
	}
}

func captureSuggestStdout(t *testing.T, fn func()) string {
	t.Helper()
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w

	fn()

	_ = w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	return buf.String()
}
