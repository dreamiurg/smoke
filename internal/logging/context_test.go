package logging

import (
	"log/slog"
	"os"
	"strings"
	"testing"
)

func TestDetectEnv(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		expected string
	}{
		{
			name:     "claude_code environment",
			envVars:  map[string]string{"CLAUDECODE": "1"},
			expected: "claude_code",
		},
		{
			name:     "github_actions CI",
			envVars:  map[string]string{"GITHUB_ACTIONS": "true"},
			expected: "ci",
		},
		{
			name:     "gitlab CI",
			envVars:  map[string]string{"GITLAB_CI": "true"},
			expected: "ci",
		},
		{
			name:     "generic CI",
			envVars:  map[string]string{"CI": "true"},
			expected: "ci",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and clear relevant env vars
			savedEnvs := make(map[string]string)
			envKeys := []string{"CLAUDECODE", "CI", "GITHUB_ACTIONS", "GITLAB_CI", "CIRCLECI", "JENKINS_URL", "BUILDKITE", "TRAVIS"}
			for _, key := range envKeys {
				savedEnvs[key] = os.Getenv(key)
				_ = os.Unsetenv(key)
			}

			// Set test env vars
			for k, v := range tt.envVars {
				_ = os.Setenv(k, v)
			}

			// Run test
			result := detectEnv()
			if result != tt.expected {
				t.Errorf("detectEnv() = %q, want %q", result, tt.expected)
			}

			// Restore env vars
			for k, v := range savedEnvs {
				if v == "" {
					_ = os.Unsetenv(k)
				} else {
					_ = os.Setenv(k, v)
				}
			}
		})
	}
}

func TestCaptureContext(t *testing.T) {
	ctx := CaptureContext()

	if ctx == nil {
		t.Fatal("CaptureContext() returned nil")
	}

	// Env should be set to something
	if ctx.Env == "" {
		t.Error("Env should not be empty")
	}

	// Cwd should be set
	if ctx.Cwd == "" {
		t.Error("Cwd should not be empty")
	}

	// Session should be set
	if ctx.Session == "" {
		t.Error("Session should not be empty")
	}
	// Caller should be set
	if ctx.Caller == "" {
		t.Error("Caller should not be empty")
	}

}

func TestContextSetIdentity(t *testing.T) {
	ctx := &Context{}
	ctx.SetIdentity("claude@swift-fox/smoke", "claude", "smoke")

	if ctx.Identity != "claude@swift-fox/smoke" {
		t.Errorf("Identity = %q, want %q", ctx.Identity, "claude@swift-fox/smoke")
	}
	if ctx.Agent != "claude" {
		t.Errorf("Agent = %q, want %q", ctx.Agent, "claude")
	}
	if ctx.Project != "smoke" {
		t.Errorf("Project = %q, want %q", ctx.Project, "smoke")
	}
}

func TestContextAttrs(t *testing.T) {
	ctx := &Context{
		Identity: "swift-fox@smoke",
		Agent:    "claude",
		Session:  "test-session",
		Env:      "terminal",
		Project:  "smoke",
		Cwd:      "/home/test",
	}

	attr := ctx.Attrs()

	if attr.Key != "ctx" {
		t.Errorf("Attr key = %q, want %q", attr.Key, "ctx")
	}

	// Check it's a group
	if attr.Value.Kind() != slog.KindGroup {
		t.Errorf("Attr value kind = %v, want Group", attr.Value.Kind())
	}
}

func TestGetSessionID(t *testing.T) {
	// Save env
	savedClaudeCode := os.Getenv("CLAUDECODE")
	savedTermSession := os.Getenv("TERM_SESSION_ID")
	savedWindowID := os.Getenv("WINDOWID")

	// Clear env
	_ = os.Unsetenv("CLAUDECODE")
	_ = os.Unsetenv("TERM_SESSION_ID")
	_ = os.Unsetenv("WINDOWID")

	t.Run("falls back to ppid", func(t *testing.T) {
		result := getSessionID()
		if !strings.HasPrefix(result, "pid-") {
			t.Errorf("getSessionID() = %q, want prefix 'pid-'", result)
		}
	})

	t.Run("uses TERM_SESSION_ID when set", func(t *testing.T) {
		_ = os.Setenv("TERM_SESSION_ID", "test-session")
		result := getSessionID()
		if result != "test-session" {
			t.Errorf("getSessionID() = %q, want %q", result, "test-session")
		}
		_ = os.Unsetenv("TERM_SESSION_ID")
	})

	t.Run("truncates long TERM_SESSION_ID", func(t *testing.T) {
		longID := "12345678901234567890" // 20 chars
		_ = os.Setenv("TERM_SESSION_ID", longID)
		result := getSessionID()
		if result != longID[:16] {
			t.Errorf("getSessionID() = %q, want %q", result, longID[:16])
		}
		_ = os.Unsetenv("TERM_SESSION_ID")
	})

	t.Run("uses WINDOWID when TERM_SESSION_ID not set", func(t *testing.T) {
		_ = os.Setenv("WINDOWID", "12345")
		result := getSessionID()
		if result != "win-12345" {
			t.Errorf("getSessionID() = %q, want %q", result, "win-12345")
		}
		_ = os.Unsetenv("WINDOWID")
	})

	t.Run("claude_code uses ppid format", func(t *testing.T) {
		_ = os.Setenv("CLAUDECODE", "1")
		result := getSessionID()
		if !strings.HasPrefix(result, "claude-") {
			t.Errorf("getSessionID() = %q, want prefix 'claude-'", result)
		}
		_ = os.Unsetenv("CLAUDECODE")
	})

	// Restore env
	if savedClaudeCode != "" {
		_ = os.Setenv("CLAUDECODE", savedClaudeCode)
	}
	if savedTermSession != "" {
		_ = os.Setenv("TERM_SESSION_ID", savedTermSession)
	}
	if savedWindowID != "" {
		_ = os.Setenv("WINDOWID", savedWindowID)
	}
}

func TestDetectCallerAgentFromEnv(t *testing.T) {
	envKeys := []string{
		"SMOKE_AGENT",
		"CLAUDECODE",
		"CLAUDE_CODE",
		"CLAUDE_CODE_SUBAGENT_MODEL",
		"ANTHROPIC_API_KEY",
		"ANTHROPIC_MODEL",
		"ANTHROPIC_DEFAULT_OPUS_MODEL",
		"ANTHROPIC_DEFAULT_SONNET_MODEL",
		"ANTHROPIC_DEFAULT_HAIKU_MODEL",
		"GEMINI_CLI",
		"CODEX",
		"CODEX_CLI",
		"OPENAI_CODEX",
		"CODEX_CI",
		"GEMINI_API_KEY",
		"GOOGLE_API_KEY",
		"GEMINI_MODEL",
		"GOOGLE_CLOUD_PROJECT",
		"GOOGLE_CLOUD_LOCATION",
		"OPENAI_API_KEY",
	}

	saveEnv := func() map[string]string {
		saved := make(map[string]string, len(envKeys))
		for _, key := range envKeys {
			saved[key] = os.Getenv(key)
			_ = os.Unsetenv(key)
		}
		return saved
	}

	restoreEnv := func(saved map[string]string) {
		for key, val := range saved {
			if val == "" {
				_ = os.Unsetenv(key)
			} else {
				_ = os.Setenv(key, val)
			}
		}
	}

	tests := []struct {
		name     string
		envVars  map[string]string
		expected string
	}{
		{
			name:     "smoke agent override",
			envVars:  map[string]string{"SMOKE_AGENT": "GeMiNi", "CLAUDECODE": "1"},
			expected: "gemini",
		},
		{
			name:     "claude via claude code",
			envVars:  map[string]string{"CLAUDECODE": "1"},
			expected: "claude",
		},
		{
			name:     "claude via anthropic key",
			envVars:  map[string]string{"ANTHROPIC_API_KEY": "x"},
			expected: "claude",
		},
		{
			name:     "gemini cli",
			envVars:  map[string]string{"GEMINI_CLI": "1"},
			expected: "gemini",
		},
		{
			name:     "codex cli",
			envVars:  map[string]string{"CODEX_CLI": "1"},
			expected: "codex",
		},
		{
			name:     "gemini api key",
			envVars:  map[string]string{"GEMINI_API_KEY": "x"},
			expected: "gemini",
		},
		{
			name:     "openai api key",
			envVars:  map[string]string{"OPENAI_API_KEY": "x"},
			expected: "codex",
		},
		{
			name:     "none",
			envVars:  map[string]string{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			saved := saveEnv()
			for k, v := range tt.envVars {
				_ = os.Setenv(k, v)
			}

			got := detectCallerAgentFromEnv()
			if got != tt.expected {
				t.Errorf("detectCallerAgentFromEnv() = %q, want %q", got, tt.expected)
			}

			restoreEnv(saved)
		})
	}
}

func TestItoa(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{0, "0"},
		{1, "1"},
		{-1, "-1"},
		{123, "123"},
		{-456, "-456"},
		{12345, "12345"},
	}

	for _, tt := range tests {
		result := itoa(tt.input)
		if result != tt.expected {
			t.Errorf("itoa(%d) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestExtractProjectFromCwd(t *testing.T) {
	tests := []struct {
		cwd      string
		expected string
	}{
		{"/home/user/projects/smoke", "smoke"},
		{"/home/user/smoke", "smoke"},
		{"", ""},
		{"/", "/"},
	}

	for _, tt := range tests {
		result := extractProjectFromCwd(tt.cwd)
		if result != tt.expected {
			t.Errorf("extractProjectFromCwd(%q) = %q, want %q", tt.cwd, result, tt.expected)
		}
	}
}
