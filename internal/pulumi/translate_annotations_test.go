package pulumi

import "testing"

func TestApplyAnnotations(t *testing.T) {
	t.Parallel()

	t.Run("copies description default and enum object values", func(t *testing.T) {
		t.Parallel()

		node := map[string]any{
			"type":        "string",
			"description": "mode selection",
			"default":     "dev",
			"enum": []any{
				map[string]any{"value": "dev"},
				"prod",
			},
		}
		translated := map[string]any{"type": "string"}

		got, err := applyAnnotations(node, translated, "pkg:index:Thing", "spec.mode")
		if err != nil {
			t.Fatalf("applyAnnotations returned error: %v", err)
		}
		if got["description"] != "mode selection" || got["default"] != "dev" {
			t.Fatalf("unexpected annotations: %#v", got)
		}
		enumValues, ok := got["enum"].([]any)
		if !ok || len(enumValues) != 2 {
			t.Fatalf("unexpected enum: %#v", got["enum"])
		}
		if enumValues[0] != "dev" || enumValues[1] != "prod" {
			t.Fatalf("unexpected enum values: %#v", enumValues)
		}
	})

	t.Run("rejects invalid description and enum compatibility", func(t *testing.T) {
		t.Parallel()

		if _, err := applyAnnotations(
			map[string]any{"description": 123},
			map[string]any{"type": "string"},
			"pkg:index:Thing",
			"spec.mode",
		); err == nil {
			t.Fatal("expected description type error")
		}

		if _, err := applyAnnotations(
			map[string]any{"enum": []any{"on", 2.0}},
			map[string]any{"type": "string"},
			"pkg:index:Thing",
			"spec.mode",
		); err == nil {
			t.Fatal("expected enum compatibility error")
		}
	})
}

func TestRejectUnsupportedKeywords(t *testing.T) {
	t.Parallel()

	err := rejectUnsupportedKeywords(
		map[string]any{"oneOf": []any{}},
		"pkg:index:Thing",
		"spec.value",
	)
	if err == nil {
		t.Fatal("expected unsupported keyword error")
	}

	unsupportedErr, ok := err.(*unsupportedError)
	if !ok {
		t.Fatalf("unexpected error type: %T", err)
	}
	if !unsupportedErr.skippable {
		t.Fatalf("expected keyword rejection to be skippable: %#v", unsupportedErr)
	}
}
