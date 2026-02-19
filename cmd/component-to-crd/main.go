package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/pulumi/component-to-crd/internal/emit"
	"github.com/pulumi/component-to-crd/internal/schema"
	selectcomponent "github.com/pulumi/component-to-crd/internal/select"
	"github.com/pulumi/component-to-crd/internal/translate"
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
	if err != nil {
		printError(stderr, err)
		return exitUserError
	}

	doc, err := schema.Load(cfg.inPath)
	if err != nil {
		printError(stderr, err)
		return exitUserError
	}

	tokens := selectcomponent.DiscoverComponentTokens(doc)
	selected, err := selectcomponent.SelectComponent(tokens, cfg.component)
	if err != nil {
		printError(stderr, err)
		return exitUserError
	}

	resource, ok := doc.Resources[selected]
	if !ok {
		printError(stderr, fmt.Errorf("selected component %q missing from schema resources", selected))
		return exitUserError
	}
	if err := schema.ValidateForTranslationComponent(doc, selected); err != nil {
		printError(stderr, err)
		return exitUserError
	}

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
		printError(stderr, err)
		return exitUserError
	}

	translatedSpec, skippedPaths, err := translate.InputPropertiesToOpenAPI(doc, selected, resource)
	for _, issue := range skippedPaths {
		printWarning(stderr, issue)
	}
	if err != nil {
		var unsupportedErr *translate.UnsupportedError
		if errors.As(err, &unsupportedErr) && !unsupportedErr.Skippable {
			printError(stderr, err)
			return exitUnsupported
		}
		printError(stderr, err)
		return exitUserError
	}

	crdYAML, err := emit.RenderCRDYAML(identity, translatedSpec)
	if err != nil {
		printError(stderr, fmt.Errorf("serialize CRD output: %w", err))
		return exitOutputError
	}
	if _, err := stdout.Write(crdYAML); err != nil {
		printError(stderr, fmt.Errorf("write CRD output: %w", err))
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
}

func parseArgs(args []string) (config, error) {
	var cfg config

	flagSet := flag.NewFlagSet("component-to-crd", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)
	flagSet.StringVar(&cfg.inPath, "in", "", "Path or URL to Pulumi schema JSON file")
	flagSet.StringVar(&cfg.component, "component", "", "Component token")
	flagSet.StringVar(&cfg.group, "group", "", "CRD API group")
	flagSet.StringVar(&cfg.version, "version", "", "CRD API version")
	flagSet.StringVar(&cfg.kind, "kind", "", "CRD kind")
	flagSet.StringVar(&cfg.plural, "plural", "", "CRD plural resource name")
	flagSet.StringVar(&cfg.singular, "singular", "", "CRD singular resource name")

	if err := flagSet.Parse(args); err != nil {
		return config{}, fmt.Errorf("invalid flags: %w", err)
	}

	if cfg.inPath == "" {
		return config{}, errors.New("missing required flag: --in")
	}

	if flagSet.NArg() > 0 {
		return config{}, fmt.Errorf("unexpected positional arguments: %v", flagSet.Args())
	}

	return cfg, nil
}

func printError(stderr io.Writer, err error) {
	fmt.Fprintf(stderr, "error: %v\n", err)
}

func printWarning(stderr io.Writer, issue translate.SkippedPathIssue) {
	fmt.Fprintf(stderr, "warn: component=%q path=%q reason=%q\n", issue.Component, issue.Path, issue.Reason)
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
	fmt.Fprintln(stdout, "Usage: component-to-crd --in <path-or-url> [--component <token>] [--group <group>] [--version <version>] [--kind <kind>] [--plural <plural>] [--singular <singular>]")
	fmt.Fprintln(stdout, "")
	fmt.Fprintln(stdout, "Converts a Pulumi component schema into a Kubernetes CRD YAML written to stdout.")
	fmt.Fprintln(stdout, "")
	fmt.Fprintln(stdout, "Flags:")
	fmt.Fprintln(stdout, "  --in string")
	fmt.Fprintln(stdout, "        Path or URL to Pulumi schema JSON file (required unless --help is used)")
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
	fmt.Fprintln(stdout, "")
	fmt.Fprintln(stdout, "Translation behavior:")
	fmt.Fprintln(stdout, "  - Untranslatable schema field paths are skipped instead of failing the whole command.")
	fmt.Fprintln(stdout, "  - Skipped-path details are reported to stderr as parseable lines:")
	fmt.Fprintln(stdout, "      warn: component=\"...\" path=\"...\" reason=\"...\"")
	fmt.Fprintln(stdout, "  - Known construct classes that may be skipped include composition keywords")
	fmt.Fprintln(stdout, "    (oneOf, anyOf, allOf), unsupported schema keywords, and unsupported shapes")
	fmt.Fprintln(stdout, "    under a property path, plus unresolved refs outside supported local handling.")
}
