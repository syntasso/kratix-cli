package integration_test

import (
	"os"
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = FDescribe("kratix platform get resources", func() {
	var r *runner
	var workingDir string
	var dir string

	BeforeEach(func() {
		var err error
		workingDir, err = os.MkdirTemp("", "kratix-test")
		Expect(err).NotTo(HaveOccurred())

		dir, err = os.MkdirTemp("", "kratix-dir")
		Expect(err).NotTo(HaveOccurred())
		r = &runner{exitCode: 0, dir: workingDir}
	})

	AfterEach(func() {
		Expect(os.RemoveAll(workingDir)).To(Succeed())
		os.RemoveAll(dir)
	})

	Describe("--help", func() {
		It("shows the help message", func() {
			sess := r.run("platform", "get", "resources", "--help")
			Expect(sess.Out).To(SatisfyAll(
				gbytes.Say("Show requests for a Promise and for a Compound Promises, its sub-requests"),
				gbytes.Say("Usage:"),
				gbytes.Say("kratix platform get resources PROMISE-NAME"),
				gbytes.Say("Flags:"),
				gbytes.Say("--context string\\s+The name of the kubeconfig context to use"),
			))
		})
	})

	FWhen("the request is a compound requests", func() {
		BeforeEach(func() {
			cmd := exec.Command("kubectl", "apply", "--filename", "assets/compound-labelled-promise/configmap-promise.yaml")
			_, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())

			cmd = exec.Command("kubectl", "apply", "--filename", "assets/compound-labelled-promise/service-promise.yaml")
			_, err = gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())

			cmd = exec.Command("kubectl", "apply", "--filename", "assets/compound-labelled-promise/promise.yaml")
			_, err = gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())

			//ensure app is available
		})

		When("there are no resource requests", func() {
			It("details that there are no requests", func() {
				sess := r.run("platform", "get", "resources", "app")
				Expect(sess.Out).To(gbytes.Say("No requests found for promise \"app\""))
			})
		})

		When("there are resource requests", func() {
			BeforeEach(func() {
				cmd := exec.Command("kubectl", "apply", "--filename", "assets/compound-labelled-promise/resource-request.yaml")
				_, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())
			})
			It("details the tree of requests and sub-requests", func() {
				sess := r.run("platform", "get", "resources", "app")
				Eventually(func() *gbytes.Buffer {
					return sess.Out
				}, 30*time.Second).Should(gbytes.Say("lkjbl \"app\""))
			})
		})
	})

	When("the request is a not compound requests", func() {
		It("displays the requests in a list", func() {

		})
	})
})
