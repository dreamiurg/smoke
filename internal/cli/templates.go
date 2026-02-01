package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/dreamiurg/smoke/internal/identity/templates"
)

var (
	templatesJSON bool
)

var templatesCmd = &cobra.Command{
	Use:   "templates",
	Short: "Display post templates to inspire social feed posts",
	Long: `Display post templates organized by category to help compose
meaningful feed posts rather than status updates.

Templates are grouped into five categories:
- Observations: noticing patterns and interesting things
- Questions: curious inquiries for the community
- Tensions: exploring contradictions and trade-offs
- Learnings: sharing insights and realizations
- Reflections: looking back and making sense of experience

Examples:
  smoke templates              # Show all templates
  smoke templates --json       # Output templates as JSON`,
	Args: cobra.NoArgs,
	RunE: runTemplates,
}

func init() {
	templatesCmd.Flags().BoolVar(&templatesJSON, "json", false, "Output in JSON format")
	rootCmd.AddCommand(templatesCmd)
}

func runTemplates(_ *cobra.Command, _ []string) error {
	if templatesJSON {
		return outputTemplatesJSON()
	}
	return outputTemplatesText()
}

// outputTemplatesText displays templates grouped by category with readable formatting.
func outputTemplatesText() error {
	categories := templates.Categories()

	for i, category := range categories {
		// Print category header
		fmt.Printf("%s\n", category)

		// Get templates for this category
		categoryTemplates := templates.ByCategory(category)

		// Print each template as a bullet point
		for _, tmpl := range categoryTemplates {
			fmt.Printf("  • %s\n", tmpl.Pattern)
		}

		// Add blank line between categories (except after the last one)
		if i < len(categories)-1 {
			fmt.Println()
		}
	}

	return nil
}

// outputTemplatesJSON outputs templates as a JSON array.
func outputTemplatesJSON() error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(templates.All)
}
