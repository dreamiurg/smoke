//go:build !cgo || (!darwin && !linux) || android

// Package feed provides clipboard functionality for the Smoke TUI.
package feed

import "errors"

// ErrClipboardUnavailable indicates clipboard is not available on this platform.
var ErrClipboardUnavailable = errors.New("clipboard unavailable: built without cgo support")

// CopyTextToClipboard copies text to the system clipboard.
// Returns an error if clipboard is unavailable.
func CopyTextToClipboard(_ string) error {
	return ErrClipboardUnavailable
}

// CopyImageToClipboard copies PNG image data to the system clipboard.
// Returns an error if clipboard is unavailable.
func CopyImageToClipboard(_ []byte) error {
	return ErrClipboardUnavailable
}
