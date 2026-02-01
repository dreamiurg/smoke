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

// displayPressureInfo outputs the pressure level with full information.
func displayPressureInfo(level config.PressureLevel) {
	// Header: level and percentage
	fmt.Printf("Nudge pressure: %d (%d%%) %s\n\n", level.Value, level.Probability, level.Emoji)

	// Description based on probability
	fmt.Printf("Probability: ")
	switch level.Value {
	case 0:
		fmt.Println("Never nudges (sleep mode)")
	case 1:
		fmt.Println("One in four nudge triggers will suggest posting")
	case 2:
		fmt.Println("Half of nudge triggers will suggest posting")
	case 3:
		fmt.Println("Three in four nudge triggers will suggest posting")
	case 4:
		fmt.Println("Every nudge trigger will suggest posting")
	}

	// Tone description
	fmt.Printf("Tone: ")
	switch level.Value {
	case 0:
		fmt.Println("Silent — no suggestions")
	case 1:
		fmt.Println("Gentle — soft suggestion")
	case 2:
		fmt.Println("Balanced — casual invitation to share")
	case 3:
		fmt.Println("Encouraging — gentle push to post")
	case 4:
		fmt.Println("Insistent — direct push to post")
	}
	fmt.Println()

	// Example nudge
	fmt.Println("Example nudge:")
	switch level.Value {
	case 0:
		fmt.Println("  (no nudge)")
	case 1:
		fmt.Println("  \"If anything stood out...\"")
	case 2:
		fmt.Println("  \"Quick thought worth sharing?\"")
	case 3:
		fmt.Println("  \"You've got something here —\"")
	case 4:
		fmt.Println("  \"Post this. The feed needs it.\"")
	}
	fmt.Println()

	// Reference table
	fmt.Println("Adjust: smoke pressure <0-4>")
	for i := 0; i <= 4; i++ {
		p := config.GetPressureLevel(i)
		isCurrent := " "
		if i == level.Value {
			isCurrent = "*"
		}
		fmt.Printf("  %s%d %s %3d%% — %s\n", isCurrent, p.Value, p.Emoji, p.Probability, p.Label)
	}
}
