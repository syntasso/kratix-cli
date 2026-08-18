package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/go-logr/logr"
	"github.com/syntasso/kratix/api/v1alpha1"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsvalidation "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/validation"
	sigsyaml "sigs.k8s.io/yaml"
)

// Level-1 deterministic gates, checked from the generated files themselves
// with no model involved: does the Promise parse, is its CRD valid and
// namespaced, does every example/Resource Request validate against that
// CRD's schema, do the workflow pipelines declare an image, and is the
// delete workflow well-formed (Kratix only supports one delete pipeline).
//
// These gates decode into the real Kratix Promise type
// (github.com/syntasso/kratix/api/v1alpha1) and the real Kubernetes CRD
// type (k8s.io/apiextensions-apiserver), and validate the CRD with the same
// validator the Kubernetes API server itself uses, rather than hand-rolled
// structs that only understand a handful of fields.

// exampleDoc is the one type here with no existing Kubernetes/Kratix
// equivalent: an example/Resource Request's shape is whatever the matching
// Promise's CRD schema says it should be, so it's decoded generically.
type exampleDoc struct {
	APIVersion string                 `yaml:"apiVersion"`
	Kind       string                 `yaml:"kind"`
	Spec       map[string]interface{} `yaml:"spec"`
}

// checkPromiseFile parses one Promise file and returns every Level-1
// structural gate failure found in it. An empty slice means it passed.
func checkPromiseFile(name string, data []byte) ([]string, *v1alpha1.Promise, error) {
	var promise v1alpha1.Promise
	if err := sigsyaml.Unmarshal(data, &promise); err != nil {
		return nil, nil, fmt.Errorf("%s: does not parse as YAML: %w", name, err)
	}

	var errs []string

	if promise.APIVersion != "platform.kratix.io/v1alpha1" || promise.Kind != "Promise" {
		errs = append(errs, fmt.Sprintf("%s: not a Promise (apiVersion=%q kind=%q)", name, promise.APIVersion, promise.Kind))
	}

	if promise.Spec.API == nil {
		errs = append(errs, fmt.Sprintf("%s: spec.api is not set", name))
		return errs, &promise, nil
	}

	_, crdErrs := decodeAndValidateCRD(name, promise.Spec.API.Raw)
	errs = append(errs, crdErrs...)

	errs = append(errs, checkWorkflowPipelines(name, &promise)...)

	return errs, &promise, nil
}

// decodeAndValidateCRD decodes spec.api into the real Kubernetes
// CustomResourceDefinition type and validates it exactly as the Kubernetes
// API server would: every required CRD field (metadata.name, spec.group,
// names.plural, version served/storage flags, a structurally valid schema,
// ...), not just a handful of top-level fields.
func decodeAndValidateCRD(promiseName string, raw []byte) (apiextensionsv1.CustomResourceDefinition, []string) {
	var crd apiextensionsv1.CustomResourceDefinition
	if err := json.Unmarshal(raw, &crd); err != nil {
		return crd, []string{fmt.Sprintf("%s: spec.api does not parse as a CustomResourceDefinition: %v", promiseName, err)}
	}

	// The real Kubernetes API server defaults fields such as names.listKind,
	// names.singular and status.storedVersions before validating a newly
	// created CRD; apply the same defaulting here so this gate only reports
	// fields a Promise author actually has to set themselves.
	apiextensionsv1.SetDefaults_CustomResourceDefinition(&crd)

	var errs []string
	if crd.APIVersion != "apiextensions.k8s.io/v1" || crd.Kind != "CustomResourceDefinition" {
		errs = append(errs, fmt.Sprintf("%s: spec.api is not a valid CustomResourceDefinition (apiVersion=%q kind=%q)", promiseName, crd.APIVersion, crd.Kind))
	}

	// Kratix itself only supports Namespaced-scope Promises; this is a
	// Kratix rule, not something the generic Kubernetes CRD validator
	// checks (Cluster scope is a perfectly valid CRD to Kubernetes).
	if crd.Spec.Scope != apiextensionsv1.NamespaceScoped {
		errs = append(errs, fmt.Sprintf("%s: CRD scope must be Namespaced, got %q (Kratix only supports Namespaced)", promiseName, crd.Spec.Scope))
	}

	var internalCRD apiextensions.CustomResourceDefinition
	if err := apiextensionsv1.Convert_v1_CustomResourceDefinition_To_apiextensions_CustomResourceDefinition(&crd, &internalCRD, nil); err != nil {
		errs = append(errs, fmt.Sprintf("%s: failed to convert CRD for validation: %v", promiseName, err))
		return crd, errs
	}

	for _, fe := range apiextensionsvalidation.ValidateCustomResourceDefinition(context.Background(), &internalCRD) {
		errs = append(errs, fmt.Sprintf("%s: CRD %s", promiseName, fe.Error()))
	}

	return crd, errs
}

// checkWorkflowPipelines checks every pipeline container in every workflow
// slot - promise.configure, promise.delete, resource.configure and
// resource.delete - using Kratix's own unstructured-to-Pipeline conversion
// (v1alpha1.NewPipelinesMap) instead of a hand-rolled container struct that
// only looked at the resource workflows.
func checkWorkflowPipelines(promiseName string, promise *v1alpha1.Promise) []string {
	pipelinesMap, err := v1alpha1.NewPipelinesMap(promise, logr.Discard())
	if err != nil {
		return []string{fmt.Sprintf("%s: failed to parse workflow pipelines: %v", promiseName, err)}
	}

	var errs []string
	for _, workflowType := range []v1alpha1.Type{v1alpha1.WorkflowTypePromise, v1alpha1.WorkflowTypeResource} {
		for _, action := range []v1alpha1.Action{v1alpha1.WorkflowActionConfigure, v1alpha1.WorkflowActionDelete} {
			pipelines := pipelinesMap[workflowType][action]
			for _, p := range pipelines {
				if len(p.Spec.Containers) == 0 {
					errs = append(errs, fmt.Sprintf("%s: %s.%s pipeline %q has no containers", promiseName, workflowType, action, p.Name))
					continue
				}
				for _, c := range p.Spec.Containers {
					if strings.TrimSpace(c.Image) == "" {
						errs = append(errs, fmt.Sprintf("%s: %s.%s pipeline container %q has no image set", promiseName, workflowType, action, c.Name))
					}
				}
			}
			if action == v1alpha1.WorkflowActionDelete && len(pipelines) > 1 {
				errs = append(errs, fmt.Sprintf("%s: %s delete workflow has %d pipelines, Kratix only supports one", promiseName, workflowType, len(pipelines)))
			}
		}
	}
	return errs
}

// checkExampleFile validates one example/Resource Request file against
// whichever loaded Promise's CRD its apiVersion+kind matches. A file that
// matches no Promise is reported as such, not silently skipped.
func checkExampleFile(name string, data []byte, promises map[string]*v1alpha1.Promise) []string {
	var ex exampleDoc
	if err := sigsyaml.Unmarshal(data, &ex); err != nil {
		return []string{fmt.Sprintf("%s: does not parse as YAML: %v", name, err)}
	}

	for promiseName, promise := range promises {
		if promise == nil || promise.Spec.API == nil {
			continue
		}
		var crd apiextensionsv1.CustomResourceDefinition
		if err := json.Unmarshal(promise.Spec.API.Raw, &crd); err != nil {
			continue
		}
		group := crd.Spec.Group
		for _, v := range crd.Spec.Versions {
			expectedAPIVersion := group + "/" + v.Name
			if ex.APIVersion != expectedAPIVersion || ex.Kind != crd.Spec.Names.Kind {
				continue
			}
			if v.Schema == nil || v.Schema.OpenAPIV3Schema == nil {
				return nil
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

func validateAgainstSchema(name string, spec map[string]interface{}, schema apiextensionsv1.JSONSchemaProps) []string {
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

	promises := map[string]*v1alpha1.Promise{}
	promiseFiles := readYAMLDir(promiseDir)
	names := make([]string, 0, len(promiseFiles))
	for name := range promiseFiles {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		errs, promise, err := checkPromiseFile(name, promiseFiles[name])
		if err != nil {
			allErrs = append(allErrs, err.Error())
			continue
		}
		allErrs = append(allErrs, errs...)
		promises[name] = promise
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
