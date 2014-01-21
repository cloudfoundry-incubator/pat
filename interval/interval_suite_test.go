package interval_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestInterval(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Interval Suite")
}
