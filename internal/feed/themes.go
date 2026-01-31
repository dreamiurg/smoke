package feed

import "github.com/charmbracelet/lipgloss"

// DefaultThemeName is the default theme when none is specified
const DefaultThemeName = "tomorrow-night"

// Theme defines a color palette for the TUI.
type Theme struct {
	// Name is the identifier for the theme (e.g., "tomorrow-night")
	Name string
	// DisplayName is the human-readable name (e.g., "Tomorrow Night")
	DisplayName string
	// Foreground is the default text color
	Foreground lipgloss.Color
	// Dim is the dimmed text color (for timestamps, etc.)
	Dim lipgloss.Color
	// AgentColors is a palette of 5 colors for agent name hashing
	AgentColors []lipgloss.Color
}

// AllThemes is the registry of available themes.
// Themes will cycle in order: tomorrow-night → monokai → dracula → solarized-light → ...
// Actual colors will be filled in T019-T022.
var AllThemes = []Theme{
	{
		Name:        "tomorrow-night",
		DisplayName: "Tomorrow Night",
		Foreground:  lipgloss.Color("#c5c8c6"),
		Dim:         lipgloss.Color("#969896"),
		AgentColors: []lipgloss.Color{
			lipgloss.Color("#81a2be"), // blue
			lipgloss.Color("#b5bd68"), // green
			lipgloss.Color("#f0c674"), // yellow
			lipgloss.Color("#8abeb7"), // cyan
			lipgloss.Color("#cc6666"), // red
		},
	},
	{
		Name:        "monokai",
		DisplayName: "Monokai",
		Foreground:  lipgloss.Color("#f8f8f2"),
		Dim:         lipgloss.Color("#75715e"),
		AgentColors: []lipgloss.Color{
			lipgloss.Color("#f92672"), // pink
			lipgloss.Color("#a6e22e"), // green
			lipgloss.Color("#fd971f"), // orange
			lipgloss.Color("#66d9ef"), // blue
			lipgloss.Color("#ae81ff"), // purple
		},
	},
	{
		Name:        "dracula",
		DisplayName: "Dracula",
		Foreground:  lipgloss.Color("#f8f8f2"),
		Dim:         lipgloss.Color("#6272a4"),
		AgentColors: []lipgloss.Color{
			lipgloss.Color("#8be9fd"), // cyan
			lipgloss.Color("#50fa7b"), // green
			lipgloss.Color("#ffb86c"), // orange
			lipgloss.Color("#ff79c6"), // pink
			lipgloss.Color("#bd93f9"), // purple
		},
	},
	{
		Name:        "solarized-light",
		DisplayName: "Solarized Light",
		Foreground:  lipgloss.Color("#657b83"),
		Dim:         lipgloss.Color("#93a1a1"),
		AgentColors: []lipgloss.Color{
			lipgloss.Color("#268bd2"), // blue
			lipgloss.Color("#2aa198"), // cyan
			lipgloss.Color("#859900"), // green
			lipgloss.Color("#cb4b16"), // orange
			lipgloss.Color("#dc322f"), // red
		},
	},
}

// GetTheme returns the theme with the given name, or the default theme if not found.
// Default theme is "tomorrow-night".
func GetTheme(name string) *Theme {
	for i := range AllThemes {
		if AllThemes[i].Name == name {
			return &AllThemes[i]
		}
	}
	// Return default theme (first one)
	return &AllThemes[0]
}

// NextTheme returns the name of the next theme for cycling.
// If current theme is not found or is the last one, returns the first theme.
func NextTheme(current string) string {
	for i, t := range AllThemes {
		if t.Name == current {
			// Return next theme, wrapping around to first
			return AllThemes[(i+1)%len(AllThemes)].Name
		}
	}
	// If not found, return first theme
	return AllThemes[0].Name
}
