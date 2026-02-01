package logging

import (
	"log/slog"
)

// LogCommand logs a command invocation at INFO level
func LogCommand(cmd string, args []string) {
	Logger().Info("command invoked",
		slog.String("cmd", cmd),
		slog.Any("args", args),
	)
	Verbose("command invoked", slog.String("cmd", cmd), slog.Any("args", args))
}

// LogPostCreated logs a post creation event at INFO level
func LogPostCreated(id, author string) {
	Logger().Info("post created",
		slog.String("id", id),
		slog.String("author", author),
	)
	Verbose("post created", slog.String("id", id), slog.String("author", author))
}

// LogError logs an error at ERROR level
func LogError(msg string, err error) {
	Logger().Error(msg,
		slog.Any("error", err),
	)
	Verbose(msg, slog.Any("error", err))
}

// LogDebug logs a debug message (only if level is debug)
func LogDebug(msg string, args ...any) {
	Logger().Debug(msg, args...)
	Verbose(msg, args...)
}

// LogWarn logs a warning message
func LogWarn(msg string, args ...any) {
	Logger().Warn(msg, args...)
	Verbose(msg, args...)
}
