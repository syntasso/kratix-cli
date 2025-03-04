package internal

import (
	"fmt"
	"strings"

	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

// TerraformVariable represents a Terraform input variable
type TerraformVariable struct {
	Name        string
	Type        string
	Description string
}

// VariablesToCRDSpecSchema converts a list of Terraform variables to a CRD JSON schema
func VariablesToCRDSpecSchema(variables []TerraformVariable) (*v1.JSONSchemaProps, error) {
	specSchema := &v1.JSONSchemaProps{
		Type:       "object",
		Properties: make(map[string]v1.JSONSchemaProps),
	}

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

			// Special handling for arrays of maps (e.g., list(map(any)))
			if strings.HasPrefix(innerType, "map(") {
				prop = v1.JSONSchemaProps{
					Type: "array",
					Items: &v1.JSONSchemaPropsOrArray{
						Schema: &v1.JSONSchemaProps{
							Type:                   "object",
							XPreserveUnknownFields: boolPtr(true),
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
			// Treat all maps as open objects that allow unknown fields
			prop = v1.JSONSchemaProps{
				Type:                   "object",
				XPreserveUnknownFields: boolPtr(true),
			}

		case strings.HasPrefix(terraformType, "object("):
			return nil, fmt.Errorf("object types are not yet supported: %s", terraformType)

		case strings.HasPrefix(terraformType, "tuple("):
			return nil, fmt.Errorf("tuple types are not yet supported: %s", terraformType)

		case strings.HasPrefix(terraformType, "set("):
			return nil, fmt.Errorf("set type is not supported: %s", terraformType)

		default:
			return nil, fmt.Errorf("unsupported Terraform type: %s", terraformType)
		}

		if v.Description != "" {
			prop.Description = v.Description
		}

		specSchema.Properties[v.Name] = prop
	}

	return specSchema, nil
}

// extractInnerType extracts the inner type from a Terraform complex type (e.g., "list(string)" -> "string").
func extractInnerType(terraformType, containerType string) string {
	inner := strings.TrimPrefix(terraformType, containerType+"(")
	inner = strings.TrimSuffix(inner, ")")
	return strings.TrimSpace(inner)
}

// boolPtr returns a pointer to a boolean value
func boolPtr(b bool) *bool {
	return &b
}
