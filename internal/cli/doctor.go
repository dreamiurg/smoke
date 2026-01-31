package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/term"
	"gopkg.in/yaml.v3"

	"github.com/dreamiurg/smoke/internal/config"
)

// isTerminal returns true if stdout is a terminal
func isTerminal() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

// color returns the colored string if color output is enabled
func color(c, s string) string {
	if !useColor {
		return s
	}
	return c + s + colorReset
}

// CheckStatus represents the result of a health check
type CheckStatus int

// CheckStatus constants
const (
	// StatusPass indicates the check passed successfully
	StatusPass CheckStatus = iota
	// StatusWarn indicates a non-critical issue
	StatusWarn
	// StatusFail indicates a critical issue
	StatusFail
)

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
	colorDim    = "\033[2m"
)

// useColor determines if color output should be used
var useColor = isTerminal()

// Check represents a single diagnostic check
type Check struct {
	Name    string
	Status  CheckStatus
	Message string
	Detail  string       // Optional additional info
	CanFix  bool         // Whether --fix can repair this
	Fix     func() error // Fix function if CanFix is true
}

// Category groups related checks
type Category struct {
	Name   string
	Checks []Check
}

var (
	doctorFix    bool
	doctorDryRun bool
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check smoke installation health",
	Long: `Diagnose smoke installation and report any issues.

Checks configuration directory, feed file, and data integrity.
Use --fix to automatically repair common problems.

Examples:
  smoke doctor              Check installation health
  smoke doctor --fix        Automatically fix problems
  smoke doctor --fix --dry-run  Preview what would be fixed`,
	RunE: runDoctor,
}

func init() {
	doctorCmd.Flags().BoolVar(&doctorFix, "fix", false, "Automatically fix problems")
	doctorCmd.Flags().BoolVar(&doctorDryRun, "dry-run", false, "Preview what would be fixed (use with --fix)")
	rootCmd.AddCommand(doctorCmd)
}

func runDoctor(_ *cobra.Command, _ []string) error {
	categories := runChecks()

	// Apply fixes if --fix was provided
	if doctorFix {
		fixCount, err := applyFixes(categories, doctorDryRun)
		if err != nil {
			return err
		}

		// Re-run checks after fixes to show updated status
		if !doctorDryRun && fixCount > 0 {
			categories = runChecks()
		}

		printReport(categories)

		switch {
		case fixCount == 0:
			fmt.Println("No problems to fix.")
		case doctorDryRun:
			fmt.Printf("\n%d issue(s) would be fixed.\n", fixCount)
		default:
			fmt.Printf("\nFixed %d issue(s).\n", fixCount)
		}
	} else {
		printReport(categories)
	}

	exitCode := computeExitCode(categories)
	if exitCode != 0 {
		exitWithCode(exitCode)
	}
	return nil
}

// checkVersion returns the smoke version as a check
func checkVersion() Check {
	return Check{
		Name:    "Smoke Version",
		Status:  StatusPass,
		Message: Version,
		CanFix:  false,
	}
}

// runChecks executes all health checks and returns categories
func runChecks() []Category {
	return []Category{
		{
			Name: "INSTALLATION",
			Checks: []Check{
				checkConfigDir(),
				checkFeedFile(),
			},
		},
		{
			Name: "DATA",
			Checks: []Check{
				checkFeedFormat(),
				checkConfigFile(),
			},
		},
		{
			Name: "VERSION",
			Checks: []Check{
				checkVersion(),
			},
		},
	}
}

// formatCheck formats a single check result for display
func formatCheck(c Check) string {
	var indicator string
	switch c.Status {
	case StatusPass:
		indicator = color(colorGreen, "✓")
	case StatusWarn:
		indicator = color(colorYellow, "⚠")
	case StatusFail:
		indicator = color(colorRed, "✗")
	}

	line := fmt.Sprintf("  %s  %s %s", indicator, c.Name, c.Message)
	if c.Detail != "" {
		line += fmt.Sprintf("\n     %s", color(colorDim, "└─ "+c.Detail))
	}
	return line
}

// formatCategory formats a category with all its checks
func formatCategory(cat Category) string {
	result := color(colorCyan, cat.Name) + "\n"
	for _, check := range cat.Checks {
		result += formatCheck(check) + "\n"
	}
	return result
}

// printReport outputs the full doctor report
func printReport(categories []Category) {
	// Version header
	fmt.Printf("smoke doctor %s\n\n", Version)

	// Print each category
	for _, cat := range categories {
		fmt.Print(formatCategory(cat))
		fmt.Println()
	}
}

// computeExitCode determines exit code based on check statuses
func computeExitCode(categories []Category) int {
	hasError := false
	hasWarning := false

	for _, cat := range categories {
		for _, check := range cat.Checks {
			switch check.Status {
			case StatusFail:
				hasError = true
			case StatusWarn:
				hasWarning = true
			}
		}
	}

	if hasError {
		return 2
	}
	if hasWarning {
		return 1
	}
	return 0
}

// exitWithCode exits with the given code, for use after doctor completes
func exitWithCode(code int) {
	os.Exit(code)
}

// fixConfigDir creates the config directory
func fixConfigDir() error {
	configDir, err := config.GetConfigDir()
	if err != nil {
		return err
	}
	return os.MkdirAll(configDir, 0755)
}

// fixFeedFile creates an empty feed file
func fixFeedFile() error {
	feedPath, err := config.GetFeedPath()
	if err != nil {
		return err
	}
	f, err := os.Create(feedPath)
	if err != nil {
		return err
	}
	return f.Close()
}

// fixConfigFile creates a default config file
func fixConfigFile() error {
	configPath, err := config.GetConfigPath()
	if err != nil {
		return err
	}
	defaultConfig := "# Smoke configuration\n# See: smoke explain\n"
	return os.WriteFile(configPath, []byte(defaultConfig), 0644)
}

// applyFixes attempts to fix all fixable issues
// Returns the number of fixes applied (or would be applied in dry-run mode)
func applyFixes(categories []Category, dryRun bool) (int, error) {
	fixCount := 0

	for _, cat := range categories {
		for _, check := range cat.Checks {
			if check.Status != StatusPass && check.CanFix && check.Fix != nil {
				if dryRun {
					fmt.Printf("Would fix: %s\n", check.Name)
				} else {
					if err := check.Fix(); err != nil {
						fmt.Printf("Failed to fix %s: %v\n", check.Name, err)
						continue
					}
					fmt.Printf("Fixed: %s\n", check.Name)
				}
				fixCount++
			}
		}
	}

	return fixCount, nil
}

// checkConfigDir verifies the config directory exists and is writable
func checkConfigDir() Check {
	check := Check{
		Name:   "Config Directory",
		CanFix: true,
		Fix:    fixConfigDir,
	}

	configDir, err := config.GetConfigDir()
	if err != nil {
		check.Status = StatusFail
		check.Message = "cannot determine config directory"
		check.Detail = err.Error()
		return check
	}

	info, err := os.Stat(configDir)
	if os.IsNotExist(err) {
		check.Status = StatusFail
		check.Message = "not found"
		check.Detail = fmt.Sprintf("Run 'smoke doctor --fix' to create %s", configDir)
		return check
	}
	if err != nil {
		check.Status = StatusFail
		check.Message = "cannot access"
		check.Detail = err.Error()
		check.CanFix = false
		return check
	}
	if !info.IsDir() {
		check.Status = StatusFail
		check.Message = "not a directory"
		check.Detail = configDir
		check.CanFix = false
		return check
	}

	// Check if writable by creating a temp file
	testFile := configDir + "/.doctor-test"
	f, err := os.Create(testFile)
	if err != nil {
		check.Status = StatusWarn
		check.Message = fmt.Sprintf("%s (not writable)", configDir)
		check.Detail = "Permission denied - check directory permissions"
		check.CanFix = false
		return check
	}
	_ = f.Close()
	_ = os.Remove(testFile)

	check.Status = StatusPass
	check.Message = configDir
	check.CanFix = false
	return check
}

// checkFeedFile verifies the feed file exists and is readable
func checkFeedFile() Check {
	check := Check{
		Name:   "Feed File",
		CanFix: true,
		Fix:    fixFeedFile,
	}

	feedPath, err := config.GetFeedPath()
	if err != nil {
		check.Status = StatusFail
		check.Message = "cannot determine feed path"
		check.Detail = err.Error()
		return check
	}

	info, err := os.Stat(feedPath)
	if os.IsNotExist(err) {
		check.Status = StatusFail
		check.Message = "not found"
		check.Detail = fmt.Sprintf("Run 'smoke doctor --fix' to create %s", feedPath)
		return check
	}
	if err != nil {
		check.Status = StatusFail
		check.Message = "cannot access"
		check.Detail = err.Error()
		check.CanFix = false
		return check
	}
	if info.IsDir() {
		check.Status = StatusFail
		check.Message = "is a directory, expected file"
		check.Detail = feedPath
		check.CanFix = false
		return check
	}

	// Check if readable
	f, err := os.Open(feedPath)
	if err != nil {
		check.Status = StatusWarn
		check.Message = fmt.Sprintf("%s (not readable)", feedPath)
		check.Detail = "Permission denied - check file permissions"
		check.CanFix = false
		return check
	}
	_ = f.Close()

	check.Status = StatusPass
	check.Message = feedPath
	check.CanFix = false
	return check
}

// checkFeedFormat validates JSONL integrity of the feed file
func checkFeedFormat() Check {
	check := Check{
		Name:   "Feed Format",
		CanFix: false,
	}

	feedPath, err := config.GetFeedPath()
	if err != nil {
		check.Status = StatusFail
		check.Message = "cannot determine feed path"
		return check
	}

	f, err := os.Open(feedPath)
	if err != nil {
		check.Status = StatusFail
		check.Message = "cannot open feed file"
		check.Detail = err.Error()
		return check
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	totalLines := 0
	validLines := 0
	invalidLines := 0

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue // Skip empty lines
		}
		totalLines++
		var js json.RawMessage
		if json.Unmarshal([]byte(line), &js) == nil {
			validLines++
		} else {
			invalidLines++
		}
	}

	if err := scanner.Err(); err != nil {
		check.Status = StatusFail
		check.Message = "error reading feed"
		check.Detail = err.Error()
		return check
	}

	if totalLines == 0 {
		check.Status = StatusPass
		check.Message = "empty (0 posts)"
		return check
	}

	if invalidLines > 0 {
		check.Status = StatusWarn
		check.Message = fmt.Sprintf("%d/%d lines valid", validLines, totalLines)
		check.Detail = "Some lines contain invalid JSON - manual inspection recommended"
		return check
	}

	check.Status = StatusPass
	check.Message = fmt.Sprintf("%d posts, all valid", validLines)
	return check
}

// checkConfigFile verifies config.yaml exists and is valid YAML
func checkConfigFile() Check {
	check := Check{
		Name:   "Config File",
		CanFix: true,
		Fix:    fixConfigFile,
	}

	configPath, err := config.GetConfigPath()
	if err != nil {
		check.Status = StatusFail
		check.Message = "cannot determine config path"
		check.Detail = err.Error()
		return check
	}

	data, err := os.ReadFile(configPath)
	if os.IsNotExist(err) {
		check.Status = StatusWarn
		check.Message = "missing (using defaults)"
		check.Detail = fmt.Sprintf("Run 'smoke doctor --fix' to create %s", configPath)
		return check
	}
	if err != nil {
		check.Status = StatusFail
		check.Message = "cannot read"
		check.Detail = err.Error()
		check.CanFix = false
		return check
	}

	// Validate YAML syntax
	var parsed interface{}
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		check.Status = StatusFail
		check.Message = "invalid YAML"
		check.Detail = err.Error()
		check.CanFix = false
		return check
	}

	check.Status = StatusPass
	check.Message = configPath
	check.CanFix = false
	return check
}
