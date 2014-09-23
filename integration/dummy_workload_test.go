package integration_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry-incubator/pat/integration/test_helpers"
)

var _ = Describe("Dummy Integration", func() {
	var (
		err     error
		tmpPath string
	)

	Context("running dummy test", func() {
		BeforeEach(func() {
			tmpPath, err = ioutil.TempDir("", "PAT")
			Ω(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			err := os.RemoveAll(tmpPath)
			Ω(err).ToNot(HaveOccurred())
		})

		It("runs a dummy workload and finishes with 0 exit", func() {
			session := RunPAT("-iterations=3", "-workload=dummy", "-silent", fmt.Sprintf("-csv-dir=%s", tmpPath))
			Ω(session.Wait(20).ExitCode()).Should(Equal(0), "exit code is not 0")

			fileInfos, err := ioutil.ReadDir(tmpPath)
			Ω(err).ToNot(HaveOccurred())
			checkForCSVFile(fileInfos)
		})
	})
})

func checkForCSVFile(fileInfos []os.FileInfo) {
	found := false
	for _, fileInfo := range fileInfos {
		if strings.Contains(fileInfo.Name(), ".csv") {
			found = true
		}
	}
	Ω(found).To(BeTrue())
}
