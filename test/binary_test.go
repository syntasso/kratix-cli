package integration_test

import (
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
			Describe("promise", func() {
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
					Expect(files).To(HaveLen((3)))

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
	Expect(err).NotTo(HaveOccurred())
	Eventually(session).Should(gexec.Exit(r.exitCode))
	return session
}
