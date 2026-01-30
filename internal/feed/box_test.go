package feed

import (
	"strings"
	"testing"
)

func TestNewBoxRenderer(t *testing.T) {
	t.Run("with valid width", func(t *testing.T) {
		box := NewBoxRenderer(60)
		if box.Width != 60 {
			t.Errorf("NewBoxRenderer(60).Width = %d, want 60", box.Width)
		}
	})

	t.Run("with zero width uses default", func(t *testing.T) {
		box := NewBoxRenderer(0)
		if box.Width != DefaultBoxWidth {
			t.Errorf("NewBoxRenderer(0).Width = %d, want %d", box.Width, DefaultBoxWidth)
		}
	})

	t.Run("with negative width uses default", func(t *testing.T) {
		box := NewBoxRenderer(-10)
		if box.Width != DefaultBoxWidth {
			t.Errorf("NewBoxRenderer(-10).Width = %d, want %d", box.Width, DefaultBoxWidth)
		}
	})
}

func TestBoxRenderer_TopBorder(t *testing.T) {
	box := NewBoxRenderer(10)
	top := box.TopBorder()

	if !strings.HasPrefix(top, BoxTopLeft) {
		t.Errorf("TopBorder should start with %q, got %q", BoxTopLeft, top)
	}
	if !strings.HasSuffix(top, BoxTopRight) {
		t.Errorf("TopBorder should end with %q, got %q", BoxTopRight, top)
	}
	if !strings.Contains(top, BoxHorizontal) {
		t.Errorf("TopBorder should contain %q", BoxHorizontal)
	}
}

func TestBoxRenderer_BottomBorder(t *testing.T) {
	box := NewBoxRenderer(10)
	bottom := box.BottomBorder()

	if !strings.HasPrefix(bottom, BoxBottomLeft) {
		t.Errorf("BottomBorder should start with %q, got %q", BoxBottomLeft, bottom)
	}
	if !strings.HasSuffix(bottom, BoxBottomRight) {
		t.Errorf("BottomBorder should end with %q, got %q", BoxBottomRight, bottom)
	}
	if !strings.Contains(bottom, BoxHorizontal) {
		t.Errorf("BottomBorder should contain %q", BoxHorizontal)
	}
}

func TestBoxRenderer_WrapLine(t *testing.T) {
	box := NewBoxRenderer(20)
	line := box.WrapLine("test")

	if !strings.HasPrefix(line, BoxVertical) {
		t.Errorf("WrapLine should start with %q, got %q", BoxVertical, line)
	}
	if !strings.HasSuffix(line, BoxVertical) {
		t.Errorf("WrapLine should end with %q, got %q", BoxVertical, line)
	}
	if !strings.Contains(line, "test") {
		t.Errorf("WrapLine should contain content")
	}
}

func TestBoxRenderer_WrapContent(t *testing.T) {
	box := NewBoxRenderer(20)
	lines := box.WrapContent([]string{"line1", "line2"})

	if len(lines) != 4 { // top + 2 content + bottom
		t.Errorf("WrapContent should return 4 lines, got %d", len(lines))
	}

	// First line should be top border
	if !strings.HasPrefix(lines[0], BoxTopLeft) {
		t.Errorf("First line should be top border")
	}

	// Last line should be bottom border
	if !strings.HasPrefix(lines[3], BoxBottomLeft) {
		t.Errorf("Last line should be bottom border")
	}

	// Middle lines should be wrapped content
	if !strings.Contains(lines[1], "line1") {
		t.Errorf("Second line should contain 'line1'")
	}
	if !strings.Contains(lines[2], "line2") {
		t.Errorf("Third line should contain 'line2'")
	}
}

func TestVisibleLength(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"plain text", "hello", 5},
		{"empty string", "", 0},
		{"with color", "\033[31mhello\033[0m", 5},
		{"with bold and color", "\033[1m\033[31mhello\033[0m", 5},
		{"multiple colored words", "\033[31mred\033[0m \033[32mgreen\033[0m", 9},
		{"unicode chars", "héllo", 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := VisibleLength(tt.input)
			if result != tt.expected {
				t.Errorf("VisibleLength(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestBoxConstants(t *testing.T) {
	// Verify box-drawing constants are UTF-8 characters
	tests := []struct {
		name     string
		constant string
	}{
		{"BoxTopLeft", BoxTopLeft},
		{"BoxTopRight", BoxTopRight},
		{"BoxBottomLeft", BoxBottomLeft},
		{"BoxBottomRight", BoxBottomRight},
		{"BoxHorizontal", BoxHorizontal},
		{"BoxVertical", BoxVertical},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.constant) == 0 {
				t.Errorf("%s should not be empty", tt.name)
			}
			// Each should be a single rune (box drawing character)
			runes := []rune(tt.constant)
			if len(runes) != 1 {
				t.Errorf("%s should be a single character, got %d", tt.name, len(runes))
			}
		})
	}
}
