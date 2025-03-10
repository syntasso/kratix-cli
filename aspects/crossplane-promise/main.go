package main

import (
	"log"

	"github.com/syntasso/kratix-cli/aspects/helm-promise/lib"
)

func main() {
	group := lib.GetEnvOrDie("XRD_GROUP")
	version := lib.GetEnvOrDie("XRD_VERSION")
	kind := lib.GetEnvOrDie("XRD_KIND")

	err := lib.TransformInputToOutput(group, version, kind)
	if err != nil {
		log.Fatalf("%v", err)
	}
}
