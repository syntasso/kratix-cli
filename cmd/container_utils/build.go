package containerutils

import (
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
