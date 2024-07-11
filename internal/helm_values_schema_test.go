package internal_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/syntasso/kratix-cli/internal"
)

var _ = Describe("HelmValuesToSchema()", func() {
	It("handles string, integer, float, and boolean type", func() {
		values := map[string]interface{}{
			"aname":   "",
			"anumber": 10,
			"afloat":  float64(10),
			"abool":   false,
		}
		schema, err := internal.HelmValuesToSchema(values)
		Expect(err).NotTo(HaveOccurred())
		Expect(schema.Properties).NotTo(BeNil())
		Expect(schema.Properties["aname"].Type).To(Equal("string"))
		Expect(schema.Properties["anumber"].Type).To(Equal("integer"))
		Expect(schema.Properties["afloat"].Type).To(Equal("number"))
		Expect(schema.Properties["abool"].Type).To(Equal("boolean"))
	})

	It("handles object", func() {
		values := map[string]interface{}{
			"abool": true,
			"object": map[string]interface{}{
				"aname": "test",
				"map":   map[string]interface{}{"test": 10},
			},
		}
		schema, err := internal.HelmValuesToSchema(values)
		Expect(err).NotTo(HaveOccurred())
		Expect(schema.Properties).NotTo(BeNil())
		Expect(schema.Properties["abool"].Type).To(Equal("boolean"))
		Expect(schema.Properties["object"].Type).To(Equal("object"))
		Expect(*schema.Properties["object"].XPreserveUnknownFields).To(BeTrue())
		Expect(schema.Properties["object"].Properties["aname"].Type).To(Equal("string"))
		Expect(schema.Properties["object"].Properties["map"].Type).To(Equal("object"))
		Expect(schema.Properties["object"].Properties["map"].Properties["test"].Type).To(Equal("integer"))
	})

	It("handles array", func() {
		values := map[string]interface{}{
			"strArr":   []interface{}{"test0", "test1"},
			"intArr":   []interface{}{10, 20, 30},
			"emptyArr": []interface{}{},
		}
		schema, err := internal.HelmValuesToSchema(values)
		Expect(err).NotTo(HaveOccurred())
		Expect(schema.Properties).NotTo(BeNil())
		Expect(schema.Properties["strArr"].Type).To(Equal("array"))
		Expect(schema.Properties["strArr"].Items.Schema.Type).To(Equal("string"))
		Expect(schema.Properties["intArr"].Type).To(Equal("array"))
		Expect(schema.Properties["intArr"].Items.Schema.Type).To(Equal("integer"))
		Expect(schema.Properties["emptyArr"].Type).To(Equal("array"))
		Expect(schema.Properties["emptyArr"].Items.Schema.XIntOrString).To(BeTrue())
	})
})
