package integration_test

import (
	"io"
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/syntasso/kratix-cli/cmd"
	mock_fetcher "github.com/syntasso/kratix-cli/test/mocks"
	"github.com/syntasso/kratix/test/kubeutils"
	"go.uber.org/mock/gomock"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var _ = Describe("kratix platform get resources", func() {
	var r *runner
	var workingDir string
	var dir string
	var mockFetcher *mock_fetcher.MockFetcher

	BeforeEach(func() {
		var err error
		workingDir, err = os.MkdirTemp("", "kratix-test")
		Expect(err).NotTo(HaveOccurred())

		dir, err = os.MkdirTemp("", "kratix-dir")
		Expect(err).NotTo(HaveOccurred())
		r = &runner{exitCode: 0, dir: workingDir}

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

	Describe("renderTree", func() {
		BeforeEach(func() {
			ctrl := gomock.NewController(GinkgoT())
			mockFetcher = mock_fetcher.NewMockFetcher(ctrl)

		})

		When("there are no resource requests", func() {
			It("details that there are no requests", func() {
				gvr := schema.GroupVersionResource{
					Group:    "marketplace.kratix.io",
					Version:  "v1alpha",
					Resource: "app",
				}
				mockFetcher.EXPECT().GVRForPromise(gomock.Any(), "app").Return(
					&schema.GroupVersionResource{
						Group:    "marketplace.kratix.io",
						Version:  "v1alpha",
						Resource: "app",
					}, nil,
				)

				mockFetcher.EXPECT().GetRequests(gomock.Any(), &gvr, "app", "").Return(
					&unstructured.UnstructuredList{}, nil,
				)

				r, w, _ := os.Pipe()
				originalStdout := os.Stdout
				os.Stdout = w

				err := cmd.RenderTree("app", mockFetcher)
				Expect(err).ToNot(HaveOccurred())

				w.Close()
				os.Stdout = originalStdout
				output, _ := io.ReadAll(r)
				r.Close()

				Expect(string(output)).To(Equal("No requests found for promise \"app\"\n"))
			})
		})

		When("a there are resource requests", func() {
			It("details the tree of requests and sub-requests", func() {
				gvr := schema.GroupVersionResource{
					Group:    "marketplace.kratix.io",
					Version:  "v1alpha",
					Resource: "app",
				}
				mockFetcher.EXPECT().GVRForPromise(gomock.Any(), "app").Return(
					&schema.GroupVersionResource{
						Group:    "marketplace.kratix.io",
						Version:  "v1alpha",
						Resource: "app",
					}, nil,
				)

				promiseRequest1 := unstructured.Unstructured{}
				promiseRequest1.SetName("my-app-1")

				promiseRequest2 := unstructured.Unstructured{}
				promiseRequest2.SetName("my-app-2")

				promiseRequest3 := unstructured.Unstructured{}
				promiseRequest3.SetName("my-app-3")

				mockFetcher.EXPECT().GetRequests(gomock.Any(), &gvr, "app", "").Return(
					&unstructured.UnstructuredList{Items: []unstructured.Unstructured{promiseRequest1, promiseRequest2, promiseRequest3}}, nil,
				)
				mockFetcher.EXPECT().GetKratixGVRs(gomock.Any()).Return(
					[]schema.GroupVersionResource{}, nil)
				r, w, _ := os.Pipe()
				originalStdout := os.Stdout
				os.Stdout = w

				err := cmd.RenderTree("app", mockFetcher)
				Expect(err).ToNot(HaveOccurred())

				w.Close()
				os.Stdout = originalStdout
				output, _ := io.ReadAll(r)
				r.Close()

				Expect(string(output)).To(SatisfyAll(
					ContainSubstring("  - my-app-1\n"),
					ContainSubstring("  - my-app-2\n"),
					ContainSubstring("  - my-app-3\n"),
				))
			})
		})

		When("a compound promise has are resource requests", func() {
			It("details the tree of requests and sub-requests", func() {
				gvr := schema.GroupVersionResource{
					Group:    "marketplace.kratix.io",
					Version:  "v1alpha",
					Resource: "app",
				}
				mockFetcher.EXPECT().GVRForPromise(gomock.Any(), "app").Return(
					&schema.GroupVersionResource{
						Group:    "marketplace.kratix.io",
						Version:  "v1alpha",
						Resource: "app",
					}, nil,
				)

				promiseRequest := unstructured.Unstructured{}
				promiseRequest.SetName("my-app")

				mockFetcher.EXPECT().GetRequests(gomock.Any(), &gvr, "app", "").Return(
					&unstructured.UnstructuredList{Items: []unstructured.Unstructured{promiseRequest}}, nil,
				)

				componentRequestGVR := schema.GroupVersionResource{
					Group:    "marketplace.kratix.io",
					Version:  "v1alpha",
					Resource: "configmap",
				}

				mockFetcher.EXPECT().GetKratixGVRs(gomock.Any()).Return(
					[]schema.GroupVersionResource{componentRequestGVR}, nil)

				componentRequest1 := unstructured.Unstructured{}
				componentRequest1.SetName("my-configmap")

				componentRequest2 := unstructured.Unstructured{}
				componentRequest2.SetName("my-service")

				mockFetcher.EXPECT().
					GetRequests(gomock.Any(), &componentRequestGVR, componentRequestGVR.Resource, gomock.Any()).
					Return(&unstructured.UnstructuredList{Items: []unstructured.Unstructured{componentRequest1, componentRequest2}}, nil)

				r, w, _ := os.Pipe()
				originalStdout := os.Stdout
				os.Stdout = w

				err := cmd.RenderTree("app", mockFetcher)
				Expect(err).ToNot(HaveOccurred())

				w.Close()
				os.Stdout = originalStdout
				output, _ := io.ReadAll(r)
				r.Close()

				Expect(string(output)).To(SatisfyAll(
					ContainSubstring("  - my-app\n"),
					ContainSubstring("    |\n"),
					ContainSubstring("    |--my-configmap\n"),
					ContainSubstring("    |\n"),
					ContainSubstring("    |--my-service\n"),
				))
			})
		})
	})
})
