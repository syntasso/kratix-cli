package integration_test

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("plugin", func() {
	var workingDir string
	var r *runner

	BeforeEach(func() {
		var err error
		workingDir, err = os.MkdirTemp("", "kratix-test")
		Expect(err).NotTo(HaveOccurred())
		r = &runner{exitCode: 0, dir: workingDir}
	})

	AfterEach(func() {
		Expect(os.RemoveAll(workingDir)).To(Succeed())
	})

	When("called without a subcommand", func() {
		It("prints the help", func() {
			session := r.run("plugin")
			Expect(session.Out).To(SatisfyAll(
				gbytes.Say("Provides utilities for interacting with plugins"),
				gbytes.Say("list        List all visible plugin executables on a user's PATH"),
			))
		})
	})

	Context("list", func() {
		When("there is no plugin found in PATH", func() {
			It("errors and prints a message", func() {
				r.Path = "does-not-contain-a-plugin"
				r.exitCode = 1
				session := r.run("plugin", "list")
				Expect(session.Err).To(gbytes.Say("error: unable to find any kratix plugins in your PATH"))
			})
		})

		When("there are plugins in PATH", func() {
			BeforeEach(func() {
				createPlugins(workingDir)
				r = &runner{exitCode: 0, dir: workingDir, Path: workingDir}
			})
			It("lists them", func() {
				sess := r.run("plugin", "list")
				Expect(sess.Out).To(SatisfyAll(
					gbytes.Say("The following compatible plugins are available:"),
					gbytes.Say("kratix-cat"),
					gbytes.Say("kratix-error"),
					gbytes.Say("kratix-hi"),
				))
			})
		})
	})

	Context("execute", func() {
		When("there are plugins in PATH", func() {
			BeforeEach(func() {
				createPlugins(workingDir)
				r = &runner{exitCode: 0, dir: workingDir, Path: workingDir}
			})

			It("can execute them", func() {
				By("running plugins successfully")
				sess := r.run("hi")
				Expect(sess.Out).To(gbytes.Say("have a nice day"))

				sess = r.run("cat")
				Expect(sess.Out).To(gbytes.Say("meow meow meow"))

				By("preserving the exit code when plugin failed")
				r.exitCode = 127
				sess = r.run("error")
				Expect(sess.Out).To(gbytes.Say("blah"))
			})
		})
	})
})

func createPlugins(dir string) {
	script1 := "#!/bin/sh\n" +
		"echo \"meow meow meow\"\n"
	script2 := "#!/bin/sh\n" +
		"echo \"have a nice day\"\n"
	script3 := "#!/bin/sh\n" +
		"echo \"blah\" && exit 127\n"

	writePluginFile(filepath.Join(dir, "kratix-cat"), script1)
	writePluginFile(filepath.Join(dir, "kratix-hi"), script2)
	writePluginFile(filepath.Join(dir, "kratix-error"), script3)
}

func writePluginFile(path, script string) {
	var f *os.File
	var err error
	if f, err = os.Create(path); err != nil {
		ExpectWithOffset(1, err).NotTo(HaveOccurred())
	}

	defer f.Close()

	_, err = f.WriteString(script)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	ExpectWithOffset(1, f.Sync()).To(Succeed())
	ExpectWithOffset(1, f.Close()).To(Succeed())

	// rwxr-xr-x
	ExpectWithOffset(1, os.Chmod(f.Name(), 0o755)).To(Succeed())
}
