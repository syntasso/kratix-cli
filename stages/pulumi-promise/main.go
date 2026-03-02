package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/syntasso/kratix-cli/internal/pulumi"
	stage "github.com/syntasso/kratix-cli/stages/pulumi-promise/internal/stage"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

const (
	programAPIVersion = "pulumi.com/v1"
	programKind       = "Program"

	pulumiComponentTokenEnvVar = "PULUMI_COMPONENT_TOKEN"
	pulumiSchemaSourceEnvVar   = "PULUMI_SCHEMA_SOURCE"
)

func main() {
	componentToken := getEnvOrDie(pulumiComponentTokenEnvVar)
	schemaSource := getEnvOrDie(pulumiSchemaSourceEnvVar)

	if err := transformInputToProgramOutput(componentToken, schemaSource); err != nil {
		log.Fatalf("%v", err)
	}
}

func transformInputToProgramOutput(componentToken, schemaSource string) error {
	inputFile := getEnvWithDefault("KRATIX_INPUT_FILE", stage.DefaultInputFilePath)
	outputFile := getEnvWithDefault("KRATIX_OUTPUT_FILE", stage.DefaultOutputFilePath)

	requestBytes, err := os.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("failed to read object file from %s: %w", inputFile, err)
	}

	request := &unstructured.Unstructured{}
	if err := yaml.Unmarshal(requestBytes, request); err != nil {
		return fmt.Errorf("failed to unmarshal object file: %w", err)
	}

	spec, ok := request.Object["spec"]
	if !ok {
		return fmt.Errorf("missing required field: spec")
	}

	specMap, ok := spec.(map[string]any)
	if !ok {
		return fmt.Errorf("invalid field: spec must be an object")
	}

	requestName := request.GetName()
	if requestName == "" {
		return fmt.Errorf("missing required field: metadata.name")
	}

	requestNamespace := request.GetNamespace()
	if requestNamespace == "" {
		requestNamespace = stage.DefaultNamespace
	}

	resourceName := stage.BuildProgramResourceName(componentToken)
	programName := stage.BuildProgramName(requestName, requestNamespace, request.GetKind(), componentToken)
	programConfiguration, err := buildProgramConfiguration(schemaSource)
	if err != nil {
		return err
	}

	output := &unstructured.Unstructured{}
	output.SetAPIVersion(programAPIVersion)
	output.SetKind(programKind)
	output.SetName(programName)
	output.SetNamespace(requestNamespace)
	output.SetLabels(request.GetLabels())
	output.SetAnnotations(request.GetAnnotations())

	if err := unstructured.SetNestedField(output.Object, map[string]any{
		resourceName: map[string]any{
			"type":       componentToken,
			"properties": specMap,
		},
	}, "program", "resources"); err != nil {
		return fmt.Errorf("failed to set program.resources: %w", err)
	}

	if len(programConfiguration) > 0 {
		if err := unstructured.SetNestedMap(output.Object, programConfiguration, "program", "configuration"); err != nil {
			return fmt.Errorf("failed to set program.configuration: %w", err)
		}
	}

	outputBytes, err := yaml.Marshal(output)
	if err != nil {
		return fmt.Errorf("failed to marshal Program object: %w", err)
	}

	if err := os.WriteFile(outputFile, outputBytes, 0o644); err != nil {
		return fmt.Errorf("failed to write object file to %s: %w", outputFile, err)
	}

	return nil
}

type schemaConfigVariable struct {
	Type    string `json:"type"`
	Default any    `json:"default"`
	Secret  *bool  `json:"secret"`
}

func buildProgramConfiguration(schemaSource string) (map[string]any, error) {
	schemaDoc, err := pulumi.LoadSchema(schemaSource)
	if err != nil {
		return nil, fmt.Errorf("load schema for Program configuration: %w", err)
	}

	if len(schemaDoc.Config.Variables) == 0 {
		return nil, nil
	}

	configuration := make(map[string]any, len(schemaDoc.Config.Variables))
	for _, key := range stage.SortedRawKeys(schemaDoc.Config.Variables) {
		variable, err := parseConfigVariable(schemaDoc.Config.Variables[key])
		if err != nil {
			return nil, fmt.Errorf("load schema for Program configuration: parse config variable %q: %w", key, err)
		}

		entry := map[string]any{}
		if variable.Type != "" {
			entry["type"] = variable.Type
		}
		if variable.Default != nil {
			entry["default"] = variable.Default
		}
		if variable.Secret != nil {
			entry["secret"] = *variable.Secret
		}

		if len(entry) == 0 {
			continue
		}
		configuration[key] = entry
	}

	if len(configuration) == 0 {
		return nil, nil
	}
	return configuration, nil
}

func parseConfigVariable(raw json.RawMessage) (schemaConfigVariable, error) {
	var variable schemaConfigVariable
	if err := json.Unmarshal(raw, &variable); err != nil {
		return schemaConfigVariable{}, err
	}
	return variable, nil
}

func getEnvWithDefault(key, defaultValue string) string {
	return stage.GetEnvWithDefault(key, defaultValue)
}

func getEnvOrDie(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("expected %s to be set", key)
	}
	return value
}
