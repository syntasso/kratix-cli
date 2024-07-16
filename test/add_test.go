package integration_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	ctrl "sigs.k8s.io/controller-runtime"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/syntasso/kratix/api/v1alpha1"

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
				script := getPipelineScript(dir, "promise", "configure", "pipeline1", "a-good-container")
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
					Expect(getPipelineScript(dir, "promise", "configure", "pipeline0", "a-great-container")).To(ContainSubstring("Hello from ${name} ${namespace}"))

					Expect(pipelineWorkflowPathExists(dir, "promise", "configure", "pipeline0", "an-even-greater-container", "Dockerfile")).To(BeTrue())
					Expect(pipelineWorkflowPathExists(dir, "promise", "configure", "pipeline0", "an-even-greater-container", "resources/")).To(BeTrue())
					Expect(getPipelineScript(dir, "promise", "configure", "pipeline0", "an-even-greater-container")).To(ContainSubstring("Hello from ${name} ${namespace}"))
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

					Expect(getPipelineScript(dir, "promise", "configure", "pipeline0", "image")).To(ContainSubstring("Hello from ${name} ${namespace}"))
					Expect(pipelineWorkflowPathExists(dir, "promise", "configure", "pipeline0", "image", "Dockerfile")).To(BeTrue())
					Expect(pipelineWorkflowPathExists(dir, "promise", "configure", "pipeline0", "image", "resources/")).To(BeTrue())

					Expect(getPipelineScript(dir, "promise", "configure", "pipeline0", "superb-image")).To(ContainSubstring("Hello from ${name} ${namespace}"))
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

	pipelines, err := v1alpha1.NewPipelinesMap(&promise, ctrl.LoggerFrom(context.Background()))

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

func getPipelineScript(dir, workflow, action, pipelineName, containerName string) string {
	promiseYAML, err := os.ReadFile(filepath.Join(dir, "workflows", workflow, action, pipelineName, containerName, "scripts", "pipeline.sh"))
	ExpectWithOffset(1, err).NotTo(HaveOccurred())

	return string(promiseYAML)
}

func pipelineWorkflowPathExists(dir, workflow, action, pipelineName, containerName, filename string) bool {
	var found = false
	_, err := os.Stat(filepath.Join(dir, "workflows", workflow, action, pipelineName, containerName, filename))
	if err == nil {
		found = true
	}
	return found
}
