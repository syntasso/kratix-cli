package cmd

import (
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Command used to initialize Kratix resources",
	Long:  `Command used to initialize Kratix resources"`,
}

func init() {
	rootCmd.AddCommand(initCmd)
}
