package logging

import (
	"log/slog"
	"strings"
	"time"
)

// CommandTracker tracks command execution for telemetry.
// Create with StartCommand, complete with Complete or Fail.
type CommandTracker struct {
	ctx       *Context
	name      string
	args      []string
	startTime time.Time
	metrics   []slog.Attr
}

// StartCommand begins tracking a command execution.
// Returns a tracker that must be completed with Complete() or Fail().
func StartCommand(name string, args []string) *CommandTracker {
	ctx := CaptureContext()
	t := &CommandTracker{
		ctx:       ctx,
		name:      name,
		args:      args,
		startTime: time.Now(),
	}

	// Log command start
	Logger().Info("command started",
		t.ctx.Attrs(),
		slog.Group("cmd",
			slog.String("name", name),
			slog.Any("args", args),
		),
	)
	Verbose("command started",
		slog.String("cmd.name", name),
		slog.Any("cmd.args", args),
	)

	return t
}

// SetIdentity sets identity fields on the tracker's context.
// Call after identity resolution.
func (t *CommandTracker) SetIdentity(identity, agent, project string) {
	t.ctx.SetIdentity(identity, agent, project)
}

// Caller returns the detected caller agent type.
func (t *CommandTracker) Caller() string {
	if t == nil || t.ctx == nil {
		return ""
	}
	return t.ctx.Caller
}

// AddMetric adds a metric to be included in the completion log.
func (t *CommandTracker) AddMetric(attr slog.Attr) {
	t.metrics = append(t.metrics, attr)
}

// AddPostMetrics adds post-related metrics.
func (t *CommandTracker) AddPostMetrics(id, author string) {
	t.metrics = append(t.metrics,
		slog.Group("post",
			slog.String("id", id),
			slog.String("author", author),
		),
	)
}

// AddFeedMetrics adds feed-related metrics.
func (t *CommandTracker) AddFeedMetrics(sizeBytes int64, postCount int) {
	t.metrics = append(t.metrics,
		slog.Group("feed",
			slog.Int64("size_bytes", sizeBytes),
			slog.Int("post_count", postCount),
		),
	)
}

// Complete logs successful command completion with duration.
func (t *CommandTracker) Complete() {
	duration := time.Since(t.startTime)

	attrs := make([]any, 0, 2+len(t.metrics))
	attrs = append(attrs,
		t.ctx.Attrs(),
		slog.Group("cmd",
			slog.String("name", t.name),
			slog.Any("args", t.args),
			slog.Int64("duration_ms", duration.Milliseconds()),
		),
	)

	// Add any collected metrics
	for _, m := range t.metrics {
		attrs = append(attrs, m)
	}

	Logger().Info("command completed", attrs...)
	Verbose("command completed",
		slog.String("cmd.name", t.name),
		slog.Int64("cmd.duration_ms", duration.Milliseconds()),
	)
}

// Fail logs command failure with error details.
func (t *CommandTracker) Fail(err error) {
	duration := time.Since(t.startTime)

	errType := categorizeError(err)

	attrs := make([]any, 0, 3+len(t.metrics))
	attrs = append(attrs,
		t.ctx.Attrs(),
		slog.Group("cmd",
			slog.String("name", t.name),
			slog.Any("args", t.args),
			slog.Int64("duration_ms", duration.Milliseconds()),
		),
		slog.Group("err",
			slog.String("message", err.Error()),
			slog.String("type", errType),
		),
	)

	// Add any collected metrics
	for _, m := range t.metrics {
		attrs = append(attrs, m)
	}

	Logger().Error("command failed", attrs...)
	Verbose("command failed",
		slog.String("cmd.name", t.name),
		slog.String("err.message", err.Error()),
		slog.Int64("cmd.duration_ms", duration.Milliseconds()),
	)
}

// categorizeError attempts to categorize an error for analysis.
func categorizeError(err error) string {
	if err == nil {
		return "none"
	}

	lowered := strings.ToLower(err.Error())

	// Check for common error patterns
	switch {
	case strings.Contains(lowered, "not initialized"):
		return "not_initialized"
	case strings.Contains(lowered, "permission"):
		return "permission"
	case strings.Contains(lowered, "not found"):
		return "not_found"
	case strings.Contains(lowered, "timeout"):
		return "timeout"
	case strings.Contains(lowered, "invalid"):
		return "invalid_input"
	case strings.Contains(lowered, "parse"):
		return "parse_error"
	case strings.Contains(lowered, "connection"):
		return "connection"
	default:
		return "unknown"
	}
}
