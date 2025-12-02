package internal_test

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/syntasso/kratix-cli/internal"
)

var _ = Describe("DownloadAndConvertTerraformToCRD", func() {
	var tempDir, variablesPath string

	BeforeEach(func() {
		var err error
		tempDir, err = os.MkdirTemp("", "test-tf-module")
		Expect(err).ToNot(HaveOccurred())

		variablesPath = filepath.Join(tempDir, ".terraform", "modules", "kratix_target", "variables.tf")

		internal.SetMkdirTempFunc(func(dir, pattern string) (string, error) {
			return tempDir, nil
		})
	})

	AfterEach(func() {
		os.RemoveAll(tempDir)
		internal.SetMkdirTempFunc(os.MkdirTemp)
		internal.SetTerraformInitFunc(internal.RunTerraformInit)
	})

	Context("when the module is successfully downloaded and parsed", func() {
		BeforeEach(func() {
			internal.SetTerraformInitFunc(func(dir string) error {
				Expect(os.MkdirAll(filepath.Dir(variablesPath), 0o755)).To(Succeed())
				manifestPath := filepath.Join(tempDir, ".terraform", "modules", "modules.json")
				expectManifest(manifestPath, ".terraform/modules/kratix_target")
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
`), 0o644)
			})
		})

		It("returns a list of variables with correct types and descriptions", func() {
			variables, err := internal.GetVariablesFromModule("mock-source", "", "")
			Expect(err).ToNot(HaveOccurred())
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
			expectDir := filepath.Join(tempDir, ".terraform", "modules", "kratix_target", "subdir")
			variablesPath = filepath.Join(expectDir, "variables.tf")
			internal.SetTerraformInitFunc(func(dir string) error {
				mainContent, err := os.ReadFile(filepath.Join(tempDir, "main.tf"))
				Expect(err).NotTo(HaveOccurred())
				expectContent := `module "kratix_target" {
  source = "git::mock-source.git//subdir"
  version = "v1.0.0"
}
`
				Expect(string(mainContent)).To(Equal(expectContent))

				expectManifest(filepath.Join(tempDir, ".terraform", "modules", "modules.json"), expectDir)
				Expect(os.MkdirAll(expectDir, 0o755)).To(Succeed())
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
`), 0o644)
			})

			variables, err := internal.GetVariablesFromModule("git::mock-source.git", "subdir", "v1.0.0")
			Expect(err).ToNot(HaveOccurred())
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

	Context("when terraform init fails", func() {
		It("errors", func() {
			internal.SetTerraformInitFunc(func(dir string) error {
				return errors.New("mock init failure")
			})

			_, err := internal.GetVariablesFromModule("mock-source", "", "")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to initialize terraform"))
		})
	})

	Context("when the module is downloaded but variable parsing fails", func() {
		It("errors", func() {
			internal.SetTerraformInitFunc(func(dir string) error {
				Expect(os.MkdirAll(filepath.Dir(variablesPath), 0o755)).To(Succeed())
				expectManifest(filepath.Join(tempDir, ".terraform", "modules", "modules.json"), ".terraform/modules/kratix_target")
				return os.WriteFile(variablesPath, []byte(`invalid hcl`), 0o644)
			})

			_, err := internal.GetVariablesFromModule("mock-source", "", "")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to parse variables"))
		})
	})
})

func expectManifest(manifestPath, moduleDir string) {
	manifest := fmt.Sprintf(`{"Modules":[{"Key":"module.%s","Dir":"%s"}]}`, "kratix_target", moduleDir)
	Expect(os.MkdirAll(filepath.Dir(manifestPath), 0o755)).To(Succeed())
	Expect(os.WriteFile(manifestPath, []byte(manifest), 0o644)).To(Succeed())
}
