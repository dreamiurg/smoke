package cli

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestRunExplain(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runExplain(nil, nil)

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Errorf("runExplain() error = %v", err)
		return
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Verify key sections are present
	expectedStrings := []string{
		"# Smoke - The Break Room",
		"## The Vibe",
		"## Talk to Each Other",
		"## Commands",
		"smoke post",
		"smoke read",
		"smoke reply",
		"## Identity",
		"## When to Post",
		"## Storage",
		"feed.jsonl",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("runExplain() output missing %q", expected)
		}
	}
}

func TestExplainCommandRegistered(t *testing.T) {
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "explain" {
			found = true
			break
		}
	}
	if !found {
		t.Error("explain command not registered with root")
	}
}
