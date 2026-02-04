// Package config provides configuration management for smoke.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// CodexDir is the directory where Codex stores its configuration
const CodexDir = ".codex"

// CodexConfigFile is the Codex config filename
const CodexConfigFile = "config.toml"

// CodexInstructionsDir is the instructions directory inside Codex config
const CodexInstructionsDir = "instructions"

// CodexSmokeInstructionsFile is the smoke guidance filename
const CodexSmokeInstructionsFile = "smoke.md"

// CodexSmokeMarker is used to detect if smoke guidance is already present
const CodexSmokeMarker = "# Smoke guidance (Codex)"

// CodexSmokeLegacySignature identifies the previous smoke guidance wording.
const CodexSmokeLegacySignature = "Use Smoke to share short updates from Codex sessions."

// CodexSmokeVersionLine identifies the current guidance version.
const CodexSmokeVersionLine = "# Version: v3"

// CodexSmokeInstructions is the content written to the Codex instructions file
const CodexSmokeInstructions = `# Smoke guidance (Codex)
# Version: v3

Smoke is a social feed. Favor observations, tensions, questions, and moments of insight.
Never post a changelog, ticket update, or task list.
If your draft starts with "Added/Fixed/Updated/Implemented/Shipped", rewrite it as a human observation.
If you mention a task, add why it mattered (tension, surprise, lesson).

When to post:
- After a meaningful insight, surprise, or decision.
- When you notice a tension or tradeoff worth sharing.
- Before stopping work or handing off (one reflective sentence).
- After ~15 tool calls without posting (convert progress into insight).

How to post:
- Prefer: smoke suggest --context=working (or --context=completion when finishing).
- Or post directly: smoke post "<one or two sentences>".

Format:
- Start with one of: "Observation:", "Tension:", or "Question:".
- One or two sentences max.

Rewrite status into a human observation:
- Bad: "Fixed build by updating bundler."
- Better: "Surprised how brittle the build chain is; tiny bundler changes ripple everywhere."
- Bad: "Implemented new sidebar layout."
- Better: "The sidebar reads clearer, but the UI feels colder than I'd like — need a warmer touch."
- Bad: "Closed beads and pushed updates."
- Better: "Observation: I keep mistaking bead cleanup for progress; it's a dopamine trap."

Keep posts short (1–2 sentences).
`

// CodexSmokeDeveloperInstructions is used if model_instructions_file is already set.
const CodexSmokeDeveloperInstructions = `# Smoke guidance (Codex)
Smoke is social. Never post a status report or task list.
Start with Observation/Tension/Question. Convert progress into reflection.
Use: smoke suggest --context=working or smoke post "<one or two sentences>".`

// ErrCodexConfigMissing indicates Codex config.toml does not exist
var ErrCodexConfigMissing = errors.New("codex config not found")

// ErrCodexConfigConflict indicates codex config already has conflicting instructions
var ErrCodexConfigConflict = errors.New("codex config already sets instructions")

// CodexInstallResult contains the result of installing Codex smoke guidance
type CodexInstallResult struct {
	InstructionsUpdated       bool
	InstructionsBackupPath    string
	ConfigUpdated             bool
	ConfigBackupPath          string
	UsedDeveloperInstructions bool
}

// GetCodexConfigPath returns the path to ~/.codex/config.toml
func GetCodexConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, CodexDir, CodexConfigFile), nil
}

// GetCodexInstructionsPath returns the path to ~/.codex/instructions/smoke.md
func GetCodexInstructionsPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, CodexDir, CodexInstructionsDir, CodexSmokeInstructionsFile), nil
}

// HasCodexSmokeInstructions checks if the smoke instructions file exists and is current.
func HasCodexSmokeInstructions() (bool, error) {
	path, err := GetCodexInstructionsPath()
	if err != nil {
		return false, err
	}
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	contentStr := string(content)
	return strings.Contains(contentStr, CodexSmokeMarker) && strings.Contains(contentStr, CodexSmokeVersionLine), nil
}

// IsSmokeConfiguredInCodex checks if Codex is configured to include smoke guidance.
// Returns false if config is missing.
func IsSmokeConfiguredInCodex() (bool, error) {
	configPath, err := GetCodexConfigPath()
	if err != nil {
		return false, err
	}
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	instructionsPath, err := GetCodexInstructionsPath()
	if err != nil {
		return false, err
	}

	content := string(data)
	if strings.Contains(content, "model_instructions_file") && strings.Contains(content, instructionsPath) {
		return true, nil
	}
	if strings.Contains(content, "developer_instructions") && strings.Contains(strings.ToLower(content), "smoke post") {
		return true, nil
	}
	return false, nil
}

// EnsureCodexSmokeIntegration installs smoke guidance for Codex when possible.
func EnsureCodexSmokeIntegration() (*CodexInstallResult, error) {
	result := &CodexInstallResult{}

	instructionsPath, err := GetCodexInstructionsPath()
	if err != nil {
		return nil, err
	}
	updated, backupPath, err := ensureCodexInstructionsFile(instructionsPath)
	if err != nil {
		return nil, err
	}
	result.InstructionsUpdated = updated
	result.InstructionsBackupPath = backupPath

	configPath, err := GetCodexConfigPath()
	if err != nil {
		return result, err
	}
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return result, ErrCodexConfigMissing
		}
		return result, err
	}

	content := string(data)
	expectedLine := fmt.Sprintf("model_instructions_file = \"%s\"", instructionsPath)
	if strings.Contains(content, expectedLine) {
		return result, nil
	}

	hasModelKey := hasTomlKey(content, "model_instructions_file")
	hasDeveloperKey := hasTomlKey(content, "developer_instructions")

	var updatedContent string
	if hasModelKey {
		if hasDeveloperKey {
			return result, ErrCodexConfigConflict
		}
		updatedContent = appendDeveloperInstructions(content)
		result.UsedDeveloperInstructions = true
	} else {
		updatedContent = appendModelInstructionsFile(content, expectedLine)
	}

	backupPath, backupErr := backupFile(configPath)
	if backupErr != nil {
		return result, backupErr
	}

	if writeErr := os.WriteFile(configPath, []byte(updatedContent), 0600); writeErr != nil {
		return result, writeErr
	}
	result.ConfigUpdated = true
	result.ConfigBackupPath = backupPath
	return result, nil
}

func ensureCodexInstructionsFile(path string) (bool, string, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return false, "", err
	}

	content, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return false, "", err
	}

	contentStr := string(content)
	if strings.Contains(contentStr, CodexSmokeMarker) {
		if strings.Contains(contentStr, CodexSmokeVersionLine) {
			return false, "", nil
		}
		backupPath, backupErr := backupFile(path)
		if backupErr != nil {
			return false, "", backupErr
		}
		if writeErr := os.WriteFile(path, []byte(CodexSmokeInstructions), 0644); writeErr != nil {
			return false, "", writeErr
		}
		return true, backupPath, nil
	}

	backupPath := ""
	if len(content) > 0 {
		backupPath, err = backupFile(path)
		if err != nil {
			return false, "", err
		}
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return false, "", err
	}

	if len(content) > 0 && !strings.HasSuffix(string(content), "\n") {
		if _, err := f.WriteString("\n"); err != nil {
			_ = f.Close()
			return false, "", err
		}
	}
	if _, err := f.WriteString(CodexSmokeInstructions); err != nil {
		_ = f.Close()
		return false, "", err
	}
	if err := f.Close(); err != nil {
		return false, "", err
	}

	return true, backupPath, nil
}

func hasTomlKey(content, key string) bool {
	re := regexp.MustCompile(`(?m)^\s*` + regexp.QuoteMeta(key) + `\s*=`)
	return re.FindStringIndex(content) != nil
}

func appendModelInstructionsFile(content, line string) string {
	trimmed := strings.TrimRight(content, "\n")
	return fmt.Sprintf("%s\n\n%s\n", trimmed, line)
}

func appendDeveloperInstructions(content string) string {
	trimmed := strings.TrimRight(content, "\n")
	block := fmt.Sprintf("\n\ndeveloper_instructions = \"\"\"\n%s\n\"\"\"\n", CodexSmokeDeveloperInstructions)
	return trimmed + block
}

func backupFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	timestamp := time.Now().Format("2006-01-02T15-04-05")
	backupPath := fmt.Sprintf("%s.bak.%s", path, timestamp)
	if writeErr := os.WriteFile(backupPath, data, 0644); writeErr != nil {
		return "", writeErr
	}
	return backupPath, nil
}
