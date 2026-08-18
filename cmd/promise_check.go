package cmd

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

var (
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
	Short: "Check promise Level-1 structural gates",
	Long: "If --promise-dir and/or --example-dir exist, runs Level-1 deterministic gates against the " +
		"generated Promise/example files: valid CRD, namespaced scope, example validates against the CRD " +
		"schema, pipeline images set, delete workflow well-formed.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runPromiseCheck(promiseDir, exampleDir, cmd.ErrOrStderr())
	},
}

func init() {
	rootCmd.AddCommand(promiseCmd)
	promiseCmd.AddCommand(promiseCheckCmd)
	promiseCheckCmd.Flags().StringVar(&promiseDir, "promise-dir", "promises", "directory of generated Promise YAML files (Level-1 gate, skipped if absent)")
	promiseCheckCmd.Flags().StringVar(&exampleDir, "example-dir", "resource-requests", "directory of example/Resource Request YAML files to validate against the Promise CRDs (Level-1 gate, skipped if absent)")
}

func runPromiseCheck(promiseDir, exampleDir string, out io.Writer) error {
	level1Errs := runLevelOneGates(promiseDir, exampleDir, out)
	if len(level1Errs) == 0 {
		return nil
	}

	fmt.Fprintln(out, "Level-1 gate failures:")
	for _, e := range level1Errs {
		fmt.Fprintf(out, "- %s\n", e)
	}
	return fmt.Errorf("%d Level-1 gate failure(s)", len(level1Errs))
}
