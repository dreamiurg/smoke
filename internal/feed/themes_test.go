package feed

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestGetTheme(t *testing.T) {
	tests := []struct {
		name      string
		themeName string
		want      string
	}{
		{
			name:      "valid theme dracula",
			themeName: "dracula",
			want:      "dracula",
		},
		{
			name:      "valid theme github",
			themeName: "github",
			want:      "github",
		},
		{
			name:      "valid theme catppuccin",
			themeName: "catppuccin",
			want:      "catppuccin",
		},
		{
			name:      "valid theme solarized",
			themeName: "solarized",
			want:      "solarized",
		},
		{
			name:      "valid theme nord",
			themeName: "nord",
			want:      "nord",
		},
		{
			name:      "valid theme gruvbox",
			themeName: "gruvbox",
			want:      "gruvbox",
		},
		{
			name:      "valid theme onedark",
			themeName: "onedark",
			want:      "onedark",
		},
		{
			name:      "valid theme tokyonight",
			themeName: "tokyonight",
			want:      "tokyonight",
		},
		{
			name:      "invalid theme name returns default",
			themeName: "nonexistent",
			want:      "dracula",
		},
		{
			name:      "empty theme name returns default",
			themeName: "",
			want:      "dracula",
		},
		{
			name:      "case sensitive - uppercase returns default",
			themeName: "DRACULA",
			want:      "dracula",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			theme := GetTheme(tt.themeName)
			if theme == nil {
				t.Fatal("GetTheme() returned nil")
			}
			if theme.Name != tt.want {
				t.Errorf("GetTheme(%q).Name = %q, want %q", tt.themeName, theme.Name, tt.want)
			}
		})
	}
}

func TestGetThemeProperties(t *testing.T) {
	theme := GetTheme("dracula")

	if theme.DisplayName != "Dracula" {
		t.Errorf("GetTheme().DisplayName = %q, want %q", theme.DisplayName, "Dracula")
	}

	// Check AdaptiveColor fields are not empty
	emptyColor := lipgloss.AdaptiveColor{}
	if theme.Text == emptyColor {
		t.Error("GetTheme().Text is empty")
	}

	if theme.TextMuted == emptyColor {
		t.Error("GetTheme().TextMuted is empty")
	}

	if theme.BackgroundSecondary == emptyColor {
		t.Error("GetTheme().BackgroundSecondary is empty")
	}

	if theme.Accent == emptyColor {
		t.Error("GetTheme().Accent is empty")
	}

	if theme.Error == emptyColor {
		t.Error("GetTheme().Error is empty")
	}

	if len(theme.AgentColors) != 5 {
		t.Errorf("GetTheme().AgentColors length = %d, want 5", len(theme.AgentColors))
	}

	for i, color := range theme.AgentColors {
		if color == "" {
			t.Errorf("GetTheme().AgentColors[%d] is empty", i)
		}
	}
}

func TestGetThemeReturnsReference(t *testing.T) {
	theme1 := GetTheme("github")
	theme2 := GetTheme("github")

	if theme1.Name != theme2.Name {
		t.Errorf("GetTheme() returned different names: %q vs %q", theme1.Name, theme2.Name)
	}

	if theme1 != theme2 {
		t.Error("GetTheme() returned different pointers for same theme")
	}
}

func TestNextTheme(t *testing.T) {
	tests := []struct {
		name    string
		current string
		want    string
	}{
		{
			name:    "next theme after dracula is github",
			current: "dracula",
			want:    "github",
		},
		{
			name:    "next theme after github is catppuccin",
			current: "github",
			want:    "catppuccin",
		},
		{
			name:    "next theme after tokyonight wraps to dracula",
			current: "tokyonight",
			want:    "dracula",
		},
		{
			name:    "invalid theme name returns first theme",
			current: "nonexistent",
			want:    "dracula",
		},
		{
			name:    "empty theme name returns first theme",
			current: "",
			want:    "dracula",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NextTheme(tt.current)
			if got != tt.want {
				t.Errorf("NextTheme(%q) = %q, want %q", tt.current, got, tt.want)
			}
		})
	}
}

func TestNextThemeChain(t *testing.T) {
	// Test cycling through all themes twice
	current := AllThemes[0].Name
	expected := make([]string, 0)

	// Build expected cycle twice
	for i := 0; i < 2; i++ {
		for _, theme := range AllThemes {
			expected = append(expected, theme.Name)
		}
	}

	// Skip the first one since we start with it
	expected = expected[1:]

	for i, exp := range expected {
		current = NextTheme(current)
		if current != exp {
			t.Errorf("NextTheme chain (step %d): expected %q, got %q", i, exp, current)
		}
	}
}

func TestAllThemesHaveUniqueNames(t *testing.T) {
	names := make(map[string]bool)
	for _, theme := range AllThemes {
		if names[theme.Name] {
			t.Errorf("Theme name %q appears more than once", theme.Name)
		}
		names[theme.Name] = true
	}
}

func TestAllThemesHaveValidProperties(t *testing.T) {
	emptyColor := lipgloss.AdaptiveColor{}
	for i, theme := range AllThemes {
		if theme.Name == "" {
			t.Errorf("AllThemes[%d].Name is empty", i)
		}
		if theme.DisplayName == "" {
			t.Errorf("AllThemes[%d].DisplayName is empty", i)
		}
		if theme.Text == emptyColor {
			t.Errorf("AllThemes[%d].Text is empty", i)
		}
		if theme.TextMuted == emptyColor {
			t.Errorf("AllThemes[%d].TextMuted is empty", i)
		}
		if theme.BackgroundSecondary == emptyColor {
			t.Errorf("AllThemes[%d].BackgroundSecondary is empty", i)
		}
		if theme.Accent == emptyColor {
			t.Errorf("AllThemes[%d].Accent is empty", i)
		}
		if theme.Error == emptyColor {
			t.Errorf("AllThemes[%d].Error is empty", i)
		}
		if len(theme.AgentColors) != 5 {
			t.Errorf("AllThemes[%d].AgentColors length = %d, want 5", i, len(theme.AgentColors))
		}
		for j, color := range theme.AgentColors {
			if color == "" {
				t.Errorf("AllThemes[%d].AgentColors[%d] is empty", i, j)
			}
		}
	}
}

func TestDefaultThemeConstant(t *testing.T) {
	if DefaultThemeName != "dracula" {
		t.Errorf("DefaultThemeName = %q, want %q", DefaultThemeName, "dracula")
	}

	// Verify that the default theme exists in AllThemes
	defaultTheme := GetTheme(DefaultThemeName)
	if defaultTheme.Name != DefaultThemeName {
		t.Errorf("GetTheme(DefaultThemeName) = %q, want %q", defaultTheme.Name, DefaultThemeName)
	}

	// Verify that GetTheme with empty string returns the default
	emptyTheme := GetTheme("")
	if emptyTheme.Name != DefaultThemeName {
		t.Errorf("GetTheme(\"\") = %q, want %q", emptyTheme.Name, DefaultThemeName)
	}
}

func TestGetThemeAllThemesAvailable(t *testing.T) {
	// Verify all themes in AllThemes can be retrieved by GetTheme
	for _, expectedTheme := range AllThemes {
		theme := GetTheme(expectedTheme.Name)
		if theme.Name != expectedTheme.Name {
			t.Errorf("GetTheme(%q).Name = %q, want %q", expectedTheme.Name, theme.Name, expectedTheme.Name)
		}
		if theme.DisplayName != expectedTheme.DisplayName {
			t.Errorf("GetTheme(%q).DisplayName = %q, want %q", expectedTheme.Name, theme.DisplayName, expectedTheme.DisplayName)
		}
	}
}

func TestThemeCount(t *testing.T) {
	// Verify we have exactly 8 themes as specified
	expected := 8
	if len(AllThemes) != expected {
		t.Errorf("AllThemes count = %d, want %d", len(AllThemes), expected)
	}
}

func TestBackwardCompatibilityMethods(t *testing.T) {
	theme := GetTheme("dracula")

	// Test Foreground() method
	fg := theme.Foreground()
	if fg != theme.Text {
		t.Error("Foreground() should return Text color")
	}

	// Test Dim() method
	dim := theme.Dim()
	if dim != theme.TextMuted {
		t.Error("Dim() should return TextMuted color")
	}
}
