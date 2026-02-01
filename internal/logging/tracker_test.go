package logging

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"testing"
)

func TestStartCommand(t *testing.T) {
	tracker := StartCommand("test", []string{"arg1", "arg2"})

	if tracker == nil {
		t.Fatal("StartCommand() returned nil")
	}

	if tracker.name != "test" {
		t.Errorf("name = %q, want %q", tracker.name, "test")
	}

	if len(tracker.args) != 2 {
		t.Errorf("args len = %d, want %d", len(tracker.args), 2)
	}

	if tracker.ctx == nil {
		t.Error("ctx should not be nil")
	}

	if tracker.startTime.IsZero() {
		t.Error("startTime should be set")
	}
}

func TestCommandTrackerSetIdentity(t *testing.T) {
	tracker := &CommandTracker{ctx: &Context{}}
	tracker.SetIdentity("claude@swift-fox/smoke", "claude", "smoke")

	if tracker.ctx.Identity != "claude@swift-fox/smoke" {
		t.Errorf("Identity = %q, want %q", tracker.ctx.Identity, "claude@swift-fox/smoke")
	}
	if tracker.ctx.Agent != "claude" {
		t.Errorf("Agent = %q, want %q", tracker.ctx.Agent, "claude")
	}
	if tracker.ctx.Project != "smoke" {
		t.Errorf("Project = %q, want %q", tracker.ctx.Project, "smoke")
	}
}

func TestCommandTrackerAddMetric(t *testing.T) {
	tracker := &CommandTracker{}
	tracker.AddMetric(slog.String("test", "value"))

	if len(tracker.metrics) != 1 {
		t.Errorf("metrics len = %d, want %d", len(tracker.metrics), 1)
	}
}

func TestCommandTrackerAddPostMetrics(t *testing.T) {
	tracker := &CommandTracker{}
	tracker.AddPostMetrics("smk-abc123", "test-author")

	if len(tracker.metrics) != 1 {
		t.Errorf("metrics len = %d, want %d", len(tracker.metrics), 1)
	}

	// Verify it's a group with post.id and post.author
	attr := tracker.metrics[0]
	if attr.Key != "post" {
		t.Errorf("metric key = %q, want %q", attr.Key, "post")
	}
}

func TestCommandTrackerAddFeedMetrics(t *testing.T) {
	tracker := &CommandTracker{}
	tracker.AddFeedMetrics(1024, 10)

	if len(tracker.metrics) != 1 {
		t.Errorf("metrics len = %d, want %d", len(tracker.metrics), 1)
	}

	attr := tracker.metrics[0]
	if attr.Key != "feed" {
		t.Errorf("metric key = %q, want %q", attr.Key, "feed")
	}
}

func TestCategorizeError(t *testing.T) {
	tests := []struct {
		err      error
		expected string
	}{
		{nil, "none"},
		{errors.New("smoke not initialized"), "not_initialized"},
		{errors.New("permission denied"), "permission"},
		{errors.New("file not found"), "not_found"},
		{errors.New("connection timeout"), "timeout"},
		{errors.New("invalid input"), "invalid_input"},
		{errors.New("parse error"), "parse_error"},
		{errors.New("connection refused"), "connection"},
		{errors.New("some random error"), "unknown"},
	}

	for _, tt := range tests {
		result := categorizeError(tt.err)
		if result != tt.expected {
			errMsg := "nil"
			if tt.err != nil {
				errMsg = tt.err.Error()
			}
			t.Errorf("categorizeError(%q) = %q, want %q", errMsg, result, tt.expected)
		}
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		s        string
		substr   string
		expected bool
	}{
		{"hello world", "world", true},
		{"Hello World", "world", true}, // case insensitive
		{"hello", "world", false},
		{"", "test", false},
		{"test", "", true},
	}

	for _, tt := range tests {
		result := contains(tt.s, tt.substr)
		if result != tt.expected {
			t.Errorf("contains(%q, %q) = %v, want %v", tt.s, tt.substr, result, tt.expected)
		}
	}
}

func TestToLower(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"HELLO", "hello"},
		{"Hello", "hello"},
		{"hello", "hello"},
		{"Hello123", "hello123"},
		{"", ""},
	}

	for _, tt := range tests {
		result := toLower(tt.input)
		if result != tt.expected {
			t.Errorf("toLower(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

// TestTrackerIntegration tests the full tracking flow with a captured log
func TestTrackerIntegration(t *testing.T) {
	// Create a buffer to capture log output
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})

	// Save and replace global logger
	oldLogger := logger
	logger = slog.New(handler)
	defer func() { logger = oldLogger }()

	// Start a command
	tracker := StartCommand("test-cmd", []string{"arg1"})
	tracker.SetIdentity("test@identity/project", "test", "project")
	tracker.AddPostMetrics("smk-test", "test-author")
	tracker.Complete()

	// Parse the log output
	lines := bytes.Split(buf.Bytes(), []byte("\n"))
	if len(lines) < 2 {
		t.Fatalf("Expected at least 2 log lines, got %d", len(lines))
	}

	// Check start message
	var startLog map[string]interface{}
	if err := json.Unmarshal(lines[0], &startLog); err != nil {
		t.Fatalf("Failed to parse start log: %v", err)
	}
	if startLog["msg"] != "command started" {
		t.Errorf("Start log msg = %v, want %q", startLog["msg"], "command started")
	}

	// Check complete message
	var completeLog map[string]interface{}
	if err := json.Unmarshal(lines[1], &completeLog); err != nil {
		t.Fatalf("Failed to parse complete log: %v", err)
	}
	if completeLog["msg"] != "command completed" {
		t.Errorf("Complete log msg = %v, want %q", completeLog["msg"], "command completed")
	}

	// Verify cmd group exists
	cmd, ok := completeLog["cmd"].(map[string]interface{})
	if !ok {
		t.Fatal("cmd group not found in complete log")
	}
	if cmd["name"] != "test-cmd" {
		t.Errorf("cmd.name = %v, want %q", cmd["name"], "test-cmd")
	}
	if cmd["duration_ms"] == nil {
		t.Error("cmd.duration_ms should be set")
	}

	// Verify ctx group exists
	ctx, ok := completeLog["ctx"].(map[string]interface{})
	if !ok {
		t.Fatal("ctx group not found in complete log")
	}
	if ctx["identity"] != "test@identity/project" {
		t.Errorf("ctx.identity = %v, want %q", ctx["identity"], "test@identity/project")
	}
}

func TestTrackerFail(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})

	oldLogger := logger
	logger = slog.New(handler)
	defer func() { logger = oldLogger }()

	tracker := StartCommand("test-cmd", []string{})
	tracker.Fail(errors.New("test error"))

	lines := bytes.Split(buf.Bytes(), []byte("\n"))
	if len(lines) < 2 {
		t.Fatalf("Expected at least 2 log lines, got %d", len(lines))
	}

	// Check fail message
	var failLog map[string]interface{}
	if err := json.Unmarshal(lines[1], &failLog); err != nil {
		t.Fatalf("Failed to parse fail log: %v", err)
	}
	if failLog["msg"] != "command failed" {
		t.Errorf("Fail log msg = %v, want %q", failLog["msg"], "command failed")
	}
	if failLog["level"] != "ERROR" {
		t.Errorf("Fail log level = %v, want %q", failLog["level"], "ERROR")
	}

	// Verify err group
	errGroup, ok := failLog["err"].(map[string]interface{})
	if !ok {
		t.Fatal("err group not found in fail log")
	}
	if errGroup["message"] != "test error" {
		t.Errorf("err.message = %v, want %q", errGroup["message"], "test error")
	}
	if errGroup["type"] == nil || errGroup["type"] == "" {
		t.Error("err.type should be set")
	}
}
