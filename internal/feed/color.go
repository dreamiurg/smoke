package feed

import (
	"hash/fnv"
	"io"
)

// ANSI escape sequences for terminal styling
const (
	Reset = "\033[0m"
	Bold  = "\033[1m"
	Dim   = "\033[2m"
)

// ANSI foreground colors
const (
	FgRed     = "\033[31m"
	FgGreen   = "\033[32m"
	FgYellow  = "\033[33m"
	FgBlue    = "\033[34m"
	FgMagenta = "\033[35m"
	FgCyan    = "\033[36m"
)

// AuthorPalette defines the colors used for author names.
// Excludes black (invisible on dark) and white (default text).
var AuthorPalette = []string{
	FgRed,
	FgGreen,
	FgYellow,
	FgBlue,
	FgMagenta,
	FgCyan,
}

// AuthorColor returns a deterministic color for the given author name.
// The same author always gets the same color.
func AuthorColor(author string) string {
	h := fnv.New32a()
	h.Write([]byte(author))
	idx := h.Sum32() % uint32(len(AuthorPalette))
	return AuthorPalette[idx]
}

// Colorize wraps text with the given ANSI codes and resets afterward.
// If color is empty, returns the text unchanged.
func Colorize(text string, codes ...string) string {
	if len(codes) == 0 {
		return text
	}
	var prefix string
	for _, code := range codes {
		prefix += code
	}
	return prefix + text + Reset
}

// ColorWriter wraps an io.Writer and conditionally applies color.
// When ColorEnabled is false, color functions return plain text.
type ColorWriter struct {
	W            io.Writer
	ColorEnabled bool
}

// NewColorWriter creates a ColorWriter with the given writer and color mode.
func NewColorWriter(w io.Writer, mode ColorMode) *ColorWriter {
	return &ColorWriter{
		W:            w,
		ColorEnabled: ShouldColorize(mode),
	}
}

// Colorize applies color codes only if color is enabled.
func (cw *ColorWriter) Colorize(text string, codes ...string) string {
	if !cw.ColorEnabled {
		return text
	}
	return Colorize(text, codes...)
}

// AuthorColor returns the colored author name if color is enabled.
func (cw *ColorWriter) AuthorColorize(author string) string {
	if !cw.ColorEnabled {
		return author
	}
	return Colorize(author, Bold, AuthorColor(author))
}

// Dim returns dimmed text if color is enabled.
func (cw *ColorWriter) Dim(text string) string {
	if !cw.ColorEnabled {
		return text
	}
	return Colorize(text, Dim)
}
