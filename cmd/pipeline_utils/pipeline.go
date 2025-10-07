package pipelineutils

import (
	"context"
	"fmt"
	"io/fs"
	"strings"

	"github.com/syntasso/kratix/api/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
)

type PipelineCmdArgs struct {
	Lifecycle string
	Action    string
	Pipeline  string
}

func ParsePipelineCmdArgs(containerPath string) (*PipelineCmdArgs, error) {
	parts := strings.Split(containerPath, "/")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid pipeline format: %s, expected format: LIFECYCLE/ACTION/PIPELINE-NAME", containerPath)
	}

	return &PipelineCmdArgs{
		Lifecycle: parts[0],
		Action:    parts[1],
		Pipeline:  parts[2],
	}, nil
}

func RetrievePipeline(promise *v1alpha1.Promise, c *PipelineCmdArgs) (*v1alpha1.Pipeline, error) {
	allPipelines, err := v1alpha1.NewPipelinesMap(promise, ctrl.LoggerFrom(context.Background()))
	if err != nil {
		return nil, err
	}

	pipelines, pipelineIdx, err := findPipelinesForLifecycleAction(c, allPipelines)
	if err != nil {
		return nil, err
	}

	return &pipelines[pipelineIdx], nil
}

func findPipelinesForLifecycleAction(c *PipelineCmdArgs, allPipelines map[v1alpha1.Type]map[v1alpha1.Action][]v1alpha1.Pipeline) ([]v1alpha1.Pipeline, int, error) {
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

func FindContainerIndex(dirEntries []fs.DirEntry, containers []v1alpha1.Container, name string) (int, error) {
	if len(dirEntries) == 0 {
		return -1, fmt.Errorf("no container found in path")
	}

	if len(dirEntries) == 1 {
		if name != "" && name != dirEntries[0].Name() {
			return -1, fmt.Errorf("container %s not found in pipeline directory", name)
		}
		name = dirEntries[0].Name()
	}

	if name == "" {
		return -1, fmt.Errorf("more than one container exists for this pipeline, please provide a name with --name")
	}

	containerIndex := -1
	for i, container := range containers {
		if container.Name == name {
			containerIndex = i
			break
		}
	}
	if containerIndex == -1 {
		return -1, fmt.Errorf("container %s not found in pipeline", name)
	}

	for _, dirEntry := range dirEntries {
		if dirEntry.Name() == name {
			return containerIndex, nil
		}
	}

	return -1, fmt.Errorf("directory entry not found for container %s", name)
}

func getContainerIdx(pipeline v1alpha1.Pipeline, containerName string) int {
	for i, container := range pipeline.Spec.Containers {
		if container.Name == containerName {
			return i
		}
	}
	return -1
}

func getPipelineIdx(pipelines []v1alpha1.Pipeline, pipelineName string) (int, error) {
	for idx, p := range pipelines {
		if p.GetName() == pipelineName {
			return idx, nil
		}
	}

	return -1, nil
}

func PipelinesToUnstructured(pipelines []v1alpha1.Pipeline) ([]unstructured.Unstructured, error) {
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

// RetrievePipeline => pipelineObj
// FindContainer =>
// FindContainerDir =>
