package pulumi

import (
	"testing"
)

func TestSelectComponent(t *testing.T) {
	t.Parallel()

	multiComponentDoc := SchemaDocument{
		Resources: map[string]SchemaResource{
			"pkg:index:Zulu":  {IsComponent: true},
			"pkg:index:Alpha": {IsComponent: true},
			"pkg:index:Skip":  {IsComponent: false},
		},
	}

	t.Run("single component auto-selected", func(t *testing.T) {
		t.Parallel()

		doc := SchemaDocument{
			Resources: map[string]SchemaResource{
				"pkg:index:Thing": {IsComponent: true},
			},
		}

		selected, err := SelectComponent(doc, "")
		if err != nil {
			t.Fatalf("SelectComponent returned error: %v", err)
		}
		if selected.Token != "pkg:index:Thing" {
			t.Fatalf("selected token mismatch: got %q, want %q", selected.Token, "pkg:index:Thing")
		}
	})

	t.Run("multiple components require explicit selection", func(t *testing.T) {
		t.Parallel()

		_, err := SelectComponent(multiComponentDoc, "")
		want := "select component: multiple components found; provide --component from: pkg:index:Alpha, pkg:index:Zulu"
		if err == nil || err.Error() != want {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("unknown component returns token list", func(t *testing.T) {
		t.Parallel()

		_, err := SelectComponent(multiComponentDoc, "pkg:index:Missing")
		want := `select component: component "pkg:index:Missing" not found; available components: pkg:index:Alpha, pkg:index:Zulu`
		if err == nil || err.Error() != want {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("explicit component is selected", func(t *testing.T) {
		t.Parallel()

		selected, err := SelectComponent(multiComponentDoc, "pkg:index:Zulu")
		if err != nil {
			t.Fatalf("SelectComponent returned error: %v", err)
		}
		if selected.Token != "pkg:index:Zulu" {
			t.Fatalf("selected token mismatch: got %q, want %q", selected.Token, "pkg:index:Zulu")
		}
	})

	t.Run("zero components returns an explicit error", func(t *testing.T) {
		t.Parallel()

		doc := SchemaDocument{
			Resources: map[string]SchemaResource{
				"pkg:index:Raw": {IsComponent: false},
			},
		}

		_, err := SelectComponent(doc, "")
		want := "select component: no component resources found in schema"
		if err == nil || err.Error() != want {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
