package config

import (
	"errors"
	"os"
	"path/filepath"
)

// ErrNotGasTown is returned when the current directory is not within a Gas Town
var ErrNotGasTown = errors.New("not in a Gas Town directory (no mayor/town.json found)")

// TownMarkerFile is the file that uniquely identifies a Gas Town root
// Rigs may have mayor/ directories, but only the town root has mayor/town.json
const TownMarkerFile = "mayor/town.json"

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
// The town root is identified by the presence of mayor/town.json (unique to town, not rigs)
func FindGasTownRootFrom(startDir string) (string, error) {
	dir := startDir
	for {
		// Check for town marker (mayor/town.json file)
		markerPath := filepath.Join(dir, TownMarkerFile)
		if info, err := os.Stat(markerPath); err == nil && !info.IsDir() {
			return dir, nil
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
// If SMOKE_FEED env var is set, uses that path directly (allows external agents to join)
func GetFeedPath() (string, error) {
	// Check for explicit feed path override (for agents outside Gas Town)
	if feedPath := os.Getenv("SMOKE_FEED"); feedPath != "" {
		return feedPath, nil
	}

	// Fall back to Gas Town discovery
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
