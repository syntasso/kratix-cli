package pulumi

import (
	"encoding/json"
	"testing"
)

func TestResolveLocalRef(t *testing.T) {
	t.Parallel()

	doc := SchemaDocument{
		Types: map[string]json.RawMessage{
			"pkg:index:Settings": json.RawMessage(`{"type":"string"}`),
		},
		Resources: map[string]SchemaResource{
			"pkg:index:Thing": {
				InputProperties: map[string]json.RawMessage{
					"name": json.RawMessage(`{"type":"string"}`),
				},
				RequiredInputs: []string{"name"},
			},
		},
	}

	t.Run("resolves local type ref", func(t *testing.T) {
		t.Parallel()

		node, key, err := resolveLocalRef(&doc, "#/types/pkg:index:Settings")
		if err != nil {
			t.Fatalf("resolveLocalRef returned error: %v", err)
		}
		if key != "#/types/pkg:index:Settings" {
			t.Fatalf("unexpected key: %q", key)
		}
		if node["type"] != "string" {
			t.Fatalf("unexpected node: %#v", node)
		}
	})

	t.Run("resolves local resource ref", func(t *testing.T) {
		t.Parallel()

		node, key, err := resolveLocalRef(&doc, "#/resources/pkg:index:Thing")
		if err != nil {
			t.Fatalf("resolveLocalRef returned error: %v", err)
		}
		if key != "#/resources/pkg:index:Thing" {
			t.Fatalf("unexpected key: %q", key)
		}
		if node["type"] != "object" {
			t.Fatalf("unexpected node: %#v", node)
		}
	})

	t.Run("returns errors for unsupported or unresolved refs", func(t *testing.T) {
		t.Parallel()

		if _, _, err := resolveLocalRef(&doc, "#/types/"); err == nil {
			t.Fatal("expected invalid local type ref error")
		}
		if _, _, err := resolveLocalRef(&doc, "#/types/pkg:index:Missing"); err == nil {
			t.Fatal("expected unresolved local type ref error")
		}
		if _, _, err := resolveLocalRef(&doc, "https://example.com/schema#/types/foo"); err == nil {
			t.Fatal("expected unsupported ref error")
		}
	})
}

func TestDecodeNode(t *testing.T) {
	t.Parallel()

	if _, err := decodeNode(nil); err == nil {
		t.Fatal("expected error for empty schema node")
	}
	if _, err := decodeNode(json.RawMessage(`null`)); err == nil {
		t.Fatal("expected error for null schema node")
	}
	if _, err := decodeNode(json.RawMessage(`{"type":"string"}`)); err != nil {
		t.Fatalf("expected valid schema node, got error: %v", err)
	}
}

func TestFallbackSchemaForNonLocalRef(t *testing.T) {
	t.Parallel()

	got := fallbackSchemaForNonLocalRef()
	if got["type"] != "object" {
		t.Fatalf("type mismatch: %#v", got)
	}
	preserve, ok := got["x-kubernetes-preserve-unknown-fields"].(bool)
	if !ok || !preserve {
		t.Fatalf("expected preserve-unknown-fields=true, got %#v", got)
	}
}
