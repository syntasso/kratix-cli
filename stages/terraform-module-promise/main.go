package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/syntasso/kratix-cli/internal"
	"gopkg.in/yaml.v3"
)

func main() {
	outputDir := GetEnv("KRATIX_OUTPUT_DIR", "/kratix/output")
	workflowType := GetEnv("KRATIX_WORKFLOW_TYPE", "/kratix/input/object.yaml")

	if workflowType == "resource" {
		runResourceworkflow(outputDir)
	}
}

// parseOutputNames splits a comma-separated list of output names, trimming whitespace.
func parseOutputNames(env string) []string {
	if env == "" {
		return nil
	}
	var names []string
	for s := range strings.SplitSeq(env, ",") {
		if trimmed := strings.TrimSpace(s); trimmed != "" {
			names = append(names, trimmed)
		}
	}
	return names
}

// GetEnv retrieves an environment variable or returns a default value if not set
func GetEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func MustHaveEnv(key string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	panic(fmt.Sprintf("Error: %s environment variable is not set", key))
}

func runResourceworkflow(outputDir string) {
	inputObjectFilepath := GetEnv("KRATIX_INPUT_FILE", "/kratix/input/object.yaml")
	moduleSource := MustHaveEnv("MODULE_SOURCE")
	moduleRegistryVersion := os.Getenv("MODULE_REGISTRY_VERSION")

	if moduleRegistryVersion != "" && !internal.IsTerraformRegistrySource(moduleSource) {
		log.Fatalf("MODULE_REGISTRY_VERSION is only valid for Terraform registry sources (e.g., \"namespace/name/provider\"). For git or local sources, embed the version ref directly in MODULE_SOURCE (e.g., \"git::https://github.com/org/repo.git?ref=v1.2.3\"). Provided module_source=%q", moduleSource)
	}

	inputObject, err := os.ReadFile(inputObjectFilepath)
	if err != nil {
		log.Fatalf("Error reading YAML file %s: %v\n", inputObjectFilepath, err)
	}

	var inputObjectData map[string]any
	err = yaml.Unmarshal(inputObject, &inputObjectData)
	if err != nil {
		log.Fatalf("Error parsing YAML file: %v\n", err)
	}

	metadata, ok := inputObjectData["metadata"].(map[string]any)
	if !ok {
		log.Fatalf("Error: metadata section not found in YAML file")
	}

	namespace, _ := metadata["namespace"].(string)
	name, _ := metadata["name"].(string)
	kind, _ := inputObjectData["kind"].(string)

	if namespace == "" || name == "" || kind == "" {
		log.Fatalf("Error: metadata.namespace, metadata.name, or kind is missing")
	}

	uniqueFileName := strings.ToLower(fmt.Sprintf("%s_%s_%s", kind, namespace, name))

	config := map[string]any{
		"module": map[string]any{
			uniqueFileName: map[string]any{
				"source": moduleSource,
			},
		},
	}
	moduleBlock := config["module"].(map[string]any)
	moduleInstance := moduleBlock[uniqueFileName].(map[string]any)

	if moduleRegistryVersion != "" {
		moduleInstance["version"] = moduleRegistryVersion
	}

	// Handle spec if it exists
	if spec, ok := inputObjectData["spec"].(map[string]any); ok {
		for key, value := range spec {
			valSlice, ok := value.([]any)
			// 1. if its not an array and its not nil, add it to the module
			// 2. if its an array and its not empty, add it to the module
			// this gets around adding a bunch of empty arrays to the module
			if (!ok && value != nil) || (ok && len(valSlice) > 0) {
				moduleInstance[key] = value
			}
		}
	}

	if outputNames := parseOutputNames(os.Getenv("MODULE_OUTPUT_NAMES")); len(outputNames) > 0 {
		outputBlock := make(map[string]any)
		for _, name := range outputNames {
			uniqueOutputName := uniqueFileName + "_" + name
			outputBlock[uniqueOutputName] = map[string]any{
				"value": fmt.Sprintf("${module.%s.%s}", uniqueFileName, name),
			}
		}
		config["output"] = outputBlock
	}

	jsonData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		log.Fatalf("Error generating JSON: %v\n", err)
	}

	err = os.MkdirAll(outputDir, os.ModePerm)
	if err != nil {
		log.Fatalf("Error creating output directory: %v\n", err)
	}

	path := filepath.Join(outputDir, uniqueFileName+".tf.json")
	err = os.WriteFile(path, jsonData, 0644)
	if err != nil {
		log.Fatalf("Error writing Terraform JSON file: %v\n", err)
	}

	fmt.Printf("Terraform JSON configuration written to %s\n", path)
}
