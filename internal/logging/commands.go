package logging

import (
	"log/slog"
)

// LogPostCreated logs a post creation event.
// Use within a tracked command to correlate with command context.
func LogPostCreated(id, author string) {
	Logger().Info("post created",
		slog.Group("post",
			slog.String("id", id),
			slog.String("author", author),
		),
	)
	Verbose("post created",
		slog.String("post.id", id),
		slog.String("post.author", author),
	)
}

// LogFeedRead logs a feed read operation with metrics.
func LogFeedRead(sizeBytes int64, postCount int) {
	Logger().Info("feed read",
		slog.Group("feed",
			slog.Int64("size_bytes", sizeBytes),
			slog.Int("post_count", postCount),
		),
	)
	Verbose("feed read",
		slog.Int64("feed.size_bytes", sizeBytes),
		slog.Int("feed.post_count", postCount),
	)
}

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

// LogCommand is deprecated. Use StartCommand() instead.
// Kept for backward compatibility during migration.
func LogCommand(cmd string, args []string) {
	Logger().Info("command invoked",
		slog.Group("cmd",
			slog.String("name", cmd),
			slog.Any("args", args),
		),
	)
	Verbose("command invoked",
		slog.String("cmd.name", cmd),
		slog.Any("cmd.args", args),
	)
}
