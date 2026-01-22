package integration_test

import (
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("InitTerraformPromise", func() {
	var r *runner
	var workingDir string
	var initPromiseCmd []string
	var dependenciesWorkflowPath string
	var session *gexec.Session

	BeforeEach(func() {
		var err error
		workingDir, err = os.MkdirTemp("", "kratix-test")
		dependenciesWorkflowPath = filepath.Join(workingDir, "workflows", "promise", "configure", "dependencies", "add-tf-dependencies")
		Expect(err).NotTo(HaveOccurred())
		r = &runner{exitCode: 0}
		r.flags = map[string]string{
			"--group":         "gcp.com",
			"--kind":          "GoogleCloudRun",
			"--version":       "v2",
			"--dir":           workingDir,
			"--module-source": "git::https://github.com/syntasso/terraform-google-cloud-run?ref=v0.16.4",
		}
		initPromiseCmd = []string{"init", "tf-module-promise", "googlecloudrun"}
	})

	AfterEach(func() {
		// Expect(os.RemoveAll(workingDir)).To(Succeed())
	})

	When("called without required flags", func() {
		It("prints an error", func() {
			r.exitCode = 1
			r.flags = map[string]string{}
			session := r.run(initPromiseCmd...)
			Expect(session.Err).To(gbytes.Say(`Error: required flag\(s\) "group", "kind", "module-source" not set`))
		})
	})

	When("called without required arguments", func() {
		It("prints an error", func() {
			r.exitCode = 1
			session := r.run("init", "tf-module-promise")
			Expect(session.Err).To(gbytes.Say(`Error: accepts 1 arg\(s\), received 0`))
		})
	})

	Describe("generating a promise from a tf module", func() {
		var generatedFiles []string
		Describe("with no additional flags", func() {
			BeforeEach(func() {
				r.timeout = time.Minute
				session = r.run(initPromiseCmd...)
				fileEntries, err := os.ReadDir(workingDir)
				generatedFiles = []string{}
				for _, fileEntry := range fileEntries {
					generatedFiles = append(generatedFiles, fileEntry.Name())
				}
				Expect(err).ToNot(HaveOccurred())
			})

			FIt("generates the expected files", func() {
				files := []string{"promise.yaml", "example-resource.yaml", "README.md", "workflows"}
				Expect(generatedFiles).To(ConsistOf(files))
				Expect(cat(filepath.Join(workingDir, "promise.yaml"))).To(Equal(cat("assets/terraform/expected-output/promise.yaml")))
				Expect(cat(filepath.Join(workingDir, "example-resource.yaml"))).To(Equal(cat("assets/terraform/expected-output/example-resource.yaml")))
				Expect(cat(filepath.Join(workingDir, "README.md"))).To(Equal(cat("assets/terraform/expected-output/README.md")))
				Expect(cat(filepath.Join(dependenciesWorkflowPath, "Dockerfile"))).To(Equal(cat("assets/terraform/expected-output/promise-workflow/Dockerfile")))
				Expect(cat(filepath.Join(dependenciesWorkflowPath, "resources", "providers.tf"))).To(Equal(cat("assets/terraform/expected-output/promise-workflow/providers.tf")))
				// todo this is not producing the content we want
				Expect(cat(filepath.Join(dependenciesWorkflowPath, "scripts", "pipeline.sh"))).To(Equal(cat("assets/terraform/expected-output/promise-workflow/pipeline.sh")))

				Expect(session.Out).To(SatisfyAll(
					gbytes.Say(`Promise generated successfully.`),
				))
			})
		})

		//TODO - update
		Describe("with the --split flag", func() {
			BeforeEach(func() {
				r.flags["--split"] = ""
				r.timeout = time.Minute
				session = r.run(initPromiseCmd...)
				fileEntries, err := os.ReadDir(workingDir)
				generatedFiles = []string{}
				for _, fileEntry := range fileEntries {
					generatedFiles = append(generatedFiles, fileEntry.Name())
				}
				Expect(err).ToNot(HaveOccurred())
			})

			It("generates the expected files", func() {
				files := []string{"api.yaml", "workflows", "example-resource.yaml", "README.md", "dependencies.yaml"}
				Expect(generatedFiles).To(ConsistOf(files))
				actualApi := cat(filepath.Join(workingDir, "api.yaml"))
				api := cat("assets/terraform/expected-output-with-split/api.yaml")
				Expect(actualApi).To(Equal(api), "actual api %s\n expected api %s\n", actualApi, api)
				Expect(cat(filepath.Join(workingDir, "workflows/resource/configure/workflow.yaml"))).To(Equal(cat("assets/terraform/expected-output-with-split/workflows/resource/configure/workflow.yaml")))
				Expect(cat(filepath.Join(workingDir, "example-resource.yaml"))).To(Equal(cat("assets/terraform/expected-output-with-split/example-resource.yaml")))
				Expect(cat(filepath.Join(workingDir, "README.md"))).To(Equal(cat("assets/terraform/expected-output-with-split/README.md")))
				Expect(cat(filepath.Join(workingDir, "dependencies.yaml"))).To(Equal(cat("assets/terraform/expected-output-with-split/dependencies.yaml")))
				Expect(session.Out).To(SatisfyAll(
					gbytes.Say(`Promise generated successfully.`),
				))
			})
		})

		Describe("with module-path on Cloud Foundation Fabric", func() {
			var vpcCmd []string
			BeforeEach(func() {
				r.flags = map[string]string{
					"--group":         "syntasso.io",
					"--kind":          "VPC",
					"--version":       "v1alpha1",
					"--dir":           workingDir,
					"--module-source": "git::https://github.com/GoogleCloudPlatform/cloud-foundation-fabric.git//modules/api-gateway?ref=v49.1.0",
				}
				vpcCmd = []string{"init", "tf-module-promise", "vpc"}
				r.timeout = time.Minute
				session = r.run(vpcCmd...)
			})

			It("generates the expected files for the module-path scenario", func() {
				fileEntries, err := os.ReadDir(workingDir)
				Expect(err).ToNot(HaveOccurred())
				generatedFiles = []string{}
				for _, fileEntry := range fileEntries {
					generatedFiles = append(generatedFiles, fileEntry.Name())
				}

				// Adjust if your fixture has different files
				files := []string{"promise.yaml", "example-resource.yaml", "README.md"}
				Expect(generatedFiles).To(ConsistOf(files))

				// Compare to prepared fixtures for this scenario
				Expect(cat(filepath.Join(workingDir, "promise.yaml"))).To(Equal(cat("assets/terraform/expected-output-vpc/promise.yaml")))
				Expect(cat(filepath.Join(workingDir, "example-resource.yaml"))).To(Equal(cat("assets/terraform/expected-output-vpc/example-resource.yaml")))
				Expect(cat(filepath.Join(workingDir, "README.md"))).To(Equal(cat("assets/terraform/expected-output-vpc/README.md")))

				Expect(session.Out).To(SatisfyAll(
					gbytes.Say(`Promise generated successfully.`),
				))
			})
		})
	})
})

func cat(file string) string {
	content, err := os.ReadFile(file)
	ExpectWithOffset(1, err).ToNot(HaveOccurred())
	return string(content)
}
