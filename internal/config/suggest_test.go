package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadSuggestConfigDefaults(t *testing.T) {
	// Should return defaults when no config file exists
	cfg := LoadSuggestConfig()

	// Verify default contexts exist
	expectedContexts := []string{"conversation", "research", "working"}
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
	expectedCategories := []string{"Observations", "Questions", "Tensions", "Learnings", "Reflections"}
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
	ctx := cfg.GetContext("conversation")
	if ctx == nil {
		t.Fatal("conversation context should exist")
	}
	if ctx.Prompt == "" {
		t.Error("conversation prompt should not be empty")
	}

	// Test non-existing context
	ctx = cfg.GetContext("nonexistent")
	if ctx != nil {
		t.Error("nonexistent context should return nil")
	}
}

func TestGetExamplesForContext(t *testing.T) {
	cfg := LoadSuggestConfig()

	// Conversation context should have Learnings and Reflections examples
	examples := cfg.GetExamplesForContext("conversation")
	if len(examples) == 0 {
		t.Error("conversation context should have examples")
	}

	// Research context should have Observations and Questions examples
	examples = cfg.GetExamplesForContext("research")
	if len(examples) == 0 {
		t.Error("research context should have examples")
	}

	// Non-existing context should return nil
	examples = cfg.GetExamplesForContext("nonexistent")
	if examples != nil {
		t.Error("nonexistent context should return nil examples")
	}
}

func TestGetAllExamples(t *testing.T) {
	cfg := LoadSuggestConfig()

	all := cfg.GetAllExamples()
	if len(all) == 0 {
		t.Error("should have examples")
	}

	// Default has 19 templates total
	if len(all) < 10 {
		t.Errorf("expected at least 10 examples, got %d", len(all))
	}
}

func TestListContextNames(t *testing.T) {
	cfg := LoadSuggestConfig()

	names := cfg.ListContextNames()
	if len(names) < 3 {
		t.Errorf("expected at least 3 contexts, got %d", len(names))
	}

	// Check that expected contexts are in the list
	expected := map[string]bool{"conversation": false, "research": false, "working": false}
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

	// Should contain all three default contexts
	contexts := []string{"conversation:", "research:", "working:"}
	for _, ctx := range contexts {
		if !contains(yaml, ctx) {
			t.Errorf("YAML should contain context %q", ctx)
		}
	}

	// Should contain all five categories
	categories := []string{"Observations:", "Questions:", "Tensions:", "Learnings:", "Reflections:"}
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
      - Questions
      - Observations
  conversation:
    prompt: "Override conversation prompt"
    categories:
      - Learnings

examples:
  Questions:
    - "Custom question example?"
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

	// Verify conversation was overridden
	conv := cfg.GetContext("conversation")
	if conv == nil {
		t.Fatal("conversation context not found")
	}
	if conv.Prompt != "Override conversation prompt" {
		t.Errorf("conversation prompt = %q, want override", conv.Prompt)
	}

	// Verify user examples were merged (appended) to existing category
	questions := cfg.Examples["Questions"]
	found := false
	for _, ex := range questions {
		if ex == "Custom question example?" {
			found = true
			break
		}
	}
	if !found {
		t.Error("custom question example not found in Questions category")
	}
	// Should still have default questions too
	if len(questions) < 2 {
		t.Error("default questions should still exist")
	}

	// Verify new category was added
	newCat := cfg.Examples["NewCategory"]
	if len(newCat) == 0 {
		t.Error("NewCategory should have examples")
	}
	if newCat[0] != "Example in new category" {
		t.Errorf("NewCategory example = %q, want %q", newCat[0], "Example in new category")
	}
}
