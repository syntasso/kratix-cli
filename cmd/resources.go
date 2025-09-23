/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// platformResourcesCmd represents the resources command
var platformResourcesCmd = &cobra.Command{
	Use:   "resources ",
	Short: "A command to fetch resource requests of a given Promise",
	Long: `A command to fetch resource requests of a given Promise
	
	For Compound Promise, it details all of the requests that make up a Compound request.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("resources called")
	},
}

func init() {
	platformGetCmd.AddCommand(platformResourcesCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// resourcesCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// resourcesCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
