package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/dreamiurg/smoke/internal/config"
	"github.com/dreamiurg/smoke/internal/logging"
)

var (
	logsLines int
	logsTail  bool
	logsClear bool
)

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "View smoke log file",
	Long: `View or manage the smoke log file.

The log file is located at ~/.config/smoke/smoke.log and contains
structured JSON entries for debugging and operational visibility.

Examples:
  smoke logs              Show last 50 lines
  smoke logs -n 100       Show last 100 lines
  smoke logs --tail       Follow log output (like tail -f)
  smoke logs --clear      Clear the log file`,
	Args: cobra.NoArgs,
	RunE: runLogs,
}

func init() {
	logsCmd.Flags().IntVarP(&logsLines, "lines", "n", 50, "Number of lines to show")
	logsCmd.Flags().BoolVarP(&logsTail, "tail", "f", false, "Follow log output")
	logsCmd.Flags().BoolVar(&logsClear, "clear", false, "Clear the log file")
	rootCmd.AddCommand(logsCmd)
}

func runLogs(_ *cobra.Command, args []string) error {
	logging.LogCommand("logs", args)

	logPath, err := config.GetLogPath()
	if err != nil {
		return fmt.Errorf("failed to get log path: %w", err)
	}

	// Handle --clear flag
	if logsClear {
		return clearLogFile(logPath)
	}

	// Check if log file exists
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		fmt.Println("No log file found. Logs are created when smoke commands run.")
		fmt.Printf("Log path: %s\n", logPath)
		return nil
	}

	if logsTail {
		return tailLogFile(logPath)
	}

	return showLogFile(logPath, logsLines)
}

// showLogFile displays the last n lines of the log file
func showLogFile(path string, lines int) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer func() { _ = file.Close() }()

	// Read all lines into a buffer
	var allLines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		allLines = append(allLines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read log file: %w", err)
	}

	// Get last N lines
	start := 0
	if len(allLines) > lines {
		start = len(allLines) - lines
	}

	for _, line := range allLines[start:] {
		fmt.Println(line)
	}

	return nil
}

// tailLogFile follows the log file for new entries
func tailLogFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer func() { _ = file.Close() }()

	// Seek to end
	_, err = file.Seek(0, io.SeekEnd)
	if err != nil {
		return fmt.Errorf("failed to seek to end: %w", err)
	}

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	fmt.Printf("Following %s (Ctrl+C to stop)\n", path)

	reader := bufio.NewReader(file)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-sigChan:
			fmt.Println()
			return nil
		case <-ticker.C:
			for {
				line, err := reader.ReadString('\n')
				if err != nil {
					break
				}
				fmt.Print(line)
			}
		}
	}
}

// clearLogFile truncates the log file
func clearLogFile(path string) error {
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Println("No log file to clear.")
		return nil
	}

	// Truncate the file
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to clear log file: %w", err)
	}
	_ = file.Close()

	fmt.Println("Log file cleared.")
	return nil
}
