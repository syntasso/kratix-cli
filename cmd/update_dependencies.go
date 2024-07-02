package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/syntasso/kratix/api/v1alpha1"
	"io"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
	"os"
	"path/filepath"
	yamlsig "sigs.k8s.io/yaml"
)

var updateDependenciesCmd = &cobra.Command{
	Use:   "dependencies",
	Short: "Commands to update promise dependencies",
	Long:  "Commands to update promise dependencies",
	Example: ` # update promise dependencies with files in 'local-dir'
Kratix update dependencies local-dir/ `,
	Args: cobra.ExactArgs(1),
	RunE: updateDependencies,
}

func init() {
	updateCmd.AddCommand(updateDependenciesCmd)
	updateDependenciesCmd.Flags().StringVarP(&dir, "dir", "d", ".", "Directory to read Promise from")
}

func updateDependencies(cmd *cobra.Command, args []string) error {
	dependenciesDir := args[0]
	dependencies, err := buildDependencies(dependenciesDir)
	if err != nil {
		return err
	}

	var depBytes []byte
	if depBytes, err = yamlsig.Marshal(dependencies); err != nil {
		return err
	}

	var bytes []byte
	file := dependencyFile()
	switch file {
	case dependenciesFileName:
		bytes = depBytes
	case promiseFileName:
		var promise v1alpha1.Promise
		if promise, err = getPromise(filepath.Join(dir, "promise.yaml")); err != nil {
			return err
		}
		promise.Spec.Dependencies = dependencies
		bytes, err = yamlsig.Marshal(promise)
		if err != nil {
			return err
		}
	}

	if err = os.WriteFile(filepath.Join(dir, file), bytes, filePerm); err != nil {
		return err
	}

	fmt.Printf("Updated %s\n", file)
	return nil
}

func dependencyFile() string {
	_, err := os.Stat(filepath.Join(dir, dependenciesFileName))
	if _, promiseErr := os.Stat(filepath.Join(dir, promiseFileName)); os.IsNotExist(err) && promiseErr == nil {
		return promiseFileName
	}
	return dependenciesFileName
}

func buildDependencies(dependenciesDir string) ([]v1alpha1.Dependency, error) {
	files, err := os.ReadDir(dependenciesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read dependency directory: %s", dependenciesDir)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no files found in directory: %s; nothing to update", dependenciesDir)
	}

	var dependencies []v1alpha1.Dependency
	var dependencyIgnored bool
	for _, fileInfo := range files {
		fileName := filepath.Join(dependenciesDir, fileInfo.Name())
		var file *os.File
		if file, err = os.Open(fileName); err != nil {
			return nil, fmt.Errorf("failed to open dependency file: %s", fileName)
		}

		decoder := yaml.NewYAMLOrJSONDecoder(file, 2048)
		for {
			var obj *unstructured.Unstructured
			if err = decoder.Decode(&obj); err == io.EOF {
				break
			} else if err != nil {
				dependencyIgnored = true
				continue
			}
			dependencies = append(dependencies, v1alpha1.Dependency{Unstructured: *obj})
		}
	}

	if len(dependencies) == 0 {
		return nil, fmt.Errorf("no valid dependencies found in directory: %s", dependenciesDir)
	}

	if dependencyIgnored {
		fmt.Println("Skipped invalid yaml documents during dependency writing")
	}

	return dependencies, nil
}

func getPromise(filePath string) (v1alpha1.Promise, error) {
	var promiseBytes []byte
	var err error
	if promiseBytes, err = os.ReadFile(filePath); err != nil {
		return v1alpha1.Promise{}, err
	}

	var promise v1alpha1.Promise
	if err = yaml.Unmarshal(promiseBytes, &promise); err != nil {
		return v1alpha1.Promise{}, err
	}
	return promise, nil
}
