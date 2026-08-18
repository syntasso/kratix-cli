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
	Short: "Validate a Promise's CRD, workflows, and examples",
	Long: "Validate the generated Promise/example files in --promise-dir and --example-dir: the CRD is a " +
		"valid, Namespaced-scope CustomResourceDefinition, every workflow pipeline container has an image " +
		"set, the delete workflow is well-formed, and every example/Resource Request validates against its " +
		"Promise's CRD schema (including CEL rules and schema defaults). Either directory may be absent, in " +
		"which case its checks are skipped.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runPromiseCheck(promiseDir, exampleDir, cmd.ErrOrStderr())
	},
}

func init() {
	rootCmd.AddCommand(promiseCmd)
	promiseCmd.AddCommand(promiseCheckCmd)
	promiseCheckCmd.Flags().StringVar(&promiseDir, "promise-dir", "promises", "directory of generated Promise YAML files, skipped if absent")
	promiseCheckCmd.Flags().StringVar(&exampleDir, "example-dir", "resource-requests", "directory of example/Resource Request YAML files to validate against the Promise CRDs, skipped if absent")
}

func runPromiseCheck(promiseDir, exampleDir string, out io.Writer) error {
	errs := runStructuralChecks(promiseDir, exampleDir, out)
	if len(errs) == 0 {
		return nil
	}

	fmt.Fprintln(out, "check failures:")
	for _, e := range errs {
		fmt.Fprintf(out, "- %s\n", e)
	}
	return fmt.Errorf("%d check failure(s)", len(errs))
}
