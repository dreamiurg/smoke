package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/dreamiurg/smoke/internal/feed"
)

func TestSuggestCommand(t *testing.T) {
	tests := []struct {
		name    string
		context string
		wantErr bool
	}{
		{
			name:    "random context",
			context: "random",
			wantErr: false,
		},
		{
			name:    "completion context",
			context: "completion",
			wantErr: false,
		},
		{
			name:    "idle context",
			context: "idle",
			wantErr: false,
		},
		{
			name:    "mention context",
			context: "mention",
			wantErr: false,
		},
		{
			name:    "invalid context",
			context: "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suggestContext = tt.context
			err := runSuggest(nil, nil)

			if (err != nil) != tt.wantErr {
				t.Errorf("runSuggest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSuggestPromptsNotEmpty(t *testing.T) {
	// Test that all prompt functions return non-empty strings
	// with various inputs (nil feed stats, with stats)

	testPost := &feed.Post{
		Author:  "test-author",
		Content: "test content here",
	}

	tests := []struct {
		name    string
		fn      func() string
		wantLen int
	}{
		{"completion no stats", func() string { return getCompletionPrompt(0, nil) }, 10},
		{"completion with count", func() string { return getCompletionPrompt(5, nil) }, 10},
		{"completion with post", func() string { return getCompletionPrompt(0, testPost) }, 10},
		{"idle no stats", func() string { return getIdlePrompt(0, nil) }, 10},
		{"idle with count", func() string { return getIdlePrompt(5, nil) }, 10},
		{"idle with post", func() string { return getIdlePrompt(0, testPost) }, 10},
		{"mention", func() string { return getMentionPrompt() }, 10},
		{"random no stats", func() string { return getRandomPrompt(0, nil) }, 10},
		{"random with count", func() string { return getRandomPrompt(5, nil) }, 10},
		{"random with post", func() string { return getRandomPrompt(0, testPost) }, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fn()
			if len(result) < tt.wantLen {
				t.Errorf("prompt too short: got %q (len %d), want at least %d chars",
					result, len(result), tt.wantLen)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input    string
		max      int
		expected string
	}{
		{"short", 10, "short"},
		{"exactly10c", 10, "exactly10c"},
		{"this is a longer string", 10, "this is..."},
		{"hello", 5, "hello"},
		{"hello world", 5, "he..."},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := truncate(tt.input, tt.max)
			if result != tt.expected {
				t.Errorf("truncate(%q, %d) = %q, want %q",
					tt.input, tt.max, result, tt.expected)
			}
		})
	}
}

func TestGetFeedStats(t *testing.T) {
	tests := []struct {
		name             string
		setupFn          func(t *testing.T) func()
		expectedCount    int
		expectedLastPost *feed.Post
		description      string
	}{
		{
			name: "no initialization",
			setupFn: func(t *testing.T) func() {
				oldDir := os.Getenv("HOME")
				tmpDir := t.TempDir()
				os.Setenv("HOME", tmpDir)
				return func() { os.Setenv("HOME", oldDir) }
			},
			expectedCount:    0,
			expectedLastPost: nil,
			description:      "When smoke is not initialized, should return 0 count and nil",
		},
		{
			name: "empty feed",
			setupFn: func(t *testing.T) func() {
				oldDir := os.Getenv("HOME")
				tmpDir := t.TempDir()
				os.Setenv("HOME", tmpDir)
				smokeDir := filepath.Join(tmpDir, ".config", "smoke")
				os.MkdirAll(smokeDir, 0700)
				// Create empty feed file
				feedFile := filepath.Join(smokeDir, "feed.jsonl")
				os.WriteFile(feedFile, []byte{}, 0600)
				return func() { os.Setenv("HOME", oldDir) }
			},
			expectedCount:    0,
			expectedLastPost: nil,
			description:      "When feed is empty, should return 0 count and nil",
		},
		{
			name: "recent posts",
			setupFn: func(t *testing.T) func() {
				oldDir := os.Getenv("HOME")
				tmpDir := t.TempDir()
				os.Setenv("HOME", tmpDir)
				smokeDir := filepath.Join(tmpDir, ".config", "smoke")
				os.MkdirAll(smokeDir, 0700)
				feedFile := filepath.Join(smokeDir, "feed.jsonl")

				// Create posts with recent timestamps
				now := time.Now()
				post1 := feed.Post{
					ID:        "smk-ts0001",
					Author:    "alice",
					Project:   "test",
					Suffix:    "test",
					Content:   "recent post 1",
					CreatedAt: now.Add(-10 * time.Minute).Format(time.RFC3339),
				}
				post2 := feed.Post{
					ID:        "smk-ts0002",
					Author:    "bob",
					Project:   "test",
					Suffix:    "test",
					Content:   "recent post 2",
					CreatedAt: now.Add(-30 * time.Minute).Format(time.RFC3339),
				}
				post3 := feed.Post{
					ID:        "smk-ts0003",
					Author:    "charlie",
					Project:   "test",
					Suffix:    "test",
					Content:   "older post",
					CreatedAt: now.Add(-90 * time.Minute).Format(time.RFC3339),
				}

				content := ""
				for _, p := range []feed.Post{post3, post2, post1} {
					data, _ := json.Marshal(p)
					content += string(data) + "\n"
				}
				os.WriteFile(feedFile, []byte(content), 0600)

				return func() { os.Setenv("HOME", oldDir) }
			},
			expectedCount: 2, // post1 and post2 are within the hour
			expectedLastPost: &feed.Post{
				Author:  "charlie",
				Content: "older post",
			},
			description: "Should count posts within last hour and return most recent",
		},
		{
			name: "posts with invalid timestamps",
			setupFn: func(t *testing.T) func() {
				oldDir := os.Getenv("HOME")
				tmpDir := t.TempDir()
				os.Setenv("HOME", tmpDir)
				smokeDir := filepath.Join(tmpDir, ".config", "smoke")
				os.MkdirAll(smokeDir, 0700)
				feedFile := filepath.Join(smokeDir, "feed.jsonl")

				now := time.Now()
				post1 := feed.Post{
					ID:        "smk-ts0001",
					Author:    "alice",
					Project:   "test",
					Suffix:    "test",
					Content:   "valid timestamp",
					CreatedAt: now.Add(-10 * time.Minute).Format(time.RFC3339),
				}
				post2 := feed.Post{
					ID:        "smk-ts0002",
					Author:    "bob",
					Project:   "test",
					Suffix:    "test",
					Content:   "invalid timestamp",
					CreatedAt: "not-a-valid-timestamp",
				}

				data1, _ := json.Marshal(post1)
				data2, _ := json.Marshal(post2)
				content := string(data2) + "\n" + string(data1) + "\n"
				os.WriteFile(feedFile, []byte(content), 0600)

				return func() { os.Setenv("HOME", oldDir) }
			},
			expectedCount: 1, // only post1 has valid timestamp and is recent
			expectedLastPost: &feed.Post{
				Author:  "bob",
				Content: "invalid timestamp",
			},
			description: "Should skip posts with invalid timestamps but still return most recent post",
		},
		{
			name: "all old posts",
			setupFn: func(t *testing.T) func() {
				oldDir := os.Getenv("HOME")
				tmpDir := t.TempDir()
				os.Setenv("HOME", tmpDir)
				smokeDir := filepath.Join(tmpDir, ".config", "smoke")
				os.MkdirAll(smokeDir, 0700)
				feedFile := filepath.Join(smokeDir, "feed.jsonl")

				now := time.Now()
				post1 := feed.Post{
					ID:        "smk-ts0001",
					Author:    "alice",
					Project:   "test",
					Suffix:    "test",
					Content:   "very old post",
					CreatedAt: now.Add(-2 * time.Hour).Format(time.RFC3339),
				}

				data, _ := json.Marshal(post1)
				content := string(data) + "\n"
				os.WriteFile(feedFile, []byte(content), 0600)

				return func() { os.Setenv("HOME", oldDir) }
			},
			expectedCount: 0, // post is outside 1-hour window
			expectedLastPost: &feed.Post{
				Author:  "alice",
				Content: "very old post",
			},
			description: "Should not count old posts but still return most recent post",
		},
		{
			name: "store creation error",
			setupFn: func(t *testing.T) func() {
				oldDir := os.Getenv("HOME")
				// Set HOME to an invalid location that will cause GetFeedPath to fail
				os.Setenv("HOME", "")
				return func() { os.Setenv("HOME", oldDir) }
			},
			expectedCount:    0,
			expectedLastPost: nil,
			description:      "When store creation fails, should return 0 count and nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := tt.setupFn(t)
			defer cleanup()

			count, lastPost := getFeedStats()

			if count != tt.expectedCount {
				t.Errorf("%s: got count=%d, want %d", tt.description, count, tt.expectedCount)
			}

			if tt.expectedLastPost == nil {
				if lastPost != nil {
					t.Errorf("%s: got lastPost=%v, want nil", tt.description, lastPost)
				}
			} else {
				if lastPost == nil {
					t.Errorf("%s: got lastPost=nil, want non-nil", tt.description)
				} else {
					if lastPost.Author != tt.expectedLastPost.Author {
						t.Errorf("%s: author mismatch: got %q, want %q",
							tt.description, lastPost.Author, tt.expectedLastPost.Author)
					}
					if lastPost.Content != tt.expectedLastPost.Content {
						t.Errorf("%s: content mismatch: got %q, want %q",
							tt.description, lastPost.Content, tt.expectedLastPost.Content)
					}
				}
			}
		})
	}
}

func TestSelectRandomPrompt(t *testing.T) {
	prompts := []string{"prompt A", "prompt B", "prompt C"}

	t.Run("no activity", func(t *testing.T) {
		result := selectRandomPrompt(prompts, 0, true)
		// Should be one of the base prompts without prefix
		found := false
		for _, p := range prompts {
			if result == p {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("selectRandomPrompt() = %q, want one of %v", result, prompts)
		}
	})

	t.Run("with activity and showActivity true", func(t *testing.T) {
		result := selectRandomPrompt(prompts, 5, true)
		// Should start with activity count
		if len(result) < 20 || result[:1] != "5" {
			t.Errorf("selectRandomPrompt() = %q, expected to start with '5 posts'", result)
		}
	})

	t.Run("with activity but showActivity false", func(t *testing.T) {
		result := selectRandomPrompt(prompts, 5, false)
		// Should not include activity count
		found := false
		for _, p := range prompts {
			if result == p {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("selectRandomPrompt() = %q, want one of %v (no activity prefix)", result, prompts)
		}
	})
}

func TestGetWorkingPromptNoActivity(t *testing.T) {
	result := getWorkingPrompt(0, nil)

	// Should return one of the base prompts
	validPrompts := []string{
		"How's it going? Any unexpected twists? smoke post \"...\"",
		"Hit any walls? Found any shortcuts? smoke post \"...\"",
		"What's on your mind right now? smoke post \"...\"",
	}

	found := false
	for _, p := range validPrompts {
		if result == p {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("getWorkingPrompt(0, nil) = %q, want one of the base prompts", result)
	}
}

func TestGetWorkingPromptWithActivity(t *testing.T) {
	result := getWorkingPrompt(7, nil)

	// Should start with activity count
	expected := "7 posts in the last hour."
	if len(result) < len(expected) || result[:len(expected)] != expected {
		t.Errorf("getWorkingPrompt(7, nil) = %q, expected to start with %q", result, expected)
	}
}

func TestGetWorkingPromptRandomness(t *testing.T) {
	// Call 100 times and verify we get at least 2 different results
	results := make(map[string]bool)
	for i := 0; i < 100; i++ {
		result := getWorkingPrompt(0, nil)
		results[result] = true
	}

	if len(results) < 2 {
		t.Errorf("getWorkingPrompt() returned only %d unique results in 100 calls, expected at least 2", len(results))
	}
}

func TestSuggestWorkingContext(t *testing.T) {
	suggestContext = "working"
	err := runSuggest(nil, nil)
	if err != nil {
		t.Errorf("runSuggest() with working context: unexpected error: %v", err)
	}
}
