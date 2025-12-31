package integration_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"sigs.k8s.io/yaml"
)

type moduleTestCase struct {
	name               string
	moduleSource       string
	moduleRegistryVer  string
	expectRegistryEnv  bool
	expectedProperties []string
	expectFailure      bool
	skip               bool
}

var _ = Describe("InitTerraformPromise source integration", func() {
	if _, err := exec.LookPath("terraform"); err != nil {
		Skip("terraform binary not found in PATH; skipping module source integration tests")
	}

	cwd, err := os.Getwd()
	Expect(err).NotTo(HaveOccurred())
	localModulePath := filepath.Join(cwd, "assets", "terraform", "modules", "local", "basic")

	cases := []moduleTestCase{
		{
			name:               "open source git repo with ref",
			moduleSource:       "git::https://github.com/terraform-aws-modules/terraform-aws-vpc.git?ref=v5.7.0",
			expectedProperties: []string{"name", "cidr"},
		},
		{
			name:               "private git repo placeholder",
			moduleSource:       "git::ssh://git@github.com/example/private-repo.git?ref=v0.1.0",
			expectedProperties: []string{},
			expectFailure:      true,
			skip:               true, // enable once credentials/repo exist
		},
		{
			name:               "registry with version",
			moduleSource:       "terraform-aws-modules/s3-bucket/aws",
			moduleRegistryVer:  "4.1.2",
			expectRegistryEnv:  true,
			expectedProperties: []string{"bucket"},
		},
		{
			name:               "nested registry module path with version",
			moduleSource:       "terraform-aws-modules/vpc/aws//modules/vpc-endpoints",
			moduleRegistryVer:  "5.7.0",
			expectRegistryEnv:  true,
			expectedProperties: []string{"vpc_id"},
		},
		{
			name:               "registry without version",
			moduleSource:       "terraform-aws-modules/vpc/aws",
			expectedProperties: []string{"name", "cidr"},
		},
		{
			name:               "local filesystem module",
			moduleSource:       localModulePath,
			expectedProperties: []string{"name", "size", "tags"},
		},
	}

	var (
		r          *runner
		workingDir string
		flags      map[string]string
	)

	BeforeEach(func() {
		var err error
		workingDir, err = os.MkdirTemp("", "kratix-test-sources")
		Expect(err).NotTo(HaveOccurred())

		r = &runner{exitCode: 0, Path: os.Getenv("PATH")}
		flags = map[string]string{
			"--group":   "example.com",
			"--kind":    "Example",
			"--version": "v1alpha1",
			"--dir":     workingDir,
			"--split":   "",
		}
	})

	AfterEach(func() {
		Expect(os.RemoveAll(workingDir)).To(Succeed())
	})

	for _, tc := range cases {
		tc := tc
		It(fmt.Sprintf("handles %s", tc.name), func() {
			if tc.skip {
				Skip("pending setup for private repo")
			}

			flags["--module-source"] = tc.moduleSource
			if tc.moduleRegistryVer != "" {
				flags["--module-registry-version"] = tc.moduleRegistryVer
			} else {
				delete(flags, "--module-registry-version")
			}
			r.flags = flags
			runnerArgs := []string{"init", "tf-module-promise", "example"}

			session := r.run(runnerArgs...)
			if tc.expectFailure {
				Expect(session.ExitCode()).NotTo(Equal(0))
				return
			}

			Expect(session).To(gexec.Exit(0))
			Expect(session.Out).To(gbytes.Say("Promise generated successfully"))

			envs := readWorkflowEnvs(workingDir)
			Expect(envs["MODULE_SOURCE"]).To(Equal(tc.moduleSource))
			if tc.expectRegistryEnv {
				Expect(envs).To(HaveKeyWithValue("MODULE_REGISTRY_VERSION", tc.moduleRegistryVer))
			} else {
				Expect(envs).NotTo(HaveKey("MODULE_REGISTRY_VERSION"))
			}

			specProps := readSpecProperties(workingDir)
			for _, prop := range tc.expectedProperties {
				Expect(specProps).To(HaveKey(prop))
			}
		})
	}
})

func readWorkflowEnvs(workingDir string) map[string]string {
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

func readSpecProperties(workingDir string) map[string]any {
	apiPath := filepath.Join(workingDir, "api.yaml")
	contents, err := os.ReadFile(apiPath)
	if err != nil {
		// fallback to promise.yaml when split not used
		contents, err = os.ReadFile(filepath.Join(workingDir, "promise.yaml"))
		Expect(err).NotTo(HaveOccurred())
	}

	var doc map[string]any
	Expect(yaml.Unmarshal(contents, &doc)).To(Succeed())

	// If promise.yaml, drill into spec.api
	if kind, _ := doc["kind"].(string); kind == "Promise" {
		spec, _ := doc["spec"].(map[string]any)
		api, _ := spec["api"].(map[string]any)
		doc = api
	}

	spec, _ := doc["spec"].(map[string]any)
	versions, _ := spec["versions"].([]any)
	Expect(versions).ToNot(BeEmpty())
	firstVersion, _ := versions[0].(map[string]any)
	schema, _ := firstVersion["schema"].(map[string]any)
	openAPISchema, _ := schema["openAPIV3Schema"].(map[string]any)
	properties, _ := openAPISchema["properties"].(map[string]any)
	specProps, _ := properties["spec"].(map[string]any)
	propMap, _ := specProps["properties"].(map[string]any)
	Expect(propMap).ToNot(BeNil())
	return propMap
}
