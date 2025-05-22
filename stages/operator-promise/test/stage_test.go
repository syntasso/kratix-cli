package run_test

import (
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var expectedOutput = `apiVersion: example.com/v1
kind: Example
metadata:
  annotations:
    image-registry: ghcr.io
  labels:
    app.kubernetes.io/name: test-object
    keyy: value
  name: test-object
  namespace: default
spec:
  arr:
  - field: value
  field: value
  nested:
    field: value
  number: 7`

func runWithEnv(envVars map[string]string) *gexec.Session {
	cmd := exec.Command(binaryPath)
	for key, value := range envVars {
		cmd.Env = append(cmd.Env, key+"="+value)
	}

	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	Eventually(session).Should(gexec.Exit())
	return session
}

var _ = Describe("From Operator to Promise Stage", func() {
	var (
		envVars map[string]string
	)

	BeforeEach(func() {
		envVars = map[string]string{
			"KRATIX_INPUT_FILE":  "assets/test-object.yaml",
			"KRATIX_OUTPUT_FILE": "/dev/stdout",
			"OPERATOR_GROUP":     "example.com",
			"OPERATOR_VERSION":   "v1",
			"OPERATOR_KIND":      "Example",
		}
	})

	It("creates an object file in the output directory", func() {
		session := runWithEnv(envVars)
		Expect(session.Out).To(gbytes.Say(expectedOutput))
	})

	It("tries to read from /kratix/input/object.yaml if KRATIX_INPUT_FILE is not set", func() {
		delete(envVars, "KRATIX_INPUT_FILE")
		session := runWithEnv(envVars)

		Expect(session).To(gexec.Exit(1))
		Expect(session.Err).To(gbytes.Say("Failed to read object file from /kratix/input/object.yaml"))
	})

	It("tries to write to from /kratix/output/object.yaml if KRATIX_OUTPUT_FILE is not set", func() {
		delete(envVars, "KRATIX_OUTPUT_FILE")
		session := runWithEnv(envVars)

		Expect(session).To(gexec.Exit(1))
		Expect(session.Err).To(gbytes.Say("Failed to write object file to /kratix/output/object.yaml"))
	})

	DescribeTable("the required env vars", func(envVar string) {
		delete(envVars, envVar)
		session := runWithEnv(envVars)

		Expect(session).To(gexec.Exit(1))
		Expect(session.Err).To(gbytes.Say("Expected %s to be set", envVar))
	},
		Entry("fails if operator group is not set", "OPERATOR_GROUP"),
		Entry("fails if operator version is not set", "OPERATOR_VERSION"),
		Entry("fails if operator kind is not set", "OPERATOR_KIND"),
	)
})
