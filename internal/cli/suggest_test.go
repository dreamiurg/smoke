package cli

import (
	"testing"

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
