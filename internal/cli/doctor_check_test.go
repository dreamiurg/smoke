package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestCheckConfigDir(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(t *testing.T) string // Returns temp dir
		wantStatus   CheckStatus
		wantCanFix   bool
		wantContains string // Substring to check in Message or Detail
	}{
		{
			name: "directory exists and writable",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				configDir := filepath.Join(tmpDir, ".config", "smoke")
				if err := os.MkdirAll(configDir, 0755); err != nil {
					t.Fatal(err)
				}
				return tmpDir
			},
			wantStatus: StatusPass,
			wantCanFix: false,
		},
		{
			name: "directory does not exist",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				// Don't create the directory
				return tmpDir
			},
			wantStatus:   StatusFail,
			wantCanFix:   true,
			wantContains: "not found",
		},
		{
			name: "path exists but is a file not directory",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, ".config", "smoke")
				os.MkdirAll(filepath.Dir(configPath), 0755)
				// Create a file instead of directory
				if err := os.WriteFile(configPath, []byte("test"), 0644); err != nil {
					t.Fatal(err)
				}
				return tmpDir
			},
			wantStatus:   StatusFail,
			wantCanFix:   false,
			wantContains: "not a directory",
		},
		{
			name: "directory not writable",
			setup: func(t *testing.T) string {
				if os.Getuid() == 0 {
					t.Skip("skipping permission test when running as root")
				}
				tmpDir := t.TempDir()
				configDir := filepath.Join(tmpDir, ".config", "smoke")
				// Create the directory first with normal permissions
				if err := os.MkdirAll(configDir, 0755); err != nil {
					t.Fatal(err)
				}
				// Then change permissions to read-only
				if err := os.Chmod(configDir, 0555); err != nil {
					t.Fatal(err)
				}
				return tmpDir
			},
			wantStatus:   StatusWarn,
			wantCanFix:   false,
			wantContains: "not writable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := tt.setup(t)

			// Set HOME to temp dir so config functions use it
			t.Setenv("HOME", tmpDir)

			check := performConfigDirCheck()

			if check.Status != tt.wantStatus {
				t.Errorf("performConfigDirCheck().Status = %v, want %v", check.Status, tt.wantStatus)
			}
			if check.CanFix != tt.wantCanFix {
				t.Errorf("performConfigDirCheck().CanFix = %v, want %v", check.CanFix, tt.wantCanFix)
			}
			if tt.wantContains != "" {
				combined := check.Message + " " + check.Detail
				if !strings.Contains(combined, tt.wantContains) {
					t.Errorf("performConfigDirCheck() message/detail should contain %q, got Message=%q Detail=%q",
						tt.wantContains, check.Message, check.Detail)
				}
			}
		})
	}
}

func TestCheckFeedFile(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(t *testing.T) string
		wantStatus   CheckStatus
		wantCanFix   bool
		wantContains string
	}{
		{
			name: "feed file exists and readable",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				configDir := filepath.Join(tmpDir, ".config", "smoke")
				os.MkdirAll(configDir, 0755)
				feedPath := filepath.Join(configDir, "feed.jsonl")
				if err := os.WriteFile(feedPath, []byte(""), 0644); err != nil {
					t.Fatal(err)
				}
				return tmpDir
			},
			wantStatus: StatusPass,
			wantCanFix: false,
		},
		{
			name: "feed file does not exist",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				configDir := filepath.Join(tmpDir, ".config", "smoke")
				os.MkdirAll(configDir, 0755)
				return tmpDir
			},
			wantStatus:   StatusFail,
			wantCanFix:   true,
			wantContains: "not found",
		},
		{
			name: "path is a directory not a file",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				configDir := filepath.Join(tmpDir, ".config", "smoke")
				feedPath := filepath.Join(configDir, "feed.jsonl")
				// Create directory instead of file
				if err := os.MkdirAll(feedPath, 0755); err != nil {
					t.Fatal(err)
				}
				return tmpDir
			},
			wantStatus:   StatusFail,
			wantCanFix:   false,
			wantContains: "is a directory",
		},
		{
			name: "feed file not readable",
			setup: func(t *testing.T) string {
				if os.Getuid() == 0 {
					t.Skip("skipping permission test when running as root")
				}
				tmpDir := t.TempDir()
				configDir := filepath.Join(tmpDir, ".config", "smoke")
				os.MkdirAll(configDir, 0755)
				feedPath := filepath.Join(configDir, "feed.jsonl")
				if err := os.WriteFile(feedPath, []byte(""), 0000); err != nil {
					t.Fatal(err)
				}
				return tmpDir
			},
			wantStatus:   StatusWarn,
			wantCanFix:   false,
			wantContains: "not readable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := tt.setup(t)

			t.Setenv("HOME", tmpDir)

			check := performFeedFileCheck()

			if check.Status != tt.wantStatus {
				t.Errorf("performFeedFileCheck().Status = %v, want %v", check.Status, tt.wantStatus)
			}
			if check.CanFix != tt.wantCanFix {
				t.Errorf("performFeedFileCheck().CanFix = %v, want %v", check.CanFix, tt.wantCanFix)
			}
			if tt.wantContains != "" {
				combined := check.Message + " " + check.Detail
				if !strings.Contains(combined, tt.wantContains) {
					t.Errorf("performFeedFileCheck() message/detail should contain %q, got Message=%q Detail=%q",
						tt.wantContains, check.Message, check.Detail)
				}
			}
		})
	}
}

func TestCheckFeedFormat(t *testing.T) {
	tests := []struct {
		name         string
		feedContent  string
		wantStatus   CheckStatus
		wantContains string
	}{
		{
			name:         "empty feed",
			feedContent:  "",
			wantStatus:   StatusPass,
			wantContains: "empty (0 posts)",
		},
		{
			name:         "valid single line JSON",
			feedContent:  `{"id":"smk-1","text":"hello"}`,
			wantStatus:   StatusPass,
			wantContains: "1 post, valid",
		},
		{
			name: "valid multiple lines JSON",
			feedContent: `{"id":"smk-1","text":"hello"}
{"id":"smk-2","text":"world"}
{"id":"smk-3","text":"test"}`,
			wantStatus:   StatusPass,
			wantContains: "3 posts, all valid",
		},
		{
			name: "empty lines ignored",
			feedContent: `{"id":"smk-1","text":"hello"}

{"id":"smk-2","text":"world"}`,
			wantStatus:   StatusPass,
			wantContains: "2 posts, all valid",
		},
		{
			name: "some invalid JSON lines",
			feedContent: `{"id":"smk-1","text":"hello"}
invalid json line
{"id":"smk-2","text":"world"}`,
			wantStatus:   StatusWarn,
			wantContains: "2/3 lines valid",
		},
		{
			name:         "all invalid JSON",
			feedContent:  "not json at all\nalso not json",
			wantStatus:   StatusWarn,
			wantContains: "0/2 lines valid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configDir := filepath.Join(tmpDir, ".config", "smoke")
			os.MkdirAll(configDir, 0755)
			feedPath := filepath.Join(configDir, "feed.jsonl")
			if err := os.WriteFile(feedPath, []byte(tt.feedContent), 0644); err != nil {
				t.Fatal(err)
			}

			t.Setenv("HOME", tmpDir)

			check := performFeedFormatCheck()

			if check.Status != tt.wantStatus {
				t.Errorf("performFeedFormatCheck().Status = %v, want %v", check.Status, tt.wantStatus)
			}
			if !strings.Contains(check.Message, tt.wantContains) {
				t.Errorf("performFeedFormatCheck().Message should contain %q, got %q",
					tt.wantContains, check.Message)
			}
		})
	}
}

func TestCheckFeedFormat_FileErrors(t *testing.T) {
	t.Run("feed file does not exist", func(t *testing.T) {
		tmpDir := t.TempDir()
		configDir := filepath.Join(tmpDir, ".config", "smoke")
		os.MkdirAll(configDir, 0755)

		t.Setenv("HOME", tmpDir)

		check := performFeedFormatCheck()

		if check.Status != StatusFail {
			t.Errorf("performFeedFormatCheck().Status = %v, want StatusFail", check.Status)
		}
		if !strings.Contains(check.Message, "cannot open") {
			t.Errorf("performFeedFormatCheck().Message should mention cannot open, got %q", check.Message)
		}
	})
}

func TestCheckConfigFile(t *testing.T) {
	tests := []struct {
		name         string
		configData   string
		fileExists   bool
		wantStatus   CheckStatus
		wantCanFix   bool
		wantContains string
	}{
		{
			name:       "valid YAML config",
			configData: "# Smoke configuration\nkey: value\n",
			fileExists: true,
			wantStatus: StatusPass,
			wantCanFix: false,
		},
		{
			name:       "empty config file",
			configData: "",
			fileExists: true,
			wantStatus: StatusPass,
			wantCanFix: false,
		},
		{
			name:         "config file does not exist",
			fileExists:   false,
			wantStatus:   StatusWarn,
			wantCanFix:   true,
			wantContains: "missing (using defaults)",
		},
		{
			name:         "invalid YAML",
			configData:   "invalid: yaml: content:\n  - bad indentation",
			fileExists:   true,
			wantStatus:   StatusFail,
			wantCanFix:   false,
			wantContains: "invalid YAML",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configDir := filepath.Join(tmpDir, ".config", "smoke")
			os.MkdirAll(configDir, 0755)

			if tt.fileExists {
				configPath := filepath.Join(configDir, "config.yaml")
				if err := os.WriteFile(configPath, []byte(tt.configData), 0644); err != nil {
					t.Fatal(err)
				}
			}

			t.Setenv("HOME", tmpDir)

			check := performConfigFileCheck()

			if check.Status != tt.wantStatus {
				t.Errorf("performConfigFileCheck().Status = %v, want %v", check.Status, tt.wantStatus)
			}
			if check.CanFix != tt.wantCanFix {
				t.Errorf("performConfigFileCheck().CanFix = %v, want %v", check.CanFix, tt.wantCanFix)
			}
			if tt.wantContains != "" {
				combined := check.Message + " " + check.Detail
				if !strings.Contains(combined, tt.wantContains) {
					t.Errorf("performConfigFileCheck() message/detail should contain %q, got Message=%q Detail=%q",
						tt.wantContains, check.Message, check.Detail)
				}
			}
		})
	}
}

func TestFixConfigDir(t *testing.T) {
	tmpDir := t.TempDir()

	t.Setenv("HOME", tmpDir)

	// Directory should not exist initially
	configDir := filepath.Join(tmpDir, ".config", "smoke")
	if _, err := os.Stat(configDir); !os.IsNotExist(err) {
		t.Fatal("config dir should not exist initially")
	}

	// Fix should create it
	result, err := fixConfigDir()
	if err != nil {
		t.Fatalf("fixConfigDir() error = %v", err)
	}
	if result == nil {
		t.Fatal("fixConfigDir() should return non-nil FixResult")
	}
	if result.Description == "" {
		t.Error("fixConfigDir() FixResult should have a description")
	}

	// Verify it was created
	info, err := os.Stat(configDir)
	if err != nil {
		t.Fatalf("config dir should exist after fix, got error: %v", err)
	}
	if !info.IsDir() {
		t.Error("fixConfigDir() should create a directory")
	}
}

func TestFixFeedFile(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".config", "smoke")
	os.MkdirAll(configDir, 0755)

	t.Setenv("HOME", tmpDir)

	feedPath := filepath.Join(configDir, "feed.jsonl")

	// File should not exist initially
	if _, err := os.Stat(feedPath); !os.IsNotExist(err) {
		t.Fatal("feed file should not exist initially")
	}

	// Fix should create it
	result, err := fixFeedFile()
	if err != nil {
		t.Fatalf("fixFeedFile() error = %v", err)
	}
	if result == nil {
		t.Fatal("fixFeedFile() should return non-nil FixResult")
	}
	if result.Description == "" {
		t.Error("fixFeedFile() FixResult should have a description")
	}

	// Verify it was created
	info, err := os.Stat(feedPath)
	if err != nil {
		t.Fatalf("feed file should exist after fix, got error: %v", err)
	}
	if info.IsDir() {
		t.Error("fixFeedFile() should create a file, not a directory")
	}
}

func TestFixConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".config", "smoke")
	os.MkdirAll(configDir, 0755)

	t.Setenv("HOME", tmpDir)

	configPath := filepath.Join(configDir, "config.yaml")

	// File should not exist initially
	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		t.Fatal("config file should not exist initially")
	}

	// Fix should create it
	result, err := fixConfigFile()
	if err != nil {
		t.Fatalf("fixConfigFile() error = %v", err)
	}
	if result == nil {
		t.Fatal("fixConfigFile() should return non-nil FixResult")
	}
	if result.Description == "" {
		t.Error("fixConfigFile() FixResult should have a description")
	}

	// Verify it was created with valid content
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("config file should exist after fix, got error: %v", err)
	}

	// Should be valid YAML
	var parsed interface{}
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		t.Errorf("fixConfigFile() should create valid YAML, got error: %v", err)
	}

	// Should contain comment
	if !strings.Contains(string(data), "Smoke configuration") {
		t.Error("fixConfigFile() should include configuration comment")
	}
}

func TestApplyFixes(t *testing.T) {
	tests := []struct {
		name          string
		categories    []Category
		dryRun        bool
		wantFixCount  int
		wantFixCalled bool
	}{
		{
			name: "no fixes needed",
			categories: []Category{
				{Checks: []Check{
					{Status: StatusPass, CanFix: false},
				}},
			},
			dryRun:       false,
			wantFixCount: 0,
		},
		{
			name: "one fix needed",
			categories: []Category{
				{Checks: []Check{
					{Status: StatusFail, CanFix: true, Fix: func() (*FixResult, error) { return &FixResult{Description: "test"}, nil }},
				}},
			},
			dryRun:        false,
			wantFixCount:  1,
			wantFixCalled: true,
		},
		{
			name: "multiple fixes needed",
			categories: []Category{
				{Checks: []Check{
					{Status: StatusFail, CanFix: true, Fix: func() (*FixResult, error) { return &FixResult{Description: "test"}, nil }},
					{Status: StatusWarn, CanFix: true, Fix: func() (*FixResult, error) { return &FixResult{Description: "test"}, nil }},
				}},
			},
			dryRun:        false,
			wantFixCount:  2,
			wantFixCalled: true,
		},
		{
			name: "dry run mode",
			categories: []Category{
				{Checks: []Check{
					{Status: StatusFail, CanFix: true, Fix: func() (*FixResult, error) {
						t.Error("Fix should not be called in dry-run mode")
						return nil, nil
					}},
				}},
			},
			dryRun:       true,
			wantFixCount: 1,
		},
		{
			name: "skip non-fixable checks",
			categories: []Category{
				{Checks: []Check{
					{Status: StatusFail, CanFix: false},
					{Status: StatusFail, CanFix: true, Fix: func() (*FixResult, error) { return &FixResult{Description: "test"}, nil }},
				}},
			},
			dryRun:       false,
			wantFixCount: 1,
		},
		{
			name: "skip passing checks even if fixable",
			categories: []Category{
				{Checks: []Check{
					{Status: StatusPass, CanFix: true, Fix: func() (*FixResult, error) {
						t.Error("Fix should not be called for passing checks")
						return nil, nil
					}},
				}},
			},
			dryRun:       false,
			wantFixCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			tmpR, tmpW, _ := os.Pipe()
			os.Stdout = tmpW
			defer func() { os.Stdout = oldStdout }()

			fixCount, err := applyFixes(tt.categories, tt.dryRun)

			tmpW.Close()
			os.Stdout = oldStdout

			// Consume output
			var buf bytes.Buffer
			buf.ReadFrom(tmpR)

			if err != nil {
				t.Errorf("applyFixes() error = %v", err)
			}
			if fixCount != tt.wantFixCount {
				t.Errorf("applyFixes() fixCount = %d, want %d", fixCount, tt.wantFixCount)
			}
		})
	}
}

func TestApplyFixes_ErrorHandling(t *testing.T) {
	categories := []Category{
		{Checks: []Check{
			{Status: StatusFail, CanFix: true, Name: "Bad Check", Fix: func() (*FixResult, error) {
				return nil, os.ErrPermission
			}},
		}},
	}

	// Capture stdout to verify error message
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	fixCount, err := applyFixes(categories, false)

	w.Close()
	os.Stdout = oldStdout

	// Read captured output
	buf := new(bytes.Buffer)
	buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Errorf("applyFixes() should not return error for individual fix failures, got %v", err)
	}
	if fixCount != 0 {
		t.Errorf("applyFixes() fixCount = %d, want 0 (failed fixes don't count)", fixCount)
	}
	if !strings.Contains(output, "Failed to fix") {
		t.Error("applyFixes() should print error message for failed fixes")
	}
}

func TestRunChecks(t *testing.T) {
	categories := runChecks()

	if len(categories) == 0 {
		t.Fatal("runChecks() should return categories")
	}

	// Verify expected categories exist
	categoryNames := make(map[string]bool)
	for _, cat := range categories {
		categoryNames[cat.Name] = true
		if len(cat.Checks) == 0 {
			t.Errorf("Category %s should have checks", cat.Name)
		}
	}

	expectedCategories := []string{"INSTALLATION", "DATA", "VERSION"}
	for _, expected := range expectedCategories {
		if !categoryNames[expected] {
			t.Errorf("runChecks() missing expected category: %s", expected)
		}
	}
}

func TestPrintReport(t *testing.T) {
	// Save original useColor and disable for predictable output
	origUseColor := useColor
	useColor = false
	defer func() { useColor = origUseColor }()

	categories := []Category{
		{
			Name: "TEST CATEGORY",
			Checks: []Check{
				{Name: "Check 1", Status: StatusPass, Message: "ok"},
				{Name: "Check 2", Status: StatusFail, Message: "failed"},
			},
		},
	}

	// Capture output using a buffer
	var buf bytes.Buffer

	// Temporarily redirect stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printReport(categories)

	w.Close()
	os.Stdout = oldStdout

	buf.ReadFrom(r)
	output := buf.String()

	// Verify output contains expected elements
	if !strings.Contains(output, "smoke doctor") {
		t.Error("printReport() should include header with version")
	}
	if !strings.Contains(output, "TEST CATEGORY") {
		t.Error("printReport() should include category name")
	}
	if !strings.Contains(output, "Check 1") {
		t.Error("printReport() should include check names")
	}
	if !strings.Contains(output, "Check 2") {
		t.Error("printReport() should include all checks")
	}
}

func TestPrintReport_CustomWriter(t *testing.T) {
	// This test demonstrates how printReport could be tested with a custom writer
	// if it accepted an io.Writer parameter (which it currently doesn't).
	// For now, this serves as documentation for future refactoring.

	origUseColor := useColor
	useColor = false
	defer func() { useColor = origUseColor }()

	categories := []Category{
		{
			Name: "CATEGORY",
			Checks: []Check{
				{Name: "Test", Status: StatusPass, Message: "msg"},
			},
		},
	}

	// Current implementation writes to stdout, so we capture it
	oldStdout := os.Stdout
	tmpR, tmpW, _ := os.Pipe()
	os.Stdout = tmpW

	printReport(categories)

	tmpW.Close()
	os.Stdout = oldStdout

	// Read and discard output
	var buf bytes.Buffer
	buf.ReadFrom(tmpR)

	// This test just ensures printReport doesn't panic and produces output
	if buf.Len() == 0 {
		t.Error("printReport() should produce output")
	}
}

// Additional edge case tests
func TestCheckConfigDir_HomeError(t *testing.T) {
	// Test when HOME directory cannot be determined
	// This is hard to test without actually breaking the environment
	// but we can verify the check handles config.GetConfigDir errors
	t.Skip("Skipping test that would require breaking HOME env var")
}

func TestCheckFeedFormat_LargeFile(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".config", "smoke")
	os.MkdirAll(configDir, 0755)
	feedPath := filepath.Join(configDir, "feed.jsonl")

	// Create a file with many valid JSON lines
	f, err := os.Create(feedPath)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 1000; i++ {
		post := map[string]interface{}{
			"id":   "smk-" + string(rune(i)),
			"text": "test message",
		}
		data, _ := json.Marshal(post)
		f.Write(data)
		f.Write([]byte("\n"))
	}
	f.Close()

	t.Setenv("HOME", tmpDir)

	check := performFeedFormatCheck()

	if check.Status != StatusPass {
		t.Errorf("performFeedFormatCheck() with large file Status = %v, want StatusPass", check.Status)
	}
	if !strings.Contains(check.Message, "1000 posts") {
		t.Errorf("performFeedFormatCheck().Message should mention 1000 posts, got %q", check.Message)
	}
}

func TestFixConfigDir_Error(t *testing.T) {
	// Test error path when HOME is not set
	t.Setenv("HOME", "")
	t.Setenv("USERPROFILE", "")

	_, err := fixConfigDir()
	if err == nil {
		t.Error("fixConfigDir() should return error when HOME is not set")
	}
}

func TestFixFeedFile_Error(t *testing.T) {
	// Test error path when HOME is not set
	t.Setenv("HOME", "")
	t.Setenv("USERPROFILE", "")

	_, err := fixFeedFile()
	if err == nil {
		t.Error("fixFeedFile() should return error when HOME is not set")
	}
}

func TestFixConfigFile_Error(t *testing.T) {
	// Test error path when HOME is not set
	t.Setenv("HOME", "")
	t.Setenv("USERPROFILE", "")

	_, err := fixConfigFile()
	if err == nil {
		t.Error("fixConfigFile() should return error when HOME is not set")
	}
}

func TestCheckFeedFile_CustomPath(t *testing.T) {
	tmpDir := t.TempDir()
	customFeed := filepath.Join(tmpDir, "custom-feed.jsonl")

	// Create custom feed file
	if err := os.WriteFile(customFeed, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	// Set HOME to temp dir so path validation passes
	t.Setenv("HOME", tmpDir)

	// Set SMOKE_FEED environment variable
	oldFeed := os.Getenv("SMOKE_FEED")
	os.Setenv("SMOKE_FEED", customFeed)
	defer func() {
		if oldFeed != "" {
			os.Setenv("SMOKE_FEED", oldFeed)
		} else {
			os.Unsetenv("SMOKE_FEED")
		}
	}()

	check := performFeedFileCheck()

	if check.Status != StatusPass {
		t.Errorf("performFeedFileCheck() with custom path Status = %v, want StatusPass", check.Status)
	}
	if !strings.Contains(check.Message, customFeed) {
		t.Errorf("performFeedFileCheck().Message should contain custom path %q, got %q", customFeed, check.Message)
	}
}

func TestFixTUIConfigStyleToLayout_BackupCreation(t *testing.T) {
	tmpDir := t.TempDir()
	tuiPath := filepath.Join(tmpDir, "tui.yaml")

	// Create a tui.yaml file with deprecated "style" field
	originalContent := "style: compact\nother: value\n"
	if err := os.WriteFile(tuiPath, []byte(originalContent), 0600); err != nil {
		t.Fatalf("Failed to create tui.yaml: %v", err)
	}

	// Call fixTUIConfigStyleToLayout
	result, err := fixTUIConfigStyleToLayout(tuiPath)
	if err != nil {
		t.Fatalf("fixTUIConfigStyleToLayout() returned error: %v", err)
	}

	// Verify backup was created
	if result == nil {
		t.Fatal("fixTUIConfigStyleToLayout() returned nil result")
	}
	if result.BackupPath == "" {
		t.Error("fixTUIConfigStyleToLayout() BackupPath should not be empty")
	}

	// Verify backup file exists
	if _, statErr := os.Stat(result.BackupPath); os.IsNotExist(statErr) {
		t.Errorf("Backup file does not exist at %q", result.BackupPath)
	}

	// Verify backup contains original content
	backupContent, readErr := os.ReadFile(result.BackupPath)
	if readErr != nil {
		t.Fatalf("Failed to read backup file: %v", readErr)
	}
	if string(backupContent) != originalContent {
		t.Errorf("Backup content mismatch.\nWant: %q\nGot: %q", originalContent, string(backupContent))
	}

	// Verify backup filename includes ".bak." pattern
	if !strings.Contains(result.BackupPath, ".bak.") {
		t.Errorf("Backup path should contain '.bak.' pattern, got %q", result.BackupPath)
	}

	// Verify description is set
	if result.Description == "" {
		t.Error("fixTUIConfigStyleToLayout() Description should not be empty")
	}
}

func TestFixTUIConfigStyleToLayout_MigrationContent(t *testing.T) {
	tmpDir := t.TempDir()
	tuiPath := filepath.Join(tmpDir, "tui.yaml")

	// Create a tui.yaml file with deprecated "style" field
	originalContent := "style: compact\nother: value\n"
	if err := os.WriteFile(tuiPath, []byte(originalContent), 0600); err != nil {
		t.Fatalf("Failed to create tui.yaml: %v", err)
	}

	// Call fixTUIConfigStyleToLayout
	result, err := fixTUIConfigStyleToLayout(tuiPath)
	if err != nil {
		t.Fatalf("fixTUIConfigStyleToLayout() returned error: %v", err)
	}

	// Verify the original file was migrated (style -> layout)
	newContent, err := os.ReadFile(tuiPath)
	if err != nil {
		t.Fatalf("Failed to read migrated tui.yaml: %v", err)
	}

	// Parse to verify structure
	var parsed map[string]interface{}
	if err := yaml.Unmarshal(newContent, &parsed); err != nil {
		t.Fatalf("Migrated content is not valid YAML: %v", err)
	}

	// Verify "style" field is gone and "layout" exists
	if _, hasStyle := parsed["style"]; hasStyle {
		t.Error("Migrated file should not contain 'style' field")
	}
	if layout, hasLayout := parsed["layout"]; !hasLayout {
		t.Error("Migrated file should contain 'layout' field")
	} else if layout != "compact" {
		t.Errorf("'layout' field should have value 'compact', got %v", layout)
	}

	// Verify description mentions the migration
	if !strings.Contains(result.Description, "style") || !strings.Contains(result.Description, "layout") {
		t.Errorf("Description should mention style->layout migration, got %q", result.Description)
	}
}

func TestApplyFixes_PrintsBackupPath(t *testing.T) {
	backupPath := "/tmp/test.bak.2024-01-01T12-00-00"
	categories := []Category{
		{
			Name: "Test",
			Checks: []Check{
				{
					Name:    "Test Check",
					Status:  StatusFail,
					CanFix:  true,
					Message: "Test message",
					Fix: func() (*FixResult, error) {
						return &FixResult{
							BackupPath:  backupPath,
							Description: "test fix",
						}, nil
					},
				},
			},
		},
	}

	// Capture stdout
	oldStdout := os.Stdout
	tmpR, tmpW, _ := os.Pipe()
	os.Stdout = tmpW

	fixCount, err := applyFixes(categories, false)

	tmpW.Close()
	os.Stdout = oldStdout

	// Read captured output
	var buf bytes.Buffer
	buf.ReadFrom(tmpR)
	output := buf.String()

	if err != nil {
		t.Fatalf("applyFixes() returned error: %v", err)
	}
	if fixCount != 1 {
		t.Errorf("applyFixes() fixCount = %d, want 1", fixCount)
	}

	// Verify backup path is printed
	if !strings.Contains(output, "Backed up to:") {
		t.Error("applyFixes() should print 'Backed up to:' when BackupPath is set")
	}
	if !strings.Contains(output, backupPath) {
		t.Errorf("applyFixes() should print the backup path %q, got output: %s", backupPath, output)
	}
}

func TestApplyFixes_PrintsDescriptionInParentheses(t *testing.T) {
	description := "migrated style to layout"
	categories := []Category{
		{
			Name: "Test",
			Checks: []Check{
				{
					Name:    "Test Check",
					Status:  StatusWarn,
					CanFix:  true,
					Message: "Test message",
					Fix: func() (*FixResult, error) {
						return &FixResult{
							Description: description,
						}, nil
					},
				},
			},
		},
	}

	// Capture stdout
	oldStdout := os.Stdout
	tmpR, tmpW, _ := os.Pipe()
	os.Stdout = tmpW

	fixCount, err := applyFixes(categories, false)

	tmpW.Close()
	os.Stdout = oldStdout

	// Read captured output
	var buf bytes.Buffer
	buf.ReadFrom(tmpR)
	output := buf.String()

	if err != nil {
		t.Fatalf("applyFixes() returned error: %v", err)
	}
	if fixCount != 1 {
		t.Errorf("applyFixes() fixCount = %d, want 1", fixCount)
	}

	// Verify description is printed in the correct format
	expectedFormat := "Fixed: Test Check (" + description + ")"
	if !strings.Contains(output, expectedFormat) {
		t.Errorf("applyFixes() should print format %q, got output: %s", expectedFormat, output)
	}
}

func TestApplyFixes_PrintsBackupPathAndDescription(t *testing.T) {
	backupPath := "/tmp/backup.bak.2024-01-01T12-00-00"
	description := "test description"
	categories := []Category{
		{
			Name: "Test",
			Checks: []Check{
				{
					Name:    "Test Check",
					Status:  StatusFail,
					CanFix:  true,
					Message: "Test message",
					Fix: func() (*FixResult, error) {
						return &FixResult{
							BackupPath:  backupPath,
							Description: description,
						}, nil
					},
				},
			},
		},
	}

	// Capture stdout
	oldStdout := os.Stdout
	tmpR, tmpW, _ := os.Pipe()
	os.Stdout = tmpW

	fixCount, err := applyFixes(categories, false)

	tmpW.Close()
	os.Stdout = oldStdout

	// Read captured output
	var buf bytes.Buffer
	buf.ReadFrom(tmpR)
	output := buf.String()

	if err != nil {
		t.Fatalf("applyFixes() returned error: %v", err)
	}
	if fixCount != 1 {
		t.Errorf("applyFixes() fixCount = %d, want 1", fixCount)
	}

	// Verify both backup path and description are printed
	if !strings.Contains(output, "Backed up to:") {
		t.Error("applyFixes() should print backup path")
	}
	if !strings.Contains(output, backupPath) {
		t.Errorf("applyFixes() should contain backup path %q", backupPath)
	}
	if !strings.Contains(output, "Fixed: Test Check ("+description+")") {
		t.Errorf("applyFixes() should print description in parentheses")
	}
}
