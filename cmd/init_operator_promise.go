package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/syntasso/kratix/api/v1alpha1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	yamlsig "sigs.k8s.io/yaml"
)

var operatorPromiseCmd = &cobra.Command{
	Use:   "operator-promise",
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
	operatorPromiseCmd.Flags().StringVarP(&targetCrdName, "api-from", "a", "", "The name of the CRD which the Promise API should be generated from.")

	operatorPromiseCmd.MarkFlagRequired("operator-manifests")
	operatorPromiseCmd.MarkFlagRequired("api-from")
}

func InitPromiseFromOperator(cmd *cobra.Command, args []string) error {
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
	updateOperatorCrd(crd, group, names, version)

	filesToWrite := map[string]interface{}{
		"dependencies.yaml": dependencies,
		"api.yaml":          crd,
	}
	return writeOperatorPromiseFiles(filesToWrite)
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

func updateOperatorCrd(crd *apiextensionsv1.CustomResourceDefinition, group string, names apiextensionsv1.CustomResourceDefinitionNames, version string) {
	crd.Spec.Names = names
	crd.Name = fmt.Sprintf("%s.%s", names.Plural, group)
	crd.Spec.Group = group

	var storedVersionIdx int
	for idx, crdVersion := range crd.Spec.Versions {
		if crdVersion.Storage {
			storedVersionIdx = idx
			break
		}
	}

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

func writeOperatorPromiseFiles(filesToWrite map[string]interface{}) error {
	for fileName, fileContent := range filesToWrite {
		fileContentBytes, err := yamlsig.Marshal(fileContent)
		if err != nil {
			return err
		}
		if err = os.WriteFile(filepath.Join(outputDir, fileName), fileContentBytes, filePerm); err != nil {
			return err
		}
	}
	return nil
}
