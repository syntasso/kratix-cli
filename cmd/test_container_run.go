package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var testContainerRunCmd = &cobra.Command{
	Use:   "run --image CONTAINER-IMAGE",
	Short: "Run tests for Kratix container images",
	Example: `  # run all testcases for a container image
  kratix test container run --image my-image

  # run specific testcases for a container image
  kratix test container run --image my-image --testcases test1,test2,test3`,
	RunE: TestContainerRun,
	Args: cobra.ExactArgs(0),
}

var testcaseNames string

func init() {
	testContainerCmd.AddCommand(testContainerRunCmd)
	testContainerRunCmd.Flags().StringVarP(&testcaseNames, "testcases", "t", "", "Comma-separated list of testcases to run")
}

func TestContainerRun(cmd *cobra.Command, args []string) error {
	testcases := strings.Split(testcaseNames, ",")
	fmt.Println("TestContainerRun: " + strings.Join(testcases, ", "))

	return nil
}
