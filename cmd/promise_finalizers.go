package cmd

import (
	"context"
	"slices"
	"time"

	"fmt"

	"github.com/syntasso/kratix/api/v1alpha1"
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
)

type Finalizer struct {
	name       string
	handleFunc func(ctx context.Context, k8sClient client.Client) error
}

var (
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
)

func loopOnFinalizers(ctx context.Context, k8sClient client.Client, _ []string) error {
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

	fmt.Println("Promise deleted")
	return nil
}

func handleRunDeleteWorkflowsFinalizer(ctx context.Context, k8sClient client.Client) error {
	fmt.Printf("  - Kratix is running any Delete workflows for the Promise. You can check the status of the workflows by running: kubectl get pods -n kratix-platform-system\n")
	fmt.Printf("    This may take a few minutes, polling..")

	pollUntilFinalizersRemoved(ctx, k8sClient, runDeleteWorkflowsFinalizer, nil)
	return nil
}

func handleRemoveAllWorkflowJobsFinalizer(ctx context.Context, k8sClient client.Client) error {
	fmt.Printf("  - Kratix is deleting all Kuberntes Jobs relating to the Promise. You can check the status of the jobs by running: kubectl get jobs -n kratix-platform-system\n")
	fmt.Printf("    This is normally very quick, polling..")

	pollUntilFinalizersRemoved(ctx, k8sClient, removeAllWorkflowJobsFinalizer, nil)
	return nil
}

func handleResourceRequestCleanupFinalizer(ctx context.Context, k8sClient client.Client) error {
	fmt.Printf("  - Kratix is deleting all resource requests for the Promise. You can check the status of the resource requests by running: kubectl get resource-requests -n kratix-platform-system\n")
	fmt.Printf("    This can take a long time, polling..")

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

		fmt.Printf("\n   %d resources remaining: %v. Polling..", len(resourceList.Items), names)
	})
	return nil
}

func handleDynamicControllerDependantResourcesCleanupFinalizer(ctx context.Context, k8sClient client.Client) error {
	fmt.Printf("  - Kratix is deleting all additional Kubernetes resoures associated with the Promise\n")
	fmt.Printf("    This is normally very quick, polling..")

	pollUntilFinalizersRemoved(ctx, k8sClient, dynamicControllerDependantResourcesCleanupFinalizer, nil)
	return nil
}

func handleDependenciesCleanupFinalizer(ctx context.Context, k8sClient client.Client) error {
	fmt.Printf("  - Kratix is deleting any dependency workloads written to any StateStores\n")
	fmt.Printf("    This is normally very quick, polling..")

	pollUntilFinalizersRemoved(ctx, k8sClient, dependenciesCleanupFinalizer, nil)
	return nil
}

func handleCRDCleanupFinalizer(ctx context.Context, k8sClient client.Client) error {
	fmt.Printf("  - Kratix is deleting the CRD for the Promise\n")
	fmt.Printf("    This is normally very quick, polling..")

	pollUntilFinalizersRemoved(ctx, k8sClient, crdCleanupFinalizer, nil)
	return nil
}

func pollUntilFinalizersRemoved(ctx context.Context, k8sClient client.Client, finalizer string, runFunc func()) {
	count := 10
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

		if runFunc != nil && count == 10 {
			runFunc()
			count = 0
		}

		time.Sleep(1 * time.Second)
		count++
	}
	fmt.Printf("\n\n")
}
