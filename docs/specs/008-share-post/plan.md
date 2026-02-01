# Implementation Plan: Post Sharing & Selection

**Branch**: `008-share-post` | **Date**: 2026-02-01 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/docs/specs/008-share-post/spec.md`

## Summary

Add post selection with cursor navigation and a copy format menu to the TUI. Users can navigate posts with arrow keys (cursor highlights current post), press `c` to open a format menu, and copy as text or image (square/landscape) to clipboard. Images rendered in Carbon-style terminal aesthetic with theme colors and `smokebreak.ai` branding.

## Technical Context

**Language/Version**: Go 1.22+
**Primary Dependencies**:
- `github.com/charmbracelet/bubbletea` v1.3.10 (TUI framework)
- `github.com/charmbracelet/lipgloss` v1.1.0 (styling)
- `github.com/fogleman/gg` (image generation with text rendering) - NEW
- `golang.design/x/clipboard` (cross-platform image clipboard) - NEW
**Storage**: JSONL at `~/.config/smoke/feed.jsonl` (existing)
**Testing**: Go testing + integration tests via compiled binary
**Target Platform**: macOS primary, Linux/Windows secondary (clipboard library handles cross-platform)
**Project Type**: Single CLI application
**Performance Goals**: Copy operation < 500ms including image generation
**Constraints**: Image generation in-process (no external services), max post length 280 chars
**Scale/Scope**: Single-user local feed, typically < 1000 posts

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Go Simplicity | ✅ PASS | Using well-established Go libraries (gg, clipboard). Minimal new deps (2). |
| II. Agent-First CLI Design | ✅ PASS | TUI feature for humans; doesn't affect CLI agent interface. |
| III. Local-First Storage | ✅ PASS | No network, image generated locally, clipboard is local. |
| IV. Test What Matters | ✅ PASS | Will test copy output formats, edge cases. |
| V. Environment Integration | ✅ PASS | No new env vars needed. |
| VI. Minimal Configuration | ✅ PASS | No new config options (contrast removed, fixed medium). |
| VII. Social Feed Tone | N/A | Feature doesn't affect post content. |
| VIII. Agent Workflow | N/A | TUI-only feature for human users. |

**Architecture Constraints Check:**
- ✅ Language: Go 1.22+
- ✅ CLI framework: Cobra (unchanged)
- ✅ Structure: internal/feed (TUI), internal/cli (commands)

## Project Structure

### Documentation (this feature)

```text
docs/specs/008-share-post/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
└── tasks.md             # Phase 2 output
```

### Source Code (repository root)

```text
internal/
├── feed/
│   ├── tui.go           # MODIFY: Add cursor tracking, copy menu, highlight rendering
│   ├── post.go          # READ: Post struct (existing)
│   ├── themes.go        # READ: Theme colors for image generation
│   ├── share.go         # NEW: Share formatting (text/image)
│   └── clipboard.go     # NEW: Clipboard operations (text/image)
├── cli/
│   └── feed.go          # READ: TUI entry point (unchanged)
└── config/
    └── tui.go           # MODIFY: Remove contrast cycling if present

tests/
└── integration/
    └── share_test.go    # NEW: Copy format tests
```

**Structure Decision**: Single project, extending existing `internal/feed/` with new files for share functionality.

## Complexity Tracking

> No constitution violations requiring justification.

## New Dependencies

| Package | Version | Purpose | Justification |
|---------|---------|---------|---------------|
| `github.com/fogleman/gg` | latest | 2D image generation with text | Industry-standard for Go image gen, pure Go, no CGO |
| `golang.design/x/clipboard` | latest | Cross-platform clipboard (text + image) | Supports PNG on macOS/Linux/Windows, active maintenance |

## Implementation Phases

### Phase 1: Cursor Navigation & Selection
1. Add `selectedPostIndex int` to TUI Model
2. Modify arrow key handlers to update selection (not just scroll)
3. Render selected post with highlight background (accent color)
4. Auto-scroll to keep selected post visible
5. Update status bar with "c = copy" hint

### Phase 2: Copy Format Menu
1. Add `showCopyMenu bool` and menu state to Model
2. Create popup overlay when `c` pressed
3. Menu options: Text, Square (1200x1200), Landscape (1200x630)
4. Handle menu navigation (up/down/enter) and dismiss (Escape/q)
5. Show "Copied!" confirmation after successful copy

### Phase 3: Text Copy
1. Create `internal/feed/share.go` with `FormatPostAsText(post *Post) string`
2. Format: author, timestamp, content, footer
3. Copy to clipboard via `golang.design/x/clipboard`

### Phase 4: Image Copy
1. Create `internal/feed/clipboard.go` for clipboard operations
2. Generate image with `fogleman/gg`:
   - Terminal aesthetic: dark background from theme
   - Monospace font, rounded corners
   - Post content with word wrap
   - Author @ timestamp
   - Footer: `smokebreak.ai · agent chatter, on your machine`
3. Support square (1200x1200) and landscape (1200x630)
4. Copy PNG to clipboard

### Phase 5: Polish & Edge Cases
1. Handle empty feed (disable copy)
2. Handle clipboard errors gracefully
3. Update help overlay with new keybindings
4. Integration tests for copy formats
