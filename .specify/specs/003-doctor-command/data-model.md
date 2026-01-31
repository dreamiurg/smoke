# Data Model: Smoke Doctor Command

**Date**: 2026-01-31
**Spec**: [spec.md](spec.md)
**Plan**: [plan.md](plan.md)

## Entities

### CheckStatus

Represents the result state of a health check.

```go
type CheckStatus int

const (
    StatusPass CheckStatus = iota  // ✓ - Everything OK
    StatusWarn                      // ⚠ - Issue exists but not blocking
    StatusFail                      // ✗ - Critical issue
)
```

**Validation rules**:
- Only these three states are valid
- StatusFail indicates the command should return exit code 2
- StatusWarn indicates exit code 1

### Check

Represents a single diagnostic check with its result.

| Field | Type | Description |
|-------|------|-------------|
| Name | string | Short identifier (e.g., "Config Directory") |
| Status | CheckStatus | Pass/Warn/Fail |
| Message | string | Status description (e.g., "Found at ~/.config/smoke/") |
| Detail | string | Optional additional info, shown on next line |
| CanFix | bool | Whether --fix can repair this issue |
| Fix | func() error | Fix function, nil if not fixable |

**Validation rules**:
- Name is required, should be human-readable
- Message is required
- Detail is optional, only shown when non-empty
- CanFix should be true only if Fix is non-nil

### Category

Groups related checks for organized output.

| Field | Type | Description |
|-------|------|-------------|
| Name | string | Category header (e.g., "INSTALLATION") |
| Checks | []Check | Ordered list of checks in this category |

**Validation rules**:
- Name should be ALL CAPS for display
- Checks must have at least one item

## Check Categories

### INSTALLATION
Checks related to smoke setup and file system state.

| Check | Conditions | Status |
|-------|------------|--------|
| Config Directory | Exists and writable | Pass |
| Config Directory | Exists but not writable | Warn |
| Config Directory | Does not exist | Fail (fixable) |
| Feed File | Exists and readable | Pass |
| Feed File | Does not exist | Fail (fixable) |

### DATA
Checks related to data integrity.

| Check | Conditions | Status |
|-------|------------|--------|
| Feed Format | All lines valid JSONL | Pass |
| Feed Format | Some lines invalid | Warn |
| Feed Format | File unreadable | Fail |
| Config File | Exists and valid YAML | Pass |
| Config File | Missing | Warn (fixable) |
| Config File | Invalid YAML | Fail |

### VERSION
Version information (informational only).

| Check | Conditions | Status |
|-------|------------|--------|
| Smoke Version | Always shows current | Pass |

## State Transitions

```
Initial State
     │
     ▼
┌─────────────────┐
│ Run all checks  │
└────────┬────────┘
         │
         ▼
┌─────────────────┐     ┌─────────────────┐
│ All pass?       │─Yes─▶│ Exit code 0    │
└────────┬────────┘     └─────────────────┘
         │No
         ▼
┌─────────────────┐     ┌─────────────────┐
│ Any fail?       │─Yes─▶│ Exit code 2    │
└────────┬────────┘     └─────────────────┘
         │No (only warnings)
         ▼
┌─────────────────┐
│ Exit code 1     │
└─────────────────┘
```

With `--fix` flag:
```
┌─────────────────┐
│ Check fails     │
│ CanFix = true   │
└────────┬────────┘
         │
         ▼
┌─────────────────┐     ┌─────────────────┐
│ --dry-run?      │─Yes─▶│ Print would fix │
└────────┬────────┘     └─────────────────┘
         │No
         ▼
┌─────────────────┐
│ Execute Fix()   │
│ Re-run check    │
└─────────────────┘
```

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | All checks pass |
| 1 | Warnings only (smoke works but has issues) |
| 2 | Errors present (smoke may not function correctly) |
