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

	testCmd.PersistentFlags().StringVarP(&testImage, "image", "i", "", "The image used by this container")
	testCmd.PersistentFlags().StringVarP(&testcaseDir, "dir", "d", ".", "Directory containing testcases")

	testCmd.MarkPersistentFlagRequired("image")
}
