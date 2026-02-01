package logging

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"sync"
)

var (
	// global logger instance
	logger     *slog.Logger
	writer     *RotatingWriter
	initOnce   sync.Once
	closeOnce  sync.Once
	warnedOnce sync.Once

	// verbose handler for stderr output (-v flag)
	verboseHandler slog.Handler
	verboseEnabled bool
)

// Init initializes the global logger with the given configuration
// Safe to call multiple times - only the first call takes effect
func Init(cfg Config) {
	initOnce.Do(func() {
		initLogger(cfg)
	})
}

// initLogger performs the actual initialization
func initLogger(cfg Config) {
	// If logging is disabled, use discard handler
	if cfg.IsDisabled() {
		logger = slog.New(slog.NewJSONHandler(io.Discard, nil))
		return
	}

	// Try to create the rotating writer
	var err error
	writer, err = NewRotatingWriter(cfg.Path, cfg.MaxSize, cfg.MaxFiles)
	if err != nil {
		// Graceful degradation: warn once and use discard handler
		warnedOnce.Do(func() {
			_, _ = fmt.Fprintf(os.Stderr, "warning: logging disabled: %v\n", err)
		})
		logger = slog.New(slog.NewJSONHandler(io.Discard, nil))
		return
	}

	// Create the JSON handler with level filtering
	opts := &slog.HandlerOptions{
		Level: cfg.Level,
	}
	logger = slog.New(slog.NewJSONHandler(writer, opts))

	// Set up verbose handler if requested
	if cfg.Verbose {
		verboseEnabled = true
		verboseHandler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
	}
}

// Logger returns the global logger instance
// If Init hasn't been called, returns a no-op logger
func Logger() *slog.Logger {
	if logger == nil {
		// Return a discard logger if not initialized
		return slog.New(slog.NewJSONHandler(io.Discard, nil))
	}
	return logger
}

// Close flushes and closes the log file
// Safe to call multiple times - only the first call takes effect
func Close() error {
	var err error
	closeOnce.Do(func() {
		if writer != nil {
			err = writer.Close()
			writer = nil
		}
	})
	return err
}

// Verbose logs a debug message to stderr if verbose mode is enabled
// This is separate from the file logger and only outputs when -v flag is used
func Verbose(msg string, args ...any) {
	if verboseEnabled && verboseHandler != nil {
		// Create a temporary logger for this message
		verboseLogger := slog.New(verboseHandler)
		verboseLogger.Debug(msg, args...)
	}
}

// SetVerbose enables or disables verbose mode
// This is typically called after parsing the -v flag
func SetVerbose(enabled bool) {
	verboseEnabled = enabled
	if enabled && verboseHandler == nil {
		verboseHandler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
	}
}

// IsVerbose returns whether verbose mode is enabled
func IsVerbose() bool {
	return verboseEnabled
}
