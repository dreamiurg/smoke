package logging

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// resetGlobalState resets the global logger state for testing
func resetGlobalState() {
	logger = nil
	writer = nil
	initOnce = sync.Once{}
	closeOnce = sync.Once{}
	warnedOnce = sync.Once{}
	verboseHandler = nil
	verboseEnabled = false
}

func TestLoggerReturnsDiscardWhenNotInitialized(t *testing.T) {
	resetGlobalState()

	l := Logger()
	if l == nil {
		t.Error("Logger() returned nil")
	}
	// Should not panic when logging
	l.Info("test message")
}

func TestInitWithDisabledLogging(t *testing.T) {
	resetGlobalState()

	cfg := Config{
		Level: LevelOff,
		Path:  "/tmp/test.log",
	}
	Init(cfg)
	defer Close()

	l := Logger()
	if l == nil {
		t.Error("Logger() returned nil")
	}
	// Should not panic when logging
	l.Info("test message")
}

func TestInitCreatesLogFile(t *testing.T) {
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

	// Log something
	Logger().Info("test message")

	// File should exist with content
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if len(data) == 0 {
		t.Error("log file is empty")
	}
	if !strings.Contains(string(data), "test message") {
		t.Errorf("log file should contain 'test message', got: %s", string(data))
	}
}

func TestInitOnlyOnce(t *testing.T) {
	resetGlobalState()

	dir := t.TempDir()
	path1 := filepath.Join(dir, "test1.log")
	path2 := filepath.Join(dir, "test2.log")

	cfg1 := Config{
		Level:    slog.LevelInfo,
		Path:     path1,
		MaxSize:  DefaultMaxSize,
		MaxFiles: DefaultMaxFiles,
	}
	cfg2 := Config{
		Level:    slog.LevelInfo,
		Path:     path2,
		MaxSize:  DefaultMaxSize,
		MaxFiles: DefaultMaxFiles,
	}

	Init(cfg1)
	Init(cfg2) // Should be ignored
	defer Close()

	Logger().Info("test message")

	// Only first path should have content
	data1, _ := os.ReadFile(path1)
	data2, _ := os.ReadFile(path2)

	if len(data1) == 0 {
		t.Error("first log file should have content")
	}
	if len(data2) != 0 {
		t.Error("second log file should be empty (second Init ignored)")
	}
}

func TestCloseOnlyOnce(t *testing.T) {
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

	// Close multiple times should not panic
	Close()
	Close()
	Close()
}

func TestVerboseMode(t *testing.T) {
	resetGlobalState()

	// Enable verbose mode
	SetVerbose(true)

	if !IsVerbose() {
		t.Error("IsVerbose() should return true after SetVerbose(true)")
	}

	// Should not panic
	Verbose("test verbose message")

	SetVerbose(false)
	if IsVerbose() {
		t.Error("IsVerbose() should return false after SetVerbose(false)")
	}
}
