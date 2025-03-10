package integration_test

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("InitCrossplanePromise", func() {
	var r *runner
	var workingDir string
	var initPromiseCmd []string
	var session *gexec.Session

	BeforeEach(func() {
		var err error
		workingDir, err = os.MkdirTemp("", "kratix-test")
		Expect(err).NotTo(HaveOccurred())
		r = &runner{exitCode: 0}
		r.flags = map[string]string{
			"--group":   "syntasso.io",
			"--kind":    "S3Bucket",
			"--version": "v2",
			"--xrd":     "assets/crossplane/xrd.yaml",
			"--dir":     workingDir,
		}
		initPromiseCmd = []string{"init", "crossplane-promise", "s3buckets"}

	})

	AfterEach(func() {
		Expect(os.RemoveAll(workingDir)).To(Succeed())
	})

	When("called without required flags", func() {
		It("prints an error", func() {
			r.exitCode = 1
			r.flags = map[string]string{}
			session := r.run(initPromiseCmd...)
			Expect(session.Err).To(gbytes.Say(`Error: required flag\(s\) "group", "kind", "xrd" not set`))
		})
	})

	When("called without required arguments", func() {
		It("prints an error", func() {
			r.exitCode = 1
			session := r.run("init", "crossplane-promise")
			Expect(session.Err).To(gbytes.Say(`Error: accepts 1 arg\(s\), received 0`))
		})
	})

	When("there is no matching XRD in the manifests directory", func() {
		It("returns an error", func() {
			r.flags["--xrd"] = "does-not-exist"
			r.exitCode = 1
			session := r.run(initPromiseCmd...)
			Expect(session.Err).To(gbytes.Say(`Error: failed to read file does-not-exist`))
		})
	})

	Describe("generating a promise from an crossplane", func() {
		var generatedFiles []string
		Describe("with the --split flag", func() {
			BeforeEach(func() {
				r.flags["--split"] = ""
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
				Expect(cat(filepath.Join(workingDir, "api.yaml"))).To(Equal(cat("assets/crossplane/expected-output-with-split/api.yaml")))
				Expect(cat(filepath.Join(workingDir, "workflows/resource/configure/workflow.yaml"))).To(Equal(cat("assets/crossplane/expected-output-with-split/workflows/resource/configure/workflow.yaml")))
				Expect(cat(filepath.Join(workingDir, "example-resource.yaml"))).To(Equal(cat("assets/crossplane/expected-output-with-split/example-resource.yaml")))
				Expect(cat(filepath.Join(workingDir, "README.md"))).To(Equal(cat("assets/crossplane/expected-output-with-split/README.md")))
				Expect(cat(filepath.Join(workingDir, "dependencies.yaml"))).To(Equal(cat("assets/crossplane/expected-output-with-split/dependencies.yaml")))
				Expect(session.Out).To(SatisfyAll(
					gbytes.Say(`Promise generated successfully.`),
				))
			})
		})

		Describe("with --compositions", func() {
			BeforeEach(func() {
				r.flags["--compositions"] = "assets/crossplane/composition.yaml"
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
				Expect(cat(filepath.Join(workingDir, "promise.yaml"))).To(Equal(cat("assets/crossplane/expected-output-with-compositions/promise.yaml")))
				Expect(cat(filepath.Join(workingDir, "example-resource.yaml"))).To(Equal(cat("assets/crossplane/expected-output-with-compositions/example-resource.yaml")))
				Expect(cat(filepath.Join(workingDir, "README.md"))).To(Equal(cat("assets/crossplane/expected-output-with-compositions/README.md")))
				Expect(session.Out).To(SatisfyAll(
					gbytes.Say(`Promise generated successfully.`),
				))
			})
		})

		Describe("with --skip-dependencies", func() {
			BeforeEach(func() {
				r.flags["--skip-dependencies"] = ""
				session = r.run(initPromiseCmd...)
				fileEntries, err := os.ReadDir(workingDir)
				generatedFiles = []string{}
				for _, fileEntry := range fileEntries {
					generatedFiles = append(generatedFiles, fileEntry.Name())
				}
				Expect(err).ToNot(HaveOccurred())
			})

			It("generates a single promise.yaml", func() {
				files := []string{"promise.yaml", "example-resource.yaml", "README.md"}
				Expect(generatedFiles).To(ConsistOf(files))
				Expect(cat(filepath.Join(workingDir, "promise.yaml"))).To(Equal(cat("assets/crossplane/expected-output-with-skip-dependencies/promise.yaml")))
				Expect(cat(filepath.Join(workingDir, "example-resource.yaml"))).To(Equal(cat("assets/crossplane/expected-output-with-skip-dependencies/example-resource.yaml")))
				Expect(cat(filepath.Join(workingDir, "README.md"))).To(Equal(cat("assets/crossplane/expected-output-with-skip-dependencies/README.md")))
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
