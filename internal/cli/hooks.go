package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/dreamiurg/smoke/internal/hooks"
)

var (
	hooksForce      bool
	hooksStatusJSON bool
)

var hooksCmd = &cobra.Command{
	Use:   "hooks",
	Short: "Manage Claude Code hook integration",
	Long: `Install, uninstall, or check status of smoke's Claude Code hooks.

Smoke uses Claude Code hooks to nudge agents to post during natural pauses.
This happens automatically during 'smoke init', but these commands allow
manual control for repairs, customization, or opt-out.`,
}

var hooksInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install or repair Claude Code hooks",
	Long: `Install smoke hooks to Claude Code.

This command:
  1. Copies hook scripts to ~/.claude/hooks/
  2. Updates ~/.claude/settings.json to register hooks
  3. Preserves any existing non-smoke hooks

By default, installation fails if hook scripts have been modified.
Use --force to overwrite customizations.`,
	Example: `  smoke hooks install         Install hooks (fail if modified)
  smoke hooks install --force  Overwrite any modifications`,
	RunE: runHooksInstall,
}

var hooksUninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Remove smoke hooks from Claude Code",
	Long: `Remove smoke hooks from Claude Code.

This command:
  1. Removes hook entries from ~/.claude/settings.json
  2. Deletes hook script files from ~/.claude/hooks/
  3. Preserves other non-smoke hooks`,
	Example: `  smoke hooks uninstall`,
	RunE:    runHooksUninstall,
}

var hooksStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show hook installation status",
	Long: `Show current status of smoke hook installation.

Displays:
  - Overall installation state
  - Per-script status (ok, missing, modified)
  - Settings.json configuration status

Use --json for machine-readable output.`,
	Example: `  smoke hooks status       Human-readable status
  smoke hooks status --json JSON output`,
	RunE: runHooksStatus,
}

func init() {
	hooksInstallCmd.Flags().BoolVarP(&hooksForce, "force", "f", false, "Overwrite modified hook scripts")
	hooksStatusCmd.Flags().BoolVarP(&hooksStatusJSON, "json", "j", false, "Output as JSON")

	hooksCmd.AddCommand(hooksInstallCmd)
	hooksCmd.AddCommand(hooksUninstallCmd)
	hooksCmd.AddCommand(hooksStatusCmd)

	rootCmd.AddCommand(hooksCmd)
}

func runHooksInstall(_ *cobra.Command, _ []string) error {
	opts := hooks.InstallOptions{
		Force: hooksForce,
	}

	result, err := hooks.Install(opts)
	if err != nil {
		if errors.Is(err, hooks.ErrScriptsModified) {
			fmt.Fprintln(os.Stderr, "Error: Hook scripts have been modified")
			fmt.Fprintln(os.Stderr, "Use --force to overwrite or update manually.")
			return nil // Don't return error to avoid double error message
		}
		if errors.Is(err, hooks.ErrPermissionDenied) {
			fmt.Fprintln(os.Stderr, "Error: Permission denied")
			fmt.Fprintln(os.Stderr, "Check directory permissions or run as appropriate user.")
			return nil
		}
		if errors.Is(err, hooks.ErrInvalidSettings) {
			fmt.Fprintln(os.Stderr, "Error: ~/.claude/settings.json contains invalid JSON")
			fmt.Fprintln(os.Stderr, "Settings backed up. Hooks installed with fresh settings.")
		} else {
			return fmt.Errorf("install hooks: %w", err)
		}
	}

	// Print backup path if created
	if result != nil && result.BackupPath != "" {
		fmt.Printf("Backed up settings to: %s\n", result.BackupPath)
	}

	// Show success message
	fmt.Println("Installed smoke hooks:")
	for _, script := range hooks.ListScripts() {
		scriptPath := hooks.GetHooksDir() + "/" + script.Name
		fmt.Printf("  %s (%s)\n", scriptPath, script.Event)
	}
	fmt.Printf("Updated %s\n", hooks.GetSettingsPath())

	return nil
}

func runHooksUninstall(_ *cobra.Command, _ []string) error {
	// Check if hooks are installed
	status, err := hooks.GetStatus()
	if err != nil {
		return fmt.Errorf("check status: %w", err)
	}

	if status.State == hooks.StateNotInstalled {
		fmt.Println("Smoke hooks are not installed.")
		return nil
	}

	// Uninstall
	result, err := hooks.Uninstall()
	if err != nil {
		if errors.Is(err, hooks.ErrPermissionDenied) {
			fmt.Fprintln(os.Stderr, "Error: Permission denied")
			fmt.Fprintln(os.Stderr, "Check directory permissions.")
			return nil
		}
		return fmt.Errorf("uninstall hooks: %w", err)
	}

	// Print backup path if created
	if result != nil && result.BackupPath != "" {
		fmt.Printf("Backed up settings to: %s\n", result.BackupPath)
	}

	// Show success message
	fmt.Println("Removed smoke hooks:")
	for _, script := range hooks.ListScripts() {
		scriptPath := hooks.GetHooksDir() + "/" + script.Name
		fmt.Printf("  %s\n", scriptPath)
	}
	fmt.Printf("Updated %s\n", hooks.GetSettingsPath())

	return nil
}

func runHooksStatus(_ *cobra.Command, _ []string) error {
	status, err := hooks.GetStatus()
	if err != nil {
		return fmt.Errorf("get status: %w", err)
	}

	if hooksStatusJSON {
		// JSON output
		type jsonScriptInfo struct {
			Path     string `json:"path"`
			Exists   bool   `json:"exists"`
			Modified bool   `json:"modified"`
		}

		type jsonStatus struct {
			Status   string                    `json:"status"`
			Scripts  map[string]jsonScriptInfo `json:"scripts"`
			Settings struct {
				Stop        bool `json:"stop"`
				PostToolUse bool `json:"postToolUse"`
			} `json:"settings"`
		}

		output := jsonStatus{
			Status:  string(status.State),
			Scripts: make(map[string]jsonScriptInfo),
		}
		output.Settings.Stop = status.Settings.Stop
		output.Settings.PostToolUse = status.Settings.PostToolUse

		for name, info := range status.Scripts {
			output.Scripts[name] = jsonScriptInfo{
				Path:     info.Path,
				Exists:   info.Exists,
				Modified: info.Modified,
			}
		}

		data, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal JSON: %w", err)
		}
		fmt.Println(string(data))
		return nil
	}

	// Human-readable output
	fmt.Printf("Status: %s\n", status.State)
	fmt.Println()

	// Show script status
	fmt.Println("Scripts:")
	for _, script := range hooks.ListScripts() {
		info, ok := status.Scripts[script.Name]
		if !ok {
			continue
		}
		fmt.Printf("  %s: %s\n", info.Path, info.Status)
	}
	fmt.Println()

	// Show settings status
	fmt.Println("Settings:")
	if status.Settings.Stop {
		fmt.Println("  Stop hook: configured")
	} else {
		fmt.Println("  Stop hook: missing")
	}
	if status.Settings.PostToolUse {
		fmt.Println("  PostToolUse hook: configured")
	} else {
		fmt.Println("  PostToolUse hook: missing")
	}
	fmt.Println()

	// Show actionable suggestions
	switch status.State {
	case hooks.StateNotInstalled:
		fmt.Println("To install: smoke hooks install")
	case hooks.StatePartiallyInstalled:
		fmt.Println("To repair: smoke hooks install")
	case hooks.StateModified:
		fmt.Println("Scripts have been customized. Use --force to overwrite.")
	}

	return nil
}
