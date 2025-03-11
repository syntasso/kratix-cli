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
		prop, warn := convertTerraformTypeToCRD(v.Type)
		if warn != "" {
			warnings = append(warnings, fmt.Sprintf("warning: unable to automatically convert %s into CRD, skipping", v.Type))
			continue
		}

		if v.Description != "" {
			prop.Description = v.Description
		}

		varSchema.Properties[v.Name] = prop
	}

	return specSchema, warnings
}

func convertTerraformTypeToCRD(terraformType string) (v1.JSONSchemaProps, string) {
	terraformType = strings.TrimSpace(terraformType)

	switch {
	case terraformType == "string":
		return v1.JSONSchemaProps{Type: "string"}, ""
	case terraformType == "number":
		return v1.JSONSchemaProps{Type: "number"}, ""
	case terraformType == "bool" || terraformType == "boolean":
		return v1.JSONSchemaProps{Type: "boolean"}, ""

	case strings.HasPrefix(terraformType, "list("):
		innerType := extractInnerType(terraformType, "list")
		prop, warn := convertTerraformTypeToCRD(innerType)
		if warn != "" {
			return v1.JSONSchemaProps{}, "unsupported list type"
		}
		return v1.JSONSchemaProps{
			Type: "array",
			Items: &v1.JSONSchemaPropsOrArray{
				Schema: &prop,
			},
		}, ""

	case strings.HasPrefix(terraformType, "map("):
		innerType := extractInnerType(terraformType, "map")
		prop, warn := convertTerraformTypeToCRD(innerType)
		if warn != "" {
			return v1.JSONSchemaProps{
				Type:                   "object",
				XPreserveUnknownFields: boolPtr(true),
			}, ""
		}
		return v1.JSONSchemaProps{
			Type: "object",
			AdditionalProperties: &v1.JSONSchemaPropsOrBool{
				Schema: &prop,
			},
		}, ""

	case strings.HasPrefix(terraformType, "object("):
		return v1.JSONSchemaProps{
			Type:                   "object",
			XPreserveUnknownFields: boolPtr(true),
		}, ""

	default:
		return v1.JSONSchemaProps{}, "unsupported type"
	}
}

func extractInnerType(terraformType, containerType string) string {
	inner := strings.TrimPrefix(terraformType, containerType+"(")
	inner = strings.TrimSuffix(inner, ")")
	return strings.TrimSpace(inner)
}

func boolPtr(b bool) *bool {
	return &b
}
