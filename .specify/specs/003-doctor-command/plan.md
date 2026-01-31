# Implementation Plan: Smoke Doctor Command

**Branch**: `003-doctor-command` | **Date**: 2026-01-31 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/docs/specs/003-doctor-command/spec.md`

## Summary

Add a `smoke doctor` command that displays health check status for the smoke installation (similar to `bd doctor`) and supports `--fix` to automatically repair common issues. The command follows existing CLI patterns using Cobra and integrates with existing config and feed packages.

## Technical Context

**Language/Version**: Go 1.22+
**Primary Dependencies**: Cobra (existing), standard library only
**Storage**: ~/.config/smoke/ (existing JSONL feed, YAML config)
**Testing**: Go testing + integration tests via compiled binary (existing pattern)
**Target Platform**: macOS, Linux (existing)
**Project Type**: Single CLI application
**Performance Goals**: Complete in under 1 second
**Constraints**: No external dependencies beyond Cobra

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Go Simplicity | ✅ PASS | Standard library only, no new dependencies |
| II. Agent-First CLI Design | ✅ PASS | Clear output, actionable errors, `--fix` for self-healing |
| III. Local-First Storage | ✅ PASS | Only checks local files in ~/.config/smoke/ |
| IV. Test What Matters | ✅ PASS | Will add integration tests for doctor command |
| V. Environment Integration | ✅ PASS | Works with existing config paths |
| VI. Minimal Configuration | ✅ PASS | No new config options, just diagnostic output |

**All gates pass. No violations to justify.**

## Project Structure

### Documentation (this feature)

```text
.specify/specs/003-doctor-command/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
└── tasks.md             # Phase 2 output (/speckit.tasks command)
```

### Source Code (repository root)

```text
internal/
├── cli/
│   └── doctor.go        # New: doctor command implementation
├── config/
│   └── (existing)       # Reuse GetConfigDir, GetFeedPath, etc.
└── feed/
    └── (existing)       # Reuse Store for validation

tests/
└── integration/
    └── smoke_test.go    # Add doctor command tests
```

**Structure Decision**: Single project structure. The doctor command fits naturally into `internal/cli/` alongside existing commands (feed.go, post.go, reply.go). It will reuse existing config and feed packages for path resolution and validation.

## Complexity Tracking

No constitution violations. No complexity justification needed.

---

## Phase 0: Research

### Research Tasks

1. **Existing config package analysis**: Understand current path resolution and initialization checks
2. **bd doctor output format**: Document exact visual format to replicate
3. **JSONL validation patterns**: How to validate feed file integrity

### Findings

#### 1. Existing Config Package

From `internal/config/root.go`:
- `GetConfigDir()` returns `~/.config/smoke/`
- `GetFeedPath()` returns feed file path (with SMOKE_FEED override support)
- `IsSmokeInitialized()` checks if feed file exists
- `EnsureInitialized()` returns error if not initialized

These functions can be reused for doctor checks.

#### 2. bd doctor Output Format

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

Pattern:
- Version header at top
- Category headers in CAPS
- 2-space indent for checks
- Status indicators: ✓ (pass), ⚠ (warning), ✗ (error)
- 5-space indent + └─ for additional details

#### 3. JSONL Validation

Feed validation should:
- Check file exists and is readable
- Attempt to parse each line as JSON
- Count valid vs invalid lines
- Report corruption but don't fail on minor issues

---

## Phase 1: Design

### Data Model

See [data-model.md](data-model.md)

### Key Types

```go
// CheckStatus represents the result of a health check
type CheckStatus int

const (
    StatusPass CheckStatus = iota
    StatusWarn
    StatusFail
)

// Check represents a single diagnostic check
type Check struct {
    Name    string
    Status  CheckStatus
    Message string
    Detail  string      // Optional additional info
    CanFix  bool        // Whether --fix can repair this
    Fix     func() error // Fix function if CanFix is true
}

// Category groups related checks
type Category struct {
    Name   string
    Checks []Check
}
```

### Check Categories

1. **INSTALLATION**
   - Config directory exists
   - Config directory writable
   - Feed file exists

2. **DATA**
   - Feed file readable
   - Feed file valid JSONL
   - Config file exists and valid

3. **VERSION**
   - Display current smoke version

### Fix Actions

| Check | Fixable | Fix Action |
|-------|---------|------------|
| Config dir missing | Yes | Create ~/.config/smoke/ |
| Feed file missing | Yes | Create empty feed.jsonl |
| Config file missing | Yes | Create default config.yaml |
| Feed invalid JSONL | No | Suggest manual inspection |
| Permission errors | No | Suggest chmod commands |

---

## Implementation Approach

1. Create `internal/cli/doctor.go` with Cobra command
2. Define check types and status constants
3. Implement individual check functions
4. Implement fix functions for fixable issues
5. Format output to match bd doctor style
6. Add integration tests
7. Update help text and documentation
