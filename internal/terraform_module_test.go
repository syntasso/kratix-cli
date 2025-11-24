package internal_test

import (
	"errors"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/syntasso/kratix-cli/internal"

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
					  default     = 10
					}

					variable "bool_var" {
					  type        = bool
					}

					variable "list_string_var" {
					  type        = list(string)
					  default     = ["stringValue"]
					}

					variable "list_object_var" {
						type      = list(map(string))
						 default = [
							{
								key1 = "value1"
								key2 = 100
							}
						]
					}
				`), 0644)
			})
		})

		It("returns a list of variables with correct types and descriptions", func() {
			variables, err := internal.GetVariablesFromModule("mock-source", "")
			Expect(err).ToNot(HaveOccurred())
			Expect(src).To(Equal("mock-source"))
			Expect(dst).To(Equal(tempDir))
			Expect(variables).To(HaveLen(6))

			Expect(variables[0].Name).To(Equal("example_var"))
			Expect(variables[0].Type).To(Equal("string"))
			Expect(variables[0].Description).To(Equal("An example variable"))

			Expect(variables[1].Name).To(Equal("complex_var"))
			Expect(variables[1].Type).To(Equal("list(map(string))"))
			Expect(variables[1].Description).To(Equal("A complex variable"))

			Expect(variables[2].Name).To(Equal("number_var"))
			Expect(variables[2].Type).To(Equal("number"))
			Expect(variables[2].Default).To(BeAssignableToTypeOf(float64(0)))
			numberVarDefault, ok := variables[2].Default.(float64)
			Expect(ok).To(BeTrue())
			Expect(numberVarDefault).To(Equal(10.0))
			Expect(variables[2].Description).To(BeEmpty())

			Expect(variables[3].Name).To(Equal("bool_var"))
			Expect(variables[3].Type).To(Equal("bool"))
			Expect(variables[3].Description).To(BeEmpty())

			Expect(variables[4].Name).To(Equal("list_string_var"))
			Expect(variables[4].Type).To(Equal("list(string)"))
			Expect(variables[4].Default).To(BeAssignableToTypeOf([]string{"stringValue"}))
			listStringVarDefault, ok := variables[4].Default.([]string)
			Expect(ok).To(BeTrue())
			Expect(listStringVarDefault).To(Equal([]string{"stringValue"}))
			Expect(variables[4].Description).To(BeEmpty())

			Expect(variables[5].Name).To(Equal("list_object_var"))
			Expect(variables[5].Type).To(Equal("list(map(string))"))
			Expect(variables[5].Default).To(BeAssignableToTypeOf([]any{}))
			listObjectVarDefault, ok := variables[5].Default.([]any)
			Expect(ok).To(BeTrue())
			Expect(len(listObjectVarDefault)).To(Equal(1))
			Expect(listObjectVarDefault[0]).To(BeAssignableToTypeOf(map[string]any{}))
			listObjectVarDefaultObj, ok := listObjectVarDefault[0].(map[string]any)
			Expect(ok).To(BeTrue())
			Expect(listObjectVarDefaultObj).To(Equal(map[string]any{"key1": "value1", "key2": float64(100)}))
			Expect(variables[5].Description).To(BeEmpty())
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
