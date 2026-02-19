package schema

import (
	"encoding/json"
	"fmt"
	"os"
)

// Document is the minimal Pulumi schema representation needed by the converter.
type Document struct {
	Resources map[string]Resource        `json:"resources"`
	Types     map[string]json.RawMessage `json:"types"`
}

// Resource is the Pulumi resource shape used for component selection and input translation.
type Resource struct {
	IsComponent     bool                       `json:"isComponent"`
	InputProperties map[string]json.RawMessage `json:"inputProperties"`
	RequiredInputs  []string                   `json:"requiredInputs"`
}

// LoadFile reads and validates that path contains parseable JSON.
func LoadFile(path string) (*Document, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read input schema: %w", err)
	}

	var doc Document
	if err := json.Unmarshal(contents, &doc); err != nil {
		return nil, fmt.Errorf("parse input schema as JSON: %w", err)
	}

	return &doc, nil
}
