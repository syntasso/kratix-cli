package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
	"github.com/syntasso/kratix/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	yamlsig "sigs.k8s.io/yaml"
)

const (
	operatorContainerName  = "from-api-to-operator"
	operatorContainerImage = "ghcr.io/syntasso/kratix-cli/from-api-to-operator:v0.1.0"
)

var operatorPromiseCmd = &cobra.Command{
	Use:   "operator-promise PROMISE-NAME --group PROMISE-API-GROUP --version PROMISE-API-VERSION --kind PROMISE-API-KIND --operator-manifests OPERATOR-MANIFESTS-DIR --api-schema-from CRD-NAME",
	Short: "Generate a Promise from a given Kubernetes Operator.",
	Long:  `Generate a Promise from a given Kubernetes Operator.`,
	Args:  cobra.ExactArgs(1),
	RunE:  InitPromiseFromOperator,
}

var (
	operatorManifestsDir, targetCrdName string
)

func init() {
	initCmd.AddCommand(operatorPromiseCmd)

	operatorPromiseCmd.Flags().StringVarP(&operatorManifestsDir, "operator-manifests", "m", "", "The path to the directory containing the operator manifests.")
	operatorPromiseCmd.Flags().StringVarP(&targetCrdName, "api-schema-from", "a", "", "The name of the CRD which the Promise API schema should be generated from.")

	operatorPromiseCmd.MarkFlagRequired("operator-manifests")
	operatorPromiseCmd.MarkFlagRequired("api-schema-from")
}

func InitPromiseFromOperator(cmd *cobra.Command, args []string) error {
	promiseName := args[0]

	if plural == "" {
		plural = fmt.Sprintf("%ss", strings.ToLower(kind))
	}

	dependencies, err := buildDependencies(operatorManifestsDir)
	if err != nil {
		return err
	}

	crd, err := findTargetCRD(targetCrdName, dependencies)
	if err != nil {
		return err
	}

	if len(crd.Spec.Versions) == 0 {
		return fmt.Errorf("no versions found in CRD")
	}

	names := apiextensionsv1.CustomResourceDefinitionNames{
		Plural:   plural,
		Singular: strings.ToLower(kind),
		Kind:     kind,
	}

	storedVersionIdx := findStoredVersionIdx(crd)
	operatorVersion := crd.Spec.Versions[storedVersionIdx].Name
	envs := []corev1.EnvVar{
		{
			Name:  "OPERATOR_GROUP",
			Value: crd.Spec.Group,
		},
		{
			Name:  "OPERATOR_VERSION",
			Value: operatorVersion,
		},
		{
			Name:  "OPERATOR_KIND",
			Value: crd.Spec.Names.Kind,
		},
	}
	updateOperatorCrd(crd, storedVersionIdx, group, names, version)

	exampleResource := &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": fmt.Sprintf("%s/%s", crd.Spec.Group, crd.Spec.Versions[0].Name),
			"kind":       kind,
			"metadata": map[string]any{
				"name":      "example-database",
				"namespace": "default",
			},
			"spec": topLevelRequiredFields(crd),
		},
	}

	pipelines := generateResourceConfigurePipelines(operatorContainerName, operatorContainerImage, envs)

	flags := fmt.Sprintf("--operator-manifests %s --api-schema-from %s", operatorManifestsDir, targetCrdName)
	filesToWrite, err := getFilesToWrite(promiseName, split, workflowDirectory, flags, nil, dependencies, crd, pipelines, exampleResource)
	if err != nil {
		return err
	}

	err = writePromiseFiles(outputDir, filesToWrite)
	if err != nil {
		return err
	}

	fmt.Println("Promise generated successfully.")
	fmt.Println("The Operator documents were added as inline dependencies in the Promise Spec.")
	fmt.Println("You can move them to a workflow by running:")
	fmt.Printf("\tkratix update dependencies %s --image yourorg/your-image:tag\n", operatorManifestsDir)
	return nil
}

func findTargetCRD(crdName string, dependencies []v1alpha1.Dependency) (*apiextensionsv1.CustomResourceDefinition, error) {
	var crd *apiextensionsv1.CustomResourceDefinition
	for _, dep := range dependencies {
		if dep.GetKind() == "CustomResourceDefinition" && dep.GetName() == crdName {
			crdAsBytes, err := json.Marshal(dep.Object)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal CRD: %w", err)
			}
			crd = &apiextensionsv1.CustomResourceDefinition{}
			if err := json.Unmarshal(crdAsBytes, crd); err != nil {
				return nil, fmt.Errorf("failed to unmarshal CRD: %w", err)
			}
			break
		}
	}
	if crd == nil {
		return nil, fmt.Errorf("no CRD found matching name: %s", targetCrdName)
	}
	return crd, nil
}

func findStoredVersionIdx(crd *apiextensionsv1.CustomResourceDefinition) int {
	var storedVersionIdx int
	for idx, crdVersion := range crd.Spec.Versions {
		if crdVersion.Storage {
			storedVersionIdx = idx
			break
		}
	}

	return storedVersionIdx
}

func updateOperatorCrd(crd *apiextensionsv1.CustomResourceDefinition, storedVersionIdx int, group string, names apiextensionsv1.CustomResourceDefinitionNames, version string) {
	crd.Spec.Names = names
	crd.Name = fmt.Sprintf("%s.%s", names.Plural, group)
	crd.Spec.Group = group

	storedVersion := crd.Spec.Versions[storedVersionIdx]

	if version == "" {
		version = storedVersion.Name
	}

	storedVersion.Name = version
	storedVersion.Storage = true
	storedVersion.Served = true
	storedVersion.Schema.OpenAPIV3Schema.Properties["kind"] = apiextensionsv1.JSONSchemaProps{
		Type: "string",
		Enum: []apiextensionsv1.JSON{{Raw: []byte(fmt.Sprintf("%q", kind))}},
	}
	storedVersion.Schema.OpenAPIV3Schema.Properties["apiVersion"] = apiextensionsv1.JSONSchemaProps{
		Type: "string",
		Enum: []apiextensionsv1.JSON{{Raw: []byte(fmt.Sprintf(`"%s/%s"`, group, version))}},
	}
	crd.Spec.Versions = []apiextensionsv1.CustomResourceDefinitionVersion{
		storedVersion,
	}
}

func writePromiseFiles(outputDir string, filesToWrite map[string]any) error {
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
			return err
		}
	}

	for key, value := range filesToWrite {
		switch v := value.(type) {
		case map[string]any:
			subdir := filepath.Join(outputDir, key)
			if err := os.MkdirAll(subdir, os.ModePerm); err != nil {
				return err
			}
			if err := writePromiseFiles(subdir, v); err != nil {
				return err
			}
		default:
			fileContentBytes, err := yamlsig.Marshal(v)
			if err != nil {
				return err
			}
			if err = os.WriteFile(filepath.Join(outputDir, key), fileContentBytes, filePerm); err != nil {
				return err
			}
		}
	}
	return nil
}

func generateResourceConfigurePipelines(containerName, containerImage string, envs []corev1.EnvVar) []unstructured.Unstructured {
	container := v1alpha1.Container{
		Name:  containerName,
		Image: containerImage,
		Env:   envs,
	}

	pipeline := unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "platform.kratix.io/v1alpha1",
			"kind":       "Pipeline",
			"metadata": map[string]any{
				"name": "instance-configure",
			},
			"spec": map[string]any{
				"containers": []any{container},
			},
		},
	}

	return []unstructured.Unstructured{pipeline}
}

func topLevelRequiredFields(crd *apiextensionsv1.CustomResourceDefinition) map[string]any {
	crdSpec := crd.Spec.Versions[0].Schema.OpenAPIV3Schema.Properties["spec"]
	requiredSpecFields := crdSpec.Required
	if len(requiredSpecFields) == 0 {
		return nil
	}

	m := map[string]any{}
	for _, field := range requiredSpecFields {
		m[field] = fmt.Sprintf("# type %s", crdSpec.Properties[field].Type)
	}
	return m
}

func getFilesToWrite(promiseName string, split bool, workflowDirectory, extraFlags string, destinationSelectors []v1alpha1.PromiseScheduling, dependencies []v1alpha1.Dependency, crd *apiextensionsv1.CustomResourceDefinition, workflow []unstructured.Unstructured, exampleResource *unstructured.Unstructured) (map[string]any, error) {
	readmeTemplate, err := template.ParseFS(promiseTemplates, "templates/promise/README.md.tpl")
	if err != nil {
		return nil, err
	}

	templatedReadme := bytes.NewBuffer([]byte{})
	err = readmeTemplate.Execute(templatedReadme, promiseTemplateValues{
		SubCommand: "operator-promise",
		ExtraFlags: extraFlags,
		Name:       promiseName,
		Group:      crd.Spec.Group,
		Kind:       crd.Spec.Names.Kind,
	})

	if err != nil {
		return nil, err
	}

	if split {
		return map[string]any{
			"dependencies.yaml":     dependencies,
			"api.yaml":              crd,
			"example-resource.yaml": exampleResource,
			workflowDirectory: map[string]any{
				"workflow.yaml": workflow,
			},
			"README.md": templatedReadme.String(),
		}, nil
	}

	promise, err := generatePromise(promiseName, destinationSelectors, dependencies, crd, workflow)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"promise.yaml":          promise,
		"example-resource.yaml": exampleResource,
		"README.md":             templatedReadme.String(),
	}, nil
}

func generatePromise(promiseName string, destinationSelectors []v1alpha1.PromiseScheduling, dependencies []v1alpha1.Dependency, crd *apiextensionsv1.CustomResourceDefinition, pipelines []unstructured.Unstructured) (v1alpha1.Promise, error) {
	promise := newPromise(promiseName)

	crdBytes, err := json.Marshal(crd)
	if err != nil {
		return v1alpha1.Promise{}, err
	}

	promise.Spec.API = &runtime.RawExtension{Raw: crdBytes}
	promise.Spec.Dependencies = dependencies
	promise.Spec.Workflows.Resource.Configure = pipelines
	promise.Spec.DestinationSelectors = destinationSelectors

	return promise, nil
}
