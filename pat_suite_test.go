package pat

import (
  . "github.com/onsi/ginkgo"
  . "github.com/onsi/gomega"

  "testing"
)

func TestPat(t *testing.T) {
  RegisterFailHandler(Fail)
  Serve()
  RunSpecs(t, "Pat Suite")
}
