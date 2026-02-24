package pulumi

import "fmt"

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
