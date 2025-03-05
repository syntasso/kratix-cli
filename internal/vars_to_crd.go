package internal

import (
	"fmt"
	"regexp"
	"strings"

	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

// TerraformVariable represents a Terraform input variable
type TerraformVariable struct {
	Name        string
	Type        string
	Description string
}

// VariablesToCRDSpecSchema converts a list of Terraform variables to a CRD JSON schema and returns warnings for unsupported types
func VariablesToCRDSpecSchema(variables []TerraformVariable) (*v1.JSONSchemaProps, []string) {
	varSchema := &v1.JSONSchemaProps{
		Type:       "object",
		Properties: make(map[string]v1.JSONSchemaProps),
	}

	specSchema := &v1.JSONSchemaProps{
		Type:       "object",
		Properties: map[string]v1.JSONSchemaProps{"vars": *varSchema},
	}

	var warnings []string

	for _, v := range variables {
		var prop v1.JSONSchemaProps
		terraformType := strings.TrimSpace(v.Type)

		switch {
		case terraformType == "string":
			prop = v1.JSONSchemaProps{Type: "string"}

		case terraformType == "number":
			prop = v1.JSONSchemaProps{Type: "number"}

		case terraformType == "bool" || terraformType == "boolean":
			prop = v1.JSONSchemaProps{Type: "boolean"}

		case strings.HasPrefix(terraformType, "list("):
			innerType := extractInnerType(terraformType, "list")

			if strings.HasPrefix(innerType, "map(") {
				// Handle `list(map(string))` as an array of open objects
				prop = v1.JSONSchemaProps{
					Type: "array",
					Items: &v1.JSONSchemaPropsOrArray{
						Schema: &v1.JSONSchemaProps{
							Type:                   "object",
							XPreserveUnknownFields: boolPtr(true),
						},
					},
				}
			} else if strings.HasPrefix(innerType, "object(") {
				objectProps, err := parseObjectType(innerType)
				if err != nil {
					warnings = append(warnings, fmt.Sprintf("warning: unable to automatically convert %s into CRD, skipping", terraformType))
					continue
				}
				prop = v1.JSONSchemaProps{
					Type: "array",
					Items: &v1.JSONSchemaPropsOrArray{
						Schema: &v1.JSONSchemaProps{
							Type:       "object",
							Properties: objectProps,
						},
					},
				}
			} else {
				prop = v1.JSONSchemaProps{
					Type: "array",
					Items: &v1.JSONSchemaPropsOrArray{
						Schema: &v1.JSONSchemaProps{Type: innerType},
					},
				}
			}

		case strings.HasPrefix(terraformType, "map("):
			innerType := extractInnerType(terraformType, "map")

			if innerType == "string" || innerType == "number" || innerType == "bool" || innerType == "boolean" {
				prop = v1.JSONSchemaProps{
					Type: "object",
					AdditionalProperties: &v1.JSONSchemaPropsOrBool{
						Schema: &v1.JSONSchemaProps{Type: innerType},
					},
				}
			} else {
				// Treat complex maps as open objects
				prop = v1.JSONSchemaProps{
					Type:                   "object",
					XPreserveUnknownFields: boolPtr(true),
				}
			}

		case strings.HasPrefix(terraformType, "object("):
			objectProps, err := parseObjectType(terraformType)
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("warning: unable to automatically convert %s into CRD, skipping", terraformType))
				continue
			}
			prop = v1.JSONSchemaProps{
				Type:       "object",
				Properties: objectProps,
			}

		default:
			warnings = append(warnings, fmt.Sprintf("warning: unable to automatically convert %s into CRD, skipping", terraformType))
			continue
		}

		if v.Description != "" {
			prop.Description = v.Description
		}

		varSchema.Properties[v.Name] = prop
	}

	return specSchema, warnings
}

// extractInnerType extracts the inner type from a Terraform complex type (e.g., "list(string)" -> "string").
func extractInnerType(terraformType, containerType string) string {
	inner := strings.TrimPrefix(terraformType, containerType+"(")
	inner = strings.TrimSuffix(inner, ")")
	return strings.TrimSpace(inner)
}

// parseObjectType parses a Terraform object type definition into OpenAPI properties
func parseObjectType(terraformType string) (map[string]v1.JSONSchemaProps, error) {
	objectBody := strings.TrimPrefix(terraformType, "object(")
	objectBody = strings.TrimSuffix(objectBody, ")")

	fieldRegex := regexp.MustCompile(`\s*(\w+)\s*=\s*(\w+)`)

	properties := make(map[string]v1.JSONSchemaProps)

	matches := fieldRegex.FindAllStringSubmatch(objectBody, -1)
	for _, match := range matches {
		if len(match) < 3 {
			return nil, fmt.Errorf("invalid object field format: %s", match)
		}
		fieldName := match[1]
		fieldType := match[2]

		mappedType, err := mapTerraformType(fieldType)
		if err != nil {
			return nil, fmt.Errorf("unsupported field type: %s", fieldType)
		}

		properties[fieldName] = v1.JSONSchemaProps{Type: mappedType}
	}

	return properties, nil
}

// mapTerraformType converts a Terraform type to a Kubernetes CRD type.
func mapTerraformType(terraformType string) (string, error) {
	switch terraformType {
	case "string":
		return "string", nil
	case "number":
		return "number", nil
	case "bool", "boolean":
		return "boolean", nil
	case "list(string)":
		return "array", nil
	case "map(string)":
		return "object", nil
	default:
		return "", fmt.Errorf("unsupported inner type: %s", terraformType)
	}
}

// boolPtr returns a pointer to a boolean value
func boolPtr(b bool) *bool {
	return &b
}
