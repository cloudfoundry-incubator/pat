package benchmarker

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestBenchmarker(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Benchmarker Suite")
}
