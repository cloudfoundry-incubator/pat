package workloads_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestWorkloads(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Workloads Suite")
}
