package main

import (
	"log"

	"github.com/syntasso/kratix-cli-plugin-investigation/stages/helm-promise/lib"
)

func main() {
	operatorGroup := lib.GetEnvOrDie("OPERATOR_GROUP")
	operatorVersion := lib.GetEnvOrDie("OPERATOR_VERSION")
	operatorKind := lib.GetEnvOrDie("OPERATOR_KIND")

	err := lib.TransformInputToOutput(operatorGroup, operatorVersion, operatorKind)
	if err != nil {
		log.Fatalf("%v", err)
	}
}
