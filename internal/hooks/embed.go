package hooks

import "embed"

// scripts contains embedded hook script files
//
//go:embed scripts/*.sh
var scripts embed.FS
