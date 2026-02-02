//go:build cgo && (darwin || (linux && !android))

// Package feed provides clipboard functionality for the Smoke TUI.
package feed

import (
	"golang.design/x/clipboard"
)

// clipboardInitialized tracks if clipboard.Init() has been called
var clipboardInitialized bool

// initClipboard initializes the clipboard library (only once)
func initClipboard() error {
	if clipboardInitialized {
		return nil
	}
	if err := clipboard.Init(); err != nil {
		return err
	}
	clipboardInitialized = true
	return nil
}

// CopyTextToClipboard copies text to the system clipboard.
// Returns an error if clipboard is unavailable.
func CopyTextToClipboard(text string) error {
	if err := initClipboard(); err != nil {
		return err
	}
	clipboard.Write(clipboard.FmtText, []byte(text))
	return nil
}

// CopyImageToClipboard copies PNG image data to the system clipboard.
// Returns an error if clipboard is unavailable.
func CopyImageToClipboard(pngData []byte) error {
	if err := initClipboard(); err != nil {
		return err
	}
	clipboard.Write(clipboard.FmtImage, pngData)
	return nil
}
