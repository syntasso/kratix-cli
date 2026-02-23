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

func TestIsLocalSchemaSource(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		source  string
		expects bool
	}{
		"relative path": {
			source:  "./schema.json",
			expects: true,
		},
		"absolute path": {
			source:  "/tmp/schema.json",
			expects: true,
		},
		"windows path": {
			source:  `C:\tmp\schema.json`,
			expects: true,
		},
		"https url": {
			source:  "https://example.com/schema.json",
			expects: false,
		},
		"http url": {
			source:  "http://example.com/schema.json",
			expects: false,
		},
		"unsupported url scheme": {
			source:  "ftp://example.com/schema.json",
			expects: false,
		},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got := IsLocalSchemaSource(tc.source); got != tc.expects {
				t.Fatalf("IsLocalSchemaSource(%q) = %t, want %t", tc.source, got, tc.expects)
			}
		})
	}
}

func TestReadSchemaURLWithClient_TestEnvOverride(t *testing.T) {
	t.Run("returns fixture body without HTTP request", func(t *testing.T) {
		rawURL := "https://schemas.example.test/schema.json"
		t.Setenv("KRATIX_TEST_SCHEMA_URL", rawURL)
		t.Setenv("KRATIX_TEST_SCHEMA_URL_BODY", `{"resources":{"pkg:index:Thing":{"isComponent":true}}}`)

		contents, err := readSchemaURLWithClient(rawURL, clientWithRoundTripper(roundTripFunc(func(_ *http.Request) (*http.Response, error) {
			t.Fatal("unexpected HTTP request when KRATIX_TEST_SCHEMA_URL is set")
			return nil, nil
		})))
		if err != nil {
			t.Fatalf("readSchemaURLWithClient returned error: %v", err)
		}
		if !strings.Contains(string(contents), `"pkg:index:Thing"`) {
			t.Fatalf("unexpected schema contents: %s", string(contents))
		}
	})

	t.Run("returns configured non-200 error", func(t *testing.T) {
		rawURL := "https://schemas.example.test/missing.json"
		t.Setenv("KRATIX_TEST_SCHEMA_URL", rawURL)
		t.Setenv("KRATIX_TEST_SCHEMA_URL_MODE", "status:404")

		_, err := readSchemaURLWithClient(rawURL, clientWithRoundTripper(roundTripFunc(func(_ *http.Request) (*http.Response, error) {
			t.Fatal("unexpected HTTP request when KRATIX_TEST_SCHEMA_URL is set")
			return nil, nil
		})))
		if err == nil || !strings.Contains(err.Error(), "load schema: fetch input schema URL: unexpected status 404 for") {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("returns configured transport error", func(t *testing.T) {
		rawURL := "https://schemas.example.test/unreachable.json"
		t.Setenv("KRATIX_TEST_SCHEMA_URL", rawURL)
		t.Setenv("KRATIX_TEST_SCHEMA_URL_MODE", "error:dial tcp: lookup schemas.example.test: no such host")

		_, err := readSchemaURLWithClient(rawURL, clientWithRoundTripper(roundTripFunc(func(_ *http.Request) (*http.Response, error) {
			t.Fatal("unexpected HTTP request when KRATIX_TEST_SCHEMA_URL is set")
			return nil, nil
		})))
		if err == nil || !strings.Contains(err.Error(), "load schema: fetch input schema URL: dial tcp: lookup schemas.example.test: no such host") {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("ignores override when URL does not match", func(t *testing.T) {
		t.Setenv("KRATIX_TEST_SCHEMA_URL", "https://schemas.example.test/schema.json")
		t.Setenv("KRATIX_TEST_SCHEMA_URL_BODY", `{"resources":{}}`)

		_, err := readSchemaURLWithClient("https://example.com/schema.json", clientWithRoundTripper(roundTripFunc(func(_ *http.Request) (*http.Response, error) {
			return nil, errors.New("network call reached client")
		})))
		if err == nil || !strings.Contains(err.Error(), "network call reached client") {
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
