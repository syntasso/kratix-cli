package integration_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/syntasso/kratix/api/v1alpha1"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
	ctrl "sigs.k8s.io/controller-runtime"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

var _ = Describe("kratix", func() {
	var r *runner
	var workingDir string

	BeforeEach(func() {
		var err error
		workingDir, err = os.MkdirTemp("", "kratix-test")
		Expect(err).NotTo(HaveOccurred())
		r = &runner{exitCode: 0, dir: workingDir}
	})

	AfterEach(func() {
		os.RemoveAll(workingDir)
	})

	Describe("help", func() {
		It("prints the help", func() {
			session := r.run("help")
			Expect(session.Out).To(gbytes.Say("A CLI tool for Kratix"))
		})
	})

	Describe("build", func() {
		var dir string
		AfterEach(func() {
			os.RemoveAll(dir)
		})

		It("builds a promise from api, dependencies and workflows files", func() {
			var err error
			dir, err = os.MkdirTemp("", "kratix-build-test")
			Expect(err).NotTo(HaveOccurred())

			r.run("init", "promise", "postgresql", "--group", "syntasso.io", "--kind", "Database", "--split", "--output-dir", dir)
			sess := r.run("build", "promise", "postgresql", "--dir", dir)
			Expect(sess.Out.Contents()).ToNot(BeEmpty())
			var promise v1alpha1.Promise
			Expect(yaml.Unmarshal(sess.Out.Contents(), &promise)).To(Succeed())
			Expect(promise.Name).To(Equal("postgresql"))
			Expect(promise.Kind).To(Equal("Promise"))
			Expect(promise.APIVersion).To(Equal(v1alpha1.GroupVersion.String()))

			promiseCRD, err := promise.GetAPIAsCRD()
			Expect(err).NotTo(HaveOccurred())
			matchCRD(promiseCRD, "syntasso.io", "v1alpha1", "Database", "database", "databases")
		})

		When("--output flag is provided", func() {
			It("outputs promise definition to provided filepath", func() {
				var err error
				dir, err = os.MkdirTemp("", "kratix-build-test")
				Expect(err).NotTo(HaveOccurred())

				r.run("init", "promise", "postgresql", "--group", "syntasso.io", "--kind", "Database", "--split", "--output-dir", dir)
				r.run("build", "promise", "postgresql", "--dir", dir, "--output", filepath.Join(dir, "promise.yaml"))
				matchPromise(dir, "postgresql", "syntasso.io", "v1alpha1", "Database", "database", "databases")
			})
		})
	})

	Describe("init", func() {
		When("called without a subcommand", func() {
			It("prints the help", func() {
				session := r.run("init")
				Expect(session.Out).To(SatisfyAll(
					gbytes.Say("Command used to initialize Kratix resources"),
					gbytes.Say(`Use "kratix init \[command\] --help" for more information about a command.`),
				))
			})
		})

		Describe("subcommands", func() {
			Context("promise", func() {
				When("called with --help", func() {
					It("prints the help", func() {
						session := r.run("init", "promise", "--help")
						Expect(session.Out).To(gbytes.Say("Initialize a new Promise"))
					})
				})

				When("called without required flags", func() {
					It("prints an error", func() {
						session := withExitCode(1).run("init", "promise", "postgresql")
						Expect(session.Err).To(gbytes.Say(`Error: required flag\(s\) "group", "kind" not set`))
					})
				})

				When("called without the required arguments", func() {
					It("prints an error", func() {
						session := withExitCode(1).run("init", "promise")
						Expect(session.Err).To(gbytes.Say(`Error: accepts 1 arg\(s\), received 0`))
					})
				})

				It("generates the promise structure", func() {
					session := r.run("init", "promise", "postgresql", "--group", "syntasso.io", "--kind", "Database")
					Expect(session.Out).To(gbytes.Say("postgresql promise bootstrapped in the current directory"))

					files, err := os.ReadDir(workingDir)
					Expect(err).NotTo(HaveOccurred())
					Expect(files).To(HaveLen(3))

					By("generating a promise.yaml file", func() {
						matchPromise(workingDir, "postgresql", "syntasso.io", "v1alpha1", "Database", "database", "databases")
					})

					By("generating an example-resource.yaml file", func() {
						matchExampleResource(workingDir, "example-postgresql", "syntasso.io", "v1alpha1", "Database")
					})

					By("including a README file", func() {
						readmeContents, err := os.ReadFile(filepath.Join(workingDir, "README.md"))
						Expect(err).NotTo(HaveOccurred())
						Expect(readmeContents).To(ContainSubstring("kratix init promise postgresql"))
					})
				})

				When("the optional flags are provided", func() {
					It("respects the provided values", func() {
						subdir := filepath.Join(workingDir, "subdir")
						Expect(os.Mkdir(subdir, 0755)).To(Succeed())

						session := r.run("init", "promise", "postgresql", "--group", "syntasso.io", "--kind", "Database", "--plural", "dbs", "--version", "v2", "--output-dir", "subdir")
						Expect(session.Out).To(gbytes.Say("postgresql promise bootstrapped in the subdir directory"))

						By("generating a promise.yaml file", func() {
							matchPromise(subdir, "postgresql", "syntasso.io", "v2", "Database", "database", "dbs")
						})

						By("generating an example-resource.yaml file", func() {
							matchExampleResource(subdir, "example-postgresql", "syntasso.io", "v2", "Database")
						})
					})

					When("--split flag is provided", func() {
						It("produces separate promise files", func() {
							session := r.run("init", "promise", "postgresql", "--group", "syntasso.io", "--kind", "Database", "--split")
							Expect(session.Out).To(gbytes.Say("postgresql promise bootstrapped in the current directory"))

							By("generating different files for api, dependencies and workflows", func() {
								files, err := os.ReadDir(workingDir)
								Expect(err).NotTo(HaveOccurred())
								Expect(files).To(HaveLen(5))
								var fileNames []string
								for _, f := range files {
									fileNames = append(fileNames, f.Name())
								}
								Expect(fileNames).To(ContainElements(
									"workflows.yaml",
									"api.yaml",
									"dependencies.yaml",
								))
							})

							By("generating api.yaml with correct values", func() {
								apiYAML, err := os.ReadFile(filepath.Join(workingDir, "api.yaml"))
								Expect(err).NotTo(HaveOccurred())
								var promiseCRD apiextensionsv1.CustomResourceDefinition
								Expect(yaml.Unmarshal(apiYAML, &promiseCRD)).To(Succeed())
								matchCRD(&promiseCRD, "syntasso.io", "v1alpha1", "Database", "database", "databases")
							})
						})
					})
				})
			})
		})
	})

	Describe("update", func() {
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

				BeforeEach(func() {
					var err error
					dir, err = os.MkdirTemp("", "kratix-update-api-test")
					Expect(err).NotTo(HaveOccurred())

					sess := r.run("init", "promise", "postgresql", "--group", "syntasso.io", "--kind", "Database", "--output-dir", dir)
					Expect(sess.Out).To(gbytes.Say("postgresql promise bootstrapped in"))
				})

				Context("api GVK", func() {
					It("updates", func() {
						sess := r.run("update", "api", "--kind", "NewKind", "--group", "newGroup", "--version", "v1beta4", "--plural", "newPlural", "--dir", dir)
						Expect(sess.Out).To(gbytes.Say("Promise updated"))
						matchPromise(dir, "postgresql", "newGroup", "v1beta4", "NewKind", "newkind", "newPlural")
					})
				})

				Context("api properties", func() {
					It("can add new properties to the promise api", func() {
						sess := r.run("update", "api", "-p", "numberField:number", "--property", "stringField:string", "--dir", dir)
						Expect(sess.Out).To(gbytes.Say("Promise updated"))
						matchPromise(dir, "postgresql", "syntasso.io", "v1alpha1", "Database", "database", "databases")
						props := getCRDProperties(dir)
						Expect(props).To(SatisfyAll(HaveKey("numberField"), HaveKey("stringField")))
						Expect(props["numberField"].Type).To(Equal("number"))
						Expect(props["stringField"].Type).To(Equal("string"))
					})

					It("can update existing properties types", func() {
						r.run("update", "api", "-p", "numberField:number", "--property", "stringField:string", "-p", "wontchange:string", "--dir", dir)
						r.run("update", "api", "-p", "numberField:string", "--property", "stringField:number", "--dir", dir)
						matchPromise(dir, "postgresql", "syntasso.io", "v1alpha1", "Database", "database", "databases")
						props := getCRDProperties(dir)
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
						props := getCRDProperties(dir)
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

		})
	})

	Describe("add", func() {
		When("it is called without a subcommand", func() {
			It("prints the help", func() {
				session := r.run("add", "--help")
				Expect(session.Out).To(gbytes.Say("Command to add to Kratix resources"))
			})
		})

		Context("container", func() {
			When("called with --help", func() {
				It("prints the help", func() {
					session := r.run("add", "container", "--help")
					Expect(session.Out).To(gbytes.Say("kratix add container LIFECYCLE/TRIGGER/PIPELINE-NAME"))
				})
			})
			When("adding a container", func() {
				var dir string
				AfterEach(func() {
					os.RemoveAll(dir)
				})

				BeforeEach(func() {
					var err error
					dir, err = os.MkdirTemp("", "kratix-update-api-test")
					Expect(err).NotTo(HaveOccurred())

					sess := r.run("init", "promise", "postgresql", "--group", "syntasso.io", "--kind", "Database", "--output-dir", dir)
					Expect(sess.Out).To(gbytes.Say("postgresql promise bootstrapped in"))
				})

				It("adds containers to promise workflows", func() {
					sess := r.run("add", "container", "promise/configure/pipeline0", "--image", "project/image:latest", "--dir", dir)
					Expect(sess.Out).To(gbytes.Say("generated the promise/configure/pipeline0/image"))
					r.run("add", "container", "promise/configure/pipeline1", "--image", "project/image1:latest", "-n", "a-good-container", "--dir", dir)
					r.run("add", "container", "promise/delete/pipeline0", "--image", "project/cleanup:latest", "--dir", dir)

					pipelines := getWorkflows(dir)
					Expect(pipelines.ConfigureResource).To(HaveLen(0))
					Expect(pipelines.DeleteResource).To(HaveLen(0))
					Expect(pipelines.ConfigurePromise).To(HaveLen(2))
					Expect(pipelines.DeletePromise).To(HaveLen(1))

					Expect(pipelines.ConfigurePromise[0].Name).To(Equal("pipeline0"))
					Expect(pipelines.ConfigurePromise[0].Spec.Containers).To(HaveLen(1))
					Expect(pipelines.ConfigurePromise[0].Spec.Containers[0].Image).To(Equal("project/image:latest"))
					Expect(pipelines.ConfigurePromise[0].Spec.Containers[0].Name).To(Equal("image"))
					Expect(pipelines.ConfigurePromise[1].Name).To(Equal("pipeline1"))
					Expect(pipelines.ConfigurePromise[1].Spec.Containers).To(HaveLen(1))
					Expect(pipelines.ConfigurePromise[1].Spec.Containers[0].Image).To(Equal("project/image1:latest"))
					Expect(pipelines.ConfigurePromise[1].Spec.Containers[0].Name).To(Equal("a-good-container"))

					Expect(pipelines.DeletePromise[0].Name).To(Equal("pipeline0"))
					Expect(pipelines.DeletePromise[0].Spec.Containers).To(HaveLen(1))
					Expect(pipelines.DeletePromise[0].Spec.Containers[0].Image).To(Equal("project/cleanup:latest"))
					Expect(pipelines.DeletePromise[0].Spec.Containers[0].Name).To(Equal("cleanup"))

					Expect(sess.Out).To(gbytes.Say("Customise your container by editing the workflows/promise/configure/pipeline0/scripts/pipeline.sh"))
					// script := getPromiseScript(dir)
					// Expect(script).To(ContainSubstring("Hello from ${name} ${namespace}"))
				})

				It("adds containers to resource workflows", func() {
					r.run("add", "container", "resource/configure/pipeline0", "--image", "project/image1:latest", "-n", "a-great-container", "--dir", dir)
					r.run("add", "container", "resource/configure/pipeline0", "--image", "project/image2:latest", "--dir", dir)
					r.run("add", "container", "resource/delete/pipeline0", "--image", "project/cleanup:latest", "--dir", dir)

					pipelines := getWorkflows(dir)
					Expect(pipelines.ConfigurePromise).To(HaveLen(0))
					Expect(pipelines.DeletePromise).To(HaveLen(0))

					Expect(pipelines.ConfigureResource).To(HaveLen(1))
					Expect(pipelines.ConfigureResource[0].Name).To(Equal("pipeline0"))
					Expect(pipelines.ConfigureResource[0].Spec.Containers).To(HaveLen(2))
					Expect(pipelines.ConfigureResource[0].Spec.Containers[0].Image).To(Equal("project/image1:latest"))
					Expect(pipelines.ConfigureResource[0].Spec.Containers[0].Name).To(Equal("a-great-container"))
					Expect(pipelines.ConfigureResource[0].Spec.Containers[1].Image).To(Equal("project/image2:latest"))
					Expect(pipelines.ConfigureResource[0].Spec.Containers[1].Name).To(Equal("image2"))

					Expect(pipelines.DeleteResource).To(HaveLen(1))
					Expect(pipelines.DeleteResource[0].Name).To(Equal("pipeline0"))
					Expect(pipelines.DeleteResource[0].Spec.Containers).To(HaveLen(1))
					Expect(pipelines.DeleteResource[0].Spec.Containers[0].Image).To(Equal("project/cleanup:latest"))
					Expect(pipelines.DeleteResource[0].Spec.Containers[0].Name).To(Equal("cleanup"))
				})
			})
		})
	})
})

func matchPromise(dir, name, group, version, kind, singular, plural string) {
	promiseYAML, err := os.ReadFile(filepath.Join(dir, "promise.yaml"))
	ExpectWithOffset(1, err).NotTo(HaveOccurred())

	var promise v1alpha1.Promise
	ExpectWithOffset(1, yaml.Unmarshal(promiseYAML, &promise)).To(Succeed())

	ExpectWithOffset(1, promise.Name).To(Equal(name))
	promiseCRD, err := promise.GetAPIAsCRD()
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	matchCRD(promiseCRD, group, version, kind, singular, plural)
}

func getCRDProperties(dir string) map[string]apiextensionsv1.JSONSchemaProps {
	promiseYAML, err := os.ReadFile(filepath.Join(dir, "promise.yaml"))
	ExpectWithOffset(1, err).NotTo(HaveOccurred())

	var promise v1alpha1.Promise
	ExpectWithOffset(1, yaml.Unmarshal(promiseYAML, &promise)).To(Succeed())
	crd, err := promise.GetAPIAsCRD()
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	return crd.Spec.Versions[0].Schema.OpenAPIV3Schema.Properties["spec"].Properties
}

func matchCRD(promiseCRD *apiextensionsv1.CustomResourceDefinition, group, version, kind, singular, plural string) {
	ExpectWithOffset(1, promiseCRD.Spec.Group).To(Equal(group))
	ExpectWithOffset(1, promiseCRD.Spec.Names).To(Equal(apiextensionsv1.CustomResourceDefinitionNames{
		Kind:     kind,
		Singular: singular,
		Plural:   plural,
	}))
	ExpectWithOffset(1, promiseCRD.Spec.Versions).To(HaveLen(1))
	ExpectWithOffset(1, promiseCRD.Spec.Versions[0].Name).To(Equal(version))
	ExpectWithOffset(1, promiseCRD.Spec.Versions[0].Served).To(BeTrue())
	ExpectWithOffset(1, promiseCRD.Spec.Versions[0].Storage).To(BeTrue())
}

func getWorkflows(dir string) v1alpha1.PromisePipelines {
	promiseYAML, err := os.ReadFile(filepath.Join(dir, "promise.yaml"))
	ExpectWithOffset(1, err).NotTo(HaveOccurred())

	var promise v1alpha1.Promise
	ExpectWithOffset(1, yaml.Unmarshal(promiseYAML, &promise)).To(Succeed())

	pipelines, err := promise.GeneratePipelines(ctrl.LoggerFrom(context.Background()))
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	return pipelines
}

func getPromiseScript(dir string) string {
	promiseYAML, err := os.ReadFile(filepath.Join(dir, "promise.sh"))
	ExpectWithOffset(1, err).NotTo(HaveOccurred())

	return string(promiseYAML)
}

func matchExampleResource(dir, name, group, version, kind string) {
	exampleResourceYAML, err := os.ReadFile(filepath.Join(dir, "example-resource.yaml"))
	ExpectWithOffset(1, err).NotTo(HaveOccurred())

	var exampleResource unstructured.Unstructured
	ExpectWithOffset(1, yaml.Unmarshal(exampleResourceYAML, &exampleResource)).To(Succeed())

	ExpectWithOffset(1, exampleResource.GetKind()).To(Equal(kind))
	ExpectWithOffset(1, exampleResource.GetAPIVersion()).To(Equal(group + "/" + version))
	ExpectWithOffset(1, exampleResource.GetName()).To(Equal(name))
}

type runner struct {
	exitCode int
	dir      string
}

func withExitCode(exitCode int) *runner {
	return &runner{exitCode: exitCode}
}

func (r *runner) run(args ...string) *gexec.Session {
	cmd := exec.Command(binaryPath, args...)
	cmd.Dir = r.dir
	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	EventuallyWithOffset(1, session).Should(gexec.Exit(r.exitCode))
	return session
}
