package feed

import (
	"testing"
)

func TestColorMode_Constants(t *testing.T) {
	// Verify the mode constants are distinct
	if ColorAuto == ColorAlways {
		t.Error("ColorAuto should not equal ColorAlways")
	}
	if ColorAuto == ColorNever {
		t.Error("ColorAuto should not equal ColorNever")
	}
	if ColorAlways == ColorNever {
		t.Error("ColorAlways should not equal ColorNever")
	}
}

func TestShouldColorize_Always(t *testing.T) {
	// ColorAlways should always return true
	result := ShouldColorize(ColorAlways)
	if !result {
		t.Error("ShouldColorize(ColorAlways) should return true")
	}
}

func TestShouldColorize_Never(t *testing.T) {
	// ColorNever should always return false
	result := ShouldColorize(ColorNever)
	if result {
		t.Error("ShouldColorize(ColorNever) should return false")
	}
}

func TestShouldColorize_Auto(_ *testing.T) {
	// ColorAuto depends on TTY status
	// In test environment, typically not a TTY (piped)
	result := ShouldColorize(ColorAuto)
	// We just verify it returns a bool without panicking
	// The actual value depends on test environment
	_ = result
}

func TestIsTerminal(_ *testing.T) {
	// In test environment, stdout is typically not a TTY
	// This just verifies the function doesn't panic
	result := IsTerminal(1) // stdout fd
	// Don't assert the value since it depends on test runner
	_ = result
}

func TestGetTerminalWidth(t *testing.T) {
	// GetTerminalWidth should return a positive value
	// Either the actual terminal width or the default
	width := GetTerminalWidth()
	if width <= 0 {
		t.Errorf("GetTerminalWidth() = %d, want > 0", width)
	}
	// In non-TTY test environment, should return default
	if width != DefaultTerminalWidth {
		t.Logf("GetTerminalWidth() returned actual width: %d", width)
	}
}

func TestDefaultTerminalWidth(t *testing.T) {
	// Verify default is a reasonable value
	if DefaultTerminalWidth < 80 {
		t.Errorf("DefaultTerminalWidth = %d, want >= 80", DefaultTerminalWidth)
	}
	if DefaultTerminalWidth > 200 {
		t.Errorf("DefaultTerminalWidth = %d, want <= 200", DefaultTerminalWidth)
	}
}
