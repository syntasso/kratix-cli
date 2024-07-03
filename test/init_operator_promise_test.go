package integration_test

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/syntasso/kratix/api/v1alpha1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
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
			"--api-from":           "postgresqls.acid.zalan.do",
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
			Expect(session.Err).To(gbytes.Say(`Error: required flag\(s\) "api-from", "group", "kind", "operator-manifests" not set`))
		})
	})

	When("called without required arguments", func() {
		It("prints an error", func() {
			r.exitCode = 1
			session := r.run("init", "operator-promise")
			Expect(session.Err).To(gbytes.Say(`Error: accepts 1 arg\(s\), received 0`))
		})
	})

	Describe("generating a promise from an operator", func() {
		When("all arguments are valid", func() {
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

			It("generates the dependencies.yaml file with the contents of the operator manifests", func() {
				Expect(generatedFiles).To(ContainElement("dependencies.yaml"))

				var dependencies v1alpha1.Dependencies
				depsContent, err := os.ReadFile(filepath.Join(workingDir, "dependencies.yaml"))
				Expect(err).ToNot(HaveOccurred())
				Expect(yaml.Unmarshal(depsContent, &dependencies)).To(Succeed())

				Expect(dependencies).To(HaveLen(6))

				var objects []string
				for _, obj := range dependencies {
					objects = append(objects, obj.GetName())
				}

				Expect(objects).To(ConsistOf(
					"operator-sa",
					"pod-reader",
					"operator-deployment",
					"postgresteams.acid.zalan.do",
					"postgresqls.acid.zalan.do",
					"operatorconfigurations.acid.zalan.do",
				))
			})

			It("generates the api.yaml file with the api-from CRD", func() {
				Expect(generatedFiles).To(ContainElement("api.yaml"))

				apiContent, err := os.ReadFile(filepath.Join(workingDir, "api.yaml"))
				Expect(err).ToNot(HaveOccurred())

				var apiCRD apiextensionsv1.CustomResourceDefinition
				Expect(yaml.Unmarshal(apiContent, &apiCRD)).To(Succeed())

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
				r.flags["--api-from"] = "does-not-exist"
			})

			It("returns an error", func() {
				r.exitCode = 1
				session := r.run(initPromiseCmd...)
				Expect(session.Err).To(gbytes.Say(`Error: no CRD found matching name: does-not-exist`))
			})
		})
	})
})
