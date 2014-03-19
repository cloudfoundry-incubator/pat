package workloads

import (
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
})
