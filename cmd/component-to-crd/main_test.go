package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	t.Parallel()

	tempDir := makeWorkspaceTempDir(t)
	singleComponentSchemaPath := filepath.Join(tempDir, "single-component.json")
	multiComponentSchemaPath := filepath.Join(tempDir, "multi-component.json")
	zeroComponentSchemaPath := filepath.Join(tempDir, "zero-component.json")
	invalidSchemaPath := filepath.Join(tempDir, "invalid.json")
	unsupportedSchemaPath := filepath.Join(tempDir, "unsupported.json")
	malformedSchemaPath := filepath.Join(tempDir, "malformed.json")
	malformedPrecedenceSchemaPath := filepath.Join(tempDir, "malformed-precedence.json")
	resourceRefSchemaPath := filepath.Join(tempDir, "resource-ref.json")

	if err := os.WriteFile(singleComponentSchemaPath, []byte(`{"resources":{"pkg:index:Thing":{"isComponent":true,"inputProperties":{"zeta":{"type":"string"},"alpha":{"type":"number","default":1.5}},"requiredInputs":["zeta","alpha"]}}}`), 0o600); err != nil {
		t.Fatalf("write single component fixture: %v", err)
	}

	if err := os.WriteFile(multiComponentSchemaPath, []byte(`{"resources":{"pkg:index:Zulu":{"isComponent":true,"inputProperties":{"v":{"type":"string"}}},"pkg:index:Alpha":{"isComponent":true,"inputProperties":{"w":{"type":"string"}}},"pkg:index:Other":{"isComponent":false}}}`), 0o600); err != nil {
		t.Fatalf("write multi component fixture: %v", err)
	}

	if err := os.WriteFile(zeroComponentSchemaPath, []byte(`{"resources":{"pkg:index:Other":{"isComponent":false}}}`), 0o600); err != nil {
		t.Fatalf("write zero component fixture: %v", err)
	}

	if err := os.WriteFile(invalidSchemaPath, []byte(`{"resources":`), 0o600); err != nil {
		t.Fatalf("write invalid schema fixture: %v", err)
	}

	if err := os.WriteFile(unsupportedSchemaPath, []byte(`{"resources":{"pkg:index:Thing":{"isComponent":true,"inputProperties":{"value":{"oneOf":[{"type":"string"},{"type":"number"}]}}}}}`), 0o600); err != nil {
		t.Fatalf("write unsupported fixture: %v", err)
	}

	if err := os.WriteFile(malformedSchemaPath, []byte(`{"resources":{"pkg:index:Thing":{"isComponent":true,"inputProperties":{"value":{"$ref":"#/types/pkg:index:Missing"}}}}}`), 0o600); err != nil {
		t.Fatalf("write malformed schema fixture: %v", err)
	}

	if err := os.WriteFile(malformedPrecedenceSchemaPath, []byte(`{"resources":{"pkg:index:Zulu":{"isComponent":true,"inputProperties":{"bad":{"$ref":"#/types/pkg:index:Missing"}}},"pkg:index:Alpha":{"isComponent":true,"inputProperties":{"value":{"oneOf":[{"type":"string"},{"type":"number"}]}}}}}`), 0o600); err != nil {
		t.Fatalf("write malformed precedence fixture: %v", err)
	}

	if err := os.WriteFile(resourceRefSchemaPath, []byte(`{"resources":{"eks:index:Addon":{"isComponent":true,"inputProperties":{"cluster":{"$ref":"#/resources/eks:index:Cluster"},"addonName":{"type":"string"}},"requiredInputs":["cluster","addonName"]},"eks:index:Cluster":{"isComponent":true,"inputProperties":{"name":{"type":"string"}},"requiredInputs":["name"]}}}`), 0o600); err != nil {
		t.Fatalf("write resource ref fixture: %v", err)
	}

	urlSchemaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"resources":{"pkg:index:Thing":{"isComponent":true,"inputProperties":{"name":{"type":"string"}}}}}`))
	}))
	t.Cleanup(urlSchemaServer.Close)

	urlNotFoundServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "missing", http.StatusNotFound)
	}))
	t.Cleanup(urlNotFoundServer.Close)

	tests := []struct {
		name            string
		args            []string
		wantExitCode    int
		wantStdoutParts []string
		wantStderrParts []string
		wantStderrNot   []string
	}{
		{
			name:         "single component auto-select",
			args:         []string{"--in", singleComponentSchemaPath},
			wantExitCode: exitSuccess,
			wantStdoutParts: []string{
				"apiVersion: apiextensions.k8s.io/v1",
				"kind: CustomResourceDefinition",
				`name: "things.pkg.components.platform"`,
				"group: pkg.components.platform",
				"kind: Thing",
				"plural: things",
				"singular: thing",
				"name: v1alpha1",
				"spec:",
				"required:",
				"- \"alpha\"",
				"- \"zeta\"",
				"default: 1.5",
			},
		},
		{
			name: "custom identity flags",
			args: []string{
				"--in", singleComponentSchemaPath,
				"--group", "apps.example.com",
				"--version", "v1",
				"--kind", "ServiceDeployment",
				"--plural", "servicedeployments",
				"--singular", "servicedeployment",
			},
			wantExitCode: exitSuccess,
			wantStdoutParts: []string{
				`name: "servicedeployments.apps.example.com"`,
				"group: apps.example.com",
				"kind: ServiceDeployment",
				"plural: servicedeployments",
				"singular: servicedeployment",
				"name: v1",
			},
		},
		{
			name: "mixed identity overrides preserve derived defaults",
			args: []string{
				"--in", singleComponentSchemaPath,
				"--group", "apps.example.com",
			},
			wantExitCode: exitSuccess,
			wantStdoutParts: []string{
				`name: "things.apps.example.com"`,
				"group: apps.example.com",
				"kind: Thing",
				"plural: things",
				"singular: thing",
				"name: v1alpha1",
			},
		},
		{
			name:         "explicit component from multi schema",
			args:         []string{"--in", multiComponentSchemaPath, "--component", "pkg:index:Zulu"},
			wantExitCode: exitSuccess,
			wantStdoutParts: []string{
				"apiVersion: apiextensions.k8s.io/v1",
				"kind: CustomResourceDefinition",
			},
		},
		{
			name:         "local resource ref translates successfully",
			args:         []string{"--in", resourceRefSchemaPath, "--component", "eks:index:Addon"},
			wantExitCode: exitSuccess,
			wantStdoutParts: []string{
				"cluster:",
				"required:",
				"- \"addonName\"",
				"- \"cluster\"",
				"name:",
				"- \"name\"",
			},
		},
		{
			name:            "unknown component",
			args:            []string{"--in", multiComponentSchemaPath, "--component", "pkg:index:Missing"},
			wantExitCode:    exitUserError,
			wantStderrParts: []string{"error:", `component "pkg:index:Missing" not found; available components: pkg:index:Alpha, pkg:index:Zulu`},
		},
		{
			name:            "multiple components require explicit token",
			args:            []string{"--in", multiComponentSchemaPath},
			wantExitCode:    exitUserError,
			wantStderrParts: []string{"error:", "multiple components found; provide --component from: pkg:index:Alpha, pkg:index:Zulu"},
		},
		{
			name:            "zero components",
			args:            []string{"--in", zeroComponentSchemaPath},
			wantExitCode:    exitUserError,
			wantStderrParts: []string{"error:", "no component resources found in schema"},
		},
		{
			name:            "unsupported construct is exit 3",
			args:            []string{"--in", unsupportedSchemaPath},
			wantExitCode:    exitUnsupported,
			wantStderrParts: []string{"error:", `component "pkg:index:Thing" path "spec.value" unsupported construct`},
		},
		{
			name:         "malformed schema returns exit 2 preflight error",
			args:         []string{"--in", malformedSchemaPath},
			wantExitCode: exitUserError,
			wantStderrParts: []string{
				"error:",
				"schema preflight path",
				`resources.pkg:index:Thing.inputProperties.value`,
				`unresolved local type ref "#/types/pkg:index:Missing"`,
			},
		},
		{
			name:         "malformed preflight error has precedence over component selection error",
			args:         []string{"--in", malformedPrecedenceSchemaPath, "--component", "pkg:index:Missing"},
			wantExitCode: exitUserError,
			wantStderrParts: []string{
				"error:",
				"schema preflight path",
				`unresolved local type ref "#/types/pkg:index:Missing"`,
			},
			wantStderrNot: []string{
				`component "pkg:index:Missing" not found`,
			},
		},
		{
			name:         "malformed preflight error has precedence over unsupported construct",
			args:         []string{"--in", malformedPrecedenceSchemaPath, "--component", "pkg:index:Alpha"},
			wantExitCode: exitUserError,
			wantStderrParts: []string{
				"error:",
				"schema preflight path",
				`unresolved local type ref "#/types/pkg:index:Missing"`,
			},
			wantStderrNot: []string{
				"unsupported construct",
			},
		},
		{
			name:            "missing in flag",
			args:            []string{},
			wantExitCode:    exitUserError,
			wantStderrParts: []string{"error:", "missing required flag: --in"},
		},
		{
			name:            "input file does not exist",
			args:            []string{"--in", filepath.Join(tempDir, "missing.json")},
			wantExitCode:    exitUserError,
			wantStderrParts: []string{"error:", "read input schema file:"},
		},
		{
			name:            "input file is invalid json",
			args:            []string{"--in", invalidSchemaPath},
			wantExitCode:    exitUserError,
			wantStderrParts: []string{"error:", "parse input schema as JSON:"},
		},
		{
			name:         "url input success",
			args:         []string{"--in", urlSchemaServer.URL},
			wantExitCode: exitSuccess,
			wantStdoutParts: []string{
				"apiVersion: apiextensions.k8s.io/v1",
				"kind: CustomResourceDefinition",
			},
		},
		{
			name:            "url input non-200",
			args:            []string{"--in", urlNotFoundServer.URL},
			wantExitCode:    exitUserError,
			wantStderrParts: []string{"error:", "fetch input schema URL: unexpected status 404 for"},
		},
		{
			name:            "invalid group flag",
			args:            []string{"--in", singleComponentSchemaPath, "--group", "bad_group"},
			wantExitCode:    exitUserError,
			wantStderrParts: []string{"error:", "invalid --group"},
		},
		{
			name:            "invalid version flag",
			args:            []string{"--in", singleComponentSchemaPath, "--version", "v1_alpha1"},
			wantExitCode:    exitUserError,
			wantStderrParts: []string{"error:", "invalid --version"},
		},
		{
			name:            "invalid plural flag",
			args:            []string{"--in", singleComponentSchemaPath, "--plural", "bad_plural"},
			wantExitCode:    exitUserError,
			wantStderrParts: []string{"error:", "invalid --plural"},
		},
		{
			name:            "invalid kind flag",
			args:            []string{"--in", singleComponentSchemaPath, "--kind", "Service-Deployment"},
			wantExitCode:    exitUserError,
			wantStderrParts: []string{"error:", "invalid --kind"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var stdout strings.Builder
			var stderr strings.Builder

			gotExitCode := run(tt.args, &stdout, &stderr)
			if gotExitCode != tt.wantExitCode {
				t.Fatalf("exit code mismatch: got %d, want %d", gotExitCode, tt.wantExitCode)
			}

			gotStdout := stdout.String()
			for _, part := range tt.wantStdoutParts {
				if !strings.Contains(gotStdout, part) {
					t.Fatalf("stdout missing %q in %q", part, gotStdout)
				}
			}

			gotStderr := stderr.String()
			for _, part := range tt.wantStderrParts {
				if !strings.Contains(gotStderr, part) {
					t.Fatalf("stderr missing %q in %q", part, gotStderr)
				}
			}
			for _, part := range tt.wantStderrNot {
				if strings.Contains(gotStderr, part) {
					t.Fatalf("stderr unexpectedly contains %q in %q", part, gotStderr)
				}
			}

			if gotStderr != "" && strings.Count(gotStderr, "\n") != 1 {
				t.Fatalf("stderr should be a single line, got %q", gotStderr)
			}
		})
	}
}

func TestRun_WritesDeterministicOutput(t *testing.T) {
	t.Parallel()

	tempDir := makeWorkspaceTempDir(t)
	schemaPath := filepath.Join(tempDir, "schema.json")

	if err := os.WriteFile(schemaPath, []byte(`{"resources":{"pkg:index:Thing":{"isComponent":true,"inputProperties":{"zeta":{"type":"string"},"alpha":{"type":"number"}},"requiredInputs":["zeta","alpha"]}}}`), 0o600); err != nil {
		t.Fatalf("write schema fixture: %v", err)
	}

	runOnce := func() string {
		t.Helper()

		var stdout strings.Builder
		var stderr strings.Builder
		code := run([]string{"--in", schemaPath}, &stdout, &stderr)
		if code != exitSuccess {
			t.Fatalf("run exit code = %d, stderr = %q", code, stderr.String())
		}
		if stderr.String() != "" {
			t.Fatalf("unexpected stderr: %q", stderr.String())
		}
		return stdout.String()
	}

	a := runOnce()
	b := runOnce()
	if a != b {
		t.Fatalf("output mismatch between runs:\nA:\n%s\nB:\n%s", a, b)
	}

	requiredSnippets := []string{
		"apiVersion: apiextensions.k8s.io/v1",
		"kind: CustomResourceDefinition",
		"openAPIV3Schema:",
		"properties:",
		"spec:",
		"alpha:",
		"zeta:",
		"required:",
	}
	output := a
	for _, snippet := range requiredSnippets {
		if !strings.Contains(output, snippet) {
			t.Fatalf("expected output to contain %q, got:\n%s", snippet, output)
		}
	}
}

func TestRun_OutputWriterFailureReturnsExit4(t *testing.T) {
	t.Parallel()

	tempDir := makeWorkspaceTempDir(t)
	schemaPath := filepath.Join(tempDir, "schema.json")
	if err := os.WriteFile(schemaPath, []byte(`{"resources":{"pkg:index:Thing":{"isComponent":true,"inputProperties":{"name":{"type":"string"}}}}}`), 0o600); err != nil {
		t.Fatalf("write schema fixture: %v", err)
	}

	var stderr strings.Builder
	code := run([]string{"--in", schemaPath}, errWriter{}, &stderr)
	if code != exitOutputError {
		t.Fatalf("exit code mismatch: got %d, want %d", code, exitOutputError)
	}
	if !strings.Contains(stderr.String(), "write CRD output:") {
		t.Fatalf("stderr mismatch: %q", stderr.String())
	}
}

type errWriter struct{}

func (errWriter) Write([]byte) (int, error) {
	return 0, errors.New("writer failure")
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
