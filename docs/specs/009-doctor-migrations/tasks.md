# Tasks: Doctor Migrations

**Input**: Design documents from `/docs/specs/009-doctor-migrations/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md

**Tests**: Tests included - this feature requires tests for migration reliability.

**Organization**: Tasks grouped by user story. US1+US2 are both P1 and tightly coupled, so grouped together.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3, US4)
- Exact file paths included in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Create migration infrastructure in config package

- [ ] T001 Define Migration struct and CurrentSchemaVersion constant in internal/config/migrations.go
- [ ] T002 Implement GetConfigAsMap() to read config.yaml as map[string]interface{} in internal/config/migrations.go
- [ ] T003 Implement WriteConfigMap() with atomic write (temp file + rename) in internal/config/migrations.go

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core migration detection logic that all user stories depend on

**CRITICAL**: No user story work can begin until this phase is complete

- [ ] T004 Implement GetSchemaVersion() to read _schema_version from config in internal/config/migrations.go
- [ ] T005 Implement GetPendingMigrations() returning []Migration slice in internal/config/migrations.go
- [ ] T006 [P] Create migrations_test.go with test helpers (temp config file setup) in internal/config/migrations_test.go

**Checkpoint**: Foundation ready - migration detection infrastructure complete

---

## Phase 3: User Story 1+2 - Detect & Apply Migrations (Priority: P1) MVP

**Goal**: `smoke doctor` detects pending migrations and `--fix` applies them

**Independent Test**: Run `smoke doctor` on config missing `pressure` field. Should report pending. Run `--fix`. Should add field and update _schema_version.

### Tests for US1+US2

- [ ] T007 [P] [US1] Unit test: GetPendingMigrations returns empty when up to date in internal/config/migrations_test.go
- [ ] T008 [P] [US1] Unit test: GetPendingMigrations returns migrations when _schema_version missing in internal/config/migrations_test.go
- [ ] T009 [P] [US2] Unit test: ApplyMigrations adds missing fields and updates schema version in internal/config/migrations_test.go
- [ ] T010 [P] [US2] Unit test: ApplyMigrations preserves existing user config values in internal/config/migrations_test.go

### Implementation for US1+US2

- [ ] T011 [US1] Define first migration (add_pressure_setting) in migrations registry in internal/config/migrations.go
- [ ] T012 [US1] Add performMigrationCheck() function in internal/cli/doctor.go
- [ ] T013 [US1] Add MIGRATIONS category to runChecks() in internal/cli/doctor.go
- [ ] T014 [US2] Implement ApplyMigrations() with backup creation in internal/config/migrations.go
- [ ] T015 [US2] Implement fixMigrations() function that calls ApplyMigrations in internal/cli/doctor.go
- [ ] T016 [US2] Wire fix function to migration check (CanFix: true) in internal/cli/doctor.go

**Checkpoint**: MVP complete - `smoke doctor` and `smoke doctor --fix` work for migrations

---

## Phase 4: User Story 3 - Track Applied Migrations (Priority: P2)

**Goal**: Migration tracking via _schema_version prevents re-application

**Independent Test**: Apply migration, verify _schema_version updated, run doctor again - no pending migrations

### Tests for US3

- [ ] T017 [P] [US3] Unit test: SetSchemaVersion updates _schema_version in config in internal/config/migrations_test.go
- [ ] T018 [P] [US3] Unit test: After ApplyMigrations, GetPendingMigrations returns empty in internal/config/migrations_test.go

### Implementation for US3

- [ ] T019 [US3] Implement SetSchemaVersion() to update config file in internal/config/migrations.go
- [ ] T020 [US3] Update ApplyMigrations() to call SetSchemaVersion after success in internal/config/migrations.go
- [ ] T021 [US3] Update performMigrationCheck() to show "up to date (version N)" when no pending in internal/cli/doctor.go

**Checkpoint**: Schema version tracking complete

---

## Phase 5: User Story 4 - Dry Run Mode (Priority: P3)

**Goal**: `smoke doctor --fix --dry-run` previews migrations without applying

**Independent Test**: Run `--fix --dry-run` with pending migrations - shows what would change, config file unchanged

### Tests for US4

- [ ] T022 [P] [US4] Unit test: ApplyMigrations with dryRun=true returns results but doesn't modify config in internal/config/migrations_test.go

### Implementation for US4

- [ ] T023 [US4] Add dryRun parameter to ApplyMigrations() in internal/config/migrations.go
- [ ] T024 [US4] Update fixMigrations() to pass doctorDryRun flag to ApplyMigrations in internal/cli/doctor.go
- [ ] T025 [US4] Format dry-run output to show "Would apply: migration_name" in internal/cli/doctor.go

**Checkpoint**: Dry run mode complete

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Integration tests, edge cases, documentation

- [ ] T026 [P] Integration test: smoke doctor with outdated config in tests/integration/doctor_test.go
- [ ] T027 [P] Integration test: smoke doctor --fix applies migrations in tests/integration/doctor_test.go
- [ ] T028 Handle edge case: missing config directory (suggest smoke init) in internal/cli/doctor.go
- [ ] T029 Handle edge case: corrupt/invalid config YAML (report error, skip migrations) in internal/config/migrations.go
- [ ] T030 Handle edge case: _schema_version > CurrentSchemaVersion (warn about future config) in internal/config/migrations.go
- [ ] T031 Update smoke init to set _schema_version: CurrentSchemaVersion in internal/cli/init.go
- [ ] T032 Run quickstart.md validation scenarios manually

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **US1+US2 (Phase 3)**: Depends on Foundational - MVP delivery
- **US3 (Phase 4)**: Can start after US1+US2 (tracks what was applied)
- **US4 (Phase 5)**: Can start after US1+US2 (adds dry-run to existing apply)
- **Polish (Phase 6)**: Depends on all user stories being complete

### User Story Dependencies

- **US1+US2 (P1)**: Combined because apply depends on detect. Core MVP.
- **US3 (P2)**: Independent - enhances reliability but US1+US2 works without it
- **US4 (P3)**: Independent - adds preview capability to existing --fix

### Within Each Phase

- Tests written first (T007-T010 before T011-T016)
- Models/structs before functions that use them
- Internal functions before CLI integration

### Parallel Opportunities

**Phase 2** (after T004, T005 complete):
- T006 can run in parallel

**Phase 3** (tests can all run in parallel):
- T007, T008, T009, T010 all parallel

**Phase 4** (tests can run in parallel):
- T017, T018 parallel

**Phase 6** (integration tests parallel):
- T026, T027 parallel

---

## Parallel Example: Phase 3 Tests

```bash
# Launch all US1+US2 tests together:
Task: "Unit test: GetPendingMigrations returns empty when up to date"
Task: "Unit test: GetPendingMigrations returns migrations when missing"
Task: "Unit test: ApplyMigrations adds missing fields"
Task: "Unit test: ApplyMigrations preserves existing config values"
```

---

## Implementation Strategy

### MVP First (US1+US2 Only)

1. Complete Phase 1: Setup (T001-T003)
2. Complete Phase 2: Foundational (T004-T006)
3. Complete Phase 3: US1+US2 (T007-T016)
4. **STOP and VALIDATE**: Test with actual outdated config
5. Deploy v1.8.0 with migration support

### Incremental Delivery

1. Setup + Foundational → Infrastructure ready
2. Add US1+US2 → Core migration detection and apply (MVP)
3. Add US3 → Schema version tracking (reliability)
4. Add US4 → Dry run preview (cautious users)
5. Polish → Integration tests, edge cases

### Suggested MVP Scope

**MVP = Phases 1-3 only (T001-T016)**

This delivers:
- Migration detection in `smoke doctor`
- Migration apply via `--fix`
- First migration: add_pressure_setting

US3 and US4 can be added in follow-up PRs.
