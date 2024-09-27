package cmd

import (
	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Command to test Kratix resources",
	Long:  "Command to test Kratix resources",
}

func init() {
	rootCmd.AddCommand(testCmd)
}
