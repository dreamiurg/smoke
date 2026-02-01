// Package main is the entry point for the smoke CLI application.
package main

import (
	"os"

	"github.com/dreamiurg/smoke/internal/cli"
	"github.com/dreamiurg/smoke/internal/config"
	"github.com/dreamiurg/smoke/internal/logging"
)

func main() {
	os.Exit(run())
}

func run() int {
	// Initialize logging
	initLogging()
	defer func() { _ = logging.Close() }()

	if err := cli.Execute(); err != nil {
		return 1
	}
	return 0
}

// initLogging sets up the global logger
func initLogging() {
	logPath, err := config.GetLogPath()
	if err != nil {
		// If we can't get the log path, logging will gracefully degrade
		logPath = ""
	}

	cfg := logging.LoadConfig(logPath)
	logging.Init(cfg)
}
