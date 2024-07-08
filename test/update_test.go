package integration_test

import (
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/syntasso/kratix/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"os"
	"path/filepath"
	yamlsig "sigs.k8s.io/yaml"
	"slices"
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
		Expect(os.RemoveAll(workingDir)).To(Succeed())
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
						matchExampleResource(dir, "example-postgresql", "newGroup", "v1beta4", "NewKind")
					})
				})

				Context("api properties", func() {
					It("can add new properties to the promise api", func() {
						sess := r.run("update", "api", "-p", "numberField:number", "--property", "stringField:string", "--property", "intValue:integer", "--dir", dir)
						Expect(sess.Out).To(gbytes.Say("Promise api updated"))
						matchPromise(dir, "postgresql", "syntasso.io", "v1alpha1", "Database", "database", "databases")
						props := getCRDProperties(dir, false)
						Expect(props).To(SatisfyAll(HaveKey("numberField"), HaveKey("stringField"), HaveKey("intValue"), HaveLen(3)))
						Expect(props["numberField"].Type).To(Equal("number"))
						Expect(props["stringField"].Type).To(Equal("string"))
						Expect(props["intValue"].Type).To(Equal("integer"))
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
					sess := r.run("update", "api", "--kind", "NewKind", "--group", "newGroup", "--version", "v2beta4", "--plural", "newPlural")
					Expect(sess.Out).To(gbytes.Say("Promise api updated"))
					matchGvkInAPIFile(workingDir, "newGroup", "v2beta4", "NewKind", "newkind", "newPlural")
					matchExampleResource(workingDir, "example-postgresql", "newGroup", "v2beta4", "NewKind")
				})

				It("can add new properties and update existing properties to the promise api", func() {
					sess := r.run("update", "api", "-p", "f1:number", "--property", "p2:string")
					Expect(sess.Out).To(gbytes.Say("Promise api updated"))
					matchGvkInAPIFile(workingDir, "syntasso.io", "v1alpha1", "Database", "database", "databases")

					props := getCRDProperties(workingDir, true)
					Expect(props).To(SatisfyAll(HaveKey("f1"), HaveKey("p2"), HaveLen(2)))
					Expect(props["f1"].Type).To(Equal("number"))
					Expect(props["p2"].Type).To(Equal("string"))
				})

				It("can remove existing properties", func() {
					r.run("update", "api", "-p", "numberField:number", "--property", "stringField:string", "-p", "keep:string")
					r.run("update", "api", "-p", "numberField-", "--property", "stringField-")
					matchGvkInAPIFile(workingDir, "syntasso.io", "v1alpha1", "Database", "database", "databases")

					props := getCRDProperties(workingDir, true)
					Expect(props).To(SatisfyAll(HaveKey("keep"), HaveLen(1)))
					Expect(props["keep"].Type).To(Equal("string"))
				})
			})
		})

	})

	Context("dependencies", func() {
		var (
			depDir      string
			ns1, ns2    *v1.Namespace
			deployment1 *appsv1.Deployment
		)

		BeforeEach(func() {
			var err error
			depDir, err = os.MkdirTemp("", "dep")
			Expect(err).NotTo(HaveOccurred())

			ns1 = namespace("test1")
			ns2 = namespace("test2")
			deployment1 = deployment("test1")

			Expect(r.run("init", "promise", "postgresql",
				"--group", "syntasso.io",
				"--kind", "Database").Out).To(gbytes.Say("postgresql promise bootstrapped in"))
		})

		AfterEach(func() {
			Expect(os.RemoveAll(depDir)).To(Succeed())
		})

		When("called without an argument", func() {
			It("errors and print a message", func() {
				r.exitCode = 1
				Expect(r.run("update", "dependencies").Err).To(gbytes.Say(`Error: accepts 1 arg\(s\), received 0`))
			})
		})

		Context("dependency directory", func() {
			When("does not exist", func() {
				It("errors and does not update promise.yaml", func() {
					r.exitCode = 1
					sess := r.run("update", "dependencies", "doesnotexistyet")
					Expect(sess.Err).To(gbytes.Say("failed to read dependency directory: doesnotexistyet"))
					matchPromise(workingDir, "postgresql", "syntasso.io", "v1alpha1", "Database", "database", "databases")
				})
			})

			When("exists but is empty", func() {
				It("errors and does not update promise.yaml", func() {
					r.exitCode = 1
					Expect(r.run("update", "dependencies", depDir).Err).To(gbytes.Say(fmt.Sprintf("no files found in directory: %s", depDir)))
					matchPromise(workingDir, "postgresql", "syntasso.io", "v1alpha1", "Database", "database", "databases")
				})
			})

			When("contains only empty files", func() {
				It("errors and does not update promise.yaml", func() {
					Expect(os.WriteFile(filepath.Join(depDir, "empty-dependencies.yaml"), []byte(""), 0644)).To(Succeed())
					r.exitCode = 1
					Expect(r.run("update", "dependencies", depDir).Err).To(gbytes.Say(fmt.Sprintf("no valid dependencies found in directory: %s", depDir)))
					matchPromise(workingDir, "postgresql", "syntasso.io", "v1alpha1", "Database", "database", "databases")
				})
			})
		})

		Context("dependencies.yaml exists", func() {
			var promiseDir string
			BeforeEach(func() {
				var err error
				promiseDir, err = os.MkdirTemp("", "split-promise")
				Expect(err).NotTo(HaveOccurred())

				Expect(r.run("init", "promise", "postgresql",
					"--group", "syntasso.io",
					"--kind", "Database",
					"--dir", promiseDir,
					"--split").Out).To(gbytes.Say("postgresql promise bootstrapped in"))
			})

			It("updates dependencies.yaml file", func() {
				Expect(os.WriteFile(filepath.Join(depDir, "deps.yaml"), slices.Concat(
					namespaceBytes(ns1),
					namespaceBytes(ns2),
					deploymentBytes(deployment1)), 0644)).To(Succeed())

				Expect(r.run("update", "dependencies", depDir, "--dir", promiseDir).Out).To(gbytes.Say("Updated dependencies.yaml"))
				generatedDeps := getDependencies(promiseDir, true)
				Expect(generatedDeps).To(HaveLen(3))
				Expect(generatedDeps[0].Object["apiVersion"]).To(Equal("v1"))
				Expect(generatedDeps[0].Object["kind"]).To(Equal("Namespace"))
				Expect(generatedDeps[1].Object["apiVersion"]).To(Equal("v1"))
				Expect(generatedDeps[1].Object["kind"]).To(Equal("Namespace"))
				Expect(generatedDeps[2].Object["apiVersion"]).To(Equal("apps/v1"))
				Expect(generatedDeps[2].Object["kind"]).To(Equal("Deployment"))
			})
		})

		Context("dependencies.yaml does not exist", func() {
			It("updates promise.yaml file", func() {
				Expect(os.WriteFile(filepath.Join(depDir, "deps.yaml"),
					slices.Concat(namespaceBytes(ns1), deploymentBytes(deployment1)), 0644)).To(Succeed())
				Expect(os.WriteFile(filepath.Join(depDir, "namespace.yaml"), namespaceBytes(ns2), 0644)).To(Succeed())

				Expect(r.run("update", "dependencies", depDir).Out).To(gbytes.Say("Updated promise.yaml"))
				generatedDeps := getDependencies(workingDir, false)
				Expect(generatedDeps).To(HaveLen(3))

				var kinds []string
				for _, d := range generatedDeps {
					kinds = append(kinds, d.Object["kind"].(string))
				}
				Expect(kinds).To(ConsistOf("Namespace", "Namespace", "Deployment"))
			})

			When("promise.yaml does not exist", func() {
				It("succeeds and writes dependencies to dependencies.yaml", func() {
					promiseDir, err := os.MkdirTemp("", "promise")
					Expect(err).NotTo(HaveOccurred())

					Expect(os.WriteFile(filepath.Join(depDir, "namespace.yaml"), namespaceBytes(ns1), 0644)).To(Succeed())
					sess := r.run("update", "dependencies", depDir, "--dir", promiseDir)
					Expect(sess.Out).To(gbytes.Say("Updated dependencies.yaml"))
				})
			})

			When("dependency directory contains file that cannot be decoded", func() {
				It("updates promise.yaml file with the valid dependencies", func() {
					Expect(os.WriteFile(filepath.Join(depDir, "deps.yaml"),
						slices.Concat(namespaceBytes(ns1), deploymentBytes(deployment1)), 0644)).To(Succeed())
					Expect(os.WriteFile(filepath.Join(depDir, "not-yaml"), []byte("not valid yaml"), 0644)).To(Succeed())
					sess := r.run("update", "dependencies", depDir)
					Expect(sess.Out).To(gbytes.Say("Skipped invalid yaml documents during dependency writing"))
					Expect(sess.Out).To(gbytes.Say("Updated promise.yaml"))
					generatedDeps := getDependencies(workingDir, false)
					Expect(generatedDeps).To(HaveLen(2))
					Expect(generatedDeps[0].Object["apiVersion"]).To(Equal("v1"))
					Expect(generatedDeps[0].Object["kind"]).To(Equal("Namespace"))
					Expect(generatedDeps[1].Object["apiVersion"]).To(Equal("apps/v1"))
					Expect(generatedDeps[1].Object["kind"]).To(Equal("Deployment"))
				})
			})
		})
	})

	Context("destination-selector", func() {
		BeforeEach(func() {
			Expect(r.run("init", "promise", "postgresql",
				"--group", "syntasso.io",
				"--kind", "Database").Out).To(gbytes.Say("postgresql promise bootstrapped in"))
		})

		When("called with --help", func() {
			It("prints the help", func() {
				session := r.run("update", "destination-selector", "--help")
				Expect(session.Out).To(gbytes.Say("Command to update destination selectors"))
			})
		})

		When("called without an argument", func() {
			It("errors and print a message", func() {
				r.exitCode = 1
				Expect(r.run("update", "destination-selector").Err).To(gbytes.Say(`Error: accepts 1 arg\(s\), received 0`))
			})
		})

		When("there is no promise.yaml", func() {
			It("errors with a helpful message", func() {
				promiseDir, err := os.MkdirTemp("", "promise")
				Expect(err).NotTo(HaveOccurred())

				r.exitCode = 1
				sess := r.run("update", "destination-selector", "zone=europe-west2", "-d", promiseDir)
				Expect(sess.Err).To(gbytes.Say("failed to find promise.yaml in directory"))
			})
		})

		It("can add new selector to the promise api", func() {
			sess := r.run("update", "destination-selector", "env=prod")
			Expect(sess.Out).To(gbytes.Say("Promise destination selector updated"))
			sess = r.run("update", "destination-selector", "zone=test-zone-b")
			Expect(sess.Out).To(gbytes.Say("Promise destination selector updated"))
			matchPromise(workingDir, "postgresql", "syntasso.io", "v1alpha1", "Database", "database", "databases")
			Expect(getDestinationSelectors(workingDir)).To(SatisfyAll(HaveLen(2), HaveKeyWithValue("env", "prod"), HaveKeyWithValue("zone", "test-zone-b")))
		})

		It("can update existing selectors", func() {
			sess := r.run("update", "destination-selector", "env=dev")
			Expect(sess.Out).To(gbytes.Say("Promise destination selector updated"))
			Expect(getDestinationSelectors(workingDir)).To(SatisfyAll(HaveLen(1), HaveKeyWithValue("env", "dev")))

			sess = r.run("update", "destination-selector", "env=prod")
			Expect(sess.Out).To(gbytes.Say("Promise destination selector updated"))
			Expect(getDestinationSelectors(workingDir)).To(SatisfyAll(HaveLen(1), HaveKeyWithValue("env", "prod")))
		})

		It("can remove existing selectors", func() {
			r.run("update", "destination-selector", "env=prod")
			r.run("update", "destination-selector", "akey=noupdate")
			Expect(getDestinationSelectors(workingDir)).To(SatisfyAll(HaveLen(2), HaveKeyWithValue("env", "prod"), HaveKeyWithValue("akey", "noupdate")))

			r.run("update", "destination-selector", "env=dev")
			Expect(getDestinationSelectors(workingDir)).To(SatisfyAll(HaveLen(2), HaveKeyWithValue("env", "dev"), HaveKeyWithValue("akey", "noupdate")))
		})

		It("errors when argument format is invalid", func() {
			r.exitCode = 1
			sess := r.run("update", "destination-selector", "akey%avalue")
			Expect(sess.Err).To(gbytes.Say("invalid"))
		})

	})
})

func getDependencies(dir string, split bool) v1alpha1.Dependencies {
	var deps v1alpha1.Dependencies
	if split {
		bytes, err := os.ReadFile(filepath.Join(dir, "dependencies.yaml"))
		ExpectWithOffset(1, err).NotTo(HaveOccurred())
		ExpectWithOffset(1, yaml.Unmarshal(bytes, &deps)).To(Succeed())
	} else {
		promiseBytes, err := os.ReadFile(filepath.Join(dir, "promise.yaml"))
		ExpectWithOffset(1, err).NotTo(HaveOccurred())
		var promise v1alpha1.Promise
		ExpectWithOffset(1, yaml.Unmarshal(promiseBytes, &promise)).To(Succeed())
		deps = promise.Spec.Dependencies
	}
	return deps
}

func matchGvkInAPIFile(dir, group, version, kind, singular, plural string) {
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

func getDestinationSelectors(dir string) map[string]string {
	promiseYAML, err := os.ReadFile(filepath.Join(dir, "promise.yaml"))
	ExpectWithOffset(1, err).NotTo(HaveOccurred())

	var promise v1alpha1.Promise
	ExpectWithOffset(1, yaml.Unmarshal(promiseYAML, &promise)).To(Succeed())
	return promise.GetSchedulingSelectors()
}

func namespaceBytes(ns *v1.Namespace) []byte {
	separator := []byte("---\n")
	bytes, err := yamlsig.Marshal(ns)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	return append(separator, bytes...)
}

func deploymentBytes(dep *appsv1.Deployment) []byte {
	separator := []byte("---\n")
	bytes, err := yamlsig.Marshal(dep)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	return append(separator, bytes...)
}

func namespace(name string) *v1.Namespace {
	return &v1.Namespace{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Namespace",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}

func deployment(name string) *appsv1.Deployment {
	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}
