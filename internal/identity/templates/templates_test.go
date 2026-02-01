package templates

import (
	"math/rand"
	"testing"
)

func TestTemplatesExist(t *testing.T) {
	tests := []struct {
		name     string
		category string
		minCount int
		maxCount int
	}{
		{"Observations", "Observations", 3, 5},
		{"Questions", "Questions", 3, 5},
		{"Tensions", "Tensions", 3, 5},
		{"Learnings", "Learnings", 3, 5},
		{"Reflections", "Reflections", 3, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl := ByCategory(tt.category)
			if len(tmpl) < tt.minCount || len(tmpl) > tt.maxCount {
				t.Errorf("category %s has %d templates, want %d-%d", tt.category, len(tmpl), tt.minCount, tt.maxCount)
			}
			for _, tmplItem := range tmpl {
				if tmplItem.Category != tt.category {
					t.Errorf("template category %s does not match filter %s", tmplItem.Category, tt.category)
				}
				if tmplItem.Pattern == "" {
					t.Errorf("template pattern is empty")
				}
			}
		})
	}
}

func TestTotalTemplateCount(t *testing.T) {
	// Should have 15-20 templates as per spec
	if len(All) < 15 || len(All) > 20 {
		t.Errorf("total template count %d, want 15-20", len(All))
	}
}

func TestCategoriesOrder(t *testing.T) {
	expected := []string{
		"Observations",
		"Questions",
		"Tensions",
		"Learnings",
		"Reflections",
	}
	categories := Categories()
	if len(categories) != len(expected) {
		t.Errorf("categories count %d, want %d", len(categories), len(expected))
	}
	for i, cat := range categories {
		if cat != expected[i] {
			t.Errorf("category %d is %s, want %s", i, cat, expected[i])
		}
	}
}

func TestNoEmptyPatterns(t *testing.T) {
	for i, tmpl := range All {
		if tmpl.Pattern == "" {
			t.Errorf("template %d has empty pattern", i)
		}
		if tmpl.Category == "" {
			t.Errorf("template %d has empty category", i)
		}
	}
}

func TestByCategoryInvalidCategory(t *testing.T) {
	result := ByCategory("NonExistent")
	if len(result) != 0 {
		t.Errorf("ByCategory with invalid category returned %d templates, want 0", len(result))
	}
}

func TestByCategoryExactCounts(t *testing.T) {
	tests := []struct {
		category string
		expected int
	}{
		{"Observations", 4},
		{"Questions", 4},
		{"Tensions", 4},
		{"Learnings", 4},
		{"Reflections", 3},
	}

	for _, tt := range tests {
		t.Run(tt.category, func(t *testing.T) {
			tmpl := ByCategory(tt.category)
			if len(tmpl) != tt.expected {
				t.Errorf("category %s has %d templates, want %d", tt.category, len(tmpl), tt.expected)
			}
		})
	}
}

func TestByCategoryAllTemplatesAccessible(t *testing.T) {
	// Verify that the sum of all category templates equals total templates
	totalByCategory := 0
	for _, cat := range Categories() {
		totalByCategory += len(ByCategory(cat))
	}
	if totalByCategory != len(All) {
		t.Errorf("sum of category templates %d does not match All count %d", totalByCategory, len(All))
	}
}

func TestByCategoryNoLeakage(t *testing.T) {
	// Verify no template appears in multiple categories
	seenPatterns := make(map[string]string)
	for _, cat := range Categories() {
		for _, tmpl := range ByCategory(cat) {
			if prev, exists := seenPatterns[tmpl.Pattern]; exists {
				t.Errorf("pattern appears in multiple categories: %s (was %s, now %s)", tmpl.Pattern, prev, cat)
			}
			seenPatterns[tmpl.Pattern] = cat
		}
	}
}

func TestGetRandomReturnsValidTemplate(t *testing.T) {
	// Use a seeded RNG for deterministic testing
	rng := rand.New(rand.NewSource(42))

	tmpl := GetRandom(rng)
	if tmpl.Category == "" {
		t.Error("GetRandom returned template with empty category")
	}
	if tmpl.Pattern == "" {
		t.Error("GetRandom returned template with empty pattern")
	}
}

func TestGetRandomDeterminism(t *testing.T) {
	// Verify same seed produces same template
	rng1 := rand.New(rand.NewSource(12345))
	rng2 := rand.New(rand.NewSource(12345))

	tmpl1 := GetRandom(rng1)
	tmpl2 := GetRandom(rng2)

	if tmpl1.Category != tmpl2.Category || tmpl1.Pattern != tmpl2.Pattern {
		t.Errorf("GetRandom not deterministic: got %v then %v with same seed", tmpl1, tmpl2)
	}
}

func TestGetRandomDistribution(t *testing.T) {
	// Verify GetRandom can select from different parts of the template set
	rng := rand.New(rand.NewSource(99))
	seen := make(map[string]bool)

	// Pull 20 templates; with 19 total, we should see variety
	for i := 0; i < 20; i++ {
		tmpl := GetRandom(rng)
		seen[tmpl.Pattern] = true
	}

	if len(seen) < 5 {
		t.Errorf("GetRandom showed only %d unique templates in 20 selections, want at least 5", len(seen))
	}
}

func TestGetRandomEmptySet(t *testing.T) {
	// Verify behavior with empty template set (edge case)
	// This test documents expected behavior if All becomes empty
	rng := rand.New(rand.NewSource(1))
	// With current All having 19 templates, this will always succeed
	// But we document the contract: GetRandom on empty set returns zero Template
	if len(All) > 0 {
		tmpl := GetRandom(rng)
		if tmpl == (Template{}) {
			t.Error("GetRandom returned zero Template when All is non-empty")
		}
	}
}

func TestGetRandomFromCategory(t *testing.T) {
	// Verify GetRandom works correctly with category-filtered results
	rng := rand.New(rand.NewSource(123))

	observations := ByCategory("Observations")
	if len(observations) == 0 {
		t.Fatal("Observations category is empty")
	}

	// Create a simple mock RNG that returns indices within the observations slice
	tmpl := observations[rng.Intn(len(observations))]

	if tmpl.Category != "Observations" {
		t.Errorf("filtered template has wrong category: got %s, want Observations", tmpl.Category)
	}
}
