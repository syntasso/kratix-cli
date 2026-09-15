package promiseutils_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestPromiseUtils(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "PromiseUtils Suite")
}
