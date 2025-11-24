package internal

import (
	"encoding/json"
	"fmt"
	"strings"

	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

// TerraformVariable represents a Terraform input variable
type TerraformVariable struct {
	Name        string
	Type        string
	Description string
	Default     any
}

func VariablesToCRDSpecSchema(variables []TerraformVariable) (*v1.JSONSchemaProps, []string) {
	varSchema := &v1.JSONSchemaProps{
		Type:       "object",
		Properties: make(map[string]v1.JSONSchemaProps),
		Default:    &v1.JSON{Raw: []byte(`{}`)},
	}

	var warnings []string

	for _, v := range variables {
		if v.Type == "" {
			inferredType := inferTypeFromDefault(v.Default)
			if inferredType == "" {
				warnings = append(warnings, fmt.Sprintf("warning: Type not set for variable %s and cannot be inferred from the default value, skipping", v.Name))
				continue
			}
			v.Type = inferredType
		}

		prop, warn := convertTerraformTypeToCRD(v.Type)
		if warn != "" {
			warnings = append(warnings, fmt.Sprintf("warning: unable to automatically convert %s of type %s into CRD, skipping", v.Name, v.Type))
			continue
		}

		if v.Description != "" {
			prop.Description = v.Description
		}

		if v.Default != nil {
			if strings.Contains(v.Type, "object") {
				warnings = append(warnings, fmt.Sprintf("warning: default value for variable %s is set but type %s does not support defaults, skipping", v.Name, v.Type))
			} else {
				raw, err := json.Marshal(v.Default)
				if err != nil {
					warnings = append(warnings, fmt.Sprintf("warning: failed to marshal default value for variable %s: %v", v.Name, err))
				} else {
					prop.Default = &v1.JSON{Raw: raw}
				}
			}
		}

		varSchema.Properties[v.Name] = prop
	}

	return varSchema, warnings
}

func inferTypeFromDefault(value any) string {
	switch v := value.(type) {
	case string:
		return "string"
	case float64, int:
		return "number"
	case bool:
		return "boolean"
	case []any:
		if len(v) > 0 {
			innerType := inferTypeFromDefault(v[0])
			if innerType != "" {
				return fmt.Sprintf("list(%s)", innerType)
			}
		}
		return "list"
	default:
		return ""
	}
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
