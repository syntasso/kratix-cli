package integration_test

import (
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("init pulumi-component-promise", func() {
	var (
		r          *runner
		workingDir string
	)

	BeforeEach(func() {
		var err error
		workingDir, err = os.MkdirTemp("", "kratix-test")
		Expect(err).NotTo(HaveOccurred())
		r = &runner{exitCode: 0, dir: workingDir, timeout: 10 * time.Second}
	})

	AfterEach(func() {
		Expect(os.RemoveAll(workingDir)).To(Succeed())
	})

	It("is discoverable under init help", func() {
		session := r.run("init", "--help")
		Expect(session.Out).To(gbytes.Say("pulumi-component-promise"))
	})

	It("prints preview help text and examples", func() {
		session := r.run("init", "pulumi-component-promise", "--help")
		Expect(session.Out).To(SatisfyAll(
			gbytes.Say("Preview: Initialize a new Promise from a Pulumi package schema"),
			gbytes.Say("kratix init pulumi-component-promise mypromise --schema ./schema.json"),
			gbytes.Say("kratix init pulumi-component-promise mypromise --schema https://www.pulumi.com/registry/packages/aws/schema.json"),
		))
		Expect(string(session.Out.Contents())).NotTo(ContainSubstring("Preview: This command is in preview"))
	})

	It("fails when promise name is missing", func() {
		session := withExitCode(1).run(
			"init", "pulumi-component-promise",
			"--schema", "./schema.json",
			"--group", "syntasso.io",
			"--kind", "Database",
		)
		Expect(session.Err).To(gbytes.Say(`Error: accepts 1 arg\(s\), received 0`))
	})

	It("fails when schema is missing", func() {
		session := withExitCode(1).run(
			"init", "pulumi-component-promise", "mypromise",
			"--group", "syntasso.io",
			"--kind", "Database",
		)
		Expect(session.Err).To(gbytes.Say(`Error: required flag\(s\) "schema" not set`))
	})

	It("fails when extra positional args are provided", func() {
		session := withExitCode(1).run(
			"init", "pulumi-component-promise", "mypromise", "extra-arg",
			"--schema", "./schema.json",
			"--group", "syntasso.io",
			"--kind", "Database",
		)
		Expect(session.Err).To(gbytes.Say(`Error: accepts 1 arg\(s\), received 2`))
	})

	It("prints preview warning on valid invocation", func() {
		session := r.run(
			"init", "pulumi-component-promise", "mypromise",
			"--schema", "./schema.json",
			"--group", "syntasso.io",
			"--kind", "Database",
		)
		Expect(session.Out).To(gbytes.Say("Preview: This command is in preview"))
	})
})
