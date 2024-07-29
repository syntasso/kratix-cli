package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"github.com/syntasso/kratix/api/v1alpha1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/yaml"
)

var updateAPICmd = &cobra.Command{
	Use:   "api --property PROPERTY-NAME:TYPE",
	Short: "Command to update promise API",
	Long:  "Command to update promise API",
	Example: `  # add a new property of type string to the API
  kratix update api --property region:string

  # add an integer 'port' property nested into a 'service' object
  kratix update api --property service.port:integer
  
  # removes the property from the API
  kratix update api --property region-
  
  # updates the API group and the Kind
  kratix update api --group myorg.com --kind Database
  
  # updates the version and the plural form
  kratix update api --version v1beta3 --plural mydbs`,
	RunE: UpdateAPI,
}

var (
	dir, apiVersion string
	properties      []string
)

func init() {
	updateCmd.AddCommand(updateAPICmd)
	updateAPICmd.Flags().StringVarP(&dir, "dir", "d", ".", "Directory to read Promise from")
	updateAPICmd.Flags().StringVarP(&group, "group", "g", "", "The API group for the Promise")
	updateAPICmd.Flags().StringVarP(&kind, "kind", "k", "", "The kind to be provided by the Promise")
	updateAPICmd.Flags().StringVarP(&apiVersion, "version", "v", "", "The group version for the Promise")
	updateAPICmd.Flags().StringVar(&plural, "plural", "", "The plural form of the kind")
	updateAPICmd.Flags().StringArrayVarP(&properties, "property", "p", []string{}, "Property of the Promise API to update")
}

func UpdateAPI(cmd *cobra.Command, args []string) error {
	var crd apiextensionsv1.CustomResourceDefinition
	var promise v1alpha1.Promise

	var splitFile bool
	filePath := filepath.Join(dir, apiFileName)
	if _, foundErr := os.Stat(filePath); foundErr == nil {
		splitFile = true
		apiBytes, err := os.ReadFile(filePath)
		if err = yaml.Unmarshal(apiBytes, &crd); err != nil {
			return err
		}
	} else {
		filePath = filepath.Join(dir, promiseFileName)
		promiseBytes, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to find %s or %s in directory. Please run 'kratix init promise' first: %s", apiFileName, promiseFileName, err)
		}
		if err = yaml.Unmarshal(promiseBytes, &promise); err != nil {
			return err
		}
		if err = yaml.Unmarshal(promise.Spec.API.Raw, &crd); err != nil {
			return err
		}
	}

	jsonBytes, err := updateCRDBytes(&crd)
	if err != nil {
		return err
	}

	apiContents := &runtime.RawExtension{Raw: jsonBytes}
	var data interface{} = apiContents
	if !splitFile {
		promise.Spec.API = apiContents
		data = promise
	}

	bytes, err := yaml.Marshal(data)
	if err != nil {
		return err
	}
	if err = os.WriteFile(filePath, bytes, filePerm); err != nil {
		return err
	}

	fmt.Println("Promise api updated")
	return nil
}

func updateCRDBytes(crd *apiextensionsv1.CustomResourceDefinition) ([]byte, error) {
	if gvkNeedsUpdate() {
		updateGVK(crd)
		if err := updateExampleResource(crd); err != nil {
			return nil, err
		}
	}

	if crd.Spec.Versions[0].Schema.OpenAPIV3Schema.Properties["spec"].Properties == nil {
		crd.Spec.Versions[0].Schema.OpenAPIV3Schema.Properties["spec"] = apiextensionsv1.JSONSchemaProps{
			Type:       "object",
			Properties: map[string]apiextensionsv1.JSONSchemaProps{},
		}
	}

	specProperties := crd.Spec.Versions[0].Schema.OpenAPIV3Schema.Properties["spec"].Properties

	if len(properties) != 0 {
		for _, prop := range properties {
			parsedProps := strings.Split(prop, ":")
			if len(parsedProps) != 2 {
				if prop[len(prop)-1:] != "-" {
					return nil, fmt.Errorf("invalid property format: %s", prop)
				}
				p := strings.TrimRight(prop, "-")

				nestedFields := strings.Split(p, ".")
				curr := specProperties
				for i := 0; i < len(nestedFields)-1; i++ {
					if curr[nestedFields[i]].Properties == nil {
						curr = nil
						break
					}
					curr = curr[nestedFields[i]].Properties
				}
				if curr != nil {
					delete(curr, nestedFields[len(nestedFields)-1])
				}
				continue
			}

			propNames := strings.Split(parsedProps[0], ".")
			propType := parsedProps[1]

			if !slices.Contains([]string{"string", "number", "integer", "object"}, propType) {
				return nil, fmt.Errorf("unsupported property type: %s", propType)
			}

			curr := specProperties
			lastProp := len(propNames) - 1
			for i := 0; i < lastProp; i++ {
				if curr[propNames[i]].Properties == nil {
					curr[propNames[i]] = apiextensionsv1.JSONSchemaProps{
						Type:       "object",
						Properties: map[string]apiextensionsv1.JSONSchemaProps{},
					}
				}
				if curr[propNames[i]].Type != "object" {
					return nil, fmt.Errorf("nested field %s is not an object", propNames[i])
				}

				curr = curr[propNames[i]].Properties
			}
			curr[propNames[lastProp]] = apiextensionsv1.JSONSchemaProps{
				Type: propType,
			}
		}
	}
	return json.Marshal(crd)
}

func gvkNeedsUpdate() bool {
	if apiVersion != "" || kind != "" || group != "" || plural != "" {
		return true
	}
	return false
}

func updateGVK(crd *apiextensionsv1.CustomResourceDefinition) {
	if kind != "" {
		crd.Spec.Names.Kind = kind
		crd.Spec.Names.Singular = strings.ToLower(kind)
	}

	if apiVersion != "" {
		crd.Spec.Versions[0].Name = apiVersion
	}

	if group != "" {
		crd.Spec.Group = group
	}

	if plural != "" {
		crd.Spec.Names.Plural = plural
	}
	crd.Name = fmt.Sprintf("%s.%s", crd.Spec.Names.Plural, crd.Spec.Group)
}

func updateExampleResource(crd *apiextensionsv1.CustomResourceDefinition) error {
	rrFilePath := filepath.Join(dir, resourceFileName)
	rrBytes, err := os.ReadFile(rrFilePath)
	if err != nil {
		return err
	}

	var rr unstructured.Unstructured
	if err = yaml.Unmarshal(rrBytes, &rr); err != nil {
		return err
	}
	rr.Object["apiVersion"] = fmt.Sprintf("%s/%s", crd.Spec.Group, crd.Spec.Versions[0].Name)
	rr.Object["kind"] = crd.Spec.Names.Kind
	updatedRR, err := yaml.Marshal(rr.Object)
	if err != nil {
		return err
	}
	if err = os.WriteFile(rrFilePath, updatedRR, filePerm); err != nil {
		return err
	}
	fmt.Println("Example resource updated")
	return nil
}
