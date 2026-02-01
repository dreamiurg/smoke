package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"golang.org/x/term"
	"gopkg.in/yaml.v3"

	"github.com/dreamiurg/smoke/internal/config"
	"github.com/dreamiurg/smoke/internal/feed"
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

// Color aliases for doctor output (using feed package constants)
const (
	colorReset  = feed.Reset
	colorRed    = feed.FgRed
	colorGreen  = feed.FgGreen
	colorYellow = feed.FgYellow
	colorCyan   = feed.FgCyan
	colorDim    = feed.Dim
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

// performVersionCheck returns the smoke version as a check
func performVersionCheck() Check {
	return Check{
		Name:    "Smoke Version",
		Status:  StatusPass,
		Message: Version,
		CanFix:  false,
	}
}

// checkVersion is kept for backward compatibility with tests
func checkVersion() Check {
	return performVersionCheck()
}

// runChecks executes all health checks and returns categories
func runChecks() []Category {
	return []Category{
		{
			Name: "INSTALLATION",
			Checks: []Check{
				performConfigDirCheck(),
				performFeedFileCheck(),
			},
		},
		{
			Name: "DATA",
			Checks: []Check{
				performFeedFormatCheck(),
				performConfigFileCheck(),
				performTUIConfigCheck(),
			},
		},
		{
			Name: "VERSION",
			Checks: []Check{
				performVersionCheck(),
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

// performConfigDirCheck verifies the config directory exists and is writable
func performConfigDirCheck() Check {
	configDir, err := config.GetConfigDir()
	if err != nil {
		return Check{
			Name:    "Config Directory",
			Status:  StatusFail,
			Message: "cannot determine config directory",
			Detail:  err.Error(),
			CanFix:  false,
		}
	}

	info, err := os.Stat(configDir)
	if os.IsNotExist(err) {
		return Check{
			Name:    "Config Directory",
			Status:  StatusFail,
			Message: "not found",
			Detail:  fmt.Sprintf("Run 'smoke doctor --fix' to create %s", configDir),
			CanFix:  true,
			Fix:     fixConfigDir,
		}
	}
	if err != nil {
		return Check{
			Name:    "Config Directory",
			Status:  StatusFail,
			Message: "cannot access",
			Detail:  err.Error(),
			CanFix:  false,
		}
	}
	if !info.IsDir() {
		return Check{
			Name:    "Config Directory",
			Status:  StatusFail,
			Message: "not a directory",
			Detail:  configDir,
			CanFix:  false,
		}
	}

	// Check if writable by creating a temp file
	testFile := filepath.Join(configDir, ".doctor-test")
	f, err := os.Create(testFile)
	if err != nil {
		return Check{
			Name:    "Config Directory",
			Status:  StatusWarn,
			Message: fmt.Sprintf("%s (not writable)", configDir),
			Detail:  "Permission denied - check directory permissions",
			CanFix:  false,
		}
	}
	_ = f.Close()
	_ = os.Remove(testFile)

	return Check{
		Name:    "Config Directory",
		Status:  StatusPass,
		Message: configDir,
		CanFix:  false,
	}
}

// checkConfigDir kept for backward compatibility with tests
func checkConfigDir() Check {
	return performConfigDirCheck()
}

// performFeedFileCheck verifies the feed file exists and is readable
func performFeedFileCheck() Check {
	feedPath, err := config.GetFeedPath()
	if err != nil {
		return Check{
			Name:    "Feed File",
			Status:  StatusFail,
			Message: "cannot determine feed path",
			Detail:  err.Error(),
			CanFix:  false,
		}
	}

	info, err := os.Stat(feedPath)
	if os.IsNotExist(err) {
		return Check{
			Name:    "Feed File",
			Status:  StatusFail,
			Message: "not found",
			Detail:  fmt.Sprintf("Run 'smoke doctor --fix' to create %s", feedPath),
			CanFix:  true,
			Fix:     fixFeedFile,
		}
	}
	if err != nil {
		return Check{
			Name:    "Feed File",
			Status:  StatusFail,
			Message: "cannot access",
			Detail:  err.Error(),
			CanFix:  false,
		}
	}
	if info.IsDir() {
		return Check{
			Name:    "Feed File",
			Status:  StatusFail,
			Message: "is a directory, expected file",
			Detail:  feedPath,
			CanFix:  false,
		}
	}

	// Check if readable
	f, err := os.Open(feedPath)
	if err != nil {
		return Check{
			Name:    "Feed File",
			Status:  StatusWarn,
			Message: fmt.Sprintf("%s (not readable)", feedPath),
			Detail:  "Permission denied - check file permissions",
			CanFix:  false,
		}
	}
	_ = f.Close()

	return Check{
		Name:    "Feed File",
		Status:  StatusPass,
		Message: feedPath,
		CanFix:  false,
	}
}

// checkFeedFile kept for backward compatibility with tests
func checkFeedFile() Check {
	return performFeedFileCheck()
}

// performFeedFormatCheck validates JSONL integrity of the feed file
func performFeedFormatCheck() Check {
	feedPath, err := config.GetFeedPath()
	if err != nil {
		return Check{
			Name:    "Feed Format",
			Status:  StatusFail,
			Message: "cannot determine feed path",
			CanFix:  false,
		}
	}

	f, err := os.Open(feedPath)
	if err != nil {
		return Check{
			Name:    "Feed Format",
			Status:  StatusFail,
			Message: "cannot open feed file",
			Detail:  err.Error(),
			CanFix:  false,
		}
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
		return Check{
			Name:    "Feed Format",
			Status:  StatusFail,
			Message: "error reading feed",
			Detail:  err.Error(),
			CanFix:  false,
		}
	}

	if totalLines == 0 {
		return Check{
			Name:    "Feed Format",
			Status:  StatusPass,
			Message: "empty (0 posts)",
			CanFix:  false,
		}
	}

	if invalidLines > 0 {
		return Check{
			Name:    "Feed Format",
			Status:  StatusWarn,
			Message: fmt.Sprintf("%d/%d lines valid", validLines, totalLines),
			Detail:  "Some lines contain invalid JSON - manual inspection recommended",
			CanFix:  false,
		}
	}

	return Check{
		Name:    "Feed Format",
		Status:  StatusPass,
		Message: fmt.Sprintf("%d posts, all valid", validLines),
		CanFix:  false,
	}
}

// checkFeedFormat kept for backward compatibility with tests
func checkFeedFormat() Check {
	return performFeedFormatCheck()
}

// performTUIConfigCheck verifies tui.yaml exists and has correct field names
func performTUIConfigCheck() Check {
	tuiPath, err := config.GetTUIConfigPath()
	if err != nil {
		return Check{
			Name:    "TUI Config",
			Status:  StatusFail,
			Message: "cannot determine TUI config path",
			Detail:  err.Error(),
			CanFix:  false,
		}
	}

	data, err := os.ReadFile(tuiPath)
	if os.IsNotExist(err) {
		// TUI config is optional, missing is fine
		return Check{
			Name:    "TUI Config",
			Status:  StatusPass,
			Message: "not present (using defaults)",
			CanFix:  false,
		}
	}
	if err != nil {
		return Check{
			Name:    "TUI Config",
			Status:  StatusFail,
			Message: "cannot read",
			Detail:  err.Error(),
			CanFix:  false,
		}
	}

	// Parse as generic map to check for deprecated fields
	var parsed map[string]interface{}
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		return Check{
			Name:    "TUI Config",
			Status:  StatusFail,
			Message: "invalid YAML",
			Detail:  err.Error(),
			CanFix:  false,
		}
	}

	// Check for deprecated "style" field (should be "layout")
	if _, hasStyle := parsed["style"]; hasStyle {
		if _, hasLayout := parsed["layout"]; !hasLayout {
			// Has style but no layout - needs migration
			return Check{
				Name:    "TUI Config",
				Status:  StatusWarn,
				Message: "deprecated 'style' field (should be 'layout')",
				Detail:  "Run 'smoke doctor --fix' to migrate to new field name",
				CanFix:  true,
				Fix: func() error {
					return fixTUIConfigStyleToLayout(tuiPath)
				},
			}
		}
	}

	return Check{
		Name:    "TUI Config",
		Status:  StatusPass,
		Message: tuiPath,
		CanFix:  false,
	}
}

// fixTUIConfigStyleToLayout migrates tui.yaml from "style" to "layout" field
func fixTUIConfigStyleToLayout(tuiPath string) error {
	data, err := os.ReadFile(tuiPath)
	if err != nil {
		return err
	}

	// Parse as generic map
	var parsed map[string]interface{}
	if err = yaml.Unmarshal(data, &parsed); err != nil {
		return err
	}

	// Migrate style -> layout
	if style, hasStyle := parsed["style"]; hasStyle {
		if _, hasLayout := parsed["layout"]; !hasLayout {
			parsed["layout"] = style
		}
		delete(parsed, "style")
	}

	// Write back
	newData, err := yaml.Marshal(parsed)
	if err != nil {
		return err
	}

	return os.WriteFile(tuiPath, newData, 0644)
}

// performConfigFileCheck verifies config.yaml exists and is valid YAML
func performConfigFileCheck() Check {
	configPath, err := config.GetConfigPath()
	if err != nil {
		return Check{
			Name:    "Config File",
			Status:  StatusFail,
			Message: "cannot determine config path",
			Detail:  err.Error(),
			CanFix:  false,
		}
	}

	data, err := os.ReadFile(configPath)
	if os.IsNotExist(err) {
		return Check{
			Name:    "Config File",
			Status:  StatusWarn,
			Message: "missing (using defaults)",
			Detail:  fmt.Sprintf("Run 'smoke doctor --fix' to create %s", configPath),
			CanFix:  true,
			Fix:     fixConfigFile,
		}
	}
	if err != nil {
		return Check{
			Name:    "Config File",
			Status:  StatusFail,
			Message: "cannot read",
			Detail:  err.Error(),
			CanFix:  false,
		}
	}

	// Validate YAML syntax
	var parsed interface{}
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		return Check{
			Name:    "Config File",
			Status:  StatusFail,
			Message: "invalid YAML",
			Detail:  err.Error(),
			CanFix:  false,
		}
	}

	return Check{
		Name:    "Config File",
		Status:  StatusPass,
		Message: configPath,
		CanFix:  false,
	}
}

// checkConfigFile kept for backward compatibility with tests
func checkConfigFile() Check {
	return performConfigFileCheck()
}
