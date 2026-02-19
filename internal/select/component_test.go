package selectcomponent

import (
	"reflect"
	"testing"

	"github.com/pulumi/component-to-crd/internal/schema"
)

func TestDiscoverComponentTokens(t *testing.T) {
	t.Parallel()

	doc := &schema.Document{
		Resources: map[string]schema.Resource{
			"pkg:index:Zeta":  {IsComponent: true},
			"pkg:index:Alpha": {IsComponent: true},
			"pkg:index:Skip":  {IsComponent: false},
		},
	}

	got := DiscoverComponentTokens(doc)
	want := []string{"pkg:index:Alpha", "pkg:index:Zeta"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("tokens mismatch: got %v, want %v", got, want)
	}
}

func TestSelectComponent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		tokens  []string
		request string
		want    string
		wantErr string
	}{
		{
			name:    "explicit valid token",
			tokens:  []string{"a", "b"},
			request: "b",
			want:    "b",
		},
		{
			name:    "explicit unknown token",
			tokens:  []string{"a", "b"},
			request: "c",
			wantErr: `component "c" not found; available components: a, b`,
		},
		{
			name:   "implicit single token",
			tokens: []string{"a"},
			want:   "a",
		},
		{
			name:    "implicit multiple tokens",
			tokens:  []string{"a", "b"},
			wantErr: "multiple components found; provide --component from: a, b",
		},
		{
			name:    "implicit zero tokens",
			tokens:  nil,
			wantErr: "no component resources found in schema",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := SelectComponent(tt.tokens, tt.request)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got != tt.want {
					t.Fatalf("selection mismatch: got %q, want %q", got, tt.want)
				}
				return
			}

			if err == nil {
				t.Fatalf("expected error %q, got nil", tt.wantErr)
			}
			if err.Error() != tt.wantErr {
				t.Fatalf("error mismatch: got %q, want %q", err.Error(), tt.wantErr)
			}
		})
	}
}
