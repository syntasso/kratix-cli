package internal_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/syntasso/kratix-cli/internal"

	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
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
			Expect(len(schema.Properties)).To(Equal(1))
			Expect(len(schema.Properties["vars"].Properties)).To(Equal(8))

			Expect(schema.Properties["vars"].Properties).To(HaveKeyWithValue("stringVar", v1.JSONSchemaProps{Type: "string"}))
			Expect(schema.Properties["vars"].Properties).To(HaveKeyWithValue("numberVar", v1.JSONSchemaProps{Type: "number"}))
			Expect(schema.Properties["vars"].Properties).To(HaveKeyWithValue("boolVar", v1.JSONSchemaProps{Type: "boolean"}))

			Expect(schema.Properties["vars"].Properties).To(HaveKey("listStringVar"))
			Expect(schema.Properties["vars"].Properties["listStringVar"].Type).To(Equal("array"))
			Expect(schema.Properties["vars"].Properties["listStringVar"].Items.Schema.Type).To(Equal("string"))

			Expect(schema.Properties["vars"].Properties).To(HaveKey("mapStringVar"))
			Expect(schema.Properties["vars"].Properties["mapStringVar"].Type).To(Equal("object"))
			Expect(schema.Properties["vars"].Properties["mapStringVar"].AdditionalProperties.Schema.Type).To(Equal("string"))

			Expect(schema.Properties["vars"].Properties).To(HaveKey("listObjectVar"))
			Expect(schema.Properties["vars"].Properties["listObjectVar"].Type).To(Equal("array"))
			Expect(schema.Properties["vars"].Properties["listObjectVar"].Items.Schema.Type).To(Equal("object"))
			Expect(schema.Properties["vars"].Properties["listObjectVar"].Items.Schema.Properties).To(HaveKeyWithValue("key1", v1.JSONSchemaProps{Type: "string"}))
			Expect(schema.Properties["vars"].Properties["listObjectVar"].Items.Schema.Properties).To(HaveKeyWithValue("key2", v1.JSONSchemaProps{Type: "number"}))

			Expect(schema.Properties["vars"].Properties).To(HaveKey("complexMap"))
			Expect(schema.Properties["vars"].Properties["complexMap"].Type).To(Equal("object"))
			Expect(schema.Properties["vars"].Properties["complexMap"].XPreserveUnknownFields).NotTo(BeNil())
			Expect(*schema.Properties["vars"].Properties["complexMap"].XPreserveUnknownFields).To(BeTrue())

			Expect(schema.Properties["vars"].Properties).To(HaveKey("complexMap2"))
			Expect(schema.Properties["vars"].Properties["complexMap2"].Type).To(Equal("object"))
			Expect(schema.Properties["vars"].Properties["complexMap2"].XPreserveUnknownFields).NotTo(BeNil())
			Expect(*schema.Properties["vars"].Properties["complexMap2"].XPreserveUnknownFields).To(BeTrue())

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
