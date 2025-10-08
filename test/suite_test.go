package integration_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var binaryPath string

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Test Suite")
}

var _ = BeforeSuite(func() {
	var err error
	binaryPath, err = gexec.Build("github.com/syntasso/kratix-cli/cmd/kratix")
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})

type runner struct {
	exitCode int
	dir      string
	flags    map[string]string
	timeout  time.Duration
	noPath   bool
}

func withExitCode(exitCode int) *runner {
	return &runner{exitCode: exitCode}
}

func (r *runner) run(args ...string) *gexec.Session {
	for key, value := range r.flags {
		if value == "" {
			args = append(args, key)
		} else {
			args = append(args, key, value)
		}
	}
	cmd := exec.Command(binaryPath, args...)
	cmd.Dir = r.dir
	cmd.Env = os.Environ()

	testBin, err := filepath.Abs("assets/binaries")
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	cmdPath := testBin + ":" + os.Getenv("PATH")
	if r.noPath {
		cmdPath = ""
	}
	cmd.Env = append(cmd.Env, "PATH="+cmdPath)

	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	t := 20 * time.Second
	if r.timeout > 0 {
		t = r.timeout
	}
	EventuallyWithOffset(1, session).WithTimeout(t).Should(gexec.Exit(r.exitCode))
	return session
}
