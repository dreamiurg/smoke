//go:build cgo

// Package internal contains internal dependencies for the smoke package.
// This file is only compiled with CGO enabled to bring in cgo-dependent libraries.
package internal

import (
	_ "github.com/fogleman/gg"
	_ "golang.design/x/clipboard"
)
