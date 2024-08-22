package integration_test

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("kratix build container", func() {
	var r *runner
	var workingDir string
	var dir string

	BeforeEach(func() {
		var err error
		workingDir, err = os.MkdirTemp("", "kratix-test")
		Expect(err).NotTo(HaveOccurred())

		dir, err = os.MkdirTemp("", "kratix-dir")
		Expect(err).NotTo(HaveOccurred())
		r = &runner{exitCode: 0, dir: workingDir}
	})

	AfterEach(func() {
		Expect(os.RemoveAll(workingDir)).To(Succeed())
		os.RemoveAll(dir)
	})

	Describe("--help", func() {
		It("shows the help message", func() {
			sess := r.run("build", "container", "--help")
			Expect(sess.Out).To(SatisfyAll(
				gbytes.Say("Usage:"),
				gbytes.Say("kratix build container LIFECYCLE/ACTION/PIPELINE-NAME"),

				gbytes.Say("Examples:"),
				gbytes.Say("# Build a container"),
				gbytes.Say("kratix build container resource/configure/mypipeline --name mycontainer"),

				gbytes.Say("Flags:"),
				gbytes.Say("-n, --name string\\s+Name of the container to build"),
			))
		})
	})

	Describe("no-split mode", func() {
		BeforeEach(func() {
			r.run("init", "promise", "postgresql", "--group", "syntasso.io", "--kind", "Database", "--dir", dir)
			r.run("add", "container", "promise/configure/postgresql", "--image", "syntasso/postgres-resource:v1.0.0", "--dir", dir)
		})

		Describe("lifecycle/action/pipeline-name", func() {
			When("there's a single container for that pipeline", func() {
				It("should build the image", func() {
					session := r.run("build", "container", "promise/configure/postgresql", "--dir", dir)

					Expect(session).To(gbytes.Say("Building container with tag syntasso/postgres-resource:v1.0.0..."))
					Expect(session).To(gbytes.Say("fake-docker build --tag syntasso/postgres-resource:v1.0.0 %s/workflows/promise/configure/postgresql/syntasso-postgres-resource", dir))
				})

				When("the incorrect container name is provided via the --name flag", func() {
					It("should raise an error", func() {
						r.exitCode = 1
						session := r.run("build", "container", "promise/configure/postgresql", "--dir", dir, "--name", "my-container")
						Expect(session.Err).To(gbytes.Say("container my-container not found in pipeline"))
					})
				})
			})

			When("there's more than one container for that pipeline", func() {
				BeforeEach(func() {
					r.flags = map[string]string{
						"--name":  "second-container",
						"--image": "syntasso/second-container:v1.0.0",
						"--dir":   dir,
					}
					r.run("add", "container", "promise/configure/postgresql")
					delete(r.flags, "--image")
				})

				It("should build when the name is provided", func() {
					session := r.run("build", "container", "promise/configure/postgresql")
					Expect(session).To(gbytes.Say("Building container with tag syntasso/second-container:v1.0.0..."))
				})

				It("should fail when no name is provided", func() {
					delete(r.flags, "--name")
					r.exitCode = 1
					session := r.run("build", "container", "promise/configure/postgresql")
					Expect(session.Err).To(gbytes.Say("more than one container exists for this pipeline, please provide a name"))
				})
			})

			When("the docker cli is not installed", func() {
				It("should fail with a helpful message", func() {
					r.noPath = true
					r.exitCode = 1
					session := r.run("build", "container", "promise/configure/postgresql", "--dir", dir)
					Expect(session.Err).To(gbytes.Say("docker CLI not found in PATH"))
				})
			})
		})

		When("--all is set", func() {
			BeforeEach(func() {
				r.run("add", "container", "resource/configure/instance", "--image", "syntasso/postgres-instance:v1.0.0", "--dir", dir)
			})

			It("builds all containers for all pipelines", func() {
				session := r.run("build", "container", "--dir", dir, "--all")
				Expect(session).To(SatisfyAll(
					gbytes.Say("Building container with tag syntasso/postgres-resource:v1.0.0..."),
					gbytes.Say("fake-docker build --tag syntasso/postgres-resource:v1.0.0 %s/workflows/promise/configure/postgresql/syntasso-postgres-resource", dir),
					gbytes.Say("Building container with tag syntasso/postgres-instance:v1.0.0..."),
					gbytes.Say("fake-docker build --tag syntasso/postgres-instance:v1.0.0 %s/workflows/resource/configure/instance/syntasso-postgres-instance", dir),
				))
			})
		})

		When("--engine is set", func() {
			Context("with a valid engine", func() {
				It("builds the container with the specified engine", func() {
					session := r.run("build", "container", "--dir", dir, "promise/configure/postgresql", "--engine", "podman")
					Expect(session).To(SatisfyAll(
						gbytes.Say("Building container with tag syntasso/postgres-resource:v1.0.0..."),
						gbytes.Say("fake-podman build --tag syntasso/postgres-resource:v1.0.0 %s/workflows/promise/configure/postgresql/syntasso-postgres-resource", dir),
					))
				})
			})
			Context("with a unsupported engine", func() {
				It("errors", func() {
					r.exitCode = 1
					session := r.run("build", "container", "--dir", dir, "promise/configure/postgresql", "--engine", "rancher")
					Expect(session.Err).To(SatisfyAll(
						gbytes.Say("unsupported container engine: rancher"),
					))
				})
			})
		})

		When("--buildx is set", func() {
			It("uses the buildx cli command", func() {
				session := r.run("build", "container", "--dir", dir, "promise/configure/postgresql", "--buildx")
				Expect(session).To(SatisfyAll(
					gbytes.Say("Building container with tag syntasso/postgres-resource:v1.0.0..."),
					gbytes.Say("fake-docker buildx build --tag syntasso/postgres-resource:v1.0.0 %s/workflows/promise/configure/postgresql/syntasso-postgres-resource", dir),
				))
			})

			When("--build-args is also provided", func() {
				It("assembles the right command", func() {
					session := r.run("build", "container", "--dir", dir, "promise/configure/postgresql", "--buildx", "--build-args", "--platform linux/amd64 --builder mybuilder")
					Expect(session).To(SatisfyAll(
						gbytes.Say("Building container with tag syntasso/postgres-resource:v1.0.0..."),
						gbytes.Say("fake-docker buildx build --tag syntasso/postgres-resource:v1.0.0 %s/workflows/promise/configure/postgresql/syntasso-postgres-resource --platform linux/amd64 --builder mybuilder", dir),
					))
				})
			})

			When("--push is also provided", func() {
				It("builds and pushes the image on a single command", func() {
					session := r.run("build", "container", "--dir", dir, "promise/configure/postgresql", "--buildx", "--push")
					Expect(session).To(SatisfyAll(
						gbytes.Say("Building container with tag syntasso/postgres-resource:v1.0.0..."),
						gbytes.Say("fake-docker buildx build --tag syntasso/postgres-resource:v1.0.0 %s/workflows/promise/configure/postgresql/syntasso-postgres-resource --push", dir),
						Not(gbytes.Say("fake-docker push syntasso/postgres-resource:v1.0.0")),
					))
				})
			})
		})

		When("--build-args is set", func() {
			It("uses the additional arguments in the build command", func() {
				session := r.run("build", "container", "--dir", dir, "promise/configure/postgresql", "--build-args", "--platform linux/amd64 --builder mybuilder")
				Expect(session).To(SatisfyAll(
					gbytes.Say("Building container with tag syntasso/postgres-resource:v1.0.0..."),
					gbytes.Say("fake-docker build --tag syntasso/postgres-resource:v1.0.0 %s/workflows/promise/configure/postgresql/syntasso-postgres-resource --platform linux/amd64 --builder mybuilder", dir),
				))
			})
		})

		When("--push is set", func() {
			It("builds and pushes the image", func() {
				session := r.run("build", "container", "--dir", dir, "promise/configure/postgresql", "--push")
				Expect(session).To(SatisfyAll(
					gbytes.Say("Building container with tag syntasso/postgres-resource:v1.0.0..."),
					gbytes.Say("fake-docker build --tag syntasso/postgres-resource:v1.0.0 %s/workflows/promise/configure/postgresql/syntasso-postgres-resource", dir),

					gbytes.Say("Pushing container with tag syntasso/postgres-resource:v1.0.0..."),
					gbytes.Say("fake-docker push syntasso/postgres-resource:v1.0.0"),
				))
			})
		})
	})

	When("--split is used on init", func() {
		BeforeEach(func() {
			r.run("init", "promise", "postgresql", "--group", "syntasso.io", "--kind", "Database", "--dir", dir, "--split")
			r.run("add", "container", "promise/configure/postgresql", "--image", "syntasso/postgres-resource:v1.0.0", "--dir", dir)
		})

		It("should build the image", func() {
			session := r.run("build", "container", "promise/configure/postgresql", "--dir", dir)
			Expect(session).To(gbytes.Say("Building container with tag syntasso/postgres-resource:v1.0.0..."))
			Expect(session).To(gbytes.Say("fake-docker build --tag syntasso/postgres-resource:v1.0.0 %s/workflows/promise/configure/postgresql/syntasso-postgres-resource", dir))
		})

		When("no workflows exists", func() {
			It("should raise an error", func() {
				r.exitCode = 1
				session := r.run("build", "container", "promise/configure/postgresql", "--dir", ".")
				Expect(session.Err).To(gbytes.Say("no workflows found"))
			})
		})
	})
})
