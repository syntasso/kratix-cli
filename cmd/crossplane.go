package cmd

import (
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

var mandatoryAdditionalClaimFields = map[string]apiextensionsv1.JSONSchemaProps{
	"compositeDeletePolicy": {
		Type:    "string",
		Enum:    []apiextensionsv1.JSON{{Raw: []byte(`"Background"`)}, {Raw: []byte(`"Foreground"`)}},
		Default: &apiextensionsv1.JSON{Raw: []byte(`"Background"`)},
	},
	"compositionRef": {
		Type: "object",
		Properties: map[string]apiextensionsv1.JSONSchemaProps{
			"name": {Type: "string"},
		},
		Required: []string{"name"},
	},
	"compositionRevisionRef": {
		Type: "object",
		Properties: map[string]apiextensionsv1.JSONSchemaProps{
			"name": {Type: "string"},
		},
		Required: []string{"name"},
	},
	"compositionRevisionSelector": {
		Type: "object",
		Properties: map[string]apiextensionsv1.JSONSchemaProps{
			"matchLabels": {
				Type:                 "object",
				AdditionalProperties: &apiextensionsv1.JSONSchemaPropsOrBool{Schema: &apiextensionsv1.JSONSchemaProps{Type: "string"}},
			},
		},
		Required: []string{"matchLabels"},
	},
	"compositionSelector": {
		Type: "object",
		Properties: map[string]apiextensionsv1.JSONSchemaProps{
			"matchLabels": {
				Type:                 "object",
				AdditionalProperties: &apiextensionsv1.JSONSchemaPropsOrBool{Schema: &apiextensionsv1.JSONSchemaProps{Type: "string"}},
			},
		},
		Required: []string{"matchLabels"},
	},
	"compositionUpdatePolicy": {
		Type: "string",
		Enum: []apiextensionsv1.JSON{
			{Raw: []byte(`"Automatic"`)},
			{Raw: []byte(`"Manual"`)},
		},
	},
	"publishConnectionDetailsTo": {
		Type: "object",
		Properties: map[string]apiextensionsv1.JSONSchemaProps{
			"configRef": {
				Type: "object",
				Properties: map[string]apiextensionsv1.JSONSchemaProps{
					"name": {Type: "string"},
				},
				Default: &apiextensionsv1.JSON{Raw: []byte(`{"name": "default"}`)},
			},
			"metadata": {
				Type: "object",
				Properties: map[string]apiextensionsv1.JSONSchemaProps{
					"annotations": {
						Type:                 "object",
						AdditionalProperties: &apiextensionsv1.JSONSchemaPropsOrBool{Schema: &apiextensionsv1.JSONSchemaProps{Type: "string"}},
					},
					"labels": {
						Type:                 "object",
						AdditionalProperties: &apiextensionsv1.JSONSchemaPropsOrBool{Schema: &apiextensionsv1.JSONSchemaProps{Type: "string"}},
					},
					"type": {Type: "string"},
				},
			},
			"name": {Type: "string"},
		},
		Required: []string{"name"},
	},
	"resourceRef": {
		Type: "object",
		Properties: map[string]apiextensionsv1.JSONSchemaProps{
			"apiVersion": {Type: "string"},
			"kind":       {Type: "string"},
			"name":       {Type: "string"},
		},
		Required: []string{"apiVersion", "kind", "name"},
	},
	"writeConnectionSecretToRef": {
		Type: "object",
		Properties: map[string]apiextensionsv1.JSONSchemaProps{
			"name": {Type: "string"},
		},
		Required: []string{"name"},
	},
}
