package selectcomponent

import (
	"fmt"
	"slices"
	"strings"

	"github.com/syntasso/pulumi-component-to-crd/internal/schema"
)

// DiscoverComponentTokens returns sorted component resource tokens from schema.
func DiscoverComponentTokens(doc *schema.Document) []string {
	if doc == nil || len(doc.Resources) == 0 {
		return nil
	}

	tokens := make([]string, 0, len(doc.Resources))
	for token, resource := range doc.Resources {
		if resource.IsComponent {
			tokens = append(tokens, token)
		}
	}

	slices.Sort(tokens)
	return tokens
}

// SelectComponent applies deterministic component selection rules.
func SelectComponent(tokens []string, requested string) (string, error) {
	if len(tokens) == 0 {
		return "", fmt.Errorf("no component resources found in schema")
	}

	if requested != "" {
		if slices.Contains(tokens, requested) {
			return requested, nil
		}

		return "", fmt.Errorf("component %q not found; available components: %s", requested, strings.Join(tokens, ", "))
	}

	if len(tokens) == 1 {
		return tokens[0], nil
	}

	return "", fmt.Errorf("multiple components found; provide --component from: %s", strings.Join(tokens, ", "))
}
