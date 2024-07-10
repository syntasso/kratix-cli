package cmd

import (
	"fmt"
	"github.com/syntasso/kratix/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
	"strings"

	"github.com/spf13/cobra"
)

// helmPromiseCmd represents the helmPromise command
var intHelmPromiseCmd = &cobra.Command{
	Use:   "helm-promise PROMISE-NAME --chart-url HELM-CHART-URL [--version]",
	Short: "Initialize a new Promise from a Helm chart",
	Long:  "Initialize a new Promise from a Helm Chart within the current directory, with all the necessary files to get started",
	Example: `  # initialize a new promise from the OCI Helm Chart
  kratix init helm-promise postgresql --chart-url oci://registry-1.docker.io/bitnamicharts/postgresql [--chart-version v1.0.0]

  # initialize a new promise from a Helm Chart repository
  kratix init helm-promise postgresql --chart-url https://fluxcd-community.github.io/helm-charts --chart-name flux2 [--chart-version v1.0.0]

  # initialize a new promise from a Helm Chart tar URL
  kratix init helm-promise postgresql --chart-url https://github.com/stefanprodan/podinfo/raw/gh-pages/podinfo-0.2.1.tgz
`,
	RunE: InitHelmPromise,
	Args: cobra.ExactArgs(1),
}

var chartURL, chartName, chartVersion string

func init() {
	initCmd.AddCommand(intHelmPromiseCmd)
	intHelmPromiseCmd.Flags().StringVarP(&chartURL, "chart-url", "", "", "The URL (supports OCI and tarball) of the Helm chart")
	intHelmPromiseCmd.Flags().StringVarP(&chartVersion, "chart-version", "", "", "The Helm chart version. Default to latest")
	intHelmPromiseCmd.Flags().StringVarP(&chartName, "chart-name", "", "", "The Helm chart name. Required when using Helm repository")
	intHelmPromiseCmd.MarkFlagRequired("chart-url")
}

func InitHelmPromise(cmd *cobra.Command, args []string) error {
	if version == "" {
		version = "v1alpha1"
	}

	if plural == "" {
		plural = fmt.Sprintf("%ss", strings.ToLower(kind))
	}

	promiseName := args[0]

	resourceConfigure, err := generateResourceConfigurePipeline()
	if err != nil {
		return err
	}

	templateValues := promiseTemplateValues{
		Name:              promiseName,
		Group:             group,
		Kind:              kind,
		Version:           version,
		Plural:            plural,
		Singular:          strings.ToLower(kind),
		SubCommand:        "helm-promise",
		ResourceConfigure: resourceConfigure,
	}

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

	dirName := "current"
	if outputDir != "." {
		dirName = outputDir
	}

	fmt.Printf("%s promise bootstrapped in the %s directory\n", promiseName, dirName)
	return nil
}

func generateResourceConfigurePipeline() (string, error) {
	envVars := []corev1.EnvVar{{Name: "CHART_URL", Value: chartURL}}
	if chartName != "" {
		envVars = append(envVars, corev1.EnvVar{Name: "CHART_NAME", Value: chartName})
	}

	if chartVersion != "" {
		envVars = append(envVars, corev1.EnvVar{Name: "CHART_VERSION", Value: chartVersion})
	}

	pipelines := []unstructured.Unstructured{
		{
			Object: map[string]interface{}{
				"apiVersion": "platform.kratix.io/v1alpha1",
				"kind":       "Pipeline",
				"metadata": map[string]interface{}{
					"name": "instance-configure",
				},
				"spec": map[string]interface{}{
					"containers": []interface{}{
						v1alpha1.Container{
							Name:  "instance-configure",
							Image: "ghcr.io/syntasso/kratix-cli/helm-instance-configure:v0.1.0",
							Env:   envVars,
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
