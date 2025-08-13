package cmd

import (
	"context"
	"crypto/sha1"
	"embed"
	"encoding/hex"
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

var supportedLanguages = []string{"go", "bash", "python"}

var languageMatrix = map[string]interface{}{
	"go": map[string]string{
		"fileExtension":       "go",
		"templateDir":         "go",
		"confirmationMessage": "run 'go mod init' and 'go mod tidy' to manage your script's dependencies",
	},
	"bash": map[string]string{
		"fileExtension": "sh",
		"templateDir":   ".",
	},
	"python": map[string]string{
		"fileExtension":       "py",
		"templateDir":         "python",
		"confirmationMessage": "",
	},
}

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

var image, containerName, language string

func init() {
	addCmd.AddCommand(addContainerCmd)
	addContainerCmd.Flags().StringVarP(&image, "image", "i", "", "The image used by this container.")
	addContainerCmd.Flags().StringVarP(&containerName, "name", "n", "", "The container name used for this container.")
	addContainerCmd.Flags().StringVarP(&dir, "dir", "d", ".", "Directory to read promise.yaml from. Default to current working directory.")
	addContainerCmd.Flags().StringVarP(&language, "language", "l", "bash", "Language to use for the scripting of the pipeline script, defaults to bash. Currently supports Bash, Go and Python.")
	addContainerCmd.MarkFlagRequired("image")
}

func AddContainer(cmd *cobra.Command, args []string) error {
	if containerName == "" {
		containerName = generateContainerName(image)
	}

	if !supportedLanguage(language) {
		return fmt.Errorf("invalid language: %s is not supported by the kratix cli", language)
	}

	pipelineInput := args[0]
	containerArgs, err := ParseContainerCmdArgs(pipelineInput)
	if err != nil {
		return err
	}

	if err := generateWorkflow(containerArgs, containerName, image, false); err != nil {
		return err
	}

	pipelineScriptFilename := pipelineScriptFilename(language)
	scriptsPath := filepath.Join("workflows", containerArgs.Lifecycle, containerArgs.Action, containerArgs.Pipeline, containerName, "scripts", pipelineScriptFilename)
	logConfirmationMessages(scriptsPath, language)
	return nil
}

func generateWorkflow(c *ContainerCmdArgs, containerName, image string, overwrite bool) error {
	if c.Lifecycle != "promise" && c.Lifecycle != "resource" {
		return fmt.Errorf("invalid lifecycle: %s, expected one of: promise, resource", c.Lifecycle)
	}

	if c.Action != "configure" && c.Action != "delete" {
		return fmt.Errorf("invalid action: %s, expected one of: configure, delete", c.Action)
	}

	if c.Pipeline == "" {
		return fmt.Errorf("pipeline name cannot be empty")
	}

	container := v1alpha1.Container{
		Name:  containerName,
		Image: image,
	}

	workflowPath := filepath.Join("workflows", c.Lifecycle, c.Action)
	var promise v1alpha1.Promise

	splitFiles := filesGeneratedWithSplit(dir)

	var filePath string
	if splitFiles {
		filePath = filepath.Join(dir, workflowPath, "workflow.yaml")
	} else {
		filePath = filepath.Join(dir, "promise.yaml")
	}

	var pipelines []v1alpha1.Pipeline
	var pipelineIdx = -1
	var fileBytes []byte
	var err error
	if splitFiles && workflowFileFound(filePath) {
		fileBytes, err = os.ReadFile(filePath)
		if err != nil {
			return err
		}
		yaml.Unmarshal(fileBytes, &pipelines)

		pipelineIdx, err = getPipelineIdx(pipelines, c.Pipeline)
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

		pipelines, pipelineIdx, err = findPipelinesForLifecycleAction(c, allPipelines)
		if err != nil {
			return err
		}
	}

	var pipelinesUnstructured []unstructured.Unstructured
	if pipelineIdx != -1 {
		containerIdx := getContainerIdx(pipelines[pipelineIdx], container.Name)
		if containerIdx == -1 {
			pipelines[pipelineIdx].Spec.Containers = append(pipelines[pipelineIdx].Spec.Containers, container)
		} else {
			if !overwrite {
				return fmt.Errorf("image '%s' already exists in Pipeline '%s'", container.Name, c.Pipeline)
			}
			pipelines[pipelineIdx].Spec.Containers[containerIdx] = container
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
					"name": c.Pipeline,
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
		fileBytes, err = yaml.Marshal(pipelinesUnstructured)
		if err != nil {
			return err
		}
	} else {
		updatePipeline(c.Lifecycle, c.Action, pipelinesUnstructured, &promise)

		fileBytes, err = yaml.Marshal(promise)
		if err != nil {
			return err
		}
	}
	if err := generatePipelineDirFiles(dir, workflowPath, c.Pipeline, containerName, language); err != nil {
		return err
	}

	if err := os.WriteFile(filePath, fileBytes, filePerm); err != nil {
		return err
	}
	fmt.Printf("generated the %s/%s/%s/%s in %s \n", c.Lifecycle, c.Action, c.Pipeline, containerName, filePath)

	return nil
}

func generateContainerName(image string) string {
	name := strings.Split(image, ":")[0]
	name = strings.NewReplacer("/", "-", ".", "-").Replace(name)
	name = strings.Trim(name, "-")

	if len(name) <= 63 {
		return name
	}

	hash := sha1.Sum([]byte(name))
	suffix := hex.EncodeToString(hash[:])[:7]
	prefix := strings.TrimRight(name[:63-len(suffix)-1], "-")
	return fmt.Sprintf("%s-%s", prefix, suffix)
}

func findPipelinesForLifecycleAction(c *ContainerCmdArgs, allPipelines map[v1alpha1.Type]map[v1alpha1.Action][]v1alpha1.Pipeline) ([]v1alpha1.Pipeline, int, error) {
	var pipelines []v1alpha1.Pipeline
	switch c.Lifecycle {
	case "promise":
		switch c.Action {
		case "configure":
			pipelines = allPipelines[v1alpha1.WorkflowTypePromise][v1alpha1.WorkflowActionConfigure]
		case "delete":
			pipelines = allPipelines[v1alpha1.WorkflowTypePromise][v1alpha1.WorkflowActionDelete]
		}
	case "resource":
		switch c.Action {
		case "configure":
			pipelines = allPipelines[v1alpha1.WorkflowTypeResource][v1alpha1.WorkflowActionConfigure]
		case "delete":
			pipelines = allPipelines[v1alpha1.WorkflowTypeResource][v1alpha1.WorkflowActionDelete]
		}
	}

	idx, err := getPipelineIdx(pipelines, c.Pipeline)
	if err != nil {
		return nil, -1, err
	}

	return pipelines, idx, nil
}

func updatePipeline(lifecycle, action string, pipelines []unstructured.Unstructured, promise *v1alpha1.Promise) {
	switch lifecycle {
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

func generatePipelineDirFiles(promiseDir, workflowDirectory, pipelineName, containerName, language string) error {
	containerFileDirectory := filepath.Join(workflowDirectory, pipelineName, containerName)
	containerScriptsDirectory := filepath.Join(containerFileDirectory, "scripts")
	resourcesDir := filepath.Join(promiseDir, containerFileDirectory, "resources")

	templates := getTemplates(containerFileDirectory, containerScriptsDirectory, language)

	if err := templateFiles(workflowTemplates, promiseDir, templates, nil); err != nil {
		return err
	}
	if _, err := os.Stat(resourcesDir); errors.Is(err, os.ErrNotExist) {
		return os.Mkdir(resourcesDir, os.ModePerm)
	}
	return nil
}

func pipelineScriptFilename(language string) string {
	extension := languageMatrix[language].(map[string]string)["fileExtension"]
	return fmt.Sprintf("pipeline.%s", extension)
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

func getPipelineIdx(pipelines []v1alpha1.Pipeline, pipelineName string) (int, error) {
	for idx, p := range pipelines {
		if p.GetName() == pipelineName {
			return idx, nil
		}
	}

	return -1, nil
}

func getContainerIdx(pipeline v1alpha1.Pipeline, containerName string) int {
	for i, container := range pipeline.Spec.Containers {
		if container.Name == containerName {
			return i
		}
	}
	return -1
}

func supportedLanguage(language string) bool {
	for _, sl := range supportedLanguages {
		if sl == language {
			return true
		}
	}
	return false
}

func getTemplates(containerFileDirectory, containerScriptsDirectory, language string) map[string]string {
	pipelineScriptFilename := pipelineScriptFilename(language)
	pipelineScriptTemplateFilepath := fmt.Sprintf("templates/workflows/%s/%s.tpl", language, pipelineScriptFilename)
	dockerfileTemplateFilepath := fmt.Sprintf("templates/workflows/%s/Dockerfile.tpl", language)

	templates := map[string]string{
		filepath.Join(containerScriptsDirectory, pipelineScriptFilename): pipelineScriptTemplateFilepath,
		filepath.Join(containerFileDirectory, "Dockerfile"):              dockerfileTemplateFilepath,
	}
	return templates
}

func logConfirmationMessages(scriptsPath, language string) {
	fmt.Printf("Customise your container by editing %s \n", scriptsPath)
	if language == "go" {
		fmt.Printf("For %s containers, %s\n", language, languageMatrix[language].(map[string]string)["confirmationMessage"])
	}
	fmt.Println("Don't forget to build and push your image!")
}
