package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetCodexConfigPath(t *testing.T) {
	path, err := GetCodexConfigPath()
	if err != nil {
		t.Fatalf("GetCodexConfigPath() error: %v", err)
	}
	if !strings.HasSuffix(path, filepath.Join(".codex", "config.toml")) {
		t.Errorf("GetCodexConfigPath() should end with .codex/config.toml, got %s", path)
	}
}

func TestGetCodexInstructionsPath(t *testing.T) {
	path, err := GetCodexInstructionsPath()
	if err != nil {
		t.Fatalf("GetCodexInstructionsPath() error: %v", err)
	}
	if !strings.HasSuffix(path, filepath.Join(".codex", "instructions", "smoke.md")) {
		t.Errorf("GetCodexInstructionsPath() should end with .codex/instructions/smoke.md, got %s", path)
	}
}

func TestIsSmokeConfiguredInCodexMissingConfig(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	found, err := IsSmokeConfiguredInCodex()
	if err != nil {
		t.Fatalf("IsSmokeConfiguredInCodex() error: %v", err)
	}
	if found {
		t.Error("IsSmokeConfiguredInCodex() = true, want false when config missing")
	}
}

func TestIsSmokeConfiguredInCodexModelFile(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	configDir := filepath.Join(tmpDir, CodexDir)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}
	instructionsPath, _ := GetCodexInstructionsPath()
	configPath := filepath.Join(configDir, CodexConfigFile)
	content := "model = \"gpt-5.2-codex\"\nmodel_instructions_file = \"" + instructionsPath + "\"\n"
	if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	found, err := IsSmokeConfiguredInCodex()
	if err != nil {
		t.Fatalf("IsSmokeConfiguredInCodex() error: %v", err)
	}
	if !found {
		t.Error("IsSmokeConfiguredInCodex() = false, want true")
	}
}

func TestEnsureCodexSmokeIntegration(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	configDir := filepath.Join(tmpDir, CodexDir)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(configDir, CodexConfigFile)
	if err := os.WriteFile(configPath, []byte("model = \"gpt-5.2-codex\"\n"), 0600); err != nil {
		t.Fatal(err)
	}

	result, err := EnsureCodexSmokeIntegration()
	if err != nil {
		t.Fatalf("EnsureCodexSmokeIntegration() error: %v", err)
	}
	if result == nil {
		t.Fatal("EnsureCodexSmokeIntegration() returned nil result")
	}

	instructionsPath, _ := GetCodexInstructionsPath()
	data, err := os.ReadFile(instructionsPath)
	if err != nil {
		t.Fatalf("expected instructions file, got error: %v", err)
	}
	if !strings.Contains(string(data), CodexSmokeMarker) {
		t.Error("instructions file missing smoke marker")
	}

	configData, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("expected config file, got error: %v", err)
	}
	if !strings.Contains(string(configData), "model_instructions_file") {
		t.Error("config file missing model_instructions_file entry")
	}
}
