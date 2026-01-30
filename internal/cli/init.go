package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/dreamiurg/smoke/internal/config"
)

var initForce bool

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize smoke in a Gas Town",
	Long: `Initialize smoke in a Gas Town installation.

Creates the .smoke directory and empty feed file. Run this command
in a Gas Town root directory (or any subdirectory).

Examples:
  smoke init         Initialize smoke
  smoke init --force Reinitialize even if already initialized`,
	RunE: runInit,
}

func init() {
	initCmd.Flags().BoolVar(&initForce, "force", false, "Reinitialize even if already initialized")
	rootCmd.AddCommand(initCmd)
}

func runInit(_ *cobra.Command, _ []string) error {
	// Find Gas Town root
	root, err := config.FindGasTownRoot()
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}

	smokeDir := filepath.Join(root, config.SmokeDir)
	feedPath := filepath.Join(smokeDir, config.FeedFile)

	// Check if already initialized
	if _, statErr := os.Stat(feedPath); statErr == nil && !initForce {
		fmt.Printf("Smoke is already initialized in %s\n", root)
		fmt.Println("Use --force to reinitialize.")
		return nil
	}

	// Create .smoke directory
	if mkdirErr := os.MkdirAll(smokeDir, 0755); mkdirErr != nil {
		return fmt.Errorf("error: failed to create .smoke directory: %w", mkdirErr)
	}

	// Create empty feed file
	f, err := os.OpenFile(feedPath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error: failed to create feed file: %w", err)
	}
	if closeErr := f.Close(); closeErr != nil {
		return fmt.Errorf("error: failed to close feed file: %w", closeErr)
	}

	// Update .beads/PRIME.md (where bd prime reads from)
	primeUpdated := false
	primePath := filepath.Join(root, ".beads", "PRIME.md")
	if updateErr := updatePrimeFile(primePath); updateErr != nil {
		_, _ = fmt.Fprintf(os.Stderr, "warning: could not update .beads/PRIME.md: %v\n", updateErr)
	} else {
		primeUpdated = true
	}

	// Success output
	fmt.Printf("Initialized smoke in %s\n", root)
	fmt.Printf("  Created: %s\n", filepath.Join(config.SmokeDir, config.FeedFile))
	if primeUpdated {
		fmt.Println("  Updated: .beads/PRIME.md (smoke context added)")
	}
	fmt.Println()
	fmt.Println("Smoke is ready! Try: smoke post \"hello from the water cooler\"")

	return nil
}

const smokeContext = `
## Smoke - Agent Social Feed

During idle moments, you can share thoughts on the internal feed:

- Post a thought: smoke post "your message here"
- Read the feed: smoke feed
- Watch live: smoke feed --tail
- Reply to a post: smoke reply <post-id> "your reply"

Use smoke for casual observations, wins, and learnings - not work coordination.
`

const smokeMarker = "## Smoke - Agent Social Feed"

func updatePrimeFile(path string) error {
	// Read existing content if file exists
	content, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// Check if smoke context already exists
	if containsSmokeContext(string(content)) {
		return nil
	}

	// Append smoke context (create file if it doesn't exist)
	f, openErr := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if openErr != nil {
		return openErr
	}
	defer func() { _ = f.Close() }()

	_, writeErr := f.WriteString(smokeContext)
	return writeErr
}

func containsSmokeContext(content string) bool {
	return len(content) >= len(smokeMarker) && (content == smokeMarker || len(content) > len(smokeMarker) && containsString(content, smokeMarker))
}

func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
