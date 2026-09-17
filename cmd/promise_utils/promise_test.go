package promiseutils_test

import (
	promiseutils "github.com/syntasso/kratix-cli/cmd/promise_utils"
	"github.com/syntasso/kratix/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("MarshalPromiseWithoutStatus", func() {
	var promise v1alpha1.Promise

	BeforeEach(func() {
		promise = v1alpha1.Promise{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Promise",
				APIVersion: "platform.kratix.io/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "my-promise",
				Labels: map[string]string{
					"kratix.io/promise-version": "v0.0.1",
				},
			},
		}
	})

	When("the Promise has a zero-value status", func() {
		It("omits the status block entirely, rather than emitting zero-value fields", func() {
			out, err := promiseutils.MarshalPromiseWithoutStatus(promise)
			Expect(err).NotTo(HaveOccurred())

			var obj map[string]interface{}
			Expect(yaml.Unmarshal(out, &obj)).To(Succeed())
			Expect(obj).NotTo(HaveKey("status"))
		})
	})

	When("the Promise has a populated status", func() {
		BeforeEach(func() {
			promise.Status = v1alpha1.PromiseStatus{
				Status:             "Available",
				Workflows:          5,
				WorkflowsSucceeded: 3,
				WorkflowsFailed:    2,
			}
		})

		It("still omits the status block, since status is server-managed", func() {
			out, err := promiseutils.MarshalPromiseWithoutStatus(promise)
			Expect(err).NotTo(HaveOccurred())

			var obj map[string]interface{}
			Expect(yaml.Unmarshal(out, &obj)).To(Succeed())
			Expect(obj).NotTo(HaveKey("status"))
		})
	})

	It("preserves the rest of the Promise", func() {
		promise.Spec.DestinationSelectors = []v1alpha1.PromiseScheduling{
			{MatchLabels: map[string]string{"environment": "prod"}},
		}

		out, err := promiseutils.MarshalPromiseWithoutStatus(promise)
		Expect(err).NotTo(HaveOccurred())

		var roundTripped v1alpha1.Promise
		Expect(yaml.Unmarshal(out, &roundTripped)).To(Succeed())

		Expect(roundTripped.Name).To(Equal("my-promise"))
		Expect(roundTripped.Labels).To(HaveKeyWithValue("kratix.io/promise-version", "v0.0.1"))
		Expect(roundTripped.Spec.DestinationSelectors).To(Equal(promise.Spec.DestinationSelectors))
	})
})
