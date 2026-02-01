// Package main is the entry point for the smoke CLI application.
package main

import (
	"os"

	"github.com/dreamiurg/smoke/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
