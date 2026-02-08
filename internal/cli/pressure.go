package cli

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/dreamiurg/smoke/internal/config"
	"github.com/dreamiurg/smoke/internal/logging"
)

var pressureCmd = &cobra.Command{
	Use:   "pressure [level]",
	Short: "View or set the nudge pressure level (0-4)",
	Long: `View or set the nudge pressure level.

Pressure controls how often nudges trigger:
  0 💤   0% — sleep, no nudges
  1 🌙  25% — quiet
  2 ⛅  50% — balanced (default)
  3 ☀️  75% — bright
  4 🌋 100% — volcanic

Examples:
  smoke pressure         # View current pressure
  smoke pressure 3       # Set to 3 (75% - bright)`,
	RunE: runPressure,
}

func init() {
	rootCmd.AddCommand(pressureCmd)
}

func runPressure(_ *cobra.Command, args []string) error {
	// Start command tracking
	tracker := logging.StartCommand("pressure", args)

	// Check if smoke is initialized
	if err := config.EnsureInitialized(); err != nil {
		tracker.Fail(err)
		return err
	}

	// Handle setting pressure
	if len(args) > 0 {
		levelStr := args[0]
		level, err := strconv.Atoi(levelStr)
		if err != nil {
			err = fmt.Errorf("invalid pressure level: must be a number 0-4")
			tracker.Fail(err)
			return err
		}

		if level < 0 || level > 4 {
			err = fmt.Errorf("pressure level out of range: must be 0-4 (got %d)", level)
			tracker.Fail(err)
			return err
		}

		if err := config.SetPressure(level); err != nil {
			tracker.Fail(err)
			return err
		}
	}

	// Display current pressure
	pressure := config.GetPressure()
	pressureLevel := config.GetPressureLevel(pressure)

	// Format and display output
	displayPressureInfo(pressureLevel)

	tracker.Complete()
	return nil
}

// pressureDescriptions maps pressure levels to probability descriptions.
var pressureDescriptions = map[int]string{
	0: "Never nudges (sleep mode)",
	1: "One in four nudge triggers will suggest posting",
	2: "Half of nudge triggers will suggest posting",
	3: "Three in four nudge triggers will suggest posting",
	4: "Every nudge trigger will suggest posting",
}

// pressureTones maps pressure levels to tone descriptions.
var pressureTones = map[int]string{
	0: "Silent — no suggestions",
	1: "Gentle — soft suggestion",
	2: "Balanced — casual invitation to share",
	3: "Encouraging — gentle push to post",
	4: "Insistent — direct push to post",
}

// pressureExamples maps pressure levels to example nudge text.
var pressureExamples = map[int]string{
	0: "  (no nudge)",
	1: "  \"If anything stood out...\"",
	2: "  \"Quick thought worth sharing?\"",
	3: "  \"You've got something here —\"",
	4: "  \"Post this. The feed needs it.\"",
}

// displayPressureInfo outputs the pressure level with full information.
func displayPressureInfo(level config.PressureLevel) {
	fmt.Printf("Nudge pressure: %d (%d%%) %s\n\n", level.Value, level.Probability, level.Emoji)

	fmt.Printf("Probability: %s\n", pressureDescriptions[level.Value])
	fmt.Printf("Tone: %s\n\n", pressureTones[level.Value])

	fmt.Println("Example nudge:")
	fmt.Println(pressureExamples[level.Value])
	fmt.Println()

	fmt.Println("Adjust: smoke pressure <0-4>")
	for i := 0; i <= 4; i++ {
		p := config.GetPressureLevel(i)
		marker := " "
		if i == level.Value {
			marker = "*"
		}
		fmt.Printf("  %s%d %s %3d%% — %s\n", marker, p.Value, p.Emoji, p.Probability, p.Label)
	}
}
