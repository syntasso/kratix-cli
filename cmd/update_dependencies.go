package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/syntasso/kratix/api/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
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

var depsAsWorkflow bool

func init() {
	updateCmd.AddCommand(updateDependenciesCmd)
	updateDependenciesCmd.Flags().StringVarP(&dir, "dir", "d", ".", "Directory to read Promise from")
	updateDependenciesCmd.Flags().BoolVar(&depsAsWorkflow, "as-workflow", false, "Whether to include dependencies as a workflow or not; default is false")
	updateDependenciesCmd.Flags().StringVarP(&image, "image", "i", "", "The image used by this container.")
}

func updateDependencies(cmd *cobra.Command, args []string) error {
	dependenciesDir := args[0]
	if depsAsWorkflow {
		return addDepsAsWorkflow(dependenciesDir)
	}

	var depBytes []byte
	var dependencies []v1alpha1.Dependency
	file := dependencyFile()
	dependencies, err = buildDependencies(dependenciesDir)
	if err != nil {
		return err
	}

	if depBytes, err = yamlsig.Marshal(dependencies); err != nil {
		return err
	}

	switch file {
	case dependenciesFileName:
		err = os.WriteFile(filepath.Join(dir, dependenciesFileName), depBytes, filePerm)
	case promiseFileName:
		err = updatePromiseDependencies(dependencies)
	}

	if err != nil {
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
		if fileInfo.IsDir() {
			subDirDependencies, err := buildDependencies(fileName)
			if err != nil {
				return nil, err
			}
			dependencies = append(dependencies, subDirDependencies...)
			continue
		}
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
			if obj.GetNamespace() == "" {
				obj.SetNamespace("default")
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

func updatePromiseDependencies(dependencies []v1alpha1.Dependency) error {
	var promise v1alpha1.Promise
	if promise, err = getPromise(filepath.Join(dir, "promise.yaml")); err != nil {
		return err
	}
	promise.Spec.Dependencies = dependencies
	bytes, err := yamlsig.Marshal(promise)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, promiseFileName), bytes, filePerm)
}

func addDepsAsWorkflow(dependenciesDir string) error {
	containerName = "configure-deps"
	if image == "" {
		return fmt.Errorf("--image is required when --as-workflow is set")
	}
	err := generateWorkflow("promise", "configure", "dependencies", true)
	if err != nil {
		return err
	}

	workflowDir := filepath.Join(dir, "workflows/promise/configure/dependencies", containerName)
	resourcesDir := filepath.Join(workflowDir, "resources")
	scriptsDir := filepath.Join(workflowDir, "scripts")
	if err := copyFiles(dependenciesDir, resourcesDir); err != nil {
		return err
	}

	pipelineScriptContent := "#!/usr/bin/env sh\n\ncp /resources/* /kratix/output"
	if err := os.WriteFile(filepath.Join(scriptsDir, "pipeline.sh"), []byte(pipelineScriptContent), filePerm); err != nil {
		return err
	}

	file := dependencyFile()
	switch file {
	case dependenciesFileName:
		err = os.Remove(filepath.Join(dir, dependenciesFileName))
	case promiseFileName:
		err = updatePromiseDependencies([]v1alpha1.Dependency{})
	}
	if err != nil {
		return err
	}

	fmt.Println("Dependencies added as a Promise workflow.")
	fmt.Println("Don't forget to build and push the image.")
	return nil
}

func copyFiles(src, dest string) error {
	files, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, f := range files {
		if f.IsDir() {
			if err := os.Mkdir(filepath.Join(dest, f.Name()), 0755); err != nil {
				return err
			}
			if err := copyFiles(filepath.Join(src, f.Name()), filepath.Join(dest, f.Name())); err != nil {
				return err
			}
		}
		fileContents, err := os.ReadFile(filepath.Join(src, f.Name()))
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(dest, f.Name()), fileContents, 0644); err != nil {
			return err
		}
	}
	return nil
}
