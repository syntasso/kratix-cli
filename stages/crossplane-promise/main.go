package main

import (
	"log"

	"github.com/syntasso/kratix-cli-plugin-investigation/cmd"
	"github.com/syntasso/kratix-cli/stages/helm-promise/lib"
)

func main() {
	group := lib.GetEnvOrDie(cmd.XRD_GROUP_ENV_VAR)
	version := lib.GetEnvOrDie(cmd.XRD_VERSION_ENV_VAR)
	kind := lib.GetEnvOrDie(cmd.XRD_KIND_ENV_VAR)

	err := lib.TransformInputToOutput(group, version, kind)
	if err != nil {
		log.Fatalf("%v", err)
	}
}
