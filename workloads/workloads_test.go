package workloads

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"

	"github.com/cloudfoundry-incubator/pat/context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type dummyWorkloadReceiver struct{ Workloads []WorkloadStep }

func (self *dummyWorkloadReceiver) AddWorkloadStep(workload WorkloadStep) {
	self.Workloads = append(self.Workloads, workload)
}

var _ = Describe("Workloads", func() {
	It("sets the full set of workloads in a worker", func() {

		testList := []WorkloadStep{
			Step("foo", func() error { return nil }, "a"),
			Step("bar", func() error { return nil }, "b"),
			Step("barry", func() error { return nil }, "c"),
			Step("fred", func() error { return nil }, "d"),
		}
		workloadList := WorkloadList{testList}

		worker := &dummyWorkloadReceiver{}
		workloadList.DescribeWorkloads(worker)

		for i, w := range testList {
			Ω(worker.Workloads[i].Name).Should(Equal(w.Name))
			Ω(worker.Workloads[i].Description).Should(Equal(w.Description))
		}
		Ω(worker.Workloads).Should(HaveLen(4))
	})

	Describe("#PopulateAppContext", func() {
		BeforeEach(func() {
			tmpDir := filepath.Join(os.TempDir(), "patTest")
			os.Chdir(tmpDir)
		})

		It("inserts the app path and manifest path into the context", func() {
			ctx := context.New()
			PopulateAppContext("foo", "manifest.yml", ctx)
			appPath, ok := ctx.GetString("app")
			Ω(Expect(ok).To(BeTrue()))
			Ω(Expect(appPath).To(Equal("foo")))
			manifestPath, ok := ctx.GetString("app:manifest")
			Ω(Expect(ok).To(BeTrue()))
			Ω(Expect(manifestPath).To(Equal("manifest.yml")))
		})

		It("inserts the app path and manifest path into the context", func() {
			ctx := context.New()
			if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
				PopulateAppContext("~/foo", "~/manifest.yml", ctx)

				usr, _ := user.Current()
				appPathActual := fmt.Sprintf("%s/foo", usr.HomeDir)
				manifestPathActual := fmt.Sprintf("%s/manifest.yml", usr.HomeDir)

				appPath, ok := ctx.GetString("app")
				Ω(Expect(ok).To(BeTrue()))
				Ω(Expect(appPath).To(Equal(appPathActual)))
				manifestPath, ok := ctx.GetString("app:manifest")
				Ω(Expect(ok).To(BeTrue()))
				Ω(Expect(manifestPath).To(Equal(manifestPathActual)))
			} else if runtime.GOOS == "windows" {
				//TODO: figure out how windows works
			}
		})

	})
})
