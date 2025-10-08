package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	promiseutils "github.com/syntasso/kratix-cli/cmd/promise_utils"
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
	promise, err := promiseutils.LoadPromiseWithWorkflows(inputDir)
	if err != nil {
		return err
	}
	promise.Kind = "Promise"
	promise.APIVersion = v1alpha1.GroupVersion.Group + "/" + v1alpha1.GroupVersion.Version
	promise.Name = promiseName

	if _, err := os.Stat(filepath.Join(inputDir, apiFileName)); err == nil {
		var apiBytes []byte
		apiBytes, err = os.ReadFile(filepath.Join(inputDir, apiFileName))
		if err != nil {
			return err
		}

		if len(apiBytes) > 0 {
			var crd apiextensionsv1.CustomResourceDefinition
			err = yaml.Unmarshal(apiBytes, &crd)
			if err != nil {
				return err
			}

			var crdBytes []byte
			crdBytes, err = json.Marshal(crd)
			if err != nil {
				return err
			}

			promise.Spec.API = &runtime.RawExtension{Raw: crdBytes}
		}
	}

	if _, err := os.Stat(filepath.Join(inputDir, dependenciesFileName)); err == nil {
		var dependencyBytes []byte
		dependencyBytes, err = os.ReadFile(filepath.Join(inputDir, dependenciesFileName))
		if err != nil {
			return err
		}

		var dependencies v1alpha1.Dependencies
		err = yaml.Unmarshal(dependencyBytes, &dependencies)
		if err != nil {
			return err
		}
		promise.Spec.Dependencies = dependencies
	}

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
			Labels: map[string]string{
				"kratix.io/promise-version": "v0.0.1",
			},
		},
	}
}
