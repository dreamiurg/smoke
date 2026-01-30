package feed

import (
	"testing"
)

func TestGenerateID(t *testing.T) {
	// Generate multiple IDs to check uniqueness
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id, err := GenerateID()
		if err != nil {
			t.Fatalf("GenerateID() unexpected error: %v", err)
		}

		// Check format
		if !ValidateID(id) {
			t.Errorf("GenerateID() produced invalid ID: %s", id)
		}

		// Check prefix
		if id[:4] != "smk-" {
			t.Errorf("GenerateID() = %s, want prefix 'smk-'", id)
		}

		// Check length
		if len(id) != 10 { // smk- (4) + 6 chars
			t.Errorf("GenerateID() = %s, length = %d, want 10", id, len(id))
		}

		// Check uniqueness
		if ids[id] {
			t.Errorf("GenerateID() produced duplicate ID: %s", id)
		}
		ids[id] = true
	}
}

func TestValidateID(t *testing.T) {
	tests := []struct {
		name string
		id   string
		want bool
	}{
		{
			name: "valid lowercase",
			id:   "smk-abc123",
			want: true,
		},
		{
			name: "valid uppercase",
			id:   "smk-ABC123",
			want: true,
		},
		{
			name: "valid mixed case",
			id:   "smk-AbC123",
			want: true,
		},
		{
			name: "valid all numbers",
			id:   "smk-123456",
			want: true,
		},
		{
			name: "valid all letters",
			id:   "smk-abcdef",
			want: true,
		},
		{
			name: "wrong prefix",
			id:   "xyz-abc123",
			want: false,
		},
		{
			name: "no prefix",
			id:   "abc123",
			want: false,
		},
		{
			name: "too short",
			id:   "smk-abc",
			want: false,
		},
		{
			name: "too long",
			id:   "smk-abc12345",
			want: false,
		},
		{
			name: "invalid characters",
			id:   "smk-abc!@#",
			want: false,
		},
		{
			name: "empty string",
			id:   "",
			want: false,
		},
		{
			name: "just prefix",
			id:   "smk-",
			want: false,
		},
		{
			name: "spaces",
			id:   "smk-ab c12",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateID(tt.id)
			if got != tt.want {
				t.Errorf("ValidateID(%q) = %v, want %v", tt.id, got, tt.want)
			}
		})
	}
}

func TestGenerateIDRandomness(t *testing.T) {
	// Generate many IDs and verify they have reasonable entropy
	ids := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		id, err := GenerateID()
		if err != nil {
			t.Fatalf("GenerateID() unexpected error: %v", err)
		}
		ids[i] = id
	}

	// Count unique IDs
	unique := make(map[string]bool)
	for _, id := range ids {
		unique[id] = true
	}

	// All 1000 IDs should be unique (statistically certain with base62^6 possibilities)
	if len(unique) != 1000 {
		t.Errorf("Generated %d unique IDs out of 1000, expected all unique", len(unique))
	}
}
