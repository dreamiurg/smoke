package feed

import (
	"testing"
	"time"
)

func TestParseLocaleTimeFormat(t *testing.T) {
	tests := []struct {
		name     string
		locale   string
		expected TimeFormat
	}{
		// 12-hour locales
		{"US English", "en_US", TimeFormat12h},
		{"US English with UTF-8", "en_US.UTF-8", TimeFormat12h},
		{"Australian English", "en_AU.UTF-8", TimeFormat12h},
		{"Philippine English", "en_PH", TimeFormat12h},
		{"Indian English", "en_IN", TimeFormat12h},
		{"US Spanish", "es_US", TimeFormat12h},
		{"Mexican Spanish", "es_MX.UTF-8", TimeFormat12h},

		// 24-hour locales
		{"UK English", "en_GB", TimeFormat24h},
		{"UK English with UTF-8", "en_GB.UTF-8", TimeFormat24h},
		{"German", "de_DE.UTF-8", TimeFormat24h},
		{"French", "fr_FR.UTF-8", TimeFormat24h},
		{"Canadian English", "en_CA", TimeFormat24h},
		{"Spanish Spain", "es_ES", TimeFormat24h},
		{"Japanese", "ja_JP.UTF-8", TimeFormat24h},
		{"Chinese", "zh_CN.UTF-8", TimeFormat24h},

		// Edge cases
		{"POSIX locale", "C", TimeFormat24h},
		{"C.UTF-8 locale", "C.UTF-8", TimeFormat24h},
		{"POSIX explicit", "POSIX", TimeFormat24h},
		{"With modifier", "sr_RS@latin", TimeFormat24h},
		{"Empty string", "", TimeFormat24h},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseLocaleTimeFormat(tt.locale)
			if got != tt.expected {
				t.Errorf("parseLocaleTimeFormat(%q) = %v, want %v", tt.locale, got, tt.expected)
			}
		})
	}
}

func TestFormatTimeWithFormat(t *testing.T) {
	// Use local timezone for testing to avoid timezone conversion issues
	loc := time.Local
	testTime := time.Date(2024, 1, 15, 14, 30, 0, 0, loc)

	tests := []struct {
		name     string
		format   TimeFormat
		expected string
	}{
		{"24-hour format", TimeFormat24h, "14:30"},
		{"12-hour format", TimeFormat12h, "2:30PM"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatTimeWithFormat(testTime, tt.format)
			if got != tt.expected {
				t.Errorf("FormatTimeWithFormat(%v, %v) = %q, want %q", testTime, tt.format, got, tt.expected)
			}
		})
	}

	// Test morning time for AM
	morningTime := time.Date(2024, 1, 15, 9, 5, 0, 0, loc)
	got := FormatTimeWithFormat(morningTime, TimeFormat12h)
	if got != "9:05AM" {
		t.Errorf("FormatTimeWithFormat(morning) = %q, want %q", got, "9:05AM")
	}

	// Test midnight
	midnight := time.Date(2024, 1, 15, 0, 0, 0, 0, loc)
	got = FormatTimeWithFormat(midnight, TimeFormat12h)
	if got != "12:00AM" {
		t.Errorf("FormatTimeWithFormat(midnight) = %q, want %q", got, "12:00AM")
	}

	// Test noon
	noon := time.Date(2024, 1, 15, 12, 0, 0, 0, loc)
	got = FormatTimeWithFormat(noon, TimeFormat12h)
	if got != "12:00PM" {
		t.Errorf("FormatTimeWithFormat(noon) = %q, want %q", got, "12:00PM")
	}
}

func TestFormatTimeWidth(t *testing.T) {
	tests := []struct {
		format   TimeFormat
		expected int
	}{
		{TimeFormat24h, 5}, // "15:04"
		{TimeFormat12h, 7}, // "3:04PM" max
	}

	for _, tt := range tests {
		got := FormatTimeWidth(tt.format)
		if got != tt.expected {
			t.Errorf("FormatTimeWidth(%v) = %d, want %d", tt.format, got, tt.expected)
		}
	}
}

func TestDayLabel(t *testing.T) {
	now := time.Now().Local()
	today := time.Date(now.Year(), now.Month(), now.Day(), 10, 0, 0, 0, now.Location())
	yesterday := today.Add(-24 * time.Hour)
	twoDaysAgo := today.Add(-48 * time.Hour)
	weekAgo := today.Add(-7 * 24 * time.Hour)
	yearAgo := today.Add(-365 * 24 * time.Hour)

	tests := []struct {
		name     string
		time     time.Time
		expected string
	}{
		{"Today", today, "Today"},
		{"Yesterday", yesterday, "Yesterday"},
		{"Two days ago", twoDaysAgo, twoDaysAgo.Weekday().String()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DayLabel(tt.time)
			if got != tt.expected {
				t.Errorf("DayLabel(%v) = %q, want %q", tt.time, got, tt.expected)
			}
		})
	}

	// Test that week-old posts show weekday
	weekAgoLabel := DayLabel(weekAgo)
	expectedWeekDay := weekAgo.Format("Mon, Jan 2")
	if weekAgoLabel != expectedWeekDay {
		t.Errorf("DayLabel(weekAgo) = %q, want %q", weekAgoLabel, expectedWeekDay)
	}

	// Test that year-old posts include year
	yearAgoLabel := DayLabel(yearAgo)
	if yearAgoLabel != yearAgo.Format("Mon, Jan 2, 2006") {
		t.Errorf("DayLabel(yearAgo) = %q, want format with year", yearAgoLabel)
	}
}
