package logging

import (
	"context"
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

// initLogger sets up the logger with lazy file creation
// Actual file creation is deferred until first log write
func initLogger(cfg Config) {
	// If logging is disabled, use discard handler immediately
	if cfg.IsDisabled() {
		logger = slog.New(slog.NewJSONHandler(io.Discard, nil))
		return
	}

	// Use a lazy handler that creates the file on first write
	logger = slog.New(&lazyHandler{cfg: cfg})

	// Set up verbose handler if requested
	if cfg.Verbose {
		verboseEnabled = true
		verboseHandler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
	}
}

// lazyHandler defers file creation until first log write
type lazyHandler struct {
	cfg     Config
	handler slog.Handler
	once    sync.Once
	mu      sync.Mutex
}

func (h *lazyHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.cfg.Level
}

func (h *lazyHandler) Handle(ctx context.Context, r slog.Record) error {
	h.ensureHandler()
	return h.handler.Handle(ctx, r)
}

func (h *lazyHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	h.ensureHandler()
	return h.handler.WithAttrs(attrs)
}

func (h *lazyHandler) WithGroup(name string) slog.Handler {
	h.ensureHandler()
	return h.handler.WithGroup(name)
}

func (h *lazyHandler) ensureHandler() {
	h.once.Do(func() {
		h.handler = h.createHandler()
	})
}

func (h *lazyHandler) createHandler() slog.Handler {
	h.mu.Lock()
	defer h.mu.Unlock()

	var err error
	writer, err = NewRotatingWriter(h.cfg.Path, h.cfg.MaxSize, h.cfg.MaxFiles)
	if err != nil {
		// Graceful degradation: warn once and use discard handler
		warnedOnce.Do(func() {
			_, _ = fmt.Fprintf(os.Stderr, "warning: logging disabled: %v\n", err)
		})
		return slog.NewJSONHandler(io.Discard, nil)
	}

	opts := &slog.HandlerOptions{
		Level: h.cfg.Level,
	}
	return slog.NewJSONHandler(writer, opts)
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
