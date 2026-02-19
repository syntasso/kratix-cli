package emit

import (
	"strings"
	"testing"
)

func TestRenderCRDYAML_IsDeterministic(t *testing.T) {
	t.Parallel()

	schema := map[string]any{
		"type": "object",
		"required": []string{
			"alpha",
			"zeta",
		},
		"properties": map[string]any{
			"zeta": map[string]any{"type": "string"},
			"alpha": map[string]any{
				"type":    "number",
				"default": float64(42),
			},
		},
	}

	first, err := RenderCRDYAML(DefaultIdentity(), schema)
	if err != nil {
		t.Fatalf("RenderCRDYAML error: %v", err)
	}
	second, err := RenderCRDYAML(DefaultIdentity(), schema)
	if err != nil {
		t.Fatalf("RenderCRDYAML error: %v", err)
	}

	if string(first) != string(second) {
		t.Fatalf("output should be deterministic; first:\n%s\nsecond:\n%s", string(first), string(second))
	}

	requiredSnippets := []string{
		"apiVersion: apiextensions.k8s.io/v1",
		"kind: CustomResourceDefinition",
		`name: "components.components.platform"`,
		"group: components.platform",
		"kind: Component",
		"plural: components",
		"singular: component",
		"name: v1alpha1",
		"openAPIV3Schema:",
		"spec:",
		"default: 42",
		"required:",
		"- \"alpha\"",
		"- \"zeta\"",
		"alpha:",
		"zeta:",
	}

	out := string(first)
	for _, snippet := range requiredSnippets {
		if !strings.Contains(out, snippet) {
			t.Fatalf("expected snippet %q in output:\n%s", snippet, out)
		}
	}
}

func TestRenderCRDYAML_NilSchema(t *testing.T) {
	t.Parallel()

	_, err := RenderCRDYAML(DefaultIdentity(), nil)
	if err == nil {
		t.Fatal("expected error but got nil")
	}
}

func TestRenderCRDYAML_UsesConfiguredIdentity(t *testing.T) {
	t.Parallel()

	schema := map[string]any{
		"type":       "object",
		"properties": map[string]any{"name": map[string]any{"type": "string"}},
	}
	identity := Identity{
		Group:    "apps.example.com",
		Version:  "v1",
		Kind:     "ServiceDeployment",
		Plural:   "servicedeployments",
		Singular: "servicedeployment",
	}

	outBytes, err := RenderCRDYAML(identity, schema)
	if err != nil {
		t.Fatalf("RenderCRDYAML error: %v", err)
	}

	out := string(outBytes)
	requiredSnippets := []string{
		`name: "servicedeployments.apps.example.com"`,
		"group: apps.example.com",
		"kind: ServiceDeployment",
		"plural: servicedeployments",
		"singular: servicedeployment",
		"name: v1",
	}
	for _, snippet := range requiredSnippets {
		if !strings.Contains(out, snippet) {
			t.Fatalf("expected snippet %q in output:\n%s", snippet, out)
		}
	}
}

func TestRenderCRDYAML_IncludesDescriptionFields(t *testing.T) {
	t.Parallel()

	schema := map[string]any{
		"type":        "object",
		"description": "spec description",
		"properties": map[string]any{
			"name": map[string]any{
				"type":        "string",
				"description": "name field",
			},
			"ports": map[string]any{
				"type":        "array",
				"description": "ports field",
				"items": map[string]any{
					"type":        "integer",
					"description": "port item",
				},
			},
		},
	}

	outBytes, err := RenderCRDYAML(DefaultIdentity(), schema)
	if err != nil {
		t.Fatalf("RenderCRDYAML error: %v", err)
	}

	out := string(outBytes)
	requiredSnippets := []string{
		`description: "spec description"`,
		`name:`,
		`description: "name field"`,
		`ports:`,
		`description: "ports field"`,
		`description: "port item"`,
	}
	for _, snippet := range requiredSnippets {
		if !strings.Contains(out, snippet) {
			t.Fatalf("expected snippet %q in output:\n%s", snippet, out)
		}
	}
}
