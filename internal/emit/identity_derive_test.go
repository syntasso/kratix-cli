package emit

import (
	"testing"

	"github.com/pulumi/component-to-crd/internal/schema"
)

func TestDeriveIdentityDefaults(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		doc           *schema.Document
		selectedToken string
		want          Identity
	}{
		{
			name: "derives from valid token and schema name",
			doc: &schema.Document{
				Name: "awsx",
			},
			selectedToken: "awsx:ecs:FargateService",
			want: Identity{
				Group:    "awsx.components.platform",
				Version:  DefaultVersion,
				Kind:     "FargateService",
				Plural:   "fargate-services",
				Singular: "fargate-service",
			},
		},
		{
			name: "falls back kind for malformed token",
			doc: &schema.Document{
				Name: "awsx",
			},
			selectedToken: "awsx:ecs",
			want: Identity{
				Group:    "awsx.components.platform",
				Version:  DefaultVersion,
				Kind:     DefaultKind,
				Plural:   DefaultPlural,
				Singular: DefaultSingular,
			},
		},
		{
			name: "group falls back to token package and sanitizes schema name",
			doc: &schema.Document{
				Name: "aws_x!!",
			},
			selectedToken: "pkg-name:index:Thing",
			want: Identity{
				Group:    "aws-x.components.platform",
				Version:  DefaultVersion,
				Kind:     "Thing",
				Plural:   "things",
				Singular: "thing",
			},
		},
		{
			name: "group falls back constant when no valid package key",
			doc: &schema.Document{
				Name: "---",
			},
			selectedToken: "::Thing",
			want: Identity{
				Group:    DefaultGroup,
				Version:  DefaultVersion,
				Kind:     "Thing",
				Plural:   "things",
				Singular: "thing",
			},
		},
		{
			name: "version maps stable major",
			doc: &schema.Document{
				Name:    "awsx",
				Version: "2.3.4",
			},
			selectedToken: "awsx:ecs:FargateService",
			want: Identity{
				Group:    "awsx.components.platform",
				Version:  "v2",
				Kind:     "FargateService",
				Plural:   "fargate-services",
				Singular: "fargate-service",
			},
		},
		{
			name: "version falls back for invalid and zero major",
			doc: &schema.Document{
				Name:    "awsx",
				Version: "v0.5.0",
			},
			selectedToken: "awsx:ecs:Class",
			want: Identity{
				Group:    "awsx.components.platform",
				Version:  DefaultVersion,
				Kind:     "Class",
				Plural:   "classes",
				Singular: "class",
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := DeriveIdentityDefaults(tt.doc, tt.selectedToken)
			if got != tt.want {
				t.Fatalf("derived identity mismatch: got %#v want %#v", got, tt.want)
			}
			if err := got.Validate(); err != nil {
				t.Fatalf("derived identity should validate: %v", err)
			}
		})
	}
}

func TestToKebabCase(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in   string
		want string
	}{
		{in: "FargateService", want: "fargate-service"},
		{in: "HTTPServer", want: "http-server"},
		{in: "Node2Group", want: "node2-group"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.in, func(t *testing.T) {
			t.Parallel()
			if got := toKebabCase(tt.in); got != tt.want {
				t.Fatalf("toKebabCase(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
