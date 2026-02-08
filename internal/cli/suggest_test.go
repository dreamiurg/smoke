package cli

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/dreamiurg/smoke/internal/config"
	"github.com/dreamiurg/smoke/internal/feed"
)

func TestFormatTimeAgo(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		time     time.Time
		expected string
	}{
		{"just now", now.Add(-30 * time.Second), "just now"},
		{"1 minute", now.Add(-1 * time.Minute), "1m ago"},
		{"5 minutes", now.Add(-5 * time.Minute), "5m ago"},
		{"59 minutes", now.Add(-59 * time.Minute), "59m ago"},
		{"1 hour", now.Add(-1 * time.Hour), "1h ago"},
		{"5 hours", now.Add(-5 * time.Hour), "5h ago"},
		{"23 hours", now.Add(-23 * time.Hour), "23h ago"},
		{"1 day", now.Add(-24 * time.Hour), "1d ago"},
		{"3 days", now.Add(-72 * time.Hour), "3d ago"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatTimeAgo(tt.time)
			if result != tt.expected {
				t.Errorf("formatTimeAgo(%v) = %q, want %q", tt.time, result, tt.expected)
			}
		})
	}
}

func TestGetRandomExamples(t *testing.T) {
	t.Run("empty input returns empty slice", func(t *testing.T) {
		result := getRandomExamples([]string{}, 2, 3)
		if len(result) != 0 {
			t.Errorf("expected empty slice, got %v", result)
		}
	})

	t.Run("respects max count when examples are plentiful", func(t *testing.T) {
		examples := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}
		result := getRandomExamples(examples, 2, 3)
		if len(result) < 2 || len(result) > 3 {
			t.Errorf("expected 2-3 examples, got %d", len(result))
		}
	})

	t.Run("returns all when fewer than minCount", func(t *testing.T) {
		examples := []string{"only one"}
		result := getRandomExamples(examples, 2, 3)
		if len(result) != 1 {
			t.Errorf("expected 1 example, got %d", len(result))
		}
		if result[0] != "only one" {
			t.Errorf("expected 'only one', got %q", result[0])
		}
	})

	t.Run("returns unique examples", func(t *testing.T) {
		examples := []string{"a", "b", "c", "d", "e"}
		for i := 0; i < 10; i++ { // Run multiple times to check for uniqueness
			result := getRandomExamples(examples, 3, 3)
			seen := make(map[string]bool)
			for _, ex := range result {
				if seen[ex] {
					t.Errorf("duplicate example found: %q", ex)
				}
				seen[ex] = true
			}
		}
	})
}

func TestFormatSuggestPost(t *testing.T) {
	// Create a temp file to capture output
	tmpFile, err := os.CreateTemp("", "suggest_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Create a test post
	post := &feed.Post{
		ID:        "smk-abc123",
		Author:    "test@project",
		Content:   "Test content for formatting",
		CreatedAt: time.Now().Add(-5 * time.Minute).Format(time.RFC3339),
	}

	// Format the post (truncated mode)
	formatSuggestPost(tmpFile, post, false)

	// Read and verify output
	tmpFile.Seek(0, 0)
	var buf bytes.Buffer
	buf.ReadFrom(tmpFile)
	output := buf.String()

	// Check that output contains expected elements
	if !contains(output, "smk-abc123") {
		t.Error("output missing post ID")
	}
	if !contains(output, "test@project") {
		t.Error("output missing author")
	}
	if !contains(output, "Test content") {
		t.Error("output missing content")
	}
	if !contains(output, "5m ago") {
		t.Error("output missing time ago")
	}
}

func TestFormatSuggestPostTruncatesLongContent(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "suggest_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Create a post with long content (>60 chars)
	longContent := "This is a very long content string that should be truncated because it exceeds the preview width limit"
	post := &feed.Post{
		ID:        "smk-xyz789",
		Author:    "test@project",
		Content:   longContent,
		CreatedAt: time.Now().Format(time.RFC3339),
	}

	// Truncated mode should cut long content
	formatSuggestPost(tmpFile, post, false)

	tmpFile.Seek(0, 0)
	var buf bytes.Buffer
	buf.ReadFrom(tmpFile)
	output := buf.String()

	// Should contain "..." indicating truncation
	if !contains(output, "...") {
		t.Error("long content should be truncated with '...'")
	}
	// Should not contain the full content
	if contains(output, "preview width limit") {
		t.Error("content should have been truncated")
	}
}

func TestFormatSuggestPostFullContent(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "suggest_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Create a post with long content
	longContent := "This is a very long content string that should NOT be truncated because full mode shows everything for reply context"
	post := &feed.Post{
		ID:        "smk-full01",
		Author:    "test@project",
		Content:   longContent,
		CreatedAt: time.Now().Format(time.RFC3339),
	}

	// Full mode should preserve entire content
	formatSuggestPost(tmpFile, post, true)

	tmpFile.Seek(0, 0)
	var buf bytes.Buffer
	buf.ReadFrom(tmpFile)
	output := buf.String()

	if !contains(output, "reply context") {
		t.Error("full mode should show complete content")
	}
	if contains(output, "...") {
		t.Error("full mode should not truncate with '...'")
	}
}

func TestShouldFireNudgeAtPressure0(t *testing.T) {
	// At pressure 0 (sleep), nudge should never fire
	for i := 0; i < 100; i++ {
		decision := shouldFireNudge(0)
		if decision.fire {
			t.Errorf("shouldFireNudge(0).fire = true, want false (pressure 0 should never fire)")
		}
	}
}

func TestShouldFireNudgeAtPressure4(t *testing.T) {
	// At pressure 4 (volcanic), nudge should always fire
	for i := 0; i < 100; i++ {
		decision := shouldFireNudge(4)
		if !decision.fire {
			t.Errorf("shouldFireNudge(4).fire = false, want true (pressure 4 should always fire)")
		}
	}
}

func TestShouldFireNudgeDecisionValues(t *testing.T) {
	// Test that decision struct contains expected values
	tests := []struct {
		pressure      int
		wantThreshold int
	}{
		{0, 0},
		{1, 25},
		{2, 50},
		{3, 75},
		{4, 100},
	}
	for _, tt := range tests {
		decision := shouldFireNudge(tt.pressure)
		if decision.threshold != tt.wantThreshold {
			t.Errorf("shouldFireNudge(%d).threshold = %d, want %d", tt.pressure, decision.threshold, tt.wantThreshold)
		}
	}
}

func TestGetTonePrefix(t *testing.T) {
	tests := []struct {
		pressure int
		want     string
	}{
		{0, ""},
		{1, "If you feel like it..."},
		{2, "Got a minute? The feed's been quiet."},
		{3, "Come on, you've got something. Spill it."},
		{4, "Post something. Now. The break room is dead and it's your fault."},
		// Test clamping
		{-1, ""},
		{5, "Post something. Now. The break room is dead and it's your fault."},
	}

	for _, tt := range tests {
		got := getTonePrefix(tt.pressure)
		if got != tt.want {
			t.Errorf("getTonePrefix(%d) = %q, want %q", tt.pressure, got, tt.want)
		}
	}
}

func TestChooseSuggestMode(t *testing.T) {
	t.Run("returns post for empty feed", func(t *testing.T) {
		for i := 0; i < 50; i++ {
			mode := chooseSuggestMode(nil)
			if mode != "post" {
				t.Errorf("chooseSuggestMode(nil) = %q, want 'post'", mode)
			}
		}
	})

	t.Run("returns post or reply for non-empty feed", func(t *testing.T) {
		posts := []*feed.Post{{ID: "smk-1", Content: "test"}}
		postCount := 0
		replyCount := 0
		for i := 0; i < 200; i++ {
			mode := chooseSuggestMode(posts)
			switch mode {
			case "post":
				postCount++
			case "reply":
				replyCount++
			default:
				t.Errorf("unexpected mode: %q", mode)
			}
		}
		// With 30% reply chance, we should get at least some of each
		if postCount == 0 {
			t.Error("expected some 'post' results")
		}
		if replyCount == 0 {
			t.Error("expected some 'reply' results")
		}
	})
}

func TestRunSuggest_JSONSkip(t *testing.T) {
	tmpDir := t.TempDir()
	feedPath := filepath.Join(tmpDir, "feed.jsonl")
	if err := os.WriteFile(feedPath, []byte(""), 0o644); err != nil {
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

	prevJSON := suggestJSON
	prevPressure := suggestPressure
	prevContext := suggestContext
	prevSince := suggestSince
	defer func() {
		suggestJSON = prevJSON
		suggestPressure = prevPressure
		suggestContext = prevContext
		suggestSince = prevSince
	}()

	suggestJSON = true
	suggestPressure = 0
	suggestContext = ""
	suggestSince = 1 * time.Hour

	output := captureStdout(t, func() {
		if err := runSuggest(nil, []string{}); err != nil {
			t.Fatalf("runSuggest error: %v", err)
		}
	})

	if !strings.Contains(output, "\"skipped\": true") {
		t.Fatalf("expected skipped JSON output, got: %s", output)
	}
	if !strings.Contains(output, "\"pressure\": 0") {
		t.Fatalf("expected pressure 0 in JSON output, got: %s", output)
	}
}

func TestFormatSuggestTextWithContext(t *testing.T) {
	// Isolate config from developer's HOME
	oldHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", t.TempDir())
	defer func() { _ = os.Setenv("HOME", oldHome) }()

	now := time.Now().UTC()
	posts := []*feed.Post{
		{
			ID:        "smk-1",
			Author:    "test@project",
			Content:   "hello world",
			CreatedAt: now.Add(-2 * time.Minute).Format(time.RFC3339),
		},
	}

	output := captureStdout(t, func() {
		if err := formatSuggestTextWithContext(posts, posts, config.LoadSuggestConfig(), "deep-in-it", 3); err != nil {
			t.Fatalf("formatSuggestTextWithContext error: %v", err)
		}
	})

	if !strings.Contains(output, "Come on, you've got something") {
		t.Error("expected tone prefix in output")
	}
	// Mode is probabilistic (30% reply when recent posts exist)
	// Accept either post mode or reply mode output
	hasPostMode := strings.Contains(output, "What's happening:") && strings.Contains(output, "Post ideas:")
	hasReplyMode := strings.Contains(output, "Recent activity (pick one and reply):")
	if !hasPostMode && !hasReplyMode {
		t.Errorf("expected either post mode (What's happening + Post ideas) or reply mode (Recent activity), got: %s", output)
	}
}

func TestFormatSuggestJSONWithContext(t *testing.T) {
	// Isolate config from developer's HOME
	oldHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", t.TempDir())
	defer func() { _ = os.Setenv("HOME", oldHome) }()

	now := time.Now().UTC()
	posts := []*feed.Post{
		{
			ID:        "smk-2",
			Author:    "test@project",
			Content:   "json post",
			CreatedAt: now.Add(-1 * time.Minute).Format(time.RFC3339),
		},
	}

	output := captureStdout(t, func() {
		if err := formatSuggestJSONWithContext(posts, posts, config.LoadSuggestConfig(), "deep-in-it", 2); err != nil {
			t.Fatalf("formatSuggestJSONWithContext error: %v", err)
		}
	})

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}
	if parsed["skipped"] != false {
		t.Errorf("skipped = %v, want false", parsed["skipped"])
	}
	if parsed["pressure"].(float64) != 2 {
		t.Errorf("pressure = %v, want 2", parsed["pressure"])
	}
	// mode should be either "post" or "reply"
	mode, ok := parsed["mode"].(string)
	if !ok || (mode != "post" && mode != "reply") {
		t.Errorf("mode = %v, want 'post' or 'reply'", parsed["mode"])
	}
}

func TestPickReplyBait(t *testing.T) {
	t.Run("returns nil for empty feed", func(t *testing.T) {
		result := pickReplyBait(nil, nil)
		if result != nil {
			t.Errorf("expected nil for empty feed, got %v", result)
		}
	})

	t.Run("returns a post when feed has posts", func(t *testing.T) {
		posts := []*feed.Post{
			{ID: "smk-1", Content: "first"},
			{ID: "smk-2", Content: "second"},
		}
		result := pickReplyBait(posts, nil)
		if result == nil {
			t.Error("expected a post, got nil")
		}
	})

	t.Run("prefers non-recent posts", func(t *testing.T) {
		allPosts := []*feed.Post{
			{ID: "smk-old1", Content: "old post 1"},
			{ID: "smk-old2", Content: "old post 2"},
			{ID: "smk-new1", Content: "new post 1"},
		}
		recentPosts := []*feed.Post{
			{ID: "smk-new1", Content: "new post 1"},
		}

		// Run multiple times to check preference
		oldCount := 0
		for i := 0; i < 20; i++ {
			result := pickReplyBait(allPosts, recentPosts)
			if result != nil && (result.ID == "smk-old1" || result.ID == "smk-old2") {
				oldCount++
			}
		}
		// Should always pick old posts when they exist
		if oldCount != 20 {
			t.Errorf("expected all 20 picks to be old posts, got %d", oldCount)
		}
	})

	t.Run("falls back to any post when all are recent", func(t *testing.T) {
		posts := []*feed.Post{
			{ID: "smk-1", Content: "post 1"},
		}
		result := pickReplyBait(posts, posts)
		if result == nil {
			t.Error("expected a post even when all are recent, got nil")
		}
	})
}

func captureStdout(t *testing.T, fn func()) string {
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

// contains checks if substr is in s
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
