package schema

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	t.Parallel()

	t.Run("local path success", func(t *testing.T) {
		t.Parallel()

		tempDir := makeWorkspaceTempDir(t)
		path := filepath.Join(tempDir, "schema.json")
		if err := os.WriteFile(path, []byte(`{"resources":{"pkg:index:Thing":{"isComponent":true}}}`), 0o600); err != nil {
			t.Fatalf("write schema fixture: %v", err)
		}

		doc, err := Load(path)
		if err != nil {
			t.Fatalf("Load returned error: %v", err)
		}

		if !doc.Resources["pkg:index:Thing"].IsComponent {
			t.Fatalf("expected pkg:index:Thing to be marked as component")
		}
	})

	t.Run("url success", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"resources":{"pkg:index:Thing":{"isComponent":true}}}`))
		}))
		t.Cleanup(server.Close)

		doc, err := Load(server.URL)
		if err != nil {
			t.Fatalf("Load returned error: %v", err)
		}
		if !doc.Resources["pkg:index:Thing"].IsComponent {
			t.Fatalf("expected pkg:index:Thing to be marked as component")
		}
	})

	t.Run("url non-200 status", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "missing", http.StatusNotFound)
		}))
		t.Cleanup(server.Close)

		_, err := Load(server.URL)
		if err == nil || !strings.Contains(err.Error(), "fetch input schema URL: unexpected status 404 for ") {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("url invalid json", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"resources":`))
		}))
		t.Cleanup(server.Close)

		_, err := Load(server.URL)
		if err == nil || !strings.Contains(err.Error(), "parse input schema as JSON:") {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("unsupported url scheme falls back to local file handling", func(t *testing.T) {
		t.Parallel()

		_, err := Load("ftp://example.com/schema.json")
		if err == nil || !strings.Contains(err.Error(), "read input schema file:") {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestLoadURLWithClient_Timeout(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(50 * time.Millisecond)
		_, _ = w.Write([]byte(`{"resources":{}}`))
	}))
	t.Cleanup(server.Close)

	client := &http.Client{Timeout: 5 * time.Millisecond}
	_, err := loadURLWithClient(server.URL, client)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if !strings.Contains(err.Error(), "fetch input schema URL:") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func makeWorkspaceTempDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp(".", ".test-tmp-")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatalf("remove temp dir %q: %v", dir, err)
		}
	})
	return dir
}
