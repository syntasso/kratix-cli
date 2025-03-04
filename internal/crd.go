package internal

import (
	"fmt"
	"strings"

	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

type TerraformVariable struct {
	Name        string
	Type        string
	Description string
}

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
			mappedType, err := mapTerraformType(innerType)
			if err != nil {
				return nil, fmt.Errorf("unsupported list type: %s", innerType)
			}
			prop = v1.JSONSchemaProps{
				Type: "array",
				Items: &v1.JSONSchemaPropsOrArray{
					Schema: &v1.JSONSchemaProps{Type: mappedType},
				},
			}

		case strings.HasPrefix(terraformType, "map("):
			innerType := extractInnerType(terraformType, "map")
			mappedType, err := mapTerraformType(innerType)
			if err != nil {
				return nil, fmt.Errorf("unsupported map value type: %s", innerType)
			}
			prop = v1.JSONSchemaProps{
				Type:                 "object",
				AdditionalProperties: &v1.JSONSchemaPropsOrBool{Schema: &v1.JSONSchemaProps{Type: mappedType}},
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

func extractInnerType(terraformType, containerType string) string {
	inner := strings.TrimPrefix(terraformType, containerType+"(")
	inner = strings.TrimSuffix(inner, ")")
	return strings.TrimSpace(inner)
}

func mapTerraformType(terraformType string) (string, error) {
	switch terraformType {
	case "string":
		return "string", nil
	case "number":
		return "number", nil
	case "bool", "boolean":
		return "boolean", nil
	default:
		return "", fmt.Errorf("unsupported inner type: %s", terraformType)
	}
}
