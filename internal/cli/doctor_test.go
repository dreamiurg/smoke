package cli

import (
	"strings"
	"testing"
)

func TestColor(t *testing.T) {
	// Save and restore useColor
	origUseColor := useColor
	defer func() { useColor = origUseColor }()

	tests := []struct {
		name     string
		useColor bool
		colorArg string
		text     string
		want     string
	}{
		{
			name:     "color enabled",
			useColor: true,
			colorArg: colorGreen,
			text:     "test",
			want:     colorGreen + "test" + colorReset,
		},
		{
			name:     "color disabled",
			useColor: false,
			colorArg: colorGreen,
			text:     "test",
			want:     "test",
		},
		{
			name:     "red color",
			useColor: true,
			colorArg: colorRed,
			text:     "error",
			want:     colorRed + "error" + colorReset,
		},
		{
			name:     "empty string",
			useColor: true,
			colorArg: colorCyan,
			text:     "",
			want:     colorCyan + "" + colorReset,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useColor = tt.useColor
			got := color(tt.colorArg, tt.text)
			if got != tt.want {
				t.Errorf("color() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatCheck(t *testing.T) {
	// Save and restore useColor
	origUseColor := useColor
	defer func() { useColor = origUseColor }()
	useColor = false // Disable colors for predictable output

	tests := []struct {
		name  string
		check Check
		want  string
	}{
		{
			name: "pass status",
			check: Check{
				Name:    "Test Check",
				Status:  StatusPass,
				Message: "all good",
			},
			want: "  ✓  Test Check all good",
		},
		{
			name: "warn status",
			check: Check{
				Name:    "Warning Check",
				Status:  StatusWarn,
				Message: "something off",
			},
			want: "  ⚠  Warning Check something off",
		},
		{
			name: "fail status",
			check: Check{
				Name:    "Failed Check",
				Status:  StatusFail,
				Message: "broken",
			},
			want: "  ✗  Failed Check broken",
		},
		{
			name: "with detail",
			check: Check{
				Name:    "Detail Check",
				Status:  StatusPass,
				Message: "ok",
				Detail:  "extra info",
			},
			want: "  ✓  Detail Check ok\n     └─ extra info",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatCheck(tt.check)
			if got != tt.want {
				t.Errorf("formatCheck() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatCategory(t *testing.T) {
	// Save and restore useColor
	origUseColor := useColor
	defer func() { useColor = origUseColor }()
	useColor = false

	cat := Category{
		Name: "TEST CATEGORY",
		Checks: []Check{
			{Name: "Check 1", Status: StatusPass, Message: "ok"},
			{Name: "Check 2", Status: StatusWarn, Message: "warning"},
		},
	}

	got := formatCategory(cat)

	if !strings.HasPrefix(got, "TEST CATEGORY\n") {
		t.Errorf("formatCategory() should start with category name, got %q", got)
	}
	if !strings.Contains(got, "Check 1") {
		t.Error("formatCategory() should contain Check 1")
	}
	if !strings.Contains(got, "Check 2") {
		t.Error("formatCategory() should contain Check 2")
	}
}

func TestComputeExitCode(t *testing.T) {
	tests := []struct {
		name       string
		categories []Category
		want       int
	}{
		{
			name: "all pass",
			categories: []Category{
				{Checks: []Check{{Status: StatusPass}, {Status: StatusPass}}},
			},
			want: 0,
		},
		{
			name: "has warning",
			categories: []Category{
				{Checks: []Check{{Status: StatusPass}, {Status: StatusWarn}}},
			},
			want: 1,
		},
		{
			name: "has error",
			categories: []Category{
				{Checks: []Check{{Status: StatusPass}, {Status: StatusFail}}},
			},
			want: 2,
		},
		{
			name: "error takes precedence over warning",
			categories: []Category{
				{Checks: []Check{{Status: StatusWarn}, {Status: StatusFail}}},
			},
			want: 2,
		},
		{
			name:       "empty categories",
			categories: []Category{},
			want:       0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := computeExitCode(tt.categories)
			if got != tt.want {
				t.Errorf("computeExitCode() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestCheckVersion(t *testing.T) {
	check := checkVersion()

	if check.Name != "Smoke Version" {
		t.Errorf("checkVersion().Name = %q, want %q", check.Name, "Smoke Version")
	}
	if check.Status != StatusPass {
		t.Errorf("checkVersion().Status = %v, want StatusPass", check.Status)
	}
	if check.CanFix {
		t.Error("checkVersion().CanFix should be false")
	}
}

func TestColorConstants(t *testing.T) {
	// Verify color codes are ANSI escape sequences
	colors := map[string]string{
		"reset":  colorReset,
		"red":    colorRed,
		"green":  colorGreen,
		"yellow": colorYellow,
		"cyan":   colorCyan,
		"dim":    colorDim,
	}

	for name, code := range colors {
		if !strings.HasPrefix(code, "\033[") {
			t.Errorf("%s color should start with ANSI escape, got %q", name, code)
		}
	}
}
