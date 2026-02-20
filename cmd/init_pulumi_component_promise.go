package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/syntasso/kratix-cli/internal/pulumi"
	"github.com/syntasso/kratix/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	pulumiComponentPromiseCommandName = "pulumi-component-promise"
	pulumiComponentPromiseExamples    = `  # initialize a new promise from a local Pulumi package schema
  kratix init pulumi-component-promise mypromise --schema ./schema.json --group syntasso.io --kind Database

  # initialize a new promise from a remote Pulumi package schema
  kratix init pulumi-component-promise mypromise --schema https://www.pulumi.com/registry/packages/aws/schema.json --group syntasso.io --kind Database
`
	pulumiComponentContainerName  = "from-api-to-pulumi-pko-program"
	pulumiComponentContainerImage = "ghcr.io/syntasso/kratix-cli/from-api-to-pulumi-pko-program:v0.1.0"
)

var (
	pulumiSchemaPath string
	pulumiComponent  string
)

var pulumiComponentPromiseCmd = &cobra.Command{
	Use:   pulumiComponentPromiseCommandName + " PROMISE-NAME --schema PATH_OR_URL --group PROMISE-API-GROUP --kind PROMISE-API-KIND [--component TOKEN] [--version] [--plural] [--split] [--dir DIR]",
	Short: "Preview: Initialize a new Promise from a Pulumi package schema",
	Long: "Preview: Initialize a new Promise from a Pulumi package schema. " +
		"This command is in preview, not supported under SLAs, and may change or break without notice.",
	Example: pulumiComponentPromiseExamples,
	Args:    cobra.ExactArgs(1),
	RunE:    InitPulumiComponentPromise,
}

func init() {
	initCmd.AddCommand(pulumiComponentPromiseCmd)

	pulumiComponentPromiseCmd.Flags().StringVar(&pulumiSchemaPath, "schema", "", "Path or URL to Pulumi package schema")
	pulumiComponentPromiseCmd.Flags().StringVar(&pulumiComponent, "component", "", "Pulumi component token to use from the schema")

	pulumiComponentPromiseCmd.MarkFlagRequired("schema")
}

func InitPulumiComponentPromise(cmd *cobra.Command, args []string) error {
	printPreviewWarning()
	if pulumi.IsLocalSchemaSource(pulumiSchemaPath) {
		printPulumiLocalSchemaWarning(pulumiSchemaPath)
	}

	schemaDoc, err := pulumi.LoadSchema(pulumiSchemaPath)
	if err != nil {
		return err
	}

	selectedComponent, err := pulumi.SelectComponent(schemaDoc, pulumiComponent)
	if err != nil {
		return err
	}

	specSchema, warnings, err := pulumi.TranslateInputsToSpecSchema(schemaDoc, selectedComponent)
	if err != nil {
		return err
	}
	for _, warning := range warnings {
		fmt.Println(warning)
	}

	return initPulumiComponentPromiseFromSelection(args[0], selectedComponent, specSchema)
}

func initPulumiComponentPromiseFromSelection(promiseName string, component pulumi.SelectedComponent, specSchema map[string]any) error {
	extraFlags := buildPulumiPromiseExtraFlags()

	if version == "" {
		version = "v1alpha1"
	}
	if plural == "" {
		plural = fmt.Sprintf("%ss", strings.ToLower(kind))
	}

	crd, err := buildPulumiCRD(specSchema)
	if err != nil {
		return err
	}

	pipelines := generateResourceConfigurePipelines(pulumiComponentContainerName, pulumiComponentContainerImage, []corev1.EnvVar{
		{
			Name:  "PULUMI_COMPONENT_TOKEN",
			Value: component.Token,
		},
		{
			Name:  "PULUMI_SCHEMA_SOURCE",
			Value: pulumiSchemaPath,
		},
	})

	exampleResource := &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": fmt.Sprintf("%s/%s", crd.Spec.Group, crd.Spec.Versions[0].Name),
			"kind":       kind,
			"metadata": map[string]string{
				"name":      "example-request",
				"namespace": "default",
			},
			"spec": topLevelRequiredFields(crd),
		},
	}

	filesToWrite, err := getFilesToWrite(
		pulumiComponentPromiseCommandName,
		promiseName,
		split,
		workflowDirectory,
		extraFlags,
		nil,
		[]v1alpha1.Dependency{},
		crd,
		pipelines,
		exampleResource,
	)
	if err != nil {
		return err
	}

	if err := writePromiseFiles(outputDir, filesToWrite); err != nil {
		return err
	}

	fmt.Println("Pulumi component Promise generated successfully.")
	return nil
}

func printPulumiLocalSchemaWarning(source string) {
	fmt.Printf("warning: local Pulumi schema source %q detected. The generated resource workflow runs in Kubernetes and cannot read files from your machine.\n", source)
	fmt.Println("warning: prefer publishing your Pulumi component/schema for remote HTTP(S) access and pass that URL with --schema.")
	fmt.Println("warning: for local iteration before publishing, make the schema reachable from the cluster (for example: bake it into the stage image, mount it via ConfigMap/volume, or host it in object storage).")
}

func buildPulumiPromiseExtraFlags() string {
	flags := []string{"--schema", shellQuoteArg(pulumiSchemaPath)}

	if pulumiComponent != "" {
		flags = append(flags, "--component", shellQuoteArg(pulumiComponent))
	}
	if version != "" {
		flags = append(flags, "--version", shellQuoteArg(version))
	}
	if plural != "" {
		flags = append(flags, "--plural", shellQuoteArg(plural))
	}
	if split {
		flags = append(flags, "--split")
	}

	return strings.Join(flags, " ")
}

func shellQuoteArg(arg string) string {
	return "'" + strings.ReplaceAll(arg, "'", `'"'"'`) + "'"
}

func buildPulumiCRD(specSchema map[string]any) (*apiextensionsv1.CustomResourceDefinition, error) {
	specSchemaBytes, err := json.Marshal(specSchema)
	if err != nil {
		return nil, fmt.Errorf("build Promise CRD: marshal translated schema: %w", err)
	}

	var specProps apiextensionsv1.JSONSchemaProps
	if err := json.Unmarshal(specSchemaBytes, &specProps); err != nil {
		return nil, fmt.Errorf("build Promise CRD: parse translated schema: %w", err)
	}
	specProps.Default = &apiextensionsv1.JSON{Raw: []byte(`{}`)}

	return &apiextensionsv1.CustomResourceDefinition{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apiextensions.k8s.io/v1",
			Kind:       "CustomResourceDefinition",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s.%s", plural, group),
		},
		Spec: apiextensionsv1.CustomResourceDefinitionSpec{
			Group: group,
			Names: apiextensionsv1.CustomResourceDefinitionNames{
				Plural:   plural,
				Singular: strings.ToLower(kind),
				Kind:     kind,
			},
			Scope: "Namespaced",
			Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
				{
					Name:    version,
					Served:  true,
					Storage: true,
					Schema: &apiextensionsv1.CustomResourceValidation{
						OpenAPIV3Schema: &apiextensionsv1.JSONSchemaProps{
							Type: "object",
							Properties: map[string]apiextensionsv1.JSONSchemaProps{
								"spec": specProps,
							},
						},
					},
				},
			},
		},
	}, nil
}
