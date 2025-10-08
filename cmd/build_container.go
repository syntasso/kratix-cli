package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
	containerutils "github.com/syntasso/kratix-cli/cmd/container_utils"
	pipelineutils "github.com/syntasso/kratix-cli/cmd/pipeline_utils"
	promiseutils "github.com/syntasso/kratix-cli/cmd/promise_utils"
)

// containerCmd represents the container command
var buildContainerCmd = &cobra.Command{
	Use:   "container LIFECYCLE/ACTION/PIPELINE-NAME [flags]",
	Short: "Command to build a container image generated with 'add container'",
	Example: `  # Build a container
  kratix build container resource/configure/mypipeline --name mycontainer

  # Build all containers for all pipelines
  kratix build container --all

  # Build and push the image
  kratix build container resource/configure/mypipeline --name mycontainer --push

  # Custom build arguments
  kratix build container resource/configure/mypipeline --build-args "--platform linux/amd64"

  # Build with buildx
  kratix build container resource/configure/mypipeline --buildx --build-args "--builder custom-builder --platform=linux/arm64,linux/amd64"

  # Build with podman
  kratix build container resource/configure/mypipeline --engine podman
  `,
	RunE: BuildContainer,
}

var buildContainerOpts = &containerutils.BuildContainerOptions{}

func init() {
	buildCmd.AddCommand(buildContainerCmd)
	buildContainerCmd.Flags().StringVarP(&buildContainerOpts.Name, "name", "n", "", "Name of the container to build")
	buildContainerCmd.Flags().StringVarP(&buildContainerOpts.Dir, "dir", "d", ".", "Directory to read the Promise from")
	buildContainerCmd.Flags().BoolVarP(&buildContainerOpts.BuildAllContainers, "all", "a", false, "Build all of the containers for the Promise across all Workflows")
	buildContainerCmd.Flags().StringVarP(&buildContainerOpts.Engine, "engine", "e", "docker", "Build all of the containers for the Promise across all Workflows")
	buildContainerCmd.Flags().BoolVar(&buildContainerOpts.Buildx, "buildx", false, "Build the container using Buildx")
	buildContainerCmd.Flags().StringVar(&buildContainerOpts.BuildArgs, "build-args", "", "Extra build arguments to pass to the container build command")
	buildContainerCmd.Flags().BoolVar(&buildContainerOpts.Push, "push", false, "Build and push the container")
}

func BuildContainer(cmd *cobra.Command, args []string) error {
	if err := validateEngine(buildContainerOpts.Engine); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	promise, err := promiseutils.LoadPromiseWithWorkflows(buildContainerOpts.Dir)
	if err != nil {
		return fmt.Errorf("error LoadPromiseWithWorkflows: %s", err)
	}

	fmt.Println("")

	if len(promise.Spec.Workflows.Resource.Configure) == 0 && len(promise.Spec.Workflows.Resource.Delete) == 0 &&
		len(promise.Spec.Workflows.Promise.Configure) == 0 && len(promise.Spec.Workflows.Promise.Delete) == 0 {
		return fmt.Errorf("no workflows found")
	}

	if len(args) == 0 && !buildContainerOpts.BuildAllContainers {
		return fmt.Errorf("expected at least 1 argument")
	}

	var containersToBuild []string
	if buildContainerOpts.BuildAllContainers {
		for _, workflowType := range []string{"promise", "resource"} {
			for _, action := range []string{"configure", "delete"} {
				workflowDir := filepath.Join(buildContainerOpts.Dir, "workflows", workflowType, action)
				if _, err := os.Stat(workflowDir); os.IsNotExist(err) {
					continue
				}

				directories, err := os.ReadDir(workflowDir)
				if err != nil {
					return err
				}

				for _, dir := range directories {
					containersToBuild = append(containersToBuild, fmt.Sprintf("%s/%s/%s", workflowType, action, dir.Name()))
				}
			}
		}
	} else {
		containersToBuild = []string{args[0]}
	}

	for _, container := range containersToBuild {
		containerArgs, err := pipelineutils.ParsePipelineCmdArgs(container)
		if err != nil {
			return fmt.Errorf("error ParsePipelineCmdArgs: %s", err)
		}

		pipeline, err := pipelineutils.RetrievePipeline(promise, containerArgs)
		if err != nil {
			return fmt.Errorf("error RetrievePipeline: %s", err)
		}

		pipelineDir := filepath.Join(buildContainerOpts.Dir, "workflows", containerArgs.Lifecycle, containerArgs.Action, containerArgs.Pipeline)
		dirEntries, err := os.ReadDir(pipelineDir)
		if err != nil {
			return fmt.Errorf("error reading dir: %s", err)
		}

		containerIndex, err := pipelineutils.FindContainerIndex(dirEntries, pipeline.Spec.Containers, buildContainerOpts.Name)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		containerImage := pipeline.Spec.Containers[containerIndex].Image
		containerName := pipeline.Spec.Containers[containerIndex].Name

		fmt.Printf("Building container with tag %s...\n", containerImage)
		if err := containerutils.ForkBuilderCommand(buildContainerOpts, containerImage, pipelineDir, containerName); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)

		}

		if buildContainerOpts.Push && !buildContainerOpts.Buildx {
			fmt.Printf("Pushing container with tag %s...\n", containerImage)
			if err := containerutils.ForkPushCommand(buildContainerOpts.Engine, containerImage); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		}
	}

	return nil
}

func validateEngine(engine string) error {
	if engine != "docker" && engine != "podman" {
		return fmt.Errorf("unsupported container engine: %s", engine)
	}

	if _, err := exec.LookPath(engine); err != nil {
		return fmt.Errorf("%s CLI not found in PATH", engine)
	}
	return nil
}
