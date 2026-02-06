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

func TestRunFeed_Normal(t *testing.T) {
	tmpDir := t.TempDir()
	feedPath := filepath.Join(tmpDir, "feed.jsonl")
	store := feed.NewStoreWithPath(feedPath)
	if err := os.WriteFile(feedPath, []byte(""), 0o600); err != nil {
		t.Fatalf("write feed file: %v", err)
	}

	post, err := feed.NewPost("tester", "project", "sfx", "hello feed")
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

	prevLimit := feedLimit
	prevTail := feedTail
	prevOneline := feedOneline
	prevQuiet := feedQuiet
	prevAuthor := feedAuthor
	prevSuffix := feedSuffix
	prevToday := feedToday
	prevSince := feedSince
	defer func() {
		feedLimit = prevLimit
		feedTail = prevTail
		feedOneline = prevOneline
		feedQuiet = prevQuiet
		feedAuthor = prevAuthor
		feedSuffix = prevSuffix
		feedToday = prevToday
		feedSince = prevSince
	}()

	feedLimit = 10
	feedTail = false
	feedOneline = true
	feedQuiet = true
	feedAuthor = ""
	feedSuffix = ""
	feedToday = false
	feedSince = 0

	output := captureFeedStdout(t, func() {
		if err := runFeed(nil, []string{}); err != nil {
			t.Fatalf("runFeed error: %v", err)
		}
	})

	if !strings.Contains(output, "hello feed") {
		t.Errorf("expected post content in output, got: %s", output)
	}
}

func captureFeedStdout(t *testing.T, fn func()) string {
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
