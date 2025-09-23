/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

// platformGetCmd represents the get command
var platformGetCmd = &cobra.Command{
	Use:   "platform",
	Short: "A command to display resources in the deployed Kratix",
	Long:  `A command to display resources in the deployed Kratix`,
}

func init() {
	platformCmd.AddCommand(platformGetCmd)
}
