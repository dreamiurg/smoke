package feed

import (
	"os"
	"strings"
	"time"
)

// TimeFormat represents 12-hour or 24-hour time format preference.
type TimeFormat int

const (
	// TimeFormat24h uses 24-hour clock (e.g., 14:30)
	TimeFormat24h TimeFormat = iota
	// TimeFormat12h uses 12-hour clock with AM/PM (e.g., 2:30 PM)
	TimeFormat12h
)

// locales12h lists locale prefixes that typically use 12-hour time format.
// This is not exhaustive but covers the most common cases.
var locales12h = map[string]bool{
	"en_US": true,
	"en_AU": true,
	"en_PH": true,
	"en_IN": true,
	"es_US": true,
	"es_MX": true,
}

// DetectTimeFormat returns the preferred time format based on locale settings.
// Checks LC_TIME, LC_ALL, and LANG environment variables in order.
func DetectTimeFormat() TimeFormat {
	// Check locale environment variables in order of specificity
	localeVars := []string{"LC_TIME", "LC_ALL", "LANG"}
	for _, varName := range localeVars {
		if locale := os.Getenv(varName); locale != "" {
			return parseLocaleTimeFormat(locale)
		}
	}
	// Default to 24h if no locale is set
	return TimeFormat24h
}

// parseLocaleTimeFormat determines time format from a locale string.
// Examples: "en_US.UTF-8", "de_DE", "C.UTF-8"
func parseLocaleTimeFormat(locale string) TimeFormat {
	// Handle POSIX/C locale - default to 24h
	if locale == "C" || locale == "POSIX" || strings.HasPrefix(locale, "C.") {
		return TimeFormat24h
	}

	// Extract the base locale (e.g., "en_US" from "en_US.UTF-8")
	base := locale
	if idx := strings.Index(locale, "."); idx > 0 {
		base = locale[:idx]
	}
	// Also strip @modifier if present (e.g., "sr_RS@latin")
	if idx := strings.Index(base, "@"); idx > 0 {
		base = base[:idx]
	}

	// Check if this locale uses 12-hour format
	if locales12h[base] {
		return TimeFormat12h
	}

	// Default to 24h for most locales (European standard)
	return TimeFormat24h
}

// FormatTime formats a time value according to the detected locale preference.
func FormatTime(t time.Time) string {
	return FormatTimeWithFormat(t, DetectTimeFormat())
}

// FormatTimeWithFormat formats a time value with the specified format preference.
func FormatTimeWithFormat(t time.Time, format TimeFormat) string {
	switch format {
	case TimeFormat12h:
		return t.Local().Format("3:04PM")
	default:
		return t.Local().Format("15:04")
	}
}

// FormatTimeWidth returns the expected width of a formatted timestamp.
// This is needed for column alignment in the TUI.
func FormatTimeWidth(format TimeFormat) int {
	switch format {
	case TimeFormat12h:
		return 7 // "3:04PM" max width (1:00AM to 12:59PM)
	default:
		return 5 // "15:04" fixed width
	}
}

// DayLabel returns a human-readable label for a date relative to today.
// Returns "Today", "Yesterday", or the formatted date for older dates.
func DayLabel(t time.Time) string {
	now := time.Now().Local()
	t = t.Local()

	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	postDay := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())

	diff := today.Sub(postDay).Hours() / 24

	switch {
	case diff < 1:
		return "Today"
	case diff < 2:
		return "Yesterday"
	case diff < 7:
		// Show weekday name for recent dates
		return t.Weekday().String()
	default:
		// Show full date for older posts
		if t.Year() == now.Year() {
			return t.Format("Mon, Jan 2")
		}
		return t.Format("Mon, Jan 2, 2006")
	}
}
