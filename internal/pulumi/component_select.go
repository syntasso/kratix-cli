package pulumi

import (
	"fmt"
	"slices"
	"strings"
)

// SelectedComponent stores the selected token and schema resource for downstream processing.
type SelectedComponent struct {
	Token    string
	Resource SchemaResource
}

// SelectComponent applies deterministic component selection rules over schema resources.
func SelectComponent(doc SchemaDocument, requestedToken string) (SelectedComponent, error) {
	componentTokens := discoverComponentTokens(doc)
	if len(componentTokens) == 0 {
		return SelectedComponent{}, fmt.Errorf("select component: no component resources found in schema")
	}

	if requestedToken != "" {
		if resource, exists := doc.Resources[requestedToken]; exists && resource.IsComponent {
			return SelectedComponent{
				Token:    requestedToken,
				Resource: resource,
			}, nil
		}

		return SelectedComponent{}, fmt.Errorf(
			"select component: component %q not found; available components: %s",
			requestedToken,
			strings.Join(componentTokens, ", "),
		)
	}

	if len(componentTokens) == 1 {
		selectedToken := componentTokens[0]
		return SelectedComponent{
			Token:    selectedToken,
			Resource: doc.Resources[selectedToken],
		}, nil
	}

	return SelectedComponent{}, fmt.Errorf(
		"select component: multiple components found; provide --component from: %s",
		strings.Join(componentTokens, ", "),
	)
}

func discoverComponentTokens(doc SchemaDocument) []string {
	if len(doc.Resources) == 0 {
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
