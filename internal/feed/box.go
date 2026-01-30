package feed

import (
	"strings"
)

// Box-drawing characters (rounded corners)
const (
	BoxTopLeft     = "╭"
	BoxTopRight    = "╮"
	BoxBottomLeft  = "╰"
	BoxBottomRight = "╯"
	BoxHorizontal  = "─"
	BoxVertical    = "│"
)

// DefaultBoxWidth is the default width for box rendering
const DefaultBoxWidth = 80

// BoxRenderer handles box-drawing for posts
type BoxRenderer struct {
	Width int
}

// NewBoxRenderer creates a box renderer with the given width
func NewBoxRenderer(width int) *BoxRenderer {
	if width <= 0 {
		width = DefaultBoxWidth
	}
	return &BoxRenderer{Width: width}
}

// TopBorder returns the top border line
func (b *BoxRenderer) TopBorder() string {
	inner := b.Width - 2 // subtract corners
	if inner < 0 {
		inner = 0
	}
	return BoxTopLeft + strings.Repeat(BoxHorizontal, inner) + BoxTopRight
}

// BottomBorder returns the bottom border line
func (b *BoxRenderer) BottomBorder() string {
	inner := b.Width - 2 // subtract corners
	if inner < 0 {
		inner = 0
	}
	return BoxBottomLeft + strings.Repeat(BoxHorizontal, inner) + BoxBottomRight
}

// WrapLine wraps content in vertical borders with padding
// Returns the line with borders on both sides
func (b *BoxRenderer) WrapLine(content string) string {
	// Calculate visible length (excluding ANSI codes)
	visibleLen := VisibleLength(content)

	// Inner width = total width - 2 borders - 2 padding spaces
	innerWidth := b.Width - 4
	if innerWidth < 0 {
		innerWidth = 0
	}

	// Pad content to fill inner width
	padding := innerWidth - visibleLen
	if padding < 0 {
		padding = 0
	}

	return BoxVertical + " " + content + strings.Repeat(" ", padding) + " " + BoxVertical
}

// WrapContent wraps multi-line content in a box
func (b *BoxRenderer) WrapContent(lines []string) []string {
	result := make([]string, 0, len(lines)+2)
	result = append(result, b.TopBorder())
	for _, line := range lines {
		result = append(result, b.WrapLine(line))
	}
	result = append(result, b.BottomBorder())
	return result
}

// VisibleLength returns the visible length of a string, excluding ANSI escape codes
func VisibleLength(s string) int {
	length := 0
	inEscape := false
	for _, r := range s {
		if r == '\033' {
			inEscape = true
			continue
		}
		if inEscape {
			if r == 'm' {
				inEscape = false
			}
			continue
		}
		length++
	}
	return length
}
