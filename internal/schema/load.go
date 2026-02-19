package schema

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// Document is the minimal Pulumi schema representation needed by the converter.
type Document struct {
	Name      string                     `json:"name"`
	Version   string                     `json:"version"`
	Resources map[string]Resource        `json:"resources"`
	Types     map[string]json.RawMessage `json:"types"`
}

// Resource is the Pulumi resource shape used for component selection and input translation.
type Resource struct {
	IsComponent     bool                       `json:"isComponent"`
	InputProperties map[string]json.RawMessage `json:"inputProperties"`
	RequiredInputs  []string                   `json:"requiredInputs"`
}

const defaultLoadURLTimeout = 15 * time.Second

// Load reads and validates that the provided input contains parseable JSON.
// The input may be a local filesystem path or an http(s) URL.
func Load(pathOrURL string) (*Document, error) {
	contents, err := loadInput(pathOrURL)
	if err != nil {
		return nil, err
	}

	var doc Document
	if err := json.Unmarshal(contents, &doc); err != nil {
		return nil, fmt.Errorf("parse input schema as JSON: %w", err)
	}

	return &doc, nil
}

// LoadFile is kept for backward compatibility in tests and callers that still use
// local-path-only naming.
func LoadFile(path string) (*Document, error) {
	return Load(path)
}

func loadInput(pathOrURL string) ([]byte, error) {
	parsedURL, err := url.Parse(pathOrURL)
	if err == nil {
		scheme := strings.ToLower(parsedURL.Scheme)
		if scheme == "http" || scheme == "https" {
			return loadURL(pathOrURL)
		}
	}
	return loadFile(pathOrURL)
}

func sanitizeURLForError(rawURL string) string {
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

func loadURL(rawURL string) ([]byte, error) {
	client := &http.Client{Timeout: defaultLoadURLTimeout}
	return loadURLWithClient(rawURL, client)
}

func loadURLWithClient(rawURL string, client *http.Client) ([]byte, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("fetch input schema URL: parse URL: %w", err)
	}

	scheme := strings.ToLower(parsedURL.Scheme)
	if scheme != "http" && scheme != "https" {
		return nil, fmt.Errorf("fetch input schema URL: unsupported scheme %q", parsedURL.Scheme)
	}

	response, err := client.Get(rawURL)
	if err != nil {
		return nil, fmt.Errorf("fetch input schema URL: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch input schema URL: unexpected status %d for %s", response.StatusCode, sanitizeURLForError(rawURL))
	}

	contents, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("fetch input schema URL: read response body: %w", err)
	}
	return contents, nil
}

func loadFile(path string) ([]byte, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read input schema file: %w", err)
	}
	return contents, nil
}
