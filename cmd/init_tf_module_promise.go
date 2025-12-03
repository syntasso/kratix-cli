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
		Short: "Initialize a Promise from a Terraform module",
		Long: `Initialize a Promise from a Terraform module.
  
This commands relies on the Terraform CLI being installed and available in your PATH. It can be used
to pull modules from Git, Terraform registry, or a local directory.

To pull modules from private registries, ensure your system is logged in to the registry with the 'terraform login' command.`,
		Example: `  # Initialize a Promise from a Terraform Module in git
  kratix init tf-module-promise vpc \
  	--module-source "git::https://github.com/terraform-aws-modules/terraform-aws-vpc.git?ref=v5.19.0" \
  	--group syntasso.io \
	--kind VPC \
	--version v1alpha1

  # Initialize a Promise from a Terraform Module in git with a specific path
  kratix init tf-module-promise gateway \
  	--module-source "git::https://github.com/GoogleCloudPlatform/cloud-foundation-fabric.git//modules/api-gateway?ref=v44.1.0" \
  	--group syntasso.io \
	--kind Gateway \
	--version v1alpha1 

  # Initialize a Promise from a Terraform Module in Terraform registry
  kratix init tf-module-promise iam \
  	--module-source terraform-aws-modules/iam/aws \
  	--module-version 6.2.3 \
  	--group syntasso.io \
	--kind IAM \
	--version v1alpha1`,
		RunE: InitFromTerraformModule,
		Args: cobra.ExactArgs(1),
	}

	moduleSource, moduleVersion string
)

func init() {
	initCmd.AddCommand(terraformModuleCmd)
	terraformModuleCmd.Flags().StringVarP(&moduleSource, "module-source", "s", "", "Source of the terraform module. \n"+
		"This can be a Git URL, Terraform registry path, or a local directory path. \n"+
		"It follows the same format as the `source` argument in the Terraform module block.",
	)
	terraformModuleCmd.Flags().StringVarP(&moduleVersion, "module-version", "m", "", "(Optional) version of the terraform module; "+
		"only use when pulling modules from Terraform registry",
	)
	terraformModuleCmd.MarkFlagRequired("module-source")
}

func InitFromTerraformModule(cmd *cobra.Command, args []string) error {
	fmt.Println("Fetching terraform module variables, this might take up to a minute...")
	variables, err := internal.GetVariablesFromModule(moduleSource, moduleVersion)
	if err != nil {
		fmt.Printf("Error: failed to download and convert terraform module to CRD: %s\n", err)
		return nil
	}

	crdSpecSchema, warnings := internal.VariablesToCRDSpecSchema(variables)

	for _, warning := range warnings {
		fmt.Println(warning)
	}

	crdSchema, err := yaml.Marshal(crdSpecSchema)
	if err != nil {
		fmt.Printf("Error: failed to marshal CRD schema: %s\n", err)
		return nil
	}

	resourceConfigure, err := generateTerraformModuleResourceConfigurePipeline()
	if err != nil {
		fmt.Printf("Error: failed to generate promise pipelines: %s\n", err)
		return nil
	}

	promiseName := args[0]
	extraFlags := fmt.Sprintf("--module-source %s", moduleSource)
	if moduleVersion != "" {
		extraFlags = fmt.Sprintf("%s --module-version %s", extraFlags, moduleVersion)
	}
	templateValues, err := generateTemplateValues(promiseName, "tf-module-promise", extraFlags, resourceConfigure, string(crdSchema))
	if err != nil {
		fmt.Printf("Error: failed to generate template values: %s\n", err)
		return nil
	}
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
		fmt.Printf("Error: failed to template files: %s\n", err)
		return nil
	}

	fmt.Println("Promise generated successfully. It is set to schedule to Destinations with the label `environment: terraform` by default. To modify this behavior, update the `.spec.destinationSelectors` field in `promise.yaml`")
	return nil
}

func generateTerraformModuleResourceConfigurePipeline() (string, error) {
	envs := []corev1.EnvVar{
		{
			Name:  "MODULE_SOURCE",
			Value: moduleSource,
		},
	}

	if moduleVersion != "" {
		envs = append(envs, corev1.EnvVar{
			Name:  "MODULE_VERSION",
			Value: moduleVersion,
		})
	}

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
							Image: "ghcr.io/syntasso/kratix-cli/terraform-generate:v0.2.0",
							Env:   envs,
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
