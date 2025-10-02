package internal_test

import (
	"errors"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/syntasso/kratix-cli-plugin-investigation/internal"

	"github.com/hashicorp/go-getter"
)

var _ = Describe("DownloadAndConvertTerraformToCRD", func() {
	var dst, src, tempDir, variablesPath string

	BeforeEach(func() {
		var err error
		tempDir, err = os.MkdirTemp("", "test-tf-module")
		Expect(err).ToNot(HaveOccurred())

		variablesPath = filepath.Join(tempDir, "variables.tf")

		internal.SetMkdirTempFunc(func(dir, pattern string) (string, error) {
			return tempDir, nil
		})
	})

	AfterEach(func() {
		os.RemoveAll(tempDir)
	})

	Context("when the module is successfully downloaded and parsed", func() {
		BeforeEach(func() {
			// Mock getter function to simulate a successful download
			internal.SetGetModuleFunc(func(givenDst, givenSrc string, opts ...getter.ClientOption) error {
				dst = givenDst
				src = givenSrc
				return os.WriteFile(variablesPath, []byte(`
					variable "example_var" {
					  type        = string
					  description = "An example variable"
					}

					variable "complex_var" {
					  type        = list(map(string))
					  description = "A complex variable"
					}

					variable "number_var" {
					  type        = number
					}

					variable "bool_var" {
					  type        = bool
					}
				`), 0644)
			})
		})

		It("returns a list of variables with correct types and descriptions", func() {
			variables, err := internal.GetVariablesFromModule("mock-source", "")
			Expect(err).ToNot(HaveOccurred())
			Expect(src).To(Equal("mock-source"))
			Expect(dst).To(Equal(tempDir))
			Expect(variables).To(HaveLen(4))

			Expect(variables[0].Name).To(Equal("example_var"))
			Expect(variables[0].Type).To(Equal("string"))
			Expect(variables[0].Description).To(Equal("An example variable"))

			Expect(variables[1].Name).To(Equal("complex_var"))
			Expect(variables[1].Type).To(Equal("list(map(string))"))
			Expect(variables[1].Description).To(Equal("A complex variable"))

			Expect(variables[2].Name).To(Equal("number_var"))
			Expect(variables[2].Type).To(Equal("number"))
			Expect(variables[2].Description).To(BeEmpty())

			Expect(variables[3].Name).To(Equal("bool_var"))
			Expect(variables[3].Type).To(Equal("bool"))
			Expect(variables[3].Description).To(BeEmpty())
		})
	})

	Context("when the variables.tf file is not at the root of the module", func() {
		BeforeEach(func() {
			Expect(os.MkdirAll(filepath.Join(tempDir, "subdir"), 0755)).To(Succeed())
			variablesPath = filepath.Join(tempDir, "subdir", "variables.tf")
			// Mock getter function to simulate a successful download
			internal.SetGetModuleFunc(func(givenDst, givenSrc string, opts ...getter.ClientOption) error {
				dst = givenDst
				src = givenSrc
				return os.WriteFile(variablesPath, []byte(`
					variable "example_var" {
					  type        = string
					  description = "An example variable"
					}

					variable "complex_var" {
					  type        = list(map(string))
					  description = "A complex variable"
					}

					variable "number_var" {
					  type        = number
					}

					variable "bool_var" {
					  type        = bool
					}
				`), 0644)
			})
		})

		It("returns a list of variables with correct types and descriptions", func() {
			variables, err := internal.GetVariablesFromModule("mock-source", "subdir")
			Expect(err).ToNot(HaveOccurred())
			Expect(src).To(Equal("mock-source"))
			Expect(dst).To(Equal(tempDir))
			Expect(variables).To(HaveLen(4))

			Expect(variables[0].Name).To(Equal("example_var"))
			Expect(variables[0].Type).To(Equal("string"))
			Expect(variables[0].Description).To(Equal("An example variable"))

			Expect(variables[1].Name).To(Equal("complex_var"))
			Expect(variables[1].Type).To(Equal("list(map(string))"))
			Expect(variables[1].Description).To(Equal("A complex variable"))

			Expect(variables[2].Name).To(Equal("number_var"))
			Expect(variables[2].Type).To(Equal("number"))
			Expect(variables[2].Description).To(BeEmpty())

			Expect(variables[3].Name).To(Equal("bool_var"))
			Expect(variables[3].Type).To(Equal("bool"))
			Expect(variables[3].Description).To(BeEmpty())
		})
	})

	Context("when the module download fails", func() {
		It("errors", func() {
			internal.SetGetModuleFunc(func(dst, src string, opts ...getter.ClientOption) error {
				return errors.New("mock download failure")
			})

			_, err := internal.GetVariablesFromModule("mock-source", "")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to download module"))
		})
	})

	Context("when the module is downloaded but variable parsing fails", func() {
		It("errors", func() {
			internal.SetGetModuleFunc(func(dst, src string, opts ...getter.ClientOption) error {
				return os.WriteFile(variablesPath, []byte(`invalid hcl`), 0644)
			})

			_, err := internal.GetVariablesFromModule("mock-source", "")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to parse variables"))
		})
	})
})
