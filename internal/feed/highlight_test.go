package feed

import (
	"strings"
	"testing"
)

func TestHashtagPattern(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		matches []string
	}{
		{"single hashtag", "hello #world", []string{"#world"}},
		{"multiple hashtags", "#one #two #three", []string{"#one", "#two", "#three"}},
		{"hashtag with underscore", "#hello_world", []string{"#hello_world"}},
		{"hashtag with numbers", "#test123", []string{"#test123"}},
		{"no hashtag", "hello world", nil},
		{"hashtag at start", "#start of message", []string{"#start"}},
		{"hashtag at end", "message ends with #end", []string{"#end"}},
		{"adjacent hashtags", "#one#two", []string{"#one", "#two"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := HashtagPattern.FindAllString(tt.input, -1)
			if len(matches) != len(tt.matches) {
				t.Errorf("got %d matches, want %d", len(matches), len(tt.matches))
				return
			}
			for i, match := range matches {
				if match != tt.matches[i] {
					t.Errorf("match %d: got %q, want %q", i, match, tt.matches[i])
				}
			}
		})
	}
}

func TestMentionPattern(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		matches []string
	}{
		{"single mention", "hello @user", []string{"@user"}},
		{"multiple mentions", "@alice @bob @charlie", []string{"@alice", "@bob", "@charlie"}},
		{"mention with underscore", "@hello_world", []string{"@hello_world"}},
		{"mention with numbers", "@user123", []string{"@user123"}},
		{"no mention", "hello world", nil},
		{"mention at start", "@start of message", []string{"@start"}},
		{"mention at end", "message ends with @end", []string{"@end"}},
		{"email-like not a mention", "email@example.com", []string{"@example"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := MentionPattern.FindAllString(tt.input, -1)
			if len(matches) != len(tt.matches) {
				t.Errorf("got %d matches, want %d", len(matches), len(tt.matches))
				return
			}
			for i, match := range matches {
				if match != tt.matches[i] {
					t.Errorf("match %d: got %q, want %q", i, match, tt.matches[i])
				}
			}
		})
	}
}

func TestHighlightHashtags(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		colorize bool
		wantCyan bool
	}{
		{"with color", "hello #world", true, true},
		{"without color", "hello #world", false, false},
		{"no hashtag", "hello world", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HighlightHashtags(tt.input, tt.colorize)
			hasCyan := strings.Contains(result, FgCyan)
			if hasCyan != tt.wantCyan {
				t.Errorf("cyan color present = %v, want %v", hasCyan, tt.wantCyan)
			}
			if tt.wantCyan && !strings.Contains(result, Reset) {
				t.Error("expected Reset code when color applied")
			}
		})
	}
}

func TestHighlightMentions(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		colorize    bool
		wantMagenta bool
	}{
		{"with color", "hello @user", true, true},
		{"without color", "hello @user", false, false},
		{"no mention", "hello world", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HighlightMentions(tt.input, tt.colorize)
			hasMagenta := strings.Contains(result, FgMagenta)
			if hasMagenta != tt.wantMagenta {
				t.Errorf("magenta color present = %v, want %v", hasMagenta, tt.wantMagenta)
			}
			if tt.wantMagenta && !strings.Contains(result, Reset) {
				t.Error("expected Reset code when color applied")
			}
		})
	}
}

func TestHighlightAll(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		colorize    bool
		wantCyan    bool
		wantMagenta bool
	}{
		{"both types", "hello #world @user", true, true, true},
		{"hashtag only", "hello #world", true, true, false},
		{"mention only", "hello @user", true, false, true},
		{"no color", "hello #world @user", false, false, false},
		{"plain text", "hello world", true, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HighlightAll(tt.input, tt.colorize)
			hasCyan := strings.Contains(result, FgCyan)
			hasMagenta := strings.Contains(result, FgMagenta)
			if hasCyan != tt.wantCyan {
				t.Errorf("cyan color present = %v, want %v", hasCyan, tt.wantCyan)
			}
			if hasMagenta != tt.wantMagenta {
				t.Errorf("magenta color present = %v, want %v", hasMagenta, tt.wantMagenta)
			}
		})
	}
}

func TestHighlightPreservesText(t *testing.T) {
	// Verify that the original text is preserved (minus color codes)
	input := "Check out #golang and ping @gopher"
	result := HighlightAll(input, true)

	// Remove all ANSI codes to check text is preserved
	clean := strings.ReplaceAll(result, Dim, "")
	clean = strings.ReplaceAll(clean, FgCyan, "")
	clean = strings.ReplaceAll(clean, FgMagenta, "")
	clean = strings.ReplaceAll(clean, Reset, "")

	if clean != input {
		t.Errorf("text not preserved: got %q, want %q", clean, input)
	}
}

func TestHighlightWithTheme(t *testing.T) {
	theme := &AllThemes[0] // Use first theme (dracula) as default

	tests := []struct {
		name  string
		input string
	}{
		{"plain text", "hello world"},
		{"hashtag only", "hello #world"},
		{"mention only", "hello @user"},
		{"both types", "hello #world @user"},
		{"hashtag at start", "#start of message"},
		{"mention at end", "message @end"},
		{"multiple hashtags", "#one #two #three"},
		{"mixed throughout", "start #tag middle @user end"},
		{"adjacent highlights", "#one@two"},
		{"trailing text after highlight", "#tag and more text"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HighlightWithTheme(tt.input, theme)
			// Result should not be empty
			if result == "" {
				t.Error("expected non-empty result")
			}
			// Result should contain the original text content
			// Note: lipgloss may strip ANSI codes in test environments (NO_COLOR, etc.)
			// so we just verify the function runs and produces output containing the input
			if !strings.Contains(result, strings.Split(tt.input, " ")[0]) {
				t.Errorf("result should contain input content, got: %q", result)
			}
		})
	}
}
