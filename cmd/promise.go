/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
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

var promiseCmd = &cobra.Command{
	Use:   "promise",
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

var group, kind, version, plural, outputDir string

func init() {
	initCmd.AddCommand(promiseCmd)

	promiseCmd.Flags().StringVarP(&group, "group", "g", "", "The API group for the Promise")
	promiseCmd.Flags().StringVarP(&kind, "kind", "k", "", "The kind to be provided by the Promise")
	promiseCmd.Flags().StringVarP(&version, "version", "v", "v1alpha1", "The group version for the Promise. Defaults to v1alpha1")
	promiseCmd.Flags().StringVarP(&plural, "plural", "p", "", "The plural form of the kind. Defaults to the kind name with an additional 's' at the end.")
	promiseCmd.Flags().StringVarP(&outputDir, "output-dir", "d", ".", "The output directory to write the Promise structure to; defaults to '.'")

	promiseCmd.MarkFlagRequired("group")
	promiseCmd.MarkFlagRequired("kind")
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

	if plural == "" {
		plural = fmt.Sprintf("%ss", strings.ToLower(kind))
	}

	templates := map[string]string{
		"promise.yaml":          "templates/promise/promise.yaml.tpl",
		"example-resource.yaml": "templates/promise/example-resource.yaml.tpl",
		"README.md":             "templates/promise/README.md",
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

		err = os.WriteFile(fmt.Sprintf("%s/%s", outputDir, name), data.Bytes(), 0644)
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
