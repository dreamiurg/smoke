package logging

import (
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLogCommand(t *testing.T) {
	resetGlobalState()

	dir := t.TempDir()
	path := filepath.Join(dir, "test.log")

	cfg := Config{
		Level:    slog.LevelInfo,
		Path:     path,
		MaxSize:  DefaultMaxSize,
		MaxFiles: DefaultMaxFiles,
	}
	Init(cfg)
	defer Close()

	// Log a command
	LogCommand("post", []string{"hello", "world"})

	// Verify log file contains the command
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if !strings.Contains(string(data), "command invoked") {
		t.Errorf("log should contain 'command invoked', got: %s", string(data))
	}
	if !strings.Contains(string(data), "post") {
		t.Errorf("log should contain 'post', got: %s", string(data))
	}
}

func TestLogPostCreated(t *testing.T) {
	resetGlobalState()

	dir := t.TempDir()
	path := filepath.Join(dir, "test.log")

	cfg := Config{
		Level:    slog.LevelInfo,
		Path:     path,
		MaxSize:  DefaultMaxSize,
		MaxFiles: DefaultMaxFiles,
	}
	Init(cfg)
	defer Close()

	// Log a post creation
	LogPostCreated("smk-abc123", "test-author")

	// Verify log file contains the post info
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if !strings.Contains(string(data), "post created") {
		t.Errorf("log should contain 'post created', got: %s", string(data))
	}
	if !strings.Contains(string(data), "smk-abc123") {
		t.Errorf("log should contain 'smk-abc123', got: %s", string(data))
	}
}

func TestLogError(t *testing.T) {
	resetGlobalState()

	dir := t.TempDir()
	path := filepath.Join(dir, "test.log")

	cfg := Config{
		Level:    slog.LevelInfo,
		Path:     path,
		MaxSize:  DefaultMaxSize,
		MaxFiles: DefaultMaxFiles,
	}
	Init(cfg)
	defer Close()

	// Log an error
	testErr := errors.New("test error")
	LogError("something failed", testErr)

	// Verify log file contains the error
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if !strings.Contains(string(data), "something failed") {
		t.Errorf("log should contain 'something failed', got: %s", string(data))
	}
	if !strings.Contains(string(data), "test error") {
		t.Errorf("log should contain 'test error', got: %s", string(data))
	}
}

func TestLogDebug(t *testing.T) {
	resetGlobalState()

	dir := t.TempDir()
	path := filepath.Join(dir, "test.log")

	cfg := Config{
		Level:    slog.LevelDebug, // Enable debug level
		Path:     path,
		MaxSize:  DefaultMaxSize,
		MaxFiles: DefaultMaxFiles,
	}
	Init(cfg)
	defer Close()

	// Log a debug message
	LogDebug("debug info", "key", "value")

	// Verify log file contains the debug message
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if !strings.Contains(string(data), "debug info") {
		t.Errorf("log should contain 'debug info', got: %s", string(data))
	}
}

func TestLogWarn(t *testing.T) {
	resetGlobalState()

	dir := t.TempDir()
	path := filepath.Join(dir, "test.log")

	cfg := Config{
		Level:    slog.LevelInfo,
		Path:     path,
		MaxSize:  DefaultMaxSize,
		MaxFiles: DefaultMaxFiles,
	}
	Init(cfg)
	defer Close()

	// Log a warning
	LogWarn("warning message", "detail", "some detail")

	// Verify log file contains the warning
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if !strings.Contains(string(data), "warning message") {
		t.Errorf("log should contain 'warning message', got: %s", string(data))
	}
}
