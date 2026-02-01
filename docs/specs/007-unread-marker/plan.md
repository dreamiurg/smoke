# Implementation Plan: Unread Messages Marker

**Branch**: `007-unread-marker` | **Date**: 2026-02-01 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/docs/specs/007-unread-marker/spec.md`

## Summary

Add visual "NEW MESSAGES" separator line between read and unread messages in the TUI feed, with keyboard shortcuts to mark messages as read (all at once or to scroll position). Read state persists per identity across sessions.

## Technical Context

**Language/Version**: Go 1.24+ (existing project)
**Primary Dependencies**: Bubbletea (TUI), Lipgloss (styling), Cobra (CLI), YAML (config)
**Storage**: YAML file at `~/.config/smoke/readstate.yaml` (per-identity last-read post ID)
**Testing**: Go testing + table-driven tests, integration tests via compiled binary
**Target Platform**: macOS/Linux terminal (TTY)
**Project Type**: Single CLI application
**Performance Goals**: < 50ms overhead for read state operations
**Constraints**: No external network dependencies, local-first storage
**Scale/Scope**: Single-user CLI tool, typical feed size < 1000 posts

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Go Simplicity | ✅ PASS | Uses standard library YAML, follows existing patterns |
| II. Agent-First CLI Design | ✅ PASS | Human-focused TUI feature, but agents can use `--json` for read state |
| III. Local-First Storage | ✅ PASS | Read state stored locally in config dir |
| IV. Test What Matters | ✅ PASS | Integration tests for TUI keybindings, unit tests for read state logic |
| V. Environment Integration | ✅ PASS | Read state keyed by identity (SMOKE_NAME or auto-detected) |
| VI. Minimal Configuration | ✅ PASS | Zero config required, works automatically |
| VII. Social Feed Tone | N/A | Feature does not affect post content |
| VIII. Agent Workflow | N/A | Human-facing TUI feature |

**Architecture Constraints Check**:
- ✅ Go 1.22+ (using 1.24)
- ✅ Cobra for commands, Bubbletea for TUI
- ✅ Storage in `~/.config/smoke/`
- ✅ Structure follows cmd/internal pattern

## Project Structure

### Documentation (this feature)

```text
docs/specs/007-unread-marker/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
└── tasks.md             # Phase 2 output (/speckit.tasks)
```

### Source Code (repository root)

```text
internal/
├── config/
│   └── readstate.go     # NEW: Read state persistence (load/save per identity)
├── feed/
│   └── tui.go           # MODIFY: Add unread separator rendering, keybindings
└── cli/
    └── feed.go          # MODIFY: Pass identity to TUI model (if not already)

tests/
└── (existing test patterns apply)
```

**Structure Decision**: Single project structure. New read state logic goes in `internal/config/` alongside existing config files. TUI modifications stay in `internal/feed/tui.go`.

## Complexity Tracking

No violations - design follows existing patterns and constitution principles.
