# Feature Specification: Doctor Migrations

**Feature Branch**: `009-doctor-migrations`
**Created**: 2026-02-01
**Status**: Draft
**Input**: User description: "Config migration system for smoke doctor to detect and apply version upgrades"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Detect Pending Migrations (Priority: P1)

When a user upgrades smoke to a new version, `smoke doctor` should detect if their configuration is missing required settings or structures that were introduced in newer versions.

**Why this priority**: Without detection, users won't know their config needs updating, leading to confusing errors or missing features.

**Independent Test**: Run `smoke doctor` after upgrading smoke binary. System reports which migrations are pending without applying any changes.

**Acceptance Scenarios**:

1. **Given** a fresh smoke installation, **When** user runs `smoke doctor`, **Then** system reports "All checks passed" with no pending migrations
2. **Given** an older config (e.g., missing `pressure` field), **When** user upgrades smoke and runs `smoke doctor`, **Then** system reports "1 pending migration: add pressure setting"
3. **Given** config at current version, **When** user runs `smoke doctor`, **Then** system reports no pending migrations

---

### User Story 2 - Apply Migrations Automatically (Priority: P1)

Users should be able to run `smoke doctor --fix` to automatically apply all pending migrations, bringing their configuration up to date with the current smoke version.

**Why this priority**: This is the core value - users need a way to upgrade their config without manual editing. Paired with detection, this completes the upgrade story.

**Independent Test**: Run `smoke doctor --fix` on outdated config. Verify config file is updated and subsequent `smoke doctor` shows no pending migrations.

**Acceptance Scenarios**:

1. **Given** config missing `pressure` field, **When** user runs `smoke doctor --fix`, **Then** pressure field is added with default value (2)
2. **Given** multiple pending migrations, **When** user runs `smoke doctor --fix`, **Then** all migrations are applied in order
3. **Given** no pending migrations, **When** user runs `smoke doctor --fix`, **Then** system reports "Nothing to fix" and makes no changes
4. **Given** config file with user customizations, **When** migrations are applied, **Then** existing user settings are preserved

---

### User Story 3 - Track Applied Migrations (Priority: P2)

The system must track which migrations have been applied to avoid re-running them and to provide visibility into config history.

**Why this priority**: Important for reliability (don't re-apply migrations) but detection logic can work by checking config state directly for MVP.

**Independent Test**: Apply a migration, verify it's recorded, run doctor again to confirm it's not flagged as pending.

**Acceptance Scenarios**:

1. **Given** a migration was applied, **When** user runs `smoke doctor` again, **Then** that migration is not listed as pending
2. **Given** migration tracking exists, **When** user views doctor output, **Then** they can see which migrations have been applied and when

---

### User Story 4 - Dry Run Mode (Priority: P3)

Users should be able to preview what migrations would be applied without actually changing their config.

**Why this priority**: Nice-to-have for cautious users who want to review changes before applying. Most users will just run `--fix`.

**Independent Test**: Run `smoke doctor --fix --dry-run` and verify no changes are made to config file.

**Acceptance Scenarios**:

1. **Given** pending migrations, **When** user runs `smoke doctor --fix --dry-run`, **Then** system shows what would be changed without modifying files
2. **Given** pending migrations, **When** user runs dry-run followed by actual fix, **Then** changes match what was previewed

---

### Edge Cases

- What happens when config file is malformed/corrupt? System should report error and not attempt migrations.
- What happens when a migration fails mid-way? System should report the failure clearly and not leave config in inconsistent state.
- What happens when user has read-only config file? System should report permission error gracefully.
- What happens when multiple smoke versions are installed? Migration state is per-config-file, not global.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST detect when configuration is missing fields or structures required by the current version
- **FR-002**: System MUST apply migrations in a defined order (oldest to newest)
- **FR-003**: System MUST preserve existing user configuration values when adding new fields
- **FR-004**: System MUST use sensible defaults for new required fields (matching what `smoke init` would create)
- **FR-005**: System MUST track which migrations have been applied to prevent re-application
- **FR-006**: System MUST report migration status as part of `smoke doctor` health check
- **FR-007**: System MUST apply all pending migrations when `--fix` flag is provided
- **FR-008**: System MUST NOT modify config files during detection (read-only until `--fix`)
- **FR-009**: System MUST handle missing config directory gracefully (suggest running `smoke init`)
- **FR-010**: System MUST validate config integrity after applying migrations

### Key Entities

- **Migration**: A named, versioned transformation that updates config from one state to another. Contains: ID (sequential number), name (descriptive), apply function, detection function.
- **MigrationState**: Record of which migrations have been applied. Contains: migration ID, applied timestamp, smoke version at time of application.
- **ConfigVersion**: A marker in the config indicating its schema version, used for quick version comparison.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can upgrade smoke and run `smoke doctor --fix` in under 10 seconds to update their config
- **SC-002**: Zero data loss - all existing user configuration values are preserved during migration
- **SC-003**: `smoke doctor` clearly indicates when config needs updating (pending migrations count)
- **SC-004**: Migrations are idempotent - running `smoke doctor --fix` multiple times produces same result
- **SC-005**: 100% of new configuration fields introduced in releases have corresponding migrations

## Assumptions

- Migrations will be defined in code, not as external files
- Migration order is determined by sequential numbering (001, 002, 003...)
- Config format remains YAML-based
- Migration state can be stored within the config file itself (e.g., `_schema_version` or `_migrations` field)
- Backwards compatibility: old smoke versions reading new config should ignore unknown fields gracefully
