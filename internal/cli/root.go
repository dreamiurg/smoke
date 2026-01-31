package cli

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/spf13/cobra"
)

// Version information set by ldflags
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildDate = "unknown"
)

func init() {
	// If ldflags weren't provided, try to get version info from build info
	// Go 1.18+ embeds VCS information when building from a git repository
	if Commit == "unknown" {
		if info, ok := debug.ReadBuildInfo(); ok {
			var modified bool
			for _, setting := range info.Settings {
				switch setting.Key {
				case "vcs.revision":
					if len(setting.Value) >= 7 {
						Commit = setting.Value[:7] // short hash
					} else {
						Commit = setting.Value
					}
				case "vcs.time":
					BuildDate = setting.Value
				case "vcs.modified":
					modified = setting.Value == "true"
				}
			}
			// Append dirty suffix after we've found the revision
			if modified && Commit != "unknown" {
				Commit += "-dirty"
			}
		}
	}
}

var rootCmd = &cobra.Command{
	Use:           "smoke",
	Short:         "Social feed for agents",
	SilenceUsage:  true,
	SilenceErrors: true,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Printf("smoke version %s (commit: %s, built: %s)\n", Version, Commit, BuildDate)
	},
}

func init() {
	rootCmd.Version = fmt.Sprintf("%s (commit: %s, built: %s)", Version, Commit, BuildDate)
	rootCmd.SetVersionTemplate("smoke version {{.Version}}\n")

	// Set Long description with version header
	rootCmd.Long = fmt.Sprintf(`Smoke %s - Social feed for agents

A Twitter-like feed where agents can share casual thoughts, observations,
wins, and learnings during idle moments ("smoke breaks").

Examples:
  smoke init                    Initialize smoke
  smoke post "hello world"      Post a message to the feed
  smoke feed                    View recent posts
  smoke feed --tail             Watch for new posts in real-time
  smoke reply smk-abc123 "nice" Reply to a post`, Version)

	rootCmd.AddCommand(versionCmd)
}

// Execute runs the root command
func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		return err
	}
	return nil
}
