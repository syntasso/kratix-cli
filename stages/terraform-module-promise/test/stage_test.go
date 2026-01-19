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

var expectedOutput = `{
  "module": {
    "testobject_non-default_test-object": {
      "source": "git::example.com?ref=1.0.0",
      "strArr": [
        {
          "field": "value"
        }
      ],
      "intArr": [
        1
      ],
      "listBool": [
        true
      ],
      "field": "value",
      "mapWithinMap": {
        "entryMap": {
          "entry": "value",
          "entry2": 2,
          "entry3": false
        },
        "entry": "value",
        "entry2": 2,
        "entry3": false
      },
      "number": 7
    }
  }
}`

var expectedOutputNoSpec = `{
  "module": {
    "testobject_non-default_test-object": {
      "source": "git::example.com?ref=1.0.0"
    }
  }
}`

var expectedRegistryOutput = `{
  "module": {
    "testobject_non-default_test-object": {
      "source": "terraform-aws-modules/iam/aws",
      "version": "6.2.3",
      "strArr": [
        {
          "field": "value"
        }
      ],
      "intArr": [
        1
      ],
      "listBool": [
        true
      ],
      "field": "value",
      "mapWithinMap": {
        "entryMap": {
          "entry": "value",
          "entry2": 2,
          "entry3": false
        },
        "entry": "value",
        "entry2": 2,
        "entry3": false
      },
      "number": 7
    }
  }
}`

func runWithEnv(envVars map[string]string) *gexec.Session {
	cmd := exec.Command(binaryPath)
	for key, value := range envVars {
		cmd.Env = append(cmd.Env, key+"="+value)
	}

	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	return session
}

var _ = Describe("From TF module to Promise Stage", func() {
	var (
		envVars map[string]string
		tmpDir  string
	)

	AfterEach(func() {
		os.RemoveAll(tmpDir)
	})

	Describe("Resource Configure Workflow", func() {
		BeforeEach(func() {
			var err error
			tmpDir, err = os.MkdirTemp("", "kratix")
			Expect(err).NotTo(HaveOccurred())
			envVars = map[string]string{
				"KRATIX_INPUT_FILE":    "assets/test-object.yaml",
				"KRATIX_OUTPUT_DIR":    tmpDir,
				"MODULE_SOURCE":        "git::example.com?ref=1.0.0",
				"KRATIX_WORKFLOW_TYPE": "resource",
			}
		})

		It("creates an object file in the output directory", func() {
			session := runWithEnv(envVars)
			Eventually(session).Should(gexec.Exit())
			Expect(session.Buffer()).To(gbytes.Say("Terraform JSON configuration written to %s/testobject_non-default_test-object.tf.json", tmpDir))
			Expect(session).To(gexec.Exit(0))
			output, err := os.ReadFile(filepath.Join(tmpDir, "testobject_non-default_test-object.tf.json"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(output)).To(MatchJSON(expectedOutput))
		})

		It("handles objects without a spec field", func() {
			envVars["KRATIX_INPUT_FILE"] = "assets/test-object-no-spec.yaml"
			session := runWithEnv(envVars)
			Eventually(session).Should(gexec.Exit())
			Expect(session.Buffer()).To(gbytes.Say("Terraform JSON configuration written to %s/testobject_non-default_test-object.tf.json", tmpDir))
			Expect(session).To(gexec.Exit(0))
			output, err := os.ReadFile(filepath.Join(tmpDir, "testobject_non-default_test-object.tf.json"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(output)).To(MatchJSON(expectedOutputNoSpec))
		})

		It("adds a registry version when provided separately", func() {
			envVars["MODULE_SOURCE"] = "terraform-aws-modules/iam/aws"
			envVars["MODULE_REGISTRY_VERSION"] = "6.2.3"
			session := runWithEnv(envVars)
			Eventually(session).Should(gexec.Exit())
			Expect(session.Buffer()).To(gbytes.Say("Terraform JSON configuration written to %s/testobject_non-default_test-object.tf.json", tmpDir))
			Expect(session).To(gexec.Exit(0))
			output, err := os.ReadFile(filepath.Join(tmpDir, "testobject_non-default_test-object.tf.json"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(output)).To(MatchJSON(expectedRegistryOutput))
		})

		It("errors when registry version is used with a non-registry source", func() {
			envVars["MODULE_REGISTRY_VERSION"] = "9.9.9"
			session := runWithEnv(envVars)
			Eventually(session).Should(gexec.Exit())
			Expect(session.ExitCode()).NotTo(Equal(0))
			Expect(session.Err).To(gbytes.Say("MODULE_REGISTRY_VERSION is only valid for Terraform registry sources"))
		})
	})

})
