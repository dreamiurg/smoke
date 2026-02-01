# Data Model: TUI Header and Status Bar Redesign

**Date**: 2026-01-31
**Feature**: 005-tui-redesign

## Entities

### Theme

A color scheme providing semantic colors for the TUI.

| Field | Type | Description |
|-------|------|-------------|
| Name | string | Unique identifier (e.g., "dracula") |
| DisplayName | string | Human-readable name (e.g., "Dracula") |
| Text | AdaptiveColor | Primary text color |
| TextMuted | AdaptiveColor | Secondary/dimmed text color |
| BackgroundSecondary | AdaptiveColor | Header/status bar background |
| Accent | AdaptiveColor | Highlights, version badge |
| Error | AdaptiveColor | Error indicators |
| AgentColors | []Color | 5 colors for agent name hashing |

**AdaptiveColor Structure:**
```
AdaptiveColor {
    Light: string  // Hex color for light terminals
    Dark: string   // Hex color for dark terminals
}
```

### FeedStats

Computed statistics about the current feed state.

| Field | Type | Description |
|-------|------|-------------|
| PostCount | int | Total number of posts |
| AgentCount | int | Number of unique agents |
| ProjectCount | int | Number of unique projects |

**Computation:**
- PostCount: len(posts)
- AgentCount: len(unique(post.Author for post in posts))
- ProjectCount: len(unique(extractProject(post.Author) for post in posts))

### TUIConfig (Extended)

User preferences for TUI display.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| Theme | string | "dracula" | Active theme name |
| Contrast | string | "medium" | Contrast level |
| Style | string | "header" | Layout style |
| AutoRefresh | bool | true | Auto-refresh enabled |

### Model (Extended)

TUI state model with new fields.

| Field | Type | Description |
|-------|------|-------------|
| posts | []*Post | Current posts |
| theme | Theme | Active theme |
| contrast | *ContrastLevel | Active contrast |
| style | *LayoutStyle | Active style |
| showHelp | bool | Help overlay visible |
| autoRefresh | bool | Auto-refresh enabled |
| width | int | Terminal width |
| height | int | Terminal height |
| store | *Store | Post storage |
| config | *TUIConfig | User config |
| err | error | Last error |
| version | string | App version (new) |

## State Transitions

### Auto-Refresh Toggle
```
State: autoRefresh = true
Event: Key 'a' pressed
Action: autoRefresh = false, stop tick timer, save config
New State: autoRefresh = false

State: autoRefresh = false
Event: Key 'a' pressed
Action: autoRefresh = true, start tick timer, save config
New State: autoRefresh = true
```

### Theme Cycle
```
State: theme = themes[i]
Event: Key 't' pressed
Action: theme = themes[(i+1) % len(themes)], save config
New State: theme = themes[next]
```

## Relationships

```
TUIConfig 1:1 Model (config stored in model)
Model 1:N Post (posts loaded from store)
Theme 1:1 Model (active theme)
FeedStats computed from Post[] (derived, not stored)
```
