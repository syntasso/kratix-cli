package translate

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/pulumi/component-to-crd/internal/schema"
)

func TestInputPropertiesToOpenAPI_SupportedMappings(t *testing.T) {
	t.Parallel()

	doc := &schema.Document{
		Types: map[string]json.RawMessage{
			"pkg:index:Resources": json.RawMessage(`{"type":"object","properties":{"cpu":{"type":"string"},"memory":{"type":"string"}},"required":["memory","cpu"]}`),
			"pkg:index:Mode":      json.RawMessage(`{"type":"string","enum":[{"name":"dev","value":"dev"},{"name":"prod","value":"prod"}]}`),
		},
	}
	resource := schema.Resource{
		InputProperties: map[string]json.RawMessage{
			"name":      json.RawMessage(`{"type":"string","default":"guestbook"}`),
			"replicas":  json.RawMessage(`{"type":"integer","default":2}`),
			"ports":     json.RawMessage(`{"type":"array","items":{"type":"integer"}}`),
			"resources": json.RawMessage(`{"$ref":"#/types/pkg:index:Resources"}`),
			"labels":    json.RawMessage(`{"type":"object","additionalProperties":{"type":"string"}}`),
			"mode":      json.RawMessage(`{"$ref":"#/types/pkg:index:Mode"}`),
		},
		RequiredInputs: []string{"resources", "name"},
	}

	got, err := InputPropertiesToOpenAPI(doc, "pkg:index:Thing", resource)
	if err != nil {
		t.Fatalf("InputPropertiesToOpenAPI error = %v", err)
	}

	want := map[string]any{
		"type": "object",
		"required": []string{
			"name",
			"resources",
		},
		"properties": map[string]any{
			"labels": map[string]any{
				"type": "object",
				"additionalProperties": map[string]any{
					"type": "string",
				},
			},
			"mode": map[string]any{
				"type": "string",
				"enum": []any{"dev", "prod"},
			},
			"name": map[string]any{
				"type":    "string",
				"default": "guestbook",
			},
			"ports": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "integer",
				},
			},
			"replicas": map[string]any{
				"type":    "integer",
				"default": float64(2),
			},
			"resources": map[string]any{
				"type": "object",
				"required": []string{
					"cpu",
					"memory",
				},
				"properties": map[string]any{
					"cpu":    map[string]any{"type": "string"},
					"memory": map[string]any{"type": "string"},
				},
			},
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("translated schema mismatch\n got: %#v\nwant: %#v", got, want)
	}
}

func TestInputPropertiesToOpenAPI_UnsupportedConstruct(t *testing.T) {
	t.Parallel()

	doc := &schema.Document{}
	resource := schema.Resource{
		InputProperties: map[string]json.RawMessage{
			"value": json.RawMessage(`{"oneOf":[{"type":"string"},{"type":"number"}]}`),
		},
	}

	_, err := InputPropertiesToOpenAPI(doc, "pkg:index:Thing", resource)
	if err == nil {
		t.Fatal("expected error but got nil")
	}

	var unsupportedErr *UnsupportedError
	if !errors.As(err, &unsupportedErr) {
		t.Fatalf("expected UnsupportedError, got %T: %v", err, err)
	}
	if unsupportedErr.Component != "pkg:index:Thing" {
		t.Fatalf("component mismatch: got %q", unsupportedErr.Component)
	}
	if unsupportedErr.Path != "spec.value" {
		t.Fatalf("path mismatch: got %q", unsupportedErr.Path)
	}
}

func TestInputPropertiesToOpenAPI_UnresolvedRefIsInvalidSchemaError(t *testing.T) {
	t.Parallel()

	doc := &schema.Document{Types: map[string]json.RawMessage{}}
	resource := schema.Resource{
		InputProperties: map[string]json.RawMessage{
			"value": json.RawMessage(`{"$ref":"#/types/pkg:index:Missing"}`),
		},
	}

	_, err := InputPropertiesToOpenAPI(doc, "pkg:index:Thing", resource)
	if err == nil {
		t.Fatal("expected error but got nil")
	}
	var unsupportedErr *UnsupportedError
	if errors.As(err, &unsupportedErr) {
		t.Fatalf("expected non-unsupported malformed-schema error, got %T: %v", err, err)
	}
	if got := err.Error(); got == "" || !containsAll(got, `component "pkg:index:Thing"`, `path "spec.value"`, `invalid schema`, `unresolved local type ref`) {
		t.Fatalf("unexpected error: %q", got)
	}
}

func containsAll(got string, parts ...string) bool {
	for _, part := range parts {
		if !strings.Contains(got, part) {
			return false
		}
	}
	return true
}
