package promiseutils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	pipelineutils "github.com/syntasso/kratix-cli/cmd/pipeline_utils"
	"github.com/syntasso/kratix-cli/cmd/utils"
	"github.com/syntasso/kratix/api/v1alpha1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"

	"sigs.k8s.io/yaml"
)

func LoadPromiseWithWorkflows(dir string) (*v1alpha1.Promise, error) {
	var promise v1alpha1.Promise

	if _, err := os.Stat(filepath.Join(dir, "promise.yaml")); err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No promise.yaml found, assuming --split was used to initialise the Promise")
			workflows, err := LoadWorkflows(dir)
			if err != nil {
				return nil, err
			}
			promise.Spec.Workflows = workflows
			return &promise, nil
		}
		return nil, err
	}

	fileBytes, err := os.ReadFile(filepath.Join(dir, "promise.yaml"))
	if err != nil {
		return nil, err
	}

	// IMPORTANT: we need "sigs.k8s.io/yaml" for this to work.
	err = yaml.Unmarshal(fileBytes, &promise)
	if err != nil {
		return nil, err
	}

	return &promise, nil
}

func LoadPromiseWithAPI(dir string) (*v1alpha1.Promise, error) {
	var promise v1alpha1.Promise

	if _, err := os.Stat(filepath.Join(dir, "promise.yaml")); err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No promise.yaml found, assuming --split was used to initialise the Promise")
			crd, err := LoadCRD(dir)
			if err != nil {
				return nil, err
			}

			var crdBytes []byte
			crdBytes, err = json.Marshal(crd)
			if err != nil {
				return nil, err
			}

			promise.Spec.API = &runtime.RawExtension{Raw: crdBytes}
			return &promise, nil
		}
		return nil, err
	}

	fileBytes, err := os.ReadFile(filepath.Join(dir, "promise.yaml"))
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(fileBytes, &promise)
	if err != nil {
		return nil, err
	}
	return &promise, nil
}

func LoadWorkflows(dir string) (v1alpha1.Workflows, error) {
	pipelineMap := map[string]map[string][]unstructured.Unstructured{}
	var workflows v1alpha1.Workflows

	missingWorkflows := 0
	for _, lifecycle := range []string{"promise", "resource"} {
		for _, action := range []string{"configure", "delete"} {
			if utils.FileExists(filepath.Join(dir, "workflows", lifecycle, action, "workflow.yaml")) {
				workflowBytes, err := os.ReadFile(filepath.Join(dir, "workflows", lifecycle, action, "workflow.yaml"))
				if err != nil {
					return workflows, err
				}

				var workflow []v1alpha1.Pipeline
				err = yaml.Unmarshal(workflowBytes, &workflow)
				if err != nil {
					return workflows, fmt.Errorf("failed to get %s %s workflow: %s", lifecycle, action, err)
				}

				uPipelines, err := pipelineutils.PipelinesToUnstructured(workflow)
				if err != nil {
					return workflows, err
				}

				if _, ok := pipelineMap[lifecycle]; !ok {
					pipelineMap[lifecycle] = make(map[string][]unstructured.Unstructured)
				}
				pipelineMap[lifecycle][action] = uPipelines
			} else {
				missingWorkflows++
			}
		}
	}

	if _, ok := pipelineMap["promise"]; ok {
		workflows.Promise.Configure = pipelineMap["promise"]["configure"]
		workflows.Promise.Delete = pipelineMap["promise"]["delete"]
	}
	if _, ok := pipelineMap["resource"]; ok {
		workflows.Resource.Configure = pipelineMap["resource"]["configure"]
		workflows.Resource.Delete = pipelineMap["resource"]["delete"]
	}

	return workflows, nil
}

func LoadCRD(dir string) (apiextensionsv1.CustomResourceDefinition, error) {
	var crd apiextensionsv1.CustomResourceDefinition

	filePath := filepath.Join(dir, "api.yaml")

	apiBytes, err := os.ReadFile(filePath)
	if err = yaml.Unmarshal(apiBytes, &crd); err != nil {
		return apiextensionsv1.CustomResourceDefinition{}, err
	}

	return crd, nil
}
