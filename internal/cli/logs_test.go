package cli

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunLogs_Show(t *testing.T) {
	tmpDir := t.TempDir()
	logDir := filepath.Join(tmpDir, ".config", "smoke")
	if err := os.MkdirAll(logDir, 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	logPath := filepath.Join(logDir, "smoke.log")
	if err := os.WriteFile(logPath, []byte("first\nsecond\n"), 0o600); err != nil {
		t.Fatalf("write log: %v", err)
	}

	oldHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() {
		if oldHome == "" {
			_ = os.Unsetenv("HOME")
		} else {
			_ = os.Setenv("HOME", oldHome)
		}
	}()

	prevLines := logsLines
	prevTail := logsTail
	prevClear := logsClear
	defer func() {
		logsLines = prevLines
		logsTail = prevTail
		logsClear = prevClear
	}()

	logsLines = 10
	logsTail = false
	logsClear = false

	output := captureLogsStdout(t, func() {
		if err := runLogs(nil, []string{}); err != nil {
			t.Fatalf("runLogs error: %v", err)
		}
	})

	if !strings.Contains(output, "first") || !strings.Contains(output, "second") {
		t.Fatalf("expected log output, got: %s", output)
	}
}

func TestRunLogs_Clear(t *testing.T) {
	tmpDir := t.TempDir()
	logDir := filepath.Join(tmpDir, ".config", "smoke")
	if err := os.MkdirAll(logDir, 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	logPath := filepath.Join(logDir, "smoke.log")
	if err := os.WriteFile(logPath, []byte("line\n"), 0o600); err != nil {
		t.Fatalf("write log: %v", err)
	}

	oldHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() {
		if oldHome == "" {
			_ = os.Unsetenv("HOME")
		} else {
			_ = os.Setenv("HOME", oldHome)
		}
	}()

	prevLines := logsLines
	prevTail := logsTail
	prevClear := logsClear
	defer func() {
		logsLines = prevLines
		logsTail = prevTail
		logsClear = prevClear
	}()

	logsLines = 10
	logsTail = false
	logsClear = true

	output := captureLogsStdout(t, func() {
		if err := runLogs(nil, []string{}); err != nil {
			t.Fatalf("runLogs error: %v", err)
		}
	})

	if !strings.Contains(output, "Log file cleared") {
		t.Fatalf("expected clear message, got: %s", output)
	}
}

func TestRunLogs_NoFile(t *testing.T) {
	tmpDir := t.TempDir()
	logDir := filepath.Join(tmpDir, ".config", "smoke")
	if err := os.MkdirAll(logDir, 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	oldHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() {
		if oldHome == "" {
			_ = os.Unsetenv("HOME")
		} else {
			_ = os.Setenv("HOME", oldHome)
		}
	}()

	prevLines := logsLines
	prevTail := logsTail
	prevClear := logsClear
	defer func() {
		logsLines = prevLines
		logsTail = prevTail
		logsClear = prevClear
	}()

	logsLines = 10
	logsTail = false
	logsClear = false

	output := captureLogsStdout(t, func() {
		if err := runLogs(nil, []string{}); err != nil {
			t.Fatalf("runLogs error: %v", err)
		}
	})

	if !strings.Contains(output, "No log file found") {
		t.Fatalf("expected no-log message, got: %s", output)
	}
}

func captureLogsStdout(t *testing.T, fn func()) string {
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
