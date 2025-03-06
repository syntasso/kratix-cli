/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"

	// importan crossplane v1
	xrdv1 "github.com/crossplane/crossplane/apis/apiextensions/v1"
)

// crossplanePromiseCmd represents the crossplanePromise command
var (
	crossplanePromiseCmd = &cobra.Command{
		Use:   "crossplane-promise",
		Short: "A brief description of your command",
		Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
		RunE: InitCrossplanePromise,
	}

	xrdPath string
)

func init() {
	initCmd.AddCommand(crossplanePromiseCmd)
	crossplanePromiseCmd.Flags().StringVarP(&xrdPath, "xrd-path", "x", "", "Path to the XRD file")
	crossplanePromiseCmd.MarkFlagRequired("xrd-path")
}

func InitCrossplanePromise(cmd *cobra.Command, args []string) error {
	promiseName := args[0]
	if plural == "" {
		plural = fmt.Sprintf("%ss", strings.ToLower(kind))
	}

	xrd, err := fetchXRD()
	if err != nil {
		return err
	}

	storedVersionIdx := findXRDStoredVersionIdx(xrd)
	if storedVersionIdx == -1 {
		return fmt.Errorf("no served version found in XRD")
	}

	crd, err := generateCRDFromXRD(xrd.Spec.Versions[storedVersionIdx])
	if err != nil {
		return err
	}

	promise, err := generatePromise(promiseName, nil, crd, nil)
	if err != nil {
		return err
	}
	content, err := yaml.Marshal(promise)
	if err != nil {
		return err
	}

	fmt.Println(string(content))

	exampleResource := &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": fmt.Sprintf("%s/%s", crd.Spec.Group, crd.Spec.Versions[0].Name),
			"kind":       kind,
			"metadata": map[string]string{
				"name":      "example-database",
				"namespace": "default",
			},
		},
	}

	envs := []corev1.EnvVar{
		{
			Name:  "GROUP",
			Value: crd.Spec.Group,
		},
		{
			Name:  "VERSION",
			Value: crd.Spec.Versions[storedVersionIdx].Name,
		},
		{
			Name:  "KIND",
			Value: crd.Spec.Names.Kind,
		},
	}
	pipelines := generateResourceConfigurePipelines("from-api-to-crossplane-claim", "ghcr.io/syntasso/kratix-cli/from-api-to-crossplane-claim:v0.1.0", envs)

	workflowDirectory := filepath.Join("workflows", "resource", "configure")
	filesToWrite, err := getFilesToWrite(promiseName, split, workflowDirectory, nil, crd, pipelines, exampleResource)
	if err != nil {
		return err
	}

	err = writePromiseFiles(outputDir, filesToWrite)
	if err != nil {
		return err
	}

	fmt.Println("Crossplane Promise generated successfully.")
	return nil

}

func generateCRDFromXRD(version xrdv1.CompositeResourceDefinitionVersion) (*apiextensionsv1.CustomResourceDefinition, error) {
	schemaRaw := version.Schema.OpenAPIV3Schema
	schema := &apiextensionsv1.JSONSchemaProps{}
	if err := yaml.Unmarshal(schemaRaw.Raw, schema); err != nil {
		return nil, fmt.Errorf("failed to unmarshal schema: %w", err)
	}

	crd := &apiextensionsv1.CustomResourceDefinition{}
	crd.Spec.Group = group
	crd.Spec.Scope = "Namespaced"
	crd.Spec.Names = apiextensionsv1.CustomResourceDefinitionNames{
		Plural:   plural,
		Singular: strings.ToLower(kind),
		Kind:     kind,
	}
	crd.Name = fmt.Sprintf("%s.%s", crd.Spec.Names.Plural, group)
	crd.Spec.Versions = []apiextensionsv1.CustomResourceDefinitionVersion{
		{
			Name:    version.Name,
			Served:  true,
			Storage: true,
			Schema: &apiextensionsv1.CustomResourceValidation{
				OpenAPIV3Schema: schema,
			},
		},
	}
	crd.APIVersion = "apiextensions.k8s.io/v1"
	crd.Kind = "CustomResourceDefinition"

	return crd, nil
}

func findXRDStoredVersionIdx(crd *xrdv1.CompositeResourceDefinition) int {
	for i, version := range crd.Spec.Versions {
		if version.Served {
			return i
		}
	}
	return -1
}

func fetchXRD() (*xrdv1.CompositeResourceDefinition, error) {
	xrd := &xrdv1.CompositeResourceDefinition{}

	// read xrdPath and unmarshal it into xrd
	contents, err := os.ReadFile(xrdPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", xrdPath, err)
	}

	if err := yaml.Unmarshal(contents, xrd); err != nil {
		return nil, fmt.Errorf("failed to unmarshal file %s: %w", xrdPath, err)
	}

	fmt.Println(xrd)
	return xrd, nil
}
