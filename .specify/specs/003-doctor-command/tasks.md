# Tasks: Smoke Doctor Command

**Input**: Design documents from `.specify/specs/003-doctor-command/`
**Prerequisites**: plan.md, spec.md, data-model.md, research.md, quickstart.md

**Tests**: Integration tests included (existing pattern in tests/integration/)

**Organization**: Tasks grouped by user story for independent implementation

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story (US1, US2, US3)
- Paths relative to repository root

---

## Phase 1: Setup

**Purpose**: Project structure for doctor command

- [ ] T001 Create internal/cli/doctor.go with empty doctorCmd Cobra command
- [ ] T002 Register doctorCmd in internal/cli/root.go

**Checkpoint**: `smoke doctor` command exists (shows help, does nothing)

---

## Phase 2: Foundational (Core Types)

**Purpose**: Define data types used by all user stories

**⚠️ CRITICAL**: These types are used by all checks and must be complete before user story work

- [ ] T003 Define CheckStatus type with StatusPass/StatusWarn/StatusFail constants in internal/cli/doctor.go
- [ ] T004 Define Check struct with Name, Status, Message, Detail, CanFix, Fix fields in internal/cli/doctor.go
- [ ] T005 Define Category struct with Name and Checks slice in internal/cli/doctor.go
- [ ] T006 Implement formatCheck() to format single check with ✓/⚠/✗ indicators in internal/cli/doctor.go
- [ ] T007 Implement formatCategory() to format category header and all checks in internal/cli/doctor.go
- [ ] T008 Implement printReport() to output version header and all categories in internal/cli/doctor.go

**Checkpoint**: Types defined, formatting functions ready for checks

---

## Phase 3: User Story 1 - Health Check Status (Priority: P1) 🎯 MVP

**Goal**: Display categorized health check report showing installation status

**Independent Test**: Run `smoke doctor` and verify all checks display with correct indicators

### Implementation for User Story 1

- [ ] T009 [P] [US1] Implement checkConfigDir() to verify ~/.config/smoke/ exists and is writable in internal/cli/doctor.go
- [ ] T010 [P] [US1] Implement checkFeedFile() to verify feed.jsonl exists and is readable in internal/cli/doctor.go
- [ ] T011 [P] [US1] Implement checkFeedFormat() to validate JSONL integrity (count valid/invalid lines) in internal/cli/doctor.go
- [ ] T012 [P] [US1] Implement checkConfigFile() to verify config.yaml exists and is valid YAML in internal/cli/doctor.go
- [ ] T013 [US1] Implement checkVersion() to display current smoke version in internal/cli/doctor.go
- [ ] T014 [US1] Implement runChecks() to collect all checks into INSTALLATION, DATA, VERSION categories in internal/cli/doctor.go
- [ ] T015 [US1] Update doctorCmd.Run to call runChecks() and printReport() in internal/cli/doctor.go
- [ ] T016 [US1] Implement exit codes: 0 for pass, 1 for warnings, 2 for errors in internal/cli/doctor.go
- [ ] T017 [US1] Add integration test for healthy installation in tests/integration/smoke_test.go
- [ ] T018 [US1] Add integration test for missing feed file in tests/integration/smoke_test.go

**Checkpoint**: `smoke doctor` displays full health check report with correct exit codes

---

## Phase 4: User Story 2 - Auto-Fix Problems (Priority: P2)

**Goal**: Automatically repair fixable issues with `--fix` flag

**Independent Test**: Delete feed file, run `smoke doctor --fix`, verify feed recreated

### Implementation for User Story 2

- [ ] T019 [US2] Add --fix flag to doctorCmd in internal/cli/doctor.go
- [ ] T020 [P] [US2] Implement fixConfigDir() to create ~/.config/smoke/ with 0755 permissions in internal/cli/doctor.go
- [ ] T021 [P] [US2] Implement fixFeedFile() to create empty feed.jsonl in internal/cli/doctor.go
- [ ] T022 [P] [US2] Implement fixConfigFile() to create default config.yaml in internal/cli/doctor.go
- [ ] T023 [US2] Wire Fix functions to Check.Fix field for fixable checks in internal/cli/doctor.go
- [ ] T024 [US2] Implement applyFixes() to execute Fix() for failed checks when --fix provided in internal/cli/doctor.go
- [ ] T025 [US2] Update printReport() to show "Fixed" status after successful fix in internal/cli/doctor.go
- [ ] T026 [US2] Display "No problems to fix" when all checks pass with --fix in internal/cli/doctor.go
- [ ] T027 [US2] Add integration test for --fix repairing missing config dir in tests/integration/smoke_test.go
- [ ] T028 [US2] Add integration test for --fix with no problems in tests/integration/smoke_test.go

**Checkpoint**: `smoke doctor --fix` repairs fixable issues automatically

---

## Phase 5: User Story 3 - Dry-Run Mode (Priority: P3)

**Goal**: Preview what --fix would do without making changes

**Independent Test**: Delete feed, run `smoke doctor --fix --dry-run`, verify feed still missing

### Implementation for User Story 3

- [ ] T029 [US3] Add --dry-run flag to doctorCmd in internal/cli/doctor.go
- [ ] T030 [US3] Update applyFixes() to print "Would fix:" instead of fixing when --dry-run in internal/cli/doctor.go
- [ ] T031 [US3] Ensure exit code reflects problems even in --dry-run mode in internal/cli/doctor.go
- [ ] T032 [US3] Add integration test for --fix --dry-run shows what would be fixed in tests/integration/smoke_test.go
- [ ] T033 [US3] Add integration test for --fix --dry-run does not modify files in tests/integration/smoke_test.go

**Checkpoint**: `smoke doctor --fix --dry-run` previews fixes without making changes

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Quality improvements and documentation

- [ ] T034 [P] Add actionable guidance messages for all failed checks (e.g., "Run 'smoke doctor --fix' to repair") in internal/cli/doctor.go
- [ ] T035 [P] Ensure output matches bd doctor visual style (spacing, indentation) in internal/cli/doctor.go
- [ ] T036 Verify all existing tests still pass with make test
- [ ] T037 Run make lint and fix any issues
- [ ] T038 Run quickstart.md validation scenarios manually

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup - BLOCKS all user stories
- **User Story 1 (Phase 3)**: Depends on Foundational
- **User Story 2 (Phase 4)**: Depends on Foundational (can parallel with US1)
- **User Story 3 (Phase 5)**: Depends on US2 (needs --fix to add --dry-run)
- **Polish (Phase 6)**: Depends on all user stories complete

### User Story Dependencies

- **US1 (P1)**: Foundation complete → can start
- **US2 (P2)**: Foundation complete → can start (parallel with US1 if desired)
- **US3 (P3)**: US2 complete (--dry-run extends --fix behavior)

### Within Each User Story

- Check functions before runChecks() integration
- Fix functions before applyFixes() integration
- Implementation before tests

### Parallel Opportunities

Within Phase 2 (Foundational):
- T003, T004, T005 can run in parallel (type definitions)
- T006, T007 can run in parallel after types defined

Within US1:
- T009, T010, T011, T012 can run in parallel (independent check functions)

Within US2:
- T020, T021, T022 can run in parallel (independent fix functions)

---

## Parallel Example: User Story 1 Checks

```bash
# Launch all check implementations together:
Task: "Implement checkConfigDir() in internal/cli/doctor.go"
Task: "Implement checkFeedFile() in internal/cli/doctor.go"
Task: "Implement checkFeedFormat() in internal/cli/doctor.go"
Task: "Implement checkConfigFile() in internal/cli/doctor.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001-T002)
2. Complete Phase 2: Foundational (T003-T008)
3. Complete Phase 3: User Story 1 (T009-T018)
4. **STOP and VALIDATE**: Run `smoke doctor` on healthy/unhealthy installations
5. MVP is ready - agents can diagnose issues

### Incremental Delivery

1. Setup + Foundational → Types and formatting ready
2. Add User Story 1 → `smoke doctor` works → MVP!
3. Add User Story 2 → `smoke doctor --fix` works
4. Add User Story 3 → `smoke doctor --fix --dry-run` works
5. Polish → Production ready

---

## Summary

- **Total tasks**: 38
- **Phase 1 (Setup)**: 2 tasks
- **Phase 2 (Foundational)**: 6 tasks
- **Phase 3 (US1 - Health Check)**: 10 tasks (MVP)
- **Phase 4 (US2 - Auto-Fix)**: 10 tasks
- **Phase 5 (US3 - Dry-Run)**: 5 tasks
- **Phase 6 (Polish)**: 5 tasks
- **Parallel opportunities**: 12 tasks marked [P]
- **Suggested MVP scope**: User Story 1 (18 tasks through Phase 3)
