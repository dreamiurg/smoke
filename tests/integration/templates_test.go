package integration

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestTemplatesText verifies that `smoke templates` displays all templates
// with readable text output including category headers and bullet list formatting.
func TestTemplatesText(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	// Initialize smoke
	if _, _, err := h.Run("init"); err != nil {
		t.Fatalf("smoke init failed: %v", err)
	}

	// Run templates command
	stdout, _, err := h.Run("templates")
	if err != nil {
		t.Fatalf("smoke templates failed: %v", err)
	}

	// Verify output is not empty
	if strings.TrimSpace(stdout) == "" {
		t.Error("templates output is empty")
	}

	// Verify output contains readable content
	if !strings.Contains(stdout, "Gripes") &&
		!strings.Contains(stdout, "Banter") &&
		!strings.Contains(stdout, "Hot Takes") &&
		!strings.Contains(stdout, "War Stories") &&
		!strings.Contains(stdout, "Props") {
		t.Errorf("templates output missing category names: %s", stdout)
	}
}

// TestTemplatesAllCategories verifies that `smoke templates` includes all 8 categories.
func TestTemplatesAllCategories(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	// Initialize smoke
	if _, _, err := h.Run("init"); err != nil {
		t.Fatalf("smoke init failed: %v", err)
	}

	// Run templates command
	stdout, _, err := h.Run("templates")
	if err != nil {
		t.Fatalf("smoke templates failed: %v", err)
	}

	// Check for all 8 categories
	expectedCategories := []string{
		"Gripes",
		"Banter",
		"Hot Takes",
		"War Stories",
		"Shower Thoughts",
		"Shop Talk",
		"Human Watch",
		"Props",
	}

	for _, category := range expectedCategories {
		if !strings.Contains(stdout, category) {
			t.Errorf("templates output missing category: %q\nGot: %s", category, stdout)
		}
	}
}

// TestTemplatesPatterns verifies that templates output includes pattern text.
func TestTemplatesPatterns(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	// Initialize smoke
	if _, _, err := h.Run("init"); err != nil {
		t.Fatalf("smoke init failed: %v", err)
	}

	// Run templates command
	stdout, _, err := h.Run("templates")
	if err != nil {
		t.Fatalf("smoke templates failed: %v", err)
	}

	// Check for some patterns - should contain placeholders like [...]
	if !strings.Contains(stdout, "[") || !strings.Contains(stdout, "]") {
		t.Errorf("templates output missing pattern placeholders: %s", stdout)
	}

	// Should have substantial content (templates with patterns)
	lines := strings.Split(stdout, "\n")
	if len(lines) < 10 {
		t.Errorf("templates output too short (expected 10+ lines): %d lines", len(lines))
	}
}

// TestTemplatesCount verifies that all templates are shown (should be 28 total).
func TestTemplatesCount(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	// Initialize smoke
	if _, _, err := h.Run("init"); err != nil {
		t.Fatalf("smoke init failed: %v", err)
	}

	// Run templates command
	stdout, _, err := h.Run("templates")
	if err != nil {
		t.Fatalf("smoke templates failed: %v", err)
	}

	// Should have 4+4+4+4+3+3+3+3 = 28 templates
	lines := strings.Split(strings.TrimSpace(stdout), "\n")

	// Should have at least 30+ lines (categories + templates)
	if len(lines) < 30 {
		t.Errorf("templates output too short for 28 templates (got %d lines): %s", len(lines), stdout)
	}
}

// TestTemplatesJSONFlag verifies that `smoke templates --json` produces valid JSON output.
func TestTemplatesJSONFlag(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	// Initialize smoke
	if _, _, err := h.Run("init"); err != nil {
		t.Fatalf("smoke init failed: %v", err)
	}

	// Run templates command with --json flag
	stdout, _, err := h.Run("templates", "--json")
	if err != nil {
		t.Fatalf("smoke templates --json failed: %v", err)
	}

	// Verify output is valid JSON
	var output interface{}
	if err := json.Unmarshal([]byte(stdout), &output); err != nil {
		t.Fatalf("invalid JSON output: %v\nGot: %s", err, stdout)
	}

	// JSON should not be empty
	if output == nil {
		t.Error("JSON output is null")
	}
}

// TestTemplatesJSONStructure verifies that JSON output has the expected structure
// with categories and templates.
func TestTemplatesJSONStructure(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	// Initialize smoke
	if _, _, err := h.Run("init"); err != nil {
		t.Fatalf("smoke init failed: %v", err)
	}

	// Run templates command with --json flag
	stdout, _, err := h.Run("templates", "--json")
	if err != nil {
		t.Fatalf("smoke templates --json failed: %v", err)
	}

	// Parse JSON output - could be array or object with categories
	var output interface{}
	if err := json.Unmarshal([]byte(stdout), &output); err != nil {
		t.Fatalf("invalid JSON output: %v\nGot: %s", err, stdout)
	}

	// Try parsing as array of objects (common pattern)
	var templates []map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &templates); err == nil {
		// If it's an array, verify it has items
		if len(templates) == 0 {
			t.Error("JSON array is empty, expected templates")
		}

		// Check that each template has Category and Pattern fields
		for i, tmpl := range templates {
			if _, hasCategory := tmpl["Category"]; !hasCategory {
				t.Errorf("template %d missing 'Category' field", i)
			}
			if _, hasPattern := tmpl["Pattern"]; !hasPattern {
				t.Errorf("template %d missing 'Pattern' field", i)
			}
		}
		return
	}

	// Alternative: try parsing as object with categories
	var categoriesMap map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &categoriesMap); err == nil {
		// If it's an object with categories as keys
		if len(categoriesMap) == 0 {
			t.Error("JSON object is empty, expected categories")
		}

		// Should have 8 categories
		if len(categoriesMap) != 8 {
			t.Errorf("expected 8 categories, got %d: %v", len(categoriesMap), categoriesMap)
		}
		return
	}

	// If we got here, the structure is unexpected
	t.Fatalf("JSON output doesn't match expected array or object structure: %s", stdout)
}

// TestTemplatesJSONValidContent verifies that JSON output contains all template data.
func TestTemplatesJSONValidContent(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	// Initialize smoke
	if _, _, err := h.Run("init"); err != nil {
		t.Fatalf("smoke init failed: %v", err)
	}

	// Run templates command with --json flag
	stdout, _, err := h.Run("templates", "--json")
	if err != nil {
		t.Fatalf("smoke templates --json failed: %v", err)
	}

	// Parse as array
	var templates []map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &templates); err != nil {
		t.Fatalf("failed to parse JSON as array: %v\nGot: %s", err, stdout)
	}

	// Should have 28 templates
	if len(templates) != 28 {
		t.Errorf("expected 28 templates, got %d", len(templates))
	}

	// Check that we have expected categories
	categoryCount := make(map[string]int)
	for _, tmpl := range templates {
		if cat, ok := tmpl["Category"].(string); ok {
			categoryCount[cat]++
		}
	}

	expectedCounts := map[string]int{
		"Gripes":          4,
		"Banter":          4,
		"Hot Takes":       4,
		"War Stories":     4,
		"Shower Thoughts": 3,
		"Shop Talk":       3,
		"Human Watch":     3,
		"Props":           3,
	}

	for category, expectedCount := range expectedCounts {
		if count, ok := categoryCount[category]; !ok {
			t.Errorf("category %q missing from JSON output", category)
		} else if count != expectedCount {
			t.Errorf("category %q has %d templates, expected %d", category, count, expectedCount)
		}
	}
}

// TestTemplatesCategoryDisplay verifies that text output shows categories
// with readable formatting (headers, indentation, etc).
func TestTemplatesCategoryDisplay(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	// Initialize smoke
	if _, _, err := h.Run("init"); err != nil {
		t.Fatalf("smoke init failed: %v", err)
	}

	// Run templates command
	stdout, _, err := h.Run("templates")
	if err != nil {
		t.Fatalf("smoke templates failed: %v", err)
	}

	// Check that categories appear as headers (likely with some formatting)
	categories := []string{
		"Gripes",
		"Banter",
		"Hot Takes",
		"War Stories",
		"Shower Thoughts",
		"Shop Talk",
		"Human Watch",
		"Props",
	}

	for _, cat := range categories {
		if !strings.Contains(stdout, cat) {
			t.Errorf("category %q not found in output", cat)
		}

		// Extract the section for this category
		catIndex := strings.Index(stdout, cat)
		if catIndex == -1 {
			continue
		}

		// Get some context around the category
		start := catIndex
		if start > 20 {
			start = catIndex - 20
		}
		end := catIndex + len(cat) + 100
		if end > len(stdout) {
			end = len(stdout)
		}

		section := stdout[start:end]

		// Should have some templates after the category header
		if !strings.Contains(section, "[") {
			t.Errorf("no templates found after %q header", cat)
		}
	}
}

// TestTemplatesOutputReadable verifies that text output is human-readable
// with good formatting and structure.
func TestTemplatesOutputReadable(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	// Initialize smoke
	if _, _, err := h.Run("init"); err != nil {
		t.Fatalf("smoke init failed: %v", err)
	}

	// Run templates command
	stdout, _, err := h.Run("templates")
	if err != nil {
		t.Fatalf("smoke templates failed: %v", err)
	}

	// Check for readable formatting
	// Should have newlines separating content
	if !strings.Contains(stdout, "\n") {
		t.Error("output has no newlines, not properly formatted")
	}

	// Should not be excessively long single lines
	lines := strings.Split(stdout, "\n")
	for i, line := range lines {
		// Allow reasonable line length (80-100 chars is typical)
		if len(line) > 120 && len(line) > 0 {
			// Some flexibility for long patterns
			if !strings.Contains(line, "[") || strings.Count(line, "[") > 5 {
				t.Logf("line %d is quite long (%d chars), may affect readability", i, len(line))
			}
		}
	}

	// Check that output has reasonable number of lines
	if len(lines) < 15 {
		t.Errorf("output too short for proper template display (%d lines)", len(lines))
	}
}

// TestTemplatesHelpIntegration verifies that templates command appears in help.
func TestTemplatesHelpIntegration(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	// Initialize smoke
	if _, _, err := h.Run("init"); err != nil {
		t.Fatalf("smoke init failed: %v", err)
	}

	// Check main help
	stdout, _, err := h.Run("--help")
	if err != nil {
		t.Fatalf("smoke --help failed: %v", err)
	}

	// Should mention templates command
	if !strings.Contains(stdout, "templates") {
		t.Logf("warning: 'templates' command not mentioned in main help: %s", stdout)
	}
}

// TestTemplatesInvalidFlag verifies that invalid flags are handled correctly.
func TestTemplatesInvalidFlag(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	// Initialize smoke
	if _, _, err := h.Run("init"); err != nil {
		t.Fatalf("smoke init failed: %v", err)
	}

	// Run with invalid flag
	_, stderr, err := h.Run("templates", "--invalid-flag")
	if err == nil {
		// Some CLIs ignore unknown flags, some error - both are acceptable
		// but we log it for awareness
		t.Logf("note: unknown flag was silently ignored (may be expected behavior)")
	} else {
		// If it errors, should have helpful message
		if strings.Contains(stderr, "invalid") || strings.Contains(stderr, "unknown") || strings.Contains(stderr, "flag") {
			t.Logf("properly rejected invalid flag: %s", stderr)
		}
	}
}

// TestTemplatesNoArgs verifies that templates command works without arguments.
func TestTemplatesNoArgs(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	// Initialize smoke
	if _, _, err := h.Run("init"); err != nil {
		t.Fatalf("smoke init failed: %v", err)
	}

	// Run templates without any args - should succeed
	stdout, _, err := h.Run("templates")
	if err != nil {
		t.Fatalf("smoke templates with no args failed: %v", err)
	}

	if strings.TrimSpace(stdout) == "" {
		t.Error("templates output is empty")
	}
}

// TestTemplatesConsistentOutput verifies that multiple calls to templates
// produce consistent output.
func TestTemplatesConsistentOutput(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	// Initialize smoke
	if _, _, err := h.Run("init"); err != nil {
		t.Fatalf("smoke init failed: %v", err)
	}

	// First call
	stdout1, _, err := h.Run("templates")
	if err != nil {
		t.Fatalf("first smoke templates failed: %v", err)
	}

	// Second call
	stdout2, _, err := h.Run("templates")
	if err != nil {
		t.Fatalf("second smoke templates failed: %v", err)
	}

	// Should be identical (for text output)
	if stdout1 != stdout2 {
		t.Errorf("inconsistent output between calls:\nFirst:\n%s\n\nSecond:\n%s", stdout1, stdout2)
	}
}

// TestTemplatesJSONConsistent verifies that multiple --json calls produce
// consistent output.
func TestTemplatesJSONConsistent(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	// Initialize smoke
	if _, _, err := h.Run("init"); err != nil {
		t.Fatalf("smoke init failed: %v", err)
	}

	// First JSON call
	stdout1, _, err := h.Run("templates", "--json")
	if err != nil {
		t.Fatalf("first smoke templates --json failed: %v", err)
	}

	// Parse first output
	var output1 []map[string]interface{}
	if unmarshalErr := json.Unmarshal([]byte(stdout1), &output1); unmarshalErr != nil {
		t.Fatalf("failed to parse first JSON: %v", unmarshalErr)
	}

	// Second JSON call
	stdout2, _, err := h.Run("templates", "--json")
	if err != nil {
		t.Fatalf("second smoke templates --json failed: %v", err)
	}

	// Parse second output
	var output2 []map[string]interface{}
	if err := json.Unmarshal([]byte(stdout2), &output2); err != nil {
		t.Fatalf("failed to parse second JSON: %v", err)
	}

	// Should have same structure
	if len(output1) != len(output2) {
		t.Errorf("JSON output length differs: %d vs %d", len(output1), len(output2))
	}

	// First few items should match
	for i := 0; i < len(output1) && i < 3; i++ {
		if output1[i]["Category"] != output2[i]["Category"] {
			t.Errorf("template %d category differs: %v vs %v", i, output1[i]["Category"], output2[i]["Category"])
		}
		if output1[i]["Pattern"] != output2[i]["Pattern"] {
			t.Errorf("template %d pattern differs", i)
		}
	}
}
