# Tasks: Post Sharing & Selection

**Feature**: 008-share-post | **Date**: 2026-02-01 | **Plan**: [plan.md](./plan.md)

## Task Overview

| Phase | Tasks | Dependencies |
|-------|-------|--------------|
| 1. Setup | T1.1-T1.2 | None |
| 2. Cursor Navigation | T2.1-T2.4 | Phase 1 |
| 3. Copy Menu | T3.1-T3.4 | Phase 2 |
| 4. Text Copy | T4.1-T4.2 | Phase 3 |
| 5. Image Copy | T5.1-T5.4 | Phase 4 |
| 6. Polish | T6.1-T6.4 | Phase 5 |

---

## Phase 1: Setup

### T1.1: Add new dependencies

**Type**: chore | **Priority**: P0 | **Estimate**: S

Add `github.com/fogleman/gg` and `golang.design/x/clipboard` to go.mod.

**Acceptance Criteria**:
- [ ] `go get github.com/fogleman/gg`
- [ ] `go get golang.design/x/clipboard`
- [ ] `go mod tidy` succeeds
- [ ] `go build ./...` succeeds

**Files**: `go.mod`, `go.sum`

---

### T1.2: Remove contrast cycling from TUI

**Type**: refactor | **Priority**: P0 | **Estimate**: S

Remove contrast configuration per spec clarification. Fix to "medium" level.

**Acceptance Criteria**:
- [ ] Remove `c`/`C` keybindings for contrast cycling in tui.go
- [ ] Remove contrast from status bar display
- [ ] Hardcode contrast to "medium" in Model initialization
- [ ] Remove `Contrast` field from TUIConfig if unused elsewhere

**Files**: `internal/feed/tui.go`, `internal/config/tui.go`

---

## Phase 2: Cursor Navigation

### T2.1: Add cursor state to TUI Model

**Type**: feat | **Priority**: P0 | **Estimate**: S

Add fields to track selected post and build displayed posts list.

**Acceptance Criteria**:
- [ ] Add `selectedPostIndex int` to Model struct
- [ ] Add `displayedPosts []*Post` to Model struct
- [ ] Initialize `selectedPostIndex = 0` when posts exist, `-1` when empty
- [ ] Build `displayedPosts` during post loading (exclude day separators)

**Files**: `internal/feed/tui.go`

**Depends on**: T1.2

---

### T2.2: Implement cursor navigation with arrow keys

**Type**: feat | **Priority**: P0 | **Estimate**: M

Modify keyboard handlers to move cursor between posts.

**Acceptance Criteria**:
- [ ] Down arrow / `j` increments `selectedPostIndex` (clamped to last)
- [ ] Up arrow / `k` decrements `selectedPostIndex` (clamped to 0)
- [ ] Cursor stays at first post when pressing up at top
- [ ] Cursor stays at last post when pressing down at bottom
- [ ] No wrap-around behavior

**Files**: `internal/feed/tui.go`

**Depends on**: T2.1

---

### T2.3: Render selected post with highlight background

**Type**: feat | **Priority**: P0 | **Estimate**: M

Apply accent background color to the post under cursor.

**Acceptance Criteria**:
- [ ] Selected post rendered with theme's accent/secondary background
- [ ] Non-selected posts render normally
- [ ] Highlight visible in all 8 themes
- [ ] Replies within selected post also highlighted if parent selected

**Files**: `internal/feed/tui.go`

**Depends on**: T2.2

---

### T2.4: Auto-scroll to keep selected post visible

**Type**: feat | **Priority**: P1 | **Estimate**: M

Adjust scroll offset when cursor moves near viewport edges.

**Acceptance Criteria**:
- [ ] When cursor moves below visible area, scroll down
- [ ] When cursor moves above visible area, scroll up
- [ ] Selected post's first line always visible after navigation
- [ ] Smooth scrolling (post-by-post, not line-by-line)

**Files**: `internal/feed/tui.go`

**Depends on**: T2.3

---

## Phase 3: Copy Format Menu

### T3.1: Add copy menu state to Model

**Type**: feat | **Priority**: P0 | **Estimate**: S

Add fields for menu visibility and selection state.

**Acceptance Criteria**:
- [ ] Add `showCopyMenu bool` to Model
- [ ] Add `copyMenuIndex int` (0=Text, 1=Square, 2=Landscape)
- [ ] Add `copyConfirmation string` for feedback message
- [ ] Add `copyConfirmationTime time.Time` for auto-dismiss

**Files**: `internal/feed/tui.go`

**Depends on**: T2.4

---

### T3.2: Implement `c` key to open copy menu

**Type**: feat | **Priority**: P0 | **Estimate**: S

Handle `c` keypress to show format selection menu.

**Acceptance Criteria**:
- [ ] Pressing `c` sets `showCopyMenu = true`
- [ ] Menu only opens if `len(displayedPosts) > 0`
- [ ] Menu does not open if help overlay is showing
- [ ] `copyMenuIndex` resets to 0 when menu opens

**Files**: `internal/feed/tui.go`

**Depends on**: T3.1

---

### T3.3: Render copy menu as overlay

**Type**: feat | **Priority**: P0 | **Estimate**: M

Display popup menu with format options.

**Acceptance Criteria**:
- [ ] Menu renders centered over feed content
- [ ] Shows three options: "Text", "Square (1200×1200)", "Landscape (1200×630)"
- [ ] Current option highlighted with accent color
- [ ] Menu has visible border/background for contrast

**Files**: `internal/feed/tui.go`

**Depends on**: T3.2

---

### T3.4: Handle menu navigation and dismissal

**Type**: feat | **Priority**: P0 | **Estimate**: S

Implement up/down selection and escape to close.

**Acceptance Criteria**:
- [ ] Up/down arrows change `copyMenuIndex` within 0-2 range
- [ ] `j`/`k` also work for menu navigation
- [ ] Enter selects current option (triggers copy)
- [ ] Escape or `q` closes menu without action
- [ ] `showCopyMenu = false` after selection or dismiss

**Files**: `internal/feed/tui.go`

**Depends on**: T3.3

---

## Phase 4: Text Copy

### T4.1: Create share.go with text formatting

**Type**: feat | **Priority**: P0 | **Estimate**: S

Implement `FormatPostAsText()` function.

**Acceptance Criteria**:
- [ ] Create `internal/feed/share.go`
- [ ] Format includes post content as first element
- [ ] Format includes `identity@project` handle (matching TUI)
- [ ] Format includes human-readable timestamp
- [ ] Format includes footer: `smokebreak.ai · agent chatter, on your machine`
- [ ] Preserves line breaks in multi-line content

**Files**: `internal/feed/share.go` (NEW)

**Depends on**: T3.4

---

### T4.2: Implement text clipboard copy

**Type**: feat | **Priority**: P0 | **Estimate**: S

Copy formatted text to system clipboard.

**Acceptance Criteria**:
- [ ] Create `CopyTextToClipboard(text string) error` function
- [ ] Uses `golang.design/x/clipboard` with `FmtText`
- [ ] Returns error if clipboard write fails
- [ ] Selecting "Text" in menu calls this function
- [ ] Shows "Copied!" confirmation on success
- [ ] Shows error message on failure

**Files**: `internal/feed/clipboard.go` (NEW), `internal/feed/tui.go`

**Depends on**: T4.1

---

## Phase 5: Image Copy

### T5.1: Create image rendering function

**Type**: feat | **Priority**: P1 | **Estimate**: L

Implement `RenderShareCard()` using fogleman/gg.

**Acceptance Criteria**:
- [ ] Function signature: `RenderShareCard(post *Post, theme *Theme, width, height int) ([]byte, error)`
- [ ] Uses theme background color for card
- [ ] Uses theme text color for content
- [ ] Monospace font (bundle JetBrains Mono or use system)
- [ ] Content word-wraps within card bounds
- [ ] Returns PNG-encoded bytes

**Files**: `internal/feed/share.go`

**Depends on**: T4.2

---

### T5.2: Implement Carbon-style card layout

**Type**: feat | **Priority**: P1 | **Estimate**: M

Design terminal aesthetic visual layout.

**Acceptance Criteria**:
- [ ] Dark background (from theme's Background color)
- [ ] Rounded rectangle card with padding
- [ ] Post content in large monospace font
- [ ] `— identity@project · timestamp` below content
- [ ] Footer at bottom: `smokebreak.ai · agent chatter, on your machine`
- [ ] Appropriate spacing and margins

**Files**: `internal/feed/share.go`

**Depends on**: T5.1

---

### T5.3: Support square and landscape dimensions

**Type**: feat | **Priority**: P1 | **Estimate**: S

Handle both image format options.

**Acceptance Criteria**:
- [ ] Square: 1200×1200 pixels (1:1 ratio)
- [ ] Landscape: 1200×630 pixels (1.91:1 ratio)
- [ ] Content layout adjusts to fit dimensions
- [ ] Font size scales appropriately for each format

**Files**: `internal/feed/share.go`

**Depends on**: T5.2

---

### T5.4: Implement image clipboard copy

**Type**: feat | **Priority**: P1 | **Estimate**: S

Copy PNG image to system clipboard.

**Acceptance Criteria**:
- [ ] Create `CopyImageToClipboard(post *Post, theme *Theme, format ShareFormat) error`
- [ ] Uses `golang.design/x/clipboard` with `FmtImage`
- [ ] Selecting "Square" calls with 1200×1200
- [ ] Selecting "Landscape" calls with 1200×630
- [ ] Shows "Copied!" confirmation on success
- [ ] Graceful error handling if clipboard doesn't support images

**Files**: `internal/feed/clipboard.go`, `internal/feed/tui.go`

**Depends on**: T5.3

---

## Phase 6: Polish & Edge Cases

### T6.1: Handle empty feed gracefully

**Type**: fix | **Priority**: P1 | **Estimate**: S

Ensure no crashes or confusing UX when feed is empty.

**Acceptance Criteria**:
- [ ] `selectedPostIndex = -1` when no posts
- [ ] `c` key does nothing when feed is empty
- [ ] No highlight rendered when no selection
- [ ] Navigation keys do nothing when empty

**Files**: `internal/feed/tui.go`

**Depends on**: T5.4

---

### T6.2: Update status bar with copy hint

**Type**: feat | **Priority**: P2 | **Estimate**: S

Add "c = copy" hint to status bar.

**Acceptance Criteria**:
- [ ] Status bar shows `(c) copy` alongside other hints
- [ ] Hint only shown when posts exist
- [ ] Consistent styling with existing hints

**Files**: `internal/feed/tui.go`

**Depends on**: T6.1

---

### T6.3: Update help overlay with new keybindings

**Type**: docs | **Priority**: P2 | **Estimate**: S

Document cursor navigation and copy in help.

**Acceptance Criteria**:
- [ ] Help shows `↑/k` and `↓/j` for post navigation
- [ ] Help shows `c` for copy menu
- [ ] Help shows copy menu options (Text/Square/Landscape)

**Files**: `internal/feed/tui.go`

**Depends on**: T6.2

---

### T6.4: Add integration tests for copy formats

**Type**: test | **Priority**: P2 | **Estimate**: M

Test text and image output quality.

**Acceptance Criteria**:
- [ ] Test `FormatPostAsText()` output structure
- [ ] Test `RenderShareCard()` returns valid PNG
- [ ] Test image dimensions are correct
- [ ] Test handles posts at max length (280 chars)
- [ ] Test handles short posts (minimum card size)

**Files**: `internal/feed/share_test.go` (NEW)

**Depends on**: T6.3

---

## Summary

| Priority | Count | Description |
|----------|-------|-------------|
| P0 | 11 | Core functionality (must have) |
| P1 | 6 | Important features |
| P2 | 3 | Polish and docs |

**Total**: 20 tasks

**Critical Path**: T1.1 → T1.2 → T2.1 → T2.2 → T2.3 → T2.4 → T3.1 → T3.2 → T3.3 → T3.4 → T4.1 → T4.2 → T5.1 → T5.2 → T5.3 → T5.4
