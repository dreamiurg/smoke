package config

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestGetTUIConfigPath(t *testing.T) {
	// Save and restore HOME env var
	origHome := os.Getenv("HOME")
	tmpHome := t.TempDir()
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	path, err := GetTUIConfigPath()
	if err != nil {
		t.Fatalf("GetTUIConfigPath() error: %v", err)
	}

	if path == "" {
		t.Error("GetTUIConfigPath() returned empty string")
	}

	if filepath.Base(path) != DefaultTUIConfigFile {
		t.Errorf("GetTUIConfigPath() should end with %s, got %s", DefaultTUIConfigFile, filepath.Base(path))
	}

	// Should be in ~/.config/smoke/
	if filepath.Base(filepath.Dir(path)) != "smoke" {
		t.Errorf("GetTUIConfigPath() parent should be 'smoke', got %s", filepath.Base(filepath.Dir(path)))
	}
}

func TestLoadTUIConfig_Default(t *testing.T) {
	// Save and restore HOME env var
	origHome := os.Getenv("HOME")
	tmpHome := t.TempDir()
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	// No config file exists yet
	cfg := LoadTUIConfig()

	if cfg == nil {
		t.Fatal("LoadTUIConfig() returned nil")
	}

	if cfg.Theme != DefaultTheme {
		t.Errorf("Expected default theme %q, got %q", DefaultTheme, cfg.Theme)
	}

	if cfg.Contrast != DefaultContrast {
		t.Errorf("Expected default contrast %q, got %q", DefaultContrast, cfg.Contrast)
	}
}

func TestLoadTUIConfig_ExistingFile(t *testing.T) {
	// Save and restore HOME env var
	origHome := os.Getenv("HOME")
	tmpHome := t.TempDir()
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	// Create config directory and file
	smokeDir := filepath.Join(tmpHome, ".config", "smoke")
	if err := os.MkdirAll(smokeDir, 0755); err != nil {
		t.Fatalf("Failed to create smoke dir: %v", err)
	}

	// Write a valid config file
	testCfg := &TUIConfig{
		Theme:    "monokai",
		Contrast: "high",
	}
	data, err := yaml.Marshal(testCfg)
	if err != nil {
		t.Fatalf("Failed to marshal YAML: %v", err)
	}

	configPath := filepath.Join(smokeDir, DefaultTUIConfigFile)
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Load and verify
	cfg := LoadTUIConfig()

	if cfg.Theme != "monokai" {
		t.Errorf("Expected theme 'monokai', got %q", cfg.Theme)
	}

	if cfg.Contrast != "high" {
		t.Errorf("Expected contrast 'high', got %q", cfg.Contrast)
	}
}

func TestLoadTUIConfig_EmptyFile(t *testing.T) {
	// Save and restore HOME env var
	origHome := os.Getenv("HOME")
	tmpHome := t.TempDir()
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	// Create empty config file
	smokeDir := filepath.Join(tmpHome, ".config", "smoke")
	if err := os.MkdirAll(smokeDir, 0755); err != nil {
		t.Fatalf("Failed to create smoke dir: %v", err)
	}

	configPath := filepath.Join(smokeDir, DefaultTUIConfigFile)
	if err := os.WriteFile(configPath, []byte{}, 0644); err != nil {
		t.Fatalf("Failed to write empty config file: %v", err)
	}

	// Load and verify defaults are returned
	cfg := LoadTUIConfig()

	if cfg.Theme != DefaultTheme {
		t.Errorf("Empty file should return default theme %q, got %q", DefaultTheme, cfg.Theme)
	}

	if cfg.Contrast != DefaultContrast {
		t.Errorf("Empty file should return default contrast %q, got %q", DefaultContrast, cfg.Contrast)
	}
}

func TestLoadTUIConfig_CorruptFile(t *testing.T) {
	// Save and restore HOME env var
	origHome := os.Getenv("HOME")
	tmpHome := t.TempDir()
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	// Create corrupt config file
	smokeDir := filepath.Join(tmpHome, ".config", "smoke")
	if err := os.MkdirAll(smokeDir, 0755); err != nil {
		t.Fatalf("Failed to create smoke dir: %v", err)
	}

	configPath := filepath.Join(smokeDir, DefaultTUIConfigFile)
	invalidYAML := "invalid: yaml: content: [["
	if err := os.WriteFile(configPath, []byte(invalidYAML), 0644); err != nil {
		t.Fatalf("Failed to write corrupt config file: %v", err)
	}

	// Load and verify defaults are returned instead of error
	cfg := LoadTUIConfig()

	if cfg == nil {
		t.Fatal("LoadTUIConfig() should return default config, not nil")
	}

	if cfg.Theme != DefaultTheme {
		t.Errorf("Corrupt file should return default theme %q, got %q", DefaultTheme, cfg.Theme)
	}

	if cfg.Contrast != DefaultContrast {
		t.Errorf("Corrupt file should return default contrast %q, got %q", DefaultContrast, cfg.Contrast)
	}
}

func TestLoadTUIConfig_PartialFields(t *testing.T) {
	// Save and restore HOME env var
	origHome := os.Getenv("HOME")
	tmpHome := t.TempDir()
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	// Create config directory and file with only theme field
	smokeDir := filepath.Join(tmpHome, ".config", "smoke")
	if err := os.MkdirAll(smokeDir, 0755); err != nil {
		t.Fatalf("Failed to create smoke dir: %v", err)
	}

	configPath := filepath.Join(smokeDir, DefaultTUIConfigFile)
	partialYAML := "theme: dracula\n"
	if err := os.WriteFile(configPath, []byte(partialYAML), 0644); err != nil {
		t.Fatalf("Failed to write partial config file: %v", err)
	}

	// Load and verify
	cfg := LoadTUIConfig()

	if cfg.Theme != "dracula" {
		t.Errorf("Expected theme 'dracula', got %q", cfg.Theme)
	}

	// Missing field should get default
	if cfg.Contrast != DefaultContrast {
		t.Errorf("Missing contrast should default to %q, got %q", DefaultContrast, cfg.Contrast)
	}
}

func TestSaveTUIConfig(t *testing.T) {
	// Save and restore HOME env var
	origHome := os.Getenv("HOME")
	tmpHome := t.TempDir()
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	// Create config directory
	smokeDir := filepath.Join(tmpHome, ".config", "smoke")
	if err := os.MkdirAll(smokeDir, 0755); err != nil {
		t.Fatalf("Failed to create smoke dir: %v", err)
	}

	// Save a config
	testCfg := &TUIConfig{
		Theme:    "solarized-dark",
		Contrast: "low",
	}

	err := SaveTUIConfig(testCfg)
	if err != nil {
		t.Fatalf("SaveTUIConfig() error: %v", err)
	}

	// Verify file exists and contains correct data
	configPath := filepath.Join(smokeDir, DefaultTUIConfigFile)
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read saved config file: %v", err)
	}

	var loadedCfg TUIConfig
	err = yaml.Unmarshal(data, &loadedCfg)
	if err != nil {
		t.Fatalf("Failed to unmarshal saved config: %v", err)
	}

	if loadedCfg.Theme != "solarized-dark" {
		t.Errorf("Saved theme should be 'solarized-dark', got %q", loadedCfg.Theme)
	}

	if loadedCfg.Contrast != "low" {
		t.Errorf("Saved contrast should be 'low', got %q", loadedCfg.Contrast)
	}
}

func TestSaveTUIConfig_Overwrites(t *testing.T) {
	// Save and restore HOME env var
	origHome := os.Getenv("HOME")
	tmpHome := t.TempDir()
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	// Create config directory
	smokeDir := filepath.Join(tmpHome, ".config", "smoke")
	if err := os.MkdirAll(smokeDir, 0755); err != nil {
		t.Fatalf("Failed to create smoke dir: %v", err)
	}

	// Save first config
	cfg1 := &TUIConfig{
		Theme:    "monokai",
		Contrast: "high",
	}
	if err := SaveTUIConfig(cfg1); err != nil {
		t.Fatalf("First SaveTUIConfig() error: %v", err)
	}

	// Save second config (should overwrite)
	cfg2 := &TUIConfig{
		Theme:    "nord",
		Contrast: "medium",
	}
	if err := SaveTUIConfig(cfg2); err != nil {
		t.Fatalf("Second SaveTUIConfig() error: %v", err)
	}

	// Load and verify second config overwrote the first
	loaded := LoadTUIConfig()

	if loaded.Theme != "nord" {
		t.Errorf("Theme should be overwritten to 'nord', got %q", loaded.Theme)
	}

	if loaded.Contrast != "medium" {
		t.Errorf("Contrast should be overwritten to 'medium', got %q", loaded.Contrast)
	}
}

func TestDefaultTUIConfig(t *testing.T) {
	cfg := defaultTUIConfig()

	if cfg == nil {
		t.Fatal("defaultTUIConfig() returned nil")
	}

	if cfg.Theme != DefaultTheme {
		t.Errorf("Expected default theme %q, got %q", DefaultTheme, cfg.Theme)
	}

	if cfg.Contrast != DefaultContrast {
		t.Errorf("Expected default contrast %q, got %q", DefaultContrast, cfg.Contrast)
	}
}

func TestDefaultTUIConfig_NotNil(t *testing.T) {
	cfg1 := defaultTUIConfig()
	cfg2 := defaultTUIConfig()

	// Should return different pointers (new instance each time)
	if cfg1 == cfg2 {
		t.Error("defaultTUIConfig() should return new instances")
	}

	// But with same values
	if cfg1.Theme != cfg2.Theme {
		t.Error("Default configs should have same theme")
	}

	if cfg1.Contrast != cfg2.Contrast {
		t.Error("Default configs should have same contrast")
	}
}

func TestLoadTUIConfig_RoundTrip(t *testing.T) {
	// Save and restore HOME env var
	origHome := os.Getenv("HOME")
	tmpHome := t.TempDir()
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	// Create config directory
	smokeDir := filepath.Join(tmpHome, ".config", "smoke")
	if err := os.MkdirAll(smokeDir, 0755); err != nil {
		t.Fatalf("Failed to create smoke dir: %v", err)
	}

	// Save a config
	original := &TUIConfig{
		Theme:    "gruvbox",
		Contrast: "high",
	}
	if err := SaveTUIConfig(original); err != nil {
		t.Fatalf("SaveTUIConfig() error: %v", err)
	}

	// Load it back
	loaded := LoadTUIConfig()

	// Verify round-trip
	if loaded.Theme != original.Theme {
		t.Errorf("Round-trip theme mismatch: saved %q, loaded %q", original.Theme, loaded.Theme)
	}

	if loaded.Contrast != original.Contrast {
		t.Errorf("Round-trip contrast mismatch: saved %q, loaded %q", original.Contrast, loaded.Contrast)
	}
}
