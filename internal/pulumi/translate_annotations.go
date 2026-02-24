package pulumi

import (
	"fmt"
	"math"
)

func applyAnnotations(node map[string]any, translated map[string]any, componentToken, path string) (map[string]any, error) {
	if descriptionValue, ok := node["description"]; ok {
		description, ok := descriptionValue.(string)
		if !ok {
			return nil, unsupported(componentToken, path, "description must be a string")
		}
		translated["description"] = description
	}

	if defaultValue, ok := node["default"]; ok {
		translated["default"] = defaultValue
	}

	enumValues, hasEnum, err := parseEnum(node)
	if err != nil {
		return nil, unsupported(componentToken, path, err.Error())
	}
	if hasEnum {
		typeName, hasType := translated["type"].(string)
		if hasType {
			if err := validateEnumValuesForType(enumValues, typeName); err != nil {
				return nil, unsupportedHard(componentToken, path, err.Error())
			}
		}
		translated["enum"] = enumValues
	}

	return translated, nil
}

func parseEnum(node map[string]any) ([]any, bool, error) {
	enumNode, ok := node["enum"]
	if !ok {
		return nil, false, nil
	}

	values, ok := enumNode.([]any)
	if !ok {
		return nil, false, fmt.Errorf("enum must be an array")
	}

	translatedValues := make([]any, 0, len(values))
	for _, item := range values {
		switch typed := item.(type) {
		case map[string]any:
			value, exists := typed["value"]
			if !exists {
				return nil, false, fmt.Errorf("enum entry object missing value field")
			}
			translatedValues = append(translatedValues, value)
		case string, bool, float64, nil:
			translatedValues = append(translatedValues, typed)
		default:
			return nil, false, fmt.Errorf("enum entry has unsupported shape")
		}
	}

	return translatedValues, true, nil
}

func validateEnumValuesForType(values []any, typeName string) error {
	for _, value := range values {
		if !isValueCompatibleWithType(value, typeName) {
			return fmt.Errorf("enum value %v is not compatible with type %q", value, typeName)
		}
	}
	return nil
}

func isValueCompatibleWithType(value any, typeName string) bool {
	switch typeName {
	case "string":
		_, ok := value.(string)
		return ok
	case "boolean":
		_, ok := value.(bool)
		return ok
	case "number":
		_, ok := value.(float64)
		return ok
	case "integer":
		asNumber, ok := value.(float64)
		if !ok {
			return false
		}
		return math.Trunc(asNumber) == asNumber
	case "array":
		_, ok := value.([]any)
		return ok
	case "object":
		_, ok := value.(map[string]any)
		return ok
	default:
		return false
	}
}

func rejectUnsupportedKeywords(node map[string]any, componentToken, path string) error {
	unsupportedKeywords := []string{"oneOf", "anyOf", "allOf", "not", "discriminator", "patternProperties", "const"}
	for _, key := range unsupportedKeywords {
		if _, exists := node[key]; exists {
			return unsupported(componentToken, path, fmt.Sprintf("keyword %q", key))
		}
	}
	return nil
}
