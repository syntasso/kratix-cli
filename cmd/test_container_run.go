package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var testContainerRunCmd = &cobra.Command{
	Use:   "run LIFECYCLE/ACTION/PIPELINE-NAME/CONTAINER-NAME",
	Short: "Run tests for Kratix container images (docker only)",
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

	imageName, err := buildImage(containerArgs)
	if err != nil {
		return err
	}

	for _, testcaseDir := range testcaseDirs {
		fmt.Printf("Running testcase: %s...\n", path.Base(testcaseDir))

		err = runTestcase(testcaseDir, imageName)
		if err != nil {
			return err
		}

		fmt.Println("Testcase passed ✅")
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

func runTestcase(testcaseDir, image string) error {
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

	// Steps:
	// 1. Copy the before/ files to a temporary directory
	// 2. Run the container image, mounting the temporary directory
	// 3. Check the temporary directory contents against the after/ files (expected output)

	// Copy the before/ files to a temporary directory
	beforeDir := path.Join(testcaseDir, "before")
	// get a tempdir in /tmp
	tmpdir := path.Join(os.TempDir(), fmt.Sprintf("kratix-test-%s-%d", path.Base(testcaseDir), time.Now().Unix()))
	err := os.MkdirAll(tmpdir, os.ModePerm)
	if err != nil {
		return err
	}

	fmt.Printf("Copying before/ files to temporary directory %s...\n", tmpdir)

	// copy the before/ files to the tempdir
	err = copyDir(beforeDir, tmpdir)
	if err != nil {
		return err
	}

	// Run the container image, mounting the temporary directory
	// TODO: Extract into a function
	cmd := fmt.Sprintf(
		"docker run --rm --volume ~/.kube:/root/.kube --network=host --volume %s:/kratix/output --volume %s:/kratix/input --volume %s:/kratix/metadata %s",
		path.Join(tmpdir, "output"), path.Join(tmpdir, "input"), path.Join(tmpdir, "metadata"),
		image,
	)
	runner := exec.Command(cmd)

	// TODO: Maybe don't show this by default? --verbose flag?
	runner.Stdout = os.Stdout
	runner.Stderr = os.Stderr
	err = runner.Run()
	if err != nil {
		return err
	}
	return nil
}

func copyDir(src, dst string) error {
	sourceDir, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range sourceDir {
		sourcePath := path.Join(src, entry.Name())
		destPath := path.Join(dst, entry.Name())

		if entry.IsDir() {
			err = os.MkdirAll(destPath, os.ModePerm)
			if err != nil {
				return err
			}
			err = copyDir(sourcePath, destPath)
			if err != nil {
				return err
			}
		} else {
			err = copyFile(sourcePath, destPath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func buildImage(containerArgs *ContainerCmdArgs) (string, error) {
	imageName := fmt.Sprintf("%s-%s-%s-%s:dev", containerArgs.Lifecycle, containerArgs.Action, containerArgs.Pipeline, containerArgs.Container)

	pipelineDir := path.Join("workflows", containerArgs.Lifecycle, containerArgs.Action, containerArgs.Pipeline)

	fmt.Println("Building test image...")
	if err := forkBuilderCommand(buildContainerOpts, imageName, pipelineDir, containerArgs.Container); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// TODO: Remove hard-coded "platform" and move to a CLI arg (?) or config
	return imageName, kindLoadImage(imageName, "platform")
}

func kindLoadImage(image, clusterName string) error {
	cmd := exec.Command("kind", "load", "docker-image", image, "--name", clusterName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
