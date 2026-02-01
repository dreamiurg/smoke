// Package logging provides structured telemetry logging for smoke CLI.
//
// The package uses Go's log/slog for structured JSON logging with
// automatic file rotation. It supports log levels via the SMOKE_LOG_LEVEL
// environment variable and verbose mode via the -v flag.
//
// # Log Levels
//
// Supported levels: debug, info (default), warn, error, off
//
// # Telemetry Schema
//
// All log entries use consistent field naming organized into groups:
//
//	ctx.identity      Full identity string (e.g., "swift-fox@smoke")
//	ctx.agent         Agent type: "claude", "human", "unknown"
//	ctx.session       Session ID for correlation
//	ctx.env           Environment: "claude_code", "ci", "terminal"
//	ctx.project       Project name
//	ctx.cwd           Working directory
//
//	cmd.name          Command name (e.g., "post", "feed")
//	cmd.args          Command arguments
//	cmd.duration_ms   Execution time in milliseconds
//
//	post.id           Post ID
//	post.author       Post author
//
//	feed.size_bytes   Feed file size
//	feed.post_count   Number of posts in feed
//	feed.mode         Feed mode: "normal", "tail", "tui"
//
//	err.message       Error message
//	err.type          Categorized error type
//
// # Example Usage
//
//	// Initialize logging
//	logging.Init(logging.LoadConfig(logPath))
//	defer logging.Close()
//
//	// Track a command with full telemetry
//	tracker := logging.StartCommand("post", args)
//	tracker.SetIdentity(identity.String(), identity.Agent, identity.Project)
//	// ... do work ...
//	tracker.AddPostMetrics(post.ID, post.Author)
//	tracker.Complete() // or tracker.Fail(err)
//
// # Event Messages
//
// Standard event messages:
//   - "command started"   - logged at command start
//   - "command completed" - logged on success with duration
//   - "command failed"    - logged on error with err.* fields
//   - "post created"      - logged when a post is created
//   - "feed read"         - logged when feed is read
package logging
