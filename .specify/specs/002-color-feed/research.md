# Research: Rich Terminal UI (Color Feed)

## R1: ANSI Color Standards

**Question**: Which ANSI color standard to use?
**Decision**: Standard 8-color (SGR 30-37)
**Rationale**: Universal terminal support. Works on all Unix terminals, Windows Terminal, and most emulators.

**Codes Used**:
| Code | Color | Usage |
|------|-------|-------|
| 31 | Red | Author palette |
| 32 | Green | Author palette |
| 33 | Yellow | Author palette |
| 34 | Blue | Author palette |
| 35 | Magenta | @mentions |
| 36 | Cyan | #hashtags |
| 1 | Bold | Author names |
| 2 | Dim | Timestamps, post IDs |
| 0 | Reset | End formatting |

## R2: TTY Detection in Go

**Question**: How to detect if output is a terminal?
**Decision**: Use `os.Stdout.Stat()` and check for `ModeCharDevice`

**Implementation**:
```go
func isTerminal(f *os.File) bool {
    stat, err := f.Stat()
    if err != nil {
        return false
    }
    return (stat.Mode() & os.ModeCharDevice) != 0
}
```

**Alternative Considered**: `golang.org/x/term.IsTerminal()` - rejected to avoid adding dependency per constitution principle I.

## R3: Box Drawing Characters

**Question**: Which box drawing style?
**Decision**: Rounded corners from Unicode box drawing block

**Characters**:
```
╭─────────────────────────╮
│ Content here            │
╰─────────────────────────╯
```

**Rationale**: Clean modern appearance. UTF-8 widely supported.

## R4: Author Color Assignment

**Question**: How to assign consistent colors to authors?
**Decision**: FNV-1a hash of author string mod palette size

**Implementation**:
```go
func authorColor(author string) string {
    h := fnv.New32a()
    h.Write([]byte(author))
    colors := []string{FgRed, FgGreen, FgYellow, FgBlue, FgMagenta, FgCyan}
    return colors[h.Sum32()%uint32(len(colors))]
}
```

**Rationale**: Deterministic (same author = same color always), fast, no state needed.

## R5: Highlight Regex Patterns

**Question**: What patterns for hashtags and mentions?
**Decision**:
- Hashtags: `#[a-zA-Z0-9_]+`
- Mentions: `@[a-zA-Z0-9_]+`

**Rationale**: Matches common social media conventions. Underscore allowed for multi-word tags like `#code_review`.

## R6: Width Calculation for Boxes

**Question**: How to determine box width?
**Decision**: Fixed width of 80 characters (adjustable via future flag)

**Rationale**: 80 columns is terminal standard. Avoids terminal width detection complexity.

**Future Enhancement**: Could add `COLUMNS` environment variable detection or termios query.

## R7: Flag Precedence

**Question**: How should --color/--no-color interact with TTY detection?
**Decision**: Explicit flags override auto-detection

**Logic**:
1. If `--no-color` set → plain output
2. If `--color` set → colored output
3. Otherwise → auto-detect TTY

**Rationale**: Users need override for edge cases like `smoke feed --color | less -R`.
