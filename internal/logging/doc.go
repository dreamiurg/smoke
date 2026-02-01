// Package logging provides structured logging for smoke CLI.
//
// The package uses Go's log/slog for structured JSON logging with
// automatic file rotation. It supports log levels via the SMOKE_LOG_LEVEL
// environment variable and verbose mode via the -v flag.
//
// Log levels: debug, info (default), warn, error, off
//
// Example usage:
//
//	logging.Init(logging.LoadConfig(logPath))
//	defer logging.Close()
//	logging.LogCommand("post", []string{"hello"})
package logging
