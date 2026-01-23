package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	pipelineutils "github.com/syntasso/kratix-cli/cmd/pipeline_utils"
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
	--module-providers "versions.tf,providers.tf" \
  	--group syntasso.io \
	--kind Gateway \
	--version v1alpha1 

  # Initialize a Promise from a Terraform Module in Terraform registry with a dedicated repository
  kratix init tf-module-promise s3-bucket \
    --module-source "terraform-aws-modules/s3-bucket/aws" \
    --module-registry-version 3.7.0 \
    --group syntasso.io \
    --kind S3Bucket \
    --version v1alpha1

  # Initialize a Promise from a Terraform Module in Terraform registry in a monorepo
  kratix init tf-module-promise iam \
  	--module-source terraform-aws-modules/iam/aws//modules/iam-account \
  	--module-registry-version 6.2.3 \
  	--group syntasso.io \
	--kind IAM \
	--version v1alpha1`,
		RunE: InitFromTerraformModule,
		Args: cobra.ExactArgs(1),
	}

	moduleSource, moduleRegistryVersion string
	moduleProviders                     []string
)

func init() {
	initCmd.AddCommand(terraformModuleCmd)
	terraformModuleCmd.Flags().StringVarP(&moduleSource, "module-source", "s", "", "Source of the terraform module. \n"+
		"This can be a Git URL, Terraform registry path, or a local directory path. \n"+
		"It follows the same format as the `source` argument in the Terraform module block.",
	)
	terraformModuleCmd.Flags().StringVarP(&moduleRegistryVersion, "module-registry-version", "r", "", "(Optional) version of the Terraform module from a registry; "+
		"only use when pulling modules from Terraform registry",
	)
	terraformModuleCmd.Flags().StringSliceVarP(&moduleProviders, "module-providers", "", []string{}, "(Optional) the names of any files containing Terraform provider block; "+
		"defaults to versions.tf and providers.tf",
	)
	terraformModuleCmd.MarkFlagRequired("module-source")
}

func InitFromTerraformModule(cmd *cobra.Command, args []string) error {
	fmt.Println("Fetching terraform module variables, this might take up to a minute...")

	if moduleRegistryVersion != "" && !internal.IsTerraformRegistrySource(moduleSource) {
		fmt.Println("Error: --module-registry-version is only valid for Terraform registry sources like 'namespace/name/provider'. For git URLs (e.g., 'git::https://github.com/org/repo.git?ref=v1.0.0') or local paths, embed the ref directly in --module-source instead.")
	}

	moduleDir, err := internal.SetupModule(moduleSource, moduleRegistryVersion)
	if err != nil {
		fmt.Printf("Error: failed to setup module : %s\n", err)
		return nil
	}
	defer os.RemoveAll(moduleDir)

	variables, err := internal.GetVariablesFromModule(moduleSource, moduleDir, moduleRegistryVersion)
	if err != nil {
		fmt.Printf("Error: failed to download and convert terraform module to CRD: %s\n", err)
		return nil
	}

	versionProviderFilepaths, err := internal.GetVersionsAndProvidersFromModule(moduleSource, moduleDir, moduleRegistryVersion, moduleProviders)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
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

	resourceConfigure, err := generateTerraformModuleResourceConfigurePipeline(moduleRegistryVersion)
	if err != nil {
		fmt.Printf("Error: failed to generate promise pipelines: %s\n", err)
		return nil
	}

	var promiseConfigure string
	if len(versionProviderFilepaths) > 1 {
		promiseConfigure, err = generateTerraformModulePromiseConfigurePipeline()
		if err != nil {
			fmt.Printf("Error: failed to generate promise configure pipelines: %s\n", err)
			return nil
		}
	}

	promiseName := args[0]
	extraFlags := fmt.Sprintf("--module-source %s", moduleSource)
	if moduleRegistryVersion != "" {
		extraFlags = fmt.Sprintf("%s --module-registry-version %s", extraFlags, moduleRegistryVersion)
	}
	templateValues, err := generateTemplateValues(promiseName, "tf-module-promise", extraFlags, resourceConfigure, promiseConfigure, string(crdSchema))
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

	err = writeDependencyFiles(versionProviderFilepaths)
	if err != nil {
		fmt.Printf("error writing promise dependencies: %s\n", err)
		return nil
	}

	fmt.Println("Promise generated successfully. It is set to schedule to Destinations with the label `environment: terraform` by default. To modify this behavior, update the `.spec.destinationSelectors` field in `promise.yaml`")
	return nil
}

func generateTerraformModuleResourceConfigurePipeline(moduleRegistryVersion string) (string, error) {
	envs := []corev1.EnvVar{
		{
			Name:  "MODULE_SOURCE",
			Value: moduleSource,
		},
	}

	if moduleRegistryVersion != "" {
		envs = append(envs, corev1.EnvVar{
			Name:  "MODULE_REGISTRY_VERSION",
			Value: moduleRegistryVersion,
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
							Image: "ghcr.io/syntasso/kratix-cli/terraform-generate:v0.4.0",
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
	return strings.TrimSuffix(string(pipelineBytes), "\n"), nil
}

func generateTerraformModulePromiseConfigurePipeline() (string, error) {
	pipelines := []unstructured.Unstructured{
		{
			Object: map[string]any{
				"apiVersion": "platform.kratix.io/v1alpha1",
				"kind":       "Pipeline",
				"metadata": map[string]any{
					"name": "dependencies",
				},
				"spec": map[string]any{
					"containers": []any{
						v1alpha1.Container{
							Name:  "add-tf-dependencies",
							Image: "my-registry.io/my-org/kratix/terraform-dependencies:v0.0.1",
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
	return strings.TrimSuffix(string(pipelineBytes), "\n"), nil
}

func writeDependencyFiles(versionProviderFilepaths []string) error {
	pipelineCmdArgs := &pipelineutils.PipelineCmdArgs{
		Lifecycle: "promise",
		Action:    "configure",
		Pipeline:  "dependencies",
	}

	containerName := "add-tf-dependencies"
	containerImage := "my-registry.io/my-org/kratix/terraform-dependencies:v0.0.1"

	err := generateWorkflow(pipelineCmdArgs, containerName, containerImage, outputDir, true)
	if err != nil {
		return fmt.Errorf("error generating workflows for %s/%s/%s: %s", pipelineCmdArgs.Lifecycle, pipelineCmdArgs.Action, pipelineCmdArgs.Pipeline, err)
	}

	containerDir := filepath.Join(dir, "workflows", pipelineCmdArgs.Lifecycle, pipelineCmdArgs.Action, pipelineCmdArgs.Pipeline, containerName)
	resourcesDir := filepath.Join(containerDir, "resources")
	for _, providerFilepath := range versionProviderFilepaths {
		sourceFile, err := os.Open(providerFilepath)
		if err != nil {
			return fmt.Errorf("error opening provider file: %s", err)
		}
		defer sourceFile.Close()

		baseFileName := filepath.Base(providerFilepath)
		destFile, err := os.Create(filepath.Join(outputDir, resourcesDir, baseFileName))
		if err != nil {
			return fmt.Errorf("error creating provider file: %s", err)
		}
		defer destFile.Close()

		_, err = io.Copy(destFile, sourceFile)
		if err != nil {
			return fmt.Errorf("error copying provider file: %s", err)
		}
	}

	scriptsDir := filepath.Join(containerDir, "scripts")
	pipelineScriptContent := "#!/usr/bin/env sh\n\ncp /resources/* /kratix/output"
	if err := os.WriteFile(filepath.Join(outputDir, scriptsDir, "pipeline.sh"), []byte(pipelineScriptContent), filePerm); err != nil {
		return err
	}

	fmt.Println("Dependencies added as a Promise workflow.")
	fmt.Println("Run the following command to build the dependencies image:")
	fmt.Printf("\n  docker build -t %s %s\n\n", image, containerDir)
	fmt.Println("Don't forget to push the image to a registry!")
	return nil
}
