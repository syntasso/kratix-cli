package cmd

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/meta"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/restmapper"
)

// reuse the global configFlags defined elsewhere
var _ = configFlags

const (
	lblPromiseName = "kratix.io/component-of-promise-name"
	lblResName     = "kratix.io/component-of-resource-name"
	lblResNS       = "kratix.io/component-of-resource-namespace"
)

func clientsFromFlags(cf *genericclioptions.ConfigFlags) (dynamic.Interface, discovery.DiscoveryInterface, meta.RESTMapper, error) {
	cfg, err := cf.ToRESTConfig()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("load REST config: %w", err)
	}
	dyn, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("dynamic client: %w", err)
	}
	disco, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("discovery client: %w", err)
	}
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(disco))
	return dyn, disco, mapper, nil
}

func gvrFor(mapper meta.RESTMapper, resource string) (schema.GroupVersionResource, error) {
	return mapper.ResourceFor(schema.GroupVersionResource{Resource: resource})
}

func listAllNamespacedResources(disco discovery.DiscoveryInterface) ([]schema.GroupVersionResource, error) {
	lists, err := disco.ServerPreferredNamespacedResources()
	if err != nil && !discovery.IsGroupDiscoveryFailedError(err) {
		return nil, err
	}
	var out []schema.GroupVersionResource
	for _, rl := range lists {
		gv, err := schema.ParseGroupVersion(rl.GroupVersion)
		if err != nil {
			continue
		}
		for i := range rl.APIResources {
			r := rl.APIResources[i]
			if strings.Contains(r.Name, "/") {
				continue // skip subresources like foos/status
			}
			hasList := false
			for _, v := range r.Verbs {
				if v == "list" {
					hasList = true
					break
				}
			}
			if !hasList {
				continue
			}
			out = append(out, gv.WithResource(r.Name))
		}
	}
	return out, nil
}

func renderTree(compoundResource string) error {
	ctx := context.Background()
	dyn, disco, mapper, err := clientsFromFlags(configFlags)
	if err != nil {
		return err
	}

	// 1) resolve the compound resource (e.g. "paved-path-demo") to its GVR
	compoundGVR, err := gvrFor(mapper, compoundResource)
	if err != nil {
		return fmt.Errorf("resolve resource %q: %w", compoundResource, err)
	}

	// 2) list all parent requests of that resource across namespaces
	parents, err := dyn.Resource(compoundGVR).Namespace(v1.NamespaceAll).List(ctx, v1.ListOptions{})
	if err != nil {
		return fmt.Errorf("list %q requests: %w", compoundResource, err)
	}
	if len(parents.Items) == 0 {
		fmt.Printf("No requests found for compound promise %q.\n", compoundResource)
		return nil
	}
	sort.Slice(parents.Items, func(i, j int) bool {
		ai := parents.Items[i].GetNamespace() + "/" + parents.Items[i].GetName()
		aj := parents.Items[j].GetNamespace() + "/" + parents.Items[j].GetName()
		return ai < aj
	})

	// 3) discover all namespaced resources to search for children
	namespacedGVRs, err := listAllNamespacedResources(disco)
	if err != nil {
		return fmt.Errorf("discover namespaced resources: %w", err)
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("kratix platform get resources %s\n", compoundResource))

	for _, parent := range parents.Items {
		pName := parent.GetName()
		pNS := parent.GetNamespace()
		if pNS == "" {
			pNS = "default"
		}
		b.WriteString(fmt.Sprintf("  - %s\n", pName))

		selector := fmt.Sprintf("%s=%s,%s=%s,%s=%s",
			lblPromiseName, compoundResource,
			lblResName, pName,
			lblResNS, pNS,
		)

		type child struct {
			Kind string
			Name string
		}
		var children []child

		// scan every namespaced resource and collect items with the linking labels
		for _, gvr := range namespacedGVRs {
			list, err := dyn.Resource(gvr).Namespace(v1.NamespaceAll).List(ctx, v1.ListOptions{LabelSelector: selector})
			if err != nil {
				continue // ignore resources we can't list
			}
			for i := range list.Items {
				// include resource kind prefix if you want: gvr.Resource + "/"
				children = append(children, child{Kind: gvr.Resource, Name: list.Items[i].GetName()})
			}
		}

		sort.Slice(children, func(i, j int) bool {
			ai := children[i].Kind + "/" + children[i].Name
			aj := children[j].Kind + "/" + children[j].Name
			return ai < aj
		})

		for _, c := range children {
			b.WriteString("    |\n")
			// choose one: with kind prefix, or just name
			// b.WriteString(fmt.Sprintf("    |--%s/%s\n", c.Kind, c.Name))
			b.WriteString(fmt.Sprintf("    |--%s\n", c.Name))
		}
	}

	fmt.Print(b.String())
	return nil
}

// kratix platform get resources <compound-resource>
// e.g. `./bin/kratix platform get resources paved-path-demo`
var platformGetResourcesCmd = &cobra.Command{
	Use:   "resources <compound-resource>",
	Short: "Show requests for a compound Promise and its labeled sub-requests",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return renderTree(args[0])
	},
}

func init() {
	platformGetCmd.AddCommand(platformGetResourcesCmd)
}
