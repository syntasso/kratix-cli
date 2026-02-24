package run_test

import (
	"os"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

func runWithEnv(envVars map[string]string) *gexec.Session {
	cmd := exec.Command(binaryPath)
	for key, value := range envVars {
		cmd.Env = append(cmd.Env, key+"="+value)
	}

	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	EventuallyWithOffset(1, session, "10s").Should(gexec.Exit())
	return session
}

var _ = Describe("From request to Pulumi Program stage", func() {
	var (
		envVars map[string]string
		tmpDir  string
	)

	BeforeEach(func() {
		var err error
		tmpDir, err = os.MkdirTemp("", "pulumi-stage")
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

	It("creates a Program CR in the output", func() {
		session := runWithEnv(envVars)
		Expect(session).To(gexec.Exit(0))
		outputBytes, err := os.ReadFile(envVars["KRATIX_OUTPUT_FILE"])
		Expect(err).NotTo(HaveOccurred())
		output := string(outputBytes)

		Expect(output).To(MatchRegexp("apiVersion: pulumi.com/v1"))
		Expect(output).To(MatchRegexp("kind: Program"))
		Expect(output).To(MatchRegexp("name: test-object-[0-9a-f]{8}"))
		Expect(output).To(MatchRegexp("namespace: non-default"))
		Expect(output).To(MatchRegexp("labels:"))
		Expect(output).To(MatchRegexp("app.kubernetes.io/name: test-object"))
		Expect(output).To(MatchRegexp("annotations:"))
		Expect(output).To(MatchRegexp("image-registry: ghcr.io"))
		Expect(output).To(MatchRegexp("resources:"))
		Expect(output).To(MatchRegexp("pkg-index-database:"))
		Expect(output).To(MatchRegexp("type: pkg:index:Database"))
		Expect(output).To(MatchRegexp("properties:"))
		Expect(output).To(MatchRegexp("field: value"))
		Expect(output).To(MatchRegexp("number: 7"))
	})

	It("returns an explicit error when spec is missing", func() {
		envVars["KRATIX_INPUT_FILE"] = "assets/test-object-no-spec.yaml"
		session := runWithEnv(envVars)

		Expect(session).To(gexec.Exit(1))
		Expect(session.Err).To(gbytes.Say("missing required field: spec"))
	})

	It("returns an explicit error when spec is not an object", func() {
		envVars["KRATIX_INPUT_FILE"] = "assets/test-object-bad-spec.yaml"
		session := runWithEnv(envVars)

		Expect(session).To(gexec.Exit(1))
		Expect(session.Err).To(gbytes.Say("invalid field: spec must be an object"))
	})

	It("tries to read from /kratix/input/object.yaml if KRATIX_INPUT_FILE is not set", func() {
		delete(envVars, "KRATIX_INPUT_FILE")
		session := runWithEnv(envVars)

		Expect(session).To(gexec.Exit(1))
		Expect(session.Err).To(gbytes.Say("failed to read object file from /kratix/input/object.yaml"))
	})

	It("tries to write to /kratix/output/object.yaml if KRATIX_OUTPUT_FILE is not set", func() {
		delete(envVars, "KRATIX_OUTPUT_FILE")
		session := runWithEnv(envVars)

		Expect(session).To(gexec.Exit(1))
		Expect(session.Err).To(gbytes.Say("failed to write object file to /kratix/output/object.yaml"))
	})

	It("fails if the Pulumi component token env var is not set", func() {
		delete(envVars, "PULUMI_COMPONENT_TOKEN")
		session := runWithEnv(envVars)

		Expect(session).To(gexec.Exit(1))
		Expect(session.Err).To(gbytes.Say("expected PULUMI_COMPONENT_TOKEN to be set"))
	})
})
