package integration_test

import (
	"context"
	"os"
	"path/filepath"

	ctrl "sigs.k8s.io/controller-runtime"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/syntasso/kratix/api/v1alpha1"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
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
				Expect(sess.Out).To(gbytes.Say("generated the promise/configure/pipeline0/image"))
				r.run("add", "container", "promise/configure/pipeline1", "--image", "project/image1:latest", "-n", "a-good-container", "--dir", dir)
				r.run("add", "container", "promise/delete/pipeline0", "--image", "project/cleanup:latest", "--dir", dir)

				pipelines := getWorkflows(dir)
				Expect(pipelines.ConfigureResource).To(HaveLen(0))
				Expect(pipelines.DeleteResource).To(HaveLen(0))
				Expect(pipelines.ConfigurePromise).To(HaveLen(2))
				Expect(pipelines.DeletePromise).To(HaveLen(1))

				Expect(pipelines.ConfigurePromise[0].Name).To(Equal("pipeline0"))
				Expect(pipelines.ConfigurePromise[0].Spec.Containers).To(HaveLen(1))
				Expect(pipelines.ConfigurePromise[0].Spec.Containers[0].Image).To(Equal("image:latest"))
				Expect(pipelines.ConfigurePromise[0].Spec.Containers[0].Name).To(Equal("image"))
				Expect(pipelines.ConfigurePromise[1].Name).To(Equal("pipeline1"))
				Expect(pipelines.ConfigurePromise[1].Spec.Containers).To(HaveLen(1))
				Expect(pipelines.ConfigurePromise[1].Spec.Containers[0].Image).To(Equal("project/image1:latest"))
				Expect(pipelines.ConfigurePromise[1].Spec.Containers[0].Name).To(Equal("a-good-container"))

				Expect(pipelines.DeletePromise[0].Name).To(Equal("pipeline0"))
				Expect(pipelines.DeletePromise[0].Spec.Containers).To(HaveLen(1))
				Expect(pipelines.DeletePromise[0].Spec.Containers[0].Image).To(Equal("project/cleanup:latest"))
				Expect(pipelines.DeletePromise[0].Spec.Containers[0].Name).To(Equal("project-cleanup"))

				Expect(sess.Out).To(gbytes.Say("Customise your container by editing the workflows/promise/configure/pipeline0/containers/scripts/pipeline.sh"))
				script := getPipelineScript(dir, "promise", "configure", "pipeline1")
				Expect(script).To(ContainSubstring("Hello from ${name} ${namespace}"))
				Expect(sess.Out).To(gbytes.Say("Don't forget to build and push your image!"))

				Expect(pipelineWorkflowPathExists(dir, "promise", "configure", "pipeline0", "Dockerfile")).To(BeTrue())
				Expect(pipelineWorkflowPathExists(dir, "promise", "configure", "pipeline0", "resources/")).To(BeTrue())
				Expect(pipelineWorkflowPathExists(dir, "promise", "configure", "pipeline1", "Dockerfile")).To(BeTrue())
				Expect(pipelineWorkflowPathExists(dir, "promise", "configure", "pipeline0", "resources/")).To(BeTrue())
				Expect(pipelineWorkflowPathExists(dir, "promise", "delete", "pipeline0", "Dockerfile")).To(BeTrue())
				Expect(pipelineWorkflowPathExists(dir, "promise", "delete", "pipeline0", "resources/")).To(BeTrue())
			})

			It("adds containers to resource workflows", func() {
				r.run("add", "container", "resource/configure/pipeline0", "--image", "project/image1:latest", "-n", "a-great-container", "--dir", dir)
				r.run("add", "container", "resource/configure/pipeline0", "--image", "project/image2:latest", "--dir", dir)
				r.run("add", "container", "resource/delete/pipeline0", "--image", "project/cleanup:latest", "--dir", dir)

				pipelines := getWorkflows(dir)
				Expect(pipelines.ConfigurePromise).To(HaveLen(0))
				Expect(pipelines.DeletePromise).To(HaveLen(0))

				Expect(pipelines.ConfigureResource).To(HaveLen(1))
				Expect(pipelines.ConfigureResource[0].Name).To(Equal("pipeline0"))
				Expect(pipelines.ConfigureResource[0].Spec.Containers).To(HaveLen(2))
				Expect(pipelines.ConfigureResource[0].Spec.Containers[0].Image).To(Equal("project/image1:latest"))
				Expect(pipelines.ConfigureResource[0].Spec.Containers[0].Name).To(Equal("a-great-container"))
				Expect(pipelines.ConfigureResource[0].Spec.Containers[1].Image).To(Equal("project/image2:latest"))
				Expect(pipelines.ConfigureResource[0].Spec.Containers[1].Name).To(Equal("project-image2"))

				Expect(pipelines.DeleteResource).To(HaveLen(1))
				Expect(pipelines.DeleteResource[0].Name).To(Equal("pipeline0"))
				Expect(pipelines.DeleteResource[0].Spec.Containers).To(HaveLen(1))
				Expect(pipelines.DeleteResource[0].Spec.Containers[0].Image).To(Equal("project/cleanup:latest"))
				Expect(pipelines.DeleteResource[0].Spec.Containers[0].Name).To(Equal("project-cleanup"))
			})
			When("the files were generated with the --split flag", func() {
				var dir string
				AfterEach(func() {
					os.RemoveAll(dir)
				})

				BeforeEach(func() {
					var err error
					dir, err = os.MkdirTemp("", "kratix-update-api-test")
					Expect(err).NotTo(HaveOccurred())

					sess := r.run("init", "promise", "postgresql", "--group", "syntasso.io", "--kind", "Database", "--dir", dir, "--split")
					Expect(sess.Out).To(gbytes.Say("postgresql promise bootstrapped in"))
				})
				It("adds containers to promise workflows", func() {
					sess := r.run("add", "container", "promise/configure/pipeline0", "--image", "image:latest", "--dir", dir)
					r.run("add", "container", "promise/configure/pipeline0", "--image", "image:latest", "-n", "superb-image", "--dir", dir)
					Expect(sess.Out).To(gbytes.Say("generated the promise/configure/pipeline0/image"))

					_, err := os.Stat(filepath.Join(dir, "workflows", "promise", "configure", "workflow.yaml"))
					Expect(err).ToNot(HaveOccurred())

					workflow := getWorkflowsFromSplitFile(dir, "promise", "configure")
					Expect(workflow.Resource.Configure).To(HaveLen(0))
					Expect(workflow.Resource.Delete).To(HaveLen(0))
					Expect(workflow.Promise.Delete).To(HaveLen(0))
					Expect(workflow.Promise.Configure).To(HaveLen(1))

					Expect(unstructuredToPipelines(workflow.Promise.Configure)[0].Name).To(Equal("pipeline0"))
					Expect(unstructuredToPipelines(workflow.Promise.Configure)[0].Spec.Containers).To(HaveLen(2))
					Expect(unstructuredToPipelines(workflow.Promise.Configure)[0].Spec.Containers[0].Name).To(Equal("image"))
					Expect(unstructuredToPipelines(workflow.Promise.Configure)[0].Spec.Containers[0].Image).To(Equal("image:latest"))
					Expect(unstructuredToPipelines(workflow.Promise.Configure)[0].Spec.Containers[1].Name).To(Equal("superb-image"))
					Expect(unstructuredToPipelines(workflow.Promise.Configure)[0].Spec.Containers[1].Image).To(Equal("image:latest"))

					Expect(sess.Out).To(gbytes.Say("Customise your container by editing the workflows/promise/configure/pipeline0/containers/scripts/pipeline.sh"))
					script := getPipelineScript(dir, "promise", "configure", "pipeline0")
					Expect(script).To(ContainSubstring("Hello from ${name} ${namespace}"))
					Expect(sess.Out).To(gbytes.Say("Don't forget to build and push your image!"))

					Expect(pipelineWorkflowPathExists(dir, "promise", "configure", "pipeline0", "Dockerfile")).To(BeTrue())
					Expect(pipelineWorkflowPathExists(dir, "promise", "configure", "pipeline0", "resources/")).To(BeTrue())
				})
			})

		})
	})
})

func getWorkflows(dir string) v1alpha1.PromisePipelines {
	promiseYAML, err := os.ReadFile(filepath.Join(dir, "promise.yaml"))
	ExpectWithOffset(1, err).NotTo(HaveOccurred())

	var promise v1alpha1.Promise
	ExpectWithOffset(1, yaml.Unmarshal(promiseYAML, &promise)).To(Succeed())

	pipelines, err := promise.GeneratePipelines(ctrl.LoggerFrom(context.Background()))
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	return pipelines
}

func getWorkflowsFromSplitFile(dir, workflowName, action string) v1alpha1.Workflows {
	workflowYAML, err := os.ReadFile(filepath.Join(dir, "workflows", workflowName, action, "workflow.yaml"))
	ExpectWithOffset(1, err).NotTo(HaveOccurred())

	var workflows v1alpha1.Workflows
	ExpectWithOffset(1, yaml.Unmarshal(workflowYAML, &workflows)).To(Succeed())

	return workflows
}

func unstructuredToPipelines(objects []unstructured.Unstructured) []v1alpha1.Pipeline {
	var pipelines = []v1alpha1.Pipeline{}
	for _, u := range objects {
		var pipeline v1alpha1.Pipeline
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &pipeline)
		if err != nil {
			return []v1alpha1.Pipeline{}
		}
		pipelines = append(pipelines, pipeline)
	}
	return pipelines
}
func getPipelineScript(dir, workflow, action, pipelineName string) string {
	promiseYAML, err := os.ReadFile(filepath.Join(dir, "workflows", workflow, action, pipelineName, "containers", "scripts", "pipeline.sh"))
	ExpectWithOffset(1, err).NotTo(HaveOccurred())

	return string(promiseYAML)
}

func pipelineWorkflowPathExists(dir, workflow, action, pipelineName, filename string) bool {
	var found = false
	_, err := os.Stat(filepath.Join(dir, "workflows", workflow, action, pipelineName, "containers", filename))
	if err == nil {
		found = true
	}
	return found
}
