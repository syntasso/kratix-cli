package translate

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/syntasso/pulumi-component-to-crd/internal/schema"
)

func TestInputPropertiesToOpenAPI_SupportedMappings(t *testing.T) {
	t.Parallel()

	doc := &schema.Document{
		Resources: map[string]schema.Resource{
			"pkg:index:ClusterSpec": {
				InputProperties: map[string]json.RawMessage{
					"id":   json.RawMessage(`{"type":"string"}`),
					"tags": json.RawMessage(`{"type":"object","additionalProperties":{"type":"string"}}`),
				},
				RequiredInputs: []string{"id"},
			},
		},
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
			"cluster":   json.RawMessage(`{"$ref":"#/resources/pkg:index:ClusterSpec"}`),
		},
		RequiredInputs: []string{"resources", "name", "cluster"},
	}

	got, skipped, err := InputPropertiesToOpenAPI(doc, "pkg:index:Thing", resource)
	if err != nil {
		t.Fatalf("InputPropertiesToOpenAPI error = %v", err)
	}
	if skipped != nil {
		t.Fatalf("expected no skipped paths, got %#v", skipped)
	}

	want := map[string]any{
		"type": "object",
		"required": []string{
			"cluster",
			"name",
			"resources",
		},
		"properties": map[string]any{
			"cluster": map[string]any{
				"type": "object",
				"required": []string{
					"id",
				},
				"properties": map[string]any{
					"id": map[string]any{"type": "string"},
					"tags": map[string]any{
						"type": "object",
						"additionalProperties": map[string]any{
							"type": "string",
						},
					},
				},
			},
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

func TestInputPropertiesToOpenAPI_UnsupportedConstructIsSkipped(t *testing.T) {
	t.Parallel()

	doc := &schema.Document{}
	resource := schema.Resource{
		InputProperties: map[string]json.RawMessage{
			"name":  json.RawMessage(`{"type":"string"}`),
			"value": json.RawMessage(`{"oneOf":[{"type":"string"},{"type":"number"}]}`),
		},
		RequiredInputs: []string{"name", "value"},
	}

	got, skipped, err := InputPropertiesToOpenAPI(doc, "pkg:index:Thing", resource)
	if err != nil {
		t.Fatalf("InputPropertiesToOpenAPI error = %v", err)
	}

	wantSkipped := []SkippedPathIssue{
		{
			Component: "pkg:index:Thing",
			Path:      "spec.value",
			Reason:    `keyword "oneOf"`,
		},
	}
	if !reflect.DeepEqual(skipped, wantSkipped) {
		t.Fatalf("skipped mismatch\n got: %#v\nwant: %#v", skipped, wantSkipped)
	}

	want := map[string]any{
		"type": "object",
		"required": []string{
			"name",
		},
		"properties": map[string]any{
			"name": map[string]any{
				"type": "string",
			},
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("translated schema mismatch\n got: %#v\nwant: %#v", got, want)
	}
}

func TestInputPropertiesToOpenAPI_DeterministicSkippedOrdering(t *testing.T) {
	t.Parallel()

	doc := &schema.Document{}
	resource := schema.Resource{
		InputProperties: map[string]json.RawMessage{
			"zeta":  json.RawMessage(`{"oneOf":[{"type":"string"},{"type":"number"}]}`),
			"alpha": json.RawMessage(`{"type":"object","properties":{"beta":{"oneOf":[{"type":"string"},{"type":"number"}]},"ok":{"type":"string"},"aardvark":{"anyOf":[{"type":"string"},{"type":"number"}]}},"required":["beta","ok","aardvark"]}`),
		},
	}

	gotA, skippedA, err := InputPropertiesToOpenAPI(doc, "pkg:index:Thing", resource)
	if err != nil {
		t.Fatalf("InputPropertiesToOpenAPI error = %v", err)
	}
	gotB, skippedB, err := InputPropertiesToOpenAPI(doc, "pkg:index:Thing", resource)
	if err == nil {
		// noop
	} else {
		t.Fatalf("InputPropertiesToOpenAPI error = %v", err)
	}

	if !reflect.DeepEqual(gotA, gotB) {
		t.Fatalf("translated schema is not deterministic\nA: %#v\nB: %#v", gotA, gotB)
	}
	if !reflect.DeepEqual(skippedA, skippedB) {
		t.Fatalf("skipped list is not deterministic\nA: %#v\nB: %#v", skippedA, skippedB)
	}

	wantSkipped := []SkippedPathIssue{
		{
			Component: "pkg:index:Thing",
			Path:      "spec.alpha.aardvark",
			Reason:    `keyword "anyOf"`,
		},
		{
			Component: "pkg:index:Thing",
			Path:      "spec.alpha.beta",
			Reason:    `keyword "oneOf"`,
		},
		{
			Component: "pkg:index:Thing",
			Path:      "spec.zeta",
			Reason:    `keyword "oneOf"`,
		},
	}
	if !reflect.DeepEqual(skippedA, wantSkipped) {
		t.Fatalf("skipped mismatch\n got: %#v\nwant: %#v", skippedA, wantSkipped)
	}
	want := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"alpha": map[string]any{
				"type": "object",
				"required": []string{
					"ok",
				},
				"properties": map[string]any{
					"ok": map[string]any{
						"type": "string",
					},
				},
			},
		},
	}
	if !reflect.DeepEqual(gotA, want) {
		t.Fatalf("translated schema mismatch\n got: %#v\nwant: %#v", gotA, want)
	}
}

func TestInputPropertiesToOpenAPI_UnresolvedRefIsInvalidSchemaError(t *testing.T) {
	t.Parallel()

	doc := &schema.Document{Types: map[string]json.RawMessage{}, Resources: map[string]schema.Resource{}}
	resource := schema.Resource{
		InputProperties: map[string]json.RawMessage{
			"value": json.RawMessage(`{"$ref":"#/resources/pkg:index:Missing"}`),
		},
	}

	_, skipped, err := InputPropertiesToOpenAPI(doc, "pkg:index:Thing", resource)
	if err == nil {
		t.Fatal("expected error but got nil")
	}
	if skipped != nil {
		t.Fatalf("expected no skipped paths, got %#v", skipped)
	}
	var unsupportedErr *UnsupportedError
	if errors.As(err, &unsupportedErr) {
		t.Fatalf("expected non-unsupported malformed-schema error, got %T: %v", err, err)
	}
	if got := err.Error(); got == "" || !containsAll(got, `component "pkg:index:Thing"`, `path "spec.value"`, `invalid schema`, `unresolved local resource ref`) {
		t.Fatalf("unexpected error: %q", got)
	}
}

func TestInputPropertiesToOpenAPI_NonLocalRefUsesFallbackSchema(t *testing.T) {
	t.Parallel()

	doc := &schema.Document{}
	resource := schema.Resource{
		InputProperties: map[string]json.RawMessage{
			"name":       json.RawMessage(`{"type":"string"}`),
			"accessData": json.RawMessage(`{"$ref":"/aws/v7.14.0/schema.json#/types/aws:eks%2FAccessScope:AccessScope"}`),
		},
		RequiredInputs: []string{"accessData", "name"},
	}

	got, skipped, err := InputPropertiesToOpenAPI(doc, "pkg:index:Thing", resource)
	if err != nil {
		t.Fatalf("InputPropertiesToOpenAPI error = %v", err)
	}
	if skipped != nil {
		t.Fatalf("expected no skipped paths, got %#v", skipped)
	}

	want := map[string]any{
		"type": "object",
		"required": []string{
			"accessData",
			"name",
		},
		"properties": map[string]any{
			"accessData": map[string]any{
				"type":                                 "object",
				"x-kubernetes-preserve-unknown-fields": true,
			},
			"name": map[string]any{
				"type": "string",
			},
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("translated schema mismatch\n got: %#v\nwant: %#v", got, want)
	}
}

func TestInputPropertiesToOpenAPI_HardUnsupportedStillFails(t *testing.T) {
	t.Parallel()

	doc := &schema.Document{
		Types: map[string]json.RawMessage{
			"pkg:index:A": json.RawMessage(`{"$ref":"#/types/pkg:index:B"}`),
			"pkg:index:B": json.RawMessage(`{"$ref":"#/types/pkg:index:A"}`),
		},
	}
	resource := schema.Resource{
		InputProperties: map[string]json.RawMessage{
			"cycle": json.RawMessage(`{"$ref":"#/types/pkg:index:A"}`),
		},
	}

	_, skipped, err := InputPropertiesToOpenAPI(doc, "pkg:index:Thing", resource)
	if err == nil {
		t.Fatal("expected error but got nil")
	}
	if skipped != nil {
		t.Fatalf("expected no skipped paths, got %#v", skipped)
	}
	var unsupportedErr *UnsupportedError
	if !errors.As(err, &unsupportedErr) {
		t.Fatalf("expected UnsupportedError, got %T: %v", err, err)
	}
	if unsupportedErr.Skippable {
		t.Fatalf("expected hard unsupported error, got skippable: %#v", unsupportedErr)
	}
	if unsupportedErr.Path != "spec.cycle" {
		t.Fatalf("path mismatch: got %q", unsupportedErr.Path)
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
