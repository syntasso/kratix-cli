/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/syntasso/kratix/api/v1alpha1"
	goyaml "gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/yaml"

	// importan crossplane v1

	xrdv1 "github.com/crossplane/crossplane/apis/apiextensions/v1"
)

var (
	mandatoryAdditionalClaimFields = map[string]apiextensionsv1.JSONSchemaProps{
		"compositeDeletePolicy": {
			Type:    "string",
			Enum:    []apiextensionsv1.JSON{{Raw: []byte(`"Background"`)}, {Raw: []byte(`"Foreground"`)}},
			Default: &apiextensionsv1.JSON{Raw: []byte(`"Background"`)},
		},
		"compositionRef": {
			Type: "object",
			Properties: map[string]apiextensionsv1.JSONSchemaProps{
				"name": {Type: "string"},
			},
			Required: []string{"name"},
		},
		"compositionRevisionRef": {
			Type: "object",
			Properties: map[string]apiextensionsv1.JSONSchemaProps{
				"name": {Type: "string"},
			},
			Required: []string{"name"},
		},
		"compositionRevisionSelector": {
			Type: "object",
			Properties: map[string]apiextensionsv1.JSONSchemaProps{
				"matchLabels": {
					Type:                 "object",
					AdditionalProperties: &apiextensionsv1.JSONSchemaPropsOrBool{Schema: &apiextensionsv1.JSONSchemaProps{Type: "string"}},
				},
			},
			Required: []string{"matchLabels"},
		},
		"compositionSelector": {
			Type: "object",
			Properties: map[string]apiextensionsv1.JSONSchemaProps{
				"matchLabels": {
					Type:                 "object",
					AdditionalProperties: &apiextensionsv1.JSONSchemaPropsOrBool{Schema: &apiextensionsv1.JSONSchemaProps{Type: "string"}},
				},
			},
			Required: []string{"matchLabels"},
		},
		"compositionUpdatePolicy": {
			Type: "string",
			Enum: []apiextensionsv1.JSON{
				{Raw: []byte(`"Automatic"`)},
				{Raw: []byte(`"Manual"`)},
			},
		},
		"publishConnectionDetailsTo": {
			Type: "object",
			Properties: map[string]apiextensionsv1.JSONSchemaProps{
				"configRef": {
					Type: "object",
					Properties: map[string]apiextensionsv1.JSONSchemaProps{
						"name": {Type: "string"},
					},
					Default: &apiextensionsv1.JSON{Raw: []byte(`{"name": "default"}`)},
				},
				"metadata": {
					Type: "object",
					Properties: map[string]apiextensionsv1.JSONSchemaProps{
						"annotations": {
							Type:                 "object",
							AdditionalProperties: &apiextensionsv1.JSONSchemaPropsOrBool{Schema: &apiextensionsv1.JSONSchemaProps{Type: "string"}},
						},
						"labels": {
							Type:                 "object",
							AdditionalProperties: &apiextensionsv1.JSONSchemaPropsOrBool{Schema: &apiextensionsv1.JSONSchemaProps{Type: "string"}},
						},
						"type": {Type: "string"},
					},
				},
				"name": {Type: "string"},
			},
			Required: []string{"name"},
		},
		"resourceRef": {
			Type: "object",
			Properties: map[string]apiextensionsv1.JSONSchemaProps{
				"apiVersion": {Type: "string"},
				"kind":       {Type: "string"},
				"name":       {Type: "string"},
			},
			Required: []string{"apiVersion", "kind", "name"},
		},
		"writeConnectionSecretToRef": {
			Type: "object",
			Properties: map[string]apiextensionsv1.JSONSchemaProps{
				"name": {Type: "string"},
			},
			Required: []string{"name"},
		},
	}

	crossplaneDestinationSelectors = []v1alpha1.PromiseScheduling{{MatchLabels: map[string]string{"crossplane": "enabled"}}}

	// crossplanePromiseCmd represents the crossplanePromise command
	crossplanePromiseCmd = &cobra.Command{
		Use:   "crossplane-promise",
		Short: "A brief description of your command",
		Args:  cobra.ExactArgs(1),
		Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
		RunE: InitCrossplanePromise,
	}

	xrdPath          string
	compositions     string
	skipDependencies bool
)

func init() {
	initCmd.AddCommand(crossplanePromiseCmd)
	crossplanePromiseCmd.Flags().StringVarP(&xrdPath, "xrd", "x", "", "Filepath to the XRD file")
	crossplanePromiseCmd.Flags().StringVarP(&compositions, "compositions", "c", "", "Filepath to the Compositions file. Can contain a single Composition or multiple Compositions.")
	crossplanePromiseCmd.Flags().BoolVarP(&skipDependencies, "skip-dependencies", "s", false, "Skip generating dependencies. For when the XRD and Compositions are already deployed to Crossplane")
	crossplanePromiseCmd.MarkFlagRequired("xrd")
}

func InitCrossplanePromise(cmd *cobra.Command, args []string) error {
	promiseName := args[0]
	if plural == "" {
		plural = fmt.Sprintf("%ss", strings.ToLower(kind))
	}

	xrd, err := getXRDFromFilepath()
	if err != nil {
		return err
	}

	var dependencies []v1alpha1.Dependency
	if !skipDependencies {
		if compositions != "" {
			dependencies, err = generateDependenciesFromCompositions(compositions)
			if err != nil {
				return fmt.Errorf("failed to generate dependencies from compositions: %w", err)
			}
		}
		objMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(xrd)
		if err != nil {
			return fmt.Errorf("Failed to parse xrd: %w", err)
		}
		dependencies = append(dependencies, v1alpha1.Dependency{Unstructured: unstructured.Unstructured{Object: objMap}})
	}

	storedVersionIdx := findXRDStoredVersionIndex(xrd)
	if storedVersionIdx == -1 {
		return fmt.Errorf("no served version found in XRD")
	}

	crd, err := generateCRDFromXRD(xrd.Spec.Versions[storedVersionIdx])
	if err != nil {
		return err
	}

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
			Value: xrd.Spec.Group,
		},
		{
			Name:  "VERSION",
			Value: xrd.Spec.Versions[storedVersionIdx].Name,
		},
		{
			Name:  "KIND",
			Value: xrd.Spec.ClaimNames.Kind,
		},
	}
	pipelines := generateResourceConfigurePipelines("from-api-to-crossplane-claim", "ghcr.io/syntasso/kratix-cli/from-api-to-crossplane-claim:v0.1.0", envs)

	workflowDirectory := filepath.Join("workflows", "resource", "configure")
	filesToWrite, err := getFilesToWrite(promiseName, split, workflowDirectory, crossplaneDestinationSelectors, dependencies, crd, pipelines, exampleResource)
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

func generateDependenciesFromCompositions(compositionsFilepath string) ([]v1alpha1.Dependency, error) {
	contents, err := os.ReadFile(compositionsFilepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", compositions, err)
	}

	var compositions []v1alpha1.Dependency
	docs := goyaml.NewDecoder(bytes.NewReader(contents))
	for {
		var comp map[string]any
		if err := docs.Decode(&comp); err != nil {
			if err.Error() == "EOF" {
				break
			}
			log.Fatalf("Failed to decode YAML: %v", err)
		}
		compositions = append(compositions, v1alpha1.Dependency{Unstructured: unstructured.Unstructured{Object: comp}})
	}

	return compositions, nil
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

	for key, value := range mandatoryAdditionalClaimFields {
		schema.Properties[key] = value
	}

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

func findXRDStoredVersionIndex(crd *xrdv1.CompositeResourceDefinition) int {
	for i, version := range crd.Spec.Versions {
		if version.Served {
			return i
		}
	}
	return -1
}

func getXRDFromFilepath() (*xrdv1.CompositeResourceDefinition, error) {
	xrd := &xrdv1.CompositeResourceDefinition{}

	// read xrdPath and unmarshal it into xrd
	contents, err := os.ReadFile(xrdPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", xrdPath, err)
	}

	if err := yaml.Unmarshal(contents, xrd); err != nil {
		return nil, fmt.Errorf("failed to unmarshal file %s: %w", xrdPath, err)
	}

	return xrd, nil
}
