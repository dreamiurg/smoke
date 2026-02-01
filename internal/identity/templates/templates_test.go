package templates

import (
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
