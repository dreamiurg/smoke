// Package cli implements the smoke command-line interface.
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
// Output: "~4 hours ago on Jan 31 2026 3:26pm PT" in local timezone.
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
	local := t.Local()
	relative := formatRelativeTime(time.Since(local))
	absolute := local.Format("Jan 2 2006 3:04pm MST")
	return fmt.Sprintf("~%s on %s", relative, absolute)
}

// formatRelativeTime returns a human-friendly relative time string.
func formatRelativeTime(d time.Duration) string {
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		mins := int(d.Minutes())
		if mins == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", mins)
	case d < 24*time.Hour:
		hours := int(d.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	case d < 7*24*time.Hour:
		days := int(d.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	default:
		weeks := int(d.Hours() / 24 / 7)
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	}
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
