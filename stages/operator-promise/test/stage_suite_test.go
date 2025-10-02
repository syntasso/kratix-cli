package run_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

func TestTest(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Operator Promise Stage Test Suite")
}

var binaryPath string

var _ = BeforeSuite(func() {
	var err error
	binaryPath, err = gexec.Build("github.com/syntasso/kratix-cli-plugin-investigation/stages/operator-promise")
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})
