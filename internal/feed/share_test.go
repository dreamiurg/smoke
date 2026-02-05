package feed

import (
	"bytes"
	"image/png"
	"strings"
	"testing"
)

func TestFormatPostAsText(t *testing.T) {
	t.Run("formats post with author and content", func(t *testing.T) {
		post, _ := NewPost("test-author", "test-project", "test-suffix", "Hello world!")
		result := FormatPostAsText(post)

		if !strings.Contains(result, "test-author") {
			t.Error("FormatPostAsText should include author")
		}
		if !strings.Contains(result, "Hello world!") {
			t.Error("FormatPostAsText should include content")
		}
		if !strings.Contains(result, ShareFooter) {
			t.Error("FormatPostAsText should include footer")
		}
	})

	t.Run("handles nil post", func(t *testing.T) {
		result := FormatPostAsText(nil)
		if result != "" {
			t.Error("FormatPostAsText(nil) should return empty string")
		}
	})

	t.Run("handles anonymous author", func(t *testing.T) {
		post := &Post{Content: "Test content"}
		result := FormatPostAsText(post)

		if !strings.Contains(result, "anonymous") {
			t.Error("FormatPostAsText should use 'anonymous' for empty author")
		}
	})
}

func TestRenderShareCard(t *testing.T) {
	post, _ := NewPost("test-author", "test-project", "test-suffix", "Hello world!")
	theme := GetTheme("dracula")

	t.Run("renders square image", func(t *testing.T) {
		data, err := RenderShareCard(post, theme, SquareImage)
		if err != nil {
			t.Fatalf("RenderShareCard failed: %v", err)
		}

		// Verify it's a valid PNG
		img, err := png.Decode(bytes.NewReader(data))
		if err != nil {
			t.Fatalf("Output is not valid PNG: %v", err)
		}

		// Verify dimensions
		bounds := img.Bounds()
		if bounds.Dx() != SquareImage.Width || bounds.Dy() != SquareImage.Height {
			t.Errorf("Square image dimensions = %dx%d, want %dx%d",
				bounds.Dx(), bounds.Dy(), SquareImage.Width, SquareImage.Height)
		}
	})

	t.Run("renders landscape image", func(t *testing.T) {
		data, err := RenderShareCard(post, theme, LandscapeImage)
		if err != nil {
			t.Fatalf("RenderShareCard failed: %v", err)
		}

		// Verify it's a valid PNG
		img, err := png.Decode(bytes.NewReader(data))
		if err != nil {
			t.Fatalf("Output is not valid PNG: %v", err)
		}

		// Verify dimensions
		bounds := img.Bounds()
		if bounds.Dx() != LandscapeImage.Width || bounds.Dy() != LandscapeImage.Height {
			t.Errorf("Landscape image dimensions = %dx%d, want %dx%d",
				bounds.Dx(), bounds.Dy(), LandscapeImage.Width, LandscapeImage.Height)
		}
	})

	t.Run("handles long content without error", func(t *testing.T) {
		longContent := strings.Repeat("This is a long line of text that should wrap across multiple lines. ", 8)
		if len(longContent) > MaxContentLength-1 {
			longContent = longContent[:MaxContentLength-1]
		}
		longPost, _ := NewPost("test-author", "test-project", "test-suffix", longContent)
		data, err := RenderShareCard(longPost, theme, LandscapeImage)
		if err != nil {
			t.Fatalf("RenderShareCard failed for long content: %v", err)
		}
		if len(data) == 0 {
			t.Fatal("RenderShareCard returned empty data for long content")
		}
	})

	t.Run("works with different themes", func(t *testing.T) {
		themes := []string{"dracula", "monokai", "nord", "gruvbox", "catppuccin"}
		for _, themeName := range themes {
			theme := GetTheme(themeName)
			_, err := RenderShareCard(post, theme, SquareImage)
			if err != nil {
				t.Errorf("RenderShareCard failed with theme %s: %v", themeName, err)
			}
		}
	})
}

func TestHexToColor(t *testing.T) {
	tests := []struct {
		hex     string
		r, g, b uint8
	}{
		{"#ff0000", 255, 0, 0},
		{"ff0000", 255, 0, 0},
		{"#00ff00", 0, 255, 0},
		{"#0000ff", 0, 0, 255},
		{"#ffffff", 255, 255, 255},
		{"#000000", 0, 0, 0},
		{"#282a36", 40, 42, 54}, // Dracula background
	}

	for _, tt := range tests {
		t.Run(tt.hex, func(t *testing.T) {
			c := hexToColor(tt.hex)
			r, g, b, _ := c.RGBA()
			// RGBA returns 16-bit values, convert to 8-bit
			r8, g8, b8 := uint8(r>>8), uint8(g>>8), uint8(b>>8)
			if r8 != tt.r || g8 != tt.g || b8 != tt.b {
				t.Errorf("hexToColor(%q) = RGB(%d,%d,%d), want RGB(%d,%d,%d)",
					tt.hex, r8, g8, b8, tt.r, tt.g, tt.b)
			}
		})
	}
}

func TestHexToColorEdgeCases(t *testing.T) {
	t.Run("empty string returns black", func(t *testing.T) {
		c := hexToColor("")
		r, g, b, _ := c.RGBA()
		if r != 0 || g != 0 || b != 0 {
			t.Error("hexToColor(\"\") should return black")
		}
	})

	t.Run("invalid length returns black", func(t *testing.T) {
		c := hexToColor("fff")
		r, g, b, _ := c.RGBA()
		if r != 0 || g != 0 || b != 0 {
			t.Error("hexToColor(\"fff\") should return black")
		}
	})
}

func TestImageDimensions(t *testing.T) {
	t.Run("square dimensions", func(t *testing.T) {
		if SquareImage.Width != 1200 || SquareImage.Height != 1200 {
			t.Errorf("SquareImage = %dx%d, want 1200x1200", SquareImage.Width, SquareImage.Height)
		}
		if SquareImage.Name != "square" {
			t.Errorf("SquareImage.Name = %q, want \"square\"", SquareImage.Name)
		}
	})

	t.Run("landscape dimensions", func(t *testing.T) {
		if LandscapeImage.Width != 1200 || LandscapeImage.Height != 630 {
			t.Errorf("LandscapeImage = %dx%d, want 1200x630", LandscapeImage.Width, LandscapeImage.Height)
		}
		if LandscapeImage.Name != "landscape" {
			t.Errorf("LandscapeImage.Name = %q, want \"landscape\"", LandscapeImage.Name)
		}
	})
}
