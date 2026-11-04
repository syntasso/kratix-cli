package integration_test

import (
	"os/exec"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// hack/check-stage-images asks the registry whether each stage image can be
// pulled. That part needs the network, so it is not tested here. What is tested
// is that the script still finds every stage image the CLI references: if a
// constant is renamed or reshaped and the script silently stops seeing it, the
// guard would pass while checking nothing.
var _ = Describe("hack/check-stage-images --list", func() {
	It("finds the image for every stage", func() {
		script, err := filepath.Abs("../hack/check-stage-images")
		Expect(err).NotTo(HaveOccurred())

		output, err := exec.Command(script, "--list").Output()
		Expect(err).NotTo(HaveOccurred())

		references := strings.Fields(string(output))
		for _, repository := range []string{
			"from-api-to-crossplane-claim",
			"from-api-to-operator",
			"helm-resource-configure",
			"pulumi-generator",
			"terraform-generate",
		} {
			matching := 0
			for _, reference := range references {
				if strings.HasPrefix(reference, "ghcr.io/syntasso/kratix-cli/"+repository+":v") {
					matching++
				}
			}
			Expect(matching).To(Equal(1), "expected exactly one reference to %s in %v", repository, references)
		}
	})
})
