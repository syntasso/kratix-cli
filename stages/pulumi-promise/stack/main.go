package main

import (
	"fmt"
	"log"
	"os"

	stage "github.com/syntasso/kratix-cli/stages/pulumi-promise/internal/stage"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

const (
	stackAPIVersion = "pulumi.com/v1"
	stackKind       = "Stack"

	pulumiComponentTokenEnvVar = "PULUMI_COMPONENT_TOKEN"
)

func main() {
	componentToken := getRequiredEnv(pulumiComponentTokenEnvVar)

	if err := transformInputToStackOutput(componentToken); err != nil {
		log.Fatalf("%v", err)
	}
}

func transformInputToStackOutput(componentToken string) error {
	inputFile := stage.GetEnvWithDefault("KRATIX_INPUT_FILE", stage.DefaultInputFilePath)
	outputFile := stage.GetEnvWithDefault("KRATIX_OUTPUT_FILE", stage.DefaultOutputFilePath)

	requestBytes, err := os.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("failed to read object file from %s: %w", inputFile, err)
	}

	request := &unstructured.Unstructured{}
	if err := yaml.Unmarshal(requestBytes, request); err != nil {
		return fmt.Errorf("failed to unmarshal object file: %w", err)
	}

	requestName := request.GetName()
	if requestName == "" {
		return fmt.Errorf("missing required field: metadata.name")
	}

	requestNamespace := request.GetNamespace()
	if requestNamespace == "" {
		requestNamespace = stage.DefaultNamespace
	}

	programName := stage.BuildProgramName(requestName, requestNamespace, request.GetKind(), componentToken)
	stackResourceName := buildStackResourceName(programName)
	stackName := stackResourceName

	output := &unstructured.Unstructured{}
	output.SetAPIVersion(stackAPIVersion)
	output.SetKind(stackKind)
	output.SetName(stackResourceName)
	output.SetNamespace(requestNamespace)
	output.SetLabels(request.GetLabels())
	output.SetAnnotations(request.GetAnnotations())

	if err := unstructured.SetNestedField(output.Object, map[string]any{
		"name": programName,
	}, "spec", "programRef"); err != nil {
		return fmt.Errorf("failed to set spec.programRef: %w", err)
	}

	if err := unstructured.SetNestedField(output.Object, stackName, "spec", "stack"); err != nil {
		return fmt.Errorf("failed to set spec.stack: %w", err)
	}

	outputBytes, err := yaml.Marshal(output)
	if err != nil {
		return fmt.Errorf("failed to marshal Stack object: %w", err)
	}

	if err := os.WriteFile(outputFile, outputBytes, 0o644); err != nil {
		return fmt.Errorf("failed to write object file to %s: %w", outputFile, err)
	}

	return nil
}

func buildStackResourceName(programName string) string {
	name := fmt.Sprintf("%s-stack", programName)
	if len(name) <= 63 {
		return name
	}
	return name[:63]
}

func getRequiredEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("missing required environment variable %s", key)
	}
	return value
}
