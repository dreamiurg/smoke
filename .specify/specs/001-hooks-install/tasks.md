# Tasks: Hooks Installation System

**Input**: Design documents from `.specify/specs/001-hooks-install/` and `docs/specs/001-hooks-install/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: Tests are included as this is a Go project with established testing patterns and the spec requires comprehensive unit and integration tests.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3, US4)
- Include exact file paths in descriptions

## Path Conventions

- **Go project**: `internal/` for packages, `cmd/` for entry point
- Hook scripts: `internal/hooks/scripts/`
- CLI commands: `internal/cli/`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Create package structure and embed hook scripts

- [ ] T001 Create internal/hooks/ package directory structure
- [ ] T002 [P] Create internal/hooks/scripts/smoke-break.sh (Stop event hook script)
- [ ] T003 [P] Create internal/hooks/scripts/smoke-nudge.sh (PostToolUse event hook script)
- [ ] T004 Create internal/hooks/embed.go with embed directive for scripts/*.sh

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core hooks package infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [ ] T005 Create internal/hooks/types.go with HookEvent, ScriptStatus, InstallState, Status types
- [ ] T006 Create internal/hooks/paths.go with GetHooksDir(), GetSettingsPath() helper functions
- [ ] T007 Create internal/hooks/scripts.go with GetScriptContent(), ListScripts(), script hash comparison
- [ ] T008 Create internal/hooks/settings.go with settings.json read/write, merge logic, and smoke hook detection
- [ ] T009 Create internal/hooks/errors.go with ErrScriptsModified, ErrPermissionDenied, ErrInvalidSettings

**Checkpoint**: Foundation ready - user story implementation can now begin

---

## Phase 3: User Story 1 - Initialize Smoke with Hooks (Priority: P1) 🎯 MVP

**Goal**: `smoke init` automatically installs Claude Code hooks so agents receive nudges during natural pauses

**Independent Test**: Run `smoke init` on fresh system, verify smoke config AND hooks are in place

### Tests for User Story 1

- [ ] T010 [P] [US1] Unit tests for Install() in internal/hooks/hooks_test.go - fresh install scenario
- [ ] T011 [P] [US1] Unit tests for Install() in internal/hooks/hooks_test.go - already initialized scenario
- [ ] T012 [P] [US1] Unit tests for init hook integration in internal/cli/init_test.go

### Implementation for User Story 1

- [ ] T013 [US1] Implement Install(opts InstallOptions) in internal/hooks/hooks.go
- [ ] T014 [US1] Integrate hooks.Install() call in internal/cli/init.go after smoke setup
- [ ] T015 [US1] Handle hook errors gracefully in init (warn but don't fail per FR-002)
- [ ] T016 [US1] Update init output to include hook installation status
- [ ] T017 [US1] Handle "already initialized but hooks missing" case - suggest `smoke hooks install`

**Checkpoint**: `smoke init` installs hooks automatically; init succeeds even if hooks fail

---

## Phase 4: User Story 2 - Reinstall/Repair Hooks (Priority: P2)

**Goal**: Users can run `smoke hooks install` to restore/repair hooks after corruption or upgrade

**Independent Test**: Remove hook files, run install, verify hooks are restored

### Tests for User Story 2

- [ ] T018 [P] [US2] Unit tests for Install() in internal/hooks/hooks_test.go - repair scenario
- [ ] T019 [P] [US2] Unit tests for Install() in internal/hooks/hooks_test.go - modified scripts scenario
- [ ] T020 [P] [US2] Unit tests for Install() in internal/hooks/hooks_test.go - force flag scenario
- [ ] T021 [P] [US2] CLI tests for hooks install command in internal/cli/hooks_test.go

### Implementation for User Story 2

- [ ] T022 [US2] Create internal/cli/hooks.go with parent `hooks` command
- [ ] T023 [US2] Implement `smoke hooks install [--force]` subcommand in internal/cli/hooks.go
- [ ] T024 [US2] Implement idempotent install (detect up-to-date, report correctly)
- [ ] T025 [US2] Implement modified script detection with hash comparison
- [ ] T026 [US2] Implement --force flag to overwrite modified scripts
- [ ] T027 [US2] Add clear output messages per cli-interface.md contract

**Checkpoint**: `smoke hooks install` repairs/reinstalls hooks; --force overwrites modifications

---

## Phase 5: User Story 3 - Check Hook Installation Status (Priority: P2)

**Goal**: Users can verify hook installation state with `smoke hooks status`

**Independent Test**: Run status in various states (not installed, installed, partial) and verify accurate reporting

### Tests for User Story 3

- [ ] T028 [P] [US3] Unit tests for GetStatus() in internal/hooks/hooks_test.go - all states
- [ ] T029 [P] [US3] CLI tests for hooks status command in internal/cli/hooks_test.go

### Implementation for User Story 3

- [ ] T030 [US3] Implement GetStatus() in internal/hooks/hooks.go
- [ ] T031 [US3] Implement `smoke hooks status [--json]` subcommand in internal/cli/hooks.go
- [ ] T032 [US3] Implement human-readable status output per cli-interface.md
- [ ] T033 [US3] Implement JSON status output (--json flag)
- [ ] T034 [US3] Show actionable instructions for each state (repair hints, etc.)

**Checkpoint**: `smoke hooks status` accurately reports all installation states

---

## Phase 6: User Story 4 - Uninstall Smoke Hooks (Priority: P3)

**Goal**: Users can opt-out by running `smoke hooks uninstall` which removes hooks cleanly

**Independent Test**: Run uninstall and verify hooks are removed while other Claude Code settings remain

### Tests for User Story 4

- [ ] T035 [P] [US4] Unit tests for Uninstall() in internal/hooks/hooks_test.go - hooks present
- [ ] T036 [P] [US4] Unit tests for Uninstall() in internal/hooks/hooks_test.go - preserves other hooks
- [ ] T037 [P] [US4] CLI tests for hooks uninstall command in internal/cli/hooks_test.go

### Implementation for User Story 4

- [ ] T038 [US4] Implement Uninstall() in internal/hooks/hooks.go
- [ ] T039 [US4] Implement `smoke hooks uninstall` subcommand in internal/cli/hooks.go
- [ ] T040 [US4] Ensure only smoke hooks are removed (FR-007)
- [ ] T041 [US4] Handle "not installed" case gracefully
- [ ] T042 [US4] Clean up state directory (~/.claude/hooks/smoke-nudge-state/)

**Checkpoint**: `smoke hooks uninstall` removes only smoke hooks, preserves others

---

## Phase 7: Integration Testing & Polish

**Purpose**: End-to-end tests and cross-cutting improvements

- [ ] T043 Create tests/integration/hooks_test.go for end-to-end hook tests
- [ ] T044 [P] Integration test: fresh init with hooks
- [ ] T045 [P] Integration test: init when hooks already exist
- [ ] T046 [P] Integration test: install/uninstall cycle
- [ ] T047 [P] Integration test: status in all states
- [ ] T048 [P] Integration test: --force flag for modified scripts
- [ ] T049 Edge case handling: invalid settings.json (backup and recover)
- [ ] T050 Edge case handling: permission denied scenarios
- [ ] T051 [P] Edge case test: smoke binary not in PATH during hook execution (graceful degradation)
- [ ] T052 [P] Edge case test: hooks already exist from previous install during fresh init (idempotency)
- [ ] T053 Run `make ci` to verify all quality gates pass
- [ ] T054 Run quickstart.md validation scenarios

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-6)**: All depend on Foundational phase completion
  - User stories can then proceed in priority order (P1 → P2 → P3)
  - Note: US1 modifies init.go, US2-4 share hooks.go - some coordination needed
- **Polish (Phase 7)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - Creates core Install() function
- **User Story 2 (P2)**: Depends on US1's Install() being complete - Adds CLI wrapper and advanced options
- **User Story 3 (P2)**: Can start after Foundational - Independent GetStatus() implementation
- **User Story 4 (P3)**: Can start after Foundational - Independent Uninstall() implementation

### Within Each User Story

- Tests should be written first (TDD approach)
- Core logic before CLI wrappers
- Story complete before moving to next priority

### Parallel Opportunities

**Phase 1** (all can run in parallel after T001):
```bash
Task: T002 "smoke-break.sh script"
Task: T003 "smoke-nudge.sh script"
```

**Phase 2** (sequential - types → helpers → implementation):
- T005 → T006 → T007, T008, T009 (last three can be parallel)

**User Story Tests** (all tests within a story can run in parallel):
```bash
# US1 tests (T010, T011, T012)
# US2 tests (T018, T019, T020, T021)
# US3 tests (T028, T029)
# US4 tests (T035, T036, T037)
```

**Integration Tests** (all can run in parallel after T043):
```bash
Task: T044-T048 all parallel
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001-T004)
2. Complete Phase 2: Foundational (T005-T009)
3. Complete Phase 3: User Story 1 (T010-T017)
4. **STOP and VALIDATE**: Test `smoke init` installs hooks automatically
5. Run `make ci` to verify quality gates

### Incremental Delivery

1. Setup + Foundational → Core types and helpers ready
2. Add User Story 1 → `smoke init` installs hooks → MVP complete
3. Add User Story 2 → `smoke hooks install` command available
4. Add User Story 3 → `smoke hooks status` command available
5. Add User Story 4 → `smoke hooks uninstall` command available
6. Polish → Integration tests, edge cases, validation

### Single Developer Strategy

Execute in strict phase order:
1. Phase 1 (Setup) - ~4 tasks
2. Phase 2 (Foundational) - ~5 tasks
3. Phase 3 (US1 - MVP) - ~8 tasks
4. Phase 4 (US2) - ~10 tasks
5. Phase 5 (US3) - ~7 tasks
6. Phase 6 (US4) - ~8 tasks
7. Phase 7 (Polish) - ~12 tasks

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Write tests first, verify they fail before implementing
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Run `make ci` frequently to catch regressions early
