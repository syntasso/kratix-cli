package run_test

import (
	"os"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func runStackWithEnv(envVars map[string]string) *gexec.Session {
	cmd := exec.Command(stackBinaryPath)
	for key, value := range envVars {
		cmd.Env = append(cmd.Env, key+"="+value)
	}

	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	EventuallyWithOffset(1, session, "10s").Should(gexec.Exit())
	return session
}

var _ = Describe("From request to Pulumi Stack stage", func() {
	var (
		envVars map[string]string
		tmpDir  string
	)

	BeforeEach(func() {
		var err error
		tmpDir, err = os.MkdirTemp("", "pulumi-generator")
		Expect(err).NotTo(HaveOccurred())

		envVars = map[string]string{
			"KRATIX_INPUT_FILE":      "assets/test-object.yaml",
			"KRATIX_OUTPUT_FILE":     filepath.Join(tmpDir, "output.yaml"),
			"PULUMI_COMPONENT_TOKEN": "pkg:index:Database",
		}
	})

	AfterEach(func() {
		Expect(os.RemoveAll(tmpDir)).To(Succeed())
	})

	It("creates a Stack CR in the output", func() {
		session := runStackWithEnv(envVars)
		Expect(session).To(gexec.Exit(0))
		Expect(session.Err).To(gbytes.Say("starting transformation"))
		Expect(session.Err).To(gbytes.Say("wrote Stack"))

		outputBytes, err := os.ReadFile(envVars["KRATIX_OUTPUT_FILE"])
		Expect(err).NotTo(HaveOccurred())

		stackObject := &unstructured.Unstructured{}
		Expect(yaml.Unmarshal(outputBytes, stackObject)).To(Succeed())

		Expect(stackObject.GetAPIVersion()).To(Equal("pulumi.com/v1"))
		Expect(stackObject.GetKind()).To(Equal("Stack"))
		Expect(stackObject.GetName()).To(MatchRegexp("test-object-[0-9a-f]{8}-stack"))
		Expect(stackObject.GetNamespace()).To(Equal("non-default"))
		Expect(stackObject.GetLabels()).To(HaveKeyWithValue("app.kubernetes.io/name", "test-object"))
		Expect(stackObject.GetAnnotations()).To(HaveKeyWithValue("image-registry", "ghcr.io"))

		programRefName, found, err := unstructured.NestedString(stackObject.Object, "spec", "programRef", "name")
		Expect(err).NotTo(HaveOccurred())
		Expect(found).To(BeTrue())
		Expect(programRefName).To(MatchRegexp("test-object-[0-9a-f]{8}"))

		stackName, found, err := unstructured.NestedString(stackObject.Object, "spec", "stack")
		Expect(err).NotTo(HaveOccurred())
		Expect(found).To(BeTrue())
		Expect(stackName).To(MatchRegexp("^test-object-[0-9a-f]{8}-stack$"))

		_, found, err = unstructured.NestedMap(stackObject.Object, "spec", "envRefs")
		Expect(err).NotTo(HaveOccurred())
		Expect(found).To(BeFalse())
	})

	It("adds Pulumi Cloud envRefs when the secret env vars are set", func() {
		envVars["PULUMI_STACK_ACCESS_TOKEN_SECRET_NAME"] = "pulumi-api-secret"
		envVars["PULUMI_STACK_ACCESS_TOKEN_SECRET_KEY"] = "accessToken"

		session := runStackWithEnv(envVars)
		Expect(session).To(gexec.Exit(0))

		outputBytes, err := os.ReadFile(envVars["KRATIX_OUTPUT_FILE"])
		Expect(err).NotTo(HaveOccurred())

		stackObject := &unstructured.Unstructured{}
		Expect(yaml.Unmarshal(outputBytes, stackObject)).To(Succeed())

		envRefs, found, err := unstructured.NestedMap(stackObject.Object, "spec", "envRefs")
		Expect(err).NotTo(HaveOccurred())
		Expect(found).To(BeTrue())
		Expect(envRefs).To(HaveKey("PULUMI_ACCESS_TOKEN"))

		accessTokenRef, ok := envRefs["PULUMI_ACCESS_TOKEN"].(map[string]any)
		Expect(ok).To(BeTrue())
		Expect(accessTokenRef).To(HaveKeyWithValue("type", "Secret"))
		Expect(accessTokenRef).To(HaveKeyWithValue("secret", map[string]any{
			"name": "pulumi-api-secret",
			"key":  "accessToken",
		}))
	})

	It("tries to read from /kratix/input/object.yaml if KRATIX_INPUT_FILE is not set", func() {
		delete(envVars, "KRATIX_INPUT_FILE")
		session := runStackWithEnv(envVars)

		Expect(session).To(gexec.Exit(1))
		Expect(session.Err).To(gbytes.Say("failed to read object file from /kratix/input/object.yaml"))
	})

	It("tries to write to /kratix/output/stack.yaml if KRATIX_OUTPUT_FILE is not set", func() {
		delete(envVars, "KRATIX_OUTPUT_FILE")
		session := runStackWithEnv(envVars)

		Expect(session).To(gexec.Exit(1))
		Expect(session.Err).To(gbytes.Say("failed to write object file to /kratix/output/stack.yaml"))
	})

	It("fails if the Pulumi component token env var is not set", func() {
		delete(envVars, "PULUMI_COMPONENT_TOKEN")
		session := runStackWithEnv(envVars)

		Expect(session).To(gexec.Exit(1))
		Expect(session.Err).To(gbytes.Say("missing required environment variable PULUMI_COMPONENT_TOKEN"))
	})

	It("fails when only one Stack access token secret env var is set", func() {
		envVars["PULUMI_STACK_ACCESS_TOKEN_SECRET_NAME"] = "pulumi-api-secret"
		session := runStackWithEnv(envVars)

		Expect(session).To(gexec.Exit(1))
		Expect(session.Err).To(gbytes.Say("invalid Pulumi Stack access token secret configuration"))
	})

})
