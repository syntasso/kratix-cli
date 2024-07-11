package integration_test

import (
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"os"
	"path/filepath"
	"sigs.k8s.io/yaml"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/syntasso/kratix/api/v1alpha1"
)

var _ = Describe("init helm-promise", func() {
	var (
		r          *runner
		workingDir string
	)

	BeforeEach(func() {
		var err error
		workingDir, err = os.MkdirTemp("", "kratix-test")
		Expect(err).NotTo(HaveOccurred())
		r = &runner{exitCode: 0, dir: workingDir, timeout: 5 * time.Second}
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

	When("called with --split", func() {
		It("bootstraps correctly", func() {
			session := r.run("init", "helm-promise", "postgresql", "--chart-url", "https://helm.github.io/examples", "--chart-name", "hello-world", "--group", "syntasso.io", "--kind", "Database", "--split")
			Expect(session.Out).To(gbytes.Say("postgresql promise bootstrapped in the current directory"))

			files, err := os.ReadDir(workingDir)
			Expect(err).NotTo(HaveOccurred())
			Expect(files).To(HaveLen(5))

			By("generating an example-resource.yaml file", func() {
				matchExampleResource(workingDir, "example-postgresql", "syntasso.io", "v1alpha1", "Database")
			})

			By("generating a README file", func() {
				readmeContents, err := os.ReadFile(filepath.Join(workingDir, "README.md"))
				Expect(err).NotTo(HaveOccurred())
				Expect(readmeContents).To(ContainSubstring("kratix init helm-promise postgresql"))
			})

			By("generating an empty dependencies file", func() {
				depContent, err := os.ReadFile(filepath.Join(workingDir, "dependencies.yaml"))
				Expect(err).NotTo(HaveOccurred())
				Expect(depContent).To(BeEmpty())
			})

			By("generating a workflow file with helm resource configure workflow", func() {
				pipelines := getPipelines(workingDir)
				Expect(pipelines).To(HaveLen(1))
				matchHelmResourceConfigurePipeline(pipelines[0], []corev1.EnvVar{
					{Name: "CHART_URL", Value: "https://helm.github.io/examples"},
					{Name: "CHART_NAME", Value: "hello-world"}})
			})

			By("including correct gvk and CRD schema in a api.yaml", func() {
				apiYAML, err := os.ReadFile(filepath.Join(workingDir, "api.yaml"))
				Expect(err).NotTo(HaveOccurred())
				var promiseCRD apiextensionsv1.CustomResourceDefinition
				Expect(yaml.Unmarshal(apiYAML, &promiseCRD)).To(Succeed())
				matchCRD(&promiseCRD, "syntasso.io", "v1alpha1", "Database", "database", "databases")

				props := getCRDProperties(workingDir, true)
				matchExampleHelloWorldHelmChartSchema(props)
			})
		})
	})

	When("called without '--split' flag", func() {
		It("bootstraps promise correctly", func() {
			session := r.run("init", "helm-promise", "--chart-url", "https://helm.github.io/examples", "--chart-version", "0.1.0", "--chart-name", "hello-world", "postgresql", "--group", "syntasso.io", "--kind", "Database")
			Expect(session.Out).To(gbytes.Say("postgresql promise bootstrapped in the current directory"))
			files, err := os.ReadDir(workingDir)
			Expect(err).NotTo(HaveOccurred())
			Expect(files).To(HaveLen(3))

			By("generating an example-resource.yaml file", func() {
				matchExampleResource(workingDir, "example-postgresql", "syntasso.io", "v1alpha1", "Database")
			})

			By("generating a README file", func() {
				readmeContents, err := os.ReadFile(filepath.Join(workingDir, "README.md"))
				Expect(err).NotTo(HaveOccurred())
				Expect(readmeContents).To(ContainSubstring("kratix init helm-promise postgresql"))
			})

			By("including GVK in promise.yaml", func() {
				matchPromise(workingDir, "postgresql", "syntasso.io", "v1alpha1", "Database", "database", "databases")
			})

			By("including CRD schema from chart values in promise.yaml", func() {
				props := getCRDProperties(workingDir, false)
				matchExampleHelloWorldHelmChartSchema(props)
			})

			By("including helm resource configure workflow in promise.yaml", func() {
				pipelines := getWorkflows(workingDir)["resource"]["configure"]
				Expect(pipelines).To(HaveLen(1))
				matchHelmResourceConfigurePipeline(pipelines[0], []corev1.EnvVar{
					{Name: "CHART_URL", Value: "https://helm.github.io/examples"},
					{Name: "CHART_NAME", Value: "hello-world"},
					{Name: "CHART_VERSION", Value: "0.1.0"},
				})
			})
		})
	})

	Context("helm chart integration", func() {
		It("works with OCI helm chart", func() {
			session := r.run("init", "helm-promise", "--chart-url", "oci://registry-1.docker.io/bitnamicharts/redis", "--chart-version", "19.6.0", "redis", "--group", "syntasso.io", "--kind", "Database")
			Expect(session.Out).To(gbytes.Say("redis promise bootstrapped in the current directory"))

			By("including correct env var", func() {
				pipelines := getWorkflows(workingDir)["resource"]["configure"]
				Expect(pipelines).To(HaveLen(1))
				matchHelmResourceConfigurePipeline(pipelines[0], []corev1.EnvVar{
					{Name: "CHART_URL", Value: "oci://registry-1.docker.io/bitnamicharts/redis"},
					{Name: "CHART_VERSION", Value: "19.6.0"},
				})
			})

			By("including CRD schema from chart values in promise.yaml", func() {
				props := getCRDProperties(workingDir, false)
				ExpectWithOffset(1, props).To(SatisfyAll(
					HaveKey("kubeVersion"),
					HaveKey("global"),
					HaveKey("image")))
			})
		})

		It("works with tar helm chart", func() {
			session := r.run("init", "helm-promise", "--chart-url", "https://github.com/fluxcd-community/helm-charts/releases/download/flux2-sync-1.9.0/flux2-sync-1.9.0.tgz", "flux", "--group", "syntasso.io", "--kind", "Cicd")
			Expect(session.Out).To(gbytes.Say("flux promise bootstrapped in the current directory"))

			By("including correct env var", func() {
				pipelines := getWorkflows(workingDir)["resource"]["configure"]
				Expect(pipelines).To(HaveLen(1))
				matchHelmResourceConfigurePipeline(pipelines[0], []corev1.EnvVar{
					{Name: "CHART_URL", Value: "https://github.com/fluxcd-community/helm-charts/releases/download/flux2-sync-1.9.0/flux2-sync-1.9.0.tgz"},
				})
			})

			By("including CRD schema from chart values in promise.yaml", func() {
				props := getCRDProperties(workingDir, false)
				ExpectWithOffset(1, props).To(SatisfyAll(
					HaveKey("secret"),
					HaveKey("cli"),
					HaveKey("gitRepository")))
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
	ExpectWithOffset(1, pipeline.Spec.Containers[0].Image).To(Equal("ghcr.io/syntasso/kratix-cli/helm-resource-configure:v0.1.0"))
	ExpectWithOffset(1, pipeline.Spec.Containers[0].Env).To(ConsistOf(vars))
}

// helm chart: https://github.com/helm/examples/tree/main/charts/hello-world
func matchExampleHelloWorldHelmChartSchema(props map[string]apiextensionsv1.JSONSchemaProps) {
	ExpectWithOffset(1, props).To(SatisfyAll(
		HaveKey("replicaCount"),
		HaveKey("nameOverride"),
		HaveKey("fullnameOverride"),
		HaveKey("serviceAccount"),
		HaveKey("service")))
	ExpectWithOffset(1, props["replicaCount"].Type).To(Equal("number"))
	ExpectWithOffset(1, props["service"].Type).To(Equal("object"))
	ExpectWithOffset(1, props["serviceAccount"].Type).To(Equal("object"))
	ExpectWithOffset(1, props["serviceAccount"].Properties["create"].Type).To(Equal("boolean"))
	ExpectWithOffset(1, props["serviceAccount"].Properties["annotations"].Type).To(Equal("object"))
	ExpectWithOffset(1, props["serviceAccount"].Properties["name"].Type).To(Equal("string"))
}
