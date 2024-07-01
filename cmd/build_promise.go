package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/syntasso/kratix/api/v1alpha1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/yaml"
)

var buildPromiseCmd = &cobra.Command{
	Use:   "promise PROMISE-NAME",
	Short: "Command to build a Kratix Promise",
	Long:  "Command to build a Kratix Promise from given api, dependencies and workflow files in a directory. Use this command if you initialized your Promise with `--split`.",
	Example: `  # build a promise from path
  kratix build promise postgresql --dir ~/path/to/promise-bundle/`,
	Args: cobra.ExactArgs(1),
	RunE: BuildPromise,
}

var inputDir, outputPath string

func init() {
	buildCmd.AddCommand(buildPromiseCmd)
	buildPromiseCmd.Flags().StringVarP(&inputDir, "dir", "d", ".", "Directory to build promise from. Default to the current working directory")
	buildPromiseCmd.Flags().StringVarP(&outputPath, "output", "o", "", "File path to write promise to. Default to output to stdout")
}

func BuildPromise(cmd *cobra.Command, args []string) error {
	promiseName := args[0]
	promise := newPromise(promiseName)

	apiBytes, err := os.ReadFile(filepath.Join(inputDir, "api.yaml"))
	if err != nil {
		return err
	}

	var crd apiextensionsv1.CustomResourceDefinition
	err = yaml.Unmarshal(apiBytes, &crd)
	if err != nil {
		return err
	}

	crdBytes, err := json.Marshal(crd)
	if err != nil {
		return err
	}

	apiContents := &runtime.RawExtension{Raw: crdBytes}
	promise.Spec.API = apiContents

	var workflows v1alpha1.Workflows
	buildWorkflows(&workflows)
	promise.Spec.Workflows = workflows

	promiseBytes, err := yaml.Marshal(promise)
	if err != nil {
		return err
	}

	if outputPath != "" {
		return os.WriteFile(outputPath, promiseBytes, filePerm)
	}

	fmt.Println(string(promiseBytes))
	return nil
}

func newPromise(promiseName string) v1alpha1.Promise {
	return v1alpha1.Promise{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Promise",
			APIVersion: v1alpha1.GroupVersion.Group + "/" + v1alpha1.GroupVersion.Version,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: promiseName,
		},
	}
}

func fileExists(filePath string) bool {
	if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) {
		return false
	}
	return true
}

func buildWorkflows(workflows *v1alpha1.Workflows) error {
	promiseConfigure, err := getWorkflow("promise", "configure")
	if err != nil {
		return err
	}

	promiseDelete, err := getWorkflow("promise", "delete")
	if err != nil {
		return err
	}

	resourceConfigure, err := getWorkflow("resource", "configure")
	if err != nil {
		return err
	}

	resourceDelete, err := getWorkflow("resource", "delete")
	if err != nil {
		return err
	}

	workflows.Promise.Configure = promiseConfigure.Promise.Configure
	workflows.Promise.Delete = promiseDelete.Promise.Delete
	workflows.Resource.Configure = resourceConfigure.Resource.Configure
	workflows.Resource.Delete = resourceDelete.Resource.Delete
	return nil
}

func getWorkflow(lifecyle, action string) (workflow v1alpha1.Workflows, err error) {
	if fileExists(filepath.Join(inputDir, "workflows", lifecyle, action, "workflow.yaml")) {
		workflowBytes, err := os.ReadFile(filepath.Join(inputDir, "workflows", lifecyle, action, "workflow.yaml"))
		if err != nil {
			return v1alpha1.Workflows{}, err
		}

		var configurePromiseWorkflow v1alpha1.Workflows
		err = yaml.Unmarshal(workflowBytes, &configurePromiseWorkflow)
		if err != nil {
			return v1alpha1.Workflows{}, err
		}
		return configurePromiseWorkflow, nil
	}
	return v1alpha1.Workflows{}, err
}
