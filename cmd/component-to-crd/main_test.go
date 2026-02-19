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
	mixedSkippableSchemaPath := filepath.Join(tempDir, "mixed-skippable.json")
	allSkippableSchemaPath := filepath.Join(tempDir, "all-skippable.json")
	hardUnsupportedSchemaPath := filepath.Join(tempDir, "hard-unsupported.json")
	malformedSchemaPath := filepath.Join(tempDir, "malformed.json")
	unreachableMalformedSchemaPath := filepath.Join(tempDir, "unreachable-malformed.json")
	selectedMalformedAndUnsupportedSchemaPath := filepath.Join(tempDir, "selected-malformed-and-unsupported.json")
	resourceRefSchemaPath := filepath.Join(tempDir, "resource-ref.json")
	nonLocalRefSchemaPath := filepath.Join(tempDir, "non-local-ref.json")

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

	if err := os.WriteFile(mixedSkippableSchemaPath, []byte(`{"resources":{"pkg:index:Thing":{"isComponent":true,"inputProperties":{"name":{"type":"string"},"badTop":{"oneOf":[{"type":"string"},{"type":"number"}]},"settings":{"type":"object","properties":{"enabled":{"type":"boolean"},"badNested":{"anyOf":[{"type":"string"},{"type":"number"}]}},"required":["enabled","badNested"]}},"requiredInputs":["name","badTop"]}}}`), 0o600); err != nil {
		t.Fatalf("write mixed skippable fixture: %v", err)
	}

	if err := os.WriteFile(allSkippableSchemaPath, []byte(`{"resources":{"pkg:index:Thing":{"isComponent":true,"inputProperties":{"value":{"oneOf":[{"type":"string"},{"type":"number"}]}}}}}`), 0o600); err != nil {
		t.Fatalf("write all skippable fixture: %v", err)
	}

	if err := os.WriteFile(hardUnsupportedSchemaPath, []byte(`{"resources":{"pkg:index:Thing":{"isComponent":true,"inputProperties":{"cycle":{"$ref":"#/types/pkg:index:A"}}}},"types":{"pkg:index:A":{"$ref":"#/types/pkg:index:B"},"pkg:index:B":{"$ref":"#/types/pkg:index:A"}}}`), 0o600); err != nil {
		t.Fatalf("write hard unsupported fixture: %v", err)
	}

	if err := os.WriteFile(malformedSchemaPath, []byte(`{"resources":{"pkg:index:Thing":{"isComponent":true,"inputProperties":{"value":{"$ref":"#/types/pkg:index:Missing"}}}}}`), 0o600); err != nil {
		t.Fatalf("write malformed schema fixture: %v", err)
	}

	if err := os.WriteFile(unreachableMalformedSchemaPath, []byte(`{"resources":{"pkg:index:Zulu":{"isComponent":true,"inputProperties":{"bad":{"$ref":"#/types/pkg:index:Missing"}}},"pkg:index:Alpha":{"isComponent":true,"inputProperties":{"value":{"type":"string"}}}}}`), 0o600); err != nil {
		t.Fatalf("write unreachable malformed fixture: %v", err)
	}

	if err := os.WriteFile(selectedMalformedAndUnsupportedSchemaPath, []byte(`{"resources":{"pkg:index:Thing":{"isComponent":true,"inputProperties":{"bad":{"$ref":"#/types/pkg:index:Missing"},"value":{"oneOf":[{"type":"string"},{"type":"number"}]}}}}}`), 0o600); err != nil {
		t.Fatalf("write selected malformed + unsupported fixture: %v", err)
	}

	if err := os.WriteFile(resourceRefSchemaPath, []byte(`{"resources":{"eks:index:Addon":{"isComponent":true,"inputProperties":{"cluster":{"$ref":"#/resources/eks:index:Cluster"},"addonName":{"type":"string"}},"requiredInputs":["cluster","addonName"]},"eks:index:Cluster":{"isComponent":true,"inputProperties":{"name":{"type":"string"}},"requiredInputs":["name"]}}}`), 0o600); err != nil {
		t.Fatalf("write resource ref fixture: %v", err)
	}

	if err := os.WriteFile(nonLocalRefSchemaPath, []byte(`{"resources":{"eks:index:Cluster":{"isComponent":true,"inputProperties":{"name":{"type":"string"},"accessScope":{"$ref":"/aws/v7.14.0/schema.json#/types/aws:eks%2FAccessScope:AccessScope"}},"requiredInputs":["name","accessScope"]}}}`), 0o600); err != nil {
		t.Fatalf("write non-local ref fixture: %v", err)
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
		wantStdoutNot   []string
		wantStdoutLines []string
		wantStderrParts []string
		wantStderrNot   []string
		wantStderrLines []string
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
			name:         "reachable non-local ref uses fallback schema",
			args:         []string{"--in", nonLocalRefSchemaPath, "--component", "eks:index:Cluster"},
			wantExitCode: exitSuccess,
			wantStdoutParts: []string{
				"accessScope:",
				"x-kubernetes-preserve-unknown-fields: true",
				"required:",
				"- \"accessScope\"",
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
			name:         "mixed translatable and untranslatable fields succeeds in default mode",
			args:         []string{"--in", mixedSkippableSchemaPath},
			wantExitCode: exitSuccess,
			wantStdoutParts: []string{
				`name: "things.pkg.components.platform"`,
				"name:",
				"settings:",
				"enabled:",
			},
			wantStdoutNot: []string{
				"badTop:",
				"badNested:",
				"- \"badTop\"",
				"- \"badNested\"",
			},
		},
		{
			name:         "all top-level untranslatable fields fail with exit 2",
			args:         []string{"--in", allSkippableSchemaPath},
			wantExitCode: exitUserError,
			wantStderrParts: []string{
				"error:",
				`no translatable spec fields remain after skipping unsupported schema paths for component "pkg:index:Thing"`,
			},
			wantStderrNot: []string{
				"warn:",
			},
		},
		{
			name:            "hard non-skippable unsupported construct is exit 3",
			args:            []string{"--in", hardUnsupportedSchemaPath},
			wantExitCode:    exitUnsupported,
			wantStderrParts: []string{"error:", `component "pkg:index:Thing" path "spec.cycle" unsupported construct: cyclic local ref "#/types/pkg:index:A"`},
		},
		{
			name:         "malformed schema returns exit 2 preflight error",
			args:         []string{"--in", malformedSchemaPath},
			wantExitCode: exitUserError,
			wantStderrParts: []string{
				"error:",
				`schema preflight component "pkg:index:Thing"`,
				"schema preflight path",
				`resources.pkg:index:Thing.inputProperties.value`,
				`unresolved local type ref "#/types/pkg:index:Missing"`,
			},
		},
		{
			name:         "component selection error has precedence over scoped preflight",
			args:         []string{"--in", unreachableMalformedSchemaPath, "--component", "pkg:index:Missing"},
			wantExitCode: exitUserError,
			wantStderrParts: []string{
				"error:",
				`component "pkg:index:Missing" not found; available components: pkg:index:Alpha, pkg:index:Zulu`,
			},
			wantStderrNot: []string{
				"schema preflight path",
			},
		},
		{
			name:         "unreachable malformed ref does not block selected component",
			args:         []string{"--in", unreachableMalformedSchemaPath, "--component", "pkg:index:Alpha"},
			wantExitCode: exitSuccess,
			wantStdoutParts: []string{
				"apiVersion: apiextensions.k8s.io/v1",
				"kind: CustomResourceDefinition",
			},
			wantStdoutNot: []string{
				"schema preflight path",
			},
		},
		{
			name:         "reachable preflight error has precedence over unsupported construct",
			args:         []string{"--in", selectedMalformedAndUnsupportedSchemaPath},
			wantExitCode: exitUserError,
			wantStderrParts: []string{
				"error:",
				`schema preflight component "pkg:index:Thing"`,
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
			for _, part := range tt.wantStdoutNot {
				if strings.Contains(gotStdout, part) {
					t.Fatalf("stdout unexpectedly contains %q in %q", part, gotStdout)
				}
			}
			if len(tt.wantStdoutLines) > 0 {
				gotLines := splitNonEmptyLines(gotStdout)
				if len(gotLines) < len(tt.wantStdoutLines) {
					t.Fatalf("stdout has %d lines, want at least %d; stdout = %q", len(gotLines), len(tt.wantStdoutLines), gotStdout)
				}
				for i, wantLine := range tt.wantStdoutLines {
					if gotLines[i] != wantLine {
						t.Fatalf("stdout line %d mismatch\n got: %q\nwant: %q\nfull stdout: %q", i+1, gotLines[i], wantLine, gotStdout)
					}
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
			if len(tt.wantStderrLines) > 0 {
				gotLines := splitNonEmptyLines(gotStderr)
				if len(gotLines) < len(tt.wantStderrLines) {
					t.Fatalf("stderr has %d lines, want at least %d; stderr = %q", len(gotLines), len(tt.wantStderrLines), gotStderr)
				}
				for i, wantLine := range tt.wantStderrLines {
					if gotLines[i] != wantLine {
						t.Fatalf("stderr line %d mismatch\n got: %q\nwant: %q\nfull stderr: %q", i+1, gotLines[i], wantLine, gotStderr)
					}
				}
			}

			for _, line := range splitNonEmptyLines(gotStderr) {
				if strings.HasPrefix(line, "error:") || strings.HasPrefix(line, "warn:") || strings.HasPrefix(line, "info:") {
					if strings.Contains(line, "\n") {
						t.Fatalf("stderr line should be single-line parseable, got %q", line)
					}
				}
			}
		})
	}
}

func TestRun_VerboseSuccessAddsStderrDiagnostics(t *testing.T) {
	t.Parallel()

	tempDir := makeWorkspaceTempDir(t)
	schemaPath := filepath.Join(tempDir, "mixed-skippable.json")
	if err := os.WriteFile(schemaPath, []byte(`{"resources":{"pkg:index:Thing":{"isComponent":true,"inputProperties":{"name":{"type":"string"},"badTop":{"oneOf":[{"type":"string"},{"type":"number"}]},"settings":{"type":"object","properties":{"enabled":{"type":"boolean"},"badNested":{"anyOf":[{"type":"string"},{"type":"number"}]}}}},"requiredInputs":["name"]}}}`), 0o600); err != nil {
		t.Fatalf("write mixed skippable fixture: %v", err)
	}

	var defaultStdout strings.Builder
	var defaultStderr strings.Builder
	defaultExit := run([]string{"--in", schemaPath}, &defaultStdout, &defaultStderr)
	if defaultExit != exitSuccess {
		t.Fatalf("default run exit code mismatch: got %d want %d", defaultExit, exitSuccess)
	}
	if defaultStderr.String() != "" {
		t.Fatalf("expected empty default stderr, got %q", defaultStderr.String())
	}

	var verboseStdout strings.Builder
	var verboseStderr strings.Builder
	verboseExit := run([]string{"--in", schemaPath, "--verbose"}, &verboseStdout, &verboseStderr)
	if verboseExit != exitSuccess {
		t.Fatalf("verbose run exit code mismatch: got %d want %d", verboseExit, exitSuccess)
	}
	if defaultStdout.String() != verboseStdout.String() {
		t.Fatalf("stdout mismatch between default and verbose runs\ndefault:\n%s\nverbose:\n%s", defaultStdout.String(), verboseStdout.String())
	}

	verboseErr := verboseStderr.String()
	requiredStderrParts := []string{
		"info: loading schema",
		"info: selecting component",
		"info: preflight validation",
		"info: translating schema",
		"info: rendering CRD",
		`warn: component="pkg:index:Thing" path="spec.badTop" reason="keyword \"oneOf\""`,
		`warn: component="pkg:index:Thing" path="spec.settings.badNested" reason="keyword \"anyOf\""`,
	}
	for _, part := range requiredStderrParts {
		if !strings.Contains(verboseErr, part) {
			t.Fatalf("verbose stderr missing %q in %q", part, verboseErr)
		}
	}
}

func TestRun_VerboseFailureMirrorsErrorToStderr(t *testing.T) {
	t.Parallel()

	tempDir := makeWorkspaceTempDir(t)
	schemaPath := filepath.Join(tempDir, "malformed.json")
	if err := os.WriteFile(schemaPath, []byte(`{"resources":{"pkg:index:Thing":{"isComponent":true,"inputProperties":{"value":{"$ref":"#/types/pkg:index:Missing"}}}}}`), 0o600); err != nil {
		t.Fatalf("write malformed schema fixture: %v", err)
	}

	var defaultStdout strings.Builder
	var defaultStderr strings.Builder
	defaultExit := run([]string{"--in", schemaPath}, &defaultStdout, &defaultStderr)
	if defaultExit != exitUserError {
		t.Fatalf("default run exit code mismatch: got %d want %d", defaultExit, exitUserError)
	}
	if defaultStdout.String() != "" {
		t.Fatalf("expected empty default stdout, got %q", defaultStdout.String())
	}
	if !strings.Contains(defaultStderr.String(), "error:") {
		t.Fatalf("expected parseable error on default stderr, got %q", defaultStderr.String())
	}

	var verboseStdout strings.Builder
	var verboseStderr strings.Builder
	verboseExit := run([]string{"--in", schemaPath, "--verbose"}, &verboseStdout, &verboseStderr)
	if verboseExit != exitUserError {
		t.Fatalf("verbose run exit code mismatch: got %d want %d", verboseExit, exitUserError)
	}
	if verboseStdout.String() != "" {
		t.Fatalf("expected empty verbose stdout, got %q", verboseStdout.String())
	}
	if !strings.Contains(verboseStderr.String(), "error:") {
		t.Fatalf("expected parseable error on verbose stderr, got %q", verboseStderr.String())
	}
	if !strings.Contains(verboseStderr.String(), "info: preflight validation") {
		t.Fatalf("expected verbose stage log on stderr, got %q", verboseStderr.String())
	}
}

func TestRun_HelpIncludesSkipGuidance(t *testing.T) {
	t.Parallel()

	var stdout strings.Builder
	var stderr strings.Builder

	code := run([]string{"--help"}, &stdout, &stderr)
	if code != exitSuccess {
		t.Fatalf("exit code mismatch: got %d, want %d", code, exitSuccess)
	}
	if stderr.String() != "" {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}

	helpText := stdout.String()
	requiredParts := []string{
		"Usage: component-to-crd --in",
		"--verbose",
		"Untranslatable schema field paths are skipped",
		"warn: component=",
		"stderr",
		"oneOf, anyOf, allOf",
		"unsupported schema keywords",
		"unresolved refs outside supported local handling",
	}
	for _, part := range requiredParts {
		if !strings.Contains(helpText, part) {
			t.Fatalf("help output missing %q in %q", part, helpText)
		}
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

func splitNonEmptyLines(value string) []string {
	raw := strings.Split(value, "\n")
	lines := make([]string, 0, len(raw))
	for _, line := range raw {
		if line == "" {
			continue
		}
		lines = append(lines, line)
	}
	return lines
}
