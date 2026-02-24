package main

import (
	"fmt"
	"hash/fnv"
	"log"
	"os"
	"regexp"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

const (
	defaultInputFilePath  = "/kratix/input/object.yaml"
	defaultOutputFilePath = "/kratix/output/object.yaml"
	defaultNamespace      = "default"

	programAPIVersion = "pulumi.com/v1"
	programKind       = "Program"

	pulumiComponentTokenEnvVar = "PULUMI_COMPONENT_TOKEN"
)

var invalidNameChars = regexp.MustCompile(`[^a-z0-9-]`)
var repeatedDashes = regexp.MustCompile(`-+`)

func main() {
	componentToken := getEnvOrDie(pulumiComponentTokenEnvVar)

	if err := transformInputToProgramOutput(componentToken); err != nil {
		log.Fatalf("%v", err)
	}
}

func transformInputToProgramOutput(componentToken string) error {
	inputFile := getEnvWithDefault("KRATIX_INPUT_FILE", defaultInputFilePath)
	outputFile := getEnvWithDefault("KRATIX_OUTPUT_FILE", defaultOutputFilePath)

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
		requestNamespace = defaultNamespace
	}

	resourceName := buildProgramResourceName(componentToken)
	programName := buildProgramName(requestName, requestNamespace, request.GetKind(), componentToken)

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
	}, "spec", "resources"); err != nil {
		return fmt.Errorf("failed to set spec.resources: %w", err)
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

func buildProgramName(requestName, requestNamespace, requestKind, componentToken string) string {
	base := sanitizeKubernetesName(requestName)
	hashValue := shortHash(fmt.Sprintf("%s/%s/%s/%s", requestNamespace, requestKind, requestName, componentToken))
	name := fmt.Sprintf("%s-%s", base, hashValue)
	if len(name) <= 63 {
		return name
	}

	maxBaseLen := 63 - len(hashValue) - 1
	if maxBaseLen < 1 {
		return hashValue
	}
	return fmt.Sprintf("%s-%s", strings.Trim(base[:maxBaseLen], "-"), hashValue)
}

func buildProgramResourceName(componentToken string) string {
	resourceName := sanitizeKubernetesName(strings.ReplaceAll(componentToken, ":", "-"))
	if len(resourceName) > 63 {
		return strings.Trim(resourceName[:63], "-")
	}
	return resourceName
}

func sanitizeKubernetesName(input string) string {
	value := strings.ToLower(input)
	value = strings.ReplaceAll(value, "_", "-")
	value = invalidNameChars.ReplaceAllString(value, "-")
	value = repeatedDashes.ReplaceAllString(value, "-")
	value = strings.Trim(value, "-")
	if value == "" {
		return "program"
	}
	return value
}

func shortHash(value string) string {
	h := fnv.New32a()
	_, _ = h.Write([]byte(value))
	return fmt.Sprintf("%08x", h.Sum32())
}

func getEnvWithDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getEnvOrDie(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("expected %s to be set", key)
	}
	return value
}
