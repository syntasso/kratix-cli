package integration_test

import (
	"os"
	"path/filepath"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/syntasso/kratix/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

var _ = Describe("InitOperatorPromise", func() {
	var r *runner
	var workingDir string
	var initPromiseCmd []string

	BeforeEach(func() {
		var err error
		workingDir, err = os.MkdirTemp("", "kratix-test")
		Expect(err).NotTo(HaveOccurred())
		r = &runner{exitCode: 0}
		r.flags = map[string]string{
			"--group":              "myorg.com",
			"--kind":               "database",
			"--operator-manifests": "assets/operator",
			"--dir":                workingDir,
			"--api-schema-from":    "postgresqls.acid.zalan.do",
			"--split":              "",
		}
		initPromiseCmd = []string{"init", "operator-promise", "postgresql"}

	})

	AfterEach(func() {
		Expect(os.RemoveAll(workingDir)).To(Succeed())
	})

	When("called without required flags", func() {
		It("prints an error", func() {
			r.exitCode = 1
			r.flags = map[string]string{}
			session := r.run(initPromiseCmd...)
			Expect(session.Err).To(gbytes.Say(`Error: required flag\(s\) "api-schema-from", "group", "kind", "operator-manifests" not set`))
		})
	})

	When("called without required arguments", func() {
		It("prints an error", func() {
			r.exitCode = 1
			r.flags = map[string]string{}
			session := r.run("init", "operator-promise", "--group", "myorg.com", "--kind", "database", "--operator-manifests", "assets/operator", "--api-schema-from", "postgresqls.acid.zalan.do")
			Expect(session.Err).To(gbytes.Say("required argument promise name not specified"))
		})
	})

	Describe("generating a promise from an operator", func() {
		Describe("the generated files", func() {
			var generatedFiles []string
			BeforeEach(func() {
				r.run(initPromiseCmd...)
				fileEntries, err := os.ReadDir(workingDir)
				generatedFiles = []string{}
				for _, fileEntry := range fileEntries {
					generatedFiles = append(generatedFiles, fileEntry.Name())
				}
				Expect(err).ToNot(HaveOccurred())
			})

			It("includes a dependencies.yaml file with the contents of the operator manifests", func() {
				Expect(generatedFiles).To(ContainElement("dependencies.yaml"))

				var dependencies v1alpha1.Dependencies
				depsContent, err := os.ReadFile(filepath.Join(workingDir, "dependencies.yaml"))
				Expect(err).ToNot(HaveOccurred())
				Expect(yaml.Unmarshal(depsContent, &dependencies)).To(Succeed())
				expectDependenciesToMatchOperatorManifests(dependencies)
			})

			It("includes an api.yaml file with the api-schema-from CRD", func() {
				Expect(generatedFiles).To(ContainElement("api.yaml"))

				apiContent, err := os.ReadFile(filepath.Join(workingDir, "api.yaml"))
				Expect(err).ToNot(HaveOccurred())

				var apiCRD apiextensionsv1.CustomResourceDefinition
				Expect(yaml.Unmarshal(apiContent, &apiCRD)).To(Succeed())
				expectCRDToMatchOperatorCRD(apiCRD)
			})

			It("includes a workflow", func() {
				expectedWorkflowFilepath := filepath.Join(workingDir, "workflows", "resource", "configure", "workflow.yaml")
				Expect(expectedWorkflowFilepath).To(BeAnExistingFile())

				workflowContent, err := os.ReadFile(expectedWorkflowFilepath)
				Expect(err).ToNot(HaveOccurred())

				var pipelines []v1alpha1.Pipeline
				Expect(yaml.Unmarshal(workflowContent, &pipelines)).To(Succeed())

				expectPipelinesToMatchOperatorPipelines(pipelines)
			})

			It("includes an example resource request", func() {
				Expect(generatedFiles).To(ContainElement("example-resource.yaml"))
				expectExampleResourceToMatchOperatorResource(workingDir)
			})

			It("includes a README.md file", func() {
				Expect(generatedFiles).To(ContainElement("README.md"))
				readmeContents, err := os.ReadFile(filepath.Join(workingDir, "README.md"))
				Expect(err).ToNot(HaveOccurred())
				Expect(readmeContents).To(ContainSubstring("init operator-promise  --group myorg.com --kind database"))
			})
		})

		When("a version is provided", func() {
			BeforeEach(func() {
				r.flags["--version"] = "v2beta1"
				r.run(initPromiseCmd...)
			})

			It("sets the api version to the provided version", func() {
				apiContent, err := os.ReadFile(filepath.Join(workingDir, "api.yaml"))
				Expect(err).ToNot(HaveOccurred())

				var apiCRD apiextensionsv1.CustomResourceDefinition
				Expect(yaml.Unmarshal(apiContent, &apiCRD)).To(Succeed())

				Expect(apiCRD.Spec.Versions).To(HaveLen(1))
				Expect(apiCRD.Spec.Versions[0].Name).To(Equal("v2beta1"))
				Expect(apiCRD.Spec.Versions[0].Schema.OpenAPIV3Schema.Properties["apiVersion"].Enum).To(HaveLen(1))
				Expect(apiCRD.Spec.Versions[0].Schema.OpenAPIV3Schema.Properties["apiVersion"].Enum[0].Raw).To(BeEquivalentTo(`"myorg.com/v2beta1"`))
			})
		})

		When("there is no matching CRD in the manifests directory", func() {
			BeforeEach(func() {
				r.flags["--api-schema-from"] = "does-not-exist"
			})

			It("returns an error", func() {
				r.exitCode = 1
				session := r.run(initPromiseCmd...)
				Expect(session.Err).To(gbytes.Say(`Error: no CRD found matching name: does-not-exist`))
			})
		})
	})

	Describe("when the --split flag is not provided", func() {
		BeforeEach(func() {
			delete(r.flags, "--split")
			r.run(initPromiseCmd...)
		})

		It("generates a single promise.yaml", func() {
			Expect(filepath.Join(workingDir, "promise.yaml")).To(BeAnExistingFile())
		})

		It("populates the promise.yaml with the right contents", func() {
			promiseContent, err := os.ReadFile(filepath.Join(workingDir, "promise.yaml"))
			Expect(err).ToNot(HaveOccurred())

			var promise v1alpha1.Promise
			Expect(yaml.Unmarshal(promiseContent, &promise)).To(Succeed())

			By("setting the promise name", func() {
				Expect(promise.GetName()).To(Equal("postgresql"))
			})

			By("setting the promise gvk", func() {
				Expect(promise.APIVersion).To(Equal(v1alpha1.GroupVersion.Group + "/" + v1alpha1.GroupVersion.Version))
				Expect(promise.Kind).To(Equal("Promise"))
			})

			By("setting the promise api", func() {
				crd, err := promise.GetAPIAsCRD()
				Expect(err).ToNot(HaveOccurred())
				expectCRDToMatchOperatorCRD(*crd)
			})

			By("setting the promise dependencies", func() {
				expectDependenciesToMatchOperatorManifests(promise.Spec.Dependencies)
			})

			By("setting the promise workflows", func() {
				pipelines, err := v1alpha1.PipelinesFromUnstructured(promise.Spec.Workflows.Resource.Configure, logr.Discard())
				Expect(err).ToNot(HaveOccurred())
				expectPipelinesToMatchOperatorPipelines(pipelines)
			})
		})

		It("includes an example resource request", func() {
			Expect(filepath.Join(workingDir, "example-resource.yaml")).To(BeAnExistingFile())
			expectExampleResourceToMatchOperatorResource(workingDir)
		})

		It("includes a README.md file", func() {
			Expect(filepath.Join(workingDir, "README.md")).To(BeAnExistingFile())
			readmeContents, err := os.ReadFile(filepath.Join(workingDir, "README.md"))
			Expect(err).ToNot(HaveOccurred())
			Expect(readmeContents).To(ContainSubstring("init operator-promise postgresql --group myorg.com --kind database"))
		})
	})

	Describe("end-to-end Promise generation", func() {
		BeforeEach(func() {
			r.flags["--group"] = "syntasso.io"
			r.flags["--kind"] = "Database"
			r.flags["--operator-manifests"] = "assets/e2e-cnpg/manifests"
			r.flags["--api-schema-from"] = "clusters.postgresql.cnpg.io"
			delete(r.flags, "--split")

			r.run(initPromiseCmd...)
		})

		It("generates the expected promise.yaml", func() {
			promiseContent, err := os.ReadFile(filepath.Join(workingDir, "promise.yaml"))
			Expect(err).ToNot(HaveOccurred())

			expectedPromiseContent, err := os.ReadFile("assets/e2e-cnpg/expected-promise.yaml")
			Expect(err).ToNot(HaveOccurred())

			Expect(promiseContent).To(MatchYAML(expectedPromiseContent))
		})

		It("generates the expected example-resource.yaml", func() {
			exampleResourceContent, err := os.ReadFile(filepath.Join(workingDir, "example-resource.yaml"))
			Expect(err).ToNot(HaveOccurred())

			expectedExampleResourceContent, err := os.ReadFile("assets/e2e-cnpg/expected-example-resource.yaml")
			Expect(err).ToNot(HaveOccurred())

			Expect(exampleResourceContent).To(MatchYAML(expectedExampleResourceContent))
		})
	})
})

func expectCRDToMatchOperatorCRD(apiCRD apiextensionsv1.CustomResourceDefinition) {
	Expect(apiCRD.Name).To(Equal("databases.myorg.com"))
	Expect(apiCRD.Spec.Group).To(Equal("myorg.com"))
	Expect(apiCRD.Spec.Names).To(Equal(apiextensionsv1.CustomResourceDefinitionNames{
		Plural:   "databases",
		Singular: "database",
		Kind:     "database",
	}))
	Expect(apiCRD.Spec.Versions).To(HaveLen(1))
	Expect(apiCRD.Spec.Versions[0].Name).To(Equal("v1Stored"))
	Expect(apiCRD.Spec.Versions[0].Schema.OpenAPIV3Schema.Properties["kind"].Enum).To(HaveLen(1))
	Expect(apiCRD.Spec.Versions[0].Schema.OpenAPIV3Schema.Properties["kind"].Enum[0].Raw).To(BeEquivalentTo(`"database"`))
	Expect(apiCRD.Spec.Versions[0].Schema.OpenAPIV3Schema.Properties["apiVersion"].Enum).To(HaveLen(1))
	Expect(apiCRD.Spec.Versions[0].Schema.OpenAPIV3Schema.Properties["apiVersion"].Enum[0].Raw).To(BeEquivalentTo(`"myorg.com/v1Stored"`))
}

func expectDependenciesToMatchOperatorManifests(dependencies v1alpha1.Dependencies) {
	Expect(dependencies).To(HaveLen(7))

	var objects []string
	var objectNamespaces []string
	for _, obj := range dependencies {
		objects = append(objects, obj.GetName())
		objectNamespaces = append(objectNamespaces, obj.GetNamespace())
	}

	Expect(objects).To(ConsistOf(
		"operator-sa",
		"subdir-sa",
		"pod-reader",
		"operator-deployment",
		"postgresteams.acid.zalan.do",
		"postgresqls.acid.zalan.do",
		"operatorconfigurations.acid.zalan.do",
	))

	Expect(objectNamespaces).To(ConsistOf(
		"default",
		"default",
		"defined-namespace",
		"default",
		"default",
		"default",
		"default",
	))

}

func expectPipelinesToMatchOperatorPipelines(pipelines []v1alpha1.Pipeline) {
	Expect(pipelines).To(HaveLen(1))
	pipeline := pipelines[0]
	Expect(pipeline.Spec.Containers).To(HaveLen(1))
	Expect(pipeline.Spec.Containers[0].Name).To(Equal("from-api-to-operator"))
	Expect(pipeline.Spec.Containers[0].Image).To(Equal("ghcr.io/syntasso/kratix-cli/from-api-to-operator:v0.1.0"))

	Expect(pipeline.Spec.Containers[0].Env).To(HaveLen(3))
	Expect(pipeline.Spec.Containers[0].Env).To(ConsistOf([]corev1.EnvVar{
		{Name: "OPERATOR_GROUP", Value: "acid.zalan.do"},
		{Name: "OPERATOR_KIND", Value: "postgresql"},
		{Name: "OPERATOR_VERSION", Value: "v1Stored"},
	}))
}

func expectExampleResourceToMatchOperatorResource(workingDir string) {
	exampleResourceContents, err := os.ReadFile(filepath.Join(workingDir, "example-resource.yaml"))
	Expect(err).ToNot(HaveOccurred())

	exampleResource := &unstructured.Unstructured{}
	Expect(yaml.Unmarshal(exampleResourceContents, exampleResource)).To(Succeed())

	Expect(exampleResource.GetName()).To(Equal("example-database"))
	Expect(exampleResource.GetNamespace()).To(Equal("default"))
	Expect(exampleResource.GetKind()).To(Equal("database"))
	Expect(exampleResource.GetAPIVersion()).To(Equal("myorg.com/v1Stored"))

	spec, found, err := unstructured.NestedMap(exampleResource.Object, "spec")
	Expect(err).ToNot(HaveOccurred())
	Expect(found).To(BeTrue())
	Expect(spec).To(Equal(map[string]interface{}{
		"numberOfInstances": "# type integer",
		"teamId":            "# type string",
		"postgresql":        "# type object",
		"volume":            "# type object",
	}))
}
