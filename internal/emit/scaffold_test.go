package emit

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/pulumi/component-to-crd/internal/schema"
)

func TestRenderScaffoldYAML_IsDeterministic(t *testing.T) {
	t.Parallel()

	resource := schema.Resource{
		InputProperties: map[string]json.RawMessage{
			"zeta":  []byte(`{"type":"string"}`),
			"alpha": []byte(`{"type":"number"}`),
		},
		RequiredInputs: []string{"zeta", "alpha"},
	}

	first := string(RenderScaffoldYAML("pkg:index:Thing", resource))
	second := string(RenderScaffoldYAML("pkg:index:Thing", resource))

	if first != second {
		t.Fatalf("scaffold output should be deterministic; first:\n%s\nsecond:\n%s", first, second)
	}

	requiredSnippets := []string{
		"apiVersion: apiextensions.k8s.io/v1",
		"kind: CustomResourceDefinition",
		"openAPIV3Schema:",
		"properties: {}",
		"translation is not implemented yet.",
		"observed inputProperties=2 (alpha, zeta), requiredInputs=2 (alpha, zeta)",
	}

	for _, snippet := range requiredSnippets {
		if !strings.Contains(first, snippet) {
			t.Fatalf("expected snippet %q in scaffold:\n%s", snippet, first)
		}
	}
}
