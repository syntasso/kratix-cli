package cmd

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/syntasso/kratix/api/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/yaml"
)

//go:embed templates/workflows/*
var workflowTemplates embed.FS

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

var (
	image, containerName string
	container            v1alpha1.Container
	pipelineIndex        = -1
	workflowTrigger      v1alpha1.Workflows
)

func init() {
	addCmd.AddCommand(addContainerCmd)
	addContainerCmd.Flags().StringVarP(&image, "image", "i", "", "The image used by this container.")
	addContainerCmd.Flags().StringVarP(&containerName, "name", "n", "", "The container name used for this container.")
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

	if err := generateWorkflow(workflow, action, pipelineName, containerName, image, false); err != nil {
		return err
	}

	pipelineScriptFilename := "pipeline.sh"
	scriptsPath := filepath.Join("workflows", workflow, action, pipelineName, containerName, "scripts", pipelineScriptFilename)
	fmt.Printf("Customise your container by editing %s \n", scriptsPath)
	fmt.Println("Don't forget to build and push your image!")
	return nil
}

func generateWorkflow(workflow, action, pipelineName, containerName, image string, overwrite bool) error {
	container = v1alpha1.Container{
		Name:  containerName,
		Image: image,
	}

	var workflowPath = filepath.Join("workflows", workflow, action)
	var filePath string
	var fileBytes []byte
	var promise v1alpha1.Promise

	splitFiles := filesGeneratedWithSplit(dir)

	var pipelines []v1alpha1.Pipeline
	if splitFiles {
		filePath = filepath.Join(dir, workflowPath, "workflow.yaml")
	} else {
		filePath = filepath.Join(dir, "promise.yaml")
	}

	var err error
	if splitFiles && workflowFileFound(filePath) {
		fileBytes, err = os.ReadFile(filePath)
		if err != nil {
			return err
		}
		yaml.Unmarshal(fileBytes, &workflowTrigger)

		pipelines, pipelineIndex, err = getPipelinesFromWorkflowYaml(workflowTrigger, workflow, action, pipelineName)
		if err != nil {
			return err
		}
	}

	if !splitFiles {
		fileBytes, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}

		err = yaml.Unmarshal(fileBytes, &promise)
		if err != nil {
			return err
		}

		allPipelines, err := v1alpha1.NewPipelinesMap(&promise, ctrl.LoggerFrom(context.Background()))
		if err != nil {
			return err
		}

		pipelines, pipelineIndex, err = findPipelinesForWorkflowAction(workflow, action, pipelineName, allPipelines)
		if err != nil {
			return err
		}
	}

	var pipelinesUnstructured []unstructured.Unstructured
	if pipelineIndex != -1 {
		containerIdx := containerIndex(pipelines[pipelineIndex], container.Name)
		if containerIdx == -1 {
			pipelines[pipelineIndex].Spec.Containers = append(pipelines[pipelineIndex].Spec.Containers, container)
		} else {
			if !overwrite {
				return fmt.Errorf("image '%s' already exists in Pipeline", container.Name)
			}
			pipelines[pipelineIndex].Spec.Containers[containerIdx] = container
		}

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

	if splitFiles {
		updateWorkflow(workflow, action, pipelinesUnstructured, &workflowTrigger)
		fileBytes, err = yaml.Marshal(workflowTrigger)
		if err != nil {
			return err
		}
	} else {
		updatePipeline(workflow, action, pipelinesUnstructured, &promise)

		fileBytes, err = yaml.Marshal(promise)
		if err != nil {
			return err
		}
	}
	if err := generatePipelineDirFiles(dir, workflowPath, pipelineName, containerName); err != nil {
		return err
	}

	if err := os.WriteFile(filePath, fileBytes, filePerm); err != nil {
		return err
	}
	fmt.Printf("generated the %s/%s/%s/%s in %s \n", workflow, action, pipelineName, containerName, filePath)

	return nil
}

func generateContainerName(image string) string {
	nameAndVersion := strings.ReplaceAll(image, "/", "-")
	return strings.Split(nameAndVersion, ":")[0]
}

func findPipelinesForWorkflowAction(workflow, action, pipelineName string, allPipelines map[v1alpha1.Type]map[v1alpha1.Action][]v1alpha1.Pipeline) ([]v1alpha1.Pipeline, int, error) {
	var pipelines []v1alpha1.Pipeline
	switch workflow {
	case "promise":
		switch action {
		case "configure":
			pipelines = allPipelines[v1alpha1.WorkflowTypePromise][v1alpha1.WorkflowActionConfigure]
		case "delete":
			pipelines = allPipelines[v1alpha1.WorkflowTypePromise][v1alpha1.WorkflowActionDelete]
		default:
			return nil, -1, fmt.Errorf("invalid action: %s", action)
		}
	case "resource":
		switch action {
		case "configure":
			pipelines = allPipelines[v1alpha1.WorkflowTypeResource][v1alpha1.WorkflowActionConfigure]
		case "delete":
			pipelines = allPipelines[v1alpha1.WorkflowTypeResource][v1alpha1.WorkflowActionDelete]
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

func generatePipelineDirFiles(promiseDir, workflowDirectory, pipelineName, containerName string) error {
	containerFileDirectory := filepath.Join(workflowDirectory, pipelineName, containerName)
	containerScriptsDirectory := filepath.Join(containerFileDirectory, "scripts")
	resourcesDir := filepath.Join(promiseDir, containerFileDirectory, "resources")

	templates := map[string]string{
		filepath.Join(containerScriptsDirectory, "pipeline.sh"): "templates/workflows/pipeline.sh.tpl",
		filepath.Join(containerFileDirectory, "Dockerfile"):     "templates/workflows/Dockerfile.tpl",
	}

	if err := templateFiles(workflowTemplates, promiseDir, templates, nil); err != nil {
		return err
	}
	if _, err := os.Stat(resourcesDir); errors.Is(err, os.ErrNotExist) {
		return os.Mkdir(resourcesDir, os.ModePerm)
	}
	return nil
}

func filesGeneratedWithSplit(dir string) bool {
	if _, err := os.Stat(dir + "/api.yaml"); errors.Is(err, os.ErrNotExist) {
		return false
	}

	if _, err := os.Stat(dir + "/dependencies.yaml"); errors.Is(err, os.ErrNotExist) {
		return false
	}

	return true
}

func workflowFileFound(workflowFilePath string) bool {
	if _, err := os.Stat(workflowFilePath); errors.Is(err, os.ErrNotExist) {
		return false
	}
	return true
}

func getPipelinesFromWorkflowYaml(workflow v1alpha1.Workflows, lifecycle string, action string, pipelineName string) (pipelines []v1alpha1.Pipeline, index int, err error) {
	var unstructuredWorkflowPipelines []unstructured.Unstructured
	switch lifecycle {
	case "promise":
		switch action {
		case "configure":
			unstructuredWorkflowPipelines = workflow.Promise.Configure
		case "delete":
			unstructuredWorkflowPipelines = workflow.Promise.Delete
		}
	case "resource":
		switch action {
		case "configure":
			unstructuredWorkflowPipelines = workflow.Resource.Configure
		case "delete":
			unstructuredWorkflowPipelines = workflow.Resource.Delete
		}
	}

	for index, p := range unstructuredWorkflowPipelines {
		if p.GetName() == pipelineName {
			workflowPipelines, err := v1alpha1.PipelinesFromUnstructured(unstructuredWorkflowPipelines, ctrl.LoggerFrom(context.Background()))
			if err != nil {
				return []v1alpha1.Pipeline{}, index, err
			}
			return workflowPipelines, index, nil
		}
	}

	workflowPipelines, err := v1alpha1.PipelinesFromUnstructured(unstructuredWorkflowPipelines, ctrl.LoggerFrom(context.Background()))
	if err != nil {
		return []v1alpha1.Pipeline{}, index, err
	}
	return workflowPipelines, -1, nil
}

func updateWorkflow(workflow, action string, pipelines []unstructured.Unstructured, workflowTrigger *v1alpha1.Workflows) {
	switch workflow {
	case "promise":
		switch action {
		case "configure":
			workflowTrigger.Promise.Configure = pipelines
		case "delete":
			workflowTrigger.Promise.Delete = pipelines
		}
	case "resource":
		switch action {
		case "configure":
			workflowTrigger.Resource.Configure = pipelines
		case "delete":
			workflowTrigger.Resource.Delete = pipelines
		}
	}
}

func containerIndex(pipeline v1alpha1.Pipeline, containerName string) int {
	for i, container := range pipeline.Spec.Containers {
		if container.Name == containerName {
			return i
		}
	}
	return -1
}
