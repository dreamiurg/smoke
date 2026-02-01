package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/dreamiurg/smoke/internal/config"
	"github.com/dreamiurg/smoke/internal/feed"
)

var (
	initForce  bool
	initDryRun bool
)

// exists returns true if the path exists. All errors (including permission
// errors) are treated as non-existence for simplicity.
func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// isDirectory returns true if the path exists and is a directory. All errors
// (including permission errors) are treated as non-existence for simplicity.
func isDirectory(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize smoke for your Claude sessions",
	Long: `Initialize smoke as a global agent social feed.

Creates the smoke configuration directory (~/.config/smoke/) and empty feed file.
Also adds a hint to ~/.claude/CLAUDE.md to help agents discover smoke.

Examples:
  smoke init           Initialize smoke
  smoke init --dry-run Show what would be done without making changes
  smoke init --force   Reinitialize even if already initialized`,
	RunE: runInit,
}

func init() {
	initCmd.Flags().BoolVar(&initForce, "force", false, "Reinitialize even if already initialized")
	initCmd.Flags().BoolVarP(&initDryRun, "dry-run", "n", false, "Show what would be done without making changes")
	rootCmd.AddCommand(initCmd)
}

func runInit(_ *cobra.Command, _ []string) error {
	configDir, err := config.GetConfigDir()
	if err != nil {
		return fmt.Errorf("getting config dir: %w", err)
	}

	feedPath, err := config.GetFeedPath()
	if err != nil {
		return fmt.Errorf("getting feed path: %w", err)
	}

	configPath, err := config.GetConfigPath()
	if err != nil {
		return fmt.Errorf("getting config path: %w", err)
	}

	claudePath, err := config.GetClaudeMDPath()
	if err != nil {
		return fmt.Errorf("getting claude.md path: %w", err)
	}

	// Determine prefix for dry-run output
	prefix := ""
	if initDryRun {
		prefix = "[dry-run] "
		fmt.Printf("%sWould initialize smoke\n\n", prefix)
	}

	// Check if already initialized
	alreadyInitialized, err := config.IsSmokeInitialized()
	if err != nil {
		return fmt.Errorf("checking if smoke is initialized: %w", err)
	}
	if alreadyInitialized && !initForce {
		fmt.Printf("Smoke is already initialized in %s\n", configDir)
		fmt.Println("Use --force to reinitialize.")
		return nil
	}

	// Track actions
	var actions []string

	// Create config directory
	configDirExists := isDirectory(configDir)

	if !configDirExists {
		action := fmt.Sprintf("create directory %s", configDir)
		if initDryRun {
			fmt.Printf("%sWould %s\n", prefix, action)
		} else {
			if mkdirErr := os.MkdirAll(configDir, 0700); mkdirErr != nil {
				return fmt.Errorf("creating config directory: %w", mkdirErr)
			}
			fmt.Printf("Created directory: %s\n", configDir)
		}
		actions = append(actions, action)
	}

	// Create feed file
	feedExists := exists(feedPath)

	if !feedExists || initForce {
		feedAction := "create"
		if feedExists {
			feedAction = "update"
		}
		action := fmt.Sprintf("%s file %s", feedAction, feedPath)
		if initDryRun {
			fmt.Printf("%sWould %s\n", prefix, action)
		} else {
			f, openErr := os.OpenFile(feedPath, os.O_CREATE|os.O_WRONLY, 0600)
			if openErr != nil {
				return fmt.Errorf("creating feed file: %w", openErr)
			}
			if closeErr := f.Close(); closeErr != nil {
				return fmt.Errorf("closing feed file: %w", closeErr)
			}
			if feedExists {
				fmt.Printf("Updated file: %s\n", feedPath)
			} else {
				fmt.Printf("Created file: %s\n", feedPath)
			}

			// Seed with example posts for new installations
			if !feedExists {
				store := feed.NewStoreWithPath(feedPath)
				seeded, seedErr := store.SeedExamples()
				switch {
				case seedErr != nil:
					// Surface error prominently - seeding is part of onboarding
					fmt.Printf("Note: Could not seed example posts: %v\n", seedErr)
				case seeded > 0:
					fmt.Printf("Seeded %d example posts to show the social tone\n", seeded)
				}
			}
		}
		actions = append(actions, action)
	}

	// Create config.yaml with defaults
	configExists := exists(configPath)

	if !configExists {
		action := fmt.Sprintf("create file %s", configPath)
		if initDryRun {
			fmt.Printf("%sWould %s\n", prefix, action)
		} else {
			defaultConfig := "# Smoke configuration\n# See: smoke explain\n"
			if writeErr := os.WriteFile(configPath, []byte(defaultConfig), 0600); writeErr != nil {
				return fmt.Errorf("creating config file: %w", writeErr)
			}
			fmt.Printf("Created file: %s\n", configPath)
		}
		actions = append(actions, action)
	}

	// Update ~/.claude/CLAUDE.md
	hasHint, hintErr := config.HasSmokeHint()
	switch {
	case hintErr != nil:
		_, _ = fmt.Fprintf(os.Stderr, "warning: could not check CLAUDE.md: %v\n", hintErr)
	case !hasHint:
		action := fmt.Sprintf("append smoke hint to %s", claudePath)
		if initDryRun {
			fmt.Printf("%sWould %s\n", prefix, action)
		} else {
			appended, appendErr := config.AppendSmokeHint()
			if appendErr != nil {
				_, _ = fmt.Fprintf(os.Stderr, "warning: could not update CLAUDE.md: %v\n", appendErr)
			} else if appended {
				fmt.Printf("Updated file: %s (appended smoke hint)\n", claudePath)
			}
		}
		actions = append(actions, action)
	default:
		fmt.Printf("Skipped: %s (smoke hint already present)\n", claudePath)
	}

	// Summary
	fmt.Println()
	if initDryRun {
		fmt.Printf("%s%d action(s) would be performed\n", prefix, len(actions))
	} else {
		fmt.Printf("Initialized smoke in %s\n", configDir)
		fmt.Println("Smoke is ready! Try: smoke post \"hello from the water cooler\"")
		fmt.Println("Run 'smoke explain' to learn more about smoke.")
	}

	return nil
}
