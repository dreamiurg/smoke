package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

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

// FixResult contains the result of a fix operation
type FixResult struct {
	BackupPath  string // Path to backup file, empty if no backup created
	Description string // Human-readable description of what was fixed
}

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
	Detail  string                     // Optional additional info
	CanFix  bool                       // Whether --fix can repair this
	Fix     func() (*FixResult, error) // Fix function if CanFix is true
}

// Helper functions for creating Check structs

// passCheck creates a passing check with the given name and message
func passCheck(name, msg string) Check {
	return Check{Name: name, Status: StatusPass, Message: msg}
}

// warnCheck creates a warning check with the given name, message, and detail
func warnCheck(name, msg, detail string) Check {
	return Check{Name: name, Status: StatusWarn, Message: msg, Detail: detail}
}

// failCheck creates a failing check with the given name, message, detail, and optional fix
func failCheck(name, msg, detail string, canFix bool, fix func() (*FixResult, error)) Check {
	return Check{Name: name, Status: StatusFail, Message: msg, Detail: detail, CanFix: canFix, Fix: fix}
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
	const name = "Smoke Version"
	return passCheck(name, Version)
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
func fixConfigDir() (*FixResult, error) {
	configDir, err := config.GetConfigDir()
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, err
	}
	return &FixResult{Description: "Created config directory"}, nil
}

// fixFeedFile creates an empty feed file
func fixFeedFile() (*FixResult, error) {
	feedPath, err := config.GetFeedPath()
	if err != nil {
		return nil, err
	}
	f, err := os.Create(feedPath)
	if err != nil {
		return nil, err
	}
	if err := f.Close(); err != nil {
		return nil, err
	}
	return &FixResult{Description: "Created empty feed file"}, nil
}

// fixConfigFile creates a default config file
func fixConfigFile() (*FixResult, error) {
	configPath, err := config.GetConfigPath()
	if err != nil {
		return nil, err
	}
	defaultConfig := "# Smoke configuration\n# See: smoke explain\n"
	if err := os.WriteFile(configPath, []byte(defaultConfig), 0600); err != nil {
		return nil, err
	}
	return &FixResult{Description: "Created default config file"}, nil
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
					result, err := check.Fix()
					if err != nil {
						fmt.Printf("Failed to fix %s: %v\n", check.Name, err)
						continue
					}
					// Print backup path first if one was created
					if result != nil && result.BackupPath != "" {
						fmt.Printf("Backed up to: %s\n", result.BackupPath)
					}
					// Print what was fixed with description
					if result != nil && result.Description != "" {
						fmt.Printf("Fixed: %s (%s)\n", check.Name, result.Description)
					} else {
						fmt.Printf("Fixed: %s\n", check.Name)
					}
				}
				fixCount++
			}
		}
	}

	return fixCount, nil
}

// performConfigDirCheck verifies the config directory exists and is writable
func performConfigDirCheck() Check {
	const name = "Config Directory"
	configDir, err := config.GetConfigDir()
	if err != nil {
		return failCheck(name, "cannot determine config directory", err.Error(), false, nil)
	}

	info, err := os.Stat(configDir)
	if os.IsNotExist(err) {
		return failCheck(name, "not found", fmt.Sprintf("Run 'smoke doctor --fix' to create %s", configDir), true, fixConfigDir)
	}
	if err != nil {
		return failCheck(name, "cannot access", err.Error(), false, nil)
	}
	if !info.IsDir() {
		return failCheck(name, "not a directory", configDir, false, nil)
	}

	// Check if writable by creating a temp file
	testFile := filepath.Join(configDir, ".doctor-test")
	f, err := os.Create(testFile)
	if err != nil {
		return warnCheck(name, fmt.Sprintf("%s (not writable)", configDir), "Permission denied - check directory permissions")
	}
	_ = f.Close()
	_ = os.Remove(testFile)

	return passCheck(name, configDir)
}

// performFeedFileCheck verifies the feed file exists and is readable
func performFeedFileCheck() Check {
	const name = "Feed File"
	feedPath, err := config.GetFeedPath()
	if err != nil {
		return failCheck(name, "cannot determine feed path", err.Error(), false, nil)
	}

	info, err := os.Stat(feedPath)
	if os.IsNotExist(err) {
		return failCheck(name, "not found", fmt.Sprintf("Run 'smoke doctor --fix' to create %s", feedPath), true, fixFeedFile)
	}
	if err != nil {
		return failCheck(name, "cannot access", err.Error(), false, nil)
	}
	if info.IsDir() {
		return failCheck(name, "is a directory, expected file", feedPath, false, nil)
	}

	// Check if readable
	f, err := os.Open(feedPath)
	if err != nil {
		return warnCheck(name, fmt.Sprintf("%s (not readable)", feedPath), "Permission denied - check file permissions")
	}
	_ = f.Close()

	return passCheck(name, feedPath)
}

// performFeedFormatCheck validates JSONL integrity of the feed file
func performFeedFormatCheck() Check {
	const name = "Feed Format"
	feedPath, err := config.GetFeedPath()
	if err != nil {
		return failCheck(name, "cannot determine feed path", "", false, nil)
	}

	f, err := os.Open(feedPath)
	if err != nil {
		return failCheck(name, "cannot open feed file", err.Error(), false, nil)
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
		return failCheck(name, "error reading feed", err.Error(), false, nil)
	}

	if totalLines == 0 {
		return passCheck(name, "empty (0 posts)")
	}

	if invalidLines > 0 {
		return warnCheck(name, fmt.Sprintf("%d/%d lines valid", validLines, totalLines), "Some lines contain invalid JSON - manual inspection recommended")
	}

	if validLines == 1 {
		return passCheck(name, "1 post, valid")
	}
	return passCheck(name, fmt.Sprintf("%d posts, all valid", validLines))
}

// performTUIConfigCheck verifies tui.yaml exists and has correct field names
func performTUIConfigCheck() Check {
	const name = "TUI Config"
	tuiPath, err := config.GetTUIConfigPath()
	if err != nil {
		return failCheck(name, "cannot determine TUI config path", err.Error(), false, nil)
	}

	data, err := os.ReadFile(tuiPath)
	if os.IsNotExist(err) {
		// TUI config is optional, missing is fine
		return passCheck(name, "not present (using defaults)")
	}
	if err != nil {
		return failCheck(name, "cannot read", err.Error(), false, nil)
	}

	// Parse as generic map to check for deprecated fields
	var parsed map[string]interface{}
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		return failCheck(name, "invalid YAML", err.Error(), false, nil)
	}

	// Check for deprecated "style" field (should be "layout")
	if _, hasStyle := parsed["style"]; hasStyle {
		if _, hasLayout := parsed["layout"]; !hasLayout {
			// Has style but no layout - needs migration
			return Check{
				Name:    name,
				Status:  StatusWarn,
				Message: "deprecated 'style' field (should be 'layout')",
				Detail:  "Run 'smoke doctor --fix' to migrate to new field name",
				CanFix:  true,
				Fix: func() (*FixResult, error) {
					return fixTUIConfigStyleToLayout(tuiPath)
				},
			}
		}
	}

	return passCheck(name, tuiPath)
}

// backupTUIConfig creates a timestamped backup of tui.yaml if it exists.
// Returns the backup path if created, empty string if file doesn't exist.
func backupTUIConfig(tuiPath string) (string, error) {
	// Check if file exists
	if _, err := os.Stat(tuiPath); os.IsNotExist(err) {
		return "", nil
	}

	data, err := os.ReadFile(tuiPath)
	if err != nil {
		return "", err
	}

	// Create timestamped backup filename
	timestamp := time.Now().Format("2006-01-02T15-04-05")
	backupPath := fmt.Sprintf("%s.bak.%s", tuiPath, timestamp)

	if err := os.WriteFile(backupPath, data, 0600); err != nil {
		return "", err
	}

	return backupPath, nil
}

// fixTUIConfigStyleToLayout migrates tui.yaml from "style" to "layout" field
func fixTUIConfigStyleToLayout(tuiPath string) (*FixResult, error) {
	// Create backup before modifying
	backupPath, err := backupTUIConfig(tuiPath)
	if err != nil {
		return nil, fmt.Errorf("backup tui config: %w", err)
	}

	data, err := os.ReadFile(tuiPath)
	if err != nil {
		return nil, err
	}

	// Parse as generic map
	var parsed map[string]interface{}
	if err = yaml.Unmarshal(data, &parsed); err != nil {
		return nil, err
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
		return nil, err
	}

	if err := os.WriteFile(tuiPath, newData, 0600); err != nil {
		return nil, err
	}

	result := &FixResult{
		Description: "Migrated 'style' field to 'layout'",
		BackupPath:  backupPath,
	}
	return result, nil
}

// performConfigFileCheck verifies config.yaml exists and is valid YAML
func performConfigFileCheck() Check {
	const name = "Config File"
	configPath, err := config.GetConfigPath()
	if err != nil {
		return failCheck(name, "cannot determine config path", err.Error(), false, nil)
	}

	data, err := os.ReadFile(configPath)
	if os.IsNotExist(err) {
		return Check{
			Name:    name,
			Status:  StatusWarn,
			Message: "missing (using defaults)",
			Detail:  fmt.Sprintf("Run 'smoke doctor --fix' to create %s", configPath),
			CanFix:  true,
			Fix:     fixConfigFile,
		}
	}
	if err != nil {
		return failCheck(name, "cannot read", err.Error(), false, nil)
	}

	// Validate YAML syntax
	var parsed interface{}
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		return failCheck(name, "invalid YAML", err.Error(), false, nil)
	}

	return passCheck(name, configPath)
}
