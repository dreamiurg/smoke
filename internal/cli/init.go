package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/dreamiurg/smoke/internal/config"
	"github.com/dreamiurg/smoke/internal/feed"
	"github.com/dreamiurg/smoke/internal/hooks"
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

		// Check if hooks are missing and suggest installation
		status, hookErr := hooks.GetStatus()
		if hookErr == nil && status.State != hooks.StateInstalled {
			fmt.Println("\nHooks not installed. Run: smoke hooks install")
		}

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

	// Create config.yaml with defaults (contexts and examples)
	configExists := exists(configPath)

	if !configExists {
		action := fmt.Sprintf("create file %s (with default contexts and examples)", configPath)
		if initDryRun {
			fmt.Printf("%sWould %s\n", prefix, action)
		} else {
			defaultConfig := config.DefaultSuggestConfigYAML()
			if writeErr := os.WriteFile(configPath, []byte(defaultConfig), 0600); writeErr != nil {
				return fmt.Errorf("creating config file: %w", writeErr)
			}
			fmt.Printf("Created file: %s (with default contexts and examples)\n", configPath)
		}
		actions = append(actions, action)
	}

	// Apply migrations and ensure config has current schema version (for new and existing configs)
	if !initDryRun {
		applied, err := config.ApplyMigrations(false)
		if err != nil {
			return fmt.Errorf("applying migrations: %w", err)
		}
		// If no migrations were applied (new config), ensure schema version is set
		if len(applied) == 0 {
			configMap, err := config.GetConfigAsMap()
			if err != nil {
				return fmt.Errorf("reading config for schema version: %w", err)
			}
			configMap[config.SchemaVersionKey] = config.CurrentSchemaVersion
			if err := config.WriteConfigMap(configMap); err != nil {
				return fmt.Errorf("writing schema version to config: %w", err)
			}
		}
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
			hintResult, appendErr := config.AppendSmokeHint()
			if appendErr != nil {
				_, _ = fmt.Fprintf(os.Stderr, "warning: could not update CLAUDE.md: %v\n", appendErr)
			} else if hintResult != nil && hintResult.Appended {
				if hintResult.BackupPath != "" {
					fmt.Printf("Backed up CLAUDE.md to: %s\n", hintResult.BackupPath)
				}
				fmt.Printf("Updated file: %s (appended smoke hint)\n", claudePath)
			}
		}
		actions = append(actions, action)
	default:
		fmt.Printf("Skipped: %s (smoke hint already present)\n", claudePath)
	}

	// Install hooks (unless dry-run)
	if !initDryRun {
		hookResult, hookErr := hooks.Install(hooks.InstallOptions{Force: false})
		if hookErr != nil {
			// Graceful degradation per FR-002: warn but don't fail init
			if errors.Is(hookErr, hooks.ErrScriptsModified) {
				fmt.Fprintf(os.Stderr, "\nNote: Hook scripts have been modified. Run 'smoke hooks install --force' to update.\n")
			} else {
				fmt.Fprintf(os.Stderr, "\nNote: Could not install hooks: %v\n", hookErr)
				fmt.Fprintf(os.Stderr, "  Run 'smoke hooks install' manually after fixing the issue.\n")
			}
		} else {
			if hookResult != nil && hookResult.BackupPath != "" {
				fmt.Printf("Backed up Claude settings to: %s\n", hookResult.BackupPath)
			}
			fmt.Printf("Hooks installed: ~/.claude/hooks/smoke-*.sh\n")
		}
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
