# Data Model: Doctor Migrations

**Date**: 2026-02-01
**Feature**: 009-doctor-migrations

## Entities

### Migration

Represents a single configuration transformation.

| Field | Type | Description |
|-------|------|-------------|
| Version | int | Sequential migration number (1, 2, 3...) |
| Name | string | Short identifier (e.g., "add_pressure_setting") |
| Description | string | Human-readable explanation |
| NeedsMigration | func | Returns true if migration should be applied |
| Apply | func | Applies the migration to config map |

**Notes**:
- Migrations are defined in code, not stored
- Order determined by Version field
- Each migration is idempotent (safe to run multiple times)

### Config Schema Version

Stored in `~/.config/smoke/config.yaml`:

```yaml
# User settings
pressure: 2
contexts:
  conversation:
    prompt: "..."

# Schema tracking (managed by migrations)
_schema_version: 1
```

| Field | Type | Description |
|-------|------|-------------|
| _schema_version | int | Current schema version applied to this config |

**Notes**:
- Prefixed with `_` to indicate internal/managed field
- Missing field treated as version 0 (needs all migrations)
- Only updated after successful migration application

## State Transitions

### Migration States

```
Config State                    Migration Action
─────────────────────────────────────────────────
_schema_version missing    →    Apply all migrations, set to CurrentSchemaVersion
_schema_version < Current  →    Apply migrations from (version+1) to Current
_schema_version = Current  →    No action needed
_schema_version > Current  →    Warning (config from future version)
```

### Config File Lifecycle

```
1. Fresh install (smoke init)
   └─> Creates config.yaml with _schema_version: CurrentSchemaVersion

2. Existing install, upgrade smoke binary
   └─> smoke doctor detects _schema_version < CurrentSchemaVersion
   └─> Reports pending migrations
   └─> smoke doctor --fix applies migrations
   └─> Updates _schema_version to CurrentSchemaVersion

3. User manually edits config
   └─> _schema_version unchanged
   └─> User values preserved on next migration
```

## Migration Registry

Initial migrations for v1.8.0:

| Version | Name | Description |
|---------|------|-------------|
| 1 | add_pressure_setting | Adds `pressure: 2` for nudge frequency control |

**Future migrations** (examples):
- Version 2: Add `theme` field for TUI theming
- Version 3: Migrate deprecated field names

## Validation Rules

1. **Version must be sequential**: No gaps (1, 2, 3 not 1, 3, 5)
2. **Migrations must be idempotent**: Running twice produces same result
3. **Migrations must preserve existing values**: Only add/modify intended fields
4. **Config must remain valid YAML**: After any migration
