package pulumi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

const schemaURLTimeout = 15 * time.Second

// SchemaDocument is the subset of the Pulumi package schema required by init.
type SchemaDocument struct {
	Name      string                     `json:"name"`
	Version   string                     `json:"version"`
	Resources map[string]SchemaResource  `json:"resources"`
	Types     map[string]json.RawMessage `json:"types"`
}

// SchemaResource contains the component metadata and input shape for a Pulumi resource.
type SchemaResource struct {
	IsComponent     bool                       `json:"isComponent"`
	InputProperties map[string]json.RawMessage `json:"inputProperties"`
	RequiredInputs  []string                   `json:"requiredInputs"`
}

// LoadSchema loads a Pulumi package schema from a local file path or HTTP(S) URL.
func LoadSchema(source string) (SchemaDocument, error) {
	rawSchema, err := readSchemaSource(source)
	if err != nil {
		return SchemaDocument{}, err
	}

	var doc SchemaDocument
	if err := json.Unmarshal(rawSchema, &doc); err != nil {
		return SchemaDocument{}, fmt.Errorf("load schema: parse input schema as JSON: %w", err)
	}

	return doc, nil
}

func readSchemaSource(source string) ([]byte, error) {
	parsedURL, err := url.Parse(source)
	if err == nil {
		scheme := strings.ToLower(parsedURL.Scheme)
		if scheme == "" || isWindowsFilePath(source) {
			return readSchemaFile(source)
		}

		switch scheme {
		case "http", "https":
			return readSchemaURL(source)
		default:
			return nil, fmt.Errorf("load schema: unsupported URL scheme %q", parsedURL.Scheme)
		}
	}

	return readSchemaFile(source)
}

// IsLocalSchemaSource reports whether a schema source should be treated as a local file path.
func IsLocalSchemaSource(source string) bool {
	parsedURL, err := url.Parse(source)
	if err != nil {
		return true
	}

	scheme := strings.ToLower(parsedURL.Scheme)
	return scheme == "" || isWindowsFilePath(source)
}

var windowsFilePath = regexp.MustCompile(`^[a-zA-Z]:[\\/]`)

func isWindowsFilePath(source string) bool {
	return windowsFilePath.MatchString(source)
}

func readSchemaFile(path string) ([]byte, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("load schema: read input schema file: %w", err)
	}
	return contents, nil
}

func readSchemaURL(rawURL string) ([]byte, error) {
	client := &http.Client{Timeout: schemaURLTimeout}
	return readSchemaURLWithClient(rawURL, client)
}

func readSchemaURLWithClient(rawURL string, client *http.Client) ([]byte, error) {
	resp, err := client.Get(rawURL)
	if err != nil {
		return nil, fmt.Errorf("load schema: fetch input schema URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("load schema: fetch input schema URL: unexpected status %d for %s", resp.StatusCode, sanitizedURL(rawURL))
	}

	contents, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("load schema: fetch input schema URL: read response body: %w", err)
	}

	return contents, nil
}

func sanitizedURL(rawURL string) string {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return rawURL
	}
	if parsedURL.Path == "" {
		return parsedURL.Scheme + "://" + parsedURL.Host
	}
	return parsedURL.Scheme + "://" + parsedURL.Host + parsedURL.Path
}
