package pipelineutils_test

import (
	"io/fs"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	pipelineutils "github.com/syntasso/kratix-cli/cmd/pipeline_utils"
	"github.com/syntasso/kratix/api/v1alpha1"
)

var _ = Describe("FindContainerIndex", func() {
	Describe("pipelineutils.FindContainerIndex", func() {
		var dirEntries []fs.DirEntry
		var containers []v1alpha1.Container
		BeforeEach(func() {
			dirEntries = []fs.DirEntry{
				&fakeDirEntry{name: "lettuce"},
				&fakeDirEntry{name: "mango"},
				&fakeDirEntry{name: "cabagge"},
			}

			containers = []v1alpha1.Container{
				{Name: "avocado"},
				{Name: "banana"},
				{Name: "mango"},
			}
		})

		When("the provided dirEntry is empty", func() {
			It("returns an error", func() {
				_, err := pipelineutils.FindContainerIndex(nil, containers, "")
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("no container found in path"))
			})
		})

		When("no name is provided", func() {
			Context("and there's a single directory", func() {
				BeforeEach(func() {
					dirEntries = []fs.DirEntry{
						&fakeDirEntry{name: "mango"},
					}
				})

				It("returns the index of the container", func() {
					index, err := pipelineutils.FindContainerIndex(dirEntries, containers, "")
					Expect(err).NotTo(HaveOccurred())
					Expect(index).To(Equal(2))
				})

				It("returns an error if the container is not found", func() {
					containers = containers[:2]
					_, err := pipelineutils.FindContainerIndex(dirEntries, containers, "")
					Expect(err).To(HaveOccurred())
					Expect(err).To(MatchError("container mango not found in pipeline"))
				})
			})

			Context("and there's more than one directory", func() {
				It("returns an error", func() {
					_, err := pipelineutils.FindContainerIndex(dirEntries, containers, "")
					Expect(err).To(HaveOccurred())
					Expect(err).To(MatchError("more than one container exists for this pipeline, please provide a name with --name"))
				})
			})
		})

		When("a name is provided", func() {
			It("returns the index of the container", func() {
				index, err := pipelineutils.FindContainerIndex(dirEntries, containers, "mango")
				Expect(err).NotTo(HaveOccurred())
				Expect(index).To(Equal(2))
			})

			It("returns an error if the container is not found", func() {
				_, err := pipelineutils.FindContainerIndex(dirEntries, containers, "lettuce")
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("container lettuce not found in pipeline"))
			})

			It("returns an error if the named container is not in the directory entries", func() {
				_, err := pipelineutils.FindContainerIndex(dirEntries, containers, "banana")
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("directory entry not found for container banana"))
			})

		})

	})
})

type fakeDirEntry struct {
	name string
}

func (f *fakeDirEntry) Name() string {
	return f.name
}

func (f *fakeDirEntry) IsDir() bool {
	return true
}

func (f *fakeDirEntry) Type() fs.FileMode {
	return fs.ModeDir
}

func (f *fakeDirEntry) Info() (fs.FileInfo, error) {
	return nil, nil
}
