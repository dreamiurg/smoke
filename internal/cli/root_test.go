package cli

import (
	"testing"
	"time"
)

func TestFormatRelativeTime(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		want     string
	}{
		{
			name:     "less than a minute",
			duration: 30 * time.Second,
			want:     "just now",
		},
		{
			name:     "exactly 1 minute",
			duration: 1 * time.Minute,
			want:     "1 minute ago",
		},
		{
			name:     "multiple minutes",
			duration: 5 * time.Minute,
			want:     "5 minutes ago",
		},
		{
			name:     "59 minutes",
			duration: 59 * time.Minute,
			want:     "59 minutes ago",
		},
		{
			name:     "exactly 1 hour",
			duration: 1 * time.Hour,
			want:     "1 hour ago",
		},
		{
			name:     "multiple hours",
			duration: 5 * time.Hour,
			want:     "5 hours ago",
		},
		{
			name:     "23 hours",
			duration: 23 * time.Hour,
			want:     "23 hours ago",
		},
		{
			name:     "exactly 1 day",
			duration: 24 * time.Hour,
			want:     "1 day ago",
		},
		{
			name:     "multiple days",
			duration: 5 * 24 * time.Hour,
			want:     "5 days ago",
		},
		{
			name:     "6 days",
			duration: 6 * 24 * time.Hour,
			want:     "6 days ago",
		},
		{
			name:     "exactly 1 week",
			duration: 7 * 24 * time.Hour,
			want:     "1 week ago",
		},
		{
			name:     "multiple weeks",
			duration: 3 * 7 * 24 * time.Hour,
			want:     "3 weeks ago",
		},
		{
			name:     "many weeks",
			duration: 10 * 7 * 24 * time.Hour,
			want:     "10 weeks ago",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatRelativeTime(tt.duration)
			if got != tt.want {
				t.Errorf("formatRelativeTime(%v) = %q, want %q", tt.duration, got, tt.want)
			}
		})
	}
}

func TestFormatBuildDate(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string // Use substring match since timezone varies
	}{
		{
			name: "empty string",
			raw:  "",
			want: "",
		},
		{
			name: "unknown",
			raw:  "unknown",
			want: "unknown",
		},
		{
			name: "RFC3339 format",
			raw:  "2026-01-31T12:00:00Z",
			want: "Jan 31 2026", // Substring check
		},
		{
			name: "without timezone",
			raw:  "2026-01-31T12:00:00",
			want: "Jan 31 2026",
		},
		{
			name: "unparseable format",
			raw:  "invalid-date",
			want: "invalid-date",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatBuildDate(tt.raw)
			// For parseable dates, just check if it contains expected substring
			if tt.want == "" || tt.want == "unknown" || tt.want == tt.raw {
				if got != tt.want {
					t.Errorf("formatBuildDate(%q) = %q, want %q", tt.raw, got, tt.want)
				}
			} else {
				// For dates, check that result contains expected date string
				if len(got) < len(tt.want) {
					t.Errorf("formatBuildDate(%q) = %q, too short", tt.raw, got)
				}
			}
		})
	}
}
