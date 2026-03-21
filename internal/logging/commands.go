package logging

import (
	"log/slog"
)

// LogError logs an error with categorization.
// Prefer using CommandTracker.Fail() for command-level errors.
func LogError(msg string, err error) {
	errType := categorizeError(err)
	Logger().Error(msg,
		slog.Group("err",
			slog.String("message", err.Error()),
			slog.String("type", errType),
		),
	)
	Verbose(msg,
		slog.String("err.message", err.Error()),
		slog.String("err.type", errType),
	)
}

// LogDebug logs a debug message.
func LogDebug(msg string, args ...any) {
	Logger().Debug(msg, args...)
	Verbose(msg, args...)
}

// LogWarn logs a warning message.
func LogWarn(msg string, args ...any) {
	Logger().Warn(msg, args...)
	Verbose(msg, args...)
}
