package pulumi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadSchema(t *testing.T) {
	t.Parallel()

	t.Run("loads local file", func(t *testing.T) {
		t.Parallel()

		tempDir := t.TempDir()
		schemaPath := filepath.Join(tempDir, "schema.json")
		if err := os.WriteFile(schemaPath, []byte(`{"resources":{"pkg:index:Thing":{"isComponent":true}}}`), 0o600); err != nil {
			t.Fatalf("write schema fixture: %v", err)
		}

		doc, err := LoadSchema(schemaPath)
		if err != nil {
			t.Fatalf("LoadSchema returned error: %v", err)
		}
		if !doc.Resources["pkg:index:Thing"].IsComponent {
			t.Fatalf("expected pkg:index:Thing to be a component")
		}
	})

	t.Run("loads URL", func(t *testing.T) {
		t.Parallel()

		client := clientWithRoundTripper(roundTripFunc(func(_ *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"resources":{"pkg:index:Thing":{"isComponent":true}}}`)),
			}, nil
		}))

		contents, err := readSchemaURLWithClient("https://example.com/schema.json", client)
		if err != nil {
			t.Fatalf("readSchemaURLWithClient returned error: %v", err)
		}

		var doc SchemaDocument
		if err := json.Unmarshal(contents, &doc); err != nil {
			t.Fatalf("json.Unmarshal returned error: %v", err)
		}
		if !doc.Resources["pkg:index:Thing"].IsComponent {
			t.Fatalf("expected pkg:index:Thing to be a component")
		}
	})

	t.Run("returns malformed JSON error", func(t *testing.T) {
		t.Parallel()

		tempDir := t.TempDir()
		schemaPath := filepath.Join(tempDir, "schema.json")
		if err := os.WriteFile(schemaPath, []byte(`{"resources":`), 0o600); err != nil {
			t.Fatalf("write schema fixture: %v", err)
		}

		_, err := LoadSchema(schemaPath)
		if err == nil || !strings.Contains(err.Error(), "load schema: parse input schema as JSON:") {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("returns non-200 URL status error", func(t *testing.T) {
		t.Parallel()

		client := clientWithRoundTripper(roundTripFunc(func(_ *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusNotFound,
				Body:       io.NopCloser(strings.NewReader("missing")),
			}, nil
		}))

		_, err := readSchemaURLWithClient("https://example.com/schema.json", client)
		if err == nil || !strings.Contains(err.Error(), "load schema: fetch input schema URL: unexpected status 404 for") {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("returns unreachable URL error", func(t *testing.T) {
		t.Parallel()

		client := clientWithRoundTripper(roundTripFunc(func(_ *http.Request) (*http.Response, error) {
			return nil, errors.New("dial tcp: lookup example.com: no such host")
		}))

		_, err := readSchemaURLWithClient("https://example.com/schema.json", client)
		if err == nil || !strings.Contains(err.Error(), "load schema: fetch input schema URL:") {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("returns unsupported URL scheme error", func(t *testing.T) {
		t.Parallel()

		_, err := LoadSchema("ftp://example.com/schema.json")
		want := `load schema: unsupported URL scheme "ftp"`
		if err == nil || err.Error() != want {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

type roundTripFunc func(req *http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func clientWithRoundTripper(rt http.RoundTripper) *http.Client {
	return &http.Client{Transport: rt}
}
