package logging

import (
	"log/slog"
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name      string
		envValue  string
		wantLevel slog.Level
	}{
		{"default", "", slog.LevelInfo},
		{"debug", "debug", slog.LevelDebug},
		{"DEBUG uppercase", "DEBUG", slog.LevelDebug},
		{"info", "info", slog.LevelInfo},
		{"warn", "warn", slog.LevelWarn},
		{"warning", "warning", slog.LevelWarn},
		{"error", "error", slog.LevelError},
		{"off", "off", LevelOff},
		{"none", "none", LevelOff},
		{"disabled", "disabled", LevelOff},
		{"invalid defaults to info", "invalid", slog.LevelInfo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set env var
			if tt.envValue != "" {
				os.Setenv("SMOKE_LOG_LEVEL", tt.envValue)
				defer os.Unsetenv("SMOKE_LOG_LEVEL")
			} else {
				os.Unsetenv("SMOKE_LOG_LEVEL")
			}

			cfg := LoadConfig("/tmp/test.log")
			if cfg.Level != tt.wantLevel {
				t.Errorf("LoadConfig().Level = %v, want %v", cfg.Level, tt.wantLevel)
			}
		})
	}
}

func TestConfigDefaults(t *testing.T) {
	os.Unsetenv("SMOKE_LOG_LEVEL")

	cfg := LoadConfig("/tmp/test.log")

	if cfg.MaxSize != DefaultMaxSize {
		t.Errorf("MaxSize = %d, want %d", cfg.MaxSize, DefaultMaxSize)
	}
	if cfg.MaxFiles != DefaultMaxFiles {
		t.Errorf("MaxFiles = %d, want %d", cfg.MaxFiles, DefaultMaxFiles)
	}
	if cfg.Path != "/tmp/test.log" {
		t.Errorf("Path = %q, want %q", cfg.Path, "/tmp/test.log")
	}
}

func TestConfigIsDisabled(t *testing.T) {
	tests := []struct {
		level    slog.Level
		disabled bool
	}{
		{slog.LevelDebug, false},
		{slog.LevelInfo, false},
		{slog.LevelWarn, false},
		{slog.LevelError, false},
		{LevelOff, true},
	}

	for _, tt := range tests {
		cfg := Config{Level: tt.level}
		if got := cfg.IsDisabled(); got != tt.disabled {
			t.Errorf("Config{Level: %v}.IsDisabled() = %v, want %v", tt.level, got, tt.disabled)
		}
	}
}
