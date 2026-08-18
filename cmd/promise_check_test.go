package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckReviewFindingsPassesWithEmptyList(t *testing.T) {
	path := writeReviewFile(t, "[]")

	var out bytes.Buffer
	if err := checkReviewFindings(path, &out); err != nil {
		t.Fatalf("expected empty review file to pass, got %v", err)
	}
	if out.String() != "" {
		t.Fatalf("expected no output, got %q", out.String())
	}
}

func TestCheckReviewFindingsPassesWithResolvedHighFinding(t *testing.T) {
	path := writeReviewFile(t, `[
  {
    "severity": "high",
    "dimension": "Pillar: Safety",
    "description": "guardrail reviewed by platform owner",
    "resolved": true
  }
]`)

	var out bytes.Buffer
	if err := checkReviewFindings(path, &out); err != nil {
		t.Fatalf("expected resolved high finding to pass, got %v", err)
	}
}

func TestCheckReviewFindingsAllowsUnresolvedMediumFinding(t *testing.T) {
	path := writeReviewFile(t, `[
  {
    "severity": "medium",
    "dimension": "Pillar: Efficiency",
    "description": "drift detection could be stronger",
    "resolved": false
  }
]`)

	var out bytes.Buffer
	if err := checkReviewFindings(path, &out); err != nil {
		t.Fatalf("expected unresolved medium finding to pass, got %v", err)
	}
}

func TestCheckReviewFindingsParsesYAML(t *testing.T) {
	path := filepath.Join(t.TempDir(), "review-findings.yaml")
	if err := os.WriteFile(path, []byte(`- severity: low
  dimension: AI-native governance
  description: skill version recorded
  resolved: false
`), 0600); err != nil {
		t.Fatalf("failed to write review file: %v", err)
	}

	var out bytes.Buffer
	if err := checkReviewFindings(path, &out); err != nil {
		t.Fatalf("expected YAML review file to pass, got %v", err)
	}
}

func TestCheckReviewFindingsFailsWithUnresolvedBlockingFinding(t *testing.T) {
	path := writeReviewFile(t, `[
  {
    "severity": "critical",
    "dimension": "Orchestration test",
    "description": "no lifecycle owner",
    "resolved": false
  },
  {
    "severity": "high",
    "dimension": "Pillar: Safety",
    "description": "production guardrails absent",
    "resolved": false
  }
]`)

	var out bytes.Buffer
	err := checkReviewFindings(path, &out)
	if err == nil {
		t.Fatal("expected unresolved critical/high findings to fail")
	}
	if !strings.Contains(out.String(), "no lifecycle owner") {
		t.Fatalf("expected unresolved findings in output, got %q", out.String())
	}
}

func TestCheckReviewFindingsRequiresExistingFile(t *testing.T) {
	var out bytes.Buffer
	err := checkReviewFindings(filepath.Join(t.TempDir(), "missing.yaml"), &out)
	if err == nil {
		t.Fatal("expected missing review file to fail")
	}
}

func TestCheckReviewFindingsRejectsInvalidSeverity(t *testing.T) {
	path := writeReviewFile(t, `[
  {
    "severity": "blocker",
    "dimension": "Pillar: Safety",
    "description": "bad severity",
    "resolved": false
  }
]`)

	var out bytes.Buffer
	err := checkReviewFindings(path, &out)
	if err == nil {
		t.Fatal("expected invalid severity to fail")
	}
}

// Regression test for review feedback: readReviewFindings only decoded the
// first JSON value in the file, so a findings array followed by trailing
// content (a second JSON value, or garbage) was silently ignored - a
// concatenated file could hide unresolved critical/high findings after the
// first document and still pass the gate.
func TestCheckReviewFindingsRejectsTrailingJSONContent(t *testing.T) {
	path := writeReviewFile(t, `[] {"unexpected": "trailing value"}`)

	var out bytes.Buffer
	err := checkReviewFindings(path, &out)
	if err == nil {
		t.Fatal("expected trailing JSON content after the findings array to fail")
	}
}

// Same regression, YAML form: a second YAML document after the findings
// array must not be silently dropped.
func TestCheckReviewFindingsRejectsTrailingYAMLDocument(t *testing.T) {
	path := filepath.Join(t.TempDir(), "review-findings.yaml")
	if err := os.WriteFile(path, []byte(`- severity: low
  dimension: AI-native governance
  description: skill version recorded
  resolved: false
---
unexpected: second document
`), 0600); err != nil {
		t.Fatalf("failed to write review file: %v", err)
	}

	var out bytes.Buffer
	err := checkReviewFindings(path, &out)
	if err == nil {
		t.Fatal("expected a trailing YAML document after the findings array to fail")
	}
}

// Regression test for review feedback: a review file containing a bare
// `null` document decodes into a nil slice with no error, which used to be
// treated the same as a clean, empty findings list - letting a malformed or
// incompletely generated artifact pass the gate silently.
func TestCheckReviewFindingsRejectsNullDocument(t *testing.T) {
	path := writeReviewFile(t, `null`)

	var out bytes.Buffer
	err := checkReviewFindings(path, &out)
	if err == nil {
		t.Fatal("expected a null review file to fail, not be treated as an empty findings list")
	}
}

func TestCheckReviewFindingsRequiresResolvedField(t *testing.T) {
	path := writeReviewFile(t, `[
  {
    "severity": "low",
    "dimension": "AI-native governance",
    "description": "missing resolved field"
  }
]`)

	var out bytes.Buffer
	err := checkReviewFindings(path, &out)
	if err == nil {
		t.Fatal("expected missing resolved field to fail")
	}
}

func writeReviewFile(t *testing.T, contents string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "review-findings.json")
	if err := os.WriteFile(path, []byte(contents), 0600); err != nil {
		t.Fatalf("failed to write review file: %v", err)
	}
	return path
}
