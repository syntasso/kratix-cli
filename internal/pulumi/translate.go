package pulumi

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
)

const (
	localTypeRefPrefix     = "#/types/"
	localResourceRefPrefix = "#/resources/"
)

type translationContext struct {
	doc            *SchemaDocument
	componentToken string
	resolvingRefs  map[string]bool
	skipped        []skippedPathIssue
}

type unsupportedError struct {
	component string
	path      string
	summary   string
	skippable bool
}

func (e *unsupportedError) Error() string {
	return fmt.Sprintf("component %q path %q unsupported construct: %s", e.component, e.path, e.summary)
}

type skippedPathIssue struct {
	component string
	path      string
	reason    string
}

// TranslateInputsToSpecSchema converts a selected Pulumi component input schema into
// Kubernetes OpenAPI schema for CRD spec.
func TranslateInputsToSpecSchema(doc SchemaDocument, component SelectedComponent) (map[string]any, []string, error) {
	ctx := &translationContext{
		doc:            &doc,
		componentToken: component.Token,
		resolvingRefs:  make(map[string]bool),
	}

	translatedProps := make(map[string]any, len(component.Resource.InputProperties))
	for _, name := range sortedRawKeys(component.Resource.InputProperties) {
		node, err := decodeNode(component.Resource.InputProperties[name])
		if err != nil {
			return nil, nil, fmt.Errorf("translate component inputs: decode input property %q: %w", name, err)
		}

		translated, err := translateNode(ctx, node, "spec."+name)
		if err != nil {
			if skipped := maybeRecordSkippedPath(ctx, err); skipped {
				continue
			}
			return nil, toWarningMessages(ctx.skipped), err
		}
		translatedProps[name] = translated
	}

	if len(translatedProps) == 0 {
		return nil, toWarningMessages(ctx.skipped), fmt.Errorf(
			"translate component inputs: no translatable spec fields remain after skipping unsupported schema paths for component %q",
			component.Token,
		)
	}

	required := normalizedRequired(component.Resource.RequiredInputs)
	required = filterRequiredForProperties(required, translatedProps)

	specSchema := map[string]any{
		"type":       "object",
		"properties": translatedProps,
	}
	if len(required) > 0 {
		specSchema["required"] = required
	}

	return specSchema, toWarningMessages(ctx.skipped), nil
}

func translateNode(ctx *translationContext, node map[string]any, path string) (map[string]any, error) {
	if err := rejectUnsupportedKeywords(node, ctx.componentToken, path); err != nil {
		return nil, err
	}

	if ref, ok := stringField(node, "$ref"); ok {
		if !isLocalRef(ref) {
			return applyAnnotations(node, fallbackSchemaForNonLocalRef(), ctx.componentToken, path)
		}

		resolvedNode, refKey, err := resolveLocalRef(ctx.doc, ref)
		if err != nil {
			return nil, fmt.Errorf("component %q path %q invalid schema: %w", ctx.componentToken, path, err)
		}

		if ctx.resolvingRefs[refKey] {
			return nil, unsupportedHard(ctx.componentToken, path, fmt.Sprintf("cyclic local ref %q", ref))
		}
		ctx.resolvingRefs[refKey] = true
		translated, err := translateNode(ctx, resolvedNode, path)
		delete(ctx.resolvingRefs, refKey)
		if err != nil {
			return nil, err
		}

		withAnnotations, err := applyAnnotations(node, translated, ctx.componentToken, path)
		if err != nil {
			return nil, err
		}
		return withAnnotations, nil
	}

	typeName, ok := stringField(node, "type")
	if !ok {
		if _, exists := node["enum"]; exists {
			return nil, unsupported(ctx.componentToken, path, "enum without explicit type")
		}
		return nil, unsupported(ctx.componentToken, path, "missing supported shape (expected one of $ref or type)")
	}

	var translated map[string]any
	switch typeName {
	case "string", "boolean", "integer", "number":
		translated = map[string]any{"type": typeName}
	case "array":
		itemsNode, ok, err := objectField(node, "items")
		if err != nil {
			return nil, unsupported(ctx.componentToken, path, "array items must be an object schema")
		}
		if !ok {
			return nil, unsupported(ctx.componentToken, path, "array type missing items schema")
		}
		translatedItems, err := translateNode(ctx, itemsNode, path+"[]")
		if err != nil {
			return nil, err
		}
		translated = map[string]any{
			"type":  "array",
			"items": translatedItems,
		}
	case "object":
		objSchema, err := translateObjectNode(ctx, node, path)
		if err != nil {
			return nil, err
		}
		translated = objSchema
	default:
		return nil, unsupported(ctx.componentToken, path, fmt.Sprintf("type %q", typeName))
	}

	return applyAnnotations(node, translated, ctx.componentToken, path)
}

func translateObjectNode(ctx *translationContext, node map[string]any, path string) (map[string]any, error) {
	translated := map[string]any{"type": "object"}

	if propertiesNode, ok := node["properties"]; ok {
		properties, ok := propertiesNode.(map[string]any)
		if !ok {
			return nil, unsupported(ctx.componentToken, path, "object properties must be an object")
		}

		translatedProps := make(map[string]any, len(properties))
		for _, name := range sortedAnyKeys(properties) {
			child, ok := properties[name].(map[string]any)
			if !ok {
				return nil, unsupported(ctx.componentToken, path+"."+name, "property schema must be an object")
			}
			childTranslated, err := translateNode(ctx, child, path+"."+name)
			if err != nil {
				if skipped := maybeRecordSkippedPath(ctx, err); skipped {
					continue
				}
				return nil, err
			}
			translatedProps[name] = childTranslated
		}
		translated["properties"] = translatedProps
	}

	if requiredNode, ok := node["required"]; ok {
		required, err := parseRequired(requiredNode)
		if err != nil {
			return nil, unsupported(ctx.componentToken, path, "object required must be an array of strings")
		}
		if props, ok := translated["properties"].(map[string]any); ok {
			required = filterRequiredForProperties(required, props)
		}
		if len(required) > 0 {
			translated["required"] = required
		}
	}

	if additionalPropertiesNode, ok := node["additionalProperties"]; ok {
		child, ok := additionalPropertiesNode.(map[string]any)
		if !ok {
			return nil, unsupported(ctx.componentToken, path, "additionalProperties must be an object schema")
		}
		translatedChild, err := translateNode(ctx, child, path+".*")
		if err != nil {
			return nil, err
		}
		translated["additionalProperties"] = translatedChild
	}

	return translated, nil
}

func applyAnnotations(node map[string]any, translated map[string]any, componentToken, path string) (map[string]any, error) {
	if descriptionValue, ok := node["description"]; ok {
		description, ok := descriptionValue.(string)
		if !ok {
			return nil, unsupported(componentToken, path, "description must be a string")
		}
		translated["description"] = description
	}

	if defaultValue, ok := node["default"]; ok {
		translated["default"] = defaultValue
	}

	enumValues, hasEnum, err := parseEnum(node)
	if err != nil {
		return nil, unsupported(componentToken, path, err.Error())
	}
	if hasEnum {
		translated["enum"] = enumValues
	}

	return translated, nil
}

func parseEnum(node map[string]any) ([]any, bool, error) {
	enumNode, ok := node["enum"]
	if !ok {
		return nil, false, nil
	}

	values, ok := enumNode.([]any)
	if !ok {
		return nil, false, fmt.Errorf("enum must be an array")
	}

	translatedValues := make([]any, 0, len(values))
	for _, item := range values {
		switch typed := item.(type) {
		case map[string]any:
			value, exists := typed["value"]
			if !exists {
				return nil, false, fmt.Errorf("enum entry object missing value field")
			}
			translatedValues = append(translatedValues, value)
		case string, bool, float64, nil:
			translatedValues = append(translatedValues, typed)
		default:
			return nil, false, fmt.Errorf("enum entry has unsupported shape")
		}
	}

	return translatedValues, true, nil
}

func rejectUnsupportedKeywords(node map[string]any, componentToken, path string) error {
	unsupportedKeywords := []string{"oneOf", "anyOf", "allOf", "not", "discriminator", "patternProperties", "const"}
	for _, key := range unsupportedKeywords {
		if _, exists := node[key]; exists {
			return unsupported(componentToken, path, fmt.Sprintf("keyword %q", key))
		}
	}
	return nil
}

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

func parseRequired(node any) ([]string, error) {
	rawValues, ok := node.([]any)
	if !ok {
		return nil, fmt.Errorf("required is not an array")
	}

	result := make([]string, 0, len(rawValues))
	for _, value := range rawValues {
		asString, ok := value.(string)
		if !ok {
			return nil, fmt.Errorf("required entry is not string")
		}
		result = append(result, asString)
	}

	return normalizedRequired(result), nil
}

func normalizedRequired(values []string) []string {
	if len(values) == 0 {
		return nil
	}

	set := make(map[string]struct{}, len(values))
	for _, value := range values {
		set[value] = struct{}{}
	}

	result := make([]string, 0, len(set))
	for value := range set {
		result = append(result, value)
	}
	sort.Strings(result)

	return result
}

func filterRequiredForProperties(required []string, translatedProps map[string]any) []string {
	if len(required) == 0 {
		return nil
	}

	filtered := make([]string, 0, len(required))
	for _, name := range required {
		if _, exists := translatedProps[name]; exists {
			filtered = append(filtered, name)
		}
	}
	return filtered
}

func maybeRecordSkippedPath(ctx *translationContext, err error) bool {
	var unsupportedErr *unsupportedError
	if !errors.As(err, &unsupportedErr) || !unsupportedErr.skippable {
		return false
	}

	ctx.skipped = append(ctx.skipped, skippedPathIssue{
		component: unsupportedErr.component,
		path:      unsupportedErr.path,
		reason:    unsupportedErr.summary,
	})
	return true
}

func sortedSkippedIssues(issues []skippedPathIssue) []skippedPathIssue {
	if len(issues) == 0 {
		return nil
	}

	result := make([]skippedPathIssue, len(issues))
	copy(result, issues)
	sort.Slice(result, func(i, j int) bool {
		if result[i].path != result[j].path {
			return result[i].path < result[j].path
		}
		if result[i].reason != result[j].reason {
			return result[i].reason < result[j].reason
		}
		return result[i].component < result[j].component
	})
	return result
}

func toWarningMessages(issues []skippedPathIssue) []string {
	sorted := sortedSkippedIssues(issues)
	if len(sorted) == 0 {
		return nil
	}

	warnings := make([]string, 0, len(sorted))
	for _, issue := range sorted {
		warnings = append(warnings, fmt.Sprintf(
			"warning: skipped unsupported schema path %q for component %q: %s",
			issue.path,
			issue.component,
			issue.reason,
		))
	}

	return warnings
}

func unsupported(componentToken, path, summary string) error {
	return &unsupportedError{
		component: componentToken,
		path:      path,
		summary:   summary,
		skippable: true,
	}
}

func unsupportedHard(componentToken, path, summary string) error {
	return &unsupportedError{
		component: componentToken,
		path:      path,
		summary:   summary,
		skippable: false,
	}
}

func sortedRawKeys(values map[string]json.RawMessage) []string {
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

func stringField(node map[string]any, key string) (string, bool) {
	value, ok := node[key]
	if !ok {
		return "", false
	}
	asString, ok := value.(string)
	if !ok {
		return "", false
	}
	return asString, true
}

func objectField(node map[string]any, key string) (map[string]any, bool, error) {
	value, ok := node[key]
	if !ok {
		return nil, false, nil
	}
	asMap, ok := value.(map[string]any)
	if !ok {
		return nil, true, fmt.Errorf("field %q is not object", key)
	}
	return asMap, true, nil
}
