package cmd

import (
	"embed"
	"fmt"
	"github.com/spf13/cobra"
	"strings"
)

//go:embed templates/promise/*
var promiseTemplates embed.FS

var initPromiseCmd = &cobra.Command{
	Use:   "promise PROMISE-NAME --group API-GROUP --kind API-KIND",
	Short: "Initialize a new Promise",
	Long:  `Initialize a new Promise within the current directory, with all the necessary files to get started`,
	Example: `  # initialize a new promise with the api group and provided kind
  kratix init promise postgresql --group syntasso.io --kind database

  # initialize a new promise with the specified version
  kratix init promise postgresql --group syntasso.io --kind database --version v1
`,
	Args: cobra.ExactArgs(1),
	RunE: InitPromise,
}

const (
	promiseFileName      = "promise.yaml"
	dependenciesFileName = "dependencies.yaml"
	apiFileName          = "api.yaml"
	resourceFileName     = "example-resource.yaml"
)

func init() {
	initCmd.AddCommand(initPromiseCmd)

}

type promiseTemplateValues struct {
	Name       string
	Group      string
	Kind       string
	Version    string
	Plural     string
	Singular   string
	SubCommand string
}

func InitPromise(cmd *cobra.Command, args []string) error {
	return templatePromiseFiles(args[0], "promise")
}

func templatePromiseFiles(promiseName, subcommand string) error {
	if version == "" {
		version = "v1alpha1"
	}

	if plural == "" {
		plural = fmt.Sprintf("%ss", strings.ToLower(kind))
	}

	templates := map[string]string{
		resourceFileName: "templates/promise/example-resource.yaml.tpl",
		"README.md":      "templates/promise/README.md",
	}

	if split {
		templates[apiFileName] = "templates/promise/api.yaml.tpl"
		templates[dependenciesFileName] = "templates/promise/dependencies.yaml"
	} else {
		templates[promiseFileName] = "templates/promise/promise.yaml.tpl"
	}

	templateValues := promiseTemplateValues{
		Name:       promiseName,
		Group:      group,
		Kind:       kind,
		Version:    version,
		Plural:     plural,
		Singular:   strings.ToLower(kind),
		SubCommand: subcommand,
	}
	if err := templateFiles(promiseTemplates, outputDir, templates, templateValues); err != nil {
		return err
	}

	dirName := "current"
	if outputDir != "." {
		dirName = outputDir
	}
	fmt.Printf("%s promise bootstrapped in the %s directory\n", promiseName, dirName)
	return nil
}
