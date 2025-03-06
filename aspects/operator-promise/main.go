package main

import "github.com/syntasso/kratix-cli/aspects/helm-promise/lib"

func main() {
	operatorGroup := lib.GetEnvOrDie("OPERATOR_GROUP")
	operatorVersion := lib.GetEnvOrDie("OPERATOR_VERSION")
	operatorKind := lib.GetEnvOrDie("OPERATOR_KIND")

	err := lib.TransformInputToOutput(operatorGroup, operatorVersion, operatorKind)
	if err != nil {
		panic(err)
	}
}
