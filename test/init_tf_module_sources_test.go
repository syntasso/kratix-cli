package integration_test

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"sigs.k8s.io/yaml"
)

var _ = Describe("InitTerraformPromise module sources", func() {
	const testVarsFile = "assets/terraform/vars/simple.hcl"

	var (
		r          *runner
		workingDir string
		flags      map[string]string
	)

	BeforeEach(func() {
		var err error
		workingDir, err = os.MkdirTemp("", "kratix-test-sources")
		Expect(err).NotTo(HaveOccurred())

		r = &runner{exitCode: 0}
		flags = map[string]string{
			"--group":   "example.com",
			"--kind":    "Example",
			"--version": "v1alpha1",
			"--dir":     workingDir,
			"--split":   "",
		}
	})

	AfterEach(func() {
		os.Unsetenv("KRATIX_TEST_TF_VARS_FILE")
		Expect(os.RemoveAll(workingDir)).To(Succeed())
	})

	getWorkflowEnvs := func() map[string]string {
		workflowsPath := filepath.Join(workingDir, "workflows", "resource", "configure", "workflow.yaml")
		bytes, err := os.ReadFile(workflowsPath)
		Expect(err).NotTo(HaveOccurred())

		var pipelines []map[string]any
		Expect(yaml.Unmarshal(bytes, &pipelines)).To(Succeed())
		Expect(pipelines).ToNot(BeEmpty())

		spec, ok := pipelines[0]["spec"].(map[string]any)
		Expect(ok).To(BeTrue())
		containers, ok := spec["containers"].([]any)
		Expect(ok).To(BeTrue())
		Expect(containers).ToNot(BeEmpty())
		firstContainer, ok := containers[0].(map[string]any)
		Expect(ok).To(BeTrue())
		envList, ok := firstContainer["env"].([]any)
		Expect(ok).To(BeTrue())

		envs := map[string]string{}
		for _, item := range envList {
			entry, ok := item.(map[string]any)
			Expect(ok).To(BeTrue())
			name, _ := entry["name"].(string)
			value, _ := entry["value"].(string)
			envs[name] = value
		}
		return envs
	}

	DescribeTable("supports module sources and registry versions",
		func(source, registryVersion string, expectVersion bool) {
			absVarsFile, err := filepath.Abs(testVarsFile)
			Expect(err).NotTo(HaveOccurred())
			os.Setenv("KRATIX_TEST_TF_VARS_FILE", absVarsFile)

			flags["--module-source"] = source
			if registryVersion != "" {
				flags["--module-registry-version"] = registryVersion
			} else {
				delete(flags, "--module-registry-version")
			}

			r.flags = flags
			session := r.run("init", "tf-module-promise", "example")
			Expect(session).To(gexec.Exit(0))
			Expect(session.Out).To(gbytes.Say("Promise generated successfully"))

			envs := getWorkflowEnvs()
			Expect(envs["MODULE_SOURCE"]).To(Equal(source))
			if expectVersion {
				Expect(envs).To(HaveKeyWithValue("MODULE_REGISTRY_VERSION", registryVersion))
			} else {
				Expect(envs).NotTo(HaveKey("MODULE_REGISTRY_VERSION"))
			}
		},
		Entry("open source git repo with embedded ref", "git::https://github.com/example/open.git?ref=v1.0.0", "", false),
		Entry("private git repo placeholder with ref", "git::ssh://git@github.com/example/private.git?ref=v0.1.0", "", false),
		Entry("registry with version", "terraform-aws-modules/vpc/aws", "5.0.0", true),
		Entry("nested registry with version", "acme/networking/vpc/aws", "3.2.1", true),
		Entry("registry without version", "terraform-providers/random/aws", "", false),
	)
})
