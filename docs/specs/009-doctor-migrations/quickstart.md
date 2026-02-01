# Quickstart: Doctor Migrations

**Date**: 2026-02-01
**Feature**: 009-doctor-migrations

## Test Scenarios

### Scenario 1: Fresh Install (No Migration Needed)

```bash
# Setup: New smoke installation
smoke init

# Test: Doctor should show no pending migrations
smoke doctor

# Expected output includes:
# MIGRATIONS
#   ✓  Config Schema    up to date (version 1)
```

### Scenario 2: Upgrade from Pre-1.7.0 (Migration Needed)

```bash
# Setup: Simulate old config without pressure field
cat > ~/.config/smoke/config.yaml << 'EOF'
contexts:
  conversation:
    prompt: "Test prompt"
EOF

# Test: Doctor detects missing migration
smoke doctor

# Expected output includes:
# MIGRATIONS
#   ⚠  Config Schema    1 pending migration
#      └─ Run 'smoke doctor --fix' to apply: add_pressure_setting
```

### Scenario 3: Apply Migrations

```bash
# Setup: Old config as in Scenario 2

# Test: Apply migrations with --fix
smoke doctor --fix

# Expected output:
# Fixed: Config Schema (Applied 1 migration: add_pressure_setting)

# Verify: Config now has pressure field and schema version
cat ~/.config/smoke/config.yaml
# Should show:
# pressure: 2
# _schema_version: 1
```

### Scenario 4: Dry Run Preview

```bash
# Setup: Old config without migrations

# Test: Preview what would be fixed
smoke doctor --fix --dry-run

# Expected output:
# Would fix: Config Schema
#   - Apply migration: add_pressure_setting (add pressure field with default 2)
#
# 1 issue(s) would be fixed.
```

### Scenario 5: Already Up to Date

```bash
# Setup: Config with _schema_version = CurrentSchemaVersion

# Test: Running --fix when no migrations needed
smoke doctor --fix

# Expected output:
# smoke doctor v1.8.0
#
# MIGRATIONS
#   ✓  Config Schema    up to date (version 1)
#
# No problems to fix.
```

### Scenario 6: Preserve User Values

```bash
# Setup: Old config with custom values
cat > ~/.config/smoke/config.yaml << 'EOF'
contexts:
  custom:
    prompt: "My custom prompt"
    categories:
      - Observations
EOF

# Test: Migration preserves existing values
smoke doctor --fix

# Verify: Custom context still present
cat ~/.config/smoke/config.yaml
# Should include:
# contexts:
#   custom:
#     prompt: "My custom prompt"
# pressure: 2
# _schema_version: 1
```

### Scenario 7: Corrupt Config File

```bash
# Setup: Invalid YAML
echo "invalid: yaml: content:" > ~/.config/smoke/config.yaml

# Test: Doctor reports error, doesn't attempt migration
smoke doctor

# Expected: Fails at Config File check before migrations
# DATA
#   ✗  Config File    invalid YAML
```

### Scenario 8: Multiple Pending Migrations

```bash
# Setup: Very old config (future scenario with 3 migrations)
# _schema_version: 0 or missing

# Test: Doctor applies all in order
smoke doctor --fix

# Expected output:
# Fixed: Config Schema (Applied 3 migrations: add_pressure_setting, add_theme, add_something)
```

## CLI Commands

| Command | Description |
|---------|-------------|
| `smoke doctor` | Check all health including pending migrations |
| `smoke doctor --fix` | Apply all pending migrations |
| `smoke doctor --fix --dry-run` | Preview migrations without applying |

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | All checks pass, no migrations needed |
| 1 | Warnings present (e.g., pending migrations) |
| 2 | Errors present (e.g., corrupt config) |
