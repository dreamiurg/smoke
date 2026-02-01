# Research: Post Sharing & Selection

**Feature**: 008-share-post | **Date**: 2026-02-01

## Research Questions & Findings

### 1. Bubbletea Popup/Modal Implementation

**Decision**: Implement simple boolean-based popup in Model struct

**Rationale**:
- Smoke's TUI already uses this pattern for the help overlay (`showHelp bool`)
- No need for external overlay library for a simple 3-option menu
- Keeps dependencies minimal per Constitution Principle I

**Alternatives Considered**:
- `github.com/rmhubbert/bubbletea-overlay` - Overkill for a simple menu, adds dependency
- Built-in popup pattern (chosen) - Simple, consistent with existing help overlay

**Implementation Pattern**:
```go
type Model struct {
    // ... existing fields
    showCopyMenu    bool
    copyMenuIndex   int  // 0=Text, 1=Square, 2=Landscape
}

func (m Model) View() string {
    if m.showCopyMenu {
        return m.renderCopyMenu()
    }
    return m.renderFeed()
}
```

### 2. Cross-Platform Image Clipboard

**Decision**: Use `golang.design/x/clipboard`

**Rationale**:
- Supports PNG images on macOS, Linux, and Windows
- Active maintenance (last update 2024)
- Simple API: `clipboard.Write(clipboard.FmtImage, pngBytes)`
- Pure Go, no CGO required on most platforms

**Alternatives Considered**:
- `github.com/atotto/clipboard` - Text only, no image support
- `osascript` via exec.Command - macOS only, fragile
- `pbcopy` - Text only on macOS

**API Usage**:
```go
import "golang.design/x/clipboard"

// Text copy
clipboard.Write(clipboard.FmtText, []byte(text))

// Image copy (PNG bytes)
clipboard.Write(clipboard.FmtImage, pngBytes)
```

### 3. Go Image Generation with Text

**Decision**: Use `github.com/fogleman/gg`

**Rationale**:
- Industry standard for 2D graphics in Go (12k+ GitHub stars)
- Built-in text rendering with word wrap: `DrawStringWrapped()`
- Font loading: `LoadFontFace(path, size)`
- Pure Go, no CGO
- Supports rounded rectangles, gradients, anti-aliasing

**Alternatives Considered**:
- Standard library `image` + `golang.org/x/image/font` - Lower level, more code
- `AdvanceGG` (fork) - Newer but less proven

**API Usage**:
```go
import "github.com/fogleman/gg"

dc := gg.NewContext(1200, 1200)
dc.SetHexColor("#282a36")  // Dracula background
dc.Clear()

// Rounded rectangle
dc.DrawRoundedRectangle(x, y, w, h, radius)
dc.Fill()

// Text with word wrap
dc.LoadFontFace("path/to/font.ttf", 48)
dc.SetHexColor("#f8f8f2")  // Dracula foreground
dc.DrawStringWrapped(text, x, y, 0.5, 0.5, maxWidth, lineSpacing, gg.AlignLeft)

// Export to PNG bytes
var buf bytes.Buffer
dc.EncodePNG(&buf)
pngBytes := buf.Bytes()
```

### 4. Font for Image Generation

**Decision**: Bundle JetBrains Mono or use system monospace font

**Rationale**:
- Monospace font fits terminal aesthetic
- JetBrains Mono is open source (OFL license), popular with developers
- Can embed font file in binary or load from common system paths

**Fallback Strategy**:
1. Try bundled font (if embedded)
2. Try system monospace: `/System/Library/Fonts/Menlo.ttc` (macOS)
3. Try `/usr/share/fonts/truetype/dejavu/DejaVuSansMono.ttf` (Linux)
4. Fall back to basic font rendering

### 5. Current TUI Post Tracking

**Decision**: Add `selectedPostIndex int` to track cursor position

**Rationale**:
- Current TUI only tracks `scrollOffset` (line-based)
- Posts can span multiple lines, no post-to-line mapping exists
- Need to build and cache list of post indices for navigation

**Implementation**:
```go
type Model struct {
    // ... existing
    selectedPostIndex int        // Index into displayedPosts
    displayedPosts    []*Post    // Posts in display order (excludes day separators)
}
```

**Navigation Logic**:
- Up/Down arrows move `selectedPostIndex` within bounds
- Calculate which lines belong to selected post
- Scroll to keep selected post's lines visible

### 6. Social Media Image Dimensions

**Decision**: Support both Square (1200x1200) and Landscape (1200x630)

**Rationale**:
- Square: Universal, works on Twitter/X, Bluesky, LinkedIn, Instagram without cropping
- Landscape: Optimal for link previews and Twitter cards (1.91:1 ratio)
- Both within platform limits (Twitter 5MB, Bluesky 1MB per image)

**Image Size Estimates**:
- 1200x1200 PNG with text: ~50-100KB (well under limits)
- 1200x630 PNG with text: ~30-60KB

## Dependencies Summary

| Package | Purpose | License |
|---------|---------|---------|
| `github.com/fogleman/gg` | Image generation | MIT |
| `golang.design/x/clipboard` | Clipboard access | MIT |

Both are MIT licensed, compatible with Smoke's license.
