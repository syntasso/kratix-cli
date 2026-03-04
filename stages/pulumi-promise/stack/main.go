package main

import (
	"fmt"
	"log"

	stage "github.com/syntasso/kratix-cli/stages/pulumi-promise/internal/stage"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
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
	inputFile := stage.ResolveInputFilePath()
	outputFile := stage.ResolveStackOutputFilePath()

	request, err := stage.ReadRequestFromFile(inputFile)
	if err != nil {
		return err
	}

	requestName, err := stage.RequireRequestName(request)
	if err != nil {
		return err
	}

	requestNamespace := stage.RequestNamespaceWithDefault(request)

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

	return stage.WriteOutputObject(outputFile, stackKind, output)
}

func buildStackResourceName(programName string) string {
	name := fmt.Sprintf("%s-stack", programName)
	if len(name) <= 63 {
		return name
	}
	return name[:63]
}

func getRequiredEnv(key string) string {
	value := stage.GetEnvWithDefault(key, "")
	if value == "" {
		log.Fatalf("missing required environment variable %s", key)
	}
	return value
}
