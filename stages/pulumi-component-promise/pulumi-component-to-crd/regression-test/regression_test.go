package regressiontest

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
)

var buildBinaryOnce sync.Once

func TestRegressionSuite_LocalFixtures(t *testing.T) {
	t.Parallel()

	binPath := buildBinary(t)

	t.Run("success_valid_schema", func(t *testing.T) {
		t.Parallel()

		stdout, stderr, code := runBinary(t, binPath,
			"--in", schemaFixture(t, "schema.valid.json"),
		)
		writeArtifacts(t, "success_valid_schema", stdout, stderr)

		if code != 0 {
			t.Fatalf("exit code mismatch: got %d, want 0, stderr=%q", code, stderr)
		}
		if stderr != "" {
			t.Fatalf("expected empty stderr, got %q", stderr)
		}
		assertContainsAll(t, stdout,
			"apiVersion: apiextensions.k8s.io/v1",
			"kind: CustomResourceDefinition",
			"openAPIV3Schema:",
			"required:",
			"default: 2",
		)
	})

	t.Run("error_multiple_components_without_component", func(t *testing.T) {
		t.Parallel()

		stdout, stderr, code := runBinary(t, binPath,
			"--in", schemaFixture(t, "schema.multi-components.json"),
		)
		writeArtifacts(t, "error_multiple_components_without_component", stdout, stderr)

		if code != 2 {
			t.Fatalf("exit code mismatch: got %d, want 2", code)
		}
		if stdout != "" {
			t.Fatalf("expected empty stdout, got %q", stdout)
		}
		assertExactLine(t, stderr, `error: multiple components found; provide --component from: pkg:index:Alpha, pkg:index:Zulu`)
	})

	t.Run("success_explicit_component_multi_schema", func(t *testing.T) {
		t.Parallel()

		stdout, stderr, code := runBinary(t, binPath,
			"--in", schemaFixture(t, "schema.with-component.json"),
			"--component", "pkg:index:Thing",
		)
		writeArtifacts(t, "success_explicit_component_multi_schema", stdout, stderr)

		if code != 0 {
			t.Fatalf("exit code mismatch: got %d, want 0, stderr=%q", code, stderr)
		}
		if stderr != "" {
			t.Fatalf("expected empty stderr, got %q", stderr)
		}
		assertContainsAll(t, stdout, "openAPIV3Schema:", "name:")
	})

	t.Run("error_unknown_component", func(t *testing.T) {
		t.Parallel()

		stdout, stderr, code := runBinary(t, binPath,
			"--in", schemaFixture(t, "schema.unknown-component.json"),
			"--component", "pkg:index:Missing",
		)
		writeArtifacts(t, "error_unknown_component", stdout, stderr)

		if code != 2 {
			t.Fatalf("exit code mismatch: got %d, want 2", code)
		}
		if stdout != "" {
			t.Fatalf("expected empty stdout, got %q", stdout)
		}
		assertExactLine(t, stderr, `error: component "pkg:index:Missing" not found; available components: pkg:index:Alpha, pkg:index:Zulu`)
	})

	t.Run("error_unsupported_construct", func(t *testing.T) {
		t.Parallel()

		stdout, stderr, code := runBinary(t, binPath,
			"--in", schemaFixture(t, "schema.unsupported.json"),
		)
		writeArtifacts(t, "error_unsupported_construct", stdout, stderr)

		if code != 2 {
			t.Fatalf("exit code mismatch: got %d, want 2", code)
		}
		if stdout != "" {
			t.Fatalf("expected empty stdout, got %q", stdout)
		}
		assertContainsAll(t, stderr,
			`error: no translatable spec fields remain after skipping unsupported schema paths for component "pkg:index:Thing"`,
		)
	})

	t.Run("error_malformed_schema_preflight", func(t *testing.T) {
		t.Parallel()

		stdout, stderr, code := runBinary(t, binPath,
			"--in", schemaFixture(t, "schema.malformed.json"),
		)
		writeArtifacts(t, "error_malformed_schema_preflight", stdout, stderr)

		if code != 2 {
			t.Fatalf("exit code mismatch: got %d, want 2", code)
		}
		if stdout != "" {
			t.Fatalf("expected empty stdout, got %q", stdout)
		}
		assertContainsAll(t, stderr,
			`error: schema preflight component "pkg:index:Thing"`,
			`resources.pkg:index:Thing.inputProperties.value`,
			`unresolved local type ref "#/types/pkg:index:Missing"`,
		)
	})
}

func TestRegressionSuite_URLInputLiveRegistry(t *testing.T) {
	t.Parallel()
	if os.Getenv("RUN_INTERNET_TESTS") != "1" {
		t.Skip("set RUN_INTERNET_TESTS=1 to enable internet-backed regression tests")
	}

	binPath := buildBinary(t)

	t.Run("url_input_live_registry", func(t *testing.T) {
		t.Parallel()

		eksStdout, eksStderr, eksCode := runBinary(t, binPath,
			"--in", "https://www.pulumi.com/registry/packages/eks/schema.json",
			"--component", "eks:index:Cluster",
		)
		writeArtifacts(t, "url_input_live_registry_eks", eksStdout, eksStderr)

		if eksCode != 0 && eksCode != 2 {
			t.Fatalf("unexpected EKS URL exit code: got %d, stderr=%q", eksCode, eksStderr)
		}
		if strings.Contains(eksStderr, "fetch input schema URL:") {
			t.Fatalf("unexpected URL fetch failure for EKS schema URL: %q", eksStderr)
		}
		if eksCode == 0 && eksStdout == "" {
			t.Fatalf("expected stdout for successful EKS URL run")
		}
		if eksCode == 2 && !strings.Contains(eksStderr, "error:") {
			t.Fatalf("expected parseable error line for EKS URL failure, got %q", eksStderr)
		}

		missingStdout, missingStderr, missingCode := runBinary(t, binPath,
			"--in", "https://www.pulumi.com/registry/packages/eks/does-not-exist.json",
			"--component", "eks:index:Cluster",
		)
		writeArtifacts(t, "url_input_live_registry_eks_missing", missingStdout, missingStderr)

		if missingCode != 2 {
			t.Fatalf("missing URL exit code mismatch: got %d, want 2", missingCode)
		}
		if missingStdout != "" {
			t.Fatalf("expected empty stdout for missing URL case, got %q", missingStdout)
		}
		assertContainsAll(t, missingStderr,
			"error: fetch input schema URL: unexpected status 404 for https://www.pulumi.com/registry/packages/eks/does-not-exist.json",
		)
	})
}

func buildBinary(t *testing.T) string {
	t.Helper()

	repoDir := componentRoot(t)
	binPath := filepath.Join(repoDir, "bin", "pulumi-component-to-crd")

	buildBinaryOnce.Do(func() {
		cmd := exec.Command(filepath.Join(repoDir, "scripts", "build_binary"))
		cmd.Dir = repoDir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("build binary: %v\n%s", err, string(out))
		}
	})

	return binPath
}

func runBinary(t *testing.T, binPath string, args ...string) (string, string, int) {
	t.Helper()

	cmd := exec.Command(binPath, args...)
	cmd.Dir = componentRoot(t)
	stdout, err := cmd.Output()
	if err == nil {
		return string(stdout), "", 0
	}

	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("run binary: %v", err)
	}
	return string(stdout), string(exitErr.Stderr), exitErr.ExitCode()
}

func writeArtifacts(t *testing.T, name, stdout, stderr string) {
	t.Helper()

	baseDir := filepath.Join(regressionTestDir(t), "work", "cases", name)
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		t.Fatalf("create artifact dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(baseDir, "stdout.txt"), []byte(stdout), 0o600); err != nil {
		t.Fatalf("write stdout artifact: %v", err)
	}
	if err := os.WriteFile(filepath.Join(baseDir, "stderr.txt"), []byte(stderr), 0o600); err != nil {
		t.Fatalf("write stderr artifact: %v", err)
	}
}

func schemaFixture(t *testing.T, name string) string {
	t.Helper()
	return filepath.Join(regressionTestDir(t), "testdata", "schemas", name)
}

func assertContainsAll(t *testing.T, text string, parts ...string) {
	t.Helper()
	for _, part := range parts {
		if !strings.Contains(text, part) {
			t.Fatalf("missing substring %q in %q", part, text)
		}
	}
}

func assertExactLine(t *testing.T, got, want string) {
	t.Helper()
	trimmed := strings.TrimSpace(got)
	if trimmed != want {
		t.Fatalf("line mismatch:\n got: %q\nwant: %q", trimmed, want)
	}
}

func componentRoot(t *testing.T) string {
	t.Helper()

	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path")
	}
	return filepath.Dir(filepath.Dir(thisFile))
}

func regressionTestDir(t *testing.T) string {
	t.Helper()
	return filepath.Join(componentRoot(t), "regression-test")
}
