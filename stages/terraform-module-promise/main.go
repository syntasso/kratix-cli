package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/syntasso/kratix-cli/internal"
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

// generateHCL produces a native HCL module block (and optional output blocks) from a
// map[string]any spec. Keys are sorted alphabetically for deterministic output.
func generateHCL(moduleName, source, version string, spec map[string]any, outputNames []string) []byte {
	var sb strings.Builder

	fmt.Fprintf(&sb, "module %q {\n", moduleName)

	// Align "source" and "version" when both are present (terraform fmt convention).
	if version != "" {
		fmt.Fprintf(&sb, "  source  = %q\n", source)
		fmt.Fprintf(&sb, "  version = %q\n", version)
	} else {
		fmt.Fprintf(&sb, "  source = %q\n", source)
	}

	if len(spec) > 0 {
		sb.WriteString("\n")
		for _, k := range sortedKeys(spec) {
			sb.WriteString("  ")
			sb.WriteString(k)
			sb.WriteString(" = ")
			writeHCLValue(&sb, spec[k], 1)
			sb.WriteString("\n")
		}
	}

	sb.WriteString("}")

	for _, name := range outputNames {
		fmt.Fprintf(&sb, "\n\noutput %q {\n", moduleName+"_"+name)
		fmt.Fprintf(&sb, "  value = \"${module.%s.%s}\"\n", moduleName, name)
		sb.WriteString("}")
	}

	return []byte(sb.String())
}

// writeHCLValue writes a single HCL attribute value at the given indent depth.
// Primitives are written inline; maps and object-bearing lists use multi-line format.
func writeHCLValue(sb *strings.Builder, v any, depth int) {
	indent := strings.Repeat("  ", depth)
	innerIndent := strings.Repeat("  ", depth+1)

	switch val := v.(type) {
	case string:
		sb.WriteString(strconv.Quote(val))
	case int:
		fmt.Fprintf(sb, "%d", val)
	case float64:
		if val == float64(int64(val)) {
			fmt.Fprintf(sb, "%d", int64(val))
		} else {
			fmt.Fprintf(sb, "%g", val)
		}
	case bool:
		if val {
			sb.WriteString("true")
		} else {
			sb.WriteString("false")
		}
	case []any:
		if len(val) == 0 {
			sb.WriteString("[]")
			return
		}
		hasObjects := false
		for _, elem := range val {
			if _, ok := elem.(map[string]any); ok {
				hasObjects = true
				break
			}
		}
		if !hasObjects {
			sb.WriteString("[")
			for i, elem := range val {
				if i > 0 {
					sb.WriteString(", ")
				}
				writeHCLValue(sb, elem, depth)
			}
			sb.WriteString("]")
		} else {
			sb.WriteString("[\n")
			for _, elem := range val {
				sb.WriteString(innerIndent)
				writeHCLValue(sb, elem, depth+1)
				sb.WriteString(",\n")
			}
			sb.WriteString(indent)
			sb.WriteString("]")
		}
	case map[string]any:
		sb.WriteString("{\n")
		for _, k := range sortedKeys(val) {
			sb.WriteString(innerIndent)
			sb.WriteString(k)
			sb.WriteString(" = ")
			writeHCLValue(sb, val[k], depth+1)
			sb.WriteString("\n")
		}
		sb.WriteString(indent)
		sb.WriteString("}")
	}
}

func sortedKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
