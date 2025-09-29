package integration_test

import (
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/syntasso/kratix/test/kubeutils"
)

var _ = Describe("kratix platform get resources", func() {
	var r *runner
	var workingDir string
	var dir string
	var platform kubeutils.Cluster

	BeforeEach(func() {
		var err error
		workingDir, err = os.MkdirTemp("", "kratix-test")
		Expect(err).NotTo(HaveOccurred())

		dir, err = os.MkdirTemp("", "kratix-dir")
		Expect(err).NotTo(HaveOccurred())
		r = &runner{exitCode: 0, dir: workingDir}

		platform = kubeutils.Cluster{
			Context: "kind-platform",
			Name:    "platform-cluster",
		}

		kubeutils.SetTimeoutAndInterval(1*time.Minute, 2*time.Second)
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

	When("the request is a compound requests", func() {
		BeforeEach(func() {
			platform.Kubectl("apply", "--filename", "assets/compound-labelled-promise/configmap-promise.yaml")
			platform.Kubectl("apply", "--filename", "assets/compound-labelled-promise/service-promise.yaml")
			platform.Kubectl("apply", "--filename", "assets/compound-labelled-promise/promise.yaml")
		})

		AfterEach(func() {
			platform.Kubectl("delete", "--filename", "assets/compound-labelled-promise/configmap-promise.yaml")
			platform.Kubectl("delete", "--filename", "assets/compound-labelled-promise/service-promise.yaml")
			platform.Kubectl("delete", "--filename", "assets/compound-labelled-promise/promise.yaml")
		})

		When("there are no resource requests", func() {
			It("details that there are no requests", func() {
				sess := r.run("platform", "get", "resources", "app")
				Expect(sess.Out).To(gbytes.Say("No requests found for promise \"app\""))
			})
		})

		When("there are resource requests", func() {
			BeforeEach(func() {
				platform.Kubectl("apply", "--filename", "assets/compound-labelled-promise/resource-request.yaml")
				Eventually(func() string {
					return platform.Kubectl("get", "app", "my-app")
				}, 2*time.Minute).Should(ContainSubstring("Reconciled"))
			})

			It("details the tree of requests and sub-requests", func() {
				sess := r.run("platform", "get", "resources", "app")
				Expect(sess.Buffer()).To(SatisfyAll(
					gbytes.Say(`- my-app`),
					gbytes.Say(`|--my-app-configmap`),
					gbytes.Say(`|--my-app-service`),
				))
			})
		})
	})

	When("the request is a not a compound requests", func() {
		BeforeEach(func() {
			platform.Kubectl("apply", "--filename", "assets/compound-labelled-promise/configmap-promise.yaml")
			platform.Kubectl("apply", "--filename", "assets/compound-labelled-promise/configmap-request-1.yaml")
			platform.Kubectl("apply", "--filename", "assets/compound-labelled-promise/configmap-request-2.yaml")
		})

		It("displays the requests in a list", func() {
			sess := r.run("platform", "get", "resources", "configmappromise")
			Expect(sess.Buffer()).To(SatisfyAll(
				gbytes.Say(`- configmap-1`),
				gbytes.Say(`- configmap-2`),
			))
		})
	})
})
