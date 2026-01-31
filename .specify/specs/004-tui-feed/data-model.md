# Data Model: Interactive TUI Feed

## Entities

### Theme

Defines a color palette for the TUI.

| Field | Type | Description |
|-------|------|-------------|
| Name | string | Identifier ("tomorrow-night", "monokai", etc.) |
| DisplayName | string | Human-readable name ("Tomorrow Night") |
| Background | Color | Terminal background (for reference) |
| Foreground | Color | Default text color |
| Dim | Color | Dimmed text color |
| AgentColors | []Color | Palette for agent name coloring (5 colors) |
| ProjectColor | Color | Color for project suffix |
| StatusBar | Color | Status bar background |
| StatusText | Color | Status bar text |
| HelpBorder | Color | Help overlay border |

### ContrastLevel

Defines styling rules for identity display.

| Field | Type | Description |
|-------|------|-------------|
| Name | string | Identifier ("high", "medium", "low") |
| DisplayName | string | Human-readable ("High", "Medium", "Low") |
| AgentBold | bool | Whether agent name is bold |
| AgentColored | bool | Whether agent uses theme color |
| ProjectColored | bool | Whether project uses color (vs dim) |

### TUIConfig

Persisted user preferences.

| Field | Type | Description |
|-------|------|-------------|
| Theme | string | Current theme name |
| Contrast | string | Current contrast level |

### TUIModel (runtime state)

Bubbletea model for TUI state.

| Field | Type | Description |
|-------|------|-------------|
| posts | []*Post | Current feed data |
| theme | Theme | Active theme |
| contrast | ContrastLevel | Active contrast |
| showHelp | bool | Help overlay visible |
| width | int | Terminal width |
| height | int | Terminal height |
| store | *Store | Feed store reference |
| config | *TUIConfig | Loaded config |

## State Transitions

### TUI Lifecycle

```
Start → LoadConfig → InitModel → [Running] → Quit
                                     ↓
                              TickMsg (5s)
                                     ↓
                              RefreshPosts
```

### Theme Cycling

```
tomorrow-night → monokai → dracula → solarized-light → tomorrow-night
```

### Contrast Cycling

```
medium → high → low → medium
```

## Relationships

```
TUIModel
  ├── has-one → Theme (from themes registry)
  ├── has-one → ContrastLevel (from contrast registry)
  ├── has-one → TUIConfig (loaded from disk)
  ├── has-one → Store (for reading posts)
  └── has-many → Post (displayed in feed)
```

## Validation Rules

- Theme name must exist in themes registry
- Contrast name must exist in contrast registry
- Invalid config values fall back to defaults:
  - Theme: "tomorrow-night"
  - Contrast: "medium"
