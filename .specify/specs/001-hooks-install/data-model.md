# Data Model: Hooks Installation System

**Branch**: `001-hooks-install` | **Date**: 2026-01-31

## Entities

### HookScript

Represents an embeddable hook script that runs in Claude Code.

| Field | Type | Description |
|-------|------|-------------|
| Name | string | Script filename (e.g., "smoke-break.sh") |
| Event | HookEvent | Claude Code event type (Stop, PostToolUse) |
| Content | []byte | Script content (from embed.FS) |
| InstallPath | string | Full path where script is installed |

**Validation Rules**:
- Name must end in ".sh"
- Name must be unique across all hooks
- Content must be non-empty
- Content must start with shebang (`#!/bin/bash`)

### HookEvent

Enumeration of supported Claude Code hook events.

| Value | Description |
|-------|-------------|
| Stop | Fires when Claude finishes responding |
| PostToolUse | Fires after each tool call |

**Note**: Other Claude Code events (PreToolUse, SessionStart, SessionEnd) are not used by smoke hooks.

### SettingsEntry

Represents a hook configuration entry in Claude Code settings.json.

| Field | Type | Description |
|-------|------|-------------|
| Matcher | string | Pattern to match (empty = all) |
| Hooks | []HookConfig | Array of hook configurations |

### HookConfig

Individual hook within a settings entry.

| Field | Type | Description |
|-------|------|-------------|
| Type | string | Always "command" for shell scripts |
| Command | string | Full path to script |

### InstallationState

Represents the current state of hook installation.

| Value | Scripts | Settings | Description |
|-------|---------|----------|-------------|
| NotInstalled | None exist | None exist | Fresh system |
| Installed | All match | All present | Up to date |
| Modified | Some differ | Present | User modified scripts |
| PartiallyInstalled | Some missing | Some missing | Incomplete state |

**State Transitions**:
```
NotInstalled --install--> Installed
Installed --uninstall--> NotInstalled
Installed --user-edit--> Modified
Modified --install-force--> Installed
Modified --install--> Error (use --force)
PartiallyInstalled --install--> Installed
```

## File Structures

### ~/.claude/settings.json (relevant portion)

```json
{
  "hooks": {
    "Stop": [
      {
        "matcher": "",
        "hooks": [
          {
            "type": "command",
            "command": "~/.claude/hooks/smoke-break.sh"
          }
        ]
      }
    ],
    "PostToolUse": [
      {
        "matcher": "",
        "hooks": [
          {
            "type": "command",
            "command": "~/.claude/hooks/smoke-nudge.sh"
          }
        ]
      }
    ]
  }
}
```

### Embedded Scripts Structure

```
internal/hooks/
├── hooks.go          # Hook management logic
├── embed.go          # Embed directives and asset access
├── scripts/
│   ├── smoke-break.sh
│   └── smoke-nudge.sh
└── hooks_test.go
```

## Relationships

```
┌──────────────┐     contains     ┌─────────────┐
│  embed.FS    │─────────────────▶│ HookScript  │
│  (scripts/)  │                  │ (2 scripts) │
└──────────────┘                  └──────┬──────┘
                                         │
                                  writes │ to disk
                                         ▼
                                  ┌──────────────┐
                                  │ ~/.claude/   │
                                  │   hooks/*.sh │
                                  └──────┬───────┘
                                         │
                              referenced │ by
                                         ▼
                                  ┌──────────────────┐
                                  │ settings.json    │
                                  │   hooks section  │
                                  └──────────────────┘
```

## Operations

### Install

1. Create `~/.claude/hooks/` if needed
2. For each HookScript:
   - Check if script exists at InstallPath
   - If exists, compare hash to embedded content
   - If modified and not --force, error
   - Write embedded content to InstallPath
   - Set executable permissions (0755)
3. Read settings.json (create if missing)
4. For each HookScript:
   - Find or create event entry
   - Add or update hook command entry
5. Write settings.json

### Uninstall

1. Read settings.json
2. For each HookScript:
   - Remove hook entry from event array
   - If event array empty, optionally remove event key
3. Write settings.json
4. For each HookScript:
   - Remove script file from InstallPath
5. Optionally clean up state directory

### Status

1. For each HookScript:
   - Check if script exists
   - If exists, compare hash
2. Read settings.json
3. For each HookScript:
   - Check if settings entry exists
4. Determine overall InstallationState
5. Report per-script status

## Error Conditions

| Condition | Detection | Recovery |
|-----------|-----------|----------|
| ~/.claude not writable | os.MkdirAll fails | Suggest sudo or permissions |
| settings.json invalid JSON | json.Unmarshal fails | Backup, create fresh |
| Script hash mismatch | SHA256 differs | Warn, suggest --force |
| Script not executable | os.Chmod fails | Suggest permissions |
| Partial read of settings | io.ReadFile fails | Return specific error |
