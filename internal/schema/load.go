package schema

import (
	"encoding/json"
	"fmt"
	"os"
)

// Document is the minimal Pulumi schema representation needed for task 01b.
type Document struct {
	Resources map[string]Resource `json:"resources"`
}

// Resource is the minimal Pulumi resource shape needed for component discovery.
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
