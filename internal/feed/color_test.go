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

func TestColorWriter_ColorEnabled(t *testing.T) {
	var buf strings.Builder

	t.Run("color always", func(t *testing.T) {
		cw := NewColorWriter(&buf, ColorAlways)
		if !cw.ColorEnabled {
			t.Error("ColorWriter with ColorAlways should have ColorEnabled=true")
		}
	})

	t.Run("color never", func(t *testing.T) {
		cw := NewColorWriter(&buf, ColorNever)
		if cw.ColorEnabled {
			t.Error("ColorWriter with ColorNever should have ColorEnabled=false")
		}
	})
}

func TestColorWriter_Colorize(t *testing.T) {
	var buf strings.Builder

	t.Run("with color enabled", func(t *testing.T) {
		cw := NewColorWriter(&buf, ColorAlways)
		result := cw.Colorize("hello", FgRed)
		if !strings.Contains(result, FgRed) {
			t.Errorf("Colorize should include color code when enabled")
		}
		if !strings.Contains(result, Reset) {
			t.Errorf("Colorize should include reset when enabled")
		}
	})

	t.Run("with color disabled", func(t *testing.T) {
		cw := NewColorWriter(&buf, ColorNever)
		result := cw.Colorize("hello", FgRed)
		if result != "hello" {
			t.Errorf("Colorize should return plain text when disabled, got %q", result)
		}
	})
}

func TestColorWriter_AuthorColorize(t *testing.T) {
	var buf strings.Builder

	t.Run("with color enabled", func(t *testing.T) {
		cw := NewColorWriter(&buf, ColorAlways)
		result := cw.AuthorColorize("ember")
		if !strings.Contains(result, Bold) {
			t.Errorf("AuthorColorize should include bold when enabled")
		}
		if !strings.Contains(result, "ember") {
			t.Errorf("AuthorColorize should include author name")
		}
	})

	t.Run("with color disabled", func(t *testing.T) {
		cw := NewColorWriter(&buf, ColorNever)
		result := cw.AuthorColorize("ember")
		if result != "ember" {
			t.Errorf("AuthorColorize should return plain author when disabled, got %q", result)
		}
	})
}

func TestColorWriter_Dim(t *testing.T) {
	var buf strings.Builder

	t.Run("with color enabled", func(t *testing.T) {
		cw := NewColorWriter(&buf, ColorAlways)
		result := cw.Dim("timestamp")
		if !strings.Contains(result, Dim) {
			t.Errorf("Dim should include dim code when enabled")
		}
	})

	t.Run("with color disabled", func(t *testing.T) {
		cw := NewColorWriter(&buf, ColorNever)
		result := cw.Dim("timestamp")
		if result != "timestamp" {
			t.Errorf("Dim should return plain text when disabled, got %q", result)
		}
	})
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

func TestColorizeIdentity(t *testing.T) {
	theme := GetTheme("tomorrow-night")
	contrast := GetContrastLevel("medium")

	t.Run("simple identity without project", func(t *testing.T) {
		result := ColorizeIdentity("agent", theme, contrast)
		if result == "" {
			t.Error("ColorizeIdentity() should return non-empty result")
		}
		// Note: lipgloss may strip ANSI codes in non-TTY test environments
		// We verify the function doesn't crash and returns the input
		if !strings.Contains(result, "agent") {
			t.Errorf("ColorizeIdentity() should contain the agent name, got: %q", result)
		}
	})

	t.Run("identity with project", func(t *testing.T) {
		result := ColorizeIdentity("agent@project", theme, contrast)
		if result == "" {
			t.Error("ColorizeIdentity() should return non-empty result")
		}
		// Should contain the @ separator somewhere in the styled output
		if !strings.Contains(result, "@") {
			t.Error("ColorizeIdentity() should preserve @ separator")
		}
	})

	t.Run("high contrast with colored project", func(t *testing.T) {
		highContrast := GetContrastLevel("high")
		result := ColorizeIdentity("agent@project", theme, highContrast)
		if result == "" {
			t.Error("ColorizeIdentity() with high contrast should return non-empty result")
		}
	})

	t.Run("low contrast", func(t *testing.T) {
		lowContrast := GetContrastLevel("low")
		result := ColorizeIdentity("agent@project", theme, lowContrast)
		if result == "" {
			t.Error("ColorizeIdentity() with low contrast should return non-empty result")
		}
	})

	t.Run("different themes", func(t *testing.T) {
		themes := []string{"tomorrow-night", "monokai", "dracula", "solarized-light"}
		for _, themeName := range themes {
			th := GetTheme(themeName)
			result := ColorizeIdentity("agent", th, contrast)
			if result == "" {
				t.Errorf("ColorizeIdentity() with theme %q should return non-empty result", themeName)
			}
		}
	})

	t.Run("deterministic for same input", func(t *testing.T) {
		result1 := ColorizeIdentity("test-agent", theme, contrast)
		result2 := ColorizeIdentity("test-agent", theme, contrast)
		if result1 != result2 {
			t.Error("ColorizeIdentity() should be deterministic for same input")
		}
	})
}
