package stage

import (
	"fmt"
	"hash/fnv"
	"os"
	"regexp"
	"sort"
	"strings"
)

const (
	DefaultInputFilePath  = "/kratix/input/object.yaml"
	DefaultOutputFilePath = "/kratix/output/object.yaml"
	DefaultNamespace      = "default"
)

var invalidNameChars = regexp.MustCompile(`[^a-z0-9-]`)
var repeatedDashes = regexp.MustCompile(`-+`)

func BuildProgramName(requestName, requestNamespace, requestKind, componentToken string) string {
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

func BuildProgramResourceName(componentToken string) string {
	resourceName := sanitizeKubernetesName(strings.ReplaceAll(componentToken, ":", "-"))
	if len(resourceName) > 63 {
		return strings.Trim(resourceName[:63], "-")
	}
	return resourceName
}

func SortedRawKeys[V any](values map[string]V) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	return sortedStrings(keys)
}

func GetEnvWithDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
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

func sortedStrings(values []string) []string {
	sortedValues := append([]string(nil), values...)
	sort.Strings(sortedValues)
	return sortedValues
}
