package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/syntasso/pulumi-component-to-crd/internal/emit"
	"github.com/syntasso/pulumi-component-to-crd/internal/schema"
	selectcomponent "github.com/syntasso/pulumi-component-to-crd/internal/select"
	"github.com/syntasso/pulumi-component-to-crd/internal/translate"
)

const (
	exitSuccess     = 0
	exitUserError   = 2
	exitUnsupported = 3
	exitOutputError = 4
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout io.Writer, stderr io.Writer) int {
	if helpRequested(args) {
		printHelp(stdout)
		return exitSuccess
	}

	cfg, err := parseArgs(args)
	out := newRunOutput(stdout, stderr, cfg.verbose)
	if err != nil {
		out.printError(err)
		return exitUserError
	}

	out.printInfo("loading schema")
	doc, err := schema.Load(cfg.inPath)
	if err != nil {
		out.printError(err)
		return exitUserError
	}

	out.printInfo("selecting component")
	tokens := selectcomponent.DiscoverComponentTokens(doc)
	selected, err := selectcomponent.SelectComponent(tokens, cfg.component)
	if err != nil {
		out.printError(err)
		return exitUserError
	}

	resource, ok := doc.Resources[selected]
	if !ok {
		out.printError(fmt.Errorf("selected component %q missing from schema resources", selected))
		return exitUserError
	}
	out.printInfo("preflight validation")
	if err := schema.ValidateForTranslationComponent(doc, selected); err != nil {
		out.printError(err)
		return exitUserError
	}

	out.printInfo("deriving identity")
	identity := emit.DeriveIdentityDefaults(doc, selected)
	if cfg.group != "" {
		identity.Group = cfg.group
	}
	if cfg.version != "" {
		identity.Version = cfg.version
	}
	if cfg.kind != "" {
		identity.Kind = cfg.kind
	}
	if cfg.plural != "" {
		identity.Plural = cfg.plural
	}
	if cfg.singular != "" {
		identity.Singular = cfg.singular
	}
	if err := identity.Validate(); err != nil {
		out.printError(err)
		return exitUserError
	}

	out.printInfo("translating schema")
	translatedSpec, skippedPaths, err := translate.InputPropertiesToOpenAPI(doc, selected, resource)
	for _, issue := range skippedPaths {
		out.printWarning(issue)
	}
	if err != nil {
		var unsupportedErr *translate.UnsupportedError
		if errors.As(err, &unsupportedErr) && !unsupportedErr.Skippable {
			out.printError(err)
			return exitUnsupported
		}
		out.printError(err)
		return exitUserError
	}

	out.printInfo("rendering CRD")
	crdYAML, err := emit.RenderCRDYAML(identity, translatedSpec)
	if err != nil {
		out.printError(fmt.Errorf("serialize CRD output: %w", err))
		return exitOutputError
	}
	if _, err := stdout.Write(crdYAML); err != nil {
		out.printError(fmt.Errorf("write CRD output: %w", err))
		return exitOutputError
	}
	return exitSuccess
}

type config struct {
	inPath    string
	component string
	group     string
	version   string
	kind      string
	plural    string
	singular  string
	verbose   bool
}

func parseArgs(args []string) (config, error) {
	var cfg config

	flagSet := flag.NewFlagSet("pulumi-component-to-crd", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)
	flagSet.StringVar(&cfg.inPath, "in", "", "Path or URL to Pulumi schema JSON file")
	flagSet.StringVar(&cfg.component, "component", "", "Component token")
	flagSet.StringVar(&cfg.group, "group", "", "CRD API group")
	flagSet.StringVar(&cfg.version, "version", "", "CRD API version")
	flagSet.StringVar(&cfg.kind, "kind", "", "CRD kind")
	flagSet.StringVar(&cfg.plural, "plural", "", "CRD plural resource name")
	flagSet.StringVar(&cfg.singular, "singular", "", "CRD singular resource name")
	flagSet.BoolVar(&cfg.verbose, "verbose", false, "Enable verbose diagnostics to stderr")

	if err := flagSet.Parse(args); err != nil {
		return cfg, fmt.Errorf("invalid flags: %w", err)
	}

	if cfg.inPath == "" {
		return config{}, errors.New("missing required flag: --in")
	}

	if flagSet.NArg() > 0 {
		return config{}, fmt.Errorf("unexpected positional arguments: %v", flagSet.Args())
	}

	return cfg, nil
}

type runOutput struct {
	stdout  io.Writer
	stderr  io.Writer
	verbose bool
}

func newRunOutput(stdout io.Writer, stderr io.Writer, verbose bool) runOutput {
	return runOutput{
		stdout:  stdout,
		stderr:  stderr,
		verbose: verbose,
	}
}

func (o runOutput) printInfo(message string) {
	if !o.verbose {
		return
	}
	fmt.Fprintf(o.stderr, "info: %s\n", message)
}

func (o runOutput) printError(err error) {
	fmt.Fprintf(o.stderr, "error: %v\n", err)
}

func (o runOutput) printWarning(issue translate.SkippedPathIssue) {
	if !o.verbose {
		return
	}
	fmt.Fprintf(o.stderr, "warn: component=%q path=%q reason=%q\n", issue.Component, issue.Path, issue.Reason)
}

func helpRequested(args []string) bool {
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			return true
		}
	}
	return false
}

func printHelp(stdout io.Writer) {
	fmt.Fprintln(stdout, "Usage: pulumi-component-to-crd --in <path-or-url> [--component <token>] [--group <group>] [--version <version>] [--kind <kind>] [--plural <plural>] [--singular <singular>] [--verbose]")
	fmt.Fprintln(stdout, "")
	fmt.Fprintln(stdout, "Converts a Pulumi component schema into a Kubernetes CRD YAML written to stdout.")
	fmt.Fprintln(stdout, "")
	fmt.Fprintln(stdout, "Flags:")
	fmt.Fprintln(stdout, "  --in string")
	fmt.Fprintln(stdout, "        Path or URL to Pulumi schema JSON file (required unless --help is used)")
	fmt.Fprintln(stdout, "        Relative paths are resolved from $PWD first, then from the process working directory")
	fmt.Fprintln(stdout, "  --component string")
	fmt.Fprintln(stdout, "        Component token")
	fmt.Fprintln(stdout, "  --group string")
	fmt.Fprintln(stdout, "        CRD API group")
	fmt.Fprintln(stdout, "  --version string")
	fmt.Fprintln(stdout, "        CRD API version")
	fmt.Fprintln(stdout, "  --kind string")
	fmt.Fprintln(stdout, "        CRD kind")
	fmt.Fprintln(stdout, "  --plural string")
	fmt.Fprintln(stdout, "        CRD plural resource name")
	fmt.Fprintln(stdout, "  --singular string")
	fmt.Fprintln(stdout, "        CRD singular resource name")
	fmt.Fprintln(stdout, "  --verbose")
	fmt.Fprintln(stdout, "        Emit stage logs and diagnostics to stderr")
	fmt.Fprintln(stdout, "")
	fmt.Fprintln(stdout, "Translation behavior:")
	fmt.Fprintln(stdout, "  - Untranslatable schema field paths are skipped instead of failing the whole command.")
	fmt.Fprintln(stdout, "  - Errors are reported to stderr as parseable lines:")
	fmt.Fprintln(stdout, "      error: ...")
	fmt.Fprintln(stdout, "  - With --verbose, stage and skipped-path details are also reported on stderr as parseable lines:")
	fmt.Fprintln(stdout, "      warn: component=\"...\" path=\"...\" reason=\"...\"")
	fmt.Fprintln(stdout, "  - Known construct classes that may be skipped include composition keywords")
	fmt.Fprintln(stdout, "    (oneOf, anyOf, allOf), unsupported schema keywords, and unsupported shapes")
	fmt.Fprintln(stdout, "    under a property path, plus unresolved refs outside supported local handling.")
}
