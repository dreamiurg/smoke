# Research: Smoke Doctor Command

**Date**: 2026-01-31
**Spec**: [spec.md](spec.md)
**Plan**: [plan.md](plan.md)

## Research Tasks

### 1. Existing Config Package Analysis

**Decision**: Reuse existing config package functions for path resolution

**Findings from `internal/config/root.go`**:
- `GetConfigDir()` returns `~/.config/smoke/`
- `GetFeedPath()` returns feed file path (supports `SMOKE_FEED` env override)
- `IsSmokeInitialized()` checks if feed file exists
- `EnsureInitialized()` returns error if not initialized

**Rationale**: These functions already encapsulate path logic and environment variable handling. No need to duplicate.

**Alternatives considered**: Creating new path functions in doctor package - rejected as unnecessary duplication.

### 2. bd doctor Output Format

**Decision**: Match bd doctor visual style exactly

**Findings from bd doctor output**:
```
bd doctor v0.49.1

CORE SYSTEM
  ✓  Installation .beads/ directory found
  ⚠  CLI Version 0.49.1 (latest: 0.49.2)
  ✗  Some Check Error message here
     └─ Additional detail on new line

DATA & CONFIG
  ✓  Config File Valid configuration
```

**Pattern analysis**:
- Version header at top: `smoke doctor vX.Y.Z`
- Category headers in ALL CAPS
- 2-space indent for checks
- Status indicators: `✓` (pass), `⚠` (warning), `✗` (error)
- Check name followed by message on same line
- 5-space indent + `└─` for additional details

**Rationale**: Familiar format for users of bd. Consistent UX across tools.

**Alternatives considered**:
- JSON output - rejected for P1, could add `--json` flag later
- Table format - less readable for quick scanning

### 3. JSONL Validation Patterns

**Decision**: Line-by-line validation with error counting

**Approach**:
1. Open file and iterate line by line
2. Attempt to unmarshal each line as JSON
3. Count valid vs invalid lines
4. Report summary (e.g., "47/50 lines valid")

**Rationale**:
- Graceful degradation - don't fail on first error
- Provide actionable info about corruption extent
- Matches agent-first design (give useful info, not just errors)

**Alternatives considered**:
- Strict validation (fail on first error) - rejected, too fragile
- Full schema validation - rejected, over-engineering for P1

### 4. Fix Actions Research

**Decision**: Support three fix actions for P1

| Issue | Fix Action | Implementation |
|-------|------------|----------------|
| Config dir missing | Create directory | `os.MkdirAll(configDir, 0755)` |
| Feed file missing | Create empty file | `os.Create(feedPath)` + close |
| Config file missing | Create default | Write minimal YAML config |

**Rationale**: These are the most common issues agents encounter. All are safe, idempotent operations.

**Non-fixable issues (require manual intervention)**:
- Permission errors → Show chmod commands
- Corrupted JSONL → Suggest backup and manual inspection
- Disk full → Inform user

## Open Questions Resolved

All NEEDS CLARIFICATION items from plan resolved:
- ✓ Config package provides all needed path functions
- ✓ bd doctor format documented
- ✓ JSONL validation approach defined
- ✓ Fix actions scoped for P1
