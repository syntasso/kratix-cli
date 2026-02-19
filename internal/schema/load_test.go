package schema

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFile(t *testing.T) {
	t.Parallel()

	tempDir := makeWorkspaceTempDir(t)
	path := filepath.Join(tempDir, "schema.json")
	if err := os.WriteFile(path, []byte(`{"resources":{"pkg:index:Thing":{"isComponent":true}}}`), 0o600); err != nil {
		t.Fatalf("write schema fixture: %v", err)
	}

	doc, err := LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile returned error: %v", err)
	}

	if !doc.Resources["pkg:index:Thing"].IsComponent {
		t.Fatalf("expected pkg:index:Thing to be marked as component")
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
