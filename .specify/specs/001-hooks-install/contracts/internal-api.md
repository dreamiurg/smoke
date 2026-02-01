# Internal API Contract: hooks Package

**Branch**: `001-hooks-install` | **Date**: 2026-01-31

## Package: `internal/hooks`

Manages Claude Code hook installation, uninstallation, and status.

---

## Types

### HookEvent

```go
type HookEvent string

const (
    EventStop        HookEvent = "Stop"
    EventPostToolUse HookEvent = "PostToolUse"
)
```

### HookScript

```go
type HookScript struct {
    Name    string    // e.g., "smoke-break.sh"
    Event   HookEvent // Which Claude Code event triggers this
}
```

### ScriptStatus

```go
type ScriptStatus string

const (
    StatusOK       ScriptStatus = "ok"       // Installed, matches embedded
    StatusMissing  ScriptStatus = "missing"  // Not installed
    StatusModified ScriptStatus = "modified" // Installed but content differs
)
```

### InstallState

```go
type InstallState string

const (
    StateNotInstalled      InstallState = "not_installed"
    StateInstalled         InstallState = "installed"
    StateModified          InstallState = "modified"
    StatePartiallyInstalled InstallState = "partially_installed"
)
```

### Status

```go
type Status struct {
    State    InstallState
    Scripts  map[string]ScriptInfo  // keyed by script name
    Settings SettingsInfo
}

type ScriptInfo struct {
    Path     string       // Full install path
    Exists   bool
    Modified bool
    Status   ScriptStatus
}

type SettingsInfo struct {
    Stop        bool // true if Stop hook configured
    PostToolUse bool // true if PostToolUse hook configured
}
```

### InstallOptions

```go
type InstallOptions struct {
    Force bool // Overwrite modified scripts
}
```

---

## Functions

### Install

```go
func Install(opts InstallOptions) error
```

Install smoke hooks to Claude Code.

**Behavior**:
1. Create `~/.claude/hooks/` directory if needed
2. For each embedded script:
   - Check if file exists at install path
   - If exists and content differs:
     - If `opts.Force`: overwrite
     - Else: return `ErrScriptsModified`
   - Write script content
   - Set executable permissions (0755)
3. Update `~/.claude/settings.json`:
   - Read existing settings (or create empty)
   - Add smoke hook entries to appropriate events
   - Write settings atomically

**Returns**:
- `nil` on success
- `ErrScriptsModified` if scripts modified and !Force
- `ErrPermissionDenied` if cannot write
- `ErrInvalidSettings` if settings.json malformed (after backup)

---

### Uninstall

```go
func Uninstall() error
```

Remove smoke hooks from Claude Code.

**Behavior**:
1. Update `~/.claude/settings.json`:
   - Read existing settings
   - Remove smoke hook entries from events
   - Write settings atomically
2. For each hook script:
   - Remove file if exists
3. Optionally remove state directory

**Returns**:
- `nil` on success (including if not installed)
- `ErrPermissionDenied` if cannot write

---

### GetStatus

```go
func GetStatus() (*Status, error)
```

Get current hook installation status.

**Behavior**:
1. For each hook script:
   - Check if file exists
   - If exists, compare hash with embedded
2. Read settings.json:
   - Check for smoke hook entries
3. Determine overall state

**Returns**:
- `*Status` with complete information
- Error only on system failures (not for missing files)

---

### GetHooksDir

```go
func GetHooksDir() string
```

Return the Claude Code hooks directory path.

**Returns**: `~/.claude/hooks/` (expanded)

---

### GetSettingsPath

```go
func GetSettingsPath() string
```

Return the Claude Code settings file path.

**Returns**: `~/.claude/settings.json` (expanded)

---

## Errors

```go
var (
    // ErrScriptsModified indicates hook scripts exist but differ from embedded
    ErrScriptsModified = errors.New("hook scripts have been modified")

    // ErrPermissionDenied indicates cannot write to hooks directory or settings
    ErrPermissionDenied = errors.New("permission denied")

    // ErrInvalidSettings indicates settings.json contains invalid JSON
    ErrInvalidSettings = errors.New("invalid settings.json")
)
```

---

## Embedded Assets

```go
package hooks

import "embed"

//go:embed scripts/*.sh
var scripts embed.FS

// GetScriptContent returns the embedded content of a hook script
func GetScriptContent(name string) ([]byte, error)

// ListScripts returns all embedded hook scripts
func ListScripts() []HookScript
```

---

## Settings JSON Structure

The package manipulates this structure within settings.json:

```go
// SettingsHooks represents the hooks section of settings.json
type SettingsHooks struct {
    Stop        []SettingsHookEntry `json:"Stop,omitempty"`
    PostToolUse []SettingsHookEntry `json:"PostToolUse,omitempty"`
    // Other events preserved but not modified
}

type SettingsHookEntry struct {
    Matcher string       `json:"matcher"`
    Hooks   []HookConfig `json:"hooks"`
}

type HookConfig struct {
    Type    string `json:"type"`    // "command"
    Command string `json:"command"` // path to script
}
```

**Smoke Hook Detection**:
A hook entry is considered a smoke hook if `Command` contains:
- `smoke-break.sh`, or
- `smoke-nudge.sh`

This allows detection regardless of absolute path variations.
