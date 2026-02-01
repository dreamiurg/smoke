package hooks

import "errors"

var (
	// ErrScriptsModified indicates hook scripts exist but differ from embedded version
	ErrScriptsModified = errors.New("hook scripts have been modified")

	// ErrPermissionDenied indicates cannot write to hooks directory or settings
	ErrPermissionDenied = errors.New("permission denied")

	// ErrInvalidSettings indicates settings.json contains invalid JSON
	ErrInvalidSettings = errors.New("invalid settings.json")
)
