package integration_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"os"
)

var _ = Describe("update", func() {
	var workingDir string
	var r *runner

	BeforeEach(func() {
		var err error
		workingDir, err = os.MkdirTemp("", "kratix-test")
		Expect(err).NotTo(HaveOccurred())
		r = &runner{exitCode: 0, dir: workingDir}
	})
	AfterEach(func() {
		os.RemoveAll(workingDir)
	})

	When("called without a subcommand", func() {
		It("prints the help", func() {
			session := r.run("update")
			Expect(session.Out).To(SatisfyAll(
				gbytes.Say("Command to update kratix resources"),
				gbytes.Say(`Use "kratix update \[command\] --help" for more information about a command.`),
			))
		})
	})

	Context("api", func() {
		When("called with --help", func() {
			It("prints the help", func() {
				session := r.run("update", "api", "--help")
				Expect(session.Out).To(gbytes.Say("Command to update promise API"))
			})
		})

		When("updating promise api", func() {
			var dir string
			AfterEach(func() {
				os.RemoveAll(dir)
			})

			BeforeEach(func() {
				var err error
				dir, err = os.MkdirTemp("", "kratix-update-api-test")
				Expect(err).NotTo(HaveOccurred())

				sess := r.run("init", "promise", "postgresql", "--group", "syntasso.io", "--kind", "Database", "--dir", dir)
				Expect(sess.Out).To(gbytes.Say("postgresql promise bootstrapped in"))
			})

			Context("api GVK", func() {
				It("updates", func() {
					sess := r.run("update", "api", "--kind", "NewKind", "--group", "newGroup", "--version", "v1beta4", "--plural", "newPlural", "--dir", dir)
					Expect(sess.Out).To(gbytes.Say("Promise updated"))
					matchPromise(dir, "postgresql", "newGroup", "v1beta4", "NewKind", "newkind", "newPlural")
				})
			})

			Context("api properties", func() {
				It("can add new properties to the promise api", func() {
					sess := r.run("update", "api", "-p", "numberField:number", "--property", "stringField:string", "--dir", dir)
					Expect(sess.Out).To(gbytes.Say("Promise updated"))
					matchPromise(dir, "postgresql", "syntasso.io", "v1alpha1", "Database", "database", "databases")
					props := getCRDProperties(dir)
					Expect(props).To(SatisfyAll(HaveKey("numberField"), HaveKey("stringField")))
					Expect(props["numberField"].Type).To(Equal("number"))
					Expect(props["stringField"].Type).To(Equal("string"))
				})

				It("can update existing properties types", func() {
					r.run("update", "api", "-p", "numberField:number", "--property", "stringField:string", "-p", "wontchange:string", "--dir", dir)
					r.run("update", "api", "-p", "numberField:string", "--property", "stringField:number", "--dir", dir)
					matchPromise(dir, "postgresql", "syntasso.io", "v1alpha1", "Database", "database", "databases")
					props := getCRDProperties(dir)
					Expect(props).To(SatisfyAll(HaveKey("numberField"), HaveKey("stringField"), HaveKey("wontchange")))
					Expect(props["numberField"].Type).To(Equal("string"))
					Expect(props["wontchange"].Type).To(Equal("string"))
					Expect(props["stringField"].Type).To(Equal("number"))
				})

				It("errors when unsupported property type is set", func() {
					r.exitCode = 1
					sess := r.run("update", "api", "--property", "unsupported:object", "--dir", dir)
					Expect(sess.Err).To(gbytes.Say("unsupported"))
				})

				It("can remove existing properties", func() {
					r.run("update", "api", "-p", "numberField:number", "--property", "stringField:string", "-p", "wontdelete:string", "--dir", dir)
					r.run("update", "api", "-p", "numberField-", "--property", "stringField-", "--dir", dir)
					matchPromise(dir, "postgresql", "syntasso.io", "v1alpha1", "Database", "database", "databases")
					props := getCRDProperties(dir)
					Expect(props).To(SatisfyAll(HaveKey("wontdelete"), HaveLen(1)))
					Expect(props["wontdelete"].Type).To(Equal("string"))
				})

				It("errors when property format is invalid", func() {
					r.exitCode = 1
					sess := r.run("update", "api", "--property", "invalid%", "--dir", dir)
					Expect(sess.Err).To(gbytes.Say("invalid"))

					r.exitCode = 1
					sess = r.run("update", "api", "--property", "invalid+string", "--dir", dir)
					Expect(sess.Err).To(gbytes.Say("invalid"))
				})
			})
		})

	})
})
