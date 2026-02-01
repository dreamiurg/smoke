# Quickstart: Hooks Installation System

**Branch**: `001-hooks-install` | **Date**: 2026-01-31

## Overview

This feature adds automatic Claude Code hook installation to smoke. After implementation, running `smoke init` will set up both smoke AND hooks that nudge agents to post during natural pauses.

## Key Files to Create/Modify

### New Files

| File | Purpose |
|------|---------|
| `internal/hooks/hooks.go` | Core hook management (install, uninstall, status) |
| `internal/hooks/embed.go` | Embed directive and asset access |
| `internal/hooks/scripts/smoke-break.sh` | Stop event hook script |
| `internal/hooks/scripts/smoke-nudge.sh` | PostToolUse event hook script |
| `internal/hooks/hooks_test.go` | Unit tests |
| `internal/cli/hooks.go` | Cobra commands (install, uninstall, status) |
| `internal/cli/hooks_test.go` | CLI tests |

### Modified Files

| File | Change |
|------|--------|
| `internal/cli/init.go` | Add hook installation call |
| `internal/cli/init_test.go` | Test hook integration |

## Implementation Order

1. **Create hooks package** (`internal/hooks/`)
   - Add embedded scripts
   - Implement Install(), Uninstall(), GetStatus()
   - Add tests

2. **Create CLI commands** (`internal/cli/hooks.go`)
   - `smoke hooks install`
   - `smoke hooks uninstall`
   - `smoke hooks status`
   - Add tests

3. **Integrate with init** (`internal/cli/init.go`)
   - Call hooks.Install() after smoke setup
   - Handle errors gracefully (warn, don't fail)
   - Update output messages

4. **Integration tests**
   - Full init → hooks flow
   - Edge cases (permissions, existing hooks)

## Testing Strategy

### Unit Tests

```go
// hooks/hooks_test.go
func TestInstall_FreshSystem(t *testing.T)
func TestInstall_AlreadyInstalled(t *testing.T)
func TestInstall_ModifiedScripts(t *testing.T)
func TestInstall_Force(t *testing.T)
func TestUninstall(t *testing.T)
func TestGetStatus_AllStates(t *testing.T)
```

### Integration Tests

```bash
# Test fresh install
smoke init
# Verify: ~/.claude/hooks/smoke-*.sh exist
# Verify: ~/.claude/settings.json has hook entries

# Test status
smoke hooks status
# Verify: shows "installed"

# Test uninstall
smoke hooks uninstall
smoke hooks status
# Verify: shows "not installed"

# Test reinstall
smoke hooks install
# Verify: hooks restored
```

## Dependencies

- Go 1.16+ (for embed directive)
- No new external dependencies

## Success Verification

After implementation, this should work:

```bash
# Fresh system
smoke init
# Output includes: "Hooks installed"

# Verify
smoke hooks status
# Shows: "Status: installed"

# Check files exist
ls ~/.claude/hooks/smoke-*.sh
# Shows both scripts

# Check settings
jq '.hooks' ~/.claude/settings.json
# Shows Stop and PostToolUse entries
```
