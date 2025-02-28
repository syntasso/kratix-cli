package cmd

import (
	"context"
	"slices"
	"time"

	"fmt"

	"github.com/syntasso/kratix/api/v1alpha1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	resourceRequestCleanupFinalizer                     = v1alpha1.KratixPrefix + "resource-request-cleanup"
	dynamicControllerDependantResourcesCleanupFinalizer = v1alpha1.KratixPrefix + "dynamic-controller-dependant-resources-cleanup"
	crdCleanupFinalizer                                 = v1alpha1.KratixPrefix + "api-crd-cleanup"
	dependenciesCleanupFinalizer                        = v1alpha1.KratixPrefix + "dependencies-cleanup"
	removeAllWorkflowJobsFinalizer                      = v1alpha1.KratixPrefix + "workflows-cleanup"
	runDeleteWorkflowsFinalizer                         = v1alpha1.KratixPrefix + "delete-workflows"

	green = "\033[32m"
	reset = "\033[0m"
)

type Finalizer struct {
	name       string
	handleFunc func(ctx context.Context, k8sClient client.Client) error
}

var (
	promiseFinalizersOrderedByExecution = []Finalizer{}
)

func init() {
	promiseFinalizersOrderedByExecution = []Finalizer{
		{
			name:       runDeleteWorkflowsFinalizer,
			handleFunc: handleRunDeleteWorkflowsFinalizer,
		},
		{
			name:       removeAllWorkflowJobsFinalizer,
			handleFunc: handleRemoveAllWorkflowJobsFinalizer,
		},
		{
			name:       resourceRequestCleanupFinalizer,
			handleFunc: handleResourceRequestCleanupFinalizer,
		},
		{
			name:       dynamicControllerDependantResourcesCleanupFinalizer,
			handleFunc: handleDynamicControllerDependantResourcesCleanupFinalizer,
		},
		{
			name:       dependenciesCleanupFinalizer,
			handleFunc: handleDependenciesCleanupFinalizer,
		},
		{
			name:       crdCleanupFinalizer,
			handleFunc: handleCRDCleanupFinalizer,
		},
	}
}

func loopOnFinalizers(ctx context.Context, k8sClient client.Client) error {
	promise := &v1alpha1.Promise{}
	err := k8sClient.Get(ctx, client.ObjectKey{Name: name}, promise)
	if err != nil {
		//if because the promise is deleted, then return nil
		if client.IgnoreNotFound(err) == nil {
			return nil
		}
	}

	if len(promise.GetFinalizers()) == 0 {
		return nil
	}

	if promise.GetDeletionTimestamp() == nil {
		return fmt.Errorf("Promise is not marked as being deleted")
	}

	for _, finalizerFunc := range promiseFinalizersOrderedByExecution {
		for _, finalizer := range promise.GetFinalizers() {
			if finalizer == finalizerFunc.name {
				err := finalizerFunc.handleFunc(ctx, k8sClient)
				if err != nil {
					return fmt.Errorf("Error executing finalizer %s: %v", finalizer, err)
				}
			}
		}
	}

	fmt.Println("✅ Promise successfully deleted.")
	return nil
}

func handleRunDeleteWorkflowsFinalizer(ctx context.Context, k8sClient client.Client) error {
	index := getIndex(runDeleteWorkflowsFinalizer)
	fmt.Printf("[%s%d/6%s] Delete workflow in progress..", green, index, reset)

	pollUntilFinalizersRemoved(ctx, k8sClient, runDeleteWorkflowsFinalizer, func() {
		labelSelector := client.MatchingLabels{
			"kratix.io/promise-name":    name,
			"kratix.io/workflow-action": "delete",
			"kratix.io/workflow-type":   "promise",
		}

		pods := &corev1.PodList{}
		err := k8sClient.List(context.TODO(), pods, labelSelector)
		if err != nil {
			fmt.Println("Error listing pods:", err)
			return
		}

		if len(pods.Items) == 0 {
			return
		}

		fmt.Printf("\n    Delete workflow Pod %s/%s still in-flight, status: %v..", pods.Items[0].Namespace, pods.Items[0].Name, pods.Items[0].Status.Phase)
	})
	return nil
}

func handleRemoveAllWorkflowJobsFinalizer(ctx context.Context, k8sClient client.Client) error {
	index := getIndex(removeAllWorkflowJobsFinalizer)
	fmt.Printf("[%s%d/6%s] Workflow cleanup in progress..", green, index, reset)

	pollUntilFinalizersRemoved(ctx, k8sClient, removeAllWorkflowJobsFinalizer, func() {
		labelSelector := client.MatchingLabels{
			"kratix.io/promise-name":    name,
			"kratix.io/workflow-action": "delete",
			"kratix.io/workflow-type":   "promise",
		}

		jobs := &batchv1.JobList{}
		err := k8sClient.List(context.TODO(), jobs, labelSelector)
		if err != nil {
			fmt.Println("Error listing pods:", err)
			return
		}

		if len(jobs.Items) == 0 {
			return
		}

		names := []string{}
		for _, resource := range jobs.Items {
			names = append(names, resource.GetNamespace()+"/"+resource.GetName())
		}

		fmt.Printf("\n      %d Remaining Jobs: %v. Polling..", len(jobs.Items), names)
	})
	return nil
}

func handleResourceRequestCleanupFinalizer(ctx context.Context, k8sClient client.Client) error {
	index := getIndex(resourceRequestCleanupFinalizer)
	fmt.Printf("[%s%d/6%s] Resource request cleanup in progress...", green, index, reset)

	promise := &v1alpha1.Promise{}
	err := k8sClient.Get(ctx, client.ObjectKey{Name: name}, promise)
	if err != nil {
		return fmt.Errorf("Error getting promise: %v", err)
	}

	gvk, _, err := promise.GetAPI()
	if err != nil {
		return fmt.Errorf("Error getting API: %v", err)
	}

	pollUntilFinalizersRemoved(ctx, k8sClient, resourceRequestCleanupFinalizer, func() {
		resourceList := &unstructured.UnstructuredList{}
		resourceList.SetGroupVersionKind(*gvk)
		err = k8sClient.List(ctx, resourceList)
		if err != nil {
			fmt.Printf("Error listing resources: %v\n", err)
		}

		names := []string{}
		for _, resource := range resourceList.Items {
			names = append(names, resource.GetNamespace()+"/"+resource.GetName())
		}

		fmt.Printf("\n      %d Remaining resource requests: %v..", len(resourceList.Items), names)
	})
	return nil
}

func handleDynamicControllerDependantResourcesCleanupFinalizer(ctx context.Context, k8sClient client.Client) error {
	index := getIndex(dynamicControllerDependantResourcesCleanupFinalizer)
	fmt.Printf("[%s%d/6%s] Additional Kubernetes Resources cleanup in progress..", green, index, reset)

	pollUntilFinalizersRemoved(ctx, k8sClient, dynamicControllerDependantResourcesCleanupFinalizer, nil)
	return nil
}

func handleDependenciesCleanupFinalizer(ctx context.Context, k8sClient client.Client) error {
	index := getIndex(dependenciesCleanupFinalizer)
	fmt.Printf("[%s%d/6%s] Dependency Workloads cleanup in progress..", green, index, reset)

	pollUntilFinalizersRemoved(ctx, k8sClient, dependenciesCleanupFinalizer, nil)
	return nil
}

func handleCRDCleanupFinalizer(ctx context.Context, k8sClient client.Client) error {
	index := getIndex(crdCleanupFinalizer)
	fmt.Printf("[%s%d/6%s] Deleting CRD..", green, index, reset)

	pollUntilFinalizersRemoved(ctx, k8sClient, crdCleanupFinalizer, nil)
	return nil
}

func pollUntilFinalizersRemoved(ctx context.Context, k8sClient client.Client, finalizer string, runFunc func()) {
	count := 5
	for {
		fmt.Printf(".")
		promise := &v1alpha1.Promise{}
		err := k8sClient.Get(ctx, client.ObjectKey{Name: name}, promise)
		if err != nil {
			//if because the promise is deleted, then return nil
			if client.IgnoreNotFound(err) == nil {
				break
			}
		}

		if !slices.Contains(promise.GetFinalizers(), finalizer) {
			break
		}

		if runFunc != nil && count == 5 {
			runFunc()
			count = 0
		}

		time.Sleep(1 * time.Second)
		count++
	}
	fmt.Printf("\n\n\n")
}

func getIndex(finalizer string) int {
	for i, f := range promiseFinalizersOrderedByExecution {
		if f.name == finalizer {
			return i + 1
		}
	}
	return -1
}
