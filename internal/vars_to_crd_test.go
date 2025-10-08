package internal_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/syntasso/kratix-cli/internal"
)

var _ = Describe("VariablesToCRDSpecSchema", func() {
	It("should generate the correct JSONSchemaProps", func() {
		vars := []internal.TerraformVariable{
			{Name: "stringVar", Type: "string"},
			{Name: "numberVar", Type: "number"},
			{Name: "boolVar", Type: "bool"},
			{Name: "listStringVar", Type: "list(string)"},
			{Name: "mapStringVar", Type: "map(string)"},
			{Name: "listObjectVar", Type: "list(object({ key1 = string, key2 = number }))"},
			{Name: "complexMap", Type: "map(map(list(string)))"},
			{Name: "complexMap2", Type: "map(object({ key1 = string, key2 = bool }))"},
			{Name: "deeplyNested", Type: "list(object({ name = string, secret = set(object({ secret_name = string, items = map(string) })) }))"},
			{Name: "probe", Type: "object({ failure_threshold = optional(number, null), initial_delay_seconds = optional(number, null), http_get = optional(object({ path = optional(string), http_headers = optional(list(object({ name = string, value = string })), null) }), null) })"},
		}

		schema, warnings := internal.VariablesToCRDSpecSchema(vars)

		Expect(warnings).To(BeEmpty(), "There should be no warnings for supported types")
		Expect(len(schema.Properties)).To(Equal(len(vars)))
	})

	Context("when supported variables have a default", func() {
		It("should correctly defaults", func() {
			vars := []internal.TerraformVariable{
				{Name: "defaultList", Default: []any{1, 2, 3}},
				{Name: "defaultString", Default: "example"},
				{Name: "defaultNumber", Default: 42},
				{Name: "defaultBoolean", Default: true},
				{Name: "defaultEmptyList", Type: "list(string)", Default: []any{}},
				{Name: "defaultEmptyMap", Type: "map(string)", Default: map[string]any{}},
				{Name: "defaultNonEmptyList", Type: "list(string)", Default: []any{"one", "two"}},
				{Name: "defaultNonEmptyMap", Type: "map(string)", Default: map[string]any{"key": "value"}},
			}

			schema, warnings := internal.VariablesToCRDSpecSchema(vars)
			Expect(warnings).To(BeEmpty(), "There should be no warnings for inferred types")

			Expect(schema.Properties).To(HaveKey("defaultList"))
			Expect(schema.Properties["defaultList"].Type).To(Equal("array"))

			var defaultList []float64
			Expect(json.Unmarshal(schema.Properties["defaultList"].Default.Raw, &defaultList)).To(Succeed())
			Expect(defaultList).To(Equal([]float64{1, 2, 3}))

			Expect(schema.Properties["defaultString"].Type).To(Equal("string"))
			var defaultString string
			Expect(json.Unmarshal(schema.Properties["defaultString"].Default.Raw, &defaultString)).To(Succeed())
			Expect(defaultString).To(Equal("example"))

			Expect(schema.Properties["defaultNumber"].Type).To(Equal("number"))
			var defaultNumber float64
			Expect(json.Unmarshal(schema.Properties["defaultNumber"].Default.Raw, &defaultNumber)).To(Succeed())
			Expect(defaultNumber).To(Equal(float64(42)))

			Expect(schema.Properties["defaultBoolean"].Type).To(Equal("boolean"))
			var defaultBool bool
			Expect(json.Unmarshal(schema.Properties["defaultBoolean"].Default.Raw, &defaultBool)).To(Succeed())
			Expect(defaultBool).To(BeTrue())

			Expect(schema.Properties["defaultEmptyList"].Type).To(Equal("array"))
			var defaultEmptyList []string
			Expect(json.Unmarshal(schema.Properties["defaultEmptyList"].Default.Raw, &defaultEmptyList)).To(Succeed())
			Expect(defaultEmptyList).To(BeEmpty())

			Expect(schema.Properties["defaultEmptyMap"].Type).To(Equal("object"))
			var defaultEmptyMap map[string]string
			Expect(json.Unmarshal(schema.Properties["defaultEmptyMap"].Default.Raw, &defaultEmptyMap)).To(Succeed())
			Expect(defaultEmptyMap).To(BeEmpty())

		})
	})

	Context("when unsupported variables have defaults", func() {
		It("should not set defaults", func() {
			vars := []internal.TerraformVariable{
				{Name: "defaultObject", Type: "object({ key = string })", Default: map[string]any{"key": "value"}},
			}

			schema, warnings := internal.VariablesToCRDSpecSchema(vars)

			Expect(warnings).To(ContainElement(
				"warning: default value for variable defaultObject is set but type object({ key = string }) does not support defaults, skipping",
			))

			Expect(schema.Properties).To(HaveKey("defaultObject"))
			Expect(schema.Properties["defaultObject"].Type).To(Equal("object"))
			Expect(schema.Properties["defaultObject"].Default).To(BeNil(), "Default should not be set for objects")
		})
	})

	Context("when processing variables with missing type and unsupported defaults", func() {
		It("should return warnings for unrecognized types", func() {
			vars := []internal.TerraformVariable{
				{Name: "unknownVar", Default: nil},
				{Name: "defaultMap", Default: map[string]any{"key": "value"}},
			}
			_, warnings := internal.VariablesToCRDSpecSchema(vars)

			Expect(warnings).To(ConsistOf(
				"warning: Type not set for variable unknownVar and cannot be inferred from the default value, skipping",
				"warning: Type not set for variable defaultMap and cannot be inferred from the default value, skipping",
			))
		})
	})
})
