package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/syntasso/kratix-cli/internal"
	"github.com/zclconf/go-cty/cty"
	"gopkg.in/yaml.v3"
)

func main() {
	outputDir := GetEnv("KRATIX_OUTPUT_DIR", "/kratix/output")
	workflowType := GetEnv("KRATIX_WORKFLOW_TYPE", "/kratix/input/object.yaml")

	if workflowType == "resource" {
		runResourceWorkflow(outputDir)
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

// goNativeToCtyValue converts Go native types (as produced by yaml.Unmarshal) to cty.Value
// for use with hclwrite. Handles strings, ints, float64s, bools, slices, and maps.
func goNativeToCtyValue(v any) cty.Value {
	switch val := v.(type) {
	case string:
		return cty.StringVal(val)
	case int:
		return cty.NumberIntVal(int64(val))
	case float64:
		return cty.NumberFloatVal(val)
	case bool:
		return cty.BoolVal(val)
	case []any:
		if len(val) == 0 {
			return cty.EmptyObjectVal
		}
		elements := make([]cty.Value, len(val))
		for i, item := range val {
			elements[i] = goNativeToCtyValue(item)
		}
		return cty.TupleVal(elements)
	case map[string]any:
		if len(val) == 0 {
			return cty.EmptyObjectVal
		}
		attrs := make(map[string]cty.Value, len(val))
		for k, v := range val {
			attrs[k] = goNativeToCtyValue(v)
		}
		return cty.ObjectVal(attrs)
	default:
		log.Fatalf("unsupported type for HCL conversion: %T (value: %v)", v, v)
		return cty.NilVal
	}
}

func runResourceWorkflow(outputDir string) {
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

	hclFile := hclwrite.NewEmptyFile()
	rootBody := hclFile.Body()

	moduleBlock := rootBody.AppendNewBlock("module", []string{uniqueFileName})
	moduleBody := moduleBlock.Body()

	moduleBody.SetAttributeValue("source", cty.StringVal(moduleSource))

	if moduleRegistryVersion != "" {
		moduleBody.SetAttributeValue("version", cty.StringVal(moduleRegistryVersion))
	}

	if spec, ok := inputObjectData["spec"].(map[string]any); ok {
		// Sort keys for deterministic output
		keys := make([]string, 0, len(spec))
		for key := range spec {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		for _, key := range keys {
			value := spec[key]
			valSlice, isSlice := value.([]any)
			// 1. if its not an array and its not nil, add it to the module
			// 2. if its an array and its not empty, add it to the module
			// this gets around adding a bunch of empty arrays to the module
			if (!isSlice && value != nil) || (isSlice && len(valSlice) > 0) {
				moduleBody.SetAttributeValue(key, goNativeToCtyValue(value))
			}
		}
	}

	if outputNames := parseOutputNames(os.Getenv("MODULE_OUTPUT_NAMES")); len(outputNames) > 0 {
		for _, outputName := range outputNames {
			uniqueOutputName := uniqueFileName + "_" + outputName
			outputBlock := rootBody.AppendNewBlock("output", []string{uniqueOutputName})
			// Use a traversal expression (module.name.output) rather than a string
			// interpolation ("${module.name.output}") — cleaner HCL and avoids
			// hclwrite escaping ${ to $${.
			traversal := hcl.Traversal{
				hcl.TraverseRoot{Name: "module"},
				hcl.TraverseAttr{Name: uniqueFileName},
				hcl.TraverseAttr{Name: outputName},
			}
			outputBlock.Body().SetAttributeRaw("value", hclwrite.TokensForTraversal(traversal))
		}
	}

	err = os.MkdirAll(outputDir, os.ModePerm)
	if err != nil {
		log.Fatalf("Error creating output directory: %v\n", err)
	}

	path := filepath.Join(outputDir, uniqueFileName+".tf")
	err = os.WriteFile(path, hclFile.Bytes(), 0644)
	if err != nil {
		log.Fatalf("Error writing Terraform HCL file: %v\n", err)
	}

	fmt.Printf("Terraform HCL configuration written to %s\n", path)
}
