package cmd

import (
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Command to update kratix resources",
	Long:  "Command to update kratix resources",
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
