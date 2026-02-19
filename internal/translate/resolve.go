package translate

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pulumi/component-to-crd/internal/schema"
)

const localTypeRefPrefix = "#/types/"

func resolveLocalType(doc *schema.Document, ref string) (map[string]any, string, error) {
	if !strings.HasPrefix(ref, localTypeRefPrefix) {
		return nil, "", fmt.Errorf("unsupported ref %q: only local type refs are supported", ref)
	}
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
	return typeNode, typeToken, nil
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
