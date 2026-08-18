package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type reviewFinding struct {
	Severity    string `json:"severity" yaml:"severity"`
	Dimension   string `json:"dimension" yaml:"dimension"`
	Description string `json:"description" yaml:"description"`
	Resolved    *bool  `json:"resolved" yaml:"resolved"`
}

var (
	reviewFile string
	promiseDir string
	exampleDir string
)

var promiseCmd = &cobra.Command{
	Use:   "promise",
	Short: "Work with Kratix Promises",
}

var promiseCheckCmd = &cobra.Command{
	Use:   "check",
	Args:  cobra.NoArgs,
	Short: "Check promise review findings and Level-1 structural gates",
	Long: "Validate a review-findings artifact and fail if unresolved critical or high findings remain. " +
		"If --promise-dir and/or --example-dir exist, also runs Level-1 deterministic gates against the " +
		"generated Promise/example files: valid CRD, namespaced scope, example validates against the CRD " +
		"schema, pipeline images set, delete workflow well-formed.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runPromiseCheck(reviewFile, promiseDir, exampleDir, cmd.ErrOrStderr())
	},
}

func init() {
	rootCmd.AddCommand(promiseCmd)
	promiseCmd.AddCommand(promiseCheckCmd)
	promiseCheckCmd.Flags().StringVar(&reviewFile, "review-file", "review-findings.yaml", "review findings YAML or JSON file")
	promiseCheckCmd.Flags().StringVar(&promiseDir, "promise-dir", "promises", "directory of generated Promise YAML files (Level-1 gate, skipped if absent)")
	promiseCheckCmd.Flags().StringVar(&exampleDir, "example-dir", "resource-requests", "directory of example/Resource Request YAML files to validate against the Promise CRDs (Level-1 gate, skipped if absent)")
}

func runPromiseCheck(reviewFile, promiseDir, exampleDir string, out io.Writer) error {
	reviewErr := checkReviewFindings(reviewFile, out)

	level1Errs := runLevelOneGates(promiseDir, exampleDir, out)
	if len(level1Errs) > 0 {
		fmt.Fprintln(out, "Level-1 gate failures:")
		for _, e := range level1Errs {
			fmt.Fprintf(out, "- %s\n", e)
		}
	}

	switch {
	case reviewErr != nil && len(level1Errs) > 0:
		return fmt.Errorf("%w; and %d Level-1 gate failure(s)", reviewErr, len(level1Errs))
	case reviewErr != nil:
		return reviewErr
	case len(level1Errs) > 0:
		return fmt.Errorf("%d Level-1 gate failure(s)", len(level1Errs))
	default:
		return nil
	}
}

func checkReviewFindings(path string, out io.Writer) error {
	findings, err := readReviewFindings(path)
	if err != nil {
		return err
	}

	var unresolved []reviewFinding
	for _, finding := range findings {
		if blocksApply(finding) {
			unresolved = append(unresolved, finding)
		}
	}

	if len(unresolved) == 0 {
		return nil
	}

	fmt.Fprintln(out, "unresolved blocking review findings:")
	for _, finding := range unresolved {
		fmt.Fprintf(out, "- [%s] %s: %s\n", finding.Severity, finding.Dimension, finding.Description)
	}
	return fmt.Errorf("%d unresolved critical/high review finding(s)", len(unresolved))
}

func readReviewFindings(path string) ([]reviewFinding, error) {
	if path == "" {
		return nil, errors.New("review file path is required")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read review file %q: %w", path, err)
	}

	var findings []reviewFinding
	switch filepath.Ext(path) {
	case ".json":
		dec := json.NewDecoder(bytes.NewReader(data))
		dec.DisallowUnknownFields()
		if err := dec.Decode(&findings); err != nil {
			return nil, fmt.Errorf("failed to parse review file %q: %w", path, err)
		}
	default:
		dec := yaml.NewDecoder(bytes.NewReader(data))
		dec.KnownFields(true)
		if err := dec.Decode(&findings); err != nil {
			return nil, fmt.Errorf("failed to parse review file %q: %w", path, err)
		}
	}

	for i, finding := range findings {
		if err := validateReviewFinding(finding); err != nil {
			return nil, fmt.Errorf("invalid finding %d: %w", i, err)
		}
	}

	return findings, nil
}

func validateReviewFinding(finding reviewFinding) error {
	switch finding.Severity {
	case "critical", "high", "medium", "low":
	default:
		return fmt.Errorf("severity must be one of critical, high, medium, low")
	}
	if finding.Dimension == "" {
		return errors.New("dimension is required")
	}
	if finding.Description == "" {
		return errors.New("description is required")
	}
	if finding.Resolved == nil {
		return errors.New("resolved is required")
	}
	return nil
}

func blocksApply(finding reviewFinding) bool {
	if *finding.Resolved {
		return false
	}
	return finding.Severity == "critical" || finding.Severity == "high"
}
