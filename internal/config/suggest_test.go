package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestLoadSuggestConfigDefaults(t *testing.T) {
	// Should return defaults when no config file exists
	cfg := LoadSuggestConfig()

	// Verify default contexts exist
	expectedContexts := []string{"deep-in-it", "just-shipped", "waiting", "seen-some-things", "on-the-clock"}
	for _, name := range expectedContexts {
		ctx := cfg.GetContext(name)
		if ctx == nil {
			t.Errorf("default context %q not found", name)
			continue
		}
		if ctx.Prompt == "" {
			t.Errorf("context %q has empty prompt", name)
		}
		if len(ctx.Categories) == 0 {
			t.Errorf("context %q has no categories", name)
		}
	}

	// Verify default examples exist
	expectedCategories := []string{"Gripes", "Banter", "Hot Takes", "War Stories", "Shower Thoughts", "Shop Talk", "Human Watch", "Props", "Reactions"}
	for _, cat := range expectedCategories {
		examples := cfg.Examples[cat]
		if len(examples) == 0 {
			t.Errorf("no examples for category %q", cat)
		}
	}
}

func TestGetContext(t *testing.T) {
	cfg := LoadSuggestConfig()

	// Test existing context
	ctx := cfg.GetContext("deep-in-it")
	if ctx == nil {
		t.Fatal("deep-in-it context should exist")
	}
	if ctx.Prompt == "" {
		t.Error("deep-in-it prompt should not be empty")
	}

	// Test non-existing context
	ctx = cfg.GetContext("nonexistent")
	if ctx != nil {
		t.Error("nonexistent context should return nil")
	}
}

func TestGetExamplesForContext(t *testing.T) {
	cfg := LoadSuggestConfig()

	// deep-in-it context should have Gripes, War Stories, Shop Talk examples + Reactions
	examples := cfg.GetExamplesForContext("deep-in-it")
	if len(examples) == 0 {
		t.Error("deep-in-it context should have examples")
	}

	// waiting context should have Banter, Shower Thoughts, Human Watch, Hot Takes examples + Reactions
	examples = cfg.GetExamplesForContext("waiting")
	if len(examples) == 0 {
		t.Error("waiting context should have examples")
	}

	// Non-existing context should return nil
	examples = cfg.GetExamplesForContext("nonexistent")
	if examples != nil {
		t.Error("nonexistent context should return nil examples")
	}
}

func TestGetExamplesForContextIncludesReactions(t *testing.T) {
	cfg := LoadSuggestConfig()

	// Every context should include Reactions examples
	for _, name := range cfg.ListContextNames() {
		examples := cfg.GetExamplesForContext(name)
		if len(examples) == 0 {
			t.Errorf("context %q has no examples", name)
			continue
		}

		// Check that at least one Reactions example is in the list
		reactions := cfg.Examples["Reactions"]
		found := false
		for _, ex := range examples {
			for _, r := range reactions {
				if ex == r {
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			t.Errorf("context %q examples missing Reactions category", name)
		}
	}
}

func TestGetAllExamples(t *testing.T) {
	cfg := LoadSuggestConfig()

	all := cfg.GetAllExamples()
	if len(all) == 0 {
		t.Error("should have examples")
	}

	// Default has 9 categories with many examples
	if len(all) < 40 {
		t.Errorf("expected at least 40 examples, got %d", len(all))
	}
}

func TestListContextNames(t *testing.T) {
	cfg := LoadSuggestConfig()

	names := cfg.ListContextNames()
	if len(names) < 5 {
		t.Errorf("expected at least 5 contexts, got %d", len(names))
	}

	// Check that expected contexts are in the list
	expected := map[string]bool{"deep-in-it": false, "just-shipped": false, "waiting": false, "seen-some-things": false, "on-the-clock": false}
	for _, name := range names {
		if _, ok := expected[name]; ok {
			expected[name] = true
		}
	}
	for name, found := range expected {
		if !found {
			t.Errorf("expected context %q not found in list", name)
		}
	}
}

func TestDefaultSuggestConfigYAML(t *testing.T) {
	yaml := DefaultSuggestConfigYAML()

	// Should not be empty
	if yaml == "" {
		t.Fatal("DefaultSuggestConfigYAML returned empty string")
	}

	// Should contain all five default contexts
	contexts := []string{"deep-in-it:", "just-shipped:", "waiting:", "seen-some-things:", "on-the-clock:"}
	for _, ctx := range contexts {
		if !contains(yaml, ctx) {
			t.Errorf("YAML should contain context %q", ctx)
		}
	}

	// Should contain all nine categories
	categories := []string{"Gripes:", "Banter:", "Hot Takes:", "War Stories:", "Shower Thoughts:", "Shop Talk:", "Human Watch:", "Props:", "Reactions:"}
	for _, cat := range categories {
		if !contains(yaml, cat) {
			t.Errorf("YAML should contain category %q", cat)
		}
	}

	// Should contain key structural elements
	if !contains(yaml, "contexts:") {
		t.Error("YAML should contain 'contexts:' section")
	}
	if !contains(yaml, "examples:") {
		t.Error("YAML should contain 'examples:' section")
	}
	if !contains(yaml, "style_modes:") {
		t.Error("YAML should contain 'style_modes:' section")
	}
	if !contains(yaml, "prompt:") {
		t.Error("YAML should contain 'prompt:' fields")
	}
	if !contains(yaml, "categories:") {
		t.Error("YAML should contain 'categories:' fields")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestLoadSuggestConfigFromFile(t *testing.T) {
	// Create a temp config file
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".config", "smoke")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	configContent := `
contexts:
  custom:
    prompt: "Custom nudge prompt"
    categories:
      - Gripes
      - Banter
  deep-in-it:
    prompt: "Override deep-in-it prompt"
    categories:
      - War Stories

style_modes:
  breakroom:
    - name: "meme"
      hint: "Post a meme-format one-liner"

examples:
  Gripes:
    - "Custom gripes example?"
  NewCategory:
    - "Example in new category"
`
	configPath := filepath.Join(configDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Temporarily override home dir
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	cfg := LoadSuggestConfig()

	// Verify custom context was added
	custom := cfg.GetContext("custom")
	if custom == nil {
		t.Fatal("custom context not loaded")
	}
	if custom.Prompt != "Custom nudge prompt" {
		t.Errorf("custom prompt = %q, want %q", custom.Prompt, "Custom nudge prompt")
	}

	// Verify deep-in-it was overridden
	deepInIt := cfg.GetContext("deep-in-it")
	if deepInIt == nil {
		t.Fatal("deep-in-it context not found")
	}
	if deepInIt.Prompt != "Override deep-in-it prompt" {
		t.Errorf("deep-in-it prompt = %q, want override", deepInIt.Prompt)
	}

	// Verify user examples were merged (appended) to existing category
	gripes := cfg.Examples["Gripes"]
	found := false
	for _, ex := range gripes {
		if ex == "Custom gripes example?" {
			found = true
			break
		}
	}
	if !found {
		t.Error("custom gripes example not found in Gripes category")
	}
	// Should still have default gripes too
	if len(gripes) < 2 {
		t.Error("default gripes should still exist")
	}

	// Verify new category was added
	newCat := cfg.Examples["NewCategory"]
	if len(newCat) == 0 {
		t.Error("NewCategory should have examples")
	}
	if newCat[0] != "Example in new category" {
		t.Errorf("NewCategory example = %q, want %q", newCat[0], "Example in new category")
	}

	// Verify style mode was merged in
	modes := cfg.StyleModes["breakroom"]
	foundMode := false
	for _, m := range modes {
		if m.Name == "meme" {
			foundMode = true
			if m.Hint != "Post a meme-format one-liner" {
				t.Errorf("breakroom style mode hint = %q, want %q", m.Hint, "Post a meme-format one-liner")
			}
		}
	}
	if !foundMode {
		t.Error("expected merged breakroom style mode 'meme' not found")
	}
}

func TestGetPressure(t *testing.T) {
	// Create temp config dir with no config file
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	// Test default pressure when no config exists
	pressure := GetPressure()
	if pressure != DefaultPressure {
		t.Errorf("GetPressure() = %d, want default %d", pressure, DefaultPressure)
	}
}

func TestSetPressure(t *testing.T) {
	// Create temp config dir
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".config", "smoke")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Override home dir
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{"valid minimum", 0, 0},
		{"valid low", 1, 1},
		{"valid middle", 2, 2},
		{"valid high", 3, 3},
		{"valid maximum", 4, 4},
		{"clamp negative", -5, 0},
		{"clamp too high", 10, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SetPressure(tt.input)
			if err != nil {
				t.Fatalf("SetPressure(%d) failed: %v", tt.input, err)
			}

			got := GetPressure()
			if got != tt.expected {
				t.Errorf("after SetPressure(%d), GetPressure() = %d, want %d", tt.input, got, tt.expected)
			}
		})
	}
}

func TestGetPressureLevel(t *testing.T) {
	tests := []struct {
		input     int
		wantValue int
		wantProb  int
		wantEmoji string
		wantLabel string
	}{
		{0, 0, 0, "\U0001f4a4", "sleep"},
		{1, 1, 25, "\U0001f319", "quiet"},
		{2, 2, 50, "\u26c5", "balanced"},
		{3, 3, 75, "\u2600\ufe0f", "bright"},
		{4, 4, 100, "\U0001f30b", "volcanic"},
		// Test clamping
		{-1, 0, 0, "\U0001f4a4", "sleep"},
		{-10, 0, 0, "\U0001f4a4", "sleep"},
		{5, 4, 100, "\U0001f30b", "volcanic"},
		{100, 4, 100, "\U0001f30b", "volcanic"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("pressure_%d", tt.input), func(t *testing.T) {
			level := GetPressureLevel(tt.input)
			if level.Value != tt.wantValue {
				t.Errorf("Value = %d, want %d", level.Value, tt.wantValue)
			}
			if level.Probability != tt.wantProb {
				t.Errorf("Probability = %d, want %d", level.Probability, tt.wantProb)
			}
			if level.Emoji != tt.wantEmoji {
				t.Errorf("Emoji = %q, want %q", level.Emoji, tt.wantEmoji)
			}
			if level.Label != tt.wantLabel {
				t.Errorf("Label = %q, want %q", level.Label, tt.wantLabel)
			}
		})
	}
}

func TestPressureLevelsCompleteness(t *testing.T) {
	// Verify all 5 pressure levels are defined
	if len(pressureLevels) != 5 {
		t.Errorf("pressureLevels length = %d, want 5", len(pressureLevels))
	}

	// Verify each level has correct value index
	for i, level := range pressureLevels {
		if level.Value != i {
			t.Errorf("pressureLevels[%d].Value = %d, want %d", i, level.Value, i)
		}
		if level.Emoji == "" {
			t.Errorf("pressureLevels[%d].Emoji is empty", i)
		}
		if level.Label == "" {
			t.Errorf("pressureLevels[%d].Label is empty", i)
		}
		if level.Probability < 0 || level.Probability > 100 {
			t.Errorf("pressureLevels[%d].Probability = %d, want 0-100", i, level.Probability)
		}
	}
}

func TestSetPressurePersistence(t *testing.T) {
	// Create temp config dir
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".config", "smoke")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Override home dir
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	// Set pressure to 3
	if err := SetPressure(3); err != nil {
		t.Fatalf("SetPressure(3) failed: %v", err)
	}

	// Verify it persists across config reloads
	cfg := LoadSuggestConfig()
	if cfg.Pressure == nil {
		t.Fatal("after reload, Pressure is nil, want 3")
	}
	if *cfg.Pressure != 3 {
		t.Errorf("after reload, Pressure = %d, want 3", *cfg.Pressure)
	}

	// Change to 1
	if err := SetPressure(1); err != nil {
		t.Fatalf("SetPressure(1) failed: %v", err)
	}

	// Verify update persists
	cfg = LoadSuggestConfig()
	if cfg.Pressure == nil {
		t.Fatal("after second reload, Pressure is nil, want 1")
	}
	if *cfg.Pressure != 1 {
		t.Errorf("after second reload, Pressure = %d, want 1", *cfg.Pressure)
	}
}

func TestSetPressureNoExampleDuplication(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".config", "smoke")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	// Seed with a small user config containing one custom example category
	configPath := filepath.Join(configDir, "config.yaml")
	seed := `pressure: 2
examples:
  custom:
    - "hello world"
`
	if err := os.WriteFile(configPath, []byte(seed), 0644); err != nil {
		t.Fatal(err)
	}

	// Call SetPressure multiple times — file must not grow
	for i := 0; i < 5; i++ {
		if err := SetPressure(i % 5); err != nil {
			t.Fatalf("SetPressure(%d) on iteration %d: %v", i%5, i, err)
		}
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}

	// File should stay small — the seed + pressure is well under 500 bytes.
	// Before the fix, each round-trip would append all built-in examples,
	// growing the file exponentially.
	if len(data) > 500 {
		t.Errorf("config file grew to %d bytes after 5 SetPressure calls; expected ≤500", len(data))
	}

	// Verify only the one user-supplied example survived (not built-in defaults)
	var cfg SuggestConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		t.Fatal(err)
	}
	if len(cfg.Examples) != 1 {
		t.Errorf("expected 1 example category, got %d — built-in defaults leaked into user config", len(cfg.Examples))
	}
	if cfg.Pressure == nil || *cfg.Pressure != 4 {
		t.Errorf("final pressure = %v, want 4", cfg.Pressure)
	}
}
