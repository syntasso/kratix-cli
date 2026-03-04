package cmd

import "testing"

func TestBuildPulumiCRDDoesNotSetInvalidSpecDefaultForRequiredFields(t *testing.T) {
	originalGroup := group
	originalKind := kind
	originalVersion := version
	originalPlural := plural
	t.Cleanup(func() {
		group = originalGroup
		kind = originalKind
		version = originalVersion
		plural = originalPlural
	})

	group = "veeam.com"
	kind = "User"
	version = "v1alpha1"
	plural = "users"

	specSchema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{
				"type": "string",
			},
		},
		"required": []string{"name"},
	}

	crd, err := buildPulumiCRD(specSchema)
	if err != nil {
		t.Fatalf("buildPulumiCRD() unexpected error: %v", err)
	}

	spec := crd.Spec.Versions[0].Schema.OpenAPIV3Schema.Properties["spec"]
	if spec.Default != nil {
		t.Fatalf("expected spec.default to be unset, got %q", string(spec.Default.Raw))
	}
}
