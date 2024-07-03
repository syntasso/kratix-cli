package cmd

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"strings"
	"text/template"

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

var split bool

const (
	promiseFileName      = "promise.yaml"
	dependenciesFileName = "dependencies.yaml"
	apiFileName          = "api.yaml"
	resourceFileName     = "example-resource.yaml"
)

func init() {
	initCmd.AddCommand(initPromiseCmd)

	initPromiseCmd.Flags().BoolVar(&split, "split", false, "Split promise.yaml file into api.yaml and dependencies.yaml")
}

type promiseTemplateValues struct {
	Name     string
	Group    string
	Kind     string
	Version  string
	Plural   string
	Singular string
}

func InitPromise(cmd *cobra.Command, args []string) error {
	promiseName := args[0]

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

	for name, tmpl := range templates {
		t, err := template.ParseFS(promiseTemplates, tmpl)
		if err != nil {
			return err
		}
		data := bytes.NewBuffer([]byte{})
		err = t.Execute(data, promiseTemplateValues{
			Name:     promiseName,
			Group:    group,
			Kind:     kind,
			Version:  version,
			Plural:   plural,
			Singular: strings.ToLower(kind),
		})
		if err != nil {
			return err
		}

		err = os.WriteFile(fmt.Sprintf("%s/%s", outputDir, name), data.Bytes(), filePerm)
		if err != nil {
			return err
		}
	}

	dirName := "current"
	if outputDir != "." {
		dirName = outputDir
	}
	fmt.Printf("%s promise bootstrapped in the %s directory\n", promiseName, dirName)
	return nil
}
