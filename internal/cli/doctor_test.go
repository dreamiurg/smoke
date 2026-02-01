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
	check := performVersionCheck()

	if check.Name != "Smoke Version" {
		t.Errorf("performVersionCheck().Name = %q, want %q", check.Name, "Smoke Version")
	}
	if check.Status != StatusPass {
		t.Errorf("performVersionCheck().Status = %v, want StatusPass", check.Status)
	}
	if check.CanFix {
		t.Error("performVersionCheck().CanFix should be false")
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

func TestPassCheck(t *testing.T) {
	check := passCheck("Test Name", "test message")

	if check.Name != "Test Name" {
		t.Errorf("passCheck().Name = %q, want %q", check.Name, "Test Name")
	}
	if check.Status != StatusPass {
		t.Errorf("passCheck().Status = %v, want StatusPass", check.Status)
	}
	if check.Message != "test message" {
		t.Errorf("passCheck().Message = %q, want %q", check.Message, "test message")
	}
	if check.Detail != "" {
		t.Errorf("passCheck().Detail should be empty, got %q", check.Detail)
	}
	if check.CanFix {
		t.Error("passCheck().CanFix should be false")
	}
	if check.Fix != nil {
		t.Error("passCheck().Fix should be nil")
	}
}

func TestWarnCheck(t *testing.T) {
	check := warnCheck("Warning Name", "warning message", "extra details")

	if check.Name != "Warning Name" {
		t.Errorf("warnCheck().Name = %q, want %q", check.Name, "Warning Name")
	}
	if check.Status != StatusWarn {
		t.Errorf("warnCheck().Status = %v, want StatusWarn", check.Status)
	}
	if check.Message != "warning message" {
		t.Errorf("warnCheck().Message = %q, want %q", check.Message, "warning message")
	}
	if check.Detail != "extra details" {
		t.Errorf("warnCheck().Detail = %q, want %q", check.Detail, "extra details")
	}
	if check.CanFix {
		t.Error("warnCheck().CanFix should be false")
	}
	if check.Fix != nil {
		t.Error("warnCheck().Fix should be nil")
	}
}

func TestFailCheck(t *testing.T) {
	fixCalled := false
	fixFunc := func() error {
		fixCalled = true
		return nil
	}

	tests := []struct {
		name    string
		canFix  bool
		fixFunc func() error
	}{
		{"with fix", true, fixFunc},
		{"without fix", false, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			check := failCheck("Fail Name", "fail message", "fail details", tt.canFix, tt.fixFunc)

			if check.Name != "Fail Name" {
				t.Errorf("failCheck().Name = %q, want %q", check.Name, "Fail Name")
			}
			if check.Status != StatusFail {
				t.Errorf("failCheck().Status = %v, want StatusFail", check.Status)
			}
			if check.Message != "fail message" {
				t.Errorf("failCheck().Message = %q, want %q", check.Message, "fail message")
			}
			if check.Detail != "fail details" {
				t.Errorf("failCheck().Detail = %q, want %q", check.Detail, "fail details")
			}
			if check.CanFix != tt.canFix {
				t.Errorf("failCheck().CanFix = %v, want %v", check.CanFix, tt.canFix)
			}
			if tt.canFix {
				if check.Fix == nil {
					t.Error("failCheck().Fix should not be nil when canFix is true")
				} else {
					check.Fix()
					if !fixCalled {
						t.Error("failCheck().Fix should call provided function")
					}
				}
			} else {
				if check.Fix != nil {
					t.Error("failCheck().Fix should be nil when canFix is false")
				}
			}
		})
	}
}
