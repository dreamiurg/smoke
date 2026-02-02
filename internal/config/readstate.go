// Package config provides configuration and initialization management for smoke.
// It handles directory paths, feed storage, and smoke initialization state.
package config

import (
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// ReadState stores the last-read post ID for the human operator.
// There's a single read marker shared across all sessions.
type ReadState struct {
	LastReadPostID string    `yaml:"last_read_post_id"`
	NudgeCount     int       `yaml:"nudge_count,omitempty"`
	Updated        time.Time `yaml:"updated"`
}

// GetReadStatePath returns the path to the readstate.yaml file
func GetReadStatePath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, DefaultReadStateFile), nil
}

// LoadReadState loads the read state from disk.
// Returns an empty state if the file doesn't exist.
// Returns an error only for parse failures.
func LoadReadState() (*ReadState, error) {
	path, err := GetReadStatePath()
	if err != nil {
		return &ReadState{}, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		// File doesn't exist or can't be read - return empty state
		if os.IsNotExist(err) {
			return &ReadState{}, nil
		}
		return nil, err
	}

	// Handle empty file
	if len(data) == 0 {
		return &ReadState{}, nil
	}

	var state ReadState
	if err := yaml.Unmarshal(data, &state); err != nil {
		return nil, err
	}

	return &state, nil
}

// SaveReadState saves the read state to disk atomically.
// Creates the config directory if it doesn't exist.
// Updates the timestamp before saving.
func SaveReadState(state *ReadState) error {
	path, err := GetReadStatePath()
	if err != nil {
		return err
	}

	// Update timestamp
	state.Updated = time.Now()

	// Ensure the directory exists
	configDir := filepath.Dir(path)
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return err
	}

	// Marshal to YAML
	data, marshalErr := yaml.Marshal(state)
	if marshalErr != nil {
		return marshalErr
	}

	// Atomic write: temp file + rename
	tmpFile := path + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0600); err != nil {
		return err
	}

	// Atomic rename
	if err := os.Rename(tmpFile, path); err != nil {
		// Best effort cleanup on error
		_ = os.Remove(tmpFile)
		return err
	}

	return nil
}

// LoadLastReadPostID loads and returns the last-read post ID,
// or an empty string if not set or file doesn't exist.
func LoadLastReadPostID() string {
	state, err := LoadReadState()
	if err != nil || state == nil {
		return ""
	}
	return state.LastReadPostID
}

// SaveLastReadPostID saves the last-read post ID to disk.
func SaveLastReadPostID(postID string) error {
	// Preserve existing nudge count when updating read state
	state, _ := LoadReadState()
	if state == nil {
		state = &ReadState{}
	}
	state.LastReadPostID = postID
	return SaveReadState(state)
}

// GetNudgeCount returns the current nudge counter value.
func GetNudgeCount() int {
	state, err := LoadReadState()
	if err != nil || state == nil {
		return 0
	}
	return state.NudgeCount
}

// IncrementNudgeCount increments the nudge counter and saves to disk.
func IncrementNudgeCount() error {
	state, _ := LoadReadState()
	if state == nil {
		state = &ReadState{}
	}
	state.NudgeCount++
	return SaveReadState(state)
}
