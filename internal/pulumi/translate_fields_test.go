package pulumi

import "testing"

func TestParseRequired(t *testing.T) {
	t.Parallel()

	t.Run("normalizes and sorts values", func(t *testing.T) {
		t.Parallel()

		got, err := parseRequired([]any{"b", "a", "b"})
		if err != nil {
			t.Fatalf("parseRequired returned error: %v", err)
		}

		want := []string{"a", "b"}
		if len(got) != len(want) {
			t.Fatalf("length mismatch: got %v want %v", got, want)
		}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("value mismatch at %d: got %q want %q", i, got[i], want[i])
			}
		}
	})

	t.Run("rejects non-array and non-string entries", func(t *testing.T) {
		t.Parallel()

		if _, err := parseRequired("not-array"); err == nil {
			t.Fatal("expected error for non-array required")
		}
		if _, err := parseRequired([]any{"ok", 2}); err == nil {
			t.Fatal("expected error for non-string required entry")
		}
	})
}

func TestFilterRequiredForProperties(t *testing.T) {
	t.Parallel()

	got := filterRequiredForProperties(
		[]string{"missing", "name", "id"},
		map[string]any{"id": map[string]any{}, "name": map[string]any{}},
	)
	want := []string{"name", "id"}
	if len(got) != len(want) {
		t.Fatalf("length mismatch: got %v want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("value mismatch at %d: got %q want %q", i, got[i], want[i])
		}
	}
}
