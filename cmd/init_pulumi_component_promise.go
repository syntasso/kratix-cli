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
	"k8s.io/apimachinery/pkg/util/validation"
)

const (
	pulumiComponentPromiseCommandName = "pulumi-component-promise"
	pulumiComponentPromiseExamples    = `  # initialize a new promise from a local Pulumi package schema
  kratix init pulumi-component-promise mypromise --schema ./schema.json --group syntasso.io --kind Database

  # initialize a new promise from a remote Pulumi package schema
  kratix init pulumi-component-promise mypromise --schema https://www.pulumi.com/registry/packages/aws-iam/schema.json --component aws-iam:index:User --group syntasso.io --kind User

  # initialize a new promise from a private Pulumi package schema
  kratix init pulumi-component-promise mypromise --schema https://github.com/acme/k8s-cluster/schema.json --component k8s:index:Cluster --group acme.io --kind User --stack-access-token-secret acme-gh:token  --kind User --schema-bearer-token-secret acme-gh:token
`
)

var (
	pulumiSchemaPath              string
	pulumiComponent               string
	pulumiSchemaBearerTokenSecret string
	pulumiStackAccessTokenSecret  string

	pulumiDestinationSelectors = []v1alpha1.PromiseScheduling{{MatchLabels: map[string]string{"environment": "pulumi"}}}
)

type pulumiPromiseTemplateValues struct {
	promiseTemplateValues
	PulumiGeneratorName      string
	PulumiStackGeneratorName string
	SchemaBearerTokenSecret  *secretKeyRef
	StackAccessTokenSecret   *secretKeyRef
}

type secretKeyRef struct {
	Name string
	Key  string
}

var pulumiComponentPromiseCmd = &cobra.Command{
	Use:   pulumiComponentPromiseCommandName + " PROMISE-NAME --schema PATH_OR_URL --group PROMISE-API-GROUP --kind PROMISE-API-KIND [--component TOKEN] [--schema-bearer-token-secret] [--stack-access-token-secret] [--version] [--plural] [--split] [--dir DIR]",
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
	pulumiComponentPromiseCmd.Flags().StringVar(&pulumiSchemaBearerTokenSecret, "schema-bearer-token-secret", "", "Secret reference in SECRET_NAME:KEY format to set PULUMI_ACCESS_TOKEN for private schema fetches")
	pulumiComponentPromiseCmd.Flags().StringVar(&pulumiStackAccessTokenSecret, "stack-access-token-secret", "", "Secret reference in SECRET_NAME:KEY format to set Stack spec.envRefs.PULUMI_ACCESS_TOKEN for Pulumi Cloud access")

	pulumiComponentPromiseCmd.MarkFlagRequired("schema")
}

func InitPulumiComponentPromise(cmd *cobra.Command, args []string) error {
	printPreviewWarning()
	if pulumi.IsLocalSchemaSource(pulumiSchemaPath) {
		printPulumiLocalSchemaWarning(pulumiSchemaPath)
	}

	schemaBearerTokenSecret, err := parseSecretKeyRefFlag(pulumiSchemaBearerTokenSecret, "schema-bearer-token-secret")
	if err != nil {
		return err
	}
	if err := pulumi.ValidateSchemaSourceAuth(pulumiSchemaPath, schemaBearerTokenSecret != nil); err != nil {
		return err
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

	return initPulumiComponentPromiseFromSelection(args[0], selectedComponent, specSchema, schemaBearerTokenSecret)
}

func initPulumiComponentPromiseFromSelection(promiseName string, component pulumi.SelectedComponent, specSchema map[string]any, schemaBearerTokenSecret *secretKeyRef) error {
	var err error
	stackAccessTokenSecret, err := parseSecretKeyRefFlag(pulumiStackAccessTokenSecret, "stack-access-token-secret")
	if err != nil {
		return err
	}

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

	programEnv := []corev1.EnvVar{
		{
			Name:  "PULUMI_COMPONENT_TOKEN",
			Value: component.Token,
		},
		{
			Name:  "PULUMI_SCHEMA_SOURCE",
			Value: pulumiSchemaPath,
		},
	}
	if schemaBearerTokenSecret != nil {
		programEnv = append(programEnv, corev1.EnvVar{
			Name: "PULUMI_ACCESS_TOKEN",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: schemaBearerTokenSecret.Name,
					},
					Key: schemaBearerTokenSecret.Key,
				},
			},
		})
	}

	pipelines := generateResourceConfigurePipelinesWithContainers([]v1alpha1.Container{
		{
			Name:  pulumiProgramGeneratorContainerName,
			Image: pulumiGeneratorImage(),
			Command: []string{
				pulumiProgramGeneratorCommand,
			},
			Env: programEnv,
		},
		{
			Name:  pulumiStackGeneratorContainerName,
			Image: pulumiGeneratorImage(),
			Command: []string{
				pulumiStackGeneratorCommand,
			},
			Env: append([]corev1.EnvVar{
				{
					Name:  "PULUMI_COMPONENT_TOKEN",
					Value: component.Token,
				},
			}, stackAccessTokenSecretEnvVars(stackAccessTokenSecret)...),
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
		pulumiDestinationSelectors,
		[]v1alpha1.Dependency{},
		crd,
		pipelines,
		exampleResource,
		pulumiPromiseTemplateValues{
			promiseTemplateValues:    baseReadmeTemplateValues(pulumiComponentPromiseCommandName, extraFlags, promiseName, crd),
			PulumiGeneratorName:      pulumiProgramGeneratorContainerName,
			PulumiStackGeneratorName: pulumiStackGeneratorContainerName,
			SchemaBearerTokenSecret:  schemaBearerTokenSecret,
			StackAccessTokenSecret:   stackAccessTokenSecret,
		},
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
	if pulumiSchemaBearerTokenSecret != "" {
		flags = append(flags, "--schema-bearer-token-secret", shellQuoteArg(pulumiSchemaBearerTokenSecret))
	}
	if pulumiStackAccessTokenSecret != "" {
		flags = append(flags, "--stack-access-token-secret", shellQuoteArg(pulumiStackAccessTokenSecret))
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

func parseSecretKeyRefFlag(value, flagName string) (*secretKeyRef, error) {
	if value == "" {
		return nil, nil
	}

	parts := strings.SplitN(value, ":", 2)
	if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
		return nil, fmt.Errorf("parse --%s: expected SECRET_NAME:KEY", flagName)
	}

	secretName := strings.TrimSpace(parts[0])
	if err := validateKubernetesSecretName(secretName, flagName); err != nil {
		return nil, err
	}

	return &secretKeyRef{
		Name: secretName,
		Key:  strings.TrimSpace(parts[1]),
	}, nil
}

func validateKubernetesSecretName(secretName, flagName string) error {
	if len(validation.IsDNS1123Subdomain(secretName)) > 0 {
		return fmt.Errorf("parse --%s: secret name %q is not a valid Kubernetes Secret name. SECRET_NAME must be a valid Kubernetes Secret name (DNS-1123 subdomain, for example pulumi-schema-auth).", flagName, secretName)
	}

	return nil
}

func stackAccessTokenSecretEnvVars(secretRef *secretKeyRef) []corev1.EnvVar {
	if secretRef == nil {
		return nil
	}

	return []corev1.EnvVar{
		{
			Name:  "PULUMI_STACK_ACCESS_TOKEN_SECRET_NAME",
			Value: secretRef.Name,
		},
		{
			Name:  "PULUMI_STACK_ACCESS_TOKEN_SECRET_KEY",
			Value: secretRef.Key,
		},
	}
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
