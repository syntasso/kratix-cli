package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/syntasso/kratix/api/v1alpha1"
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
    metadata:
      name: inventorysyncservices.shopco.platform.syntasso.io
    spec:
      group: shopco.platform.syntasso.io
      scope: Namespaced
      names:
        kind: InventorySyncService
        plural: inventorysyncservices
        singular: inventorysyncservice
      versions:
        - name: v1alpha1
          served: true
          storage: true
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
        - apiVersion: platform.kratix.io/v1alpha1
          kind: Pipeline
          metadata:
            name: configure-pipeline
          spec:
            containers:
              - name: configure
                image: busybox
      delete:
        - apiVersion: platform.kratix.io/v1alpha1
          kind: Pipeline
          metadata:
            name: delete-pipeline
          spec:
            containers:
              - name: delete
                image: busybox
    promise:
      configure:
        - apiVersion: platform.kratix.io/v1alpha1
          kind: Pipeline
          metadata:
            name: promise-configure-pipeline
          spec:
            containers:
              - name: configure
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
	errs, promise, err := checkPromiseFile("inventory-sync-service.yaml", []byte(validPromiseYAML))
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if len(errs) != 0 {
		t.Fatalf("expected no errors, got %v", errs)
	}
	if promise.Name != "inventory-sync-service" {
		t.Fatalf("expected promise to be parsed, got name %q", promise.Name)
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

// Regression test for review feedback: the pipeline-image gate only
// inspected spec.workflows.resource.configure/.delete, so a missing image
// under spec.workflows.promise.configure (as seen in generated Terraform
// fixture output) silently passed.
func TestCheckPromiseFileFlagsMissingPromiseWorkflowImage(t *testing.T) {
	bad := strings.Replace(validPromiseYAML,
		"name: promise-configure-pipeline\n          spec:\n            containers:\n              - name: configure\n                image: busybox",
		"name: promise-configure-pipeline\n          spec:\n            containers:\n              - name: configure\n                image: \"\"", 1)
	errs, _, err := checkPromiseFile("bad-promise-image.yaml", []byte(bad))
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if !containsSubstring(errs, "promise.configure pipeline container \"configure\" has no image set") {
		t.Fatalf("expected a missing-image error for promise.configure, got %v", errs)
	}
}

// Regression test: a pipeline declared with an empty containers list isn't
// runnable, so it should be flagged the same way a container missing an
// image is.
func TestCheckPromiseFileFlagsPipelineWithNoContainers(t *testing.T) {
	bad := strings.Replace(validPromiseYAML,
		"name: promise-configure-pipeline\n          spec:\n            containers:\n              - name: configure\n                image: busybox",
		"name: promise-configure-pipeline\n          spec:\n            containers: []", 1)
	errs, _, err := checkPromiseFile("bad-empty-containers.yaml", []byte(bad))
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if !containsSubstring(errs, "promise.configure pipeline \"promise-configure-pipeline\" has no containers") {
		t.Fatalf("expected a no-containers error for promise.configure, got %v", errs)
	}
}

func TestCheckPromiseFileFlagsMultipleDeletePipelines(t *testing.T) {
	bad := strings.Replace(validPromiseYAML, `      delete:
        - apiVersion: platform.kratix.io/v1alpha1
          kind: Pipeline
          metadata:
            name: delete-pipeline
          spec:
            containers:
              - name: delete
                image: busybox`, `      delete:
        - apiVersion: platform.kratix.io/v1alpha1
          kind: Pipeline
          metadata:
            name: delete-pipeline-one
          spec:
            containers:
              - name: delete-one
                image: busybox
        - apiVersion: platform.kratix.io/v1alpha1
          kind: Pipeline
          metadata:
            name: delete-pipeline-two
          spec:
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

// Regression test for review feedback: the CRD gate accepted a Promise as
// long as apiVersion/kind/scope/a versions-entry were present, even when
// required CRD fields like names.plural were missing - a Promise that
// cannot actually be installed. Decoding into the real CRD type and running
// it through the real Kubernetes CRD validator must catch this.
func TestCheckPromiseFileFlagsCRDMissingRequiredNamesPlural(t *testing.T) {
	bad := strings.Replace(validPromiseYAML, "        plural: inventorysyncservices\n", "", 1)
	errs, _, err := checkPromiseFile("bad-crd-names.yaml", []byte(bad))
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if !containsSubstring(errs, "plural") {
		t.Fatalf("expected a missing-names.plural error, got %v", errs)
	}
}

// Regression test: a CRD version with neither served nor storage set to
// true cannot be installed (there must be exactly one storage version).
func TestCheckPromiseFileFlagsCRDMissingStorageVersion(t *testing.T) {
	bad := strings.Replace(validPromiseYAML, "          storage: true\n", "          storage: false\n", 1)
	errs, _, err := checkPromiseFile("bad-crd-storage.yaml", []byte(bad))
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if len(errs) == 0 {
		t.Fatalf("expected an error for a CRD with no storage version, got none")
	}
}

func TestCheckExampleFileValidatesRequiredAndPattern(t *testing.T) {
	_, promise, err := checkPromiseFile("inventory-sync-service.yaml", []byte(validPromiseYAML))
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	promises := map[string]*v1alpha1.Promise{"inventory-sync-service.yaml": promise}
	gvkIndex, gvkErrs := buildGVKIndex([]string{"inventory-sync-service.yaml"}, promises)
	if len(gvkErrs) != 0 {
		t.Fatalf("unexpected GVK index errors: %v", gvkErrs)
	}

	goodExample := `
apiVersion: shopco.platform.syntasso.io/v1alpha1
kind: InventorySyncService
spec:
  environment: prod
`
	errs := checkExampleFile("good.yaml", []byte(goodExample), gvkIndex)
	if len(errs) != 0 {
		t.Fatalf("expected no errors for a valid example, got %v", errs)
	}

	missingRequired := `
apiVersion: shopco.platform.syntasso.io/v1alpha1
kind: InventorySyncService
spec: {}
`
	errs = checkExampleFile("missing-required.yaml", []byte(missingRequired), gvkIndex)
	if !containsSubstring(errs, "environment") {
		t.Fatalf("expected a missing-required-field error, got %v", errs)
	}

	// Kubernetes structural schemas forbid combining "properties" with
	// "additionalProperties" on the same object (they'd contradict the API
	// convention of pruning unknown fields rather than rejecting them), so
	// a real CRD can never reject an unrecognised field the way the old
	// hand-rolled validator did. An extra field is allowed here, matching
	// real Kubernetes CRD validation behaviour.
	unknownField := `
apiVersion: shopco.platform.syntasso.io/v1alpha1
kind: InventorySyncService
spec:
  environment: prod
  bogusField: true
`
	errs = checkExampleFile("unknown-field.yaml", []byte(unknownField), gvkIndex)
	if len(errs) != 0 {
		t.Fatalf("expected an unrecognised field to be allowed (pruned), got %v", errs)
	}

	badPattern := `
apiVersion: shopco.platform.syntasso.io/v1alpha1
kind: InventorySyncService
spec:
  environment: production
`
	errs = checkExampleFile("bad-pattern.yaml", []byte(badPattern), gvkIndex)
	if !containsSubstring(errs, "environment") {
		t.Fatalf("expected a pattern-mismatch error, got %v", errs)
	}
}

// Regression test for review feedback: a property with no pattern accepted
// any value regardless of its declared type, so a YAML string could satisfy
// a schema that declares `type: integer`. Kubernetes would reject this.
func TestCheckExampleFileFlagsNonStringValueForIntegerType(t *testing.T) {
	withReplicas := strings.Replace(validPromiseYAML,
		"                    environment:\n                      type: string\n                      pattern: \"^(prod|staging)$\"\n",
		"                    environment:\n                      type: string\n                      pattern: \"^(prod|staging)$\"\n                    replicas:\n                      type: integer\n", 1)
	_, promise, err := checkPromiseFile("inventory-sync-service.yaml", []byte(withReplicas))
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	promises := map[string]*v1alpha1.Promise{"inventory-sync-service.yaml": promise}
	gvkIndex, _ := buildGVKIndex([]string{"inventory-sync-service.yaml"}, promises)

	stringForInteger := `
apiVersion: shopco.platform.syntasso.io/v1alpha1
kind: InventorySyncService
spec:
  environment: prod
  replicas: "3"
`
	errs := checkExampleFile("bad-type.yaml", []byte(stringForInteger), gvkIndex)
	if !containsSubstring(errs, "replicas") {
		t.Fatalf("expected a type-mismatch error for replicas, got %v", errs)
	}
}

// Regression test for review feedback: required fields nested below the top
// level of spec were never recursed into, so a missing nested required
// field silently passed.
func TestCheckExampleFileFlagsMissingNestedRequiredField(t *testing.T) {
	withConfig := strings.Replace(validPromiseYAML,
		"                    environment:\n                      type: string\n                      pattern: \"^(prod|staging)$\"\n",
		"                    environment:\n                      type: string\n                      pattern: \"^(prod|staging)$\"\n                    config:\n                      type: object\n                      required: [\"region\"]\n                      properties:\n                        region:\n                          type: string\n", 1)
	_, promise, err := checkPromiseFile("inventory-sync-service.yaml", []byte(withConfig))
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	promises := map[string]*v1alpha1.Promise{"inventory-sync-service.yaml": promise}
	gvkIndex, _ := buildGVKIndex([]string{"inventory-sync-service.yaml"}, promises)

	missingNestedRequired := `
apiVersion: shopco.platform.syntasso.io/v1alpha1
kind: InventorySyncService
spec:
  environment: prod
  config: {}
`
	errs := checkExampleFile("missing-nested-required.yaml", []byte(missingNestedRequired), gvkIndex)
	if !containsSubstring(errs, "region") {
		t.Fatalf("expected a missing-nested-required-field error for config.region, got %v", errs)
	}
}

// Regression test for review feedback: an example whose apiVersion names a
// CRD version that exists but has served: false cannot actually be
// submitted to that API (Kubernetes would 404 it), so it must not be
// treated as a valid match.
func TestCheckExampleFileFlagsUnservedVersion(t *testing.T) {
	twoVersions := strings.Replace(validPromiseYAML, `      versions:
        - name: v1alpha1
          served: true
          storage: true`, `      versions:
        - name: v1alpha1
          served: true
          storage: true
        - name: v1beta1
          served: false
          storage: false`, 1)

	_, promise, err := checkPromiseFile("inventory-sync-service.yaml", []byte(twoVersions))
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	promises := map[string]*v1alpha1.Promise{"inventory-sync-service.yaml": promise}
	gvkIndex, gvkErrs := buildGVKIndex([]string{"inventory-sync-service.yaml"}, promises)
	if len(gvkErrs) != 0 {
		t.Fatalf("unexpected GVK index errors: %v", gvkErrs)
	}

	unservedExample := `
apiVersion: shopco.platform.syntasso.io/v1beta1
kind: InventorySyncService
spec:
  environment: prod
`
	errs := checkExampleFile("unserved-version.yaml", []byte(unservedExample), gvkIndex)
	if !containsSubstring(errs, "not served") {
		t.Fatalf("expected a not-served error, got %v", errs)
	}
}

// Regression test: validating only the extracted spec sub-schema missed
// root-level schema rules such as a top-level `required: ["spec"]` - an
// example with no spec at all would pass. Validating the whole document
// against the whole schema catches this.
func TestCheckExampleFileFlagsMissingRootRequiredSpec(t *testing.T) {
	promiseYAML := `
apiVersion: platform.kratix.io/v1alpha1
kind: Promise
metadata:
  name: widget-service
spec:
  api:
    apiVersion: apiextensions.k8s.io/v1
    kind: CustomResourceDefinition
    metadata:
      name: widgets.shopco.platform.syntasso.io
    spec:
      group: shopco.platform.syntasso.io
      scope: Namespaced
      names:
        kind: Widget
        plural: widgets
        singular: widget
      versions:
        - name: v1alpha1
          served: true
          storage: true
          schema:
            openAPIV3Schema:
              type: object
              required: ["spec"]
              properties:
                spec:
                  type: object
  workflows:
    resource:
      configure:
        - apiVersion: platform.kratix.io/v1alpha1
          kind: Pipeline
          metadata:
            name: configure-pipeline
          spec:
            containers:
              - name: configure
                image: busybox
`
	_, promise, err := checkPromiseFile("widget-service.yaml", []byte(promiseYAML))
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	promises := map[string]*v1alpha1.Promise{"widget-service.yaml": promise}
	gvkIndex, gvkErrs := buildGVKIndex([]string{"widget-service.yaml"}, promises)
	if len(gvkErrs) != 0 {
		t.Fatalf("unexpected GVK index errors: %v", gvkErrs)
	}

	missingSpec := `
apiVersion: shopco.platform.syntasso.io/v1alpha1
kind: Widget
`
	errs := checkExampleFile("missing-spec.yaml", []byte(missingSpec), gvkIndex)
	if !containsSubstring(errs, "spec") {
		t.Fatalf("expected a missing-root-required-spec error, got %v", errs)
	}
}

// Regression test for review feedback: a multi-document example file
// (several `---`-separated Resource Requests, a common and valid pattern)
// used to have only its first document checked - a broken second document
// silently passed the gate. Every document must be checked.
func TestCheckExampleFileValidatesEveryDocumentInAMultiDocFile(t *testing.T) {
	_, promise, err := checkPromiseFile("inventory-sync-service.yaml", []byte(validPromiseYAML))
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	promises := map[string]*v1alpha1.Promise{"inventory-sync-service.yaml": promise}
	gvkIndex, gvkErrs := buildGVKIndex([]string{"inventory-sync-service.yaml"}, promises)
	if len(gvkErrs) != 0 {
		t.Fatalf("unexpected GVK index errors: %v", gvkErrs)
	}

	multiDoc := `
apiVersion: shopco.platform.syntasso.io/v1alpha1
kind: InventorySyncService
spec:
  environment: prod
---
apiVersion: shopco.platform.syntasso.io/v1alpha1
kind: InventorySyncService
spec: {}
`
	errs := checkExampleFile("multi.yaml", []byte(multiDoc), gvkIndex)
	if !containsSubstring(errs, "environment") {
		t.Fatalf("expected the second document's missing-required-field error to be reported, got %v", errs)
	}
}

func TestCheckExampleFileFlagsNoMatchingPromise(t *testing.T) {
	errs := checkExampleFile("orphan.yaml", []byte(`
apiVersion: some.other.group/v1
kind: SomethingElse
spec: {}
`), map[string]gvkPromise{})
	if !containsSubstring(errs, "matches no loaded Promise") {
		t.Fatalf("expected a no-matching-Promise error, got %v", errs)
	}
}

// Regression test for review feedback: when two loaded Promises share the
// same group/version/kind but different schemas, map iteration used to
// pick an arbitrary one to validate an example against - nondeterministic
// pass/fail depending on Go map order. buildGVKIndex must instead flag the
// duplicate as a gate failure, and must do so the same way every time.
func TestBuildGVKIndexFlagsDuplicateGVKAcrossPromises(t *testing.T) {
	secondPromise := strings.NewReplacer(
		"name: inventory-sync-service", "name: inventory-sync-service-v2",
		"name: inventorysyncservices.shopco.platform.syntasso.io", "name: inventorysyncservices2.shopco.platform.syntasso.io",
		"plural: inventorysyncservices", "plural: inventorysyncservices2",
		"singular: inventorysyncservice", "singular: inventorysyncservice2",
	).Replace(validPromiseYAML)

	_, promiseOne, err := checkPromiseFile("a.yaml", []byte(validPromiseYAML))
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	_, promiseTwo, err := checkPromiseFile("b.yaml", []byte(secondPromise))
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	promises := map[string]*v1alpha1.Promise{"a.yaml": promiseOne, "b.yaml": promiseTwo}

	for i := 0; i < 10; i++ {
		gvkIndex, errs := buildGVKIndex([]string{"a.yaml", "b.yaml"}, promises)
		if !containsSubstring(errs, "declared by both a.yaml and b.yaml") {
			t.Fatalf("expected a deterministic duplicate-GVK error naming both files, got %v", errs)
		}
		if entry, ok := gvkIndex["shopco.platform.syntasso.io/v1alpha1/InventorySyncService"]; !ok || entry.promiseName != "a.yaml" {
			t.Fatalf("expected the index to deterministically keep the first-seen Promise (a.yaml), got %+v", gvkIndex)
		}
	}
}

func TestRunLevelOneGatesSkippedWhenDirsAbsent(t *testing.T) {
	errs := runLevelOneGates("/nonexistent/promises", "/nonexistent/examples", os.Stderr)
	if len(errs) != 0 {
		t.Fatalf("expected gate to be skipped with no errors when dirs are absent, got %v", errs)
	}
}

func TestRunLevelOneGatesEndToEnd(t *testing.T) {
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

	errs := runLevelOneGates(promiseDir, exampleDir, os.Stderr)
	if len(errs) != 0 {
		t.Fatalf("expected a clean end-to-end run, got %v", errs)
	}
}

// Regression test for review feedback: readYAMLDir used to silently drop
// any file it couldn't read (permission denied, corrupt, ...), so the gate
// could report success without having actually validated everything it was
// asked to check. A missing directory must still be a no-op (the gate is
// intentionally skipped), but an unreadable *file* inside an existing
// directory must surface as a gate failure.
func TestReadYAMLDirSkipsOnlyMissingDirectory(t *testing.T) {
	files, errs := readYAMLDir(filepath.Join(t.TempDir(), "does-not-exist"))
	if len(errs) != 0 {
		t.Fatalf("expected a missing directory to be skipped with no errors, got %v", errs)
	}
	if len(files) != 0 {
		t.Fatalf("expected no files from a missing directory, got %v", files)
	}
}

func TestReadYAMLDirReportsUnreadableFile(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("file permissions are not enforced when running as root")
	}

	dir := t.TempDir()
	path := writeTemp(t, dir, "unreadable.yaml", validPromiseYAML)
	if err := os.Chmod(path, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(path, 0o644) })

	_, errs := readYAMLDir(dir)
	if !containsSubstring(errs, "unreadable.yaml") {
		t.Fatalf("expected an unreadable file to be reported as a gate failure, got %v", errs)
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
