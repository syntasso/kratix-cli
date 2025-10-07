package containerutils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type BuildContainerOptions struct {
	Name               string
	Dir                string
	BuildAllContainers bool

	Engine    string
	Buildx    bool
	Push      bool
	BuildArgs string
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

func ForkRunCommand(opts *BuildContainerOptions, containerImage, inputVolume, outputVolume, metadataVolume, command string) error {
	runCommand := "run"

	buildArgs := []string{"--volume", fmt.Sprintf("%s:%s", inputVolume, "/kratix/input/"), "--volume", fmt.Sprintf("%s:%s", outputVolume, "/kratix/output/"), "--volume", fmt.Sprintf("%s:%s", metadataVolume, "/kratix/metadata/"), containerImage}
	commandArgs := []string{"-c", command}

	if command != "" {
		buildArgs = append(buildArgs, commandArgs...)
	}

	buildArgs = append(buildArgs, strings.Fields(opts.BuildArgs)...)
	buildArgs = append(strings.Fields(runCommand), buildArgs...)

	builder := exec.Command(opts.Engine, buildArgs...)
	builder.Stdout = os.Stdout
	builder.Stderr = os.Stderr
	if err := builder.Run(); err != nil {
		return err
	}
	return nil
}
