# Tasks: Rich Terminal UI (Color Feed)

**Input**: Design documents from `.specify/specs/002-color-feed/`
**Prerequisites**: plan.md, spec.md, research.md

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3, US4)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Color Infrastructure)

**Purpose**: Create color utilities and TTY detection foundation

- [ ] T001 [P] Create ANSI color constants and utilities in internal/feed/color.go
- [ ] T002 [P] Create TTY detection function in internal/feed/tty.go
- [ ] T003 [P] Add unit tests for color utilities in internal/feed/color_test.go
- [ ] T004 [P] Add unit tests for TTY detection in internal/feed/tty_test.go

**Checkpoint**: Color infrastructure ready

---

## Phase 2: User Story 1 - Visual Hierarchy (Priority: P1) MVP

**Goal**: Box-drawing borders, author coloring, dim timestamps

**Independent Test**: `smoke feed` shows bordered posts with colored authors and dim timestamps

### Implementation for User Story 1

- [ ] T005 [US1] Create box-drawing renderer in internal/feed/box.go
- [ ] T006 [US1] Add author color assignment function (hash-based) in internal/feed/color.go
- [ ] T007 [US1] Create ColorWriter wrapper to handle ANSI output in internal/feed/color.go
- [ ] T008 [US1] Modify formatDefault() in internal/feed/format.go to use box borders
- [ ] T009 [US1] Add author coloring to formatDefault() in internal/feed/format.go
- [ ] T010 [US1] Add dim timestamps to formatDefault() in internal/feed/format.go
- [ ] T011 [US1] Add dim post IDs to output in internal/feed/format.go
- [ ] T012 [P] [US1] Add unit tests for box drawing in internal/feed/box_test.go
- [ ] T013 [US1] Add integration test for colored feed in tests/integration/smoke_test.go

**Checkpoint**: Basic colored feed with visual hierarchy working

---

## Phase 3: User Story 2 - Hashtag & Mention Highlighting (Priority: P1)

**Goal**: Detect and colorize #hashtags (cyan) and @mentions (magenta)

**Independent Test**: Post with "#test @user" shows highlighted hashtags and mentions

### Implementation for User Story 2

- [ ] T014 [P] [US2] Create highlight patterns and functions in internal/feed/highlight.go
- [ ] T015 [P] [US2] Add unit tests for highlight detection in internal/feed/highlight_test.go
- [ ] T016 [US2] Integrate highlighting into formatDefault() in internal/feed/format.go
- [ ] T017 [US2] Add integration test for hashtag highlighting in tests/integration/smoke_test.go

**Checkpoint**: Hashtags and mentions highlighted in feed

---

## Phase 4: User Story 3 - Graceful Degradation (Priority: P1)

**Goal**: Auto-detect TTY, provide --color/--no-color flags

**Independent Test**: `smoke feed | cat` produces plain text, `smoke feed --color | cat` produces colored

### Implementation for User Story 3

- [ ] T018 [US3] Add --color and --no-color flags to feed command in internal/cli/feed.go
- [ ] T019 [US3] Implement color mode logic (flag precedence over TTY) in internal/cli/feed.go
- [ ] T020 [US3] Pass color mode through FormatOptions in internal/feed/format.go
- [ ] T021 [US3] Ensure all color output respects ColorEnabled flag in internal/feed/format.go
- [ ] T022 [P] [US3] Add integration test for --no-color flag in tests/integration/smoke_test.go
- [ ] T023 [P] [US3] Add integration test for piped output in tests/integration/smoke_test.go

**Checkpoint**: Color output degrades gracefully

---

## Phase 5: User Story 4 - Oneline Format Enhancement (Priority: P2)

**Goal**: Add colors to compact oneline format

**Independent Test**: `smoke feed --oneline` shows colored post IDs, authors, and highlighted content

### Implementation for User Story 4

- [ ] T024 [US4] Add coloring to formatOneline() in internal/feed/format.go
- [ ] T025 [US4] Ensure highlighting works in oneline format in internal/feed/format.go
- [ ] T026 [US4] Add integration test for colored oneline in tests/integration/smoke_test.go

**Checkpoint**: Oneline format has colors

---

## Phase 6: Polish & Final Integration

**Purpose**: Final cleanup and verification

- [ ] T027 Run full test suite and verify all tests pass
- [ ] T028 Run golangci-lint and fix any issues
- [ ] T029 Update smoke --help text to document --color/--no-color flags
- [ ] T030 Manual testing: verify output on different terminals

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: No dependencies - start immediately
- **Phase 2 (US1)**: Depends on Phase 1 completion
- **Phase 3 (US2)**: Depends on Phase 1, can parallel with Phase 2
- **Phase 4 (US3)**: Depends on Phases 1-3 (needs color output to control)
- **Phase 5 (US4)**: Depends on Phase 1-3
- **Phase 6 (Polish)**: Depends on all user stories

### Parallel Opportunities

Phase 1 tasks (T001-T004) can all run in parallel - different files

Phase 2/3 can partially overlap:
- T005-T007 (box/color infrastructure) can run while T014-T015 (highlighting) proceed
- Integration depends on both completing

### Task Dependencies Within Stories

```
T001 (color.go) ─┬─> T006 (author color) ─> T009 (integrate)
                 └─> T007 (ColorWriter) ─> T008 (box format)

T014 (highlight.go) ─> T016 (integrate into format)

T018-T019 (flags) ─> T020-T021 (pass through format)
```

---

## Implementation Strategy

### MVP (User Stories 1-3)

1. Complete Phase 1: Color infrastructure
2. Complete Phase 2: Visual hierarchy (US1)
3. Complete Phase 3: Highlighting (US2)
4. Complete Phase 4: Graceful degradation (US3)
5. **VALIDATE**: Test all three stories work together
6. Merge to main

### Full Feature (Add US4)

1. Complete MVP
2. Add Phase 5: Oneline enhancement (US4)
3. Complete Phase 6: Polish
4. Final validation and merge

---

## Notes

- All new files should have corresponding test files
- Commit after each logical group of tasks
- Keep color constants centralized in color.go
- Test with both dark and light terminal themes
- Verify pipe behavior before marking US3 complete
