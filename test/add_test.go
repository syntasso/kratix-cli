package integration_test

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/syntasso/kratix/api/v1alpha1"

	"github.com/go-logr/logr"

	"k8s.io/apimachinery/pkg/util/yaml"
)

var _ = Describe("add", func() {
	var r *runner
	var workingDir string

	BeforeEach(func() {
		var err error
		workingDir, err = os.MkdirTemp("", "kratix-test")
		Expect(err).NotTo(HaveOccurred())
		r = &runner{exitCode: 0, dir: workingDir}
	})

	AfterEach(func() {
		os.RemoveAll(workingDir)
	})

	When("it is called without a subcommand", func() {
		It("prints the help", func() {
			session := r.run("add", "--help")
			Expect(session.Out).To(gbytes.Say("Command to add to Kratix resources"))
		})
	})

	Context("container", func() {
		When("called without an argument", func() {
			It("fails and prints the help", func() {
				r.exitCode = 1
				session := r.run("add", "container", "--image", "animage:latest")
				Expect(session.Err).To(gbytes.Say("kratix add container LIFECYCLE/ACTION/PIPELINE-NAME"))
			})
		})

		When("called without --image", func() {
			It("fails with message", func() {
				r.exitCode = 1
				session := r.run("add", "container", "promise/delete/instance")
				Expect(session.Err).To(gbytes.Say(`required flag\(s\) \"image\" not set`))
			})
		})

		When("called without 3 parts to the pipeline input", func() {
			It("fails with message", func() {
				r.exitCode = 1
				session := r.run("add", "container", "promise/delete", "--image", "animage:latest")
				Expect(session.Err).To(gbytes.Say(`invalid pipeline format: promise/delete, expected format: LIFECYCLE/ACTION/PIPELINE-NAME`))
			})
		})

		When("called with an invalid LIFECYCLE", func() {
			It("fails with message", func() {
				r.exitCode = 1
				session := r.run("add", "container", "invalid/delete/instance", "--image", "animage:latest")
				Expect(session.Err).To(gbytes.Say(`invalid lifecycle: invalid, expected one of: promise, resource`))
			})
		})

		When("called with an invalid ACTION", func() {
			It("fails with message", func() {
				r.exitCode = 1
				session := r.run("add", "container", "promise/invalid/instance", "--image", "animage:latest")
				Expect(session.Err).To(gbytes.Say(`invalid action: invalid, expected one of: configure, delete`))
			})
		})

		When("called with an empty PIPELINE-NAME", func() {
			It("fails with message", func() {
				r.exitCode = 1
				session := r.run("add", "container", "promise/configure/", "--image", "animage:latest")
				Expect(session.Err).To(gbytes.Say(`pipeline name cannot be empty`))
			})
		})

		When("called with --help", func() {
			It("prints the help", func() {
				session := r.run("add", "container", "--help")
				Expect(session.Out).To(gbytes.Say("kratix add container LIFECYCLE/ACTION/PIPELINE-NAME"))
			})
		})

		When("adding a container", func() {
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

			It("adds containers to promise workflows", func() {
				sess := r.run("add", "container", "promise/configure/pipeline0", "--image", "image:latest", "--dir", dir)
				Expect(sess.Out).To(gbytes.Say(fmt.Sprintf("generated the promise/configure/pipeline0/image in %s/promise.yaml", dir)))
				r.run("add", "container", "promise/configure/pipeline1", "--image", "project/image1:latest", "-n", "a-good-container", "--dir", dir)
				r.run("add", "container", "promise/delete/pipeline0", "--image", "project/cleanup:latest", "--dir", dir)

				pipelines := getWorkflows(dir)
				Expect(pipelines[v1alpha1.WorkflowTypeResource][v1alpha1.WorkflowActionConfigure]).To(HaveLen(0))
				Expect(pipelines[v1alpha1.WorkflowTypeResource][v1alpha1.WorkflowActionDelete]).To(HaveLen(0))
				Expect(pipelines[v1alpha1.WorkflowTypePromise][v1alpha1.WorkflowActionConfigure]).To(HaveLen(2))
				Expect(pipelines[v1alpha1.WorkflowTypePromise][v1alpha1.WorkflowActionDelete]).To(HaveLen(1))

				Expect(pipelines[v1alpha1.WorkflowTypePromise][v1alpha1.WorkflowActionConfigure][0].Name).To(Equal("pipeline0"))
				Expect(pipelines[v1alpha1.WorkflowTypePromise][v1alpha1.WorkflowActionConfigure][0].Spec.Containers).To(HaveLen(1))
				Expect(pipelines[v1alpha1.WorkflowTypePromise][v1alpha1.WorkflowActionConfigure][0].Spec.Containers[0].Image).To(Equal("image:latest"))
				Expect(pipelines[v1alpha1.WorkflowTypePromise][v1alpha1.WorkflowActionConfigure][0].Spec.Containers[0].Name).To(Equal("image"))
				Expect(pipelines[v1alpha1.WorkflowTypePromise][v1alpha1.WorkflowActionConfigure][1].Name).To(Equal("pipeline1"))
				Expect(pipelines[v1alpha1.WorkflowTypePromise][v1alpha1.WorkflowActionConfigure][1].Spec.Containers).To(HaveLen(1))
				Expect(pipelines[v1alpha1.WorkflowTypePromise][v1alpha1.WorkflowActionConfigure][1].Spec.Containers[0].Image).To(Equal("project/image1:latest"))
				Expect(pipelines[v1alpha1.WorkflowTypePromise][v1alpha1.WorkflowActionConfigure][1].Spec.Containers[0].Name).To(Equal("a-good-container"))

				Expect(pipelines[v1alpha1.WorkflowTypePromise][v1alpha1.WorkflowActionDelete][0].Name).To(Equal("pipeline0"))
				Expect(pipelines[v1alpha1.WorkflowTypePromise][v1alpha1.WorkflowActionDelete][0].Spec.Containers).To(HaveLen(1))
				Expect(pipelines[v1alpha1.WorkflowTypePromise][v1alpha1.WorkflowActionDelete][0].Spec.Containers[0].Image).To(Equal("project/cleanup:latest"))
				Expect(pipelines[v1alpha1.WorkflowTypePromise][v1alpha1.WorkflowActionDelete][0].Spec.Containers[0].Name).To(Equal("project-cleanup"))

				Expect(sess.Out).To(gbytes.Say("Customise your container by editing workflows/promise/configure/pipeline0/image/scripts/pipeline.sh"))
				script := getPipelineScriptContents(dir, "promise", "configure", "pipeline1", "a-good-container")
				Expect(script).To(ContainSubstring("Hello from ${name} ${namespace}"))
				Expect(sess.Out).To(gbytes.Say("Don't forget to build and push your image!"))

				Expect(pipelineWorkflowPathExists(dir, "promise", "configure", "pipeline0", "image", "Dockerfile")).To(BeTrue())
				Expect(pipelineWorkflowPathExists(dir, "promise", "configure", "pipeline0", "image", "resources/")).To(BeTrue())
				Expect(pipelineWorkflowPathExists(dir, "promise", "configure", "pipeline1", "a-good-container", "Dockerfile")).To(BeTrue())
				Expect(pipelineWorkflowPathExists(dir, "promise", "configure", "pipeline1", "a-good-container", "resources/")).To(BeTrue())
				Expect(pipelineWorkflowPathExists(dir, "promise", "delete", "pipeline0", "project-cleanup", "Dockerfile")).To(BeTrue())
				Expect(pipelineWorkflowPathExists(dir, "promise", "delete", "pipeline0", "project-cleanup", "resources/")).To(BeTrue())
			})

			When("multiple containers are added to the same pipeline", func() {
				It("adds containers to promise workflows", func() {
					r.run("add", "container", "promise/configure/pipeline0", "--image", "project/image1:latest", "-n", "a-great-container", "--dir", dir)
					r.run("add", "container", "promise/configure/pipeline0", "--image", "project/image2:latest", "-n", "an-even-greater-container", "--dir", dir)

					pipelines := getWorkflows(dir)
					Expect(pipelines[v1alpha1.WorkflowTypeResource][v1alpha1.WorkflowActionConfigure]).To(HaveLen(0))
					Expect(pipelines[v1alpha1.WorkflowTypeResource][v1alpha1.WorkflowActionDelete]).To(HaveLen(0))
					Expect(pipelines[v1alpha1.WorkflowTypePromise][v1alpha1.WorkflowActionConfigure]).To(HaveLen(1))
					Expect(pipelines[v1alpha1.WorkflowTypePromise][v1alpha1.WorkflowActionDelete]).To(HaveLen(0))

					Expect(pipelines[v1alpha1.WorkflowTypePromise][v1alpha1.WorkflowActionConfigure][0].Name).To(Equal("pipeline0"))
					Expect(pipelines[v1alpha1.WorkflowTypePromise][v1alpha1.WorkflowActionConfigure][0].Spec.Containers).To(HaveLen(2))
					Expect(pipelines[v1alpha1.WorkflowTypePromise][v1alpha1.WorkflowActionConfigure][0].Spec.Containers[0].Name).To(Equal("a-great-container"))
					Expect(pipelines[v1alpha1.WorkflowTypePromise][v1alpha1.WorkflowActionConfigure][0].Spec.Containers[0].Image).To(Equal("project/image1:latest"))
					Expect(pipelines[v1alpha1.WorkflowTypePromise][v1alpha1.WorkflowActionConfigure][0].Spec.Containers[1].Name).To(Equal("an-even-greater-container"))
					Expect(pipelines[v1alpha1.WorkflowTypePromise][v1alpha1.WorkflowActionConfigure][0].Spec.Containers[1].Image).To(Equal("project/image2:latest"))

					Expect(pipelineWorkflowPathExists(dir, "promise", "configure", "pipeline0", "a-great-container", "Dockerfile")).To(BeTrue())
					Expect(pipelineWorkflowPathExists(dir, "promise", "configure", "pipeline0", "a-great-container", "resources/")).To(BeTrue())
					Expect(getPipelineScriptContents(dir, "promise", "configure", "pipeline0", "a-great-container")).To(ContainSubstring("Hello from ${name} ${namespace}"))

					Expect(pipelineWorkflowPathExists(dir, "promise", "configure", "pipeline0", "an-even-greater-container", "Dockerfile")).To(BeTrue())
					Expect(pipelineWorkflowPathExists(dir, "promise", "configure", "pipeline0", "an-even-greater-container", "resources/")).To(BeTrue())
					Expect(getPipelineScriptContents(dir, "promise", "configure", "pipeline0", "an-even-greater-container")).To(ContainSubstring("Hello from ${name} ${namespace}"))
				})
			})

			It("adds containers to resource workflows", func() {
				r.run("add", "container", "resource/configure/pipeline0", "--image", "project/image1:latest", "-n", "a-great-container", "--dir", dir)
				r.run("add", "container", "resource/configure/pipeline0", "--image", "project/image2:latest", "--dir", dir)
				r.run("add", "container", "resource/delete/pipeline0", "--image", "project/cleanup:latest", "--dir", dir)

				pipelines := getWorkflows(dir)
				Expect(pipelines[v1alpha1.WorkflowTypePromise][v1alpha1.WorkflowActionConfigure]).To(HaveLen(0))
				Expect(pipelines[v1alpha1.WorkflowTypePromise][v1alpha1.WorkflowActionDelete]).To(HaveLen(0))

				Expect(pipelines[v1alpha1.WorkflowTypeResource][v1alpha1.WorkflowActionConfigure]).To(HaveLen(1))
				Expect(pipelines[v1alpha1.WorkflowTypeResource][v1alpha1.WorkflowActionConfigure][0].Name).To(Equal("pipeline0"))
				Expect(pipelines[v1alpha1.WorkflowTypeResource][v1alpha1.WorkflowActionConfigure][0].Spec.Containers).To(HaveLen(2))
				Expect(pipelines[v1alpha1.WorkflowTypeResource][v1alpha1.WorkflowActionConfigure][0].Spec.Containers[0].Image).To(Equal("project/image1:latest"))
				Expect(pipelines[v1alpha1.WorkflowTypeResource][v1alpha1.WorkflowActionConfigure][0].Spec.Containers[0].Name).To(Equal("a-great-container"))
				Expect(pipelines[v1alpha1.WorkflowTypeResource][v1alpha1.WorkflowActionConfigure][0].Spec.Containers[1].Image).To(Equal("project/image2:latest"))
				Expect(pipelines[v1alpha1.WorkflowTypeResource][v1alpha1.WorkflowActionConfigure][0].Spec.Containers[1].Name).To(Equal("project-image2"))

				Expect(pipelines[v1alpha1.WorkflowTypeResource][v1alpha1.WorkflowActionDelete]).To(HaveLen(1))
				Expect(pipelines[v1alpha1.WorkflowTypeResource][v1alpha1.WorkflowActionDelete][0].Name).To(Equal("pipeline0"))
				Expect(pipelines[v1alpha1.WorkflowTypeResource][v1alpha1.WorkflowActionDelete][0].Spec.Containers).To(HaveLen(1))
				Expect(pipelines[v1alpha1.WorkflowTypeResource][v1alpha1.WorkflowActionDelete][0].Spec.Containers[0].Image).To(Equal("project/cleanup:latest"))
				Expect(pipelines[v1alpha1.WorkflowTypeResource][v1alpha1.WorkflowActionDelete][0].Spec.Containers[0].Name).To(Equal("project-cleanup"))
			})

			When("adding a container that matches the name of an existing container in the pipeline", func() {
				It("raises an error", func() {
					r.run("add", "container", "promise/configure/pipeline0", "--image", "image:latest", "--name", "my-image", "--dir", dir)
					sess := withExitCode(1).run("add", "container", "promise/configure/pipeline0", "--image", "another-image:latest", "--name", "my-image", "--dir", dir)
					Expect(sess.Err).To(gbytes.Say("image 'my-image' already exists in Pipeline"))
				})
			})

			When("the container image name is very long", func() {
				It("truncates the generated container name", func() {
					longImage := "ghcr.io/blahblah/idp-reference-architecture/database-configure-pipeline:v0.1.0"
					expectedName := generateExpectedContainerName(longImage)
					sess := r.run("add", "container", "promise/configure/pipeline0", "--image", longImage, "--dir", dir)
					Expect(sess.Out).To(gbytes.Say(fmt.Sprintf("generated the promise/configure/pipeline0/%s in %s/promise.yaml", expectedName, dir)))

					pipelines := getWorkflows(dir)
					Expect(pipelines[v1alpha1.WorkflowTypePromise][v1alpha1.WorkflowActionConfigure][0].Spec.Containers[0].Name).To(Equal(expectedName))
					Expect(len(pipelines[v1alpha1.WorkflowTypePromise][v1alpha1.WorkflowActionConfigure][0].Spec.Containers[0].Name)).To(BeNumerically("<=", 63))
				})

				It("generates unique names for similar long images", func() {
					longImageA := "ghcr.io/blahblah/idp-reference-architecture/database-configure-pipelineaa:v0.1.0"
					longImageB := "ghcr.io/blahblah/idp-reference-architecture/database-configure-pipelinebb:v0.1.0"

					r.run("add", "container", "promise/configure/pipeline0", "--image", longImageA, "--dir", dir)
					r.run("add", "container", "promise/configure/pipeline0", "--image", longImageB, "--dir", dir)

					pipelines := getWorkflows(dir)
					names := pipelines[v1alpha1.WorkflowTypePromise][v1alpha1.WorkflowActionConfigure][0].Spec.Containers
					Expect(names[0].Name).NotTo(Equal(names[1].Name))
					Expect(len(names[0].Name)).To(BeNumerically("<=", 63))
					Expect(len(names[1].Name)).To(BeNumerically("<=", 63))
				})
			})

			When("the --language flag is provided", func() {
				It("raises an error if the specified language is not supported", func() {
					r.exitCode = 1
					sess := r.run("add", "container", "promise/configure/pipeline0", "--image", "image:latest", "--dir", dir, "--language", "clojure")
					Expect(sess.Err).To(gbytes.Say("invalid language: clojure is not supported by the kratix cli"))
				})

				Context("bash", func() {
					It("generates the expected files and adds containers to promise workflows", func() {
						sess := r.run("add", "container", "promise/configure/pipeline0", "--image", "image:latest", "--dir", dir, "--language", "bash")

						pipelines := getWorkflows(dir)
						Expect(pipelines[v1alpha1.WorkflowTypeResource][v1alpha1.WorkflowActionConfigure]).To(HaveLen(0))
						Expect(pipelines[v1alpha1.WorkflowTypeResource][v1alpha1.WorkflowActionDelete]).To(HaveLen(0))
						Expect(pipelines[v1alpha1.WorkflowTypePromise][v1alpha1.WorkflowActionConfigure]).To(HaveLen(1))
						Expect(pipelines[v1alpha1.WorkflowTypePromise][v1alpha1.WorkflowActionDelete]).To(HaveLen(0))

						Expect(pipelines[v1alpha1.WorkflowTypePromise][v1alpha1.WorkflowActionConfigure][0].Name).To(Equal("pipeline0"))
						Expect(pipelines[v1alpha1.WorkflowTypePromise][v1alpha1.WorkflowActionConfigure][0].Spec.Containers).To(HaveLen(1))
						Expect(pipelines[v1alpha1.WorkflowTypePromise][v1alpha1.WorkflowActionConfigure][0].Spec.Containers[0].Image).To(Equal("image:latest"))
						Expect(pipelines[v1alpha1.WorkflowTypePromise][v1alpha1.WorkflowActionConfigure][0].Spec.Containers[0].Name).To(Equal("image"))

						Expect(sess.Out).To(gbytes.Say(fmt.Sprintf("generated the promise/configure/pipeline0/image in %s/promise.yaml", dir)))
						Expect(sess.Out).To(gbytes.Say("Customise your container by editing workflows/promise/configure/pipeline0/image/scripts/pipeline.sh"))

						script := getPipelineScriptContents(dir, "promise", "configure", "pipeline0", "image")
						Expect(script).To(ContainSubstring("#!/usr/bin/env sh"))
						Expect(script).To(ContainSubstring("Hello from ${name} ${namespace}"))

						dockerfile := getPipelineDockerfile(dir, "promise", "configure", "pipeline0", "image")
						Expect(dockerfile).To(ContainSubstring("FROM \"alpine\""))

						Expect(sess.Out).To(gbytes.Say("Don't forget to build and push your image!"))
					})
				})

				Context("go", func() {
					It("generates the expected files and adds containers to the promise workflows", func() {
						sess := r.run("add", "container", "promise/configure/pipeline0", "--image", "image:latest", "--dir", dir, "--language", "go")

						pipelines := getWorkflows(dir)
						Expect(pipelines[v1alpha1.WorkflowTypeResource][v1alpha1.WorkflowActionConfigure]).To(HaveLen(0))
						Expect(pipelines[v1alpha1.WorkflowTypeResource][v1alpha1.WorkflowActionDelete]).To(HaveLen(0))
						Expect(pipelines[v1alpha1.WorkflowTypePromise][v1alpha1.WorkflowActionConfigure]).To(HaveLen(1))
						Expect(pipelines[v1alpha1.WorkflowTypePromise][v1alpha1.WorkflowActionDelete]).To(HaveLen(0))

						Expect(pipelines[v1alpha1.WorkflowTypePromise][v1alpha1.WorkflowActionConfigure][0].Name).To(Equal("pipeline0"))
						Expect(pipelines[v1alpha1.WorkflowTypePromise][v1alpha1.WorkflowActionConfigure][0].Spec.Containers).To(HaveLen(1))
						Expect(pipelines[v1alpha1.WorkflowTypePromise][v1alpha1.WorkflowActionConfigure][0].Spec.Containers[0].Image).To(Equal("image:latest"))
						Expect(pipelines[v1alpha1.WorkflowTypePromise][v1alpha1.WorkflowActionConfigure][0].Spec.Containers[0].Name).To(Equal("image"))

						Expect(sess.Out).To(gbytes.Say(fmt.Sprintf("generated the promise/configure/pipeline0/image in %s/promise.yaml", dir)))
						Expect(sess.Out).To(gbytes.Say("Customise your container by editing workflows/promise/configure/pipeline0/image/scripts/pipeline.go"))

						scriptFilename := getPipelineScriptFilename(dir, "promise", "configure", "pipeline0", "image")
						Expect(scriptFilename).To(Equal("pipeline.go"))
						script := getPipelineScriptContents(dir, "promise", "configure", "pipeline0", "image")
						Expect(script).To(ContainSubstring("github.com/syntasso/kratix-go"))
						Expect(script).To(ContainSubstring(`fmt.Printf("Hello from %s", sdk.PromiseName())`))

						dockerfile := getPipelineDockerfile(dir, "promise", "configure", "pipeline0", "image")
						Expect(dockerfile).To(ContainSubstring("FROM golang"))

						Expect(sess.Out).To(gbytes.Say("For go containers, run 'go mod init' and 'go mod tidy' to manage your script's dependencies"))
						Expect(sess.Out).To(gbytes.Say("Don't forget to build and push your image!"))
					})
				})

				Context("python", func() {
					It("generates the expected files and adds containers to the promise workflows", func() {
						sess := r.run("add", "container", "promise/configure/pipeline0", "--image", "image:latest", "--dir", dir, "--language", "python")

						pipelines := getWorkflows(dir)
						Expect(pipelines[v1alpha1.WorkflowTypeResource][v1alpha1.WorkflowActionConfigure]).To(HaveLen(0))
						Expect(pipelines[v1alpha1.WorkflowTypeResource][v1alpha1.WorkflowActionDelete]).To(HaveLen(0))
						Expect(pipelines[v1alpha1.WorkflowTypePromise][v1alpha1.WorkflowActionConfigure]).To(HaveLen(1))
						Expect(pipelines[v1alpha1.WorkflowTypePromise][v1alpha1.WorkflowActionDelete]).To(HaveLen(0))

						Expect(pipelines[v1alpha1.WorkflowTypePromise][v1alpha1.WorkflowActionConfigure][0].Name).To(Equal("pipeline0"))
						Expect(pipelines[v1alpha1.WorkflowTypePromise][v1alpha1.WorkflowActionConfigure][0].Spec.Containers).To(HaveLen(1))
						Expect(pipelines[v1alpha1.WorkflowTypePromise][v1alpha1.WorkflowActionConfigure][0].Spec.Containers[0].Image).To(Equal("image:latest"))
						Expect(pipelines[v1alpha1.WorkflowTypePromise][v1alpha1.WorkflowActionConfigure][0].Spec.Containers[0].Name).To(Equal("image"))

						Expect(sess.Out).To(gbytes.Say(fmt.Sprintf("generated the promise/configure/pipeline0/image in %s/promise.yaml", dir)))
						Expect(sess.Out).To(gbytes.Say("Customise your container by editing workflows/promise/configure/pipeline0/image/scripts/pipeline.py"))

						scriptFilename := getPipelineScriptFilename(dir, "promise", "configure", "pipeline0", "image")
						Expect(scriptFilename).To(Equal("pipeline.py"))
						script := getPipelineScriptContents(dir, "promise", "configure", "pipeline0", "image")
						Expect(script).To(ContainSubstring("import kratix_sdk as ks"))
						Expect(script).To(ContainSubstring(`f'Hello from {sdk.promise_name()}'`))

						dockerfile := getPipelineDockerfile(dir, "promise", "configure", "pipeline0", "image")
						Expect(dockerfile).To(ContainSubstring("FROM python"))
						Expect(sess.Out).To(gbytes.Say("Don't forget to build and push your image!"))
					})
				})
			})

			When("the files were generated with the --split flag", func() {
				var dir string

				BeforeEach(func() {
					var err error
					dir, err = os.MkdirTemp("", "kratix-update-api-test")
					Expect(err).NotTo(HaveOccurred())

					sess := r.run("init", "promise", "postgresql", "--group", "syntasso.io", "--kind", "Database", "--dir", dir, "--split")
					Expect(sess.Out).To(gbytes.Say("postgresql promise bootstrapped in"))
				})

				AfterEach(func() {
					os.RemoveAll(dir)
				})

				It("adds containers to promise workflows", func() {
					sess := r.run("add", "container", "promise/configure/pipeline0", "--image", "image:latest", "--dir", dir)
					r.run("add", "container", "promise/configure/pipeline0", "--image", "image:latest", "-n", "superb-image", "--dir", dir)
					Expect(sess.Out).To(gbytes.Say(fmt.Sprintf("generated the promise/configure/pipeline0/image in %s/workflows/promise/configure/workflow.yaml", dir)))

					_, err := os.Stat(filepath.Join(dir, "workflows", "promise", "configure", "workflow.yaml"))
					Expect(err).ToNot(HaveOccurred())

					pipelines := getWorkflowsFromSplitFile(dir, "promise", "configure")

					Expect(pipelines[0].Name).To(Equal("pipeline0"))
					Expect(pipelines[0].Spec.Containers).To(HaveLen(2))
					Expect(pipelines[0].Spec.Containers[0].Name).To(Equal("image"))
					Expect(pipelines[0].Spec.Containers[0].Image).To(Equal("image:latest"))
					Expect(pipelines[0].Spec.Containers[1].Name).To(Equal("superb-image"))
					Expect(pipelines[0].Spec.Containers[1].Image).To(Equal("image:latest"))

					Expect(sess.Out).To(gbytes.Say("Customise your container by editing workflows/promise/configure/pipeline0/image/scripts/pipeline.sh"))
					Expect(sess.Out).To(gbytes.Say("Don't forget to build and push your image!"))

					Expect(getPipelineScriptContents(dir, "promise", "configure", "pipeline0", "image")).To(ContainSubstring("Hello from ${name} ${namespace}"))
					Expect(pipelineWorkflowPathExists(dir, "promise", "configure", "pipeline0", "image", "Dockerfile")).To(BeTrue())
					Expect(pipelineWorkflowPathExists(dir, "promise", "configure", "pipeline0", "image", "resources/")).To(BeTrue())

					Expect(getPipelineScriptContents(dir, "promise", "configure", "pipeline0", "superb-image")).To(ContainSubstring("Hello from ${name} ${namespace}"))
					Expect(pipelineWorkflowPathExists(dir, "promise", "configure", "pipeline0", "superb-image", "Dockerfile")).To(BeTrue())
					Expect(pipelineWorkflowPathExists(dir, "promise", "configure", "pipeline0", "superb-image", "resources/")).To(BeTrue())
				})
			})
		})
	})
})

func getWorkflows(dir string) map[v1alpha1.Type]map[v1alpha1.Action][]v1alpha1.Pipeline {
	promiseYAML, err := os.ReadFile(filepath.Join(dir, "promise.yaml"))
	ExpectWithOffset(1, err).NotTo(HaveOccurred())

	var promise v1alpha1.Promise
	ExpectWithOffset(1, yaml.Unmarshal(promiseYAML, &promise)).To(Succeed())

	pipelines, err := v1alpha1.NewPipelinesMap(&promise, logr.FromContextOrDiscard(context.Background()))

	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	return pipelines
}

func getWorkflowsFromSplitFile(dir, workflowName, action string) []v1alpha1.Pipeline {
	workflowYAML, err := os.ReadFile(filepath.Join(dir, "workflows", workflowName, action, "workflow.yaml"))
	ExpectWithOffset(1, err).NotTo(HaveOccurred())

	var workflows []v1alpha1.Pipeline
	ExpectWithOffset(1, yaml.Unmarshal(workflowYAML, &workflows)).To(Succeed())

	return workflows
}

func getPipelineScriptContents(dir, workflow, action, pipelineName, containerName string) string {
	filename := getPipelineScriptFilename(dir, workflow, action, pipelineName, containerName)
	script, err := os.ReadFile(filepath.Join(dir, "workflows", workflow, action, pipelineName, containerName, "scripts", filename))
	ExpectWithOffset(1, err).NotTo(HaveOccurred())

	return string(script)
}

func getPipelineScriptFilename(dir, workflow, action, pipelineName, containerName string) string {
	var filename string
	content, err := os.ReadDir(filepath.Join(dir, "workflows", workflow, action, pipelineName, containerName, "scripts"))
	Expect(err).To(Not(HaveOccurred()), "error listing content in scripts directory")
	for _, file := range content {
		if strings.HasPrefix(file.Name(), "pipeline.") {
			filename = file.Name()
		}
	}
	return filename
}

func getPipelineDockerfile(dir, workflow, action, pipelineName, containerName string) string {
	dockerfile, err := os.ReadFile(filepath.Join(dir, "workflows", workflow, action, pipelineName, containerName, "Dockerfile"))
	ExpectWithOffset(1, err).NotTo(HaveOccurred())

	return string(dockerfile)
}

func pipelineWorkflowPathExists(dir, workflow, action, pipelineName, containerName, filename string) bool {
	var found = false
	_, err := os.Stat(filepath.Join(dir, "workflows", workflow, action, pipelineName, containerName, filename))
	if err == nil {
		found = true
	}
	return found
}

func generateExpectedContainerName(image string) string {
	name := strings.Split(image, ":")[0]
	name = strings.NewReplacer("/", "-", ".", "-").Replace(name)
	name = strings.Trim(name, "-")
	if len(name) <= 63 {
		return name
	}
	h := sha1.Sum([]byte(name))
	suffix := hex.EncodeToString(h[:])[:7]
	prefix := strings.TrimRight(name[:63-len(suffix)-1], "-")
	return fmt.Sprintf("%s-%s", prefix, suffix)
}
