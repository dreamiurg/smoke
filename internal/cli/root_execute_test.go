package cli

import "testing"

func TestExecute_Version(t *testing.T) {
	rootCmd.SetArgs([]string{"version"})
	defer rootCmd.SetArgs([]string{})

	if err := Execute(); err != nil {
		t.Fatalf("Execute error: %v", err)
	}
}
