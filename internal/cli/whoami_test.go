package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestWhoamiCommand(t *testing.T) {
	// Save original env vars
	origBDActor := os.Getenv("BD_ACTOR")
	origSmokeAuthor := os.Getenv("SMOKE_AUTHOR")
	defer func() {
		os.Setenv("BD_ACTOR", origBDActor)
		os.Setenv("SMOKE_AUTHOR", origSmokeAuthor)
	}()

	tests := []struct {
		name       string
		bdActor    string
		jsonFlag   bool
		nameFlag   bool
		wantOutput string
		wantJSON   map[string]string
	}{
		{
			name:       "default format with BD_ACTOR",
			bdActor:    "testbot@myproject",
			jsonFlag:   false,
			nameFlag:   false,
			wantOutput: "testbot@myproject",
		},
		{
			name:       "name only with BD_ACTOR",
			bdActor:    "testbot@myproject",
			jsonFlag:   false,
			nameFlag:   true,
			wantOutput: "testbot",
		},
		{
			name:     "json format with BD_ACTOR",
			bdActor:  "testbot@myproject",
			jsonFlag: true,
			nameFlag: false,
			wantJSON: map[string]string{
				"name":    "testbot",
				"project": "myproject",
			},
		},
		{
			name:       "agent-suffix format",
			bdActor:    "claude-swift-fox@smoke",
			jsonFlag:   false,
			nameFlag:   false,
			wantOutput: "claude-swift-fox@smoke",
		},
		{
			name:       "agent-suffix name only",
			bdActor:    "claude-swift-fox@smoke",
			jsonFlag:   false,
			nameFlag:   true,
			wantOutput: "claude-swift-fox",
		},
		{
			name:     "agent-suffix json format",
			bdActor:  "claude-swift-fox@smoke",
			jsonFlag: true,
			nameFlag: false,
			wantJSON: map[string]string{
				"name":    "claude-swift-fox",
				"project": "smoke",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment
			os.Setenv("BD_ACTOR", tt.bdActor)
			os.Setenv("SMOKE_AUTHOR", "")

			// Set flags
			whoamiJSON = tt.jsonFlag
			whoamiName = tt.nameFlag

			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := runWhoami(nil, nil)

			w.Close()
			os.Stdout = oldStdout

			if err != nil {
				t.Errorf("runWhoami() error = %v", err)
				return
			}

			var buf bytes.Buffer
			buf.ReadFrom(r)
			output := strings.TrimSpace(buf.String())

			if tt.jsonFlag {
				var got map[string]string
				if err := json.Unmarshal([]byte(output), &got); err != nil {
					t.Errorf("failed to parse JSON output: %v", err)
					return
				}
				for k, v := range tt.wantJSON {
					if got[k] != v {
						t.Errorf("JSON field %q = %q, want %q", k, got[k], v)
					}
				}
			} else {
				if output != tt.wantOutput {
					t.Errorf("output = %q, want %q", output, tt.wantOutput)
				}
			}
		})
	}
}

func TestWhoamiFlagsRegistered(t *testing.T) {
	// Test that flags are properly registered
	jsonFlag := whoamiCmd.Flags().Lookup("json")
	if jsonFlag == nil {
		t.Error("--json flag not registered")
	}

	nameFlag := whoamiCmd.Flags().Lookup("name")
	if nameFlag == nil {
		t.Error("--name flag not registered")
	}
}

func TestWhoamiCommandRegistered(t *testing.T) {
	// Test that whoami command is registered with root
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "whoami" {
			found = true
			break
		}
	}
	if !found {
		t.Error("whoami command not registered with root")
	}
}
