//go:build !cgo

// Package feed provides clipboard functionality for the Smoke TUI.
// This is a stub for builds without CGO support.
package feed

import "errors"

// ErrClipboardNotAvailable is returned when clipboard is not available (no CGO)
var ErrClipboardNotAvailable = errors.New("clipboard not available: built without CGO support")

// CopyTextToClipboard copies text to the system clipboard.
// Returns an error if clipboard is unavailable.
func CopyTextToClipboard(text string) error {
	return ErrClipboardNotAvailable
}

// CopyImageToClipboard copies PNG image data to the system clipboard.
// Returns an error if clipboard is unavailable.
func CopyImageToClipboard(pngData []byte) error {
	return ErrClipboardNotAvailable
}
