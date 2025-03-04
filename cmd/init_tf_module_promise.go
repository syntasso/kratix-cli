package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/syntasso/kratix-cli/internal"
	"github.com/syntasso/kratix/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

// terraformModuleCmd represents the terraformModule command
var (
	terraformModuleCmd = &cobra.Command{
		Use:   "tf-module-promise",
		Short: "Initialize a new promise from a terraform module",
		Long:  "Initialize a new promise from a terraform module",
		Example: ` # Initialize a new promise from a terraform module in Git
	  kratix init tf-module-promise vpc --version v5.19.0 --source https://github.com/terraform-aws-modules/terraform-aws-vpc.git --group syntasso.io --kind VPC
		`,
		RunE: InitFromTerraformModule,
		Args: cobra.ExactArgs(1),
	}

	moduleSource, moduleVersion string
)

func init() {
	initCmd.AddCommand(terraformModuleCmd)
	terraformModuleCmd.Flags().StringVarP(&moduleSource, "source", "s", "", "source of the terraform module")
	terraformModuleCmd.Flags().StringVarP(&moduleVersion, "version", "v", "", "version of the terraform module")
}

func InitFromTerraformModule(cmd *cobra.Command, args []string) error {
	versionedModuleSourceURL := fmt.Sprintf("git::%s?ref=%s", moduleSource, moduleVersion)
	variables, err := internal.DownloadAndConvertTerraformToCRD(versionedModuleSourceURL)
	if err != nil {
		return fmt.Errorf("failed to download and convert terraform module to CRD: %w", err)
	}

	crdSpecSchema, warnings := internal.VariablesToCRDSpecSchema(variables)

	for _, warning := range warnings {
		fmt.Println(warning)
	}

	crdSchema, err := yaml.Marshal(crdSpecSchema)
	if err != nil {
		return fmt.Errorf("failed to marshal CRD schema: %w", err)
	}

	resourceConfigure, err := generateTerraformModuleResourceConfigurePipeline()
	if err != nil {
		return err
	}

	promiseName := args[0]
	templateValues := generateTemplateValues(promiseName, "terraform-module-promise", resourceConfigure, string(crdSchema))
	templateValues.DestinationSelectors = "- matchLabels:\n    environment: terraform"

	templates := map[string]string{
		resourceFileName: "templates/promise/example-resource.yaml.tpl",
		"README.md":      "templates/promise/README.md.tpl",
	}

	if split {
		templates[apiFileName] = "templates/promise/api.yaml.tpl"
		templates[dependenciesFileName] = "templates/promise/dependencies.yaml"
		templates[resourceConfigureWorkflowFileName] = "templates/promise/workflow.yaml.tpl"
	} else {
		templates[promiseFileName] = "templates/promise/promise.yaml.tpl"
	}

	err = templateFiles(promiseTemplates, outputDir, templates, templateValues)
	if err != nil {
		return err
	}

	dirName := "the current directory"
	if outputDir != "." {
		dirName = outputDir
	}

	fmt.Printf("%s promise bootstrapped in %s\n", promiseName, dirName)
	return err
}

func generateTerraformModuleResourceConfigurePipeline() (string, error) {
	pipelines := []unstructured.Unstructured{
		{
			Object: map[string]any{
				"apiVersion": "platform.kratix.io/v1alpha1",
				"kind":       "Pipeline",
				"metadata": map[string]any{
					"name": "instance-configure",
				},
				"spec": map[string]any{
					"containers": []any{
						v1alpha1.Container{
							Name:  "terraform-generate",
							Image: "ghcr.io/syntasso/kratix-cli/terraform-generate:v0.1.0",
							Env: []corev1.EnvVar{
								{
									Name:  "MODULE_SOURCE",
									Value: moduleSource,
								},
								{
									Name:  "MODULE_VERSION",
									Value: moduleVersion,
								},
							},
						},
					},
				},
			},
		},
	}
	pipelineBytes, err := yaml.Marshal(pipelines)
	if err != nil {
		return "", err
	}
	return string(pipelineBytes), nil
}
