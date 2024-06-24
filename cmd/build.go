package cmd

import (
	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Command to build kratix resources",
	Long:  "Command to build kratix resources",
}

func init() {
	rootCmd.AddCommand(buildCmd)
}
