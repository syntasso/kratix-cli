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

var _ = FDescribe("InitterraformPromise", func() {
	var r *runner
	var workingDir string
	var initPromiseCmd []string
	var session *gexec.Session

	//./bin/kratix init tf-module-promise gcr --version v0.16.4 --source https://github.com/GoogleCloudPlatform/terraform-google-cloud-run --group gcp.com --kind cloudrun --dir tmp-cr
	BeforeEach(func() {
		var err error
		workingDir, err = os.MkdirTemp("", "kratix-test")
		Expect(err).NotTo(HaveOccurred())
		r = &runner{exitCode: 0}
		r.flags = map[string]string{
			"--group":          "gcp.com",
			"--kind":           "GoogleCloudRun",
			"--version":        "v2",
			"--dir":            workingDir,
			"--module-version": "v0.16.4",
			"--module-source":  "https://github.com/GoogleCloudPlatform/terraform-google-cloud-run",
		}
		initPromiseCmd = []string{"init", "tf-module-promise", "googlecloudrun"}

	})

	AfterEach(func() {
		Expect(os.RemoveAll(workingDir)).To(Succeed())
	})

	When("called without required flags", func() {
		It("prints an error", func() {
			r.exitCode = 1
			r.flags = map[string]string{}
			session := r.run(initPromiseCmd...)
			Expect(session.Err).To(gbytes.Say(`Error: required flag\(s\) "group", "kind", "module-source", "module-version" not set`))
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

			It("generates the expected files", func() {
				files := []string{"promise.yaml", "example-resource.yaml", "README.md"}
				Expect(generatedFiles).To(ConsistOf(files))
				Expect(cat(filepath.Join(workingDir, "promise.yaml"))).To(Equal(cat("assets/terraform/expected-output/promise.yaml")))
				Expect(cat(filepath.Join(workingDir, "example-resource.yaml"))).To(Equal(cat("assets/terraform/expected-output/example-resource.yaml")))
				Expect(cat(filepath.Join(workingDir, "README.md"))).To(Equal(cat("assets/terraform/expected-output/README.md")))
				Expect(session.Out).To(SatisfyAll(
					gbytes.Say(`Promise generated successfully.`),
				))
			})
		})

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
				Expect(cat(filepath.Join(workingDir, "api.yaml"))).To(Equal(cat("assets/terraform/expected-output-with-split/api.yaml")))
				Expect(cat(filepath.Join(workingDir, "workflows/resource/configure/workflow.yaml"))).To(Equal(cat("assets/terraform/expected-output-with-split/workflows/resource/configure/workflow.yaml")))
				Expect(cat(filepath.Join(workingDir, "example-resource.yaml"))).To(Equal(cat("assets/terraform/expected-output-with-split/example-resource.yaml")))
				Expect(cat(filepath.Join(workingDir, "README.md"))).To(Equal(cat("assets/terraform/expected-output-with-split/README.md")))
				Expect(cat(filepath.Join(workingDir, "dependencies.yaml"))).To(Equal(cat("assets/terraform/expected-output-with-split/dependencies.yaml")))
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
