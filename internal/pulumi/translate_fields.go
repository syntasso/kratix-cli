package pulumi

import (
	"encoding/json"
	"fmt"
	"sort"
)

func parseRequired(node any) ([]string, error) {
	rawValues, ok := node.([]any)
	if !ok {
		return nil, fmt.Errorf("required is not an array")
	}

	result := make([]string, 0, len(rawValues))
	for _, value := range rawValues {
		asString, ok := value.(string)
		if !ok {
			return nil, fmt.Errorf("required entry is not string")
		}
		result = append(result, asString)
	}

	return normalizedRequired(result), nil
}

func normalizedRequired(values []string) []string {
	if len(values) == 0 {
		return nil
	}

	set := make(map[string]struct{}, len(values))
	for _, value := range values {
		set[value] = struct{}{}
	}

	result := make([]string, 0, len(set))
	for value := range set {
		result = append(result, value)
	}
	sort.Strings(result)

	return result
}

func filterRequiredForProperties(required []string, translatedProps map[string]any) []string {
	if len(required) == 0 {
		return nil
	}

	filtered := make([]string, 0, len(required))
	for _, name := range required {
		if _, exists := translatedProps[name]; exists {
			filtered = append(filtered, name)
		}
	}
	return filtered
}

func sortedRawKeys(values map[string]json.RawMessage) []string {
	if len(values) == 0 {
		return nil
	}

	result := make([]string, 0, len(values))
	for key := range values {
		result = append(result, key)
	}
	sort.Strings(result)
	return result
}

func sortedAnyKeys(values map[string]any) []string {
	if len(values) == 0 {
		return nil
	}

	result := make([]string, 0, len(values))
	for key := range values {
		result = append(result, key)
	}
	sort.Strings(result)
	return result
}

func stringField(node map[string]any, key string) (string, bool) {
	value, ok := node[key]
	if !ok {
		return "", false
	}
	asString, ok := value.(string)
	if !ok {
		return "", false
	}
	return asString, true
}

func objectField(node map[string]any, key string) (map[string]any, bool, error) {
	value, ok := node[key]
	if !ok {
		return nil, false, nil
	}
	asMap, ok := value.(map[string]any)
	if !ok {
		return nil, true, fmt.Errorf("field %q is not object", key)
	}
	return asMap, true, nil
}
