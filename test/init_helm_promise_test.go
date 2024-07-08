package integration_test

import (
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"os"
	"path/filepath"
	"sigs.k8s.io/yaml"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/syntasso/kratix/api/v1alpha1"
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

	When("called without the required flag", func() {
		It("raises an error", func() {
			r.exitCode = 1
			Expect(r.run("init", "helm-promise", "postgresql", "--group", "syntasso.io", "--kind", "Database").Err).To(SatisfyAll(
				gbytes.Say(`required flag\(s\) "chart-url" not set`),
			))
		})
	})

	Context("--split", func() {
		When("called with the required arguments", func() {
			It("generates correct files", func() {
				session := r.run("init", "helm-promise", "postgresql", "--chart-url", "postgres.io/chart", "--group", "syntasso.io", "--kind", "Database", "--split")
				Expect(session.Out).To(gbytes.Say("postgresql promise bootstrapped in the current directory"))

				files, err := os.ReadDir(workingDir)
				Expect(err).NotTo(HaveOccurred())
				Expect(files).To(HaveLen(5))

				By("generating a api.yaml file", func() {
					apiYAML, err := os.ReadFile(filepath.Join(workingDir, "api.yaml"))
					Expect(err).NotTo(HaveOccurred())
					var promiseCRD apiextensionsv1.CustomResourceDefinition
					Expect(yaml.Unmarshal(apiYAML, &promiseCRD)).To(Succeed())
					matchCRD(&promiseCRD, "syntasso.io", "v1alpha1", "Database", "database", "databases")
				})

				By("generating an example-resource.yaml file", func() {
					matchExampleResource(workingDir, "example-postgresql", "syntasso.io", "v1alpha1", "Database")
				})

				By("including a README file", func() {
					readmeContents, err := os.ReadFile(filepath.Join(workingDir, "README.md"))
					Expect(err).NotTo(HaveOccurred())
					Expect(readmeContents).To(ContainSubstring("kratix init helm-promise postgresql"))
				})

				By("writing a workflow file", func() {
					pipelines := getPipelines(workingDir)
					Expect(pipelines).To(HaveLen(1))
					matchHelmResourceConfigurePipeline(pipelines[0], []corev1.EnvVar{{Name: "CHART_URL", Value: "postgres.io/chart"}})
				})
			})
		})
	})

	Context("promise.yaml", func() {
		When("called with the required arguments", func() {
			It("generates correct files", func() {
				session := r.run("init", "helm-promise", "--chart-url", "postgres.io/chart", "--chart-version", "v110.0.1", "--chart-name", "greatchart", "postgresql", "--group", "syntasso.io", "--kind", "Database")
				Expect(session.Out).To(gbytes.Say("postgresql promise bootstrapped in the current directory"))

				files, err := os.ReadDir(workingDir)
				Expect(err).NotTo(HaveOccurred())
				Expect(files).To(HaveLen(3))

				matchPromise(workingDir, "postgresql", "syntasso.io", "v1alpha1", "Database", "database", "databases")
				pipelines := getWorkflows(workingDir)["resource"]["configure"]
				Expect(pipelines).To(HaveLen(1))
				matchHelmResourceConfigurePipeline(pipelines[0], []corev1.EnvVar{
					{Name: "CHART_URL", Value: "postgres.io/chart"},
					{Name: "CHART_NAME", Value: "greatchart"},
					{Name: "CHART_VERSION", Value: "v110.0.1"},
				})
			})
		})
	})
})

func getPipelines(dir string) []v1alpha1.Pipeline {
	workflowContent, err := os.ReadFile(filepath.Join(dir, "workflows", "resource", "configure", "workflow.yaml"))
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	var pipelines []v1alpha1.Pipeline
	Expect(yaml.Unmarshal(workflowContent, &pipelines)).To(Succeed())
	return pipelines
}

func matchHelmResourceConfigurePipeline(pipeline v1alpha1.Pipeline, vars []corev1.EnvVar) {
	ExpectWithOffset(1, pipeline.Spec.Containers).To(HaveLen(1))
	ExpectWithOffset(1, pipeline.Spec.Containers[0].Name).To(Equal("instance-configure"))
	ExpectWithOffset(1, pipeline.Spec.Containers[0].Image).To(Equal("ghcr.io/syntasso/kratix-cli/helm-instance-configure:v0.1.0"))
	ExpectWithOffset(1, pipeline.Spec.Containers[0].Env).To(ConsistOf(vars))
}
