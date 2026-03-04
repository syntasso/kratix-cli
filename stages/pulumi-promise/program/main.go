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
	log.Printf("starting transformation (componentToken=%q, schemaSource=%q)", componentToken, schemaSource)

	if err := transformInputToProgramOutput(componentToken, schemaSource); err != nil {
		log.Printf("failed: %v", err)
		log.Fatalf("%v", err)
	}
}

func transformInputToProgramOutput(componentToken, schemaSource string) error {
	inputFile := stage.ResolveInputFilePath()
	outputFile := stage.ResolveProgramOutputFilePath()
	log.Printf("using input file %q and output file %q", inputFile, outputFile)

	request, err := stage.ReadRequestFromFile(inputFile)
	if err != nil {
		log.Printf("unable to read request from %q: %v", inputFile, err)
		return err
	}
	log.Printf("loaded request (name=%q, kind=%q, namespace=%q)", request.GetName(), request.GetKind(), request.GetNamespace())

	specMap, err := stage.RequireSpecMap(request)
	if err != nil {
		log.Printf("request validation failed: %v", err)
		return err
	}
	log.Printf("request spec is present with %d top-level field(s)", len(specMap))

	requestName, err := stage.RequireRequestName(request)
	if err != nil {
		log.Printf("request validation failed: %v", err)
		return err
	}

	requestNamespace := stage.RequestNamespaceWithDefault(request)

	resourceName := stage.BuildProgramResourceName(componentToken)
	programName := stage.BuildProgramName(requestName, requestNamespace, request.GetKind(), componentToken)
	log.Printf("computed names (programName=%q, resourceName=%q)", programName, resourceName)
	programConfiguration, err := buildProgramConfiguration(schemaSource)
	if err != nil {
		log.Printf("failed to build configuration from schema %q: %v", schemaSource, err)
		return err
	}
	log.Printf("built configuration entries=%d", len(programConfiguration))

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

	if err := stage.WriteOutputObject(outputFile, programKind, output); err != nil {
		log.Printf("failed to write Program to %q: %v", outputFile, err)
		return err
	}

	log.Printf("wrote Program %q to %q", programName, outputFile)
	return nil
}

type schemaConfigVariable struct {
	Type    string `json:"type"`
	Default any    `json:"default"`
	Secret  *bool  `json:"secret"`
}

func buildProgramConfiguration(schemaSource string) (map[string]any, error) {
	log.Printf("loading schema from %q", schemaSource)
	schemaDoc, err := pulumi.LoadSchema(schemaSource)
	if err != nil {
		return nil, fmt.Errorf("load schema for Program configuration: %w", err)
	}

	if len(schemaDoc.Config.Variables) == 0 {
		log.Printf("schema has no config variables")
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
		log.Printf("no trusted configuration values found in schema")
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
