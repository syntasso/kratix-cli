package integration_test

import (
	"context"
	"os"
	"path/filepath"
	"slices"

	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/syntasso/kratix/api/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"

	"k8s.io/apimachinery/pkg/util/yaml"
)

var _ = Describe("build", func() {
	var r *runner
	var promiseDir, depDir string

	BeforeEach(func() {
		var err error
		promiseDir, err = os.MkdirTemp("", "kratix-build-test")
		Expect(err).NotTo(HaveOccurred())

		depDir, err = os.MkdirTemp("", "dep")
		Expect(err).NotTo(HaveOccurred())

		r = &runner{exitCode: 0}
		r.run("init", "promise", "postgresql", "--group", "syntasso.io", "--kind", "Database", "--split", "--dir", promiseDir)

		Expect(os.WriteFile(filepath.Join(depDir, "deps.yaml"), slices.Concat(
			namespaceBytes(namespace("test1")),
			deploymentBytes(deployment("test1"))), 0644)).To(Succeed())
		r.run("update", "dependencies", depDir, "--dir", promiseDir)
	})

	AfterEach(func() {
		Expect(os.RemoveAll(promiseDir)).To(Succeed())
		Expect(os.RemoveAll(depDir)).To(Succeed())
	})

	Context("promise", func() {
		It("builds a promise from api, dependencies and workflows files", func() {
			sess := r.run("build", "promise", "postgresql", "--dir", promiseDir)
			Expect(sess.Out.Contents()).ToNot(BeEmpty())

			var promise v1alpha1.Promise
			Expect(yaml.Unmarshal(sess.Out.Contents(), &promise)).To(Succeed())
			Expect(promise.Name).To(Equal("postgresql"))
			Expect(promise.Kind).To(Equal("Promise"))
			Expect(promise.APIVersion).To(Equal(v1alpha1.GroupVersion.String()))

			promiseCRD, err := promise.GetAPIAsCRD()
			Expect(err).NotTo(HaveOccurred())
			matchCRD(promiseCRD, "syntasso.io", "v1alpha1", "Database", "database", "databases")

			promiseDependencies := promise.Spec.Dependencies
			Expect(promiseDependencies).To(HaveLen(2))
			Expect(promiseDependencies[0].Object["apiVersion"]).To(Equal("v1"))
			Expect(promiseDependencies[0].Object["kind"]).To(Equal("Namespace"))
			Expect(promiseDependencies[1].Object["apiVersion"]).To(Equal("apps/v1"))
			Expect(promiseDependencies[1].Object["kind"]).To(Equal("Deployment"))
		})

		When("workflow files exist in the workflows directory", func() {
			It("builds a promise from api, dependencies and workflow files", func() {
				r.run("init", "promise", "postgresql", "--group", "syntasso.io", "--kind", "Database", "--split", "--dir", promiseDir)
				r.run("add", "container", "promise/configure/pipeline0", "--image", "psql:latest", "-n", "configure-image", "--dir", promiseDir)
				r.run("add", "container", "resource/delete/pipeline0", "--image", "psql:latest", "-n", "delete-image", "--dir", promiseDir)
				sess := r.run("build", "promise", "postgresql", "--dir", promiseDir)

				var promise v1alpha1.Promise
				Expect(yaml.Unmarshal(sess.Out.Contents(), &promise)).To(Succeed())
				Expect(promise.Name).To(Equal("postgresql"))
				Expect(promise.Kind).To(Equal("Promise"))
				Expect(promise.APIVersion).To(Equal(v1alpha1.GroupVersion.String()))
				pipelines, err := v1alpha1.NewPipelinesMap(&promise, ctrl.LoggerFrom(context.Background()))
				Expect(err).ToNot(HaveOccurred())
				Expect(pipelines[v1alpha1.WorkflowTypePromise][v1alpha1.WorkflowActionConfigure]).To(HaveLen(1))
				Expect(pipelines[v1alpha1.WorkflowTypePromise][v1alpha1.WorkflowActionDelete]).To(HaveLen(0))
				Expect(pipelines[v1alpha1.WorkflowTypeResource][v1alpha1.WorkflowActionConfigure]).To(HaveLen(0))
				Expect(pipelines[v1alpha1.WorkflowTypeResource][v1alpha1.WorkflowActionDelete]).To(HaveLen(1))
				Expect(pipelines[v1alpha1.WorkflowTypePromise][v1alpha1.WorkflowActionConfigure][0].Name).To(Equal("pipeline0"))
				Expect(pipelines[v1alpha1.WorkflowTypePromise][v1alpha1.WorkflowActionConfigure][0].Spec.Containers).To(HaveLen(1))
				Expect(pipelines[v1alpha1.WorkflowTypePromise][v1alpha1.WorkflowActionConfigure][0].Spec.Containers[0].Name).To(Equal("configure-image"))
				Expect(pipelines[v1alpha1.WorkflowTypePromise][v1alpha1.WorkflowActionConfigure][0].Spec.Containers[0].Image).To(Equal("psql:latest"))
				Expect(pipelines[v1alpha1.WorkflowTypeResource][v1alpha1.WorkflowActionDelete][0].Spec.Containers[0].Name).To(Equal("delete-image"))
				Expect(pipelines[v1alpha1.WorkflowTypeResource][v1alpha1.WorkflowActionDelete][0].Spec.Containers[0].Image).To(Equal("psql:latest"))

				promiseCRD, err := promise.GetAPIAsCRD()
				Expect(err).NotTo(HaveOccurred())
				matchCRD(promiseCRD, "syntasso.io", "v1alpha1", "Database", "database", "databases")
			})
		})

		Context("dependencies.yaml file", func() {
			When("does not exist", func() {
				It("skips adding dependencies", func() {
					Expect(os.RemoveAll(filepath.Join(promiseDir, "dependencies.yaml"))).To(Succeed())

					sess := r.run("build", "promise", "postgresql", "--dir", promiseDir)
					Expect(sess.Out.Contents()).ToNot(BeEmpty())
					var promise v1alpha1.Promise
					Expect(yaml.Unmarshal(sess.Out.Contents(), &promise)).To(Succeed())
					Expect(promise.Spec.Dependencies).To(BeNil())
				})
			})

			When("is empty", func() {
				It("skips adding dependencies", func() {
					Expect(os.WriteFile(filepath.Join(promiseDir, "dependencies.yaml"), []byte(""), 0644)).To(Succeed())

					sess := r.run("build", "promise", "postgresql", "--dir", promiseDir)
					Expect(sess.Out.Contents()).ToNot(BeEmpty())
					var promise v1alpha1.Promise
					Expect(yaml.Unmarshal(sess.Out.Contents(), &promise)).To(Succeed())
					Expect(promise.Spec.Dependencies).To(BeNil())
				})
			})
		})

		Context("api.yaml file", func() {
			When("does not exist", func() {
				It("skips adding the API", func() {
					Expect(os.RemoveAll(filepath.Join(promiseDir, "api.yaml"))).To(Succeed())

					sess := r.run("build", "promise", "postgresql", "--dir", promiseDir)
					Expect(sess.Out.Contents()).ToNot(BeEmpty())
					var promise v1alpha1.Promise
					Expect(yaml.Unmarshal(sess.Out.Contents(), &promise)).To(Succeed())
					Expect(promise.Spec.API).To(BeNil())
				})
			})
			When("is empty", func() {
				It("skips adding the API", func() {
					Expect(os.WriteFile(filepath.Join(promiseDir, "api.yaml"), []byte(""), 0644)).To(Succeed())

					sess := r.run("build", "promise", "postgresql", "--dir", promiseDir)
					Expect(sess.Out.Contents()).ToNot(BeEmpty())
					var promise v1alpha1.Promise
					Expect(yaml.Unmarshal(sess.Out.Contents(), &promise)).To(Succeed())
					Expect(promise.Spec.API).To(BeNil())
				})
			})
		})

		DescribeTable("split file is not valid", func(fileName, objType string) {
			Expect(os.WriteFile(filepath.Join(promiseDir, fileName), []byte("not valid"), 0644)).To(Succeed())

			r.exitCode = 1
			sess := r.run("build", "promise", "postgresql", "--dir", promiseDir)
			Expect(sess.Err).To(gbytes.Say(fmt.Sprintf("json: cannot unmarshal string into Go value of type %s", objType)))
		},
			Entry("dependencies file", "dependencies.yaml", "v1alpha1.Dependencies"),
			Entry("api file", "api.yaml", "v1.CustomResourceDefinition"),
		)

		When("--output flag is provided", func() {
			It("outputs promise definition to provided filepath", func() {
				r.run("init", "promise", "postgresql", "--group", "syntasso.io", "--kind", "Database", "--split", "--dir", promiseDir)
				r.run("build", "promise", "postgresql", "--dir", promiseDir, "--output", filepath.Join(promiseDir, "promise.yaml"))
				matchPromise(promiseDir, "postgresql", "syntasso.io", "v1alpha1", "Database", "database", "databases")
			})
		})
	})
})
