package cli

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/dreamiurg/smoke/internal/identity/templates"
)

func TestOutputTemplatesText(t *testing.T) {
	output := captureTemplatesStdout(t, func() {
		if err := outputTemplatesText(); err != nil {
			t.Fatalf("outputTemplatesText error: %v", err)
		}
	})

	if !strings.Contains(output, "Observations") {
		t.Error("expected Observations category in output")
	}
	if !strings.Contains(output, "•") {
		t.Error("expected bullet points in output")
	}
}

func TestOutputTemplatesJSON(t *testing.T) {
	output := captureTemplatesStdout(t, func() {
		if err := outputTemplatesJSON(); err != nil {
			t.Fatalf("outputTemplatesJSON error: %v", err)
		}
	})

	var parsed []templates.Template
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}
	if len(parsed) == 0 {
		t.Fatal("expected non-empty templates JSON")
	}
}

func TestRunTemplatesJSON(t *testing.T) {
	prev := templatesJSON
	templatesJSON = true
	defer func() { templatesJSON = prev }()

	output := captureTemplatesStdout(t, func() {
		if err := runTemplates(nil, []string{}); err != nil {
			t.Fatalf("runTemplates error: %v", err)
		}
	})

	if !strings.HasPrefix(strings.TrimSpace(output), "[") {
		t.Errorf("expected JSON output, got: %s", output)
	}
}

func TestRunTemplatesText(t *testing.T) {
	prev := templatesJSON
	templatesJSON = false
	defer func() { templatesJSON = prev }()

	output := captureTemplatesStdout(t, func() {
		if err := runTemplates(nil, []string{}); err != nil {
			t.Fatalf("runTemplates error: %v", err)
		}
	})

	if !strings.Contains(output, "Observations") {
		t.Errorf("expected text output, got: %s", output)
	}
}

func captureTemplatesStdout(t *testing.T, fn func()) string {
	t.Helper()
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w

	fn()

	_ = w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	return buf.String()
}
