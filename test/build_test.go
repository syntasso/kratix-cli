package integration_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/syntasso/kratix/api/v1alpha1"
	"os"
	"path/filepath"
	"slices"

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

		When("dependencies.yaml file", func() {
			It("does not exist should skip adding dependencies", func() {
				Expect(os.RemoveAll(filepath.Join(promiseDir, "dependencies.yaml"))).To(Succeed())

				sess := r.run("build", "promise", "postgresql", "--dir", promiseDir)
				Expect(sess.Out.Contents()).ToNot(BeEmpty())
				var promise v1alpha1.Promise
				Expect(yaml.Unmarshal(sess.Out.Contents(), &promise)).To(Succeed())
				Expect(promise.Spec.Dependencies).To(BeNil())
			})

			It("is empty, build should add an empty dependencies key", func() {
				Expect(os.WriteFile(filepath.Join(promiseDir, "dependencies.yaml"), []byte(""), 0644)).To(Succeed())

				sess := r.run("build", "promise", "postgresql", "--dir", promiseDir)
				Expect(sess.Out.Contents()).ToNot(BeEmpty())
				var promise v1alpha1.Promise
				Expect(yaml.Unmarshal(sess.Out.Contents(), &promise)).To(Succeed())
				Expect(promise.Spec.Dependencies).To(HaveLen(0))
			})

			When("is not a valid dependencies list", func() {
				It("should error and not build the promise", func() {
					Expect(os.WriteFile(filepath.Join(promiseDir, "dependencies.yaml"), []byte("not valid"), 0644)).To(Succeed())

					r.exitCode = 1
					sess := r.run("build", "promise", "postgresql", "--dir", promiseDir)
					Expect(sess.Err).To(gbytes.Say("json: cannot unmarshal string into Go value of type v1alpha1.Dependencies"))
				})
			})
		})

		When("--output flag is provided", func() {
			It("outputs promise definition to provided filepath", func() {
				r.run("init", "promise", "postgresql", "--group", "syntasso.io", "--kind", "Database", "--split", "--dir", promiseDir)
				r.run("build", "promise", "postgresql", "--dir", promiseDir, "--output", filepath.Join(promiseDir, "promise.yaml"))
				matchPromise(promiseDir, "postgresql", "syntasso.io", "v1alpha1", "Database", "database", "databases")
			})
		})
	})
})
