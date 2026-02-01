# Research: Doctor Migrations

**Date**: 2026-02-01
**Feature**: 009-doctor-migrations

## Research Questions

### Q1: How should migrations be tracked?

**Decision**: Schema version number in config.yaml

**Rationale**:
- Simple integer comparison (current vs required)
- Single field to maintain (`_schema_version`)
- Easy to debug: `cat config.yaml | grep _schema_version`
- Matches common patterns (Rails, Django, Alembic)

**Alternatives Considered**:
1. **Applied migrations list** (`_applied_migrations: [001, 002]`)
   - More granular but adds complexity
   - Requires set operations to find pending
   - Overkill for smoke's simple config
2. **No tracking, check each field**
   - Every migration checks if its field exists
   - Works but slower with many migrations
   - Can't distinguish "never had" from "manually removed"

### Q2: Where should migrations be defined?

**Decision**: Go code in `internal/config/migrations.go`

**Rationale**:
- Type-safe, compile-time checked
- No external files to distribute
- Easy to test
- Migrations are code, not data

**Alternatives Considered**:
1. **YAML/JSON migration files**
   - More flexible but requires parser
   - Can't express complex logic
   - Distribution complexity
2. **SQL-style migration files**
   - Overkill for YAML config changes
   - Adds file management complexity

### Q3: How to preserve existing config values during migration?

**Decision**: Read entire config, modify in-memory, write back

**Rationale**:
- YAML preserves unknown fields by default with `yaml.v3`
- Single atomic write operation
- Creates backup before modifying (existing pattern in doctor.go)

**Implementation**:
```go
// Read existing config
data, _ := os.ReadFile(configPath)
var config map[string]interface{}
yaml.Unmarshal(data, &config)

// Apply migration (adds new field, preserves existing)
if _, exists := config["pressure"]; !exists {
    config["pressure"] = DefaultPressure
}

// Write back (atomic via temp file + rename)
newData, _ := yaml.Marshal(config)
os.WriteFile(configPath, newData, 0600)
```

### Q4: How to integrate with existing doctor checks?

**Decision**: Add "MIGRATIONS" category to `runChecks()` output

**Rationale**:
- Reuses existing `--fix` infrastructure
- Consistent UX with other doctor checks
- No changes to check/fix flow

**Implementation**:
```go
// In runChecks()
{
    Name: "MIGRATIONS",
    Checks: []Check{
        performMigrationCheck(),
    },
},
```

### Q5: What should the first migration be?

**Decision**: Migration 001: Ensure `pressure` field exists

**Rationale**:
- `pressure` was added in v1.7.0
- Users upgrading from older versions won't have it
- Good test case for migration system

**Migration Definition**:
```go
{
    Version: 1,
    Name:    "add_pressure_setting",
    Description: "Add pressure field for nudge frequency control",
    Check: func(cfg map[string]interface{}) bool {
        _, exists := cfg["pressure"]
        return !exists // needs migration if missing
    },
    Apply: func(cfg map[string]interface{}) {
        cfg["pressure"] = DefaultPressure // 2
    },
},
```

## Existing Code Analysis

### Doctor Infrastructure

From `internal/cli/doctor.go`:
- `Check` struct with `CanFix` and `Fix` function
- `applyFixes()` iterates checks and calls Fix functions
- `--dry-run` support already exists
- Backup pattern exists (`backupTUIConfig`)

### Config Package

From `internal/config/`:
- `GetConfigPath()` returns `~/.config/smoke/config.yaml`
- `suggest.go` already loads/saves config.yaml
- YAML parsing with `gopkg.in/yaml.v3`

## Conclusions

The migration system will:
1. Store schema version in config.yaml as `_schema_version`
2. Define migrations in Go code as `[]Migration` slice
3. Integrate as new "MIGRATIONS" category in doctor output
4. Reuse existing `--fix` and `--dry-run` flags
5. Create backup before modifying config (existing pattern)
6. First migration adds `pressure` field for users upgrading from pre-1.7.0
