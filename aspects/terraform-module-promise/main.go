package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// GetEnv retrieves an environment variable or returns a default value if not set
func GetEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func main() {
	// Default paths
	yamlFile := GetEnv("KRATIX_INPUT_FILE", "/kratix/input/object.yaml")
	outputDir := GetEnv("KRATIX_OUTPUT_DIR", "/kratix/output")
	moduleSource := GetEnv("MODULE_SOURCE", "")

	// Ensure MODEL_SOURCE is set
	if moduleSource == "" {
		fmt.Println("Error: MODEL_SOURCE environment variable is not set")
		os.Exit(1)
	}

	// Read YAML file
	yamlContent, err := os.ReadFile(yamlFile)
	if err != nil {
		fmt.Printf("Error reading YAML file %s: %v\n", yamlFile, err)
		os.Exit(1)
	}

	// Parse YAML into a map
	var data map[string]any
	err = yaml.Unmarshal(yamlContent, &data)
	if err != nil {
		fmt.Printf("Error parsing YAML file: %v\n", err)
		os.Exit(1)
	}

	// Extract metadata fields
	metadata, ok := data["metadata"].(map[string]any)
	if !ok {
		fmt.Println("Error: metadata section not found in YAML file")
		os.Exit(1)
	}

	namespace, _ := metadata["namespace"].(string)
	name, _ := metadata["name"].(string)
	kind, _ := data["kind"].(string)

	if namespace == "" || name == "" || kind == "" {
		fmt.Println("Error: metadata.namespace, metadata.name, or kind is missing")
		os.Exit(1)
	}

	// Construct the output filename
	uniqueName := strings.ToLower(fmt.Sprintf("%s_%s_%s", kind, namespace, name))

	// Extract .spec section
	spec, ok := data["spec"].(map[string]any)
	if !ok {
		fmt.Println("Error: .spec section not found in YAML file")
		os.Exit(1)
	}

	// Create a valid Terraform JSON module block
	module := map[string]map[string]map[string]any{
		"module": {
			uniqueName: {
				"source": moduleSource,
			},
		},
	}

	// Add parameters from .spec
	for key, value := range spec {
		module["module"][uniqueName][key] = value
	}

	// Convert to JSON
	jsonData, err := json.MarshalIndent(module, "", "  ")
	if err != nil {
		fmt.Printf("Error generating JSON: %v\n", err)
		os.Exit(1)
	}

	// Ensure output directory exists
	err = os.MkdirAll(outputDir, os.ModePerm)
	if err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	// Write JSON to the dynamically generated filename
	path := filepath.Join(outputDir, uniqueName+".tf.json")
	err = os.WriteFile(path, jsonData, 0644)
	if err != nil {
		fmt.Printf("Error writing Terraform JSON file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Terraform JSON configuration written to %s\n", path)
}
