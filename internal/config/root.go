package config

import (
	"errors"
	"os"
	"path/filepath"
)

// ErrNotGasTown is returned when the current directory is not within a Gas Town
var ErrNotGasTown = errors.New("not in a Gas Town directory (no .beads/ found)")

// GasTownMarkers are directories that indicate a Gas Town root
var GasTownMarkers = []string{".beads", "mayor"}

// SmokeDir is the name of the smoke data directory
const SmokeDir = ".smoke"

// FeedFile is the name of the feed file
const FeedFile = "feed.jsonl"

// FindGasTownRoot walks up the directory tree to find the Gas Town root
// It returns the path to the Gas Town root directory or an error if not found
func FindGasTownRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return FindGasTownRootFrom(dir)
}

// FindGasTownRootFrom walks up from the given directory to find the Gas Town root
func FindGasTownRootFrom(startDir string) (string, error) {
	dir := startDir
	for {
		// Check for Gas Town markers
		for _, marker := range GasTownMarkers {
			markerPath := filepath.Join(dir, marker)
			if info, err := os.Stat(markerPath); err == nil && info.IsDir() {
				return dir, nil
			}
		}

		// Move up one directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root
			return "", ErrNotGasTown
		}
		dir = parent
	}
}

// GetSmokeDir returns the path to the .smoke directory
func GetSmokeDir() (string, error) {
	root, err := FindGasTownRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, SmokeDir), nil
}

// GetFeedPath returns the path to the feed.jsonl file
func GetFeedPath() (string, error) {
	smokeDir, err := GetSmokeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(smokeDir, FeedFile), nil
}

// IsSmokeInitialized checks if smoke has been initialized in the current Gas Town
func IsSmokeInitialized() (bool, error) {
	feedPath, err := GetFeedPath()
	if err != nil {
		return false, err
	}
	_, err = os.Stat(feedPath)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
