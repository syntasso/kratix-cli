package emit

import (
	"fmt"
	"strings"
)

const (
	DefaultGroup    = "components.platform"
	DefaultVersion  = "v1alpha1"
	DefaultKind     = "Component"
	DefaultPlural   = "components"
	DefaultSingular = "component"
)

// Identity configures CRD wrapper identity fields.
type Identity struct {
	Group    string
	Version  string
	Kind     string
	Plural   string
	Singular string
}

func DefaultIdentity() Identity {
	return Identity{
		Group:    DefaultGroup,
		Version:  DefaultVersion,
		Kind:     DefaultKind,
		Plural:   DefaultPlural,
		Singular: DefaultSingular,
	}
}

func (i Identity) MetadataName() string {
	return i.Plural + "." + i.Group
}

func (i Identity) Validate() error {
	if err := validateDNSSubdomainLike(i.Group); err != nil {
		return fmt.Errorf("invalid --group: %w", err)
	}
	if err := validateDNSLabelLike(i.Version); err != nil {
		return fmt.Errorf("invalid --version: %w", err)
	}
	if err := validateKubernetesKindLike(i.Kind); err != nil {
		return fmt.Errorf("invalid --kind: %w", err)
	}
	if err := validateDNSLabelLike(i.Plural); err != nil {
		return fmt.Errorf("invalid --plural: %w", err)
	}
	if err := validateDNSLabelLike(i.Singular); err != nil {
		return fmt.Errorf("invalid --singular: %w", err)
	}
	return nil
}

func validateDNSSubdomainLike(value string) error {
	if value == "" {
		return fmt.Errorf("must be non-empty")
	}
	if len(value) > 253 {
		return fmt.Errorf("must be 253 characters or fewer")
	}
	parts := strings.Split(value, ".")
	for _, part := range parts {
		if err := validateDNSLabelLike(part); err != nil {
			return fmt.Errorf("must be a DNS subdomain-like name")
		}
	}
	return nil
}

func validateDNSLabelLike(value string) error {
	if value == "" {
		return fmt.Errorf("must be non-empty")
	}
	if len(value) > 63 {
		return fmt.Errorf("must be 63 characters or fewer")
	}
	for idx, r := range value {
		isLower := r >= 'a' && r <= 'z'
		isDigit := r >= '0' && r <= '9'
		isDash := r == '-'
		if !isLower && !isDigit && !isDash {
			return fmt.Errorf("must be a DNS label-like name")
		}
		if (idx == 0 || idx == len(value)-1) && isDash {
			return fmt.Errorf("must be a DNS label-like name")
		}
	}
	return nil
}

func validateKubernetesKindLike(value string) error {
	if value == "" {
		return fmt.Errorf("must be non-empty")
	}
	for idx, r := range value {
		isUpper := r >= 'A' && r <= 'Z'
		isLower := r >= 'a' && r <= 'z'
		isDigit := r >= '0' && r <= '9'
		if idx == 0 {
			if !isUpper && !isLower {
				return fmt.Errorf("must match ^[A-Za-z][A-Za-z0-9]*$")
			}
			continue
		}
		if !isUpper && !isLower && !isDigit {
			return fmt.Errorf("must match ^[A-Za-z][A-Za-z0-9]*$")
		}
	}
	return nil
}
