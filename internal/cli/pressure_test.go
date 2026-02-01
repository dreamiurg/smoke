package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/dreamiurg/smoke/internal/config"
)

func setupPressureEnv(t *testing.T) (cleanup func()) {
	t.Helper()

	tempDir := t.TempDir()
	origHome := os.Getenv("HOME")
	origSmokeName := os.Getenv("SMOKE_NAME")

	os.Setenv("HOME", tempDir)
	os.Setenv("SMOKE_NAME", "testbot@testproject")

	// Create smoke config
	configDir := filepath.Join(tempDir, ".config", "smoke")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}
	feedPath := filepath.Join(configDir, "feed.jsonl")
	if err := os.WriteFile(feedPath, []byte{}, 0644); err != nil {
		t.Fatalf("failed to create feed file: %v", err)
	}

	return func() {
		os.Setenv("HOME", origHome)
		os.Setenv("SMOKE_NAME", origSmokeName)
	}
}

func TestPressureCommandView(t *testing.T) {
	cleanup := setupPressureEnv(t)
	defer cleanup()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runPressure(nil, []string{})

	w.Close()
	os.Stdout = oldStdout

	assert.NoError(t, err)

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Verify output contains pressure information
	assert.Contains(t, output, "Nudge pressure:")
	assert.Contains(t, output, "Probability:")
	assert.Contains(t, output, "Tone:")
	assert.Contains(t, output, "Example nudge:")
	assert.Contains(t, output, "Adjust: smoke pressure")
	// Default is 2 (balanced)
	assert.Contains(t, output, "(50%)")
	assert.Contains(t, output, "⛅")
}

func TestPressureCommandSet(t *testing.T) {
	cleanup := setupPressureEnv(t)
	defer cleanup()

	tests := []struct {
		name          string
		level         string
		wantPressure  int
		wantEmoji     string
		wantPercent   string
		shouldContain string
	}{
		{"set to 0", "0", 0, "💤", "(0%)", "sleep"},
		{"set to 1", "1", 1, "🌙", "(25%)", "quiet"},
		{"set to 2", "2", 2, "⛅", "(50%)", "balanced"},
		{"set to 3", "3", 3, "☀️", "(75%)", "bright"},
		{"set to 4", "4", 4, "🌋", "(100%)", "volcanic"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset pressure to default first
			config.SetPressure(config.DefaultPressure)

			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := runPressure(nil, []string{tt.level})

			w.Close()
			os.Stdout = oldStdout

			assert.NoError(t, err)

			var buf bytes.Buffer
			buf.ReadFrom(r)
			output := buf.String()

			// Verify pressure was set
			assert.Contains(t, output, tt.shouldContain)
			assert.Contains(t, output, tt.wantEmoji)
			assert.Contains(t, output, tt.wantPercent)

			// Verify config was saved
			saved := config.GetPressure()
			assert.Equal(t, tt.wantPressure, saved)
		})
	}
}

func TestPressureCommandInvalid(t *testing.T) {
	cleanup := setupPressureEnv(t)
	defer cleanup()

	tests := []struct {
		name       string
		level      string
		wantError  bool
		wantErrMsg string
	}{
		{"not a number", "abc", true, "invalid pressure level"},
		{"negative", "-1", true, "out of range"},
		{"too high", "5", true, "out of range"},
		{"way too high", "100", true, "out of range"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := runPressure(nil, []string{tt.level})

			if tt.wantError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPressureCommandNotInitialized(t *testing.T) {
	tempDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", origHome)

	// Don't create smoke config
	err := runPressure(nil, []string{})

	assert.Error(t, err)
}

func TestPressureCommandRegistered(t *testing.T) {
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "pressure [level]" {
			found = true
			break
		}
	}
	assert.True(t, found, "pressure command not registered with root")
}

func TestPressureOutputFormat(t *testing.T) {
	cleanup := setupPressureEnv(t)
	defer cleanup()

	// Set pressure to known value
	config.SetPressure(2)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runPressure(nil, []string{})

	w.Close()
	os.Stdout = oldStdout

	assert.NoError(t, err)

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Verify sections are present and in order
	sections := []string{
		"Nudge pressure:",
		"Probability:",
		"Tone:",
		"Example nudge:",
		"Adjust: smoke pressure",
	}

	lastPos := -1
	for _, section := range sections {
		pos := strings.Index(output, section)
		assert.NotEqual(t, -1, pos, "section %q not found", section)
		assert.Greater(t, pos, lastPos, "section %q out of order", section)
		lastPos = pos
	}
}

func TestPressureReferenceTable(t *testing.T) {
	cleanup := setupPressureEnv(t)
	defer cleanup()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runPressure(nil, []string{})

	w.Close()
	os.Stdout = oldStdout

	assert.NoError(t, err)

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Verify all 5 levels are shown in table
	allEmojis := []string{"💤", "🌙", "⛅", "☀️", "🌋"}
	allLabels := []string{"sleep", "quiet", "balanced", "bright", "volcanic"}
	allPercents := []string{"0%", "25%", "50%", "75%", "100%"}

	for _, emoji := range allEmojis {
		assert.Contains(t, output, emoji, "emoji %q not in table", emoji)
	}

	for _, label := range allLabels {
		assert.Contains(t, output, label, "label %q not in table", label)
	}

	for _, percent := range allPercents {
		assert.Contains(t, output, percent, "percent %q not in table", percent)
	}
}

func TestPressurePersistence(t *testing.T) {
	cleanup := setupPressureEnv(t)
	defer cleanup()

	// Set pressure to 3
	err := runPressure(nil, []string{"3"})
	assert.NoError(t, err)

	// Verify it persists
	saved := config.GetPressure()
	assert.Equal(t, 3, saved)

	// Set to different value
	err = runPressure(nil, []string{"1"})
	assert.NoError(t, err)

	// Verify new value persists
	saved = config.GetPressure()
	assert.Equal(t, 1, saved)
}
