package cmd

import (
	"embed"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"sigs.k8s.io/yaml"
)

//go:embed templates/promise/*
var promiseTemplates embed.FS

var initPromiseCmd = &cobra.Command{
	Use:   "promise PROMISE-NAME --group PROMISE-API-GROUP --kind PROMISE-API-KIND",
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
	Name                 string
	Group                string
	Kind                 string
	Version              string
	Plural               string
	Singular             string
	SubCommand           string
	ResourceConfigure    string
	PromiseConfigure     string
	CRDSchema            string
	DestinationSelectors string
	ExtraFlags           string
}

func InitPromise(cmd *cobra.Command, args []string) error {
	promiseName := args[0]

	templateValues, err := generateTemplateValues(promiseName, "promise", "", "[]", "[]", "")
	if err != nil {
		return err
	}

	templates := map[string]string{
		resourceFileName: fmt.Sprintf("templates/promise/%s.tpl", resourceFileName),
		"README.md":      "templates/promise/README.md.tpl",
	}

	if split {
		templates[apiFileName] = fmt.Sprintf("templates/promise/%s.tpl", apiFileName)
		templates[dependenciesFileName] = fmt.Sprintf("templates/promise/%s", dependenciesFileName)
	} else {
		templates[promiseFileName] = fmt.Sprintf("templates/promise/%s.tpl", promiseFileName)
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

func generateTemplateValues(promiseName, subCommand, extraFlags, resourceConfigure, promiseConfigure, crdSchema string) (promiseTemplateValues, error) {
	if version == "" {
		version = "v1alpha1"
	}

	if plural == "" {
		plural = fmt.Sprintf("%ss", strings.ToLower(kind))
	}

	if crdSchema == "" {
		schema := &apiextensionsv1.JSONSchemaProps{
			Type:       "object",
			Default:    &apiextensionsv1.JSON{Raw: []byte(`{}`)},
			Properties: map[string]apiextensionsv1.JSONSchemaProps{},
		}

		crdSchemaBytes, err := yaml.Marshal(schema)
		if err != nil {
			return promiseTemplateValues{}, err
		}
		crdSchema = string(crdSchemaBytes)
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
		PromiseConfigure:  promiseConfigure,
		CRDSchema:         crdSchema,
		ExtraFlags:        extraFlags,
	}, nil
}
