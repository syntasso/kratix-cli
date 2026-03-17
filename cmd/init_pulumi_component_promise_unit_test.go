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

func TestParseSecretKeyRefFlag(t *testing.T) {
	t.Run("returns nil when flag is empty", func(t *testing.T) {
		ref, err := parseSecretKeyRefFlag("", "schema-bearer-token-secret")
		if err != nil {
			t.Fatalf("parseSecretKeyRefFlag() unexpected error: %v", err)
		}
		if ref != nil {
			t.Fatalf("expected nil ref, got %#v", ref)
		}
	})

	t.Run("parses secret name and key", func(t *testing.T) {
		ref, err := parseSecretKeyRefFlag("pulumi-schema-auth:accessToken", "schema-bearer-token-secret")
		if err != nil {
			t.Fatalf("parseSecretKeyRefFlag() unexpected error: %v", err)
		}
		if ref == nil {
			t.Fatal("expected parsed secret ref, got nil")
		}
		if ref.Name != "pulumi-schema-auth" || ref.Key != "accessToken" {
			t.Fatalf("unexpected parsed ref: %#v", ref)
		}
	})

	t.Run("returns a helpful error for invalid input", func(t *testing.T) {
		_, err := parseSecretKeyRefFlag("pulumi-schema-auth", "schema-bearer-token-secret")
		if err == nil || err.Error() != "parse --schema-bearer-token-secret: expected SECRET_NAME:KEY" {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
