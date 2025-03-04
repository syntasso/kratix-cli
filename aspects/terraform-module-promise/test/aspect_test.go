package run_test

import (
	"fmt"
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
      "source": "example.com",
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

var _ = Describe("From Operator to Promise Aspect", func() {
	var (
		envVars map[string]string
		tmpDir  string
	)

	BeforeEach(func() {
		var err error
		tmpDir, err = os.MkdirTemp("", "kratix")
		Expect(err).NotTo(HaveOccurred())
		envVars = map[string]string{
			"KRATIX_INPUT_FILE": "assets/test-object.yaml",
			"KRATIX_OUTPUT_DIR": tmpDir,
			"MODULE_SOURCE":     "example.com",
		}

		fmt.Println("tmpDir: ", tmpDir)
	})

	AfterEach(func() {
		os.RemoveAll(tmpDir)
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
})
