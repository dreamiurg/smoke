# Research: Hooks Installation System

**Branch**: `001-hooks-install` | **Date**: 2026-01-31

## Research Summary

This document resolves technical questions identified during planning for the hooks installation feature.

---

## 1. Go Embed for Script Assets

**Decision**: Use `//go:embed` directive to bundle hook scripts in the smoke binary.

**Rationale**:
- Go 1.16+ provides native `embed` package for embedding files at compile time
- Zero runtime dependencies - scripts are part of the binary
- Simple pattern matching for multiple files
- Works with `embed.FS` for filesystem-like access

**Implementation Pattern**:
```go
package hooks

import "embed"

//go:embed scripts/*.sh
var Scripts embed.FS

// Access: Scripts.ReadFile("scripts/smoke-break.sh")
```

**Alternatives Considered**:
- Raw string constants: Rejected - harder to maintain, no file validation
- External files: Rejected - adds distribution complexity, violates "zero config" principle
- Build-time generation: Rejected - adds complexity, embed is simpler

---

## 2. Claude Code Settings.json Format

**Decision**: Use the documented hook format with array-based event handlers.

**Observed Structure** (from ~/.claude/settings.json):
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
    "PostToolUse": [...]
  }
}
```

**Key Observations**:
- Each event type maps to an array of matcher objects
- Each matcher has a `matcher` string (empty = match all) and `hooks` array
- Each hook has `type: "command"` and `command` field
- Multiple matchers per event type are supported
- Must merge with existing hooks (FR-006, FR-007)

**Alternatives Considered**:
- Replacing entire hooks section: Rejected - would destroy user's other hooks
- Using hook scripts directly without settings entry: Rejected - won't fire

---

## 3. Script Location Strategy

**Decision**: Install to `~/.claude/hooks/` directory.

**Rationale**:
- Standard location for Claude Code hooks
- Matches current manual installation
- Easy to find for advanced users who want to customize
- Consistent with Claude Code conventions

**Path Structure**:
- `~/.claude/hooks/smoke-break.sh` - Stop hook (completion nudge)
- `~/.claude/hooks/smoke-nudge.sh` - PostToolUse hook (mid-session)
- `~/.claude/hooks/smoke-nudge-state/` - State directory (created by nudge hook)

**Alternatives Considered**:
- `~/.config/smoke/hooks/`: Rejected - Claude Code expects hooks in `~/.claude/hooks/`
- Embedding scripts directly in settings.json: Not supported by Claude Code

---

## 4. Idempotent Installation Logic

**Decision**: Check both file content (hash) and settings entry presence.

**Installation States**:
| State | Scripts | Settings | Action |
|-------|---------|----------|--------|
| not-installed | Missing | Missing | Full install |
| installed | Match | Present | Report up-to-date |
| modified | Differ | Present | Warn, offer --force |
| partially-installed | One missing | One missing | Repair (install missing) |

**Hash Comparison**:
- Compare SHA256 of installed script vs embedded script
- Different hash = modified by user
- --force flag overwrites regardless of modification

**Alternatives Considered**:
- Always overwrite: Rejected - would lose user customizations silently
- Version markers in scripts: Rejected - adds complexity, hash is sufficient

---

## 5. JSON Merging Strategy

**Decision**: Use Go's encoding/json with careful field preservation.

**Merge Algorithm**:
1. Read existing settings.json (handle missing/invalid)
2. Parse into `map[string]interface{}` for flexibility
3. Navigate to `hooks` section (create if missing)
4. For each smoke event (Stop, PostToolUse):
   - Get existing array for event (or empty)
   - Check if smoke hook entry exists (by command path)
   - Add if missing, update if present
5. Marshal back with `MarshalIndent` for readability
6. Write atomically (temp file + rename)

**Hook Detection**:
- Match on command containing `smoke-break.sh` or `smoke-nudge.sh`
- This allows detecting smoke hooks regardless of full path

**Alternatives Considered**:
- External JSON library (jsonparser, gjson): Rejected - standard library sufficient
- Full settings struct: Rejected - too brittle if Claude adds new fields

---

## 6. Error Handling Strategy

**Decision**: Graceful degradation with clear messaging per FR-002.

**Error Categories**:

| Error | During init | During hooks install |
|-------|-------------|---------------------|
| Can't create ~/.claude/ | Warn, continue init | Error |
| Can't write hook scripts | Warn, continue init | Error |
| Invalid settings.json | Warn, create backup, continue | Error with recovery hint |
| Permission denied | Warn, suggest fix | Error with suggestion |

**Recovery Actions**:
- Backup invalid settings.json to settings.json.backup
- Suggest `smoke hooks install --force` for repairs
- Never lose user data silently

---

## 7. Hook Script Design

**Decision**: Use existing proven scripts with minimal modification.

**Current Scripts** (already working in production):

**smoke-break.sh** (Stop hook):
- Fires on completion
- Counts tool calls since last human message
- Threshold: >15 tools triggers nudge
- Calls `smoke suggest --context=completion`
- Falls back gracefully if smoke not available

**smoke-nudge.sh** (PostToolUse hook):
- Fires after each tool use
- Tracks state per session in `~/.claude/hooks/smoke-nudge-state/`
- Dual threshold: >50 tools AND >10 mins since last prompt
- Calls `smoke suggest --context=working`
- State cleanup: per-session files

**Required Changes for Embedding**:
- None significant - scripts are stable
- May add version comment for tracking
- Ensure `set -euo pipefail` for robustness (already in nudge, add to break)

---

## 8. Command Structure

**Decision**: Use `smoke hooks <subcommand>` pattern.

**Commands**:
```
smoke hooks install   # Install/repair hooks (--force to overwrite)
smoke hooks uninstall # Remove hooks
smoke hooks status    # Show installation state
```

**Integration with init**:
- `smoke init` calls hooks install internally
- Failure doesn't fail init (FR-002)
- Reports "hooks installed" or "hooks failed: <reason>"

**Alternatives Considered**:
- `smoke install-hooks`: Rejected - less discoverable than subcommand
- `smoke config hooks`: Rejected - hooks aren't config, they're integration

---

## Open Questions Resolved

All NEEDS CLARIFICATION items have been resolved:
- [x] Embed mechanism: Go embed directive
- [x] Settings format: Documented array structure
- [x] Script location: ~/.claude/hooks/
- [x] Idempotency: Hash comparison + settings check
- [x] JSON handling: Standard library with map[string]interface{}
- [x] Error strategy: Graceful degradation for init, strict for hooks commands
