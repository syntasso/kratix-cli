package emit

import (
	"strings"
	"testing"
)

func TestDefaultIdentity_Valid(t *testing.T) {
	t.Parallel()

	id := DefaultIdentity()
	if err := id.Validate(); err != nil {
		t.Fatalf("default identity should be valid, got %v", err)
	}
	if got, want := id.MetadataName(), "components.components.platform"; got != want {
		t.Fatalf("metadata name mismatch: got %q want %q", got, want)
	}
}

func TestIdentityValidate(t *testing.T) {
	t.Parallel()

	valid := Identity{
		Group:    "apps.example.com",
		Version:  "v1",
		Kind:     "ServiceDeployment",
		Plural:   "servicedeployments",
		Singular: "servicedeployment",
	}

	tests := []struct {
		name        string
		identity    Identity
		wantErrPart string
	}{
		{name: "valid", identity: valid},
		{
			name: "invalid group",
			identity: Identity{
				Group:    "bad_group",
				Version:  valid.Version,
				Kind:     valid.Kind,
				Plural:   valid.Plural,
				Singular: valid.Singular,
			},
			wantErrPart: "invalid --group",
		},
		{
			name: "invalid version",
			identity: Identity{
				Group:    valid.Group,
				Version:  "v1_alpha1",
				Kind:     valid.Kind,
				Plural:   valid.Plural,
				Singular: valid.Singular,
			},
			wantErrPart: "invalid --version",
		},
		{
			name: "empty kind",
			identity: Identity{
				Group:    valid.Group,
				Version:  valid.Version,
				Kind:     "",
				Plural:   valid.Plural,
				Singular: valid.Singular,
			},
			wantErrPart: "invalid --kind",
		},
		{
			name: "kind with dash",
			identity: Identity{
				Group:    valid.Group,
				Version:  valid.Version,
				Kind:     "Service-Deployment",
				Plural:   valid.Plural,
				Singular: valid.Singular,
			},
			wantErrPart: "invalid --kind",
		},
		{
			name: "kind starts with digit",
			identity: Identity{
				Group:    valid.Group,
				Version:  valid.Version,
				Kind:     "1ServiceDeployment",
				Plural:   valid.Plural,
				Singular: valid.Singular,
			},
			wantErrPart: "invalid --kind",
		},
		{
			name: "invalid plural",
			identity: Identity{
				Group:    valid.Group,
				Version:  valid.Version,
				Kind:     valid.Kind,
				Plural:   "bad_plural",
				Singular: valid.Singular,
			},
			wantErrPart: "invalid --plural",
		},
		{
			name: "invalid singular",
			identity: Identity{
				Group:    valid.Group,
				Version:  valid.Version,
				Kind:     valid.Kind,
				Plural:   valid.Plural,
				Singular: "bad.singular",
			},
			wantErrPart: "invalid --singular",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.identity.Validate()
			if tt.wantErrPart == "" {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tt.wantErrPart)
			}
			if got := err.Error(); got == "" || !strings.Contains(got, tt.wantErrPart) {
				t.Fatalf("error mismatch: got %q, want contains %q", got, tt.wantErrPart)
			}
		})
	}
}
