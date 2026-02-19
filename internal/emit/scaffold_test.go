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

	first, err := RenderCRDYAML("pkg:index:Thing", schema)
	if err != nil {
		t.Fatalf("RenderCRDYAML error: %v", err)
	}
	second, err := RenderCRDYAML("pkg:index:Thing", schema)
	if err != nil {
		t.Fatalf("RenderCRDYAML error: %v", err)
	}

	if string(first) != string(second) {
		t.Fatalf("output should be deterministic; first:\n%s\nsecond:\n%s", string(first), string(second))
	}

	requiredSnippets := []string{
		"apiVersion: apiextensions.k8s.io/v1",
		"kind: CustomResourceDefinition",
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

	_, err := RenderCRDYAML("pkg:index:Thing", nil)
	if err == nil {
		t.Fatal("expected error but got nil")
	}
}
