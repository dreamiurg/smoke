package feed

import "hash/fnv"

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
