package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/syntasso/kratix/api/v1alpha1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"os"
	"path/filepath"
	"sigs.k8s.io/yaml"
	"strings"
)

var updateAPICmd = &cobra.Command{
	Use:   "api --property PROPERTY-NAME:TYPE",
	Short: "Command to update promise API",
	Long:  "Command to update promise API",
	Example: `  # add a new property of type string to the API
  kratix update api --property region:string
  # removes the property from the API 
  kratix update api --property region-
  # updates the API group and the Kind
  kratix update api --group myorg.com --kind Database
  # updates the version and the plural form
  kratix update api --version v1beta3 --plural mydbs`,
	RunE: UpdateAPI,
}

var (
	dir        string
	properties []string
)

func init() {
	updateCmd.AddCommand(updateAPICmd)
	updateAPICmd.Flags().StringVarP(&dir, "dir", "d", ".", "Directory to read Promise from")
	updateAPICmd.Flags().StringVarP(&group, "group", "g", "", "The API group for the Promise")
	updateAPICmd.Flags().StringVarP(&kind, "kind", "k", "", "The kind to be provided by the Promise")
	updateAPICmd.Flags().StringVarP(&version, "version", "v", "v1alpha1", "The group version for the Promise")
	updateAPICmd.Flags().StringVar(&plural, "plural", "", "The plural form of the kind")
	updateAPICmd.Flags().StringArrayVarP(&properties, "property", "p", []string{}, "Property of the Promise API to update")
}

func UpdateAPI(cmd *cobra.Command, args []string) error {
	promiseFilePath := filepath.Join(dir, "promise.yaml")
	promiseBytes, err := os.ReadFile(promiseFilePath)
	if err != nil {
		return err
	}

	var promise v1alpha1.Promise
	err = yaml.Unmarshal(promiseBytes, &promise)
	if err != nil {
		return err
	}

	var crd apiextensionsv1.CustomResourceDefinition
	err = yaml.Unmarshal(promise.Spec.API.Raw, &crd)
	if err != nil {
		return err
	}

	updateGVK(&crd)

	if len(properties) != 0 {
		for _, prop := range properties {
			parsedProps := strings.Split(prop, ":")
			if len(parsedProps) != 2 {
				if prop[len(prop)-1:] != "-" {
					return fmt.Errorf("invalid property format: %s", prop)
				}
				p := strings.TrimRight(prop, "-")
				delete(crd.Spec.Versions[0].Schema.OpenAPIV3Schema.Properties["spec"].Properties, p)
				continue
			}

			propName := parsedProps[0]
			propType := parsedProps[1]

			if propType != "string" && propType != "number" {
				return fmt.Errorf("unsupported property type: %s", propType)
			}
			if crd.Spec.Versions[0].Schema.OpenAPIV3Schema.Properties["spec"].Properties == nil {
				crd.Spec.Versions[0].Schema.OpenAPIV3Schema.Properties["spec"] = apiextensionsv1.JSONSchemaProps{
					Type:       "object",
					Properties: map[string]apiextensionsv1.JSONSchemaProps{},
				}
			}
			crd.Spec.Versions[0].Schema.OpenAPIV3Schema.Properties["spec"].Properties[propName] = apiextensionsv1.JSONSchemaProps{
				Type: propType,
			}
		}
	}

	crdBytes, err := json.Marshal(crd)
	if err != nil {
		return err
	}

	apiContents := &runtime.RawExtension{Raw: crdBytes}
	promise.Spec.API = apiContents

	promiseBytes, err = yaml.Marshal(promise)
	if err != nil {
		return err
	}

	err = os.WriteFile(promiseFilePath, promiseBytes, filePerm)
	if err != nil {
		return err
	}
	fmt.Println("Promise updated")
	return nil
}

func updateGVK(crd *apiextensionsv1.CustomResourceDefinition) {
	if kind != "" {
		crd.Spec.Names.Kind = kind
		crd.Spec.Names.Singular = strings.ToLower(kind)
	}

	if version != "" {
		crd.Spec.Versions[0].Name = version
	}

	if group != "" {
		crd.Spec.Group = group
	}

	if plural != "" {
		crd.Spec.Names.Plural = plural
	}

	crd.Name = fmt.Sprintf("%s.%s", crd.Spec.Names.Plural, crd.Spec.Group)
}
