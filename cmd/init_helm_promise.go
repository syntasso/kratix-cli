package cmd

import (
	"fmt"
	helmclient "github.com/mittwald/go-helm-client"
	"github.com/spf13/cobra"
	"github.com/syntasso/kratix-cli/internal"
	"github.com/syntasso/kratix/api/v1alpha1"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/registry"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

var intHelmPromiseCmd = &cobra.Command{
	Use:   "helm-promise PROMISE-NAME --chart-url HELM-CHART-URL --group PROMISE-API-GROUP --kind PROMISE-API-KIND [--chart-version]",
	Short: "Initialize a new Promise from a Helm chart",
	Long:  "Initialize a new Promise from a Helm Chart",
	Example: `  # initialize a new promise from an OCI Helm Chart
  kratix init helm-promise postgresql --chart-url oci://registry-1.docker.io/bitnamicharts/postgresql [--chart-version]

  # initialize a new promise from a Helm Chart repository
  kratix init helm-promise postgresql --chart-url https://fluxcd-community.github.io/helm-charts --chart-name flux2 [--chart-version]

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
	promiseName := args[0]
	resourceConfigure, err := generateResourceConfigurePipeline()
	if err != nil {
		return err
	}

	crdSchema, err := schemaFromChart()
	if err != nil {
		return err
	}

	templateValues := generateTemplateValues(promiseName, "helm-promise", resourceConfigure, crdSchema)

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
							Image: "ghcr.io/syntasso/kratix-cli/helm-resource-configure:v0.1.0",
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

func schemaFromChart() (string, error) {
	values, err := valuesFromChart()
	if err != nil {
		return "", err
	}
	schema, err := internal.HelmValuesToSchema(values)
	if err != nil {
		return "", fmt.Errorf("failed to convert helm values to schema: %w", err)
	}

	bytes, err := yaml.Marshal(*schema)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

func valuesFromChart() (map[string]interface{}, error) {
	client, err := helmclient.New(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create helm client: %w", err)
	}

	install := action.NewInstall(&action.Configuration{})
	registryClient, _ := registry.NewClient()
	install.SetRegistryClient(registryClient)

	if chartName != "" {
		install.RepoURL = chartURL
	}

	if chartVersion != "" {
		install.ChartPathOptions.Version = chartVersion
	}

	chart, _, err := client.GetChart(getChartName(), &install.ChartPathOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch helm chart: %w", err)
	}

	return chart.Values, nil
}

// when provided --chart-url is a chart repo and --chart-name is provided, getChartName() returns chart-name
// when provided --chart-url is OCI or a tar chart, getChartName() returns chart url
func getChartName() string {
	if chartName != "" {
		return chartName
	}
	return chartURL
}
