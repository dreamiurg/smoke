package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/dreamiurg/smoke/internal/config"
)

func TestWhoamiCommand(t *testing.T) {
	// Save original env var
	origSmokeName := os.Getenv("SMOKE_NAME")
	defer os.Setenv("SMOKE_NAME", origSmokeName)

	// Get the actual auto-detected project for assertions
	// @project override is now ignored, always auto-detected
	actualProject := detectProject()

	tests := []struct {
		name       string
		smokeName  string
		jsonFlag   bool
		nameFlag   bool
		wantOutput string
		wantJSON   map[string]string
	}{
		{
			name:       "default format with SMOKE_NAME (project override ignored)",
			smokeName:  "testbot@ignored-project",
			jsonFlag:   false,
			nameFlag:   false,
			wantOutput: "testbot@" + actualProject, // @project is auto-detected
		},
		{
			name:       "name only with SMOKE_NAME",
			smokeName:  "testbot@ignored-project",
			jsonFlag:   false,
			nameFlag:   true,
			wantOutput: "testbot", // agent is "", suffix is full name
		},
		{
			name:      "json format with SMOKE_NAME (project override ignored)",
			smokeName: "testbot@ignored-project",
			jsonFlag:  true,
			nameFlag:  false,
			wantJSON: map[string]string{
				"name":    "testbot",
				"project": actualProject, // @project is auto-detected
			},
		},
		{
			name:       "agent-suffix format (project override ignored)",
			smokeName:  "claude-swift-fox@ignored-project",
			jsonFlag:   false,
			nameFlag:   false,
			wantOutput: "claude-swift-fox@" + actualProject, // Full name as suffix, project auto-detected
		},
		{
			name:       "agent-suffix name only",
			smokeName:  "claude-swift-fox@ignored-project",
			jsonFlag:   false,
			nameFlag:   true,
			wantOutput: "claude-swift-fox", // Full name as suffix
		},
		{
			name:      "agent-suffix json format (project override ignored)",
			smokeName: "claude-swift-fox@ignored-project",
			jsonFlag:  true,
			nameFlag:  false,
			wantJSON: map[string]string{
				"name":    "claude-swift-fox",
				"project": actualProject, // @project is auto-detected
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment
			os.Setenv("SMOKE_NAME", tt.smokeName)

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

// detectProject is a test helper to get the auto-detected project name
func detectProject() string {
	id, err := config.GetIdentity("")
	if err != nil {
		return "unknown"
	}
	return id.Project
}
