package hooks

// HookEvent represents a Claude Code hook event type
type HookEvent string

const (
	// EventStop fires when Claude finishes responding
	EventStop HookEvent = "Stop"
	// EventPostToolUse fires after each tool call
	EventPostToolUse HookEvent = "PostToolUse"
)

// ScriptStatus represents the status of an installed hook script
type ScriptStatus string

const (
	// StatusOK means script is installed and matches embedded version
	StatusOK ScriptStatus = "ok"
	// StatusMissing means script is not installed
	StatusMissing ScriptStatus = "missing"
	// StatusModified means script exists but content differs from embedded
	StatusModified ScriptStatus = "modified"
)

// InstallState represents the overall installation state
type InstallState string

const (
	// StateNotInstalled means no hooks are installed
	StateNotInstalled InstallState = "not_installed"
	// StateInstalled means all hooks are installed and up-to-date
	StateInstalled InstallState = "installed"
	// StateModified means hooks are installed but some have been customized
	StateModified InstallState = "modified"
	// StatePartiallyInstalled means some hooks are missing or incomplete
	StatePartiallyInstalled InstallState = "partially_installed"
)

// HookScript represents an embeddable hook script
type HookScript struct {
	Name  string    // Script filename (e.g., "smoke-break.sh")
	Event HookEvent // Which Claude Code event triggers this
}

// Status represents the complete hook installation status
type Status struct {
	State    InstallState
	Scripts  map[string]ScriptInfo
	Settings SettingsInfo
}

// ScriptInfo contains status information for a single script
type ScriptInfo struct {
	Path     string       // Full install path
	Exists   bool         // Whether file exists
	Modified bool         // Whether content differs from embedded
	Status   ScriptStatus // Overall status
}

// SettingsInfo contains Claude Code settings.json hook configuration status
type SettingsInfo struct {
	Stop        bool // true if Stop hook is configured
	PostToolUse bool // true if PostToolUse hook is configured
}

// InstallOptions configures hook installation behavior
type InstallOptions struct {
	Force bool // Overwrite modified scripts
}

// SettingsHooks represents the hooks section of Claude Code settings.json
type SettingsHooks struct {
	Stop        []SettingsHookEntry `json:"Stop,omitempty"`
	PostToolUse []SettingsHookEntry `json:"PostToolUse,omitempty"`
	// Other events are preserved but not modified by smoke
}

// SettingsHookEntry represents a hook configuration entry
type SettingsHookEntry struct {
	Matcher string       `json:"matcher"`
	Hooks   []HookConfig `json:"hooks"`
}

// HookConfig represents an individual hook within an entry
type HookConfig struct {
	Type    string `json:"type"`    // Always "command" for shell scripts
	Command string `json:"command"` // Full path to script
}
