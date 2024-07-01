package integration_test

import (
	"context"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/syntasso/kratix/api/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"

	"k8s.io/apimachinery/pkg/util/yaml"
)

var _ = Describe("build", func() {
	var r *runner
	var dir string

	BeforeEach(func() {
		var err error
		dir, err = os.MkdirTemp("", "kratix-build-test")
		Expect(err).NotTo(HaveOccurred())
		r = &runner{exitCode: 0}
	})

	AfterEach(func() {
		os.RemoveAll(dir)
	})

	Context("promise", func() {
		It("builds a promise from api and dependencies files", func() {
			r.run("init", "promise", "postgresql", "--group", "syntasso.io", "--kind", "Database", "--split", "--dir", dir)
			sess := r.run("build", "promise", "postgresql", "--dir", dir)
			Expect(sess.Out.Contents()).ToNot(BeEmpty())

			var promise v1alpha1.Promise
			Expect(yaml.Unmarshal(sess.Out.Contents(), &promise)).To(Succeed())
			Expect(promise.Name).To(Equal("postgresql"))
			Expect(promise.Kind).To(Equal("Promise"))
			Expect(promise.APIVersion).To(Equal(v1alpha1.GroupVersion.String()))

			promiseCRD, err := promise.GetAPIAsCRD()
			Expect(err).NotTo(HaveOccurred())
			matchCRD(promiseCRD, "syntasso.io", "v1alpha1", "Database", "database", "databases")
		})

		When("workflow files exist in the workflows directory", func() {
			It("builds a promise from api, dependencies and workflow files", func() {
				r.run("init", "promise", "postgresql", "--group", "syntasso.io", "--kind", "Database", "--split", "--dir", dir)
				r.run("add", "container", "promise/configure/pipeline0", "--image", "psql:latest", "-n", "configure-image", "--dir", dir)
				r.run("add", "container", "resource/delete/pipeline0", "--image", "psql:latest", "-n", "delete-image", "--dir", dir)
				sess := r.run("build", "promise", "postgresql", "--dir", dir)

				var promise v1alpha1.Promise
				Expect(yaml.Unmarshal(sess.Out.Contents(), &promise)).To(Succeed())
				Expect(promise.Name).To(Equal("postgresql"))
				Expect(promise.Kind).To(Equal("Promise"))
				Expect(promise.APIVersion).To(Equal(v1alpha1.GroupVersion.String()))
				pipelines, err := promise.GeneratePipelines(ctrl.LoggerFrom(context.Background()))
				Expect(err).ToNot(HaveOccurred())
				Expect(pipelines.ConfigurePromise).To(HaveLen(1))
				Expect(pipelines.DeletePromise).To(HaveLen(0))
				Expect(pipelines.ConfigureResource).To(HaveLen(0))
				Expect(pipelines.DeleteResource).To(HaveLen(1))
				Expect(pipelines.ConfigurePromise[0].Name).To(Equal("pipeline0"))
				Expect(pipelines.ConfigurePromise[0].Spec.Containers).To(HaveLen(1))
				Expect(pipelines.ConfigurePromise[0].Spec.Containers[0].Name).To(Equal("configure-image"))
				Expect(pipelines.ConfigurePromise[0].Spec.Containers[0].Image).To(Equal("psql:latest"))
				Expect(pipelines.DeleteResource[0].Spec.Containers[0].Name).To(Equal("delete-image"))
				Expect(pipelines.DeleteResource[0].Spec.Containers[0].Image).To(Equal("psql:latest"))

				promiseCRD, err := promise.GetAPIAsCRD()
				Expect(err).NotTo(HaveOccurred())
				matchCRD(promiseCRD, "syntasso.io", "v1alpha1", "Database", "database", "databases")
			})
		})

		When("--output flag is provided", func() {
			It("outputs promise definition to provided filepath", func() {
				r.run("init", "promise", "postgresql", "--group", "syntasso.io", "--kind", "Database", "--split", "--dir", dir)
				r.run("build", "promise", "postgresql", "--dir", dir, "--output", filepath.Join(dir, "promise.yaml"))
				matchPromise(dir, "postgresql", "syntasso.io", "v1alpha1", "Database", "database", "databases")
			})
		})
	})
})
