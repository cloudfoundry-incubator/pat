package pat

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestPat(t *testing.T) {
	RegisterFailHandler(Fail)
	ServeWithArgs("/tmp/pats-acceptance-test-runs")
	RunSpecs(t, "Pat Suite")
}
