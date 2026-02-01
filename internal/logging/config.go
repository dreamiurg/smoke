package logging

import (
	"log/slog"
	"os"
	"strings"
)

// Default logging configuration values
const (
	// DefaultMaxSize is the maximum log file size before rotation (1MB)
	DefaultMaxSize = 1 << 20 // 1MB

	// DefaultMaxFiles is the maximum number of rotated log files to keep
	DefaultMaxFiles = 5
)

// LevelOff represents disabled logging (higher than any slog level)
const LevelOff = slog.Level(100)

// Config holds logging configuration
type Config struct {
	// Level is the minimum log level to record
	Level slog.Level

	// Path is the log file path
	Path string

	// MaxSize is the maximum log file size before rotation
	MaxSize int64

	// MaxFiles is the maximum number of rotated log files
	MaxFiles int

	// Verbose enables debug output to stderr (from -v flag)
	Verbose bool
}

// LoadConfig loads logging configuration from environment variables
// SMOKE_LOG_LEVEL: debug, info, warn, error, off (default: info)
func LoadConfig(logPath string) Config {
	cfg := Config{
		Level:    slog.LevelInfo,
		Path:     logPath,
		MaxSize:  DefaultMaxSize,
		MaxFiles: DefaultMaxFiles,
		Verbose:  false,
	}

	levelStr := strings.ToLower(os.Getenv("SMOKE_LOG_LEVEL"))
	switch levelStr {
	case "debug":
		cfg.Level = slog.LevelDebug
	case "info", "":
		cfg.Level = slog.LevelInfo
	case "warn", "warning":
		cfg.Level = slog.LevelWarn
	case "error":
		cfg.Level = slog.LevelError
	case "off", "none", "disabled":
		cfg.Level = LevelOff
	default:
		// Invalid level - default to info, but this is worth noting
		cfg.Level = slog.LevelInfo
	}

	return cfg
}

// IsDisabled returns true if logging is disabled
func (c Config) IsDisabled() bool {
	return c.Level >= LevelOff
}
