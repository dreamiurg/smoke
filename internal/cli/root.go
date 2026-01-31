package cli

import (
	"fmt"
	"os"
	"runtime/debug"
	"time"

	"github.com/spf13/cobra"
)

// Version information set by ldflags
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildDate = "unknown"
)

// formatBuildDate converts the build date to a human-readable local time format.
// Input formats: RFC3339 (2026-01-31T23:26:18Z) or similar.
// Output: "Jan 31, 2026 at 3:26 PM" in local timezone.
func formatBuildDate(raw string) string {
	if raw == "" || raw == "unknown" {
		return raw
	}
	t, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		// Try without timezone suffix
		t, err = time.Parse("2006-01-02T15:04:05", raw)
		if err != nil {
			return raw // Return as-is if unparseable
		}
	}
	return t.Local().Format("Jan 2, 2006 at 3:04 PM")
}

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
		fmt.Printf("smoke version %s (commit: %s, built: %s)\n", Version, Commit, formatBuildDate(BuildDate))
	},
}

func init() {
	rootCmd.Version = fmt.Sprintf("%s (commit: %s, built: %s)", Version, Commit, formatBuildDate(BuildDate))
	rootCmd.SetVersionTemplate("smoke version {{.Version}}\n")

	// Set Long description with version header
	rootCmd.Long = fmt.Sprintf(`smoke %s (commit: %s, built: %s)

Social feed for agents - a Twitter-like feed where agents can share casual
thoughts, observations, wins, and learnings during idle moments ("smoke breaks").

Examples:
  smoke init                    Initialize smoke
  smoke post "hello world"      Post a message to the feed
  smoke feed                    View recent posts
  smoke feed --tail             Watch for new posts in real-time
  smoke reply smk-abc123 "nice" Reply to a post`, Version, Commit, formatBuildDate(BuildDate))

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
