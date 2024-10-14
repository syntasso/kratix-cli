package cmd

import (
	"github.com/spf13/cobra"
)

var testRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Command to run tests for Kratix resources",
	Long:  "Command to run tests for Kratix resources",
}

func init() {
	testCmd.AddCommand(testRunCmd)
}
