package cmd

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/spf13/cobra"
)

var testContainerRunCmd = &cobra.Command{
	Use:   "run LIFECYCLE/ACTION/PIPELINE-NAME/CONTAINER-NAME",
	Short: "Run tests for Kratix container images",
	Example: `  # run all testcases for a container image
  kratix test container run resource/configure/instance/syntasso-postgres-resource

  # run specific testcases for a container image
  kratix test container run resource/configure/instance/syntasso-postgres-resource --testcases test1,test2,test3`,
	RunE: TestContainerRun,
	Args: cobra.ExactArgs(1),
}

var testcaseNames, command string

func init() {
	testContainerCmd.AddCommand(testContainerRunCmd)
	testContainerRunCmd.Flags().StringVarP(&testcaseNames, "testcases", "t", "", "Comma-separated list of testcases to run")
	testContainerRunCmd.Flags().StringVarP(&command, "command", "c", "", "Command to start the image with")
}

func TestContainerRun(cmd *cobra.Command, args []string) error {
	if testcaseNames == "" {
		fmt.Println("Running all testcases...")
	} else {
		fmt.Printf("Running testcases: %s\n", testcaseNames)
	}

	pipelineInput := args[0]
	containerArgs, err := ParseContainerCmdArgs(pipelineInput, 4)
	if err != nil {
		return err
	}

	imageTestDir, err := getImageTestDir(containerArgs)
	if err != nil {
		return err
	}

	testcaseDirs, err := getTestcaseDirs(imageTestDir, testcaseNames)
	if err != nil {
		return err
	}

	for _, testcaseDir := range testcaseDirs {
		fmt.Printf("Running testcase: %s...", path.Base(testcaseDir))

		err = runTestcase(testcaseDir)
		if err != nil {
			return err
		}

		fmt.Println(" ✅")
	}

	return nil
}

func getTestcaseDirs(imageDir, testcaseNames string) ([]string, error) {
	if testcaseNames == "" {
		return getDirs(imageDir)
	}

	testcaseNamesList := strings.Split(testcaseNames, ",")
	testcaseDirs := make([]string, 0, len(testcaseNamesList))

	for _, testcaseName := range testcaseNamesList {
		testcaseDir := path.Join(imageDir, testcaseName)
		if _, err := os.Stat(testcaseDir); os.IsNotExist(err) {
			return nil, fmt.Errorf("testcase directory %q does not exist", testcaseDir)
		}
		testcaseDirs = append(testcaseDirs, testcaseDir)
	}

	return testcaseDirs, nil
}

func getDirs(dir string) ([]string, error) {
	var dirs []string

	directories, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, d := range directories {
		if d.IsDir() {
			dirs = append(dirs, path.Join(dir, d.Name()))
		}
	}

	return dirs, nil
}

func runTestcase(testcaseDir string) error {
	// Placeholder for running the testcases
	// This function should be replaced with the actual test runner

	/*
		Something like:

		docker run \
		--rm \
		--volume ~/.kube:/root/.kube \
		--network=host \
		--volume ./test/output:/kratix/output \
		--volume ./test/input:/kratix/input \
		--volume ./test/metadata:/kratix/metadata \
		kratix-workshop/app-promise-pipeline:v0.1.0

		See: https://docs.kratix.io/workshop/part-ii/promise-workflows#test-driving-your-workflows
	*/
	return nil
}
