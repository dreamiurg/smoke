package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/dreamiurg/smoke/internal/config"
)

var (
	whoamiJSON bool
	whoamiName bool
)

var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Print the current identity",
	Long: `Print the current identity.

By default, outputs the full identity in name@project format.
The identity is resolved from SMOKE_NAME environment variable,
or auto-detected from the session.

Examples:
  smoke whoami                  # Output: swift-fox@smoke
  smoke whoami --name           # Output: swift-fox
  smoke whoami --json           # Output: {"name":"swift-fox","agent":"claude","project":"smoke"}`,
	Args: cobra.NoArgs,
	RunE: runWhoami,
}

func init() {
	whoamiCmd.Flags().BoolVar(&whoamiJSON, "json", false, "Output in JSON format")
	whoamiCmd.Flags().BoolVar(&whoamiName, "name", false, "Output name only (without project)")
	rootCmd.AddCommand(whoamiCmd)
}

func runWhoami(_ *cobra.Command, _ []string) error {
	// Get identity
	identity, err := config.GetIdentity("")
	if err != nil {
		return err
	}

	if whoamiJSON {
		output := map[string]string{
			"name":    identity.Suffix,
			"project": identity.Project,
		}
		if identity.Agent != "" {
			output["agent"] = identity.Agent
		}
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(output)
	}

	if whoamiName {
		fmt.Println(identity.Suffix)
		return nil
	}

	// Default: suffix@project
	fmt.Println(identity.String())
	return nil
}
