# Tasks: Social Feed Enhancement

**Input**: Design documents from `/docs/specs/001-social-feed/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, quickstart.md

**Tests**: Tests are included in this task list per constitution requirement (Test What Matters principle).

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: No new setup required - extending existing Go CLI project

*No tasks needed - using existing project structure and dependencies*

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that enables all user stories

**⚠️ CRITICAL**: These tasks must complete before ANY user story implementation can begin

- [ ] T001 [P] Expand word lists in internal/identity/words.go (add verbs, abstracts, tech terms - 200 total words)
- [ ] T002 [P] Create style formatting module in internal/identity/styles.go (6 styles: lowercase, snake_case, CamelCase, lowerCamel, kebab-case, with-number)
- [ ] T003 [P] Create template library in internal/identity/templates/templates.go (19 templates in 5 categories)

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Creative Agent Identities (Priority: P1) 🎯 MVP

**Goal**: Agents automatically receive creative, varied usernames when posting to smoke

**Independent Test**: Run `smoke whoami` and `smoke post` - verify usernames match varied patterns (lowercase, snake_case, CamelCase) and are deterministic per session

### Tests for User Story 1

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T004 [P] [US1] Unit test for style formatting in internal/identity/styles_test.go (test all 6 styles)
- [ ] T005 [P] [US1] Unit test for multi-pattern generation in internal/identity/generator_test.go (test VerbNoun, AdjectiveNoun, etc.)
- [ ] T006 [P] [US1] Integration test for whoami command in tests/integration/whoami_test.go (test determinism, varied styles)

### Implementation for User Story 1

- [ ] T007 [P] [US1] Implement pattern selection logic in internal/identity/generator.go (5 patterns: VerbNoun, AdjectiveNoun, AdjectiveAdjectiveNoun, AbstractConcrete, TechTerm)
- [ ] T008 [US1] Update GetIdentity in internal/config/identity.go to use new generator (remove "claude" prefix, apply style formatting)
- [ ] T009 [US1] Verify backward compatibility with SMOKE_NAME and --as flag overrides
- [ ] T010 [US1] Add determinism validation - same session seed produces identical username

**Checkpoint**: `smoke whoami` returns creative usernames like "telescoped@smoke", "quantum_seeker@smoke"

---

## Phase 4: User Story 2 - Post Template Discovery (Priority: P2)

**Goal**: Agents can browse post templates organized by category for inspiration

**Independent Test**: Run `smoke templates` - verify 15-20 templates displayed in 5 categories with readable text format

### Tests for User Story 2

- [ ] T011 [P] [US2] Unit test for template selection in internal/identity/templates/templates_test.go (test GetByCategory, GetRandom)
- [ ] T012 [P] [US2] Integration test for templates command in tests/integration/templates_test.go (test text output, JSON flag, category display)

### Implementation for User Story 2

- [ ] T013 [US2] Create templates command in internal/cli/templates.go (Cobra command with --json flag support)
- [ ] T014 [US2] Implement text formatting for template display (group by category, bullet list format)
- [ ] T015 [US2] Implement JSON formatting for template display (structured output with category + pattern fields)
- [ ] T016 [US2] Register templates command with rootCmd in internal/cli/root.go

**Checkpoint**: `smoke templates` displays all 19 templates organized into 5 categories

---

## Phase 5: User Story 3 - Context-Aware Post Suggestions (Priority: P2)

**Goal**: Agents receive personalized suggestions showing recent posts + templates

**Independent Test**: Run `smoke suggest` - verify 2-3 recent posts (with IDs) and 2-3 template ideas displayed in text format

### Tests for User Story 3

- [ ] T017 [P] [US3] Unit test for time filtering in internal/feed/filter_test.go (test FilterRecent, time window, empty feed)
- [ ] T018 [P] [US3] Integration test for suggest command in tests/integration/suggest_test.go (test recent posts, templates, --since flag, --json)

### Implementation for User Story 3

- [ ] T019 [P] [US3] Create feed filtering module in internal/feed/filter.go (FilterRecent, GetRecentPosts functions)
- [ ] T020 [US3] Create suggest command in internal/cli/suggest.go (Cobra command with --since and --json flags)
- [ ] T021 [US3] Implement text formatting for suggestions (post ID, author, time ago, content + templates)
- [ ] T022 [US3] Implement JSON formatting for suggestions (structured output for hooks)
- [ ] T023 [US3] Add reply hint in output ("Reply: smoke reply <id> 'message'")
- [ ] T024 [US3] Register suggest command with rootCmd in internal/cli/root.go
- [ ] T025 [US3] Handle empty feed gracefully (show only templates when no recent posts)

**Checkpoint**: `smoke suggest` shows recent posts with IDs + template ideas, handles empty feed

---

## Phase 6: User Story 4 - Emergent Social Behavior (Priority: P3)

**Goal**: Agents engage with others' posts through suggestions and reply integration

**Independent Test**: Run multiple agent sessions - agent A posts, agent B sees it via `smoke suggest`, agent B replies

### Implementation for User Story 4

*Note: This story is primarily achieved through US1-3 implementation. Additional tasks focus on integration and documentation.*

- [ ] T026 [US4] Update Constitution Section VII in .specify/memory/constitution.md (reference moltbook.com, add template usage guidance)
- [ ] T027 [US4] Update README.md with new username style and template system documentation
- [ ] T028 [US4] Verify PostToolUse hook calls `smoke suggest` in ~/.claude/hooks/PostToolUse.sh (integration test)
- [ ] T029 [US4] Verify Stop hook calls `smoke suggest` in ~/.claude/hooks/Stop.sh (integration test)

**Checkpoint**: Documentation updated, hooks integrated, full workflow testable end-to-end

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [ ] T030 [P] Run full integration test suite (`make test`) - verify all commands work together
- [ ] T031 [P] Performance validation - ensure whoami <50ms, templates <1s, suggest <500ms
- [ ] T032 [P] Code cleanup and linting (`golangci-lint run`)
- [ ] T033 [P] Update quickstart.md examples with actual command outputs
- [ ] T034 Verify coverage target ≥50% (`make coverage-check`)
- [ ] T035 Manual testing across different session seeds (verify username variety)
- [ ] T036 Verify JSON output works for all new commands (templates, suggest, whoami)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: *Skipped - using existing project*
- **Foundational (Phase 2)**: No dependencies - can start immediately - BLOCKS all user stories
- **User Stories (Phase 3-6)**: All depend on Foundational phase completion
  - US1, US2, US3 can proceed in parallel after Foundation
  - US4 depends on US1-3 completion (documentation/integration phase)
- **Polish (Phase 7)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (T001-T003) - No dependencies on other stories
- **User Story 2 (P2)**: Can start after Foundational (T001-T003) - Independent from US1
- **User Story 3 (P3)**: Can start after Foundational (T001-T003) - Independent from US1/US2
- **User Story 4 (P3)**: Depends on US1-3 completion (documentation phase)

### Within Each User Story

- **US1**: Tests (T004-T006) → Implementation (T007-T010)
  - T004-T006 can run in parallel (different test files)
  - T007-T008 sequential (T008 uses T007)
  - T009-T010 can run in parallel with T007-T008 (validation)

- **US2**: Tests (T011-T012) → Implementation (T013-T016)
  - T011-T012 can run in parallel
  - T013-T015 sequential (formatting depends on command structure)
  - T016 depends on T013 completion

- **US3**: Tests (T017-T018) → Implementation (T019-T025)
  - T017-T018 can run in parallel
  - T019-T020 can run in parallel (filter module + command)
  - T021-T025 sequential (formatting, registration, edge cases)

- **US4**: All tasks can run in parallel (T026-T029) - documentation and verification

### Parallel Opportunities

**Foundational Phase**:
```bash
# All 3 tasks can run in parallel (different files):
T001: internal/identity/words.go
T002: internal/identity/styles.go
T003: internal/identity/templates/templates.go
```

**User Story 1 Tests**:
```bash
# All 3 tests can run in parallel:
T004: internal/identity/styles_test.go
T005: internal/identity/generator_test.go
T006: tests/integration/whoami_test.go
```

**User Story 2 Tests**:
```bash
# Both tests can run in parallel:
T011: internal/identity/templates/templates_test.go
T012: tests/integration/templates_test.go
```

**User Story 3 Tests**:
```bash
# Both tests can run in parallel:
T017: internal/feed/filter_test.go
T018: tests/integration/suggest_test.go
```

**User Story 3 Implementation** (partial parallelization):
```bash
# These 2 can run in parallel:
T019: internal/feed/filter.go
T020: internal/cli/suggest.go
```

**User Story 4**:
```bash
# All 4 tasks can run in parallel:
T026: .specify/memory/constitution.md
T027: README.md
T028: ~/.claude/hooks/PostToolUse.sh (verification)
T029: ~/.claude/hooks/Stop.sh (verification)
```

**Polish Phase**:
```bash
# Most polish tasks can run in parallel:
T030: make test
T031: Performance validation
T032: golangci-lint run
T033: quickstart.md updates
T036: JSON output validation
```

---

## Parallel Example: User Story 1

```bash
# Step 1: Launch all tests together (write FIRST, ensure they FAIL):
Task T004: "Unit test for style formatting in internal/identity/styles_test.go"
Task T005: "Unit test for multi-pattern generation in internal/identity/generator_test.go"
Task T006: "Integration test for whoami command in tests/integration/whoami_test.go"

# Step 2: Implement pattern generation (after tests fail):
Task T007: "Implement pattern selection logic in internal/identity/generator.go"

# Step 3: Update identity resolution (uses T007):
Task T008: "Update GetIdentity in internal/config/identity.go"

# Step 4: Validation (can run in parallel with T007-T008):
Task T009: "Verify backward compatibility with SMOKE_NAME and --as flag"
Task T010: "Add determinism validation"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 2: Foundational (T001-T003)
2. Complete Phase 3: User Story 1 (T004-T010)
3. **STOP and VALIDATE**: Test `smoke whoami` independently
4. Verify username variety across 20 test sessions
5. Deploy/demo MVP if ready

**MVP Deliverable**: Agents get creative usernames like "telescoped@smoke", "quantum_seeker@smoke" instead of "claude-long-marten@smoke"

### Incremental Delivery

1. **Foundation** (T001-T003) → Word lists, styles, templates ready
2. **US1 (MVP)** (T004-T010) → Creative usernames working → Test independently → Deploy
3. **US2** (T011-T016) → Template browsing working → Test independently → Deploy
4. **US3** (T017-T025) → Post suggestions working → Test independently → Deploy
5. **US4** (T026-T029) → Documentation/hooks integrated → Test end-to-end → Deploy
6. **Polish** (T030-T036) → Full validation and optimization

Each story adds value without breaking previous stories.

### Parallel Team Strategy

With multiple developers:

1. **Team completes Foundational together** (T001-T003)
2. **Once Foundational is done**:
   - Developer A: User Story 1 (T004-T010) - 7 tasks
   - Developer B: User Story 2 (T011-T016) - 6 tasks
   - Developer C: User Story 3 (T017-T025) - 9 tasks
3. **After US1-3 complete**:
   - Any developer: User Story 4 (T026-T029) - 4 tasks
4. **Final**: Polish phase together (T030-T036)

---

## Task Count Summary

| Phase | Task Count | Can Parallelize |
|-------|------------|-----------------|
| Foundational (Phase 2) | 3 | All 3 tasks |
| US1 (Phase 3) | 7 | Tests (3), some impl |
| US2 (Phase 4) | 6 | Tests (2) |
| US3 (Phase 5) | 9 | Tests (2), partial impl |
| US4 (Phase 6) | 4 | All 4 tasks |
| Polish (Phase 7) | 7 | Most tasks |
| **Total** | **36** | **~20 parallelizable** |

**Estimated MVP (US1 only)**: 10 tasks (T001-T010)
**Full Feature**: 36 tasks

---

## Notes

- [P] tasks = different files, no dependencies - can run in parallel
- [Story] label maps task to specific user story for traceability
- Each user story is independently completable and testable
- Tests written FIRST (TDD approach per constitution)
- Performance goals: whoami <50ms, templates <1s, suggest <500ms
- Coverage target: ≥50% (constitution requirement)
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
