package cmd

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	containerutils "github.com/syntasso/kratix-cli/cmd/container_utils"
	"github.com/syntasso/kratix/api/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/yaml"
)

type ContainerCmdArgs struct {
	Lifecycle string
	Action    string
	Pipeline  string
}

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

	promise, err := LoadPromiseWithWorkflows(buildContainerOpts.Dir)
	if err != nil {
		return err
	}

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
		containerArgs, err := ParseContainerCmdArgs(container)
		if err != nil {
			return err
		}

		pipeline, err := RetrievePipeline(promise, containerArgs)
		if err != nil {
			return err
		}

		pipelineDir := filepath.Join(buildContainerOpts.Dir, "workflows", containerArgs.Lifecycle, containerArgs.Action, containerArgs.Pipeline)
		dirEntries, err := os.ReadDir(pipelineDir)
		if err != nil {
			return err
		}

		containerIndex, err := FindContainer(dirEntries, pipeline.Spec.Containers, buildContainerOpts.Name)
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
			if err := forkPushCommand(buildContainerOpts.Engine, containerImage); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		}
	}

	return nil
}

func LoadWorkflows(dir string) (v1alpha1.Workflows, error) {
	pipelineMap := map[string]map[string][]unstructured.Unstructured{}
	var workflows v1alpha1.Workflows

	missingWorkflows := 0
	for _, lifecycle := range []string{"promise", "resource"} {
		for _, action := range []string{"configure", "delete"} {
			if fileExists(filepath.Join(dir, "workflows", lifecycle, action, "workflow.yaml")) {
				workflowBytes, err := os.ReadFile(filepath.Join(dir, "workflows", lifecycle, action, "workflow.yaml"))
				if err != nil {
					return workflows, err
				}

				var workflow []v1alpha1.Pipeline
				err = yaml.Unmarshal(workflowBytes, &workflow)
				if err != nil {
					return workflows, fmt.Errorf("failed to get %s %s workflow: %s", lifecycle, action, err)
				}

				uPipelines, err := pipelinesToUnstructured(workflow)
				if err != nil {
					return workflows, err
				}

				if _, ok := pipelineMap[lifecycle]; !ok {
					pipelineMap[lifecycle] = make(map[string][]unstructured.Unstructured)
				}
				pipelineMap[lifecycle][action] = uPipelines
			} else {
				missingWorkflows++
			}
		}
	}

	if _, ok := pipelineMap["promise"]; ok {
		workflows.Promise.Configure = pipelineMap["promise"]["configure"]
		workflows.Promise.Delete = pipelineMap["promise"]["delete"]
	}
	if _, ok := pipelineMap["resource"]; ok {
		workflows.Resource.Configure = pipelineMap["resource"]["configure"]
		workflows.Resource.Delete = pipelineMap["resource"]["delete"]
	}

	return workflows, nil
}

func LoadPromiseWithWorkflows(dir string) (*v1alpha1.Promise, error) {
	var promise v1alpha1.Promise

	if _, err := os.Stat(filepath.Join(dir, "promise.yaml")); err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No promise.yaml found, assuming --split was used to initialise the Promise")
			workflows, err := LoadWorkflows(dir)
			if err != nil {
				return nil, err
			}
			promise.Spec.Workflows = workflows
			return &promise, nil
		}
		return nil, err
	}

	fileBytes, err := os.ReadFile(filepath.Join(dir, "promise.yaml"))
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(fileBytes, &promise)
	if err != nil {
		return nil, err
	}

	return &promise, nil
}

func RetrievePipeline(promise *v1alpha1.Promise, c *ContainerCmdArgs) (*v1alpha1.Pipeline, error) {
	allPipelines, err := v1alpha1.NewPipelinesMap(promise, ctrl.LoggerFrom(context.Background()))
	if err != nil {
		return nil, err
	}

	pipelines, pipelineIdx, err := findPipelinesForLifecycleAction(c, allPipelines)
	if err != nil {
		return nil, err
	}

	return &pipelines[pipelineIdx], nil
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

func ForkBuilderCommand(opts *BuildContainerOptions, containerImage, pipelineDir, containerName string) error {
	buildCommand := "build"

	buildArgs := []string{"--tag", containerImage, filepath.Join(pipelineDir, containerName)}
	if opts.Buildx {
		buildCommand = "buildx build"
		if opts.Push {
			buildArgs = append(buildArgs, "--push")
		}
	}
	buildArgs = append(buildArgs, strings.Fields(opts.BuildArgs)...)
	buildArgs = append(strings.Fields(buildCommand), buildArgs...)

	builder := exec.Command(opts.Engine, buildArgs...)
	builder.Stdout = os.Stdout
	builder.Stderr = os.Stderr
	if err := builder.Run(); err != nil {
		return err
	}
	return nil
}

func forkPushCommand(engine, containerImage string) error {
	builder := exec.Command(engine, "push", containerImage)
	builder.Stdout = os.Stdout
	builder.Stderr = os.Stderr
	if err := builder.Run(); err != nil {
		return err
	}
	return nil
}

func FindContainer(dirEntries []fs.DirEntry, containers []v1alpha1.Container, name string) (int, error) {
	if len(dirEntries) == 0 {
		return -1, fmt.Errorf("no container found in path")
	}

	if len(dirEntries) == 1 {
		if name != "" && name != dirEntries[0].Name() {
			return -1, fmt.Errorf("container %s not found in pipeline", name)
		}
		name = dirEntries[0].Name()
	}

	if name == "" {
		return -1, fmt.Errorf("more than one container exists for this pipeline, please provide a name with --name")
	}

	containerIndex := -1
	for i, container := range containers {
		if container.Name == name {
			containerIndex = i
			break
		}
	}
	if containerIndex == -1 {
		return -1, fmt.Errorf("container %s not found in pipeline", name)
	}

	for _, dirEntry := range dirEntries {
		if dirEntry.Name() == name {
			return containerIndex, nil
		}
	}

	return -1, fmt.Errorf("directory entry not found for container %s", name)
}
