package internal_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/syntasso/kratix-cli/internal"
)

var _ = Describe("VariablesToCRDSpecSchema", func() {
	Context("when processing supported Terraform types", func() {
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
			}

			schema, warnings := internal.VariablesToCRDSpecSchema(vars)

			Expect(warnings).To(BeEmpty(), "There should be no warnings for supported types")
			Expect(len(schema.Properties)).To(Equal(len(vars)))

			Expect(schema.Properties["stringVar"].Type).To(Equal("string"))
			Expect(schema.Properties["numberVar"].Type).To(Equal("number"))
			Expect(schema.Properties["boolVar"].Type).To(Equal("boolean"))

			Expect(schema.Properties["listStringVar"].Type).To(Equal("array"))
			Expect(schema.Properties["listStringVar"].Items.Schema.Type).To(Equal("string"))

			Expect(schema.Properties["mapStringVar"].Type).To(Equal("object"))
			Expect(schema.Properties["mapStringVar"].AdditionalProperties.Schema.Type).To(Equal("string"))

			Expect(schema.Properties["listObjectVar"].Type).To(Equal("array"))
			Expect(schema.Properties["listObjectVar"].Items.Schema.Type).To(Equal("object"))
			Expect(schema.Properties["listObjectVar"].Items.Schema.XPreserveUnknownFields).NotTo(BeNil())

			Expect(schema.Properties["complexMap"].Type).To(Equal("object"))
			Expect(schema.Properties["complexMap"].XPreserveUnknownFields).To(BeNil())
			Expect(schema.Properties["complexMap"].AdditionalProperties.Schema.Type).To(Equal("object"))
			Expect(schema.Properties["complexMap"].AdditionalProperties.Schema.AdditionalProperties.Schema.Type).To(Equal("array"))
			Expect(schema.Properties["complexMap"].AdditionalProperties.Schema.AdditionalProperties.Schema.Items.Schema.Type).To(Equal("string"))

			Expect(schema.Properties["complexMap2"].Type).To(Equal("object"))
			Expect(schema.Properties["complexMap2"].XPreserveUnknownFields).To(BeNil())
			Expect(schema.Properties["complexMap2"].AdditionalProperties.Schema.Type).To(Equal("object"))
			Expect(schema.Properties["complexMap2"].AdditionalProperties.Schema.XPreserveUnknownFields).NotTo(BeNil())
		})
	})

	Context("when processing complex nested Terraform types", func() {
		It("should correctly handle deeply nested structures", func() {
			vars := []internal.TerraformVariable{
				{Name: "deeplyNested", Type: "list(object({ name = string, secret = set(object({ secret_name = string, items = map(string) })) }))"},
				{Name: "probe", Type: "object({ failure_threshold = optional(number, null), initial_delay_seconds = optional(number, null), http_get = optional(object({ path = optional(string), http_headers = optional(list(object({ name = string, value = string })), null) }), null) })"},
			}

			schema, warnings := internal.VariablesToCRDSpecSchema(vars)

			Expect(warnings).To(BeEmpty(), "There should be no warnings for supported nested types")
			Expect(schema.Properties).To(HaveKey("deeplyNested"))
			Expect(schema.Properties["deeplyNested"].Type).To(Equal("array"))
			Expect(schema.Properties["deeplyNested"].Items.Schema.Type).To(Equal("object"))
			Expect(schema.Properties["deeplyNested"].Items.Schema.XPreserveUnknownFields).NotTo(BeNil())

			Expect(schema.Properties).To(HaveKey("probe"))
			Expect(schema.Properties["probe"].Type).To(Equal("object"))
			Expect(schema.Properties["probe"].XPreserveUnknownFields).NotTo(BeNil())
		})
	})

	Context("when processing unsupported Terraform types", func() {
		It("should return warnings for unsupported types", func() {
			vars := []internal.TerraformVariable{
				{Name: "setVar", Type: "set(string)"},
				{Name: "tupleVar", Type: "tuple([string, number])"},
			}

			_, warnings := internal.VariablesToCRDSpecSchema(vars)

			Expect(warnings).To(ConsistOf(
				"warning: unable to automatically convert set(string) into CRD, skipping",
				"warning: unable to automatically convert tuple([string, number]) into CRD, skipping",
			))
		})
	})
})
