/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"

	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/scheme"

	"github.com/syntasso/kratix/api/v1alpha1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	"sigs.k8s.io/controller-runtime/pkg/client/config"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// promiseCmd represents the promise command
var promiseCmd = &cobra.Command{
	Use:   "promise",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: deletePromise,
}

var name string

func init() {
	deleteCmd.AddCommand(promiseCmd)
	promiseCmd.Flags().StringVarP(&name, "name", "", "", "The name of the promise")
	promiseCmd.MarkFlagRequired("name")
	utilruntime.Must(v1alpha1.AddToScheme(scheme.Scheme))

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// promiseCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// promiseCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func deletePromise(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	fmt.Println("Deleting promise " + name)

	k8sClient, err := getControllerRuntimeClient()
	if err != nil {
		return fmt.Errorf("Error getting clientset: %v", err)
	}

	promise := &v1alpha1.Promise{}
	err = k8sClient.Get(ctx, client.ObjectKey{Name: name, Namespace: "default"}, promise)
	if err != nil {
		return fmt.Errorf("Error getting promise: %v", err)
	}

	initialFinalizers := promise.GetFinalizers()

	fmt.Printf("There are %d stages of deletion to execute before the promise is deleted\n", len(initialFinalizers))

	err = k8sClient.Delete(ctx, promise)
	if err != nil {
		return fmt.Errorf("Error deleting promise: %v", err)
	}

	return loopOnFinalizers(ctx, k8sClient, initialFinalizers)
}

func getControllerRuntimeClient() (client.Client, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("Failed to get Kubernetes config: %w", err)
	}

	// Create a new client
	k8sClient, err := client.New(cfg, client.Options{
		Scheme: scheme.Scheme,
	})

	if err != nil {
		return nil, fmt.Errorf("Failed to create Kubernetes client: $w", err)
	}
	return k8sClient, nil
}
