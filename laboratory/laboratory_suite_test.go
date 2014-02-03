package laboratory

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestLaboratory(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Laboratory Suite")
}
