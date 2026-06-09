package main

import (
	"fmt"
	"log"
	"maps"
	"os"
	"path/filepath"
	"slices"
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

	spec := make(map[string]any)
	if specMap, ok := inputObjectData["spec"].(map[string]any); ok {
		for key, value := range specMap {
			valSlice, ok := value.([]any)
			// 1. if its not an array and its not nil, add it to the module
			// 2. if its an array and its not empty, add it to the module
			// this gets around adding a bunch of empty arrays to the module
			if (!ok && value != nil) || (ok && len(valSlice) > 0) {
				spec[key] = value
			}
		}
	}

	outputNames := parseOutputNames(os.Getenv("MODULE_OUTPUT_NAMES"))

	hclData := generateHCL(uniqueFileName, moduleSource, moduleRegistryVersion, spec, outputNames)

	err = os.MkdirAll(outputDir, os.ModePerm)
	if err != nil {
		log.Fatalf("Error creating output directory: %v\n", err)
	}

	path := filepath.Join(outputDir, uniqueFileName+".tf")
	err = os.WriteFile(path, hclData, 0644)
	if err != nil {
		log.Fatalf("Error writing Terraform file: %v\n", err)
	}

	fmt.Printf("Terraform configuration written to %s\n", path)
}

func generateHCL(moduleName, source, version string, spec map[string]any, outputNames []string) []byte {
	f := hclwrite.NewEmptyFile()
	rootBody := f.Body()

	moduleBlock := rootBody.AppendNewBlock("module", []string{moduleName})
	body := moduleBlock.Body()

	body.SetAttributeValue("source", cty.StringVal(source))
	if version != "" {
		body.SetAttributeValue("version", cty.StringVal(version))
	}

	if len(spec) > 0 {
		body.AppendNewline()
		for _, key := range slices.Sorted(maps.Keys(spec)) {
			body.SetAttributeValue(key, goToCty(spec[key]))
		}
	}

	for _, name := range outputNames {
		rootBody.AppendNewline()
		outputBlock := rootBody.AppendNewBlock("output", []string{moduleName + "_" + name})
		traversal := hcl.Traversal{
			hcl.TraverseRoot{Name: "module"},
			hcl.TraverseAttr{Name: moduleName},
			hcl.TraverseAttr{Name: name},
		}
		outputBlock.Body().SetAttributeRaw("value", hclwrite.TokensForTraversal(traversal))
	}

	return f.Bytes()
}

func goToCty(v any) cty.Value {
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
		vals := make([]cty.Value, len(val))
		for i, elem := range val {
			vals[i] = goToCty(elem)
		}
		return cty.TupleVal(vals)
	case map[string]any:
		if len(val) == 0 {
			return cty.EmptyObjectVal
		}
		attrs := make(map[string]cty.Value, len(val))
		for k, v := range val {
			attrs[k] = goToCty(v)
		}
		return cty.ObjectVal(attrs)
	default:
		log.Fatalf("unsupported value type %T", v)
		return cty.NilVal
	}
}
