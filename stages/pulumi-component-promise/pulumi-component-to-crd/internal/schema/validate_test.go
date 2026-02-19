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
							"mode":             json.RawMessage(`{"$ref":"#/types/pkg:index:Mode"}`),
							"specFromResource": json.RawMessage(`{"$ref":"#/resources/pkg:index:ClusterSpec"}`),
							"spec":             json.RawMessage(`{"type":"object","properties":{"replicas":{"type":"integer"},"labels":{"type":"object","additionalProperties":{"type":"string"}},"ports":{"type":"array","items":{"type":"number"}}}}`),
						},
					},
					"pkg:index:ClusterSpec": {
						InputProperties: map[string]json.RawMessage{
							"name": json.RawMessage(`{"type":"string"}`),
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
							"value": json.RawMessage(`{"$ref":"#/providers/pkg:index:Mode"}`),
						},
					},
				},
			},
			wantErrPart: `unsupported ref`,
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
			name: "unresolved local resource ref fails",
			doc: &Document{
				Resources: map[string]Resource{
					"pkg:index:Thing": {
						IsComponent: true,
						InputProperties: map[string]json.RawMessage{
							"value": json.RawMessage(`{"$ref":"#/resources/pkg:index:Missing"}`),
						},
					},
				},
			},
			wantErrPart: `unresolved local resource ref "#/resources/pkg:index:Missing"`,
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

func TestValidateForTranslationComponent(t *testing.T) {
	t.Parallel()

	t.Run("reachable malformed ref fails", func(t *testing.T) {
		t.Parallel()

		doc := &Document{
			Resources: map[string]Resource{
				"pkg:index:Selected": {
					IsComponent: true,
					InputProperties: map[string]json.RawMessage{
						"value": json.RawMessage(`{"$ref":"#/types/pkg:index:Missing"}`),
					},
				},
				"pkg:index:Other": {
					IsComponent: true,
					InputProperties: map[string]json.RawMessage{
						"ok": json.RawMessage(`{"type":"string"}`),
					},
				},
			},
			Types: map[string]json.RawMessage{},
		}

		err := ValidateForTranslationComponent(doc, "pkg:index:Selected")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), `schema preflight component "pkg:index:Selected"`) {
			t.Fatalf("expected selected component context in error, got %q", err)
		}
		if !strings.Contains(err.Error(), `unresolved local type ref "#/types/pkg:index:Missing"`) {
			t.Fatalf("expected unresolved ref error, got %q", err)
		}
	})

	t.Run("unreachable malformed ref does not fail selected component", func(t *testing.T) {
		t.Parallel()

		doc := &Document{
			Resources: map[string]Resource{
				"pkg:index:Selected": {
					IsComponent: true,
					InputProperties: map[string]json.RawMessage{
						"ok": json.RawMessage(`{"type":"string"}`),
					},
				},
				"pkg:index:Other": {
					IsComponent: true,
					InputProperties: map[string]json.RawMessage{
						"bad": json.RawMessage(`{"$ref":"#/types/pkg:index:Missing"}`),
					},
				},
			},
			Types: map[string]json.RawMessage{},
		}

		if err := ValidateForTranslationComponent(doc, "pkg:index:Selected"); err != nil {
			t.Fatalf("ValidateForTranslationComponent error = %v", err)
		}
	})

	t.Run("reachable non-local ref does not fail selected component preflight", func(t *testing.T) {
		t.Parallel()

		doc := &Document{
			Resources: map[string]Resource{
				"pkg:index:Selected": {
					IsComponent: true,
					InputProperties: map[string]json.RawMessage{
						"value": json.RawMessage(`{"$ref":"/aws/v7.14.0/schema.json#/types/aws:eks%2FAccessScope:AccessScope"}`),
					},
				},
			},
		}

		if err := ValidateForTranslationComponent(doc, "pkg:index:Selected"); err != nil {
			t.Fatalf("ValidateForTranslationComponent error = %v", err)
		}
	})

	t.Run("deterministic traversal for equivalent input map ordering", func(t *testing.T) {
		t.Parallel()

		buildDoc := func(assignmentOrder []string) *Document {
			inputProperties := make(map[string]json.RawMessage, 2)
			for _, key := range assignmentOrder {
				switch key {
				case "alpha":
					inputProperties[key] = json.RawMessage(`{"$ref":"#/types/pkg:index:Missing"}`)
				case "zeta":
					inputProperties[key] = json.RawMessage(`{"type":"string"}`)
				}
			}

			return &Document{
				Resources: map[string]Resource{
					"pkg:index:Selected": {
						IsComponent:     true,
						InputProperties: inputProperties,
					},
				},
				Types: map[string]json.RawMessage{},
			}
		}

		errA := ValidateForTranslationComponent(buildDoc([]string{"zeta", "alpha"}), "pkg:index:Selected")
		errB := ValidateForTranslationComponent(buildDoc([]string{"alpha", "zeta"}), "pkg:index:Selected")
		if errA == nil || errB == nil {
			t.Fatalf("expected both validations to fail, got errA=%v errB=%v", errA, errB)
		}
		if errA.Error() != errB.Error() {
			t.Fatalf("expected deterministic error ordering, got errA=%q errB=%q", errA, errB)
		}
	})
}
