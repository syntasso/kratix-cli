package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/syntasso/kratix-cli/internal/pulumi"
	stage "github.com/syntasso/kratix-cli/stages/pulumi-promise/internal/stage"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
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
	inputFile := stage.ResolveInputFilePath()
	outputFile := stage.ResolveOutputFilePath()

	request, err := stage.ReadRequestFromFile(inputFile)
	if err != nil {
		return err
	}

	specMap, err := stage.RequireSpecMap(request)
	if err != nil {
		return err
	}

	requestName, err := stage.RequireRequestName(request)
	if err != nil {
		return err
	}

	requestNamespace := stage.RequestNamespaceWithDefault(request)

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

	return stage.WriteOutputObject(outputFile, programKind, output)
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

func getEnvOrDie(key string) string {
	value := stage.GetEnvWithDefault(key, "")
	if value == "" {
		log.Fatalf("expected %s to be set", key)
	}
	return value
}
