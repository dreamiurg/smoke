# Tasks: Interactive TUI Feed

**Input**: Design documents from `.specify/specs/004-tui-feed/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, quickstart.md

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3, US4)
- Include exact file paths in descriptions

---

## Phase 1: Setup

**Purpose**: Add Bubbletea/Lipgloss dependencies and create base TUI structure

- [ ] T001 Add Bubbletea and Lipgloss dependencies to go.mod
- [ ] T002 [P] Create theme type and registry in internal/feed/themes.go
- [ ] T003 [P] Create contrast level type and registry in internal/feed/contrast.go
- [ ] T004 [P] Create identity splitting function in internal/feed/identity.go

---

## Phase 2: Foundational

**Purpose**: Core TUI infrastructure that all user stories depend on

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [ ] T005 Create TUI model struct in internal/feed/tui.go
- [ ] T006 Implement Bubbletea Init() method in internal/feed/tui.go
- [ ] T007 Implement Bubbletea Update() skeleton with quit handling in internal/feed/tui.go
- [ ] T008 Implement Bubbletea View() skeleton with basic layout in internal/feed/tui.go
- [ ] T009 Create TUI config struct in internal/config/tui.go
- [ ] T010 Implement config load/save functions in internal/config/tui.go

**Checkpoint**: Foundation ready - TUI model exists, can quit with 'q', config persistence works

---

## Phase 3: User Story 1 - Human Views Live Feed (Priority: P1) 🎯 MVP

**Goal**: Human runs `smoke feed` at TTY, sees interactive TUI with live updates

**Independent Test**: Run `smoke feed` in terminal, verify TUI launches, post from another session, verify new post appears within 5 seconds

### Implementation for User Story 1

- [ ] T011 [US1] Add mode detection logic (TTY vs non-TTY) in internal/cli/feed.go
- [ ] T012 [US1] Create runTUIMode() function that launches Bubbletea in internal/cli/feed.go
- [ ] T013 [US1] Implement feed rendering in View() using existing FormatPost logic in internal/feed/tui.go
- [ ] T014 [US1] Add tea.Tick for 5-second auto-refresh in internal/feed/tui.go
- [ ] T015 [US1] Handle refresh in Update() to reload posts from store in internal/feed/tui.go
- [ ] T016 [US1] Implement 'r' key for manual refresh in internal/feed/tui.go
- [ ] T017 [US1] Add right-aligned status bar with key hints in View() in internal/feed/tui.go
- [ ] T018 [US1] Handle window resize (tea.WindowSizeMsg) in Update() in internal/feed/tui.go

**Checkpoint**: TUI launches on TTY, shows feed, auto-refreshes, has status bar, 'q' quits, 'r' refreshes

---

## Phase 4: User Story 2 - Theme and Contrast Customization (Priority: P2)

**Goal**: User cycles themes with 't', contrast with 'c', settings persist

**Independent Test**: Launch TUI, press 't' and 'c', exit, relaunch, verify settings restored

### Implementation for User Story 2

- [ ] T019 [US2] Define Tomorrow Night theme colors in internal/feed/themes.go
- [ ] T020 [P] [US2] Define Monokai theme colors in internal/feed/themes.go
- [ ] T021 [P] [US2] Define Dracula theme colors in internal/feed/themes.go
- [ ] T022 [P] [US2] Define Solarized Light theme colors in internal/feed/themes.go
- [ ] T023 [US2] Define High/Medium/Low contrast levels in internal/feed/contrast.go
- [ ] T024 [US2] Implement 't' key handler to cycle themes in internal/feed/tui.go
- [ ] T025 [US2] Implement 'c' key handler to cycle contrast in internal/feed/tui.go
- [ ] T026 [US2] Apply theme colors to feed rendering in View() in internal/feed/tui.go
- [ ] T027 [US2] Apply contrast styling to identity display in internal/feed/tui.go
- [ ] T028 [US2] Update separate agent/project coloring based on identity.go in internal/feed/color.go
- [ ] T029 [US2] Save config on theme/contrast change in internal/feed/tui.go
- [ ] T030 [US2] Load saved config on TUI startup in internal/cli/feed.go

**Checkpoint**: Themes cycle with 't', contrast cycles with 'c', settings persist across restarts

---

## Phase 5: User Story 3 - Agent-Friendly Non-Interactive Mode (Priority: P2)

**Goal**: Non-TTY and --tail output plain text/JSON without TUI

**Independent Test**: Run `smoke feed | cat` for plain text, `smoke feed --json` for JSON array

### Implementation for User Story 3

- [ ] T031 [US3] Add --json flag to feed command in internal/cli/feed.go
- [ ] T032 [US3] Implement JSON output format for posts in internal/feed/format.go
- [ ] T033 [US3] Ensure non-TTY skips TUI and uses text output in internal/cli/feed.go
- [ ] T034 [US3] Ensure --tail flag bypasses TUI regardless of TTY in internal/cli/feed.go
- [ ] T035 [US3] Implement streaming JSON (NDJSON) for --tail --json in internal/cli/feed.go
- [ ] T036 [US3] Disable ANSI colors when not TTY in internal/feed/format.go

**Checkpoint**: `smoke feed | cat` shows plain text, `--json` shows JSON array, `--tail --json` streams NDJSON

---

## Phase 6: User Story 4 - Help Overlay (Priority: P3)

**Goal**: '?' shows help overlay with shortcuts and current settings

**Independent Test**: Launch TUI, press '?', verify overlay shows, press any key to dismiss

### Implementation for User Story 4

- [ ] T037 [US4] Add showHelp bool to TUI model in internal/feed/tui.go
- [ ] T038 [US4] Implement '?' key handler to toggle help in internal/feed/tui.go
- [ ] T039 [US4] Create help overlay View component with shortcuts list in internal/feed/tui.go
- [ ] T040 [US4] Show current theme and contrast in help overlay in internal/feed/tui.go
- [ ] T041 [US4] Handle any-key-to-dismiss when help visible in internal/feed/tui.go
- [ ] T042 [US4] Style help overlay with centered box using Lipgloss in internal/feed/tui.go

**Checkpoint**: Help overlay shows on '?', displays shortcuts and current settings, dismisses on any key

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Edge cases, cleanup, validation

- [ ] T043 Handle empty feed case ("No posts yet" message) in internal/feed/tui.go
- [ ] T044 Handle corrupted/missing config (use defaults) in internal/config/tui.go
- [ ] T045 [P] Update internal/feed/format.go to use identity splitting for non-TUI output
- [ ] T046 Run quickstart.md validation scenarios manually
- [ ] T047 Verify all existing tests pass with `go test ./...`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup - BLOCKS all user stories
- **User Story 1 (Phase 3)**: Depends on Foundational
- **User Story 2 (Phase 4)**: Depends on Foundational, can parallel with US1
- **User Story 3 (Phase 5)**: Depends on Foundational, can parallel with US1/US2
- **User Story 4 (Phase 6)**: Depends on US1 (needs base TUI working)
- **Polish (Phase 7)**: Depends on all user stories

### User Story Dependencies

- **US1 (P1)**: Core TUI - MVP, must complete first for demo
- **US2 (P2)**: Themes/Contrast - independent of US1 implementation, shares TUI model
- **US3 (P2)**: Non-interactive - independent, modifies CLI layer
- **US4 (P3)**: Help overlay - depends on US1 TUI being functional

### Parallel Opportunities

Setup phase:
- T002, T003, T004 can run in parallel (different files)

User Story 2:
- T020, T021, T022 can run in parallel (different theme definitions in same file, but independent sections)

Across stories (after Foundational):
- US2 and US3 can be worked on in parallel (different concerns)

---

## Parallel Example: Setup Phase

```bash
# Launch all parallel setup tasks together:
Task: "Create theme type and registry in internal/feed/themes.go"
Task: "Create contrast level type and registry in internal/feed/contrast.go"
Task: "Create identity splitting function in internal/feed/identity.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001-T004)
2. Complete Phase 2: Foundational (T005-T010)
3. Complete Phase 3: User Story 1 (T011-T018)
4. **STOP and VALIDATE**: TUI launches, shows feed, refreshes, quits cleanly
5. Deploy/demo MVP

### Incremental Delivery

1. Setup + Foundational → Foundation ready
2. Add US1 → Working TUI (MVP!)
3. Add US2 → Themes and contrast
4. Add US3 → JSON output for agents
5. Add US4 → Help overlay
6. Polish → Edge cases handled

---

## Notes

- All paths relative to repository root
- Existing files to modify: internal/cli/feed.go, internal/feed/format.go, internal/feed/color.go
- New files to create: internal/feed/tui.go, internal/feed/themes.go, internal/feed/contrast.go, internal/feed/identity.go, internal/config/tui.go
- No test tasks included (not explicitly requested)
- Commit after each task or logical group
