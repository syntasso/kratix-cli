package cmd

import (
	"fmt"
	"io"
	"os"
	"path"
	"regexp"

	"github.com/spf13/cobra"
)

var testContainerAddCmd = &cobra.Command{
	Use:   "add LIFECYCLE/ACTION/PIPELINE-NAME/CONTAINER-NAME --testcase TESTCASE-NAME",
	Short: "Adds a new testcase for the Kratix container image",
	Example: `  # add a container testcase directory for a given image
  > kratix test container add resource/configure/instance/syntasso-postgres-resource \
     --testcase handles_empty_metadata`,
	RunE: TestContainerAdd,
	Args: cobra.ExactArgs(1),
}

var (
	inputObject, testcaseName string
)

func init() {
	testContainerCmd.AddCommand(testContainerAddCmd)

	testContainerAddCmd.Flags().StringVarP(&inputObject, "input-object", "o", "", "The path to the input object to use for this testcase")
	testContainerAddCmd.Flags().StringVarP(&testcaseName, "testcase", "t", "", "The name of the testcase to add")

	testContainerAddCmd.MarkFlagRequired("testcase")
}

func TestContainerAdd(cmd *cobra.Command, args []string) error {
	var err error

	pipelineInput := args[0]
	containerArgs, err := ParseContainerCmdArgs(pipelineInput, 4)
	if err != nil {
		return err
	}

	imageTestDir, err := getImageTestDir(containerArgs)
	if err != nil {
		return err
	}

	var testcaseDir string
	if testcaseDir, err = validateTestcaseName(testcaseName, imageTestDir); err != nil {
		return err
	}

	beforeDir := path.Join(testcaseDir, "before")
	afterDir := path.Join(testcaseDir, "after")

	dirsToCreate := []string{
		path.Join(beforeDir, "metadata"),
		path.Join(beforeDir, "input"),
		path.Join(beforeDir, "output"),
		path.Join(afterDir, "metadata"),
		path.Join(afterDir, "input"),
		path.Join(afterDir, "output"),
	}

	for _, dir := range dirsToCreate {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	if inputObject != "" {
		err := copyFile(inputObject, path.Join(beforeDir, "input", "object.yaml"))
		if err != nil {
			return err
		}
	} else {
		var objectFile string
		if containerArgs.Lifecycle == "resource" {
			objectFile = path.Join(dir, "example-resource.yaml")
		} else {
			objectFile = path.Join(dir, "promise.yaml")
		}
		fmt.Printf("No input object provided, copying %s\n", objectFile)
		err = copyFile(objectFile, path.Join(beforeDir, "input", "object.yaml"))
		if err != nil {
			return err
		}
	}

	fmt.Printf("Testcase %s added successfully! ✅\n\n", testcaseName)
	fmt.Printf("Customise your testcase by editing the files in %s\n", testcaseDir)

	return nil
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destinationFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		return err
	}

	err = destinationFile.Sync()
	if err != nil {
		return err
	}

	return nil
}

func validateTestcaseName(testcaseName, imageDir string) (string, error) {
	testcaseDir := path.Join(imageDir, testcaseName)

	if testcaseName == "" {
		return "", fmt.Errorf("testcase name cannot be empty")
	}

	if _, err := os.Stat(testcaseDir); err == nil {
		return "", fmt.Errorf("testcase directory already exists: %s", testcaseDir)
	}

	if !regexp.MustCompile(`^[a-zA-Z0-9_-]+$`).MatchString(testcaseName) {
		return "", fmt.Errorf("invalid testcase name: %s, only alphanumeric characters, hyphens, and underscores are allowed", testcaseName)
	}

	return testcaseDir, nil
}

func getImageTestDir(containerArgs *ContainerCmdArgs) (string, error) {
	containerDir := getContainerDir(containerArgs)

	if _, err := os.Stat(containerDir); os.IsNotExist(err) {
		return "", fmt.Errorf("container directory does not exist: %s", containerDir)
	}

	return path.Join(containerDir, "test"), nil
}

func getContainerDir(containerArgs *ContainerCmdArgs) string {
	return path.Join(outputDir, "workflows", containerArgs.Lifecycle, containerArgs.Action, containerArgs.Pipeline, containerArgs.Container)
}
