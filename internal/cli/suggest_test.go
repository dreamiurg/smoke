package cli

import (
	"testing"
)

func TestSuggestCommand(t *testing.T) {
	tests := []struct {
		name    string
		context string
		wantErr bool
	}{
		{
			name:    "random context",
			context: "random",
			wantErr: false,
		},
		{
			name:    "completion context",
			context: "completion",
			wantErr: false,
		},
		{
			name:    "idle context",
			context: "idle",
			wantErr: false,
		},
		{
			name:    "mention context",
			context: "mention",
			wantErr: false,
		},
		{
			name:    "invalid context",
			context: "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suggestContext = tt.context
			err := runSuggest(nil, nil)

			if (err != nil) != tt.wantErr {
				t.Errorf("runSuggest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSuggestPromptsNotEmpty(t *testing.T) {
	// Verify all prompt arrays are non-empty
	if len(completionPrompts) == 0 {
		t.Error("completionPrompts should not be empty")
	}
	if len(idlePrompts) == 0 {
		t.Error("idlePrompts should not be empty")
	}
	if len(mentionPrompts) == 0 {
		t.Error("mentionPrompts should not be empty")
	}
	if len(randomPrompts) == 0 {
		t.Error("randomPrompts should not be empty")
	}
}
