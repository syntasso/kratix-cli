package main

import (
	"errors"
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

	if err := os.WriteFile(singleComponentSchemaPath, []byte(`{"resources":{"pkg:index:Thing":{"isComponent":true,"inputProperties":{"zeta":{"type":"string"},"alpha":{"type":"number"}},"requiredInputs":["zeta","alpha"]}}}`), 0o600); err != nil {
		t.Fatalf("write single component fixture: %v", err)
	}

	if err := os.WriteFile(multiComponentSchemaPath, []byte(`{"resources":{"pkg:index:Zulu":{"isComponent":true},"pkg:index:Alpha":{"isComponent":true},"pkg:index:Other":{"isComponent":false}}}`), 0o600); err != nil {
		t.Fatalf("write multi component fixture: %v", err)
	}

	if err := os.WriteFile(zeroComponentSchemaPath, []byte(`{"resources":{"pkg:index:Other":{"isComponent":false}}}`), 0o600); err != nil {
		t.Fatalf("write zero component fixture: %v", err)
	}

	if err := os.WriteFile(invalidSchemaPath, []byte(`{"resources":`), 0o600); err != nil {
		t.Fatalf("write invalid schema fixture: %v", err)
	}

	tests := []struct {
		name            string
		args            []string
		wantExitCode    int
		wantStdoutParts []string
		wantStderrParts []string
	}{
		{
			name:            "single component auto-select",
			args:            []string{"--in", singleComponentSchemaPath},
			wantExitCode:    exitSuccess,
			wantStdoutParts: []string{"apiVersion: apiextensions.k8s.io/v1", "kind: CustomResourceDefinition", "placeholder scaffold for pkg:index:Thing"},
		},
		{
			name:            "explicit component from multi schema",
			args:            []string{"--in", multiComponentSchemaPath, "--component", "pkg:index:Zulu"},
			wantExitCode:    exitSuccess,
			wantStdoutParts: []string{"placeholder scaffold for pkg:index:Zulu"},
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
			name:            "missing in flag",
			args:            []string{},
			wantExitCode:    exitUserError,
			wantStderrParts: []string{"error:", "missing required flag: --in"},
		},
		{
			name:            "input file does not exist",
			args:            []string{"--in", filepath.Join(tempDir, "missing.json")},
			wantExitCode:    exitUserError,
			wantStderrParts: []string{"error:", "read input schema:"},
		},
		{
			name:            "input file is invalid json",
			args:            []string{"--in", invalidSchemaPath},
			wantExitCode:    exitUserError,
			wantStderrParts: []string{"error:", "parse input schema as JSON:"},
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

			if gotStderr != "" && strings.Count(gotStderr, "\n") != 1 {
				t.Fatalf("stderr should be a single line, got %q", gotStderr)
			}
		})
	}
}

func TestRun_WritesDeterministicScaffold(t *testing.T) {
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
		"properties: {}",
		"TODO: Pulumi inputProperties/requiredInputs to OpenAPI translation is not implemented yet.",
		"observed inputProperties=2 (alpha, zeta), requiredInputs=2 (alpha, zeta)",
	}
	output := a
	for _, snippet := range requiredSnippets {
		if !strings.Contains(output, snippet) {
			t.Fatalf("expected scaffold to contain %q, got:\n%s", snippet, output)
		}
	}
}

func TestRun_OutputWriterFailureReturnsExit4(t *testing.T) {
	t.Parallel()

	tempDir := makeWorkspaceTempDir(t)
	schemaPath := filepath.Join(tempDir, "schema.json")
	if err := os.WriteFile(schemaPath, []byte(`{"resources":{"pkg:index:Thing":{"isComponent":true}}}`), 0o600); err != nil {
		t.Fatalf("write schema fixture: %v", err)
	}

	var stderr strings.Builder
	code := run([]string{"--in", schemaPath}, errWriter{}, &stderr)
	if code != exitOutputError {
		t.Fatalf("exit code mismatch: got %d, want %d", code, exitOutputError)
	}
	if !strings.Contains(stderr.String(), "write scaffold output:") {
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
