package feed

import (
	"testing"
)

func TestGetTheme(t *testing.T) {
	tests := []struct {
		name      string
		themeName string
		want      string
	}{
		{
			name:      "valid theme tomorrow-night",
			themeName: "tomorrow-night",
			want:      "tomorrow-night",
		},
		{
			name:      "valid theme monokai",
			themeName: "monokai",
			want:      "monokai",
		},
		{
			name:      "valid theme dracula",
			themeName: "dracula",
			want:      "dracula",
		},
		{
			name:      "valid theme solarized-light",
			themeName: "solarized-light",
			want:      "solarized-light",
		},
		{
			name:      "invalid theme name returns default",
			themeName: "nonexistent",
			want:      "tomorrow-night",
		},
		{
			name:      "empty theme name returns default",
			themeName: "",
			want:      "tomorrow-night",
		},
		{
			name:      "case sensitive - lowercase monokai",
			themeName: "monokai",
			want:      "monokai",
		},
		{
			name:      "case sensitive - uppercase returns default",
			themeName: "MONOKAI",
			want:      "tomorrow-night",
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
	theme := GetTheme("tomorrow-night")

	if theme.DisplayName != "Tomorrow Night" {
		t.Errorf("GetTheme().DisplayName = %q, want %q", theme.DisplayName, "Tomorrow Night")
	}

	if theme.Foreground == "" {
		t.Error("GetTheme().Foreground is empty")
	}

	if theme.Dim == "" {
		t.Error("GetTheme().Dim is empty")
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
	theme1 := GetTheme("monokai")
	theme2 := GetTheme("monokai")

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
			name:    "next theme after tomorrow-night is monokai",
			current: "tomorrow-night",
			want:    "monokai",
		},
		{
			name:    "next theme after monokai is dracula",
			current: "monokai",
			want:    "dracula",
		},
		{
			name:    "next theme after dracula is solarized-light",
			current: "dracula",
			want:    "solarized-light",
		},
		{
			name:    "next theme after solarized-light wraps to tomorrow-night",
			current: "solarized-light",
			want:    "tomorrow-night",
		},
		{
			name:    "invalid theme name returns first theme",
			current: "nonexistent",
			want:    "tomorrow-night",
		},
		{
			name:    "empty theme name returns first theme",
			current: "",
			want:    "tomorrow-night",
		},
		{
			name:    "case sensitive - invalid case returns first theme",
			current: "MONOKAI",
			want:    "tomorrow-night",
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

func TestNextThemeCycling(t *testing.T) {
	// Verify that cycling through all themes returns to the first
	current := "tomorrow-night"
	expected := []string{"monokai", "dracula", "solarized-light", "tomorrow-night"}

	for _, exp := range expected {
		current = NextTheme(current)
		if current != exp {
			t.Errorf("NextTheme cycling: expected %q, got %q", exp, current)
		}
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
	for i, theme := range AllThemes {
		if theme.Name == "" {
			t.Errorf("AllThemes[%d].Name is empty", i)
		}
		if theme.DisplayName == "" {
			t.Errorf("AllThemes[%d].DisplayName is empty", i)
		}
		if theme.Foreground == "" {
			t.Errorf("AllThemes[%d].Foreground is empty", i)
		}
		if theme.Dim == "" {
			t.Errorf("AllThemes[%d].Dim is empty", i)
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
	if DefaultThemeName != "tomorrow-night" {
		t.Errorf("DefaultThemeName = %q, want %q", DefaultThemeName, "tomorrow-night")
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
