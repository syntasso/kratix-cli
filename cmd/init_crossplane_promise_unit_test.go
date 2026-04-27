package cmd

import (
	"encoding/json"
	"testing"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

func TestBuildSpecDefault(t *testing.T) {
	tests := []struct {
		name        string
		specProp    apiextensionsv1.JSONSchemaProps
		wantDefault map[string]any // nil means no default should be set
	}{
		{
			name: "no required fields and no defaults",
			specProp: apiextensionsv1.JSONSchemaProps{
				Type: "object",
				Properties: map[string]apiextensionsv1.JSONSchemaProps{
					"image": {Type: "string"},
				},
			},
			wantDefault: map[string]any{},
		},
		{
			name: "defaults but no required fields",
			specProp: apiextensionsv1.JSONSchemaProps{
				Type: "object",
				Properties: map[string]apiextensionsv1.JSONSchemaProps{
					"image": {
						Type:    "string",
						Default: &apiextensionsv1.JSON{Raw: []byte(`"nginx"`)},
					},
				},
			},
			wantDefault: map[string]any{},
		},
		{
			name: "required fields that all have defaults",
			specProp: apiextensionsv1.JSONSchemaProps{
				Type: "object",
				Properties: map[string]apiextensionsv1.JSONSchemaProps{
					"image": {
						Type:    "string",
						Default: &apiextensionsv1.JSON{Raw: []byte(`"nginx"`)},
					},
				},
				Required: []string{"image"},
			},
			wantDefault: map[string]any{"image": "nginx"},
		},
		{
			name: "required fields that do not have defaults",
			specProp: apiextensionsv1.JSONSchemaProps{
				Type: "object",
				Properties: map[string]apiextensionsv1.JSONSchemaProps{
					"resourceConfig": {Type: "object"},
				},
				Required: []string{"resourceConfig"},
			},
			wantDefault: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildSpecDefault(tt.specProp)
			if tt.wantDefault == nil {
				if got != nil {
					t.Fatalf("expected no spec.default, got %s", string(got.Raw))
				}
				return
			}
			if got == nil {
				t.Fatal("expected spec.default to be set, got nil")
			}
			var gotMap map[string]any
			if err := json.Unmarshal(got.Raw, &gotMap); err != nil {
				t.Fatalf("failed to unmarshal result: %v", err)
			}
			gotJSON, _ := json.Marshal(gotMap)
			wantJSON, _ := json.Marshal(tt.wantDefault)
			if string(gotJSON) != string(wantJSON) {
				t.Fatalf("expected spec.default %s, got %s", wantJSON, gotJSON)
			}
		})
	}
}
