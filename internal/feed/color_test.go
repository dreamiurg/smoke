package feed

import (
	"strings"
	"testing"
)

func TestAuthorColor_Deterministic(t *testing.T) {
	// Same author should always get the same color
	author := "ember"
	color1 := AuthorColor(author)
	color2 := AuthorColor(author)

	if color1 != color2 {
		t.Errorf("AuthorColor not deterministic: got %q and %q for same author", color1, color2)
	}
}

func TestAuthorColor_DifferentAuthors(t *testing.T) {
	// Test that we get colors from the palette for various authors
	authors := []string{"ember", "slit", "toast", "witness", "chrome", "furiosa"}
	colors := make(map[string]bool)

	for _, author := range authors {
		color := AuthorColor(author)
		// Verify it's a valid palette color
		found := false
		for _, c := range AuthorPalette {
			if color == c {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("AuthorColor(%q) returned invalid color %q", author, color)
		}
		colors[color] = true
	}

	// With 6 authors and 6 colors, we should have good distribution
	// (not testing exact distribution, just that multiple colors are used)
	if len(colors) < 2 {
		t.Errorf("Expected multiple colors for different authors, got %d unique colors", len(colors))
	}
}

func TestAuthorPalette_ValidColors(t *testing.T) {
	// Verify palette contains valid ANSI codes
	for i, color := range AuthorPalette {
		if !strings.HasPrefix(color, "\033[") {
			t.Errorf("AuthorPalette[%d] = %q is not a valid ANSI escape", i, color)
		}
	}
}

func TestColorize_Basic(t *testing.T) {
	text := "hello"
	result := Colorize(text, FgRed)

	if !strings.HasPrefix(result, FgRed) {
		t.Errorf("Colorize should prefix with color code")
	}
	if !strings.Contains(result, text) {
		t.Errorf("Colorize should contain original text")
	}
	if !strings.HasSuffix(result, Reset) {
		t.Errorf("Colorize should suffix with reset")
	}
}

func TestColorize_MultipleCodes(t *testing.T) {
	text := "hello"
	result := Colorize(text, Bold, FgCyan)

	if !strings.HasPrefix(result, Bold) {
		t.Errorf("Colorize should start with first code")
	}
	if !strings.Contains(result, FgCyan) {
		t.Errorf("Colorize should include second code")
	}
	if !strings.HasSuffix(result, Reset) {
		t.Errorf("Colorize should suffix with reset")
	}
}

func TestColorize_NoCodes(t *testing.T) {
	text := "hello"
	result := Colorize(text)

	if result != text {
		t.Errorf("Colorize with no codes should return text unchanged, got %q", result)
	}
}

func TestColorize_Empty(t *testing.T) {
	result := Colorize("", FgRed)
	expected := FgRed + Reset

	if result != expected {
		t.Errorf("Colorize empty string: expected %q, got %q", expected, result)
	}
}

func TestANSIConstants(t *testing.T) {
	// Verify ANSI constants are correctly defined
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{"Reset", Reset, "\033[0m"},
		{"Bold", Bold, "\033[1m"},
		{"Dim", Dim, "\033[2m"},
		{"FgRed", FgRed, "\033[31m"},
		{"FgGreen", FgGreen, "\033[32m"},
		{"FgYellow", FgYellow, "\033[33m"},
		{"FgBlue", FgBlue, "\033[34m"},
		{"FgMagenta", FgMagenta, "\033[35m"},
		{"FgCyan", FgCyan, "\033[36m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("%s = %q, want %q", tt.name, tt.constant, tt.expected)
			}
		})
	}
}
