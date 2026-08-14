package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// Level-1 deterministic gates, checked from the generated files themselves
// with no model involved: does the Promise parse, is its CRD valid and
// namespaced, does every example/Resource Request validate against that
// CRD's schema, do the workflow pipelines declare an image, and is the
// delete workflow well-formed (Kratix only supports one delete pipeline).

type schemaNode struct {
	Type       string                `yaml:"type"`
	Properties map[string]schemaNode `yaml:"properties"`
	Required   []string              `yaml:"required"`
	Pattern    string                `yaml:"pattern"`
}

type pipelineContainer struct {
	Name  string `yaml:"name"`
	Image string `yaml:"image"`
}

type pipelineDoc struct {
	Spec struct {
		Containers []pipelineContainer `yaml:"containers"`
	} `yaml:"spec"`
}

type promiseCRD struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Spec       struct {
		Group string `yaml:"group"`
		Scope string `yaml:"scope"`
		Names struct {
			Kind string `yaml:"kind"`
		} `yaml:"names"`
		Versions []struct {
			Name   string `yaml:"name"`
			Schema struct {
				OpenAPIV3Schema schemaNode `yaml:"openAPIV3Schema"`
			} `yaml:"schema"`
		} `yaml:"versions"`
	} `yaml:"spec"`
}

type promiseDoc struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Metadata   struct {
		Name string `yaml:"name"`
	} `yaml:"metadata"`
	Spec struct {
		API       promiseCRD `yaml:"api"`
		Workflows struct {
			Resource struct {
				Configure []pipelineDoc `yaml:"configure"`
				Delete    []pipelineDoc `yaml:"delete"`
			} `yaml:"resource"`
		} `yaml:"workflows"`
	} `yaml:"spec"`
}

type exampleDoc struct {
	APIVersion string                 `yaml:"apiVersion"`
	Kind       string                 `yaml:"kind"`
	Spec       map[string]interface{} `yaml:"spec"`
}

// checkPromiseFile parses one Promise file and returns every Level-1
// structural gate failure found in it. An empty slice means it passed.
func checkPromiseFile(name string, data []byte) ([]string, promiseDoc, error) {
	var doc promiseDoc
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, doc, fmt.Errorf("%s: does not parse as YAML: %w", name, err)
	}

	var errs []string

	if doc.APIVersion != "platform.kratix.io/v1alpha1" || doc.Kind != "Promise" {
		errs = append(errs, fmt.Sprintf("%s: not a Promise (apiVersion=%q kind=%q)", name, doc.APIVersion, doc.Kind))
	}

	crd := doc.Spec.API
	if crd.APIVersion != "apiextensions.k8s.io/v1" || crd.Kind != "CustomResourceDefinition" {
		errs = append(errs, fmt.Sprintf("%s: spec.api is not a valid CustomResourceDefinition (apiVersion=%q kind=%q)", name, crd.APIVersion, crd.Kind))
	}
	if crd.Spec.Scope != "Namespaced" {
		errs = append(errs, fmt.Sprintf("%s: CRD scope must be Namespaced, got %q (Kratix only supports Namespaced)", name, crd.Spec.Scope))
	}
	if len(crd.Spec.Versions) == 0 {
		errs = append(errs, fmt.Sprintf("%s: CRD declares no versions", name))
	}

	errs = append(errs, checkPipelines(name, "configure", doc.Spec.Workflows.Resource.Configure)...)
	errs = append(errs, checkPipelines(name, "delete", doc.Spec.Workflows.Resource.Delete)...)
	if len(doc.Spec.Workflows.Resource.Delete) > 1 {
		errs = append(errs, fmt.Sprintf("%s: delete workflow has %d pipelines, Kratix only supports one", name, len(doc.Spec.Workflows.Resource.Delete)))
	}

	return errs, doc, nil
}

func checkPipelines(promiseName, workflow string, pipelines []pipelineDoc) []string {
	var errs []string
	for _, p := range pipelines {
		for _, c := range p.Spec.Containers {
			if strings.TrimSpace(c.Image) == "" {
				errs = append(errs, fmt.Sprintf("%s: %s pipeline container %q has no image set", promiseName, workflow, c.Name))
			}
		}
	}
	return errs
}

// checkExampleFile validates one example/Resource Request file against
// whichever loaded Promise's CRD its apiVersion+kind matches. A file that
// matches no Promise is reported as such, not silently skipped.
func checkExampleFile(name string, data []byte, promises map[string]promiseDoc) []string {
	var ex exampleDoc
	if err := yaml.Unmarshal(data, &ex); err != nil {
		return []string{fmt.Sprintf("%s: does not parse as YAML: %v", name, err)}
	}

	for promiseName, doc := range promises {
		crd := doc.Spec.API
		group := crd.Spec.Group
		for _, v := range crd.Spec.Versions {
			expectedAPIVersion := group + "/" + v.Name
			if ex.APIVersion != expectedAPIVersion || ex.Kind != crd.Spec.Names.Kind {
				continue
			}
			specSchema, ok := v.Schema.OpenAPIV3Schema.Properties["spec"]
			if !ok {
				return nil
			}
			return validateAgainstSchema(name, ex.Spec, specSchema)
		}
		_ = promiseName
	}

	return []string{fmt.Sprintf("%s: apiVersion=%q kind=%q matches no loaded Promise's CRD", name, ex.APIVersion, ex.Kind)}
}

func validateAgainstSchema(name string, spec map[string]interface{}, schema schemaNode) []string {
	var errs []string

	required := make(map[string]bool, len(schema.Required))
	for _, r := range schema.Required {
		required[r] = true
	}
	for field := range required {
		if _, ok := spec[field]; !ok {
			errs = append(errs, fmt.Sprintf("%s: missing required field %q", name, field))
		}
	}

	for field, value := range spec {
		fieldSchema, known := schema.Properties[field]
		if !known {
			errs = append(errs, fmt.Sprintf("%s: unknown field %q", name, field))
			continue
		}
		if fieldSchema.Pattern != "" {
			str, isString := value.(string)
			if !isString {
				continue
			}
			matched, err := regexp.MatchString(fieldSchema.Pattern, str)
			if err == nil && !matched {
				errs = append(errs, fmt.Sprintf("%s: field %q value %q does not match pattern %q", name, field, str, fieldSchema.Pattern))
			}
		}
	}

	return errs
}

// runLevel1Gates loads every Promise in promiseDir and every example in
// exampleDir and runs all Level-1 checks. Either directory may be absent
// (returns no errors, gate is skipped) so this stays backward compatible
// with a review-file-only check.
func runLevel1Gates(promiseDir, exampleDir string, out io.Writer) []string {
	var allErrs []string

	promises := map[string]promiseDoc{}
	promiseFiles := readYAMLDir(promiseDir)
	names := make([]string, 0, len(promiseFiles))
	for name := range promiseFiles {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		errs, doc, err := checkPromiseFile(name, promiseFiles[name])
		if err != nil {
			allErrs = append(allErrs, err.Error())
			continue
		}
		allErrs = append(allErrs, errs...)
		promises[name] = doc
	}

	exampleFiles := readYAMLDir(exampleDir)
	exNames := make([]string, 0, len(exampleFiles))
	for name := range exampleFiles {
		exNames = append(exNames, name)
	}
	sort.Strings(exNames)
	for _, name := range exNames {
		allErrs = append(allErrs, checkExampleFile(name, exampleFiles[name], promises)...)
	}

	return allErrs
}

func readYAMLDir(dir string) map[string][]byte {
	files := map[string][]byte{}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return files
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := filepath.Ext(entry.Name())
		if ext != ".yaml" && ext != ".yml" {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		files[path] = data
	}
	return files
}
