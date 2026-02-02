//go:build cgo && (darwin || (linux && !android))

// Package internal contains internal dependencies for the smoke package.
// This file imports cgo-dependent packages to ensure they stay in go.mod.
package internal

import (
	_ "github.com/fogleman/gg"
	_ "golang.design/x/clipboard"
)
