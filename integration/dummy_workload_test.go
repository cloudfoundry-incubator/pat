package integration_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry-incubator/pat/integration/test_helpers"
)

var _ = Describe("Dummy Integration", func() {
	var (
		err            error
		tmpPath        string
		newAppPath     string
		newAppManifest string
	)

	Context("running dummy workload with -silent", func() {
		BeforeEach(func() {
			tmpPath, err = ioutil.TempDir("", "PAT")
			Ω(err).ToNot(HaveOccurred())

			newAppPath = filepath.Join("assets", "hello-world")
			newAppManifest = filepath.Join("assets", "manifests", "hello-world-manifest.yml")
		})

		AfterEach(func() {
			err := os.RemoveAll(tmpPath)
			Ω(err).ToNot(HaveOccurred())
		})

		It("and finishes with 0 exit", func() {
			csvDir := fmt.Sprintf("-csv-dir=%s", tmpPath)
			session := RunPAT("-iterations=3", "-workload=dummy", "-silent", csvDir)
			Ω(session.Wait(20).ExitCode()).Should(Equal(0), "exit code is not 0")

			fileInfos, err := ioutil.ReadDir(tmpPath)
			Ω(err).ToNot(HaveOccurred())
			checkForCSVFile(fileInfos)
		})

		Context("and -app specified", func() {
			It("and finishes with 0 exit", func() {
				csvDir := fmt.Sprintf("-csv-dir=%s", tmpPath)
				appPath := fmt.Sprintf("-app=%s", newAppPath)
				session := RunPAT("-iterations=1", "-workload=dummy", "-silent", appPath, csvDir)
				Ω(session.Wait(20).ExitCode()).Should(Equal(0), "exit code is not 0")

				fileInfos, err := ioutil.ReadDir(tmpPath)
				Ω(err).ToNot(HaveOccurred())
				checkForCSVFile(fileInfos)
			})

			It("and -app:manifest specified and finishes with 0 exit", func() {
				csvDir := fmt.Sprintf("-csv-dir=%s", tmpPath)
				appPath := fmt.Sprintf("-app=%s", newAppPath)
				manifestPath := fmt.Sprintf("-app:manifest=%s", newAppManifest)
				session := RunPAT("-iterations=1", "-workload=dummy", "-silent", appPath, manifestPath, csvDir)
				Ω(session.Wait(20).ExitCode()).Should(Equal(0), "exit code is not 0")

				fileInfos, err := ioutil.ReadDir(tmpPath)
				Ω(err).ToNot(HaveOccurred())
				checkForCSVFile(fileInfos)
			})
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
