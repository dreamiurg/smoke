package config

// Default directory and file names
const (
	// DefaultSmokeDir is the name of the smoke data directory within ~/.config/
	DefaultSmokeDir = "smoke"

	// DefaultFeedFile is the name of the feed file
	DefaultFeedFile = "feed.jsonl"

	// DefaultConfigFile is the name of the config file
	DefaultConfigFile = "config.yaml"

	// DefaultTUIConfigFile is the name of the TUI config file
	DefaultTUIConfigFile = "tui.yaml"

	// DefaultLogFile is the name of the log file
	DefaultLogFile = "smoke.log"
)

// Default TUI configuration values
const (
	// DefaultTheme is the default TUI theme
	DefaultTheme = "dracula"

	// DefaultContrast is the default contrast level
	DefaultContrast = "medium"

	// DefaultLayout is the default TUI layout
	DefaultLayout = "comfy"

	// DefaultAutoRefresh determines if auto-refresh is enabled by default
	DefaultAutoRefresh = true
)

// Default suggest configuration values
const (
	// DefaultPressure is the default pressure level for suggest nudges (0-4 scale)
	// Level 2 (balanced) provides a 50% nudge probability
	DefaultPressure = 2
)
