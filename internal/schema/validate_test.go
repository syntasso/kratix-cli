package schema

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestValidateForTranslation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		doc         *Document
		wantErrPart string
	}{
		{
			name: "valid schema passes",
			doc: &Document{
				Resources: map[string]Resource{
					"pkg:index:Thing": {
						IsComponent: true,
						InputProperties: map[string]json.RawMessage{
							"mode": json.RawMessage(`{"$ref":"#/types/pkg:index:Mode"}`),
							"spec": json.RawMessage(`{"type":"object","properties":{"replicas":{"type":"integer"},"labels":{"type":"object","additionalProperties":{"type":"string"}},"ports":{"type":"array","items":{"type":"number"}}}}`),
						},
					},
				},
				Types: map[string]json.RawMessage{
					"pkg:index:Mode": json.RawMessage(`{"type":"string","enum":["dev","prod"]}`),
				},
			},
		},
		{
			name: "invalid ref format fails",
			doc: &Document{
				Resources: map[string]Resource{
					"pkg:index:Thing": {
						IsComponent: true,
						InputProperties: map[string]json.RawMessage{
							"value": json.RawMessage(`{"$ref":"#/resources/pkg:index:Mode"}`),
						},
					},
				},
			},
			wantErrPart: `unsupported ref "#/resources/pkg:index:Mode"`,
		},
		{
			name: "unresolved local ref fails",
			doc: &Document{
				Resources: map[string]Resource{
					"pkg:index:Thing": {
						IsComponent: true,
						InputProperties: map[string]json.RawMessage{
							"value": json.RawMessage(`{"$ref":"#/types/pkg:index:Missing"}`),
						},
					},
				},
				Types: map[string]json.RawMessage{},
			},
			wantErrPart: `unresolved local type ref "#/types/pkg:index:Missing"`,
		},
		{
			name: "invalid properties container fails",
			doc: &Document{
				Types: map[string]json.RawMessage{
					"pkg:index:Meta": json.RawMessage(`{"type":"object","properties":[]}`),
				},
			},
			wantErrPart: "properties must be an object schema map",
		},
		{
			name: "invalid items container fails",
			doc: &Document{
				Types: map[string]json.RawMessage{
					"pkg:index:List": json.RawMessage(`{"type":"array","items":[]}`),
				},
			},
			wantErrPart: "items must be an object schema",
		},
		{
			name: "invalid additionalProperties container fails",
			doc: &Document{
				Types: map[string]json.RawMessage{
					"pkg:index:Labels": json.RawMessage(`{"type":"object","additionalProperties":[]}`),
				},
			},
			wantErrPart: "additionalProperties must be an object schema",
		},
		{
			name: "property schema must be object",
			doc: &Document{
				Types: map[string]json.RawMessage{
					"pkg:index:ThingSpec": json.RawMessage(`{"type":"object","properties":{"name":"string"}}`),
				},
			},
			wantErrPart: "property schema must be an object",
		},
		{
			name: "ref target must decode to object node",
			doc: &Document{
				Resources: map[string]Resource{
					"pkg:index:Thing": {
						IsComponent: true,
						InputProperties: map[string]json.RawMessage{
							"value": json.RawMessage(`{"$ref":"#/types/pkg:index:Mode"}`),
						},
					},
				},
				Types: map[string]json.RawMessage{
					"pkg:index:Mode": json.RawMessage(`[]`),
				},
			},
			wantErrPart: "invalid ref target",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := ValidateForTranslation(tt.doc)
			if tt.wantErrPart == "" {
				if err != nil {
					t.Fatalf("ValidateForTranslation error = %v", err)
				}
				return
			}

			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tt.wantErrPart)
			}

			got := err.Error()
			if !strings.Contains(got, tt.wantErrPart) {
				t.Fatalf("error mismatch: got %q, want substring %q", got, tt.wantErrPart)
			}
			if !strings.Contains(got, "schema preflight path") {
				t.Fatalf("expected path-aware preflight error, got %q", got)
			}
			if strings.Contains(got, "\n") {
				t.Fatalf("expected single-line error, got %q", got)
			}
		})
	}
}
