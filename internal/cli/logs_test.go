package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestShowLogFile(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "test.log")

	// Create a log file with 10 lines
	var lines []string
	for i := 1; i <= 10; i++ {
		lines = append(lines, `{"level":"info","msg":"line `+string(rune('0'+i))+`"}`)
	}
	content := strings.Join(lines, "\n") + "\n"
	if err := os.WriteFile(logPath, []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	// Test showing last 5 lines
	err := showLogFile(logPath, 5)
	if err != nil {
		t.Errorf("showLogFile() error = %v", err)
	}
}

func TestShowLogFileNotExists(t *testing.T) {
	err := showLogFile("/nonexistent/path/to/log", 10)
	if err == nil {
		t.Error("showLogFile() expected error for non-existent file")
	}
}

func TestClearLogFile(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "test.log")

	// Create a log file with content
	content := `{"level":"info","msg":"test"}`
	if err := os.WriteFile(logPath, []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	// Clear the file
	err := clearLogFile(logPath)
	if err != nil {
		t.Errorf("clearLogFile() error = %v", err)
	}

	// Verify file is empty
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if len(data) != 0 {
		t.Errorf("clearLogFile() file should be empty, got %d bytes", len(data))
	}
}

func TestClearLogFileNotExists(t *testing.T) {
	// Should not error when file doesn't exist
	err := clearLogFile("/nonexistent/path/to/log")
	if err != nil {
		t.Errorf("clearLogFile() should not error for non-existent file, got %v", err)
	}
}
