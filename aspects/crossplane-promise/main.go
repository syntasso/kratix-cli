package main

import "github.com/syntasso/kratix-cli/aspects/helm-promise/lib"

func main() {
	group := lib.GetEnvOrDie("GROUP")
	version := lib.GetEnvOrDie("VERSION")
	kind := lib.GetEnvOrDie("KIND")

	err := lib.TransformInputToOutput(group, version, kind)
	if err != nil {
		panic(err)
	}
}
