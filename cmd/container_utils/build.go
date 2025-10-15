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
	return builder.Run()
}

func ForkRunCommand(opts *BuildContainerOptions, containerImage, inputVolume, outputVolume, metadataVolume string, envvars []string, command string) error {
	args := []string{
		"run",
		"--volume", fmt.Sprintf("%s:/kratix/input/", inputVolume),
		"--volume", fmt.Sprintf("%s:/kratix/output/", outputVolume),
		"--volume", fmt.Sprintf("%s:/kratix/metadata/", metadataVolume),
	}

	if len(envvars) > 0 {
		for _, evar := range envvars {
			args = append(args, "--env")
			args = append(args, evar)
		}
	}

	args = append(args, strings.Fields(opts.BuildArgs)...)
	args = append(args, containerImage)

	if command != "" {
		args = append(args, "sh", "-c", command)
	}

	cmd := exec.Command(opts.Engine, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func ForkPushCommand(engine, containerImage string) error {
	builder := exec.Command(engine, "push", containerImage)
	builder.Stdout = os.Stdout
	builder.Stderr = os.Stderr
	return builder.Run()
}
