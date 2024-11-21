package main

import (
	"log"
	"os"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

func main() {
	inputFile := os.Getenv("KRATIX_INPUT_FILE")
	if inputFile == "" {
		inputFile = "/kratix/input/object.yaml"
	}

	outputFile := os.Getenv("KRATIX_OUTPUT_FILE")
	if outputFile == "" {
		outputFile = "/kratix/output/object.yaml"
	}

	requestContents, err := os.ReadFile(inputFile)
	if err != nil {
		log.Fatalf("Failed to read object file from %s: %v", inputFile, err)
	}

	uRequestObj := &unstructured.Unstructured{}
	err = yaml.Unmarshal(requestContents, uRequestObj)
	if err != nil {
		log.Fatalf("Failed to unmarshal object file: %v", err)
	}

	operatorGroup := getEnvOrDie("OPERATOR_GROUP")
	operatorVersion := getEnvOrDie("OPERATOR_VERSION")
	operatorKind := getEnvOrDie("OPERATOR_KIND")

	outputObject := &unstructured.Unstructured{}
	outputObject.SetName(uRequestObj.GetName())
	outputObject.SetNamespace("default")
	outputObject.SetKind(operatorKind)
	outputObject.SetAPIVersion(operatorGroup + "/" + operatorVersion)
	outputObject.SetLabels(uRequestObj.GetLabels())
	outputObject.SetAnnotations(uRequestObj.GetAnnotations())

	unstructured.SetNestedField(outputObject.Object, uRequestObj.Object["spec"], "spec")

	outputObjectBytes, _ := yaml.Marshal(outputObject)
	if err := os.WriteFile(outputFile, outputObjectBytes, 0644); err != nil {
		log.Fatalf("Failed to write object file to %s: %v", outputFile, err)
	}
}

func getEnvOrDie(envVar string) string {
	value := os.Getenv(envVar)
	if value == "" {
		log.Fatalf("Expected %s to be set", envVar)
	}

	return value
}
