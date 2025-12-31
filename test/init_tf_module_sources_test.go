package integration_test

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"sigs.k8s.io/yaml"
)

type moduleTestCase struct {
	name              string
	moduleSource      string
	moduleRegistryVer string
	expectRegistryEnv bool
	expectedAPIPath   string
}

var _ = Describe("InitTerraformPromise source integration", func() {
	cwd, err := os.Getwd()
	Expect(err).NotTo(HaveOccurred())

	localModulePath := filepath.Join(cwd, "assets", "terraform", "modules", "local", "basic")
	vpcAPI := filepath.Join(cwd, "assets", "terraform", "api", "git.yaml")
	vpcSubdirAPI := filepath.Join(cwd, "assets", "terraform", "api", "git-subdir.yaml")
	s3API := filepath.Join(cwd, "assets", "terraform", "api", "registry.yaml")
	nestedAPI := filepath.Join(cwd, "assets", "terraform", "api", "nested-registry.yaml")
	localAPI := filepath.Join(cwd, "assets", "terraform", "api", "local.yaml")

	DescribeTable("generates promise schema and workflow envs",
		func(tc moduleTestCase) {
			workingDir, err := os.MkdirTemp("", "kratix-test-sources")
			Expect(err).NotTo(HaveOccurred())
			defer os.RemoveAll(workingDir)

			r := &runner{
				exitCode: 0,
				Path:     "/opt/homebrew/bin:" + os.Getenv("PATH"),
				flags: map[string]string{
					"--group":   "example.com",
					"--kind":    "Example",
					"--version": "v1alpha1",
					"--dir":     workingDir,
					"--split":   "",
				},
			}

			r.flags["--module-source"] = tc.moduleSource
			if tc.moduleRegistryVer != "" {
				r.flags["--module-registry-version"] = tc.moduleRegistryVer
			}

			session := r.run("init", "tf-module-promise", "example")
			Expect(session).To(gexec.Exit(0))
			Expect(session.Out).To(gbytes.Say("Promise generated successfully"))

			envs := readWorkflowEnvs(workingDir)
			Expect(envs["MODULE_SOURCE"]).To(Equal(tc.moduleSource))
			if tc.expectRegistryEnv {
				Expect(envs).To(HaveKeyWithValue("MODULE_REGISTRY_VERSION", tc.moduleRegistryVer))
			} else {
				Expect(envs).NotTo(HaveKey("MODULE_REGISTRY_VERSION"))
			}

			actual := readSpecTypes(workingDir)
			expected := readCRDTypes(tc.expectedAPIPath)
			Expect(actual).To(Equal(expected))

			actualCRD := readCRD(workingDir)
			expectedCRD := readCRDFromPath(tc.expectedAPIPath)
			Expect(actualCRD).To(Equal(expectedCRD))
		},
		Entry("open source git repo with ref",
			moduleTestCase{
				name:            "git vpc",
				moduleSource:    "git::https://github.com/terraform-aws-modules/terraform-aws-vpc.git?ref=v5.7.0",
				expectedAPIPath: vpcAPI,
			},
		),
		Entry("private git repo placeholder with subdir (mono-repo style)",
			moduleTestCase{
				name:            "git subdir",
				moduleSource:    "git::ssh://git@github.com/syntasso/kratix-cli-private-tf-module-test-fixture.git//modules/vpc-endpoints?ref=v5.7.0",
				expectedAPIPath: vpcSubdirAPI,
			},
		),
		Entry("registry with version",
			moduleTestCase{
				name:              "registry s3 bucket",
				moduleSource:      "terraform-aws-modules/s3-bucket/aws",
				moduleRegistryVer: "4.1.2",
				expectRegistryEnv: true,
				expectedAPIPath:   s3API,
			},
		),
		Entry("nested registry with version",
			moduleTestCase{
				name:              "nested registry vpc endpoints",
				moduleSource:      "terraform-aws-modules/vpc/aws//modules/vpc-endpoints",
				moduleRegistryVer: "5.7.0",
				expectRegistryEnv: true,
				expectedAPIPath:   nestedAPI,
			},
		),
		Entry("local filesystem module",
			moduleTestCase{
				name:            "local module",
				moduleSource:    localModulePath,
				expectedAPIPath: localAPI,
			},
		),
	)
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

func readSpecTypes(workingDir string) map[string]string {
	crd := readCRD(workingDir)
	specProps := crd.Spec.Versions[0].Schema.OpenAPIV3Schema.Properties["spec"].Properties

	result := map[string]string{}
	for name, prop := range specProps {
		result[name] = prop.Type
	}
	return result
}

func readCRDTypes(fixturePath string) map[string]string {
	expected := readCRDFromPath(fixturePath)
	specProps := expected.Spec.Versions[0].Schema.OpenAPIV3Schema.Properties["spec"].Properties

	result := map[string]string{}
	for name, prop := range specProps {
		result[name] = prop.Type
	}
	return result
}

func readCRD(workingDir string) apiextensionsv1.CustomResourceDefinition {
	path := filepath.Join(workingDir, "api.yaml")
	if _, err := os.Stat(path); err != nil {
		path = filepath.Join(workingDir, "promise.yaml")
	}
	return readCRDFromPath(path)
}

func readCRDFromPath(path string) apiextensionsv1.CustomResourceDefinition {
	data, err := os.ReadFile(path)
	Expect(err).NotTo(HaveOccurred())
	crd := apiextensionsv1.CustomResourceDefinition{}
	Expect(yaml.Unmarshal(data, &crd)).To(Succeed())
	return crd
}
