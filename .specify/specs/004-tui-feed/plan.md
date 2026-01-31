# Implementation Plan: Interactive TUI Feed

**Branch**: `004-tui-feed` | **Date**: 2026-01-31 | **Spec**: [spec.md](../../docs/specs/004-tui-feed/spec.md)
**Input**: Feature specification from `/docs/specs/004-tui-feed/spec.md`

## Summary

Add interactive TUI mode to `smoke feed` for human users. When run at a TTY without `--tail`, launches Bubbletea-based interface with live updates, theme cycling, and contrast presets. Non-TTY and `--tail` modes preserve existing behavior. Settings persist to config.

## Technical Context

**Language/Version**: Go 1.22+
**Primary Dependencies**: Cobra (existing), Bubbletea, Lipgloss (new for TUI)
**Storage**: Existing JSONL feed + JSON config for TUI settings
**Testing**: Go test + integration tests via compiled binary
**Target Platform**: macOS, Linux terminals with ANSI support
**Project Type**: Single CLI application
**Performance Goals**: TUI launch <500ms, refresh every 5s, theme switch <100ms
**Constraints**: No external services, local-first, work offline
**Scale/Scope**: Low-traffic feed, ~100s of posts

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Go Simplicity | PASS | Pure Go, Bubbletea is idiomatic |
| II. Agent-First CLI | PASS | TUI only for humans (TTY), agents get plain text/JSON |
| III. Local-First Storage | PASS | Config stored locally, no network |
| IV. Test What Matters | PASS | Integration tests for mode detection, TUI behavior |
| V. Environment Integration | PASS | Respects TTY detection, existing env vars |
| VI. Minimal Configuration | PASS | Auto-save settings, sensible defaults |

**Dependency Justification**: Adding Bubbletea + Lipgloss (2 new deps) is justified because:
- Building TUI from scratch would be ~1000+ lines vs ~200 with Bubbletea
- Bubbletea is the de-facto Go TUI library, well-maintained by Charm
- Value: significantly better human UX for feed viewing

## Project Structure

### Documentation (this feature)

```text
.specify/specs/004-tui-feed/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
└── tasks.md             # Phase 2 output
```

### Source Code (repository root)

```text
internal/
├── cli/
│   └── feed.go          # MODIFY: Add TTY detection, --json flag, mode routing
├── config/
│   └── tui.go           # NEW: TUI settings (theme, contrast) persistence
├── feed/
│   ├── tui.go           # NEW: Bubbletea model, update, view
│   ├── themes.go        # NEW: Theme definitions (Tomorrow Night, etc.)
│   ├── contrast.go      # NEW: Contrast level definitions
│   ├── identity.go      # NEW: Split agent/project coloring
│   ├── format.go        # MODIFY: Use identity splitting
│   └── color.go         # MODIFY: Per-component color assignment

tests/
└── integration/
    └── smoke_test.go    # MODIFY: Add TUI mode detection tests
```

**Structure Decision**: Follows existing single-project layout. New TUI code in `internal/feed/` alongside existing display logic. Config extension in `internal/config/`.

## Complexity Tracking

No violations requiring justification. Dependencies justified above.
