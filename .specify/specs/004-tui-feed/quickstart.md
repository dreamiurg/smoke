# Quickstart: Interactive TUI Feed

## For Users

### Launching the TUI

```bash
# From a terminal (TTY), just run:
smoke feed

# TUI launches automatically, showing live feed
```

### Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `q` | Quit |
| `t` | Cycle through themes |
| `c` | Cycle contrast levels |
| `r` | Force refresh |
| `?` | Show/hide help |

### Non-Interactive Mode

```bash
# Force text output (no TUI)
smoke feed --tail

# JSON output for scripts
smoke feed --json

# Streaming JSON
smoke feed --tail --json
```

## For Developers

### Dependencies

```bash
go get github.com/charmbracelet/bubbletea
go get github.com/charmbracelet/lipgloss
```

### Key Files

| File | Purpose |
|------|---------|
| `internal/feed/tui.go` | Bubbletea Model/Update/View |
| `internal/feed/themes.go` | Theme definitions |
| `internal/feed/contrast.go` | Contrast level definitions |
| `internal/config/tui.go` | Settings persistence |

### Adding a Theme

1. Define colors in `themes.go`:

```go
var MyTheme = Theme{
    Name:        "my-theme",
    DisplayName: "My Theme",
    Foreground:  lipgloss.Color("#ffffff"),
    // ... other colors
}
```

2. Register in `AllThemes` slice
3. Theme becomes available via `t` key cycling

### Testing

```bash
# Run all tests
go test ./...

# Test TUI mode detection
go test ./tests/integration -run TestFeedTUI
```
