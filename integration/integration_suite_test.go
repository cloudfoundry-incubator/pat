package integration_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"github.com/cloudfoundry-incubator/pat/integration/test_helpers"

	"testing"
)

func TestIntegration(t *testing.T) {
	BeforeSuite(test_helpers.BuildExecutable)

	AfterSuite(func() {
		gexec.CleanupBuildArtifacts()
	})

	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Suite")
}
