package integration_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/syntasso/kratix/api/v1alpha1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"os"
	"path/filepath"
)

var _ = Describe("update", func() {
	var workingDir string
	var r *runner

	BeforeEach(func() {
		var err error
		workingDir, err = os.MkdirTemp("", "kratix-test")
		Expect(err).NotTo(HaveOccurred())
		r = &runner{exitCode: 0, dir: workingDir}
	})
	AfterEach(func() {
		os.RemoveAll(workingDir)
	})

	When("called without a subcommand", func() {
		It("prints the help", func() {
			session := r.run("update")
			Expect(session.Out).To(SatisfyAll(
				gbytes.Say("Command to update kratix resources"),
				gbytes.Say(`Use "kratix update \[command\] --help" for more information about a command.`),
			))
		})
	})

	Context("api", func() {
		When("called with --help", func() {
			It("prints the help", func() {
				session := r.run("update", "api", "--help")
				Expect(session.Out).To(gbytes.Say("Command to update promise API"))
			})
		})

		When("updating promise api", func() {
			var dir string
			AfterEach(func() {
				os.RemoveAll(dir)
			})

			When("there is no api.yaml or promise.yaml present", func() {
				It("errors with a helpful message", func() {
					r.exitCode = 1
					sess := r.run("update", "api", "-p", "test:string")
					Expect(sess.Err).To(gbytes.Say("failed to find api.yaml or promise.yaml in directory. Please run 'kratix init promise' first"))
				})
			})

			When("working with promise.yaml", func() {
				BeforeEach(func() {
					var err error
					dir, err = os.MkdirTemp("", "kratix-update-api-test")
					Expect(err).NotTo(HaveOccurred())

					sess := r.run("init", "promise", "postgresql", "--group", "syntasso.io", "--kind", "Database", "--dir", dir)
					Expect(sess.Out).To(gbytes.Say("postgresql promise bootstrapped in"))
				})

				Context("api GVK", func() {
					It("updates", func() {
						sess := r.run("update", "api", "--kind", "NewKind", "--group", "newGroup", "--version", "v1beta4", "--plural", "newPlural", "--dir", dir)
						Expect(sess.Out).To(gbytes.Say("Promise api updated"))
						matchPromise(dir, "postgresql", "newGroup", "v1beta4", "NewKind", "newkind", "newPlural")
					})
				})

				Context("api properties", func() {
					It("can add new properties to the promise api", func() {
						sess := r.run("update", "api", "-p", "numberField:number", "--property", "stringField:string", "intvalue:integer", "--dir", dir)
						Expect(sess.Out).To(gbytes.Say("Promise api updated"))
						matchPromise(dir, "postgresql", "syntasso.io", "v1alpha1", "Database", "database", "databases")
						props := getCRDProperties(dir, false)
						Expect(props).To(SatisfyAll(HaveKey("numberField"), HaveKey("stringField")))
						Expect(props["numberField"].Type).To(Equal("number"))
						Expect(props["stringField"].Type).To(Equal("string"))
					})

					It("can update existing properties types", func() {
						r.run("update", "api", "-p", "numberField:number", "--property", "stringField:string", "-p", "wontchange:string", "--dir", dir)
						r.run("update", "api", "-p", "numberField:string", "--property", "stringField:number", "--dir", dir)
						matchPromise(dir, "postgresql", "syntasso.io", "v1alpha1", "Database", "database", "databases")
						props := getCRDProperties(dir, false)
						Expect(props).To(SatisfyAll(HaveKey("numberField"), HaveKey("stringField"), HaveKey("wontchange")))
						Expect(props["numberField"].Type).To(Equal("string"))
						Expect(props["wontchange"].Type).To(Equal("string"))
						Expect(props["stringField"].Type).To(Equal("number"))
					})

					It("errors when unsupported property type is set", func() {
						r.exitCode = 1
						sess := r.run("update", "api", "--property", "unsupported:object", "--dir", dir)
						Expect(sess.Err).To(gbytes.Say("unsupported"))
					})

					It("can remove existing properties", func() {
						r.run("update", "api", "-p", "numberField:number", "--property", "stringField:string", "-p", "wontdelete:string", "--dir", dir)
						r.run("update", "api", "-p", "numberField-", "--property", "stringField-", "--dir", dir)
						matchPromise(dir, "postgresql", "syntasso.io", "v1alpha1", "Database", "database", "databases")
						props := getCRDProperties(dir, false)
						Expect(props).To(SatisfyAll(HaveKey("wontdelete"), HaveLen(1)))
						Expect(props["wontdelete"].Type).To(Equal("string"))
					})

					It("errors when property format is invalid", func() {
						r.exitCode = 1
						sess := r.run("update", "api", "--property", "invalid%", "--dir", dir)
						Expect(sess.Err).To(gbytes.Say("invalid"))

						r.exitCode = 1
						sess = r.run("update", "api", "--property", "invalid+string", "--dir", dir)
						Expect(sess.Err).To(gbytes.Say("invalid"))
					})
				})
			})

			When("working with promise generated with --split flag", func() {
				BeforeEach(func() {
					var err error
					dir, err = os.MkdirTemp("", "kratix-update-api-test")
					Expect(err).NotTo(HaveOccurred())

					sess := r.run("init", "promise", "postgresql", "--group", "syntasso.io", "--kind", "Database", "--split")
					Expect(sess.Out).To(gbytes.Say("postgresql promise bootstrapped in"))
				})

				It("can update gvk of the api", func() {
					sess := r.run("update", "api", "--kind", "NewKind", "--group", "newGroup", "--version", "v1beta4", "--plural", "newPlural")
					Expect(sess.Out).To(gbytes.Say("Promise api updated"))
					matchApiGvk(workingDir, "newGroup", "v1beta4", "NewKind", "newkind", "newPlural")
				})

				It("can add new properties and update existing properties to the promise api", func() {
					sess := r.run("update", "api", "-p", "f1:number", "--property", "p2:string")
					Expect(sess.Out).To(gbytes.Say("Promise api updated"))
					matchApiGvk(workingDir, "syntasso.io", "v1alpha1", "Database", "database", "databases")

					props := getCRDProperties(workingDir, true)
					Expect(props).To(SatisfyAll(HaveKey("f1"), HaveKey("p2"), HaveLen(2)))
					Expect(props["f1"].Type).To(Equal("number"))
					Expect(props["p2"].Type).To(Equal("string"))
				})

				It("can remove existing properties", func() {
					r.run("update", "api", "-p", "numberField:number", "--property", "stringField:string", "-p", "keep:string")
					r.run("update", "api", "-p", "numberField-", "--property", "stringField-")
					matchApiGvk(workingDir, "syntasso.io", "v1alpha1", "Database", "database", "databases")

					props := getCRDProperties(workingDir, true)
					Expect(props).To(SatisfyAll(HaveKey("keep"), HaveLen(1)))
					Expect(props["keep"].Type).To(Equal("string"))
				})
			})
		})

	})
})

func matchApiGvk(dir, group, version, kind, singular, plural string) {
	apiYAML, err := os.ReadFile(filepath.Join(dir, "api.yaml"))
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	var crd apiextensionsv1.CustomResourceDefinition
	ExpectWithOffset(1, yaml.Unmarshal(apiYAML, &crd)).To(Succeed())
	matchCRD(&crd, group, version, kind, singular, plural)
}

func getCRDProperties(dir string, split bool) map[string]apiextensionsv1.JSONSchemaProps {
	var crd *apiextensionsv1.CustomResourceDefinition
	if split {
		apiYAML, err := os.ReadFile(filepath.Join(dir, "api.yaml"))
		ExpectWithOffset(1, err).NotTo(HaveOccurred())
		ExpectWithOffset(1, yaml.Unmarshal(apiYAML, &crd)).To(Succeed())
	} else {
		promiseYAML, err := os.ReadFile(filepath.Join(dir, "promise.yaml"))
		ExpectWithOffset(1, err).NotTo(HaveOccurred())

		var promise v1alpha1.Promise
		ExpectWithOffset(1, yaml.Unmarshal(promiseYAML, &promise)).To(Succeed())
		crd, err = promise.GetAPIAsCRD()
		ExpectWithOffset(1, err).NotTo(HaveOccurred())
	}
	return crd.Spec.Versions[0].Schema.OpenAPIV3Schema.Properties["spec"].Properties
}
