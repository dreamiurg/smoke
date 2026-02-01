# Tasks: Unread Messages Marker

**Input**: Design documents from `/docs/specs/007-unread-marker/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, quickstart.md

**Tests**: This feature requires unit tests for read state logic and the existing integration test patterns apply. Tests should follow table-driven patterns per CLAUDE.md.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Project type**: Single Go CLI application
- **Source**: `internal/` at repository root
- **Tests**: Co-located `*_test.go` files (Go convention)

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Create the ReadState persistence layer that all user stories depend on

- [ ] T001 [P] Create ReadState struct and YAML schema in internal/config/readstate.go
- [ ] T002 [P] Implement LoadReadState() function in internal/config/readstate.go
- [ ] T003 [P] Implement SaveReadState() with atomic write in internal/config/readstate.go
- [ ] T004 Implement GetLastReadPostID(identity) and SetLastReadPostID(identity, postID) in internal/config/readstate.go
- [ ] T005 Add unit tests for ReadState CRUD operations in internal/config/readstate_test.go

**Checkpoint**: ReadState persistence layer complete and tested

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Extend TUI Model with unread tracking fields needed by all stories

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [ ] T006 Add lastReadPostID, unreadCount, identity fields to feed.Model struct in internal/feed/tui.go
- [ ] T007 Update NewModel() to accept identity parameter in internal/feed/tui.go
- [ ] T008 Load read state on TUI init, set lastReadPostID from stored value in internal/feed/tui.go
- [ ] T009 Implement countUnread() helper function in internal/feed/tui.go
- [ ] T010 Update cli/feed.go to pass identity to NewModel (resolve via SMOKE_NAME or auto-detect)

**Checkpoint**: Foundation ready - TUI has unread tracking fields populated from read state

---

## Phase 3: User Story 1 - See Unread Message Indicator (Priority: P1) 🎯 MVP

**Goal**: Display visual "NEW" separator line between read and unread messages

**Independent Test**: Post new messages, reopen TUI, verify separator appears at correct position

### Implementation for User Story 1

- [ ] T011 [US1] Implement formatUnreadSeparator() method in internal/feed/tui.go (follow formatDaySeparator pattern)
- [ ] T012 [US1] Modify buildAllContentLines() to insert separator when crossing lastReadPostID boundary in internal/feed/tui.go
- [ ] T013 [US1] Add unread count display to renderStatusBar() in internal/feed/tui.go (show "N new" when unreadCount > 0)
- [ ] T014 [US1] Handle first-time user case: no separator when lastReadPostID is empty in internal/feed/tui.go
- [ ] T015 [US1] Handle missing post ID case: treat as all-read if stored ID not in current feed in internal/feed/tui.go
- [ ] T016 [US1] Ensure separator position unchanged during auto-refresh in internal/feed/tui.go

**Checkpoint**: User Story 1 complete - separator visible, count in status bar, stable during refresh

---

## Phase 4: User Story 2 - Mark All as Read (Priority: P2)

**Goal**: Single keypress (`m`) marks all messages as read, separator disappears

**Independent Test**: Open feed with unread messages, press `m`, verify separator disappears and persists across sessions

### Implementation for User Story 2

- [ ] T017 [US2] Add `m` key handler in Update() to mark all as read in internal/feed/tui.go
- [ ] T018 [US2] On mark-all: save latest post ID to read state, clear lastReadPostID, reset unreadCount in internal/feed/tui.go
- [ ] T019 [US2] Add `m` keybinding documentation to renderHelpOverlay() in internal/feed/tui.go
- [ ] T020 [US2] Update status bar to show "(m)ark" or "(m)ark: N new" based on unread count in internal/feed/tui.go

**Checkpoint**: User Story 2 complete - mark-all-read works, state persists, help updated

---

## Phase 5: User Story 3 - Mark Read to Current Scroll Position (Priority: P3)

**Goal**: Shift+M marks messages as read only up to current scroll position

**Independent Test**: Scroll partway through unread messages, press `M`, verify separator moves to that position

### Implementation for User Story 3

- [ ] T021 [US3] Implement getPostAtScrollPosition() helper to find visible post ID in internal/feed/tui.go
- [ ] T022 [US3] Add `M` (Shift+M) key handler in Update() to mark to scroll position in internal/feed/tui.go
- [ ] T023 [US3] On mark-to-here: save visible post ID, update lastReadPostID, recalculate unreadCount in internal/feed/tui.go
- [ ] T024 [US3] Add `M` keybinding documentation to renderHelpOverlay() in internal/feed/tui.go
- [ ] T025 [US3] Handle edge case: at bottom of feed, behave same as mark-all in internal/feed/tui.go

**Checkpoint**: User Story 3 complete - mark-to-position works, help updated

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Edge cases, tests, and final validation

- [ ] T026 [P] Add edge case tests: empty feed, single post, corrupted state in internal/config/readstate_test.go
- [ ] T027 [P] Add edge case tests: missing post ID, multiple identities in internal/config/readstate_test.go
- [ ] T028 Verify separator styling works with all themes (Ember, Ocean, Forest, Mono) in internal/feed/tui.go
- [ ] T029 Verify separator styling works with all contrast levels (Low, Medium, High) in internal/feed/tui.go
- [ ] T030 Run quickstart.md validation sequence manually
- [ ] T031 Run full test suite: make ci

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-5)**: All depend on Foundational phase completion
  - User stories can proceed in priority order (P1 → P2 → P3)
  - P1 should complete first as it's the MVP
- **Polish (Phase 6)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P2)**: Can start after User Story 1 - Builds on separator rendering
- **User Story 3 (P3)**: Can start after User Story 2 - Extends mark-as-read functionality

### Within Each Phase

- Tasks marked [P] can run in parallel (different files or independent logic)
- Sequential tasks depend on prior tasks in the same phase
- Core implementation before integration
- Phase complete before moving to next

### Parallel Opportunities

- Phase 1: T001, T002, T003 can run in parallel (independent functions)
- Phase 6: T026, T027 can run in parallel (independent test files)

---

## Parallel Example: Phase 1 Setup

```bash
# Launch ReadState implementation tasks in parallel:
Task: "Create ReadState struct and YAML schema in internal/config/readstate.go"
Task: "Implement LoadReadState() function in internal/config/readstate.go"
Task: "Implement SaveReadState() with atomic write in internal/config/readstate.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (ReadState persistence)
2. Complete Phase 2: Foundational (TUI fields and init)
3. Complete Phase 3: User Story 1 (separator rendering)
4. **STOP and VALIDATE**: Test separator appears correctly
5. Deploy/demo if ready - core value delivered

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Test separator → MVP delivered
3. Add User Story 2 → Test mark-all → Enhanced UX
4. Add User Story 3 → Test mark-to-position → Power user feature
5. Each story adds value without breaking previous stories

---

## Notes

- [P] tasks = different files or independent logic, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Follow existing code patterns in internal/feed/tui.go and internal/config/*.go
- Avoid: vague tasks, same file conflicts that prevent parallelization
