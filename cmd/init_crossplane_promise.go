package cmd

import (
	"bytes"
	"fmt"
	"log"
	"maps"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/syntasso/kratix/api/v1alpha1"
	goyaml "gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"

	xrdv1 "github.com/crossplane/crossplane/apis/apiextensions/v1"
)

const (
	crossplaneContainerName  = "from-api-to-crossplane-claim"
	crossplaneContainerImage = "ghcr.io/syntasso/kratix-cli/from-api-to-crossplane-claim:v0.2.2"

	workflowDirectory = "workflows/resource/configure"

	XRD_GROUP_ENV_VAR   = "XRD_GROUP"
	XRD_VERSION_ENV_VAR = "XRD_VERSION"
	XRD_KIND_ENV_VAR    = "XRD_KIND"
)

var (
	crossplaneDestinationSelectors = []v1alpha1.PromiseScheduling{{MatchLabels: map[string]string{"crossplane": "enabled"}}}

	// crossplanePromiseCmd represents the crossplanePromise command
	crossplanePromiseCmd = &cobra.Command{
		Use:   "crossplane-promise",
		Short: "Preview: Initialize a new Promise from a Crossplane XRD",
		Long: "Preview: Initialize a new Promise from a Crossplane XRD. " +
			"This command is in preview, not supported under SLAs, and may change or break without notice.",
		Example: `  # initialize a new promise from a Crossplane XRD and Composition
  kratix init crossplane-promise s3buckets --xrd xrd.yaml --group syntasso.io --kind S3Bucket --dir --compositions composition.yaml
`,

		Args: cobra.ExactArgs(1),
		RunE: InitCrossplanePromise,
	}

	xrdPath          string
	compositions     string
	functions        string
	skipDependencies bool
)

func init() {
	initCmd.AddCommand(crossplanePromiseCmd)
	crossplanePromiseCmd.Flags().StringVarP(&xrdPath, "xrd", "x", "", "Filepath to the XRD file")
	crossplanePromiseCmd.Flags().StringVarP(&compositions, "compositions", "c", "", "Filepath to the Compositions file. Can contain a single Composition or multiple Compositions.")
	crossplanePromiseCmd.Flags().StringVarP(&functions, "functions", "f", "", "Filepath to the Functions file. Can contain a single Function or multiple Functions.")
	crossplanePromiseCmd.Flags().BoolVarP(&skipDependencies, "skip-dependencies", "s", false, "Skip generating dependencies. For when the XRD and Compositions are already deployed to Crossplane")
	crossplanePromiseCmd.MarkFlagRequired("xrd")
}

func InitCrossplanePromise(cmd *cobra.Command, args []string) error {
	printPreviewWarning()
	promiseName := args[0]
	if plural == "" {
		plural = fmt.Sprintf("%ss", strings.ToLower(kind))
	}

	xrd, err := getXRD(xrdPath)
	if err != nil {
		return err
	}

	var dependencies []v1alpha1.Dependency
	if !skipDependencies {
		if functions != "" {
			functionDeps, err := generateDependenciesFromFile(functions)
			if err != nil {
				return fmt.Errorf("failed to generate dependencies from functions: %w", err)
			}
			dependencies = append(dependencies, functionDeps...)
		}
		if compositions != "" {
			compDeps, err := generateDependenciesFromFile(compositions)
			if err != nil {
				return fmt.Errorf("failed to generate dependencies from compositions: %w", err)
			}
			dependencies = append(dependencies, compDeps...)
		}
		xrdRaw, err := readFileAsUnstructured(xrdPath)
		if err != nil {
			return fmt.Errorf("failed to parse xrd: %w", err)
		}
		dependencies = append(dependencies, v1alpha1.Dependency{Unstructured: unstructured.Unstructured{Object: xrdRaw}})
	}

	xrdStoredVersion, err := getXRDStoredVersion(xrd)
	if err != nil {
		return err
	}

	crd, err := generateCRDFromXRD(xrdStoredVersion)
	if err != nil {
		return err
	}

	xrdKind := xrd.Spec.Names.Kind
	if xrd.Spec.ClaimNames != nil {
		xrdKind = xrd.Spec.ClaimNames.Kind
	}

	pipelines := generateResourceConfigurePipelines(crossplaneContainerName, crossplaneContainerImage, []corev1.EnvVar{
		{
			Name:  XRD_GROUP_ENV_VAR,
			Value: xrd.Spec.Group,
		},
		{
			Name:  XRD_VERSION_ENV_VAR,
			Value: xrdStoredVersion.Name,
		},
		{
			Name:  XRD_KIND_ENV_VAR,
			Value: xrdKind,
		},
	})

	exampleResource := generateExampleResource(crd)
	warnFieldsWithoutDefaults(crd)
	flags := fmt.Sprintf("--xrd %s", xrdPath)
	if functions != "" {
		flags = fmt.Sprintf("%s --functions %s", flags, functions)
	}
	if compositions != "" {
		flags = fmt.Sprintf("%s --compositions %s", flags, compositions)
	}
	if skipDependencies {
		flags = fmt.Sprintf("%s --skip-dependencies", flags)
	}
	filesToWrite, err := getFilesToWrite("crossplane-promise", promiseName, split, workflowDirectory, flags, crossplaneDestinationSelectors, dependencies, crd, pipelines, exampleResource, nil)
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

func getXRDStoredVersion(xrd *xrdv1.CompositeResourceDefinition) (*xrdv1.CompositeResourceDefinitionVersion, error) {
	for i, version := range xrd.Spec.Versions {
		if version.Served {
			return &xrd.Spec.Versions[i], nil
		}
	}
	return nil, fmt.Errorf("no served version found in XRD")
}

func generateExampleResource(crd *apiextensionsv1.CustomResourceDefinition) *unstructured.Unstructured {
	return &unstructured.Unstructured{
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
}

func warnFieldsWithoutDefaults(crd *apiextensionsv1.CustomResourceDefinition) {
	crdSpec := crd.Spec.Versions[0].Schema.OpenAPIV3Schema.Properties["spec"]
	for _, field := range crdSpec.Required {
		if crdSpec.Properties[field].Default == nil {
			fmt.Printf("warning: required field %q has no default value; a default is required for top level required fields in a CRD\nYou will need to add a default to make the Promise API valid.\n", field)
		}
	}
}

func generateDependenciesFromFile(filepath string) ([]v1alpha1.Dependency, error) {
	contents, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filepath, err)
	}

	var deps []v1alpha1.Dependency
	docs := goyaml.NewDecoder(bytes.NewReader(contents))
	for {
		var obj map[string]any
		if err := docs.Decode(&obj); err != nil {
			if err.Error() == "EOF" {
				break
			}
			log.Fatalf("Failed to decode YAML: %v", err)
		}
		deps = append(deps, v1alpha1.Dependency{Unstructured: unstructured.Unstructured{Object: obj}})
	}

	return deps, nil
}

func generateCRDFromXRD(version *xrdv1.CompositeResourceDefinitionVersion) (*apiextensionsv1.CustomResourceDefinition, error) {
	if version.Schema == nil {
		return nil, fmt.Errorf("version %s has no schema", version.Name)
	}
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

	if schema.Properties == nil {
		schema.Properties = make(map[string]apiextensionsv1.JSONSchemaProps)
	}
	specProp := schema.Properties["spec"]
	specProp.Default = &apiextensionsv1.JSON{Raw: []byte(`{}`)}
	if specProp.Properties == nil {
		specProp.Properties = make(map[string]apiextensionsv1.JSONSchemaProps)
	}
	schema.Properties["spec"] = specProp
	maps.Copy(schema.Properties["spec"].Properties, mandatoryAdditionalClaimFields)

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

func getXRD(path string) (*xrdv1.CompositeResourceDefinition, error) {
	xrd := &xrdv1.CompositeResourceDefinition{}
	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", path, err)
	}

	if err := yaml.Unmarshal(contents, xrd); err != nil {
		return nil, fmt.Errorf("failed to unmarshal file %s: %w", path, err)
	}

	return xrd, nil
}

func readFileAsUnstructured(path string) (map[string]any, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", path, err)
	}
	var obj map[string]any
	if err := yaml.Unmarshal(contents, &obj); err != nil {
		return nil, fmt.Errorf("failed to unmarshal file %s: %w", path, err)
	}
	return obj, nil
}
