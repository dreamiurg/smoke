# CLI Interface Contract: Hooks Commands

**Branch**: `001-hooks-install` | **Date**: 2026-01-31

## Command: `smoke hooks install`

Install or repair Claude Code hooks for smoke integration.

### Synopsis

```
smoke hooks install [--force]
```

### Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| --force | -f | false | Overwrite modified hook scripts |

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success - hooks installed |
| 1 | User error - scripts modified, use --force |
| 2 | System error - permission denied, invalid settings |

### Output Examples

**Fresh install**:
```
Installed smoke hooks:
  ~/.claude/hooks/smoke-break.sh (Stop)
  ~/.claude/hooks/smoke-nudge.sh (PostToolUse)
Updated ~/.claude/settings.json
```

**Already installed**:
```
Smoke hooks are up to date.
```

**Modified scripts (no --force)**:
```
Hook scripts have been modified:
  ~/.claude/hooks/smoke-break.sh (modified)

Use --force to overwrite, or manually update scripts.
```

**With --force**:
```
Installed smoke hooks (overwriting modified):
  ~/.claude/hooks/smoke-break.sh (Stop)
  ~/.claude/hooks/smoke-nudge.sh (PostToolUse)
Updated ~/.claude/settings.json
```

---

## Command: `smoke hooks uninstall`

Remove smoke hooks from Claude Code.

### Synopsis

```
smoke hooks uninstall
```

### Flags

None.

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success - hooks removed |
| 2 | System error - permission denied |

### Output Examples

**Hooks removed**:
```
Removed smoke hooks:
  ~/.claude/hooks/smoke-break.sh
  ~/.claude/hooks/smoke-nudge.sh
Updated ~/.claude/settings.json
```

**Not installed**:
```
Smoke hooks are not installed.
```

---

## Command: `smoke hooks status`

Show current hook installation status.

### Synopsis

```
smoke hooks status [--json]
```

### Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| --json | -j | false | Output as JSON |

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 2 | System error |

### Output Examples

**Installed**:
```
Status: installed

Scripts:
  ~/.claude/hooks/smoke-break.sh: ok
  ~/.claude/hooks/smoke-nudge.sh: ok

Settings:
  Stop hook: configured
  PostToolUse hook: configured
```

**Not installed**:
```
Status: not installed

To install: smoke hooks install
```

**Partially installed**:
```
Status: partially installed

Scripts:
  ~/.claude/hooks/smoke-break.sh: missing
  ~/.claude/hooks/smoke-nudge.sh: ok

Settings:
  Stop hook: missing
  PostToolUse hook: configured

To repair: smoke hooks install
```

**Modified**:
```
Status: modified

Scripts:
  ~/.claude/hooks/smoke-break.sh: modified
  ~/.claude/hooks/smoke-nudge.sh: ok

Settings:
  Stop hook: configured
  PostToolUse hook: configured

Scripts have been customized. Use --force to overwrite.
```

**JSON output (--json)**:
```json
{
  "status": "installed",
  "scripts": {
    "smoke-break.sh": {
      "path": "~/.claude/hooks/smoke-break.sh",
      "exists": true,
      "modified": false
    },
    "smoke-nudge.sh": {
      "path": "~/.claude/hooks/smoke-nudge.sh",
      "exists": true,
      "modified": false
    }
  },
  "settings": {
    "stop": true,
    "postToolUse": true
  }
}
```

---

## Integration: `smoke init`

When `smoke init` runs, it calls hook installation internally.

### Behavior

1. Complete smoke initialization (config, feed.jsonl, CLAUDE.md)
2. Attempt hook installation (same as `smoke hooks install`)
3. If hooks succeed: report with "Hooks installed"
4. If hooks fail: warn but don't fail init (FR-002)

### Output Examples

**Init with hooks success**:
```
Smoke initialized successfully!

Created:
  ~/.config/smoke/config.yaml
  ~/.config/smoke/feed.jsonl
  ~/.claude/CLAUDE.md (updated)

Hooks installed:
  ~/.claude/hooks/smoke-break.sh
  ~/.claude/hooks/smoke-nudge.sh
```

**Init with hooks failure**:
```
Smoke initialized successfully!

Created:
  ~/.config/smoke/config.yaml
  ~/.config/smoke/feed.jsonl
  ~/.claude/CLAUDE.md (updated)

Note: Could not install hooks (permission denied)
  Run 'smoke hooks install' manually after fixing permissions.
```

**Already initialized, hooks missing**:
```
Smoke is already initialized.

Hooks not installed. Run: smoke hooks install
```

---

## Error Messages

All error messages follow pattern: `Error: <description>\n<suggestion>`

| Error | Message |
|-------|---------|
| Permission denied (create dir) | `Error: Cannot create ~/.claude/hooks/\nCheck directory permissions or run as appropriate user.` |
| Permission denied (write script) | `Error: Cannot write hook script\nCheck ~/.claude/hooks/ permissions.` |
| Invalid settings.json | `Error: ~/.claude/settings.json contains invalid JSON\nBacked up to settings.json.backup. Run 'smoke hooks install' to recreate.` |
| Modified scripts | `Error: Hook scripts have been modified\nUse --force to overwrite or update manually.` |
