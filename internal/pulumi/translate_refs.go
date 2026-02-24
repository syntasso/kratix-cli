package pulumi

import (
	"encoding/json"
	"fmt"
	"strings"
)

func resolveLocalRef(doc *SchemaDocument, ref string) (map[string]any, string, error) {
	switch {
	case strings.HasPrefix(ref, localTypeRefPrefix):
		typeToken := strings.TrimPrefix(ref, localTypeRefPrefix)
		if typeToken == "" {
			return nil, "", fmt.Errorf("invalid local type ref %q", ref)
		}
		rawType, ok := doc.Types[typeToken]
		if !ok {
			return nil, "", fmt.Errorf("unresolved local type ref %q", ref)
		}

		typeNode, err := decodeNode(rawType)
		if err != nil {
			return nil, "", fmt.Errorf("decode local type ref %q: %w", ref, err)
		}
		return typeNode, localTypeRefPrefix + typeToken, nil
	case strings.HasPrefix(ref, localResourceRefPrefix):
		resourceToken := strings.TrimPrefix(ref, localResourceRefPrefix)
		if resourceToken == "" {
			return nil, "", fmt.Errorf("invalid local resource ref %q", ref)
		}
		resource, ok := doc.Resources[resourceToken]
		if !ok {
			return nil, "", fmt.Errorf("unresolved local resource ref %q", ref)
		}
		resourceNode, err := buildResourceRefNode(resource)
		if err != nil {
			return nil, "", fmt.Errorf("decode local resource ref %q: %w", ref, err)
		}
		return resourceNode, localResourceRefPrefix + resourceToken, nil
	default:
		return nil, "", fmt.Errorf("unsupported ref %q: only local type and resource refs are supported", ref)
	}
}

func buildResourceRefNode(resource SchemaResource) (map[string]any, error) {
	properties := make(map[string]any, len(resource.InputProperties))
	for _, name := range sortedRawKeys(resource.InputProperties) {
		node, err := decodeNode(resource.InputProperties[name])
		if err != nil {
			return nil, fmt.Errorf("input property %q: %w", name, err)
		}
		properties[name] = node
	}

	requiredValues := normalizedRequired(resource.RequiredInputs)
	required := make([]any, 0, len(requiredValues))
	for _, value := range requiredValues {
		required = append(required, value)
	}

	result := map[string]any{
		"type":       "object",
		"properties": properties,
	}
	if len(required) > 0 {
		result["required"] = required
	}

	return result, nil
}

func decodeNode(raw json.RawMessage) (map[string]any, error) {
	if len(raw) == 0 {
		return nil, fmt.Errorf("empty schema node")
	}

	var node map[string]any
	if err := json.Unmarshal(raw, &node); err != nil {
		return nil, err
	}
	if node == nil {
		return nil, fmt.Errorf("schema node is null")
	}

	return node, nil
}

func isLocalRef(ref string) bool {
	return strings.HasPrefix(ref, localTypeRefPrefix) || strings.HasPrefix(ref, localResourceRefPrefix)
}

func fallbackSchemaForNonLocalRef() map[string]any {
	return map[string]any{
		"type":                                 "object",
		"x-kubernetes-preserve-unknown-fields": true,
	}
}
