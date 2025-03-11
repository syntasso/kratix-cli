package internal_test

import (
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

		// Run schema conversion
		schema, warnings := internal.VariablesToCRDSpecSchema(vars)

		Expect(warnings).To(BeEmpty(), "There should be no warnings for supported types")
		Expect(len(schema.Properties)).To(Equal(len(vars)))
	})

	Context("when processing variables with inferred types from defaults", func() {
		It("should correctly infer types from default values", func() {
			vars := []internal.TerraformVariable{
				{Name: "defaultList", Default: []any{1, 2, 3}},
				{Name: "defaultString", Default: "example"},
				{Name: "defaultNumber", Default: 42},
				{Name: "defaultBoolean", Default: true},
			}

			schema, warnings := internal.VariablesToCRDSpecSchema(vars)

			Expect(warnings).To(BeEmpty(), "There should be no warnings for inferred types")
			Expect(schema.Properties).To(HaveKey("defaultList"))
			Expect(schema.Properties["defaultList"].Type).To(Equal("array"))
			Expect(schema.Properties["defaultList"].Items.Schema.Type).To(Equal("number"))

			Expect(schema.Properties["defaultString"].Type).To(Equal("string"))
			Expect(schema.Properties["defaultNumber"].Type).To(Equal("number"))
			Expect(schema.Properties["defaultBoolean"].Type).To(Equal("boolean"))
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
