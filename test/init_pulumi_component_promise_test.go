package integration_test

import (
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("init pulumi-component-promise", func() {
	var (
		r                   *runner
		workingDir          string
		singleSchemaPath    string
		multiSchemaPath     string
		malformedSchemaPath string
		onlyUnsupportedPath string
		mixedSchemaPath     string
	)

	BeforeEach(func() {
		var err error
		workingDir, err = os.MkdirTemp("", "kratix-test")
		Expect(err).NotTo(HaveOccurred())
		r = &runner{exitCode: 0, dir: workingDir, timeout: 10 * time.Second}

		singleSchemaPath = filepath.Join(workingDir, "single-component-schema.json")
		Expect(os.WriteFile(singleSchemaPath, []byte(`{"resources":{"pkg:index:Thing":{"isComponent":true,"inputProperties":{"name":{"type":"string"}},"requiredInputs":["name"]}}}`), 0o600)).To(Succeed())

		multiSchemaPath = filepath.Join(workingDir, "multi-component-schema.json")
		Expect(os.WriteFile(multiSchemaPath, []byte(`{"resources":{"pkg:index:Zulu":{"isComponent":true,"inputProperties":{"name":{"type":"string"}}},"pkg:index:Alpha":{"isComponent":true,"inputProperties":{"name":{"type":"string"}}}}}`), 0o600)).To(Succeed())

		malformedSchemaPath = filepath.Join(workingDir, "malformed-schema.json")
		Expect(os.WriteFile(malformedSchemaPath, []byte(`{"resources":`), 0o600)).To(Succeed())

		onlyUnsupportedPath = filepath.Join(workingDir, "unsupported-schema.json")
		Expect(os.WriteFile(onlyUnsupportedPath, []byte(`{"resources":{"pkg:index:Thing":{"isComponent":true,"inputProperties":{"value":{"oneOf":[{"type":"string"},{"type":"number"}]}}}}}`), 0o600)).To(Succeed())

		mixedSchemaPath = filepath.Join(workingDir, "mixed-schema.json")
		Expect(os.WriteFile(mixedSchemaPath, []byte(`{"resources":{"pkg:index:Thing":{"isComponent":true,"inputProperties":{"name":{"type":"string"},"value":{"oneOf":[{"type":"string"},{"type":"number"}]}}}}}`), 0o600)).To(Succeed())
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
			"--schema", singleSchemaPath,
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
			"--schema", singleSchemaPath,
			"--group", "syntasso.io",
			"--kind", "Database",
		)
		Expect(session.Err).To(gbytes.Say(`Error: accepts 1 arg\(s\), received 2`))
	})

	It("prints preview warning on valid invocation", func() {
		session := r.run(
			"init", "pulumi-component-promise", "mypromise",
			"--schema", singleSchemaPath,
			"--group", "syntasso.io",
			"--kind", "Database",
		)
		Expect(session.Out).To(gbytes.Say("Preview: This command is in preview"))
	})

	It("fails when schema has multiple components and --component is not provided", func() {
		session := withExitCode(1).run(
			"init", "pulumi-component-promise", "mypromise",
			"--schema", multiSchemaPath,
			"--group", "syntasso.io",
			"--kind", "Database",
		)
		Expect(session.Err).To(gbytes.Say(`Error: select component: multiple components found; provide --component from: pkg:index:Alpha, pkg:index:Zulu`))
	})

	It("fails when provided --component token is unknown", func() {
		session := withExitCode(1).run(
			"init", "pulumi-component-promise", "mypromise",
			"--schema", multiSchemaPath,
			"--component", "pkg:index:Missing",
			"--group", "syntasso.io",
			"--kind", "Database",
		)
		Expect(session.Err).To(gbytes.Say(`Error: select component: component "pkg:index:Missing" not found; available components: pkg:index:Alpha, pkg:index:Zulu`))
	})

	It("succeeds when --component selects from a multi-component schema", func() {
		session := r.run(
			"init", "pulumi-component-promise", "mypromise",
			"--schema", multiSchemaPath,
			"--component", "pkg:index:Alpha",
			"--group", "syntasso.io",
			"--kind", "Database",
		)
		Expect(session.Out).To(gbytes.Say("Preview: This command is in preview"))
		Expect(session.Err).NotTo(gbytes.Say("Error:"))
	})

	It("prints deterministic warnings for skipped unsupported paths", func() {
		session := r.run(
			"init", "pulumi-component-promise", "mypromise",
			"--schema", mixedSchemaPath,
			"--group", "syntasso.io",
			"--kind", "Database",
		)
		Expect(session.Out).To(gbytes.Say(`warning: skipped unsupported schema path "spec.value" for component "pkg:index:Thing": keyword "oneOf"`))
		Expect(session.Err).NotTo(gbytes.Say("Error:"))
	})

	It("fails when translated spec is empty after skipping unsupported paths", func() {
		session := withExitCode(1).run(
			"init", "pulumi-component-promise", "mypromise",
			"--schema", onlyUnsupportedPath,
			"--group", "syntasso.io",
			"--kind", "Database",
		)
		Expect(session.Err).To(gbytes.Say(`Error: translate component inputs: no translatable spec fields remain after skipping unsupported schema paths for component "pkg:index:Thing"`))
	})

	It("fails for malformed JSON schema", func() {
		session := withExitCode(1).run(
			"init", "pulumi-component-promise", "mypromise",
			"--schema", malformedSchemaPath,
			"--group", "syntasso.io",
			"--kind", "Database",
		)
		Expect(session.Err).To(gbytes.Say(`Error: load schema: parse input schema as JSON:`))
	})

	It("fails for unsupported schema URL scheme", func() {
		session := withExitCode(1).run(
			"init", "pulumi-component-promise", "mypromise",
			"--schema", "ftp://example.com/schema.json",
			"--group", "syntasso.io",
			"--kind", "Database",
		)
		Expect(session.Err).To(gbytes.Say(`Error: load schema: unsupported URL scheme "ftp"`))
	})

})
