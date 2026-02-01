package feed

import "github.com/charmbracelet/lipgloss"

// DefaultThemeName is the default theme when none is specified
const DefaultThemeName = "dracula"

// Theme defines a color palette for the TUI with AdaptiveColor support.
type Theme struct {
	// Name is the identifier for the theme (e.g., "dracula")
	Name string
	// DisplayName is the human-readable name (e.g., "Dracula")
	DisplayName string
	// Text is the primary text color
	Text lipgloss.AdaptiveColor
	// TextMuted is for timestamps and secondary text
	TextMuted lipgloss.AdaptiveColor
	// Background is the main content area background
	Background lipgloss.AdaptiveColor
	// BackgroundSecondary is for header/status bar backgrounds
	BackgroundSecondary lipgloss.AdaptiveColor
	// Accent is for highlights, version badge
	Accent lipgloss.AdaptiveColor
	// Error is for error indicators
	Error lipgloss.AdaptiveColor
	// AgentColors is a palette of 5 colors for agent name hashing
	AgentColors []lipgloss.Color
}

// AllThemes is the registry of available themes.
// Themes cycle in order: dracula → github → catppuccin → solarized → nord → gruvbox → onedark → tokyonight
var AllThemes = []Theme{
	// Dracula - High contrast, vibrant purples/pinks
	{
		Name:                "dracula",
		DisplayName:         "Dracula",
		Text:                lipgloss.AdaptiveColor{Light: "#212121", Dark: "#f8f8f2"},
		TextMuted:           lipgloss.AdaptiveColor{Light: "#757575", Dark: "#6272a4"},
		Background:          lipgloss.AdaptiveColor{Light: "#fafafa", Dark: "#282a36"},
		BackgroundSecondary: lipgloss.AdaptiveColor{Light: "#e0e0e0", Dark: "#44475a"},
		Accent:              lipgloss.AdaptiveColor{Light: "#7e57c2", Dark: "#bd93f9"},
		Error:               lipgloss.AdaptiveColor{Light: "#d32f2f", Dark: "#ff5555"},
		AgentColors: []lipgloss.Color{
			lipgloss.Color("#8be9fd"), // cyan
			lipgloss.Color("#50fa7b"), // green
			lipgloss.Color("#ffb86c"), // orange
			lipgloss.Color("#ff79c6"), // pink
			lipgloss.Color("#bd93f9"), // purple
		},
	},
	// GitHub - Official GitHub theme, WCAG compliant
	{
		Name:                "github",
		DisplayName:         "GitHub",
		Text:                lipgloss.AdaptiveColor{Light: "#24292f", Dark: "#c9d1d9"},
		TextMuted:           lipgloss.AdaptiveColor{Light: "#57606a", Dark: "#8b949e"},
		Background:          lipgloss.AdaptiveColor{Light: "#ffffff", Dark: "#0d1117"},
		BackgroundSecondary: lipgloss.AdaptiveColor{Light: "#f6f8fa", Dark: "#161b22"},
		Accent:              lipgloss.AdaptiveColor{Light: "#0969da", Dark: "#58a6ff"},
		Error:               lipgloss.AdaptiveColor{Light: "#cf222e", Dark: "#f85149"},
		AgentColors: []lipgloss.Color{
			lipgloss.Color("#58a6ff"), // blue
			lipgloss.Color("#3fb950"), // green
			lipgloss.Color("#d29922"), // yellow
			lipgloss.Color("#a371f7"), // purple
			lipgloss.Color("#f78166"), // orange
		},
	},
	// Catppuccin - Modern pastels
	{
		Name:                "catppuccin",
		DisplayName:         "Catppuccin",
		Text:                lipgloss.AdaptiveColor{Light: "#4c4f69", Dark: "#cdd6f4"},
		TextMuted:           lipgloss.AdaptiveColor{Light: "#9ca0b0", Dark: "#6c7086"},
		Background:          lipgloss.AdaptiveColor{Light: "#eff1f5", Dark: "#1e1e2e"},
		BackgroundSecondary: lipgloss.AdaptiveColor{Light: "#e6e9ef", Dark: "#313244"},
		Accent:              lipgloss.AdaptiveColor{Light: "#1e66f5", Dark: "#89b4fa"},
		Error:               lipgloss.AdaptiveColor{Light: "#d20f39", Dark: "#f38ba8"},
		AgentColors: []lipgloss.Color{
			lipgloss.Color("#89b4fa"), // blue
			lipgloss.Color("#a6e3a1"), // green
			lipgloss.Color("#fab387"), // peach
			lipgloss.Color("#cba6f7"), // mauve
			lipgloss.Color("#f5c2e7"), // pink
		},
	},
	// Solarized - Scientifically designed, colorblind-friendly
	{
		Name:                "solarized",
		DisplayName:         "Solarized",
		Text:                lipgloss.AdaptiveColor{Light: "#657b83", Dark: "#839496"},
		TextMuted:           lipgloss.AdaptiveColor{Light: "#93a1a1", Dark: "#586e75"},
		Background:          lipgloss.AdaptiveColor{Light: "#fdf6e3", Dark: "#002b36"},
		BackgroundSecondary: lipgloss.AdaptiveColor{Light: "#eee8d5", Dark: "#073642"},
		Accent:              lipgloss.AdaptiveColor{Light: "#268bd2", Dark: "#268bd2"},
		Error:               lipgloss.AdaptiveColor{Light: "#dc322f", Dark: "#dc322f"},
		AgentColors: []lipgloss.Color{
			lipgloss.Color("#268bd2"), // blue
			lipgloss.Color("#2aa198"), // cyan
			lipgloss.Color("#859900"), // green
			lipgloss.Color("#cb4b16"), // orange
			lipgloss.Color("#d33682"), // magenta
		},
	},
	// Nord - Arctic-inspired, cool tones
	{
		Name:                "nord",
		DisplayName:         "Nord",
		Text:                lipgloss.AdaptiveColor{Light: "#2e3440", Dark: "#eceff4"},
		TextMuted:           lipgloss.AdaptiveColor{Light: "#4c566a", Dark: "#d8dee9"},
		Background:          lipgloss.AdaptiveColor{Light: "#eceff4", Dark: "#2e3440"},
		BackgroundSecondary: lipgloss.AdaptiveColor{Light: "#e5e9f0", Dark: "#3b4252"},
		Accent:              lipgloss.AdaptiveColor{Light: "#5e81ac", Dark: "#88c0d0"},
		Error:               lipgloss.AdaptiveColor{Light: "#bf616a", Dark: "#bf616a"},
		AgentColors: []lipgloss.Color{
			lipgloss.Color("#88c0d0"), // frost cyan
			lipgloss.Color("#a3be8c"), // aurora green
			lipgloss.Color("#ebcb8b"), // aurora yellow
			lipgloss.Color("#b48ead"), // aurora purple
			lipgloss.Color("#81a1c1"), // frost blue
		},
	},
	// Gruvbox - Retro warm tones
	{
		Name:                "gruvbox",
		DisplayName:         "Gruvbox",
		Text:                lipgloss.AdaptiveColor{Light: "#3c3836", Dark: "#ebdbb2"},
		TextMuted:           lipgloss.AdaptiveColor{Light: "#7c6f64", Dark: "#a89984"},
		Background:          lipgloss.AdaptiveColor{Light: "#fbf1c7", Dark: "#282828"},
		BackgroundSecondary: lipgloss.AdaptiveColor{Light: "#f2e5bc", Dark: "#3c3836"},
		Accent:              lipgloss.AdaptiveColor{Light: "#d65d0e", Dark: "#fe8019"},
		Error:               lipgloss.AdaptiveColor{Light: "#cc241d", Dark: "#fb4934"},
		AgentColors: []lipgloss.Color{
			lipgloss.Color("#83a598"), // aqua
			lipgloss.Color("#b8bb26"), // green
			lipgloss.Color("#fabd2f"), // yellow
			lipgloss.Color("#d3869b"), // purple
			lipgloss.Color("#fe8019"), // orange
		},
	},
	// One Dark - Atom's default theme
	{
		Name:                "onedark",
		DisplayName:         "One Dark",
		Text:                lipgloss.AdaptiveColor{Light: "#383a42", Dark: "#abb2bf"},
		TextMuted:           lipgloss.AdaptiveColor{Light: "#a0a1a7", Dark: "#5c6370"},
		Background:          lipgloss.AdaptiveColor{Light: "#fafafa", Dark: "#282c34"},
		BackgroundSecondary: lipgloss.AdaptiveColor{Light: "#f0f0f0", Dark: "#3e4451"},
		Accent:              lipgloss.AdaptiveColor{Light: "#4078f2", Dark: "#61afef"},
		Error:               lipgloss.AdaptiveColor{Light: "#e45649", Dark: "#e06c75"},
		AgentColors: []lipgloss.Color{
			lipgloss.Color("#61afef"), // blue
			lipgloss.Color("#98c379"), // green
			lipgloss.Color("#e5c07b"), // yellow
			lipgloss.Color("#c678dd"), // purple
			lipgloss.Color("#56b6c2"), // cyan
		},
	},
	// Tokyo Night - Modern anime-inspired
	{
		Name:                "tokyonight",
		DisplayName:         "Tokyo Night",
		Text:                lipgloss.AdaptiveColor{Light: "#343b58", Dark: "#c0caf5"},
		TextMuted:           lipgloss.AdaptiveColor{Light: "#9699a3", Dark: "#565f89"},
		Background:          lipgloss.AdaptiveColor{Light: "#e1e2e7", Dark: "#1a1b26"},
		BackgroundSecondary: lipgloss.AdaptiveColor{Light: "#d5d6db", Dark: "#1f2335"},
		Accent:              lipgloss.AdaptiveColor{Light: "#2e7de9", Dark: "#7aa2f7"},
		Error:               lipgloss.AdaptiveColor{Light: "#f52a65", Dark: "#f7768e"},
		AgentColors: []lipgloss.Color{
			lipgloss.Color("#7aa2f7"), // blue
			lipgloss.Color("#9ece6a"), // green
			lipgloss.Color("#e0af68"), // yellow
			lipgloss.Color("#bb9af7"), // purple
			lipgloss.Color("#7dcfff"), // cyan
		},
	},
}

// GetTheme returns the theme with the given name, or the default theme if not found.
func GetTheme(name string) *Theme {
	for i := range AllThemes {
		if AllThemes[i].Name == name {
			return &AllThemes[i]
		}
	}
	return &AllThemes[0]
}

// NextTheme returns the name of the next theme for cycling.
func NextTheme(current string) string {
	for i, t := range AllThemes {
		if t.Name == current {
			return AllThemes[(i+1)%len(AllThemes)].Name
		}
	}
	return AllThemes[0].Name
}

// PrevTheme returns the name of the previous theme for reverse cycling.
func PrevTheme(current string) string {
	for i, t := range AllThemes {
		if t.Name == current {
			return AllThemes[(i-1+len(AllThemes))%len(AllThemes)].Name
		}
	}
	return AllThemes[0].Name
}

// Foreground returns the Text color for backward compatibility
func (t *Theme) Foreground() lipgloss.AdaptiveColor {
	return t.Text
}

// Dim returns the TextMuted color for backward compatibility
func (t *Theme) Dim() lipgloss.AdaptiveColor {
	return t.TextMuted
}
