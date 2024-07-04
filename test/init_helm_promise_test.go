package integration_test

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("init helm-promise", func() {
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

	When("called without any arguments", func() {
		It("raises an error", func() {
			r.exitCode = 1

			Expect(r.run("init", "helm-promise").Err).To(SatisfyAll(
				gbytes.Say(`Error: accepts 1 arg\(s\), received 0`),
			))
		})
	})

	When("called without the --url flag", func() {
		It("raises an error", func() {
			r.exitCode = 1

			Expect(r.run("init", "helm-promise", "postgresql", "--group", "syntasso.io", "--kind", "Database").Err).To(SatisfyAll(
				gbytes.Say(`required flag\(s\) "url" not set`),
			))
		})
	})

	Context("Promise", func() {
		When("called with the required arguments", func() {
			It("generates a promise file, Readme and example resource request file", func() {
				session := r.run("init", "helm-promise", "--url", "postgres.io/chart", "postgresql", "--group", "syntasso.io", "--kind", "Database")
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
					Expect(readmeContents).To(ContainSubstring("kratix init helm-promise postgresql"))
				})
			})
		})
	})

})
