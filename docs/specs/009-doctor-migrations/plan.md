# Implementation Plan: Doctor Migrations

**Branch**: `009-doctor-migrations` | **Date**: 2026-02-01 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/docs/specs/009-doctor-migrations/spec.md`

## Summary

Add a migration system to `smoke doctor` that detects when configuration needs updating after smoke version upgrades, and applies migrations via `--fix`. Migrations are defined in code, ordered sequentially, and tracked via a schema version in config.yaml.

## Technical Context

**Language/Version**: Go 1.22+
**Primary Dependencies**: gopkg.in/yaml.v3 (already in use), standard library
**Storage**: YAML at `~/.config/smoke/config.yaml`
**Testing**: Go testing + integration tests via compiled binary
**Target Platform**: macOS, Linux (cross-platform CLI)
**Project Type**: Single project (existing smoke CLI)
**Performance Goals**: Migrations complete in <1 second
**Constraints**: Must preserve existing user config values, idempotent operations

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Go Simplicity | ✅ Pass | Standard library + yaml.v3 only |
| II. Agent-First CLI | ✅ Pass | `--fix` flag already exists, adding migration detection |
| III. Local-First Storage | ✅ Pass | All in ~/.config/smoke/, no network |
| IV. Test What Matters | ✅ Pass | Will test CLI behavior + edge cases |
| V. Environment Integration | ✅ Pass | Extends existing config system |
| VI. Minimal Configuration | ✅ Pass | Auto-detection, sensible defaults |
| VII. Social Feed Tone | N/A | Not feed-related |
| VIII. Agent Workflow | ✅ Pass | Doctor is agent-friendly already |

**Gate Status**: PASS - No violations.

## Project Structure

### Documentation (this feature)

```text
docs/specs/009-doctor-migrations/
├── plan.md              # This file
├── spec.md              # Feature specification
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
└── checklists/          # Quality validation
    └── requirements.md
```

### Source Code (repository root)

```text
internal/
├── cli/
│   ├── doctor.go        # MODIFY: Add migration checks to runChecks()
│   └── doctor_test.go   # MODIFY: Add migration test cases
├── config/
│   ├── migrations.go    # NEW: Migration definitions and runner
│   └── migrations_test.go # NEW: Migration unit tests
└── ...

tests/integration/
└── doctor_test.go       # MODIFY: Add migration integration tests
```

**Structure Decision**: Extend existing `internal/config` and `internal/cli/doctor.go` packages. Migrations live in config package since they operate on config data. No new packages needed.

## Design Decisions

### Migration Detection Strategy

**Option A: Schema version number** (Selected)
- Store `_schema_version: N` in config.yaml
- Compare against `CurrentSchemaVersion` constant in code
- Simple, explicit, easy to debug

**Option B: Feature flags per migration**
- Store `_applied_migrations: [001, 002, 003]` as list
- More granular but more complex
- Rejected: Overkill for simple config changes

### Migration Definition Approach

Migrations defined as Go functions in a slice, ordered by version number:

```go
var migrations = []Migration{
    {Version: 1, Name: "add_pressure", Check: hasPressure, Apply: addPressure},
    {Version: 2, Name: "future_field", Check: hasFuture, Apply: addFuture},
}
```

Each migration has:
- `Check()`: Returns true if migration is needed (field missing)
- `Apply()`: Adds the field with default value

### Integration with Existing Doctor

Add new check category "MIGRATIONS" to `runChecks()`:
- Check function calls `GetPendingMigrations()` from config package
- If pending > 0, return check with `CanFix: true`
- Fix function calls `ApplyMigrations()` from config package

This reuses existing `--fix` infrastructure without modification.

## Complexity Tracking

No constitution violations to justify.
