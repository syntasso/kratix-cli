package emit

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/pulumi/component-to-crd/internal/schema"
)

const defaultDerivedGroupSuffix = ".components.platform"

// DeriveIdentityDefaults builds deterministic identity defaults from schema
// metadata and the selected component token.
func DeriveIdentityDefaults(doc *schema.Document, selectedToken string) Identity {
	kind := deriveKind(selectedToken)
	singular := deriveSingular(kind)
	plural := derivePlural(singular)
	group := deriveGroup(doc, selectedToken)

	return Identity{
		Group:    group,
		Version:  deriveVersion(doc),
		Kind:     kind,
		Plural:   plural,
		Singular: singular,
	}
}

func deriveKind(selectedToken string) string {
	parts := strings.Split(selectedToken, ":")
	if len(parts) < 3 {
		return DefaultKind
	}
	candidate := strings.TrimSpace(parts[2])
	if err := validateKubernetesKindLike(candidate); err != nil {
		return DefaultKind
	}
	return candidate
}

func deriveSingular(kind string) string {
	candidate := toKebabCase(kind)
	if err := validateDNSLabelLike(candidate); err != nil {
		return DefaultSingular
	}
	return candidate
}

func derivePlural(singular string) string {
	candidate := singular + "s"
	if strings.HasSuffix(singular, "s") {
		candidate = singular + "es"
	}
	if err := validateDNSLabelLike(candidate); err != nil {
		return DefaultPlural
	}
	return candidate
}

func deriveGroup(doc *schema.Document, selectedToken string) string {
	packageKey := ""
	if doc != nil {
		packageKey = sanitizePackageKey(doc.Name)
	}
	if packageKey == "" {
		parts := strings.Split(selectedToken, ":")
		if len(parts) >= 1 {
			packageKey = sanitizePackageKey(parts[0])
		}
	}
	if packageKey == "" {
		return DefaultGroup
	}
	candidate := packageKey + defaultDerivedGroupSuffix
	if err := validateDNSSubdomainLike(candidate); err != nil {
		return DefaultGroup
	}
	return candidate
}

func deriveVersion(doc *schema.Document) string {
	if doc == nil {
		return DefaultVersion
	}
	raw := strings.TrimSpace(doc.Version)
	if raw == "" {
		return DefaultVersion
	}
	if strings.HasPrefix(raw, "v") {
		raw = raw[1:]
	}
	parts := strings.Split(raw, ".")
	if len(parts) == 0 {
		return DefaultVersion
	}
	major := strings.TrimSpace(parts[0])
	if major == "" {
		return DefaultVersion
	}
	for _, r := range major {
		if r < '0' || r > '9' {
			return DefaultVersion
		}
	}
	if major == "0" {
		return DefaultVersion
	}
	candidate := fmt.Sprintf("v%s", major)
	if err := validateDNSLabelLike(candidate); err != nil {
		return DefaultVersion
	}
	return candidate
}

func sanitizePackageKey(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return ""
	}

	var b strings.Builder
	lastDash := false
	for _, r := range value {
		isValid := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
		if isValid {
			b.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash {
			b.WriteByte('-')
			lastDash = true
		}
	}

	candidate := strings.Trim(b.String(), "-")
	if candidate == "" {
		return ""
	}
	if err := validateDNSLabelLike(candidate); err != nil {
		return ""
	}
	return candidate
}

func toKebabCase(value string) string {
	if value == "" {
		return ""
	}
	var b strings.Builder
	runes := []rune(value)
	for idx, r := range runes {
		isUpper := r >= 'A' && r <= 'Z'
		isLower := r >= 'a' && r <= 'z'
		isDigit := r >= '0' && r <= '9'

		if isUpper {
			if idx > 0 {
				prev := runes[idx-1]
				nextLower := idx+1 < len(runes) && unicode.IsLower(runes[idx+1])
				if (prev >= 'a' && prev <= 'z') || (prev >= '0' && prev <= '9') || (prev >= 'A' && prev <= 'Z' && nextLower) {
					b.WriteByte('-')
				}
			}
			b.WriteRune(r + ('a' - 'A'))
			continue
		}

		if isLower || isDigit {
			b.WriteRune(r)
			continue
		}

		if b.Len() > 0 && !strings.HasSuffix(b.String(), "-") {
			b.WriteByte('-')
		}
	}
	return strings.Trim(b.String(), "-")
}
