package lib

import (
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TransformInputToOutput(group, version, kind string) error {
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
		return fmt.Errorf("Failed to read object file from %s: %w", inputFile, err)
	}

	uRequestObj := &unstructured.Unstructured{}
	err = yaml.Unmarshal(requestContents, uRequestObj)
	if err != nil {
		return fmt.Errorf("Failed to unmarshal object file: %w", err)
	}

	outputObject := &unstructured.Unstructured{}
	outputObject.SetName(uRequestObj.GetName())
	outputObject.SetNamespace("default")
	outputObject.SetKind(kind)
	outputObject.SetAPIVersion(group + "/" + version)
	outputObject.SetLabels(uRequestObj.GetLabels())
	outputObject.SetAnnotations(uRequestObj.GetAnnotations())

	unstructured.SetNestedField(outputObject.Object, uRequestObj.Object["spec"], "spec")

	outputObjectBytes, _ := yaml.Marshal(outputObject)
	if err := os.WriteFile(outputFile, outputObjectBytes, 0644); err != nil {
		return fmt.Errorf("Failed to write object file to %s: %w", outputFile, err)
	}

	return nil
}

func GetEnvOrDie(envVar string) string {
	value := os.Getenv(envVar)
	if value == "" {
		log.Fatalf("Expected %s to be set", envVar)
	}

	return value
}
