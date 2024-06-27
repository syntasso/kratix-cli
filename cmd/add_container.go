package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	"github.com/syntasso/kratix/api/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/yaml"
)

var addContainerCmd = &cobra.Command{
	Use:   "container LIFECYCLE/ACTION/PIPELINE-NAME --image CONTAINER-IMAGE",
	Short: "Adds a container to the named workflow",
	Example: `  # LIFECYCLE is one of: promise, resource
  # ACTION is one of: configure, delete

  # add a new resource configure container to pipeline 'instance'
  kratix add container resource/configure/instance --image syntasso/postgres-resource:v1.0.0

  # add a new promise configure container to pipeline 'pipeline0', with the container name 'deploy-deps'
  kratix add container promise/configure/pipeline0 --image syntasso/postgres-resource:v1.0.0 --name deploy-deps`,
	RunE: AddContainer,
	Args: cobra.ExactArgs(1),
}

var image, containerName string

func init() {
	addCmd.AddCommand(addContainerCmd)
	addContainerCmd.Flags().StringVarP(&image, "image", "i", "", "The image used by this container.")
	addContainerCmd.Flags().StringVarP(&containerName, "containerName", "n", "", "The container name used for this container.")
	addContainerCmd.Flags().StringVarP(&dir, "dir", "d", ".", "Directory to read promise.yaml from. Default to current working directory.")
	addContainerCmd.MarkFlagRequired("image")
}

func AddContainer(cmd *cobra.Command, args []string) error {
	if containerName == "" {
		containerName = generateContainerName(image)
	}

	pipelineInput := args[0]
	pipelineParts := strings.Split(pipelineInput, "/")
	workflow, action, pipelineName := pipelineParts[0], pipelineParts[1], pipelineParts[2]

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

	allPipelines, err := promise.GeneratePipelines(logr.Logger{})
	if err != nil {
		return err
	}

	container := v1alpha1.Container{
		Name:  containerName,
		Image: image,
	}

	pipelines, pipelineIndex, err := findPipelinesForWorkflowAction(workflow, action, pipelineName, allPipelines)
	if err != nil {
		return err
	}

	var pipelinesUnstructured []unstructured.Unstructured
	if pipelineIndex != -1 {
		pipelines[pipelineIndex].Spec.Containers = append(pipelines[pipelineIndex].Spec.Containers, container)
		pipelinesUnstructured, err = pipelinesToUnstructured(pipelines)
		if err != nil {
			return err
		}

	} else {
		pipeline := unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "platform.kratix.io/v1alpha1",
				"kind":       "Pipeline",
				"metadata": map[string]interface{}{
					"name": pipelineName,
				},
				"spec": map[string]interface{}{
					"containers": []interface{}{container},
				},
			},
		}
		pipelinesUnstructured, err = pipelinesToUnstructured(pipelines)
		if err != nil {
			return err
		}

		pipelinesUnstructured = append(pipelinesUnstructured, pipeline)
	}

	updatePipeline(workflow, action, pipelinesUnstructured, &promise)

	promiseBytes, err = yaml.Marshal(promise)
	if err != nil {
		return err
	}

	err = os.WriteFile(promiseFilePath, promiseBytes, filePerm)
	if err != nil {
		return err
	}
	fmt.Printf("generated the %s/%s/%s/%s \n", workflow, action, pipelineName, containerName)

	pipelineScriptFilename := "pipeline.sh"
	generatePipelineDirFiles(dir, workflow, action, pipelineName)
	fmt.Printf("Customise your container by editing the workflows/%s/%s/%s/scripts/%s \n", workflow, action, pipelineName, pipelineScriptFilename)
	fmt.Println("Don't forget to build and push your image!")
	return nil
}

func generateContainerName(image string) string {
	nameAndVersion := strings.ReplaceAll(image, "/", "-")
	return strings.Split(nameAndVersion, ":")[0]
}

func findPipelinesForWorkflowAction(workflow, action, pipelineName string, allPipelines v1alpha1.PromisePipelines) ([]v1alpha1.Pipeline, int, error) {
	var pipelines []v1alpha1.Pipeline
	switch workflow {
	case "promise":
		switch action {
		case "configure":
			pipelines = allPipelines.ConfigurePromise
		case "delete":
			pipelines = allPipelines.DeletePromise
		default:
			return nil, -1, fmt.Errorf("invalid action: %s", action)
		}
	case "resource":
		switch action {
		case "configure":
			pipelines = allPipelines.ConfigureResource
		case "delete":
			pipelines = allPipelines.DeleteResource
		default:
			return nil, -1, fmt.Errorf("invalid action: %s", action)
		}
	default:
		return nil, -1, fmt.Errorf("invalid workflow: %s", workflow)
	}

	for i, p := range pipelines {
		if p.Name == pipelineName {
			return pipelines, i, nil
		}
	}
	return pipelines, -1, nil
}

func updatePipeline(workflow, action string, pipelines []unstructured.Unstructured, promise *v1alpha1.Promise) {
	switch workflow {
	case "promise":
		switch action {
		case "configure":
			promise.Spec.Workflows.Promise.Configure = pipelines
		case "delete":
			promise.Spec.Workflows.Promise.Delete = pipelines
		}
	case "resource":
		switch action {
		case "configure":
			promise.Spec.Workflows.Resource.Configure = pipelines
		case "delete":
			promise.Spec.Workflows.Resource.Delete = pipelines
		}
	}
}

func pipelinesToUnstructured(pipelines []v1alpha1.Pipeline) ([]unstructured.Unstructured, error) {
	var pipelinesUnstructured []unstructured.Unstructured
	for _, p := range pipelines {
		pMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&p)
		if err != nil {
			return nil, err
		}
		pMap["kind"] = "Pipeline"
		pMap["apiVersion"] = "platform.kratix.io/v1alpha1"
		pUnstructured := unstructured.Unstructured{Object: pMap}
		pipelinesUnstructured = append(pipelinesUnstructured, pUnstructured)
	}
	return pipelinesUnstructured, nil
}

func generatePipelineDirFiles(dir, workflow, action, pipelineName string) error {
	pipelineScriptContents := []byte(`#!/usr/bin/env sh

	set -xe
	
	name="$(yq eval '.metadata.name' /kratix/input/object.yaml)"
	namespace=$(yq '.metadata.namespace' /kratix/input/object.yaml)
	
	echo "Hello from ${name} ${namespace}"`)

	pipelineScriptFilename := "pipeline.sh"
	pipelineFileDirectory := fmt.Sprintf("%s/workflows/%s/%s/%s/", dir, workflow, action, pipelineName)
	pipelineScriptDirectory := fmt.Sprintf("%s/workflows/%s/%s/%s/scripts/", dir, workflow, action, pipelineName)
	err := os.MkdirAll(pipelineScriptDirectory, os.ModePerm)
	if err != nil {
		return err
	}

	err = os.WriteFile(pipelineScriptDirectory+pipelineScriptFilename, pipelineScriptContents, filePerm)
	if err != nil {
		return err
	}

	_, err = os.Create(pipelineFileDirectory + "Dockerfile")
	if err != nil {
		return err
	}

	if _, err := os.Stat(pipelineFileDirectory + "resources/"); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(pipelineFileDirectory+"resources/", os.ModePerm)
		if err != nil {
			return err
		}
	}

	return nil
}
