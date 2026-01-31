package feed

import (
	"os"

	"golang.org/x/term"
)

// DefaultTerminalWidth is the fallback width when detection fails
const DefaultTerminalWidth = 100

// GetTerminalWidth returns the current terminal width, or a default if detection fails
func GetTerminalWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || width <= 0 {
		return DefaultTerminalWidth
	}
	return width
}

// ColorMode represents the color output mode
type ColorMode int

const (
	// ColorAuto detects TTY and enables color only for interactive terminals
	ColorAuto ColorMode = iota
	// ColorAlways forces color output regardless of TTY
	ColorAlways
	// ColorNever disables color output
	ColorNever
)

// IsTerminal reports whether the given file descriptor is a terminal.
func IsTerminal(fd uintptr) bool {
	return term.IsTerminal(int(fd))
}

// ShouldColorize determines whether to use color based on the mode and TTY status.
func ShouldColorize(mode ColorMode) bool {
	switch mode {
	case ColorAlways:
		return true
	case ColorNever:
		return false
	default: // ColorAuto
		return IsTerminal(os.Stdout.Fd())
	}
}
