# Tasks: TUI Header and Status Bar Redesign

**Input**: Design documents from `.specify/specs/005-tui-redesign/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md

**Organization**: Tasks are grouped by user story to enable independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2)
- Include exact file paths in descriptions

## Path Conventions

- Single project Go structure at repository root
- Source: `internal/feed/`, `internal/config/`, `internal/cli/`
- Tests: `internal/feed/*_test.go`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Theme system refactor - foundational for all user stories

- [ ] T001 Define new Theme interface with AdaptiveColor support in internal/feed/theme.go
- [ ] T002 [P] Create Dracula theme in internal/feed/theme_dracula.go
- [ ] T003 [P] Create GitHub theme in internal/feed/theme_github.go
- [ ] T004 [P] Create Catppuccin theme in internal/feed/theme_catppuccin.go
- [ ] T005 [P] Create Solarized theme in internal/feed/theme_solarized.go
- [ ] T006 [P] Create Nord theme in internal/feed/theme_nord.go
- [ ] T007 [P] Create Gruvbox theme in internal/feed/theme_gruvbox.go
- [ ] T008 [P] Create One Dark theme in internal/feed/theme_onedark.go
- [ ] T009 [P] Create Tokyo Night theme in internal/feed/theme_tokyonight.go
- [ ] T010 Implement theme registry and cycling in internal/feed/theme.go
- [ ] T011 Migrate existing code to use new Theme interface in internal/feed/tui.go

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure needed before any user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [ ] T012 Add AutoRefresh field to TUIConfig in internal/config/tui.go
- [ ] T013 Create FeedStats struct and ComputeStats function in internal/feed/stats.go
- [ ] T014 Add version field to Model struct in internal/feed/tui.go
- [ ] T015 Update NewModel to accept version parameter in internal/feed/tui.go
- [ ] T016 Pass version from CLI to TUI model in internal/cli/feed.go

**Checkpoint**: Foundation ready - user story implementation can now begin

---

## Phase 3: User Story 1 - View Feed with Context Information (Priority: P1) 🎯 MVP

**Goal**: Display header bar with version, post/agent/project counts, and locale clock

**Independent Test**: Launch TUI and verify header shows all stats and updating clock

### Implementation for User Story 1

- [ ] T017 [US1] Implement renderHeader method in internal/feed/tui.go
- [ ] T018 [US1] Add version badge styling using theme Accent color in internal/feed/tui.go
- [ ] T019 [US1] Integrate FeedStats into header (Posts/Agents/Projects counts) in internal/feed/tui.go
- [ ] T020 [US1] Implement locale-aware clock display in header in internal/feed/tui.go
- [ ] T021 [US1] Add clock tick message for real-time updates in internal/feed/tui.go
- [ ] T022 [US1] Style header bar with BackgroundSecondary in internal/feed/tui.go

**Checkpoint**: Header bar fully functional with version, stats, and updating clock

---

## Phase 4: User Story 2 - View and Understand Current Settings (Priority: P1)

**Goal**: Display status bar with all settings and keybindings

**Independent Test**: Launch TUI and verify status bar shows all settings with (key) format

### Implementation for User Story 2

- [ ] T023 [US2] Refactor renderStatusBar to show settings with keybindings in internal/feed/tui.go
- [ ] T024 [US2] Add auto-refresh state display "(a) auto: ON/OFF" in internal/feed/tui.go
- [ ] T025 [US2] Add style display "(s) style: [name]" in internal/feed/tui.go
- [ ] T026 [US2] Add theme display "(t) theme: [name]" in internal/feed/tui.go
- [ ] T027 [US2] Add contrast display "(c) contrast: [name]" in internal/feed/tui.go
- [ ] T028 [US2] Add help and quit keybindings "(?) help (q) quit" in internal/feed/tui.go
- [ ] T029 [US2] Style status bar with BackgroundSecondary matching header in internal/feed/tui.go

**Checkpoint**: Status bar shows all settings with keybindings

---

## Phase 5: Layout Integration (Combines US1 & US2)

**Goal**: Fixed header/status layout with scrollable content

**Independent Test**: Verify header stays at top, status at bottom during scrolling

- [ ] T030 Refactor View() to use three-section vertical layout in internal/feed/tui.go
- [ ] T031 Calculate content area height as (terminal height - 2) in internal/feed/tui.go
- [ ] T032 Ensure content scrolls independently with fixed bars in internal/feed/tui.go
- [ ] T033 Handle narrow terminal gracefully (truncation) in internal/feed/tui.go

**Checkpoint**: Complete Abacus-style layout working

---

## Phase 6: User Story 3 - Toggle Auto-Refresh (Priority: P2)

**Goal**: Enable/disable auto-refresh with (a) key

**Independent Test**: Press (a) and verify status bar updates and refresh behavior changes

### Implementation for User Story 3

- [ ] T034 [US3] Add autoRefresh field to Model struct in internal/feed/tui.go
- [ ] T035 [US3] Implement (a) key handler to toggle autoRefresh in internal/feed/tui.go
- [ ] T036 [US3] Return conditional tickCmd based on autoRefresh state in internal/feed/tui.go
- [ ] T037 [US3] Save AutoRefresh to config on toggle in internal/feed/tui.go
- [ ] T038 [US3] Load AutoRefresh from config on startup in internal/feed/tui.go

**Checkpoint**: Auto-refresh toggle works and persists

---

## Phase 7: User Story 4 - Switch Between Themes (Priority: P2)

**Goal**: Cycle through 8 themes with (t) key

**Independent Test**: Press (t) repeatedly and verify all 8 themes cycle correctly

### Implementation for User Story 4

- [ ] T039 [US4] Update (t) key handler to use new theme cycling in internal/feed/tui.go
- [ ] T040 [US4] Verify theme persists to config on change in internal/feed/tui.go
- [ ] T041 [US4] Ensure all UI elements update when theme changes in internal/feed/tui.go

**Checkpoint**: Theme cycling works across all 8 themes

---

## Phase 8: User Story 5 - Readable Status Bars (Priority: P2)

**Goal**: Ensure all themes have readable header/status bars

**Independent Test**: Cycle through all themes and visually verify contrast

### Implementation for User Story 5

- [ ] T042 [P] [US5] Verify Dracula theme contrast for bars in internal/feed/theme_dracula.go
- [ ] T043 [P] [US5] Verify GitHub theme contrast for bars in internal/feed/theme_github.go
- [ ] T044 [P] [US5] Verify Catppuccin theme contrast for bars in internal/feed/theme_catppuccin.go
- [ ] T045 [P] [US5] Verify Solarized theme contrast for bars in internal/feed/theme_solarized.go
- [ ] T046 [P] [US5] Verify Nord theme contrast for bars in internal/feed/theme_nord.go
- [ ] T047 [P] [US5] Verify Gruvbox theme contrast for bars in internal/feed/theme_gruvbox.go
- [ ] T048 [P] [US5] Verify One Dark theme contrast for bars in internal/feed/theme_onedark.go
- [ ] T049 [P] [US5] Verify Tokyo Night theme contrast for bars in internal/feed/theme_tokyonight.go

**Checkpoint**: All themes have readable header/status bars

---

## Phase 9: Polish & Testing

**Purpose**: Final quality verification

- [ ] T050 [P] Add unit tests for FeedStats computation in internal/feed/stats_test.go
- [ ] T051 [P] Add unit tests for theme registry in internal/feed/theme_test.go
- [ ] T052 Update help overlay with new keybindings in internal/feed/tui.go
- [ ] T053 Run full test suite and fix any failures
- [ ] T054 Build and manual verification across all themes

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - theme system must be built first
- **Foundational (Phase 2)**: Depends on Setup - config and stats infrastructure
- **US1 Header (Phase 3)**: Depends on Foundational - needs stats and version
- **US2 Status Bar (Phase 4)**: Depends on Foundational - needs config fields
- **Layout (Phase 5)**: Depends on US1 and US2 - combines both bars
- **US3-5 (Phase 6-8)**: Depends on Layout - refinements to working system
- **Polish (Phase 9)**: Depends on all user stories complete

### User Story Dependencies

- **US1 (P1)**: Independent after Foundational
- **US2 (P1)**: Independent after Foundational
- **US3 (P2)**: Can be done after Layout integration
- **US4 (P2)**: Can be done after Layout integration
- **US5 (P2)**: Can be done after all themes exist

### Parallel Opportunities

**Phase 1 - Themes (T002-T009)**: All 8 theme files can be created in parallel

**Phase 8 - Contrast Verification (T042-T049)**: All theme verifications can run in parallel

---

## Parallel Example: Theme Creation

```bash
# Launch all theme file creation in parallel:
Task: "Create Dracula theme in internal/feed/theme_dracula.go"
Task: "Create GitHub theme in internal/feed/theme_github.go"
Task: "Create Catppuccin theme in internal/feed/theme_catppuccin.go"
Task: "Create Solarized theme in internal/feed/theme_solarized.go"
Task: "Create Nord theme in internal/feed/theme_nord.go"
Task: "Create Gruvbox theme in internal/feed/theme_gruvbox.go"
Task: "Create One Dark theme in internal/feed/theme_onedark.go"
Task: "Create Tokyo Night theme in internal/feed/theme_tokyonight.go"
```

---

## Implementation Strategy

### MVP First (Header + Status Bar)

1. Complete Phase 1: Theme System
2. Complete Phase 2: Foundational
3. Complete Phase 3: US1 Header Bar
4. Complete Phase 4: US2 Status Bar
5. Complete Phase 5: Layout Integration
6. **STOP and VALIDATE**: Test basic TUI with header/status bars
7. Deploy as MVP

### Incremental Delivery

1. MVP (Phases 1-5) → Working header/status layout
2. Add US3 (Auto-refresh toggle) → Enhanced control
3. Add US4 (Theme cycling) → Customization
4. Add US5 (Contrast verification) → Quality assurance
5. Polish → Production ready

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story
- Theme files can be created in parallel (T002-T009)
- Layout integration (Phase 5) is the key milestone
- Each story should be independently testable
- Commit after each task or logical group
