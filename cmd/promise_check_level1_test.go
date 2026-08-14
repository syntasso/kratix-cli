package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const validPromiseYAML = `
apiVersion: platform.kratix.io/v1alpha1
kind: Promise
metadata:
  name: inventory-sync-service
spec:
  api:
    apiVersion: apiextensions.k8s.io/v1
    kind: CustomResourceDefinition
    spec:
      group: shopco.platform.syntasso.io
      scope: Namespaced
      names:
        kind: InventorySyncService
      versions:
        - name: v1alpha1
          schema:
            openAPIV3Schema:
              type: object
              properties:
                spec:
                  type: object
                  required: ["environment"]
                  properties:
                    environment:
                      type: string
                      pattern: "^(prod|staging)$"
  workflows:
    resource:
      configure:
        - spec:
            containers:
              - name: configure
                image: busybox
      delete:
        - spec:
            containers:
              - name: delete
                image: busybox
`

func writeTemp(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	return path
}

func TestCheckPromiseFilePassesOnWellFormedPromise(t *testing.T) {
	errs, doc, err := checkPromiseFile("inventory-sync-service.yaml", []byte(validPromiseYAML))
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if len(errs) != 0 {
		t.Fatalf("expected no errors, got %v", errs)
	}
	if doc.Metadata.Name != "inventory-sync-service" {
		t.Fatalf("expected doc to be parsed, got name %q", doc.Metadata.Name)
	}
}

func TestCheckPromiseFileFlagsClusterScope(t *testing.T) {
	bad := strings.Replace(validPromiseYAML, "scope: Namespaced", "scope: Cluster", 1)
	errs, _, err := checkPromiseFile("bad-scope.yaml", []byte(bad))
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if !containsSubstring(errs, "must be Namespaced") {
		t.Fatalf("expected a Namespaced-scope error, got %v", errs)
	}
}

func TestCheckPromiseFileFlagsMissingConfigureImage(t *testing.T) {
	bad := strings.Replace(validPromiseYAML, "              - name: configure\n                image: busybox", "              - name: configure\n                image: \"\"", 1)
	errs, _, err := checkPromiseFile("bad-image.yaml", []byte(bad))
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if !containsSubstring(errs, "no image set") {
		t.Fatalf("expected a missing-image error, got %v", errs)
	}
}

func TestCheckPromiseFileFlagsMultipleDeletePipelines(t *testing.T) {
	bad := strings.Replace(validPromiseYAML, `      delete:
        - spec:
            containers:
              - name: delete
                image: busybox`, `      delete:
        - spec:
            containers:
              - name: delete-one
                image: busybox
        - spec:
            containers:
              - name: delete-two
                image: busybox`, 1)
	errs, _, err := checkPromiseFile("bad-delete.yaml", []byte(bad))
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if !containsSubstring(errs, "only supports one") {
		t.Fatalf("expected a multiple-delete-pipeline error, got %v", errs)
	}
}

func TestCheckPromiseFileFlagsNotAPromise(t *testing.T) {
	bad := strings.Replace(validPromiseYAML, "kind: Promise", "kind: NotAPromise", 1)
	errs, _, err := checkPromiseFile("not-a-promise.yaml", []byte(bad))
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if !containsSubstring(errs, "not a Promise") {
		t.Fatalf("expected a not-a-Promise error, got %v", errs)
	}
}

func TestCheckExampleFileValidatesRequiredAndPattern(t *testing.T) {
	_, doc, err := checkPromiseFile("inventory-sync-service.yaml", []byte(validPromiseYAML))
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	promises := map[string]promiseDoc{"inventory-sync-service.yaml": doc}

	goodExample := `
apiVersion: shopco.platform.syntasso.io/v1alpha1
kind: InventorySyncService
spec:
  environment: prod
`
	errs := checkExampleFile("good.yaml", []byte(goodExample), promises)
	if len(errs) != 0 {
		t.Fatalf("expected no errors for a valid example, got %v", errs)
	}

	missingRequired := `
apiVersion: shopco.platform.syntasso.io/v1alpha1
kind: InventorySyncService
spec: {}
`
	errs = checkExampleFile("missing-required.yaml", []byte(missingRequired), promises)
	if !containsSubstring(errs, "missing required field") {
		t.Fatalf("expected a missing-required-field error, got %v", errs)
	}

	unknownField := `
apiVersion: shopco.platform.syntasso.io/v1alpha1
kind: InventorySyncService
spec:
  environment: prod
  bogusField: true
`
	errs = checkExampleFile("unknown-field.yaml", []byte(unknownField), promises)
	if !containsSubstring(errs, "unknown field") {
		t.Fatalf("expected an unknown-field error, got %v", errs)
	}

	badPattern := `
apiVersion: shopco.platform.syntasso.io/v1alpha1
kind: InventorySyncService
spec:
  environment: production
`
	errs = checkExampleFile("bad-pattern.yaml", []byte(badPattern), promises)
	if !containsSubstring(errs, "does not match pattern") {
		t.Fatalf("expected a pattern-mismatch error, got %v", errs)
	}
}

func TestCheckExampleFileFlagsNoMatchingPromise(t *testing.T) {
	errs := checkExampleFile("orphan.yaml", []byte(`
apiVersion: some.other.group/v1
kind: SomethingElse
spec: {}
`), map[string]promiseDoc{})
	if !containsSubstring(errs, "matches no loaded Promise") {
		t.Fatalf("expected a no-matching-Promise error, got %v", errs)
	}
}

func TestRunLevel1GatesSkippedWhenDirsAbsent(t *testing.T) {
	errs := runLevel1Gates("/nonexistent/promises", "/nonexistent/examples", os.Stderr)
	if len(errs) != 0 {
		t.Fatalf("expected gate to be skipped with no errors when dirs are absent, got %v", errs)
	}
}

func TestRunLevel1GatesEndToEnd(t *testing.T) {
	dir := t.TempDir()
	promiseDir := filepath.Join(dir, "promises")
	exampleDir := filepath.Join(dir, "resource-requests")
	if err := os.Mkdir(promiseDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(exampleDir, 0o755); err != nil {
		t.Fatal(err)
	}

	writeTemp(t, promiseDir, "inventory-sync-service.yaml", validPromiseYAML)
	writeTemp(t, exampleDir, "good.yaml", `
apiVersion: shopco.platform.syntasso.io/v1alpha1
kind: InventorySyncService
spec:
  environment: prod
`)

	errs := runLevel1Gates(promiseDir, exampleDir, os.Stderr)
	if len(errs) != 0 {
		t.Fatalf("expected a clean end-to-end run, got %v", errs)
	}
}

func containsSubstring(errs []string, substr string) bool {
	for _, e := range errs {
		if strings.Contains(e, substr) {
			return true
		}
	}
	return false
}
