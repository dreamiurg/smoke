package feed

import (
	"testing"
)

func TestGetContrastLevel(t *testing.T) {
	tests := []struct {
		name         string
		contrastName string
		want         string
	}{
		{
			name:         "valid contrast medium",
			contrastName: "medium",
			want:         "medium",
		},
		{
			name:         "valid contrast high",
			contrastName: "high",
			want:         "high",
		},
		{
			name:         "valid contrast low",
			contrastName: "low",
			want:         "low",
		},
		{
			name:         "invalid contrast name returns default",
			contrastName: "nonexistent",
			want:         "medium",
		},
		{
			name:         "empty contrast name returns default",
			contrastName: "",
			want:         "medium",
		},
		{
			name:         "case sensitive - uppercase returns default",
			contrastName: "HIGH",
			want:         "medium",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contrast := GetContrastLevel(tt.contrastName)
			if contrast == nil {
				t.Fatal("GetContrastLevel() returned nil")
			}
			if contrast.Name != tt.want {
				t.Errorf("GetContrastLevel(%q).Name = %q, want %q", tt.contrastName, contrast.Name, tt.want)
			}
		})
	}
}

func TestGetContrastLevelProperties(t *testing.T) {
	tests := []struct {
		name           string
		contrastName   string
		wantAgentBold  bool
		wantAgentColor bool
		wantProjColor  bool
	}{
		{
			name:           "medium contrast",
			contrastName:   "medium",
			wantAgentBold:  true,
			wantAgentColor: true,
			wantProjColor:  false,
		},
		{
			name:           "high contrast",
			contrastName:   "high",
			wantAgentBold:  true,
			wantAgentColor: true,
			wantProjColor:  true,
		},
		{
			name:           "low contrast",
			contrastName:   "low",
			wantAgentBold:  false,
			wantAgentColor: false,
			wantProjColor:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contrast := GetContrastLevel(tt.contrastName)

			if contrast.AgentBold != tt.wantAgentBold {
				t.Errorf("GetContrastLevel(%q).AgentBold = %v, want %v", tt.contrastName, contrast.AgentBold, tt.wantAgentBold)
			}
			if contrast.AgentColored != tt.wantAgentColor {
				t.Errorf("GetContrastLevel(%q).AgentColored = %v, want %v", tt.contrastName, contrast.AgentColored, tt.wantAgentColor)
			}
			if contrast.ProjectColored != tt.wantProjColor {
				t.Errorf("GetContrastLevel(%q).ProjectColored = %v, want %v", tt.contrastName, contrast.ProjectColored, tt.wantProjColor)
			}
		})
	}
}

func TestGetContrastLevelReturnsReference(t *testing.T) {
	contrast1 := GetContrastLevel("medium")
	contrast2 := GetContrastLevel("medium")

	if contrast1.Name != contrast2.Name {
		t.Errorf("GetContrastLevel() returned different names: %q vs %q", contrast1.Name, contrast2.Name)
	}

	if contrast1 != contrast2 {
		t.Error("GetContrastLevel() returned different pointers for same contrast level")
	}
}

func TestNextContrastLevel(t *testing.T) {
	tests := []struct {
		name    string
		current string
		want    string
	}{
		{
			name:    "next contrast after medium is high",
			current: "medium",
			want:    "high",
		},
		{
			name:    "next contrast after high is low",
			current: "high",
			want:    "low",
		},
		{
			name:    "next contrast after low wraps to medium",
			current: "low",
			want:    "medium",
		},
		{
			name:    "invalid contrast name returns first contrast",
			current: "nonexistent",
			want:    "medium",
		},
		{
			name:    "empty contrast name returns first contrast",
			current: "",
			want:    "medium",
		},
		{
			name:    "case sensitive - invalid case returns first contrast",
			current: "HIGH",
			want:    "medium",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NextContrastLevel(tt.current)
			if got != tt.want {
				t.Errorf("NextContrastLevel(%q) = %q, want %q", tt.current, got, tt.want)
			}
		})
	}
}

func TestNextContrastLevelCycling(t *testing.T) {
	// Verify that cycling through all contrast levels returns to the first
	current := "medium"
	expected := []string{"high", "low", "medium"}

	for _, exp := range expected {
		current = NextContrastLevel(current)
		if current != exp {
			t.Errorf("NextContrastLevel cycling: expected %q, got %q", exp, current)
		}
	}
}

func TestNextContrastLevelChain(t *testing.T) {
	// Test cycling through all contrast levels twice
	current := AllContrastLevels[0].Name
	expected := make([]string, 0)

	// Build expected cycle twice
	for i := 0; i < 2; i++ {
		for _, contrast := range AllContrastLevels {
			expected = append(expected, contrast.Name)
		}
	}

	// Skip the first one since we start with it
	expected = expected[1:]

	for i, exp := range expected {
		current = NextContrastLevel(current)
		if current != exp {
			t.Errorf("NextContrastLevel chain (step %d): expected %q, got %q", i, exp, current)
		}
	}
}

func TestAllContrastLevelsHaveUniqueNames(t *testing.T) {
	names := make(map[string]bool)
	for _, contrast := range AllContrastLevels {
		if names[contrast.Name] {
			t.Errorf("Contrast level name %q appears more than once", contrast.Name)
		}
		names[contrast.Name] = true
	}
}

func TestAllContrastLevelsHaveValidProperties(t *testing.T) {
	for i, contrast := range AllContrastLevels {
		if contrast.Name == "" {
			t.Errorf("AllContrastLevels[%d].Name is empty", i)
		}
		if contrast.DisplayName == "" {
			t.Errorf("AllContrastLevels[%d].DisplayName is empty", i)
		}
	}
}

func TestDefaultContrastConstant(t *testing.T) {
	if DefaultContrastName != "medium" {
		t.Errorf("DefaultContrastName = %q, want %q", DefaultContrastName, "medium")
	}

	// Verify that the default contrast exists in AllContrastLevels
	defaultContrast := GetContrastLevel(DefaultContrastName)
	if defaultContrast.Name != DefaultContrastName {
		t.Errorf("GetContrastLevel(DefaultContrastName) = %q, want %q", defaultContrast.Name, DefaultContrastName)
	}

	// Verify that GetContrastLevel with empty string returns the default
	emptyContrast := GetContrastLevel("")
	if emptyContrast.Name != DefaultContrastName {
		t.Errorf("GetContrastLevel(\"\") = %q, want %q", emptyContrast.Name, DefaultContrastName)
	}
}

func TestGetContrastLevelAllLevelsAvailable(t *testing.T) {
	// Verify all contrast levels in AllContrastLevels can be retrieved by GetContrastLevel
	for _, expectedContrast := range AllContrastLevels {
		contrast := GetContrastLevel(expectedContrast.Name)
		if contrast.Name != expectedContrast.Name {
			t.Errorf("GetContrastLevel(%q).Name = %q, want %q", expectedContrast.Name, contrast.Name, expectedContrast.Name)
		}
		if contrast.DisplayName != expectedContrast.DisplayName {
			t.Errorf("GetContrastLevel(%q).DisplayName = %q, want %q", expectedContrast.Name, contrast.DisplayName, expectedContrast.DisplayName)
		}
	}
}
