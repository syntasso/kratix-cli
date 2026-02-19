package schema

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

const localTypeRefPrefix = "#/types/"
const localResourceRefPrefix = "#/resources/"

// ValidateForTranslation checks schema shape constraints needed for translation traversal.
func ValidateForTranslation(doc *Document) error {
	if doc == nil {
		return fmt.Errorf("schema preflight path %q: document is nil", "root")
	}

	validator := newPreflightValidator(doc, false)

	for _, token := range sortedResourceKeys(doc.Resources) {
		resource := doc.Resources[token]
		if err := validator.validateResourceInputs(token, resource); err != nil {
			return err
		}
	}

	for _, token := range sortedRawMessageKeys(doc.Types) {
		nodePath := fmt.Sprintf("types.%s", token)
		if err := validator.validateRawNode(doc.Types[token], nodePath); err != nil {
			return err
		}
	}

	return nil
}

// ValidateForTranslationComponent checks schema shape constraints reachable from the selected component.
func ValidateForTranslationComponent(doc *Document, componentToken string) error {
	if doc == nil {
		return fmt.Errorf("schema preflight component %q path %q: document is nil", componentToken, "root")
	}
	if componentToken == "" {
		return fmt.Errorf("schema preflight component %q path %q: selected component token is empty", componentToken, "root")
	}

	resource, ok := doc.Resources[componentToken]
	if !ok {
		return fmt.Errorf("schema preflight component %q path %q: selected component not found in schema resources", componentToken, "resources."+componentToken)
	}

	validator := newPreflightValidator(doc, true)
	if err := validator.validateResourceInputs(componentToken, resource); err != nil {
		return fmt.Errorf("schema preflight component %q: %w", componentToken, err)
	}

	return nil
}

type preflightValidator struct {
	doc               *Document
	validatedRef      map[string]bool
	visitingRef       map[string]bool
	allowNonLocalRefs bool
}

func newPreflightValidator(doc *Document, allowNonLocalRefs bool) *preflightValidator {
	return &preflightValidator{
		doc:               doc,
		validatedRef:      make(map[string]bool),
		visitingRef:       make(map[string]bool),
		allowNonLocalRefs: allowNonLocalRefs,
	}
}

func (v *preflightValidator) validateResourceInputs(resourceToken string, resource Resource) error {
	for _, prop := range sortedRawMessageKeys(resource.InputProperties) {
		nodePath := fmt.Sprintf("resources.%s.inputProperties.%s", resourceToken, prop)
		if err := v.validateRawNode(resource.InputProperties[prop], nodePath); err != nil {
			return err
		}
	}
	return nil
}

func (v *preflightValidator) validateRawNode(raw json.RawMessage, path string) error {
	node, err := decodeNode(raw)
	if err != nil {
		return fmt.Errorf("schema preflight path %q: %w", path, err)
	}
	return v.validateNode(node, path)
}

func (v *preflightValidator) validateNode(node map[string]any, path string) error {
	refNode, hasRef := node["$ref"]
	if hasRef {
		ref, ok := refNode.(string)
		if !ok {
			return fmt.Errorf("schema preflight path %q: $ref must be a string", path)
		}
		if err := v.validateRef(path, ref); err != nil {
			return err
		}
	}

	if propertiesNode, ok := node["properties"]; ok {
		properties, ok := propertiesNode.(map[string]any)
		if !ok {
			return fmt.Errorf("schema preflight path %q: properties must be an object schema map", path)
		}

		for _, key := range sortedAnyKeys(properties) {
			propertyNode, ok := properties[key].(map[string]any)
			if !ok {
				return fmt.Errorf("schema preflight path %q: property schema must be an object", path+".properties."+key)
			}
			if err := v.validateNode(propertyNode, path+".properties."+key); err != nil {
				return err
			}
		}
	}

	if itemsNode, ok := node["items"]; ok {
		items, ok := itemsNode.(map[string]any)
		if !ok {
			return fmt.Errorf("schema preflight path %q: items must be an object schema", path)
		}
		if err := v.validateNode(items, path+".items"); err != nil {
			return err
		}
	}

	if additionalPropertiesNode, ok := node["additionalProperties"]; ok {
		additionalProperties, ok := additionalPropertiesNode.(map[string]any)
		if !ok {
			return fmt.Errorf("schema preflight path %q: additionalProperties must be an object schema", path)
		}
		if err := v.validateNode(additionalProperties, path+".additionalProperties"); err != nil {
			return err
		}
	}

	return nil
}

func (v *preflightValidator) validateRef(path, ref string) error {
	switch {
	case strings.HasPrefix(ref, localTypeRefPrefix):
		return v.validateTypeRef(path, ref)
	case strings.HasPrefix(ref, localResourceRefPrefix):
		return v.validateResourceRef(path, ref)
	default:
		if v.allowNonLocalRefs {
			return nil
		}
		return fmt.Errorf("schema preflight path %q: unsupported ref %q (expected local ref prefix %q or %q)", path, ref, localTypeRefPrefix, localResourceRefPrefix)
	}
}

func (v *preflightValidator) validateTypeRef(path, ref string) error {
	typeToken := strings.TrimPrefix(ref, localTypeRefPrefix)
	if typeToken == "" {
		return fmt.Errorf("schema preflight path %q: invalid local type ref %q", path, ref)
	}

	rawType, ok := v.doc.Types[typeToken]
	if !ok {
		return fmt.Errorf("schema preflight path %q: unresolved local type ref %q", path, ref)
	}

	cacheKey := localTypeRefPrefix + typeToken
	if v.validatedRef[cacheKey] || v.visitingRef[cacheKey] {
		return nil
	}

	v.visitingRef[cacheKey] = true
	defer delete(v.visitingRef, cacheKey)

	typePath := "types." + typeToken
	if err := v.validateRawNode(rawType, typePath); err != nil {
		return fmt.Errorf("schema preflight path %q: invalid ref target %q: %w", path, ref, err)
	}

	v.validatedRef[cacheKey] = true
	return nil
}

func (v *preflightValidator) validateResourceRef(path, ref string) error {
	resourceToken := strings.TrimPrefix(ref, localResourceRefPrefix)
	if resourceToken == "" {
		return fmt.Errorf("schema preflight path %q: invalid local resource ref %q", path, ref)
	}

	resource, ok := v.doc.Resources[resourceToken]
	if !ok {
		return fmt.Errorf("schema preflight path %q: unresolved local resource ref %q", path, ref)
	}

	cacheKey := localResourceRefPrefix + resourceToken
	if v.validatedRef[cacheKey] || v.visitingRef[cacheKey] {
		return nil
	}

	v.visitingRef[cacheKey] = true
	defer delete(v.visitingRef, cacheKey)

	if err := v.validateResourceInputs(resourceToken, resource); err != nil {
		return fmt.Errorf("schema preflight path %q: invalid ref target %q: %w", path, ref, err)
	}

	v.validatedRef[cacheKey] = true
	return nil
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

func sortedResourceKeys(values map[string]Resource) []string {
	if len(values) == 0 {
		return nil
	}
	result := make([]string, 0, len(values))
	for key := range values {
		result = append(result, key)
	}
	sort.Strings(result)
	return result
}

func sortedRawMessageKeys(values map[string]json.RawMessage) []string {
	if len(values) == 0 {
		return nil
	}
	result := make([]string, 0, len(values))
	for key := range values {
		result = append(result, key)
	}
	sort.Strings(result)
	return result
}

func sortedAnyKeys(values map[string]any) []string {
	if len(values) == 0 {
		return nil
	}
	result := make([]string, 0, len(values))
	for key := range values {
		result = append(result, key)
	}
	sort.Strings(result)
	return result
}
