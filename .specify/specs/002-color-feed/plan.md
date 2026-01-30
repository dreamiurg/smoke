# Implementation Plan: Rich Terminal UI (Color Feed)

**Branch**: `002-color-feed` | **Date**: 2026-01-30 | **Spec**: [spec.md](./spec.md)

## Summary

Add colorful terminal output to the smoke feed with box-drawing borders, author coloring, hashtag highlighting (#cyan), mention highlighting (@magenta), and graceful degradation when piped. Implementation uses Go standard library ANSI escape codes with no external dependencies.

## Technical Context

**Language/Version**: Go 1.22+
**Primary Dependencies**: Standard library only (no new dependencies)
**Storage**: N/A (rendering only, no storage changes)
**Testing**: Go testing (`go test`), integration tests via binary
**Target Platform**: Unix terminals (macOS, Linux), Windows Terminal
**Project Type**: Single CLI application
**Performance Goals**: <500ms for 100 posts with full formatting
**Constraints**: Standard 8-color ANSI codes only, UTF-8 box characters
**Scale/Scope**: Small feature addition to existing CLI

## Constitution Check

*GATE: Pass. All principles satisfied.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Go Simplicity | ✅ | Standard library only, no new dependencies |
| II. CLI-First Design | ✅ | TTY detection, --color/--no-color flags, graceful degradation |
| III. Local-First Storage | ✅ | N/A - rendering only |
| IV. Test What Matters | ✅ | Unit tests for formatters, integration tests for CLI flags |
| V. Gas Town Integration | ✅ | No changes to identity/storage |
| VI. Minimal Configuration | ✅ | Auto-detect TTY, flags only for override |

## Project Structure

### Documentation (this feature)

```text
.specify/specs/002-color-feed/
├── spec.md              # Feature specification
├── plan.md              # This file
├── research.md          # Research decisions
├── checklists/
│   └── requirements.md  # Quality checklist
└── tasks.md             # Task list (after /speckit.tasks)
```

### Source Code (existing structure)

```text
internal/
├── feed/
│   ├── format.go        # MODIFY: Add color formatting
│   ├── format_test.go   # MODIFY: Add color tests
│   ├── color.go         # NEW: ANSI color utilities
│   ├── color_test.go    # NEW: Color unit tests
│   ├── highlight.go     # NEW: Hashtag/mention detection
│   └── highlight_test.go # NEW: Highlight tests
└── cli/
    └── feed.go          # MODIFY: Add --color/--no-color flags

tests/
└── integration/
    └── smoke_test.go    # MODIFY: Add color integration tests
```

**Structure Decision**: Extend existing `internal/feed/` package with new files for color and highlight logic. Keep formatting in `format.go`, add helpers in separate files.

## Complexity Tracking

*No constitution violations. Standard feature addition.*

## Research Decisions

### Terminal Color Support

**Decision**: Use standard 8-color ANSI escape codes (30-37, 40-47)
**Rationale**: Maximum compatibility across terminals. 256-color and true color require detection and fallback logic.
**Alternatives Rejected**:
- 256-color mode: Not universally supported, adds complexity
- True color (24-bit): Limited terminal support

### Box Drawing Characters

**Decision**: Use Unicode box-drawing (U+2500 block) with rounded corners
**Rationale**: Modern terminals support UTF-8. Provides clean visual appearance.
**Alternatives Rejected**:
- ASCII art (+, -, |): Less visually appealing, same compatibility
- No borders: Doesn't meet spec requirement

### TTY Detection

**Decision**: Use `os.Stdout.Fd()` with `golang.org/x/term` or `isatty` check
**Rationale**: Standard Go pattern. The `x/term` package is quasi-standard.
**Alternatives Rejected**:
- Always color: Breaks pipes
- Environment variable only: Less automatic

**Amendment**: Per constitution principle I (minimal dependencies), prefer checking `os.Stdout.Fd()` directly without importing `x/term`. Can use `syscall.Isatty` on Unix or check `os.ModeCharDevice`.

### Author Color Assignment

**Decision**: Hash author name to palette index for deterministic coloring
**Rationale**: Same author always gets same color, no state needed
**Algorithm**: `hash(author) % len(palette)`

### Color Palette

**Decision**: 6 distinct foreground colors (red, green, yellow, blue, magenta, cyan)
**Rationale**: Avoids black (invisible on dark bg) and white (default text)
**Colors**: Red(31), Green(32), Yellow(33), Blue(34), Magenta(35), Cyan(36)

## Key Implementation Details

### ANSI Escape Sequences

```go
const (
    Reset     = "\033[0m"
    Bold      = "\033[1m"
    Dim       = "\033[2m"

    FgRed     = "\033[31m"
    FgGreen   = "\033[32m"
    FgYellow  = "\033[33m"
    FgBlue    = "\033[34m"
    FgMagenta = "\033[35m"
    FgCyan    = "\033[36m"
)
```

### Box Drawing Characters

```go
const (
    BoxTopLeft     = "╭"
    BoxTopRight    = "╮"
    BoxBottomLeft  = "╰"
    BoxBottomRight = "╯"
    BoxHorizontal  = "─"
    BoxVertical    = "│"
)
```

### Highlight Patterns

```go
var (
    hashtagPattern = regexp.MustCompile(`#[a-zA-Z0-9_]+`)
    mentionPattern = regexp.MustCompile(`@[a-zA-Z0-9_]+`)
)
```

## Testing Strategy

1. **Unit Tests** (`internal/feed/*_test.go`):
   - Color code generation
   - Author hash → color mapping
   - Hashtag/mention detection
   - Box drawing width calculation
   - ANSI stripping for plain mode

2. **Integration Tests** (`tests/integration/smoke_test.go`):
   - `smoke feed` with TTY shows colors (visual verification)
   - `smoke feed | cat` strips colors
   - `smoke feed --no-color` forces plain
   - `smoke feed --color` forces color
   - `smoke feed --oneline` with colors

## Dependencies

No new external dependencies. Standard library only.
