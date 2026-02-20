package pulumi

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestTranslateInputsToSpecSchema(t *testing.T) {
	t.Parallel()

	t.Run("translates required fields nested object arrays maps and local refs", func(t *testing.T) {
		t.Parallel()

		doc := SchemaDocument{
			Resources: map[string]SchemaResource{
				"pkg:index:NestedResource": {
					InputProperties: map[string]json.RawMessage{
						"id": json.RawMessage(`{"type":"string"}`),
					},
					RequiredInputs: []string{"id"},
				},
			},
			Types: map[string]json.RawMessage{
				"pkg:index:Settings": json.RawMessage(`{
					"type":"object",
					"properties":{
						"region":{"type":"string"},
						"labels":{"type":"object","additionalProperties":{"type":"string"}}
					},
					"required":["region"]
				}`),
			},
		}

		component := SelectedComponent{
			Token: "pkg:index:Thing",
			Resource: SchemaResource{
				InputProperties: map[string]json.RawMessage{
					"name":     json.RawMessage(`{"type":"string"}`),
					"replicas": json.RawMessage(`{"type":"integer"}`),
					"ports":    json.RawMessage(`{"type":"array","items":{"type":"number"}}`),
					"config":   json.RawMessage(`{"$ref":"#/types/pkg:index:Settings"}`),
					"resource": json.RawMessage(`{"$ref":"#/resources/pkg:index:NestedResource"}`),
				},
				RequiredInputs: []string{"resource", "name"},
			},
		}

		specSchema, warnings, err := TranslateInputsToSpecSchema(doc, component)
		if err != nil {
			t.Fatalf("TranslateInputsToSpecSchema returned error: %v", err)
		}
		if len(warnings) != 0 {
			t.Fatalf("expected no warnings, got: %v", warnings)
		}

		assertSchemaType(t, specSchema, "object")
		assertRequired(t, specSchema, []string{"name", "resource"})

		props := mustProperties(t, specSchema)
		assertSchemaType(t, props["name"].(map[string]any), "string")
		assertSchemaType(t, props["replicas"].(map[string]any), "integer")

		ports := props["ports"].(map[string]any)
		assertSchemaType(t, ports, "array")
		assertSchemaType(t, ports["items"].(map[string]any), "number")

		config := props["config"].(map[string]any)
		assertSchemaType(t, config, "object")
		assertRequired(t, config, []string{"region"})
		configProps := mustProperties(t, config)
		labels := configProps["labels"].(map[string]any)
		assertSchemaType(t, labels, "object")
		additional := labels["additionalProperties"].(map[string]any)
		assertSchemaType(t, additional, "string")

		resource := props["resource"].(map[string]any)
		assertSchemaType(t, resource, "object")
		assertRequired(t, resource, []string{"id"})
	})

	t.Run("skips unsupported nodes and returns deterministic warnings", func(t *testing.T) {
		t.Parallel()

		component := SelectedComponent{
			Token: "pkg:index:Thing",
			Resource: SchemaResource{
				InputProperties: map[string]json.RawMessage{
					"zeta":  json.RawMessage(`{"oneOf":[{"type":"string"},{"type":"number"}]}`),
					"alpha": json.RawMessage(`{"type":"object","properties":{"beta":{"oneOf":[{"type":"string"},{"type":"number"}]},"ok":{"type":"string"},"aardvark":{"anyOf":[{"type":"string"},{"type":"number"}]}},"required":["beta","ok","aardvark"]}`),
				},
			},
		}

		specSchema, warnings, err := TranslateInputsToSpecSchema(SchemaDocument{}, component)
		if err != nil {
			t.Fatalf("TranslateInputsToSpecSchema returned error: %v", err)
		}

		wantWarnings := []string{
			`warning: skipped unsupported schema path "spec.alpha.aardvark" for component "pkg:index:Thing": keyword "anyOf"`,
			`warning: skipped unsupported schema path "spec.alpha.beta" for component "pkg:index:Thing": keyword "oneOf"`,
			`warning: skipped unsupported schema path "spec.zeta" for component "pkg:index:Thing": keyword "oneOf"`,
		}
		if len(warnings) != len(wantWarnings) {
			t.Fatalf("warning count mismatch: got %d want %d (%v)", len(warnings), len(wantWarnings), warnings)
		}
		for i := range wantWarnings {
			if warnings[i] != wantWarnings[i] {
				t.Fatalf("warning mismatch at index %d: got %q want %q", i, warnings[i], wantWarnings[i])
			}
		}

		props := mustProperties(t, specSchema)
		if _, found := props["zeta"]; found {
			t.Fatalf("expected unsupported top-level property to be skipped")
		}
		alpha := props["alpha"].(map[string]any)
		assertRequired(t, alpha, []string{"ok"})
	})

	t.Run("returns error when nothing translatable remains", func(t *testing.T) {
		t.Parallel()

		component := SelectedComponent{
			Token: "pkg:index:Thing",
			Resource: SchemaResource{
				InputProperties: map[string]json.RawMessage{
					"value": json.RawMessage(`{"oneOf":[{"type":"string"},{"type":"number"}]}`),
				},
			},
		}

		_, warnings, err := TranslateInputsToSpecSchema(SchemaDocument{}, component)
		if err == nil {
			t.Fatal("expected an error but got nil")
		}
		if got := err.Error(); !strings.Contains(got, `no translatable spec fields remain`) {
			t.Fatalf("unexpected error: %q", got)
		}
		if len(warnings) != 1 {
			t.Fatalf("expected one warning, got %v", warnings)
		}
	})

	t.Run("uses fallback schema for non-local refs", func(t *testing.T) {
		t.Parallel()

		component := SelectedComponent{
			Token: "pkg:index:Thing",
			Resource: SchemaResource{
				InputProperties: map[string]json.RawMessage{
					"accessData": json.RawMessage(`{"$ref":"/aws/v7.14.0/schema.json#/types/aws:eks%2FAccessScope:AccessScope"}`),
				},
			},
		}

		specSchema, warnings, err := TranslateInputsToSpecSchema(SchemaDocument{}, component)
		if err != nil {
			t.Fatalf("TranslateInputsToSpecSchema returned error: %v", err)
		}
		if len(warnings) != 0 {
			t.Fatalf("expected no warnings, got %v", warnings)
		}

		accessData := mustProperties(t, specSchema)["accessData"].(map[string]any)
		assertSchemaType(t, accessData, "object")
		preserve, ok := accessData["x-kubernetes-preserve-unknown-fields"].(bool)
		if !ok || !preserve {
			t.Fatalf("expected x-kubernetes-preserve-unknown-fields=true, got %#v", accessData["x-kubernetes-preserve-unknown-fields"])
		}
	})

	t.Run("skips refs that include unsupported sibling keywords", func(t *testing.T) {
		t.Parallel()

		doc := SchemaDocument{
			Types: map[string]json.RawMessage{
				"pkg:index:Settings": json.RawMessage(`{"type":"object","properties":{"region":{"type":"string"}}}`),
			},
		}
		component := SelectedComponent{
			Token: "pkg:index:Thing",
			Resource: SchemaResource{
				InputProperties: map[string]json.RawMessage{
					"name":   json.RawMessage(`{"type":"string"}`),
					"config": json.RawMessage(`{"$ref":"#/types/pkg:index:Settings","oneOf":[{"type":"string"},{"type":"number"}]}`),
				},
				RequiredInputs: []string{"config", "name"},
			},
		}

		specSchema, warnings, err := TranslateInputsToSpecSchema(doc, component)
		if err != nil {
			t.Fatalf("TranslateInputsToSpecSchema returned error: %v", err)
		}

		wantWarnings := []string{
			`warning: skipped unsupported schema path "spec.config" for component "pkg:index:Thing": keyword "oneOf"`,
		}
		if len(warnings) != len(wantWarnings) {
			t.Fatalf("warning count mismatch: got %d want %d (%v)", len(warnings), len(wantWarnings), warnings)
		}
		for i := range wantWarnings {
			if warnings[i] != wantWarnings[i] {
				t.Fatalf("warning mismatch at index %d: got %q want %q", i, warnings[i], wantWarnings[i])
			}
		}

		props := mustProperties(t, specSchema)
		if _, found := props["config"]; found {
			t.Fatalf("expected property with unsupported keyword to be skipped")
		}
		assertRequired(t, specSchema, []string{"name"})
	})

	t.Run("returns hard error for cyclic local refs", func(t *testing.T) {
		t.Parallel()

		doc := SchemaDocument{
			Types: map[string]json.RawMessage{
				"pkg:index:A": json.RawMessage(`{"$ref":"#/types/pkg:index:B"}`),
				"pkg:index:B": json.RawMessage(`{"$ref":"#/types/pkg:index:A"}`),
			},
		}
		component := SelectedComponent{
			Token: "pkg:index:Thing",
			Resource: SchemaResource{
				InputProperties: map[string]json.RawMessage{
					"cycle": json.RawMessage(`{"$ref":"#/types/pkg:index:A"}`),
				},
			},
		}

		_, warnings, err := TranslateInputsToSpecSchema(doc, component)
		if err == nil {
			t.Fatal("expected error but got nil")
		}
		if len(warnings) != 0 {
			t.Fatalf("expected no warnings, got %v", warnings)
		}
		if got := err.Error(); !strings.Contains(got, `cyclic local ref`) {
			t.Fatalf("unexpected error: %q", got)
		}
	})
}

func assertSchemaType(t *testing.T, node map[string]any, want string) {
	t.Helper()
	got, ok := node["type"].(string)
	if !ok {
		t.Fatalf("expected schema type string, got %#v", node["type"])
	}
	if got != want {
		t.Fatalf("schema type mismatch: got %q want %q", got, want)
	}
}

func assertRequired(t *testing.T, node map[string]any, want []string) {
	t.Helper()
	raw, ok := node["required"]
	if !ok {
		if len(want) == 0 {
			return
		}
		t.Fatalf("required field missing, expected %v", want)
	}
	values, ok := raw.([]string)
	if !ok {
		t.Fatalf("required is not []string: %#v", raw)
	}
	if len(values) != len(want) {
		t.Fatalf("required length mismatch: got %v want %v", values, want)
	}
	for i := range want {
		if values[i] != want[i] {
			t.Fatalf("required mismatch at index %d: got %q want %q", i, values[i], want[i])
		}
	}
}

func mustProperties(t *testing.T, node map[string]any) map[string]any {
	t.Helper()
	raw, ok := node["properties"]
	if !ok {
		t.Fatalf("properties field missing")
	}
	props, ok := raw.(map[string]any)
	if !ok {
		t.Fatalf("properties is not map[string]any: %#v", raw)
	}
	return props
}
