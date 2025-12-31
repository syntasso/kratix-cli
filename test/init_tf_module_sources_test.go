package integration_test

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/syntasso/kratix-cli/internal"
	"sigs.k8s.io/yaml"
)

type moduleTestCase struct {
	name              string
	moduleSource      string
	moduleRegistryVer string
	expectRegistryEnv bool
	expectedTypesFile string
}

var _ = Describe("InitTerraformPromise source integration", func() {
	cwd, err := os.Getwd()
	Expect(err).NotTo(HaveOccurred())

	localModulePath := filepath.Join(cwd, "assets", "terraform", "modules", "local", "basic")
	vpcFixture := filepath.Join(cwd, "assets", "terraform", "vars", "vpc-variables.tf")
	s3Fixture := filepath.Join(cwd, "assets", "terraform", "vars", "registry-variables.tf")
	nestedFixture := filepath.Join(cwd, "assets", "terraform", "vars", "nested-registry-variables.tf")

	DescribeTable("generates promise schema and workflow envs",
		func(tc moduleTestCase) {
			workingDir, err := os.MkdirTemp("", "kratix-test-sources")
			Expect(err).NotTo(HaveOccurred())
			defer os.RemoveAll(workingDir)

			r := &runner{
				exitCode: 0,
				Path:     os.Getenv("PATH"),
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
			expected := expectedTypesFromFixture(tc.expectedTypesFile)
			Expect(actual).To(Equal(expected))
		},
		Entry("open source git repo with ref",
			moduleTestCase{
				name:              "git vpc",
				moduleSource:      "git::https://github.com/terraform-aws-modules/terraform-aws-vpc.git?ref=v5.7.0",
				expectedTypesFile: vpcFixture,
			},
		),
		Entry("git repo subdir (mono-repo style)",
			moduleTestCase{
				name:              "git subdir",
				moduleSource:      "git::https://github.com/terraform-aws-modules/terraform-aws-vpc.git//modules/vpc-endpoints?ref=v5.7.0",
				expectedTypesFile: nestedFixture,
			},
		),
		Entry("registry with version",
			moduleTestCase{
				name:              "registry s3 bucket",
				moduleSource:      "terraform-aws-modules/s3-bucket/aws",
				moduleRegistryVer: "4.1.2",
				expectRegistryEnv: true,
				expectedTypesFile: s3Fixture,
			},
		),
		Entry("nested registry with version",
			moduleTestCase{
				name:              "nested registry vpc endpoints",
				moduleSource:      "terraform-aws-modules/vpc/aws//modules/vpc-endpoints",
				moduleRegistryVer: "5.7.0",
				expectRegistryEnv: true,
				expectedTypesFile: nestedFixture,
			},
		),
		Entry("registry without version",
			moduleTestCase{
				name:              "registry without version",
				moduleSource:      "terraform-aws-modules/vpc/aws",
				expectedTypesFile: vpcFixture,
			},
		),
		Entry("local filesystem module",
			moduleTestCase{
				name:              "local module",
				moduleSource:      localModulePath,
				expectedTypesFile: filepath.Join(localModulePath, "variables.tf"),
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
	apiPath := filepath.Join(workingDir, "api.yaml")
	contents, err := os.ReadFile(apiPath)
	if err != nil {
		contents, err = os.ReadFile(filepath.Join(workingDir, "promise.yaml"))
		Expect(err).NotTo(HaveOccurred())
	}

	var doc map[string]any
	Expect(yaml.Unmarshal(contents, &doc)).To(Succeed())

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

	result := map[string]string{}
	for name, raw := range propMap {
		rawMap, _ := raw.(map[string]any)
		typ, _ := rawMap["type"].(string)
		result[name] = typ
	}

	return result
}

func expectedTypesFromFixture(fixturePath string) map[string]string {
	vars, err := internal.GetVariablesFromModule(fixturePath, "")
	Expect(err).NotTo(HaveOccurred())
	schema, warnings := internal.VariablesToCRDSpecSchema(vars)
	Expect(warnings).To(BeEmpty())

	result := map[string]string{}
	for name, prop := range schema.Properties {
		result[name] = prop.Type
	}
	return result
}
