package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/syntasso/kratix/api/v1alpha1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/api/meta"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/restmapper"
)

// reuse the global configFlags defined elsewhere
var _ = configFlags

const (
	kindLabel              = "kratix.io/component-of-promise-name"
	resourceNameLabel      = "kratix.io/component-of-resource-name"
	resourceNamespaceLabel = "kratix.io/component-of-resource-namespace"
)

// kratix platform get resources <promise-kind>
// e.g. `./bin/kratix platform get resources paved-path-demo`
var platformGetResourcesCmd = &cobra.Command{
	Use:   "resources PROMISE-NAME",
	Short: "Show requests for a Promise and its labeled sub-requests",
	Long:  "Show requests for a Promise and for a Compound Promises, its sub-requests",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return renderTree(args[0])
	},
}

func init() {
	platformGetCmd.AddCommand(platformGetResourcesCmd)
}

func clientsFromFlags(cf *genericclioptions.ConfigFlags) (dynamic.Interface, discovery.DiscoveryInterface, *apiextensionsclient.Clientset, meta.RESTMapper, error) {
	cfg, err := cf.ToRESTConfig()
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("load REST config: %w", err)
	}

	k8sClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("error generating client: %w", err)
	}

	dynamicClient, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("dynamic client: %w", err)
	}

	crdClient, err := apiextensionsclient.NewForConfig(cfg)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("create CRD client: %w", err)
	}

	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(discoveryClient))
	return dynamicClient, discoveryClient, crdClient, mapper, nil
}

func gvrFor(mapper meta.RESTMapper, resource string) (schema.GroupVersionResource, error) {
	return mapper.ResourceFor(schema.GroupVersionResource{Resource: resource})
}

func listAllKratixGVRs(crdClient apiextensionsclient.Interface) ([]schema.GroupVersionResource, error) {
	ctx := context.Background()
	crds, err := crdClient.
		ApiextensionsV1().
		CustomResourceDefinitions().
		List(ctx, v1.ListOptions{LabelSelector: v1alpha1.PromiseNameLabel})
	if err != nil {
		return nil, fmt.Errorf("error listing promise CRDs: %w", err)
	}

	var out []schema.GroupVersionResource
	for _, crd := range crds.Items {
		var storageVersion string
		for _, v := range crd.Spec.Versions {
			if v.Storage {
				storageVersion = v.Name
				break
			}
		}
		if storageVersion == "" {
			continue
		}
		out = append(out, schema.GroupVersionResource{
			Group:    crd.Spec.Group,
			Version:  storageVersion,
			Resource: crd.Spec.Names.Plural,
		})
	}
	return out, nil
}

func renderTree(promiseName string) error {
	ctx := context.Background()
	dynamicClient, _, crdClient, mapper, err := clientsFromFlags(configFlags)
	if err != nil {
		return err
	}

	promise := 


	gvr, err := gvrFor(mapper, promiseName)
	if err != nil {
		return fmt.Errorf("error generating GroupVersionResource for %q, is the specified kind correct? %s", err)
	}

	promiseRequests, err := dynamicClient.Resource(gvr).Namespace(v1.NamespaceAll).List(ctx, v1.ListOptions{})
	if err != nil {
		return fmt.Errorf("list %q : %w", kind, err)
	}
	if len(promiseRequests.Items) == 0 {
		fmt.Printf("No requests found for promise %q\n", kind)
		return nil
	}

	// fetch all available CRDs installed by promises
	kratixGVRs, err := listAllKratixGVRs(crdClient)
	if err != nil {
		return fmt.Errorf("discover namespaced resources: %w", err)
	}

	var b strings.Builder

	for _, request := range promiseRequests.Items {
		requestName := request.GetName()
		requestNamespace := request.GetNamespace()
		if requestNamespace == "" {
			requestNamespace = "default"
		}

		b.WriteString(fmt.Sprintf("  - %s\n", requestName))

		selector := fmt.Sprintf("%s=%s,%s=%s,%s=%s",
			kindLabel, kind,
			resourceNameLabel, requestName,
			resourceNamespaceLabel, requestNamespace,
		)

		type subRequest struct {
			Kind string
			Name string
		}
		var subRequests []subRequest

		// scan every namespaced resource and collect items with the matching labels
		for _, gvr := range kratixGVRs {
			list, err := dynamicClient.Resource(gvr).Namespace(v1.NamespaceAll).List(ctx, v1.ListOptions{LabelSelector: selector})
			if err != nil {
				return fmt.Errorf("error listing resources %s %s: %s", gvr.Resource, err)
			}
			for i := range list.Items {
				subRequests = append(subRequests, subRequest{Kind: gvr.Resource, Name: list.Items[i].GetName()})
			}
		}

		for _, c := range subRequests {
			b.WriteString("    |\n")
			b.WriteString(fmt.Sprintf("    |--%s\n", c.Name))
		}
	}

	fmt.Print(b.String())
	return nil
}
