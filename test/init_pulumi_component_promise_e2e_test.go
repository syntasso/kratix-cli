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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
)

var _ = Describe("init pulumi-component-promise end-to-end preview flow", func() {
	var (
		r               *runner
		workingDir      string
		stageBinaryPath string
	)

	BeforeEach(func() {
		var err error
		workingDir, err = os.MkdirTemp("", "kratix-pulumi-e2e")
		Expect(err).NotTo(HaveOccurred())
		r = &runner{exitCode: 0, dir: workingDir}

		schemaBytes, err := os.ReadFile("assets/pulumi/schema.valid.json")
		Expect(err).NotTo(HaveOccurred())
		Expect(os.WriteFile(filepath.Join(workingDir, "schema.valid.json"), schemaBytes, 0o600)).To(Succeed())
	})

	AfterEach(func() {
		Expect(os.RemoveAll(workingDir)).To(Succeed())
	})

	It("generates split output wired to the stage runtime and emits a Program CR from the generated example resource", func() {
		session := r.run(
			"init", "pulumi-component-promise", "mypromise",
			"--schema", "./schema.valid.json",
			"--group", "syntasso.io",
			"--kind", "Database",
			"--split",
		)

		Expect(session.Out).To(SatisfyAll(
			gbytes.Say("Preview: This command is in preview"),
			gbytes.Say("Pulumi component Promise generated successfully."),
		))

		workflowBytes, err := os.ReadFile(filepath.Join(workingDir, "workflows/resource/configure/workflow.yaml"))
		Expect(err).NotTo(HaveOccurred())

		var pipelines []v1alpha1.Pipeline
		Expect(yaml.Unmarshal(workflowBytes, &pipelines)).To(Succeed())
		Expect(pipelines).To(HaveLen(1))
		Expect(pipelines[0].Spec.Containers).To(HaveLen(1))

		container := pipelines[0].Spec.Containers[0]
		Expect(container.Name).To(Equal("from-api-to-pulumi-pko-program"))
		Expect(container.Image).To(Equal("ghcr.io/syntasso/kratix-cli/from-api-to-pulumi-pko-program:v0.1.0"))
		Expect(container.Env).To(ContainElements(
			corev1.EnvVar{Name: "PULUMI_COMPONENT_TOKEN", Value: "pkg:index:Database"},
			corev1.EnvVar{Name: "PULUMI_SCHEMA_SOURCE", Value: "./schema.valid.json"},
		))

		stageBinaryPath, err = gexec.Build("github.com/syntasso/kratix-cli/stages/pulumi-promise")
		Expect(err).NotTo(HaveOccurred())

		outputPath := filepath.Join(workingDir, "program-output.yaml")
		stageCmd := exec.Command(stageBinaryPath)
		stageCmd.Env = append(os.Environ(),
			"KRATIX_INPUT_FILE="+filepath.Join(workingDir, "example-resource.yaml"),
			"KRATIX_OUTPUT_FILE="+outputPath,
			"PULUMI_COMPONENT_TOKEN=pkg:index:Database",
		)

		stageSession, err := gexec.Start(stageCmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(stageSession, "10s").Should(gexec.Exit(0))

		outputBytes, err := os.ReadFile(outputPath)
		Expect(err).NotTo(HaveOccurred())

		inputBytes, err := os.ReadFile(filepath.Join(workingDir, "example-resource.yaml"))
		Expect(err).NotTo(HaveOccurred())

		inputObject := &unstructured.Unstructured{}
		Expect(yaml.Unmarshal(inputBytes, inputObject)).To(Succeed())
		inputSpec, found, err := unstructured.NestedMap(inputObject.Object, "spec")
		Expect(err).NotTo(HaveOccurred())
		Expect(found).To(BeTrue())

		programObject := &unstructured.Unstructured{}
		Expect(yaml.Unmarshal(outputBytes, programObject)).To(Succeed())
		Expect(programObject.GetAPIVersion()).To(Equal("pulumi.com/v1"))
		Expect(programObject.GetKind()).To(Equal("Program"))
		Expect(programObject.GetName()).To(MatchRegexp(`^example-request-[0-9a-f]{8}$`))
		Expect(programObject.GetNamespace()).To(Equal("default"))

		resources, found, err := unstructured.NestedMap(programObject.Object, "spec", "resources")
		Expect(err).NotTo(HaveOccurred())
		Expect(found).To(BeTrue())
		Expect(resources).To(HaveKey("pkg-index-database"))

		componentResource, ok := resources["pkg-index-database"].(map[string]any)
		Expect(ok).To(BeTrue())
		Expect(componentResource).To(HaveKeyWithValue("type", "pkg:index:Database"))
		Expect(componentResource).To(HaveKeyWithValue("properties", inputSpec))
	})
})
