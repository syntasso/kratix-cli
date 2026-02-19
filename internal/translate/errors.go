package translate

import "fmt"

// UnsupportedError represents a schema construct that is intentionally not supported.
type UnsupportedError struct {
	Component string
	Path      string
	Summary   string
}

func (e *UnsupportedError) Error() string {
	return fmt.Sprintf("component %q path %q unsupported construct: %s", e.Component, e.Path, e.Summary)
}

func unsupported(componentToken, path, summary string) error {
	return &UnsupportedError{
		Component: componentToken,
		Path:      path,
		Summary:   summary,
	}
}
