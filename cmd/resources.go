/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// platformGetResourcesCmd represents the get resources command
var platformGetResourcesCmd = &cobra.Command{
	Use:   "resources",
	Short: "A command to fetch resource requests of a given Promise",
	Long: `A command to fetch resource requests of a given Promise
	
	For Compound Promise, it details all of the requests that make up a Compound request.`,
	RunE: GetResources,
}

func GetResources(cmd *cobra.Command, args []string) error {
	restConfig, err := configFlags.ToRESTConfig()
	if err != nil {
		return err
	}

	// k8s client setup with kubectl flags
	client, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return err
	}

	fmt.Print(client.AppsV1().Deployments("kratix-platform-system").List(context.TODO(), v1.ListOptions{}))

	return nil
}

func init() {
	platformGetCmd.AddCommand(platformGetResourcesCmd)
}
