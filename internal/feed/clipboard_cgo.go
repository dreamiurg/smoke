//go:build cgo && !linux

// Package feed provides clipboard functionality for the Smoke TUI.
package feed

import (
	"sync"

	"golang.design/x/clipboard"
)

var (
	clipboardOnce    sync.Once
	clipboardInitErr error
)

// initClipboard initializes the clipboard library (only once)
func initClipboard() error {
	clipboardOnce.Do(func() {
		clipboardInitErr = clipboard.Init()
	})
	return clipboardInitErr
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
