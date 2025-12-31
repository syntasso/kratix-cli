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
	yamlFile := GetEnv("KRATIX_INPUT_FILE", "/kratix/input/object.yaml")
	outputDir := GetEnv("KRATIX_OUTPUT_DIR", "/kratix/output")
	moduleSource := MustHaveEnv("MODULE_SOURCE")
	moduleRegistryVersion := os.Getenv("MODULE_REGISTRY_VERSION")

	if moduleRegistryVersion != "" && !internal.IsTerraformRegistrySource(moduleSource) {
		log.Fatalf("MODULE_REGISTRY_VERSION is only valid for Terraform registry sources (e.g., \"namespace/name/provider\"). For git or local sources, embed the ref directly in MODULE_SOURCE (e.g., \"git::https://github.com/org/repo.git?ref=v1.2.3\"). Provided module_source=%q", moduleSource)
	}

	yamlContent, err := os.ReadFile(yamlFile)
	if err != nil {
		log.Fatalf("Error reading YAML file %s: %v\n", yamlFile, err)
	}

	var data map[string]any
	err = yaml.Unmarshal(yamlContent, &data)
	if err != nil {
		log.Fatalf("Error parsing YAML file: %v\n", err)
	}

	metadata, ok := data["metadata"].(map[string]any)
	if !ok {
		log.Fatalf("Error: metadata section not found in YAML file")
	}

	namespace, _ := metadata["namespace"].(string)
	name, _ := metadata["name"].(string)
	kind, _ := data["kind"].(string)

	if namespace == "" || name == "" || kind == "" {
		log.Fatalf("Error: metadata.namespace, metadata.name, or kind is missing")
	}

	uniqueFileName := strings.ToLower(fmt.Sprintf("%s_%s_%s", kind, namespace, name))

	module := map[string]map[string]map[string]any{
		"module": {
			uniqueFileName: {
				"source": moduleSource,
			},
		},
	}

	if moduleRegistryVersion != "" {
		module["module"][uniqueFileName]["version"] = moduleRegistryVersion
	}

	// Handle spec if it exists
	if spec, ok := data["spec"].(map[string]any); ok {
		for key, value := range spec {
			valSlice, ok := value.([]any)
			// 1. if its not an array and its not nil, add it to the module
			// 2. if its an array and its not empty, add it to the module
			// this gets around adding a bunch of empty arrays to the module
			if (!ok && value != nil) || (ok && len(valSlice) > 0) {
				module["module"][uniqueFileName][key] = value
			}
		}
	}

	jsonData, err := json.MarshalIndent(module, "", "  ")
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
