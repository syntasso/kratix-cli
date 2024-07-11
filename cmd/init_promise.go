package cmd

import (
	"embed"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
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
	promiseFileName                   = "promise.yaml"
	dependenciesFileName              = "dependencies.yaml"
	apiFileName                       = "api.yaml"
	resourceFileName                  = "example-resource.yaml"
	resourceConfigureWorkflowFileName = "workflows/resource/configure/workflow.yaml"
)

func init() {
	initCmd.AddCommand(initPromiseCmd)

}

type promiseTemplateValues struct {
	Name              string
	Group             string
	Kind              string
	Version           string
	Plural            string
	Singular          string
	SubCommand        string
	ResourceConfigure string
	CRDSchema         string
}

func InitPromise(cmd *cobra.Command, args []string) error {
	promiseName := args[0]

	templateValues := generateTemplateValues(promiseName, "promise", "", "")

	templates := map[string]string{
		resourceFileName: "templates/promise/example-resource.yaml.tpl",
		"README.md":      "templates/promise/README.md.tpl",
	}

	if split {
		templates[apiFileName] = "templates/promise/api.yaml.tpl"
		templates[dependenciesFileName] = "templates/promise/dependencies.yaml"
	} else {
		templates[promiseFileName] = "templates/promise/promise.yaml.tpl"
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

func generateTemplateValues(promiseName, subCommand, resourceConfigure, crdSchema string) promiseTemplateValues {
	if version == "" {
		version = "v1alpha1"
	}

	if plural == "" {
		plural = fmt.Sprintf("%ss", strings.ToLower(kind))
	}

	return promiseTemplateValues{
		Name:              promiseName,
		Group:             group,
		Kind:              kind,
		Version:           version,
		Plural:            plural,
		Singular:          strings.ToLower(kind),
		SubCommand:        subCommand,
		ResourceConfigure: resourceConfigure,
		CRDSchema:         crdSchema,
	}
}
