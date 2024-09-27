package cmd

import (
	"fmt"
	"io"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

var testContainerAddCmd = &cobra.Command{
	Use:   "add --image CONTAINER-IMAGE TESTCASE-NAME",
	Short: "Adds a new testcase for the Kratix container",
	Example: `  # add a container testcase directory for a given image
  > kratix test container add \
     --image ghcr.io/syntasso/my-image:v0.1.0
     --testcase handles_empty_metadata`,
	RunE: TestContainerAdd,
	Args: cobra.ExactArgs(1),
}

var (
	inputObject string
)

func init() {
	testContainerCmd.AddCommand(testContainerAddCmd)

	testContainerAddCmd.Flags().StringVarP(&inputObject, "input-object", "o", "", "The path to the input object to use for this testcase")

	testContainerAddCmd.MarkFlagRequired("testcase")
}

func TestContainerAdd(cmd *cobra.Command, args []string) error {
	testcaseName := args[0]

	imageName := strings.Split(testImage, ":")[0]
	formattedImageName := strings.ReplaceAll(imageName, "/", "_")
	formattedImageName = strings.ReplaceAll(formattedImageName, ".", "_")

	imageDir := path.Join(testcaseDir, formattedImageName)

	// validate testcase name
	var testcaseDir string
	var err error
	if testcaseDir, err = validateTestcaseName(testcaseName, imageDir); err != nil {
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
	}

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
