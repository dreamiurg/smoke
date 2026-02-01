package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadReadState_NonExistent(t *testing.T) {
	// Use a temp directory
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	t.Cleanup(func() { os.Setenv("HOME", originalHome) })
	os.Setenv("HOME", tmpDir)

	// LoadReadState on non-existent file returns empty state
	state, err := LoadReadState()
	if err != nil {
		t.Fatalf("LoadReadState failed: %v", err)
	}
	if state == nil {
		t.Fatal("LoadReadState returned nil state")
	}
	if state.LastReadPostID != "" {
		t.Fatalf("Expected empty LastReadPostID, got %s", state.LastReadPostID)
	}
}

func TestSaveLastReadPostID(t *testing.T) {
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	t.Cleanup(func() { os.Setenv("HOME", originalHome) })
	os.Setenv("HOME", tmpDir)

	// SaveLastReadPostID saves
	err := SaveLastReadPostID("post-123")
	if err != nil {
		t.Fatalf("SaveLastReadPostID failed: %v", err)
	}

	// Can be loaded
	state, err := LoadReadState()
	if err != nil {
		t.Fatalf("LoadReadState after save failed: %v", err)
	}
	if state.LastReadPostID != "post-123" {
		t.Fatalf("Expected 'post-123', got '%s'", state.LastReadPostID)
	}
}

func TestLoadLastReadPostID(t *testing.T) {
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	t.Cleanup(func() { os.Setenv("HOME", originalHome) })
	os.Setenv("HOME", tmpDir)

	// Empty when file doesn't exist
	postID := LoadLastReadPostID()
	if postID != "" {
		t.Fatalf("Expected empty string for non-existent file, got '%s'", postID)
	}

	// Save and load
	err := SaveLastReadPostID("post-456")
	if err != nil {
		t.Fatalf("SaveLastReadPostID failed: %v", err)
	}

	postID = LoadLastReadPostID()
	if postID != "post-456" {
		t.Fatalf("Expected 'post-456', got '%s'", postID)
	}
}

func TestSaveLastReadPostID_Overwrite(t *testing.T) {
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	t.Cleanup(func() { os.Setenv("HOME", originalHome) })
	os.Setenv("HOME", tmpDir)

	// Save first value
	err := SaveLastReadPostID("post-123")
	if err != nil {
		t.Fatalf("First SaveLastReadPostID failed: %v", err)
	}

	// Save second value (overwrites)
	err = SaveLastReadPostID("post-456")
	if err != nil {
		t.Fatalf("Second SaveLastReadPostID failed: %v", err)
	}

	// Second value is saved
	postID := LoadLastReadPostID()
	if postID != "post-456" {
		t.Fatalf("Expected 'post-456' after overwrite, got '%s'", postID)
	}
}

func TestTimestampUpdate(t *testing.T) {
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	t.Cleanup(func() { os.Setenv("HOME", originalHome) })
	os.Setenv("HOME", tmpDir)

	err := SaveLastReadPostID("post-123")
	if err != nil {
		t.Fatalf("SaveLastReadPostID failed: %v", err)
	}

	state, err := LoadReadState()
	if err != nil {
		t.Fatalf("LoadReadState failed: %v", err)
	}

	// Timestamp is updated
	if state.Updated.IsZero() {
		t.Fatal("Updated timestamp is zero")
	}
	if time.Since(state.Updated) > 1*time.Second {
		t.Fatal("Updated timestamp is too old")
	}
}

func TestSaveReadState_AtomicWrite(t *testing.T) {
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	t.Cleanup(func() { os.Setenv("HOME", originalHome) })
	os.Setenv("HOME", tmpDir)

	state := &ReadState{
		LastReadPostID: "post-123",
	}

	err := SaveReadState(state)
	if err != nil {
		t.Fatalf("SaveReadState failed: %v", err)
	}

	// Verify file was created
	path, err := GetReadStatePath()
	if err != nil {
		t.Fatalf("GetReadStatePath failed: %v", err)
	}

	_, err = os.Stat(path)
	if err != nil {
		t.Fatalf("File was not created: %v", err)
	}

	// No temp file should remain
	tmpPath := path + ".tmp"
	_, err = os.Stat(tmpPath)
	if err == nil {
		t.Fatal("Temp file was not cleaned up")
	}
	if !os.IsNotExist(err) {
		t.Fatalf("Unexpected error checking temp file: %v", err)
	}
}

func TestGetReadStatePath(t *testing.T) {
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	t.Cleanup(func() { os.Setenv("HOME", originalHome) })
	os.Setenv("HOME", tmpDir)

	path, err := GetReadStatePath()
	if err != nil {
		t.Fatalf("GetReadStatePath failed: %v", err)
	}

	expected := tmpDir + "/.config/smoke/readstate.yaml"
	if path != expected {
		t.Fatalf("Expected %s, got %s", expected, path)
	}
}

func TestLoadReadState_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	t.Cleanup(func() { os.Setenv("HOME", originalHome) })
	os.Setenv("HOME", tmpDir)

	// Create empty file
	path, err := GetReadStatePath()
	if err != nil {
		t.Fatalf("GetReadStatePath failed: %v", err)
	}

	err = os.MkdirAll(tmpDir+"/.config/smoke", 0700)
	if err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	err = os.WriteFile(path, []byte{}, 0600)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	// Should return empty state, not error
	state, err := LoadReadState()
	if err != nil {
		t.Fatalf("LoadReadState on empty file failed: %v", err)
	}
	if state.LastReadPostID != "" {
		t.Fatalf("Expected empty LastReadPostID, got %s", state.LastReadPostID)
	}
}
