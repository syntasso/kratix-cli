package stage

import (
	"fmt"
	"os"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

func ResolveInputFilePath() string {
	return GetEnvWithDefault("KRATIX_INPUT_FILE", DefaultInputFilePath)
}

func ResolveOutputFilePath() string {
	return GetEnvWithDefault("KRATIX_OUTPUT_FILE", DefaultOutputFilePath)
}

func ReadRequestFromFile(inputFile string) (*unstructured.Unstructured, error) {
	requestBytes, err := os.ReadFile(inputFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read object file from %s: %w", inputFile, err)
	}

	request := &unstructured.Unstructured{}
	if err := yaml.Unmarshal(requestBytes, request); err != nil {
		return nil, fmt.Errorf("failed to unmarshal object file: %w", err)
	}

	return request, nil
}

func RequireRequestName(request *unstructured.Unstructured) (string, error) {
	requestName := request.GetName()
	if requestName == "" {
		return "", fmt.Errorf("missing required field: metadata.name")
	}
	return requestName, nil
}

func RequestNamespaceWithDefault(request *unstructured.Unstructured) string {
	requestNamespace := request.GetNamespace()
	if requestNamespace == "" {
		return DefaultNamespace
	}
	return requestNamespace
}

func RequireSpecMap(request *unstructured.Unstructured) (map[string]any, error) {
	spec, ok := request.Object["spec"]
	if !ok {
		return nil, fmt.Errorf("missing required field: spec")
	}

	specMap, ok := spec.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid field: spec must be an object")
	}

	return specMap, nil
}

func WriteOutputObject(outputFile, objectKind string, output *unstructured.Unstructured) error {
	outputBytes, err := yaml.Marshal(output)
	if err != nil {
		return fmt.Errorf("failed to marshal %s object: %w", objectKind, err)
	}

	if err := os.WriteFile(outputFile, outputBytes, 0o644); err != nil {
		return fmt.Errorf("failed to write object file to %s: %w", outputFile, err)
	}

	return nil
}
