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
		DescribeTable("generating a promise with different XRDs",
			func(xrdPath, expectedOutputDir string) {
				if xrdPath != "" {
					r.flags["--xrd"] = xrdPath
				}
				session = r.run(initPromiseCmd...)
				generatedFiles = getFiles(workingDir)
				files := []string{"promise.yaml", "example-resource.yaml", "README.md"}
				Expect(generatedFiles).To(ConsistOf(files))
				expectFilesEqual(workingDir, expectedOutputDir, files)
				Expect(session.Out).To(SatisfyAll(
					gbytes.Say(`Promise generated successfully.`),
				))
			},
			Entry("with spec.properties", "", "assets/crossplane/expected-output"),
			Entry("with empty openAPIV3Schema", "assets/crossplane/xrd-with-empty-openAPIV3Schema.yaml", "assets/crossplane/expected-output-with-empty-openAPIV3Schema"),
			Entry("with no spec.properties", "assets/crossplane/xrd-with-no-spec-properties.yaml", "assets/crossplane/expected-output-with-no-spec-properties"),
		)

		Describe("with the --split flag", func() {
			BeforeEach(func() {
				r.flags["--split"] = ""
				session = r.run(initPromiseCmd...)
				generatedFiles = getFiles(workingDir)
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
				generatedFiles = getFiles(workingDir)
			})

			It("generates the expected files", func() {
				files := []string{"promise.yaml", "example-resource.yaml", "README.md"}
				Expect(generatedFiles).To(ConsistOf(files))
				expectFilesEqual(workingDir, "assets/crossplane/expected-output-with-compositions", files)
				Expect(session.Out).To(SatisfyAll(
					gbytes.Say(`Promise generated successfully.`),
				))
			})
		})

		Describe("with --skip-dependencies", func() {
			BeforeEach(func() {
				r.flags["--skip-dependencies"] = ""
				session = r.run(initPromiseCmd...)
				generatedFiles = getFiles(workingDir)
			})

			It("generates a single promise.yaml", func() {
				files := []string{"promise.yaml", "example-resource.yaml", "README.md"}
				Expect(generatedFiles).To(ConsistOf(files))
				expectFilesEqual(workingDir, "assets/crossplane/expected-output-with-skip-dependencies", files)
				Expect(session.Out).To(SatisfyAll(
					gbytes.Say(`Promise generated successfully.`),
				))
			})
		})
	})
})

func getFiles(dir string) []string {
	fileEntries, err := os.ReadDir(dir)
	Expect(err).ToNot(HaveOccurred())
	var files []string
	for _, fileEntry := range fileEntries {
		files = append(files, fileEntry.Name())
	}
	return files
}

func expectFilesEqual(actualDir, expectedDir string, files []string) {
	for _, file := range files {
		Expect(cat(filepath.Join(actualDir, file))).To(Equal(cat(filepath.Join(expectedDir, file))))
	}
}
