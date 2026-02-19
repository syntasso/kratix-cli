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

func TestResolveFileReadCandidates(t *testing.T) {
	t.Parallel()

	t.Run("absolute input path", func(t *testing.T) {
		t.Parallel()

		got := resolveFileReadCandidates("/tmp/schema.json", "/Users/example/project")
		if len(got) != 1 || got[0] != "/tmp/schema.json" {
			t.Fatalf("unexpected candidates: %#v", got)
		}
	})

	t.Run("relative input path with absolute PWD", func(t *testing.T) {
		t.Parallel()

		got := resolveFileReadCandidates("schema.json", "/Users/example/project")
		if len(got) != 2 {
			t.Fatalf("expected 2 candidates, got %#v", got)
		}
		if got[0] != "/Users/example/project/schema.json" {
			t.Fatalf("unexpected first candidate: %q", got[0])
		}
		if got[1] != "schema.json" {
			t.Fatalf("unexpected second candidate: %q", got[1])
		}
	})

	t.Run("relative input path with non-absolute PWD", func(t *testing.T) {
		t.Parallel()

		got := resolveFileReadCandidates("schema.json", "project")
		if len(got) != 1 || got[0] != "schema.json" {
			t.Fatalf("unexpected candidates: %#v", got)
		}
	})
}

func TestLoadFile_UsesPWDForRelativePath(t *testing.T) {
	tempDir := makeWorkspaceTempDir(t)
	absTempDir, err := filepath.Abs(tempDir)
	if err != nil {
		t.Fatalf("resolve temp dir absolute path: %v", err)
	}
	path := filepath.Join(absTempDir, "schema.json")
	if err := os.WriteFile(path, []byte(`{"resources":{}}`), 0o600); err != nil {
		t.Fatalf("write schema fixture: %v", err)
	}

	t.Setenv("PWD", absTempDir)

	contents, err := loadFile("schema.json")
	if err != nil {
		t.Fatalf("loadFile returned error: %v", err)
	}
	if !strings.Contains(string(contents), `"resources"`) {
		t.Fatalf("unexpected schema content: %q", string(contents))
	}
}

func TestLoadFile_FallsBackToProcessCWD(t *testing.T) {
	cwdDir := makeWorkspaceTempDir(t)
	absCWDDir, err := filepath.Abs(cwdDir)
	if err != nil {
		t.Fatalf("resolve cwd dir absolute path: %v", err)
	}
	pwdDir := makeWorkspaceTempDir(t)
	absPWDDir, err := filepath.Abs(pwdDir)
	if err != nil {
		t.Fatalf("resolve pwd dir absolute path: %v", err)
	}

	path := filepath.Join(absCWDDir, "schema.json")
	if err := os.WriteFile(path, []byte(`{"resources":{"from":"cwd"}}`), 0o600); err != nil {
		t.Fatalf("write schema fixture: %v", err)
	}

	withCurrentDir(t, absCWDDir, func() {
		t.Setenv("PWD", absPWDDir)

		contents, err := loadFile("schema.json")
		if err != nil {
			t.Fatalf("loadFile returned error: %v", err)
		}
		if !strings.Contains(string(contents), `"cwd"`) {
			t.Fatalf("expected fallback to process cwd file, got %q", string(contents))
		}
	})
}

func TestLoadFile_PrefersPWDOverProcessCWD(t *testing.T) {
	cwdDir := makeWorkspaceTempDir(t)
	absCWDDir, err := filepath.Abs(cwdDir)
	if err != nil {
		t.Fatalf("resolve cwd dir absolute path: %v", err)
	}
	pwdDir := makeWorkspaceTempDir(t)
	absPWDDir, err := filepath.Abs(pwdDir)
	if err != nil {
		t.Fatalf("resolve pwd dir absolute path: %v", err)
	}

	if err := os.WriteFile(filepath.Join(absCWDDir, "schema.json"), []byte(`{"resources":{"from":"cwd"}}`), 0o600); err != nil {
		t.Fatalf("write cwd schema fixture: %v", err)
	}
	if err := os.WriteFile(filepath.Join(absPWDDir, "schema.json"), []byte(`{"resources":{"from":"pwd"}}`), 0o600); err != nil {
		t.Fatalf("write pwd schema fixture: %v", err)
	}

	withCurrentDir(t, absCWDDir, func() {
		t.Setenv("PWD", absPWDDir)

		contents, err := loadFile("schema.json")
		if err != nil {
			t.Fatalf("loadFile returned error: %v", err)
		}
		if !strings.Contains(string(contents), `"pwd"`) {
			t.Fatalf("expected PWD file to be preferred, got %q", string(contents))
		}
	})
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

func withCurrentDir(t *testing.T, dir string, fn func()) {
	t.Helper()
	previousDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get current directory: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("change directory to %q: %v", dir, err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(previousDir); err != nil {
			t.Fatalf("restore current directory to %q: %v", previousDir, err)
		}
	})
	fn()
}
