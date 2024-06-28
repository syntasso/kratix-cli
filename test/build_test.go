package integration_test

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/syntasso/kratix/api/v1alpha1"

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
		It("builds a promise from api, dependencies and workflows files", func() {
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

		When("--output flag is provided", func() {
			It("outputs promise definition to provided filepath", func() {
				r.run("init", "promise", "postgresql", "--group", "syntasso.io", "--kind", "Database", "--split", "--dir", dir)
				r.run("build", "promise", "postgresql", "--dir", dir, "--output", filepath.Join(dir, "promise.yaml"))
				matchPromise(dir, "postgresql", "syntasso.io", "v1alpha1", "Database", "database", "databases")
			})
		})
	})
})
