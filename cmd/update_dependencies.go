package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	pipelineutils "github.com/syntasso/kratix-cli-plugin-investigation/cmd/pipeline_utils"
	"github.com/syntasso/kratix/api/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
	yamlsig "sigs.k8s.io/yaml"
)

var updateDependenciesCmd = &cobra.Command{
	Use:   "dependencies PATH",
	Short: "Commands to update promise dependencies",
	Long:  "Commands to update promise dependencies, by default dependencies are stored in the Promise spec.dependencies field",
	Example: `# update promise dependencies with all files in 'local-dir'
kratix update dependencies path/to/dir/

# update promise dependencies with single file
kratix update dependencies path/to/file.yaml`,
	Args: cobra.ExactArgs(1),
	RunE: updateDependencies,
}

func init() {
	updateCmd.AddCommand(updateDependenciesCmd)
	updateDependenciesCmd.Flags().StringVarP(&dir, "dir", "d", ".", "Directory to read Promise from")
	updateDependenciesCmd.Flags().StringVarP(&image, "image", "i", "", "Store dependencies to a Promise Configure workflow image with this image/tag")
}

func updateDependencies(cmd *cobra.Command, args []string) error {
	dependenciesDir := args[0]
	if image != "" {
		return addDepsAsWorkflow(dependenciesDir)
	}

	var depBytes []byte
	mode, fileToUpdate := promiseFileMode()
	dependencies, err := buildDependencies(dependenciesDir)
	if err != nil {
		return err
	}

	if depBytes, err = yamlsig.Marshal(dependencies); err != nil {
		return err
	}

	switch mode {
	case "split":
		err = os.WriteFile(filepath.Join(dir, dependenciesFileName), depBytes, filePerm)
	case "flat":
		err = updatePromiseDependencies(dependencies)
	}

	if err != nil {
		return err
	}
	fmt.Printf("Updated %s\n", fileToUpdate)
	return nil
}

func promiseFileMode() (mode string, fileToUpdate string) {
	_, dependencyFileErr := os.Stat(filepath.Join(dir, dependenciesFileName))
	if _, promiseErr := os.Stat(filepath.Join(dir, promiseFileName)); os.IsNotExist(dependencyFileErr) && promiseErr == nil {
		return "flat", promiseFileName
	}
	return "split", dependenciesFileName
}

func buildDependencies(dependenciesDir string) ([]v1alpha1.Dependency, error) {
	dependenciesDirInfo, err := os.Stat(dependenciesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to stat dependency: %s", dependenciesDir)
	}

	var dependencies []v1alpha1.Dependency
	if !dependenciesDirInfo.IsDir() {
		dependencies, err = extractDepFromFile(dependenciesDir)
		if err != nil {
			return nil, err
		}
		if len(dependencies) == 0 {
			return nil, fmt.Errorf("no valid dependencies found in directory: %s", dependenciesDir)
		}
		return dependencies, nil
	}

	files, err := os.ReadDir(dependenciesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read dependency directory: %s", dependenciesDir)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no files found in directory: %s; nothing to update", dependenciesDir)
	}

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
		if !isYAML(fileName) {
			continue
		}
		dep, err := extractDepFromFile(fileName)
		if err != nil {
			return nil, err
		}
		dependencies = append(dependencies, dep...)
	}

	if len(dependencies) == 0 {
		return nil, fmt.Errorf("no valid dependencies found in directory: %s", dependenciesDir)
	}

	return dependencies, nil
}

func extractDepFromFile(fileName string) ([]v1alpha1.Dependency, error) {
	var dependencies []v1alpha1.Dependency
	file, err := os.Open(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to open dependency file %s: %s", fileName, err)
	}

	decoder := yaml.NewYAMLOrJSONDecoder(file, 2048)
	for {
		var obj *unstructured.Unstructured
		err = decoder.Decode(&obj)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to decode dependency file %s: %s", fileName, err)
		}
		if obj == nil {
			continue
		}
		if obj.GetNamespace() == "" {
			obj.SetNamespace("default")
		}
		dependencies = append(dependencies, v1alpha1.Dependency{Unstructured: *obj})
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
	promise, err := getPromise(filepath.Join(dir, "promise.yaml"))
	if err != nil {
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
	c := &pipelineutils.PipelineCmdArgs{
		Lifecycle: "promise",
		Action:    "configure",
		Pipeline:  "dependencies",
	}

	err := generateWorkflow(c, containerName, image, true)
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

	mode, _ := promiseFileMode()
	switch mode {
	case "split":
		err = os.Remove(filepath.Join(dir, dependenciesFileName))
	case "flat":
		err = updatePromiseDependencies([]v1alpha1.Dependency{})
	}
	if err != nil {
		return err
	}

	fmt.Println("Dependencies added as a Promise workflow.")
	fmt.Println("Run the following command to build the dependencies image:")
	fmt.Printf("\n  docker build -t %s %s\n\n", image, workflowDir)
	fmt.Println("Don't forget to push the image to a registry!")
	return nil
}

func copyFiles(src, dest string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if srcInfo.Mode().IsDir() {
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

				continue
			}

			if err := writeToFile(filepath.Join(src, f.Name()), dest, f.Name()); err != nil {
				return err
			}
		}
	} else if srcInfo.Mode().IsRegular() {
		fileName := filepath.Base(src)
		return writeToFile(src, dest, fileName)
	} else {
		return errors.New("unsupported type for dependencies: must be file or directory")
	}

	return nil
}

func writeToFile(src string, dest string, fileName string) error {
	fileContents, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dest, fileName), fileContents, 0644); err != nil {
		return err
	}
	return nil
}

func isYAML(fileName string) bool {
	return filepath.Ext(fileName) == ".yaml" || filepath.Ext(fileName) == ".yml"
}
