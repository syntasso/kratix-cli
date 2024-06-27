package cmd

import (
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Command to add to Kratix resources",
	Long:  "Command to add to Kratix resources",
}

func init() {
	rootCmd.AddCommand(addCmd)
}
