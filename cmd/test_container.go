package cmd

import (
	"github.com/spf13/cobra"
)

var testContainerCmd = &cobra.Command{
	Use:   "container",
	Short: "Command to test Kratix container images",
	Long:  "Command to test Kratix container images",
}

var (
	testImage   string
	testcaseDir string
)

func init() {
	testCmd.AddCommand(testContainerCmd)
}
