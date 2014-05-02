package benchmarker

import (
	. "github.com/cloudfoundry-incubator/pat/workloads"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"time"
)

var _ = Describe("LocalWorker", func() {
	It("Sets a function by name", func() {
		worker := &defaultWorker{make(map[string]WorkloadStep)}
		worker.AddWorkloadStep(Step("foo", func() error { time.Sleep(1 * time.Second); return nil }, ""))
		Ω(worker.Experiments["foo"]).ShouldNot(BeNil())
	})

	It("Visits all of the added experiements", func() {
		worker := &defaultWorker{make(map[string]WorkloadStep)}
		experiements := []string{"foo", "bar", "barry"}
		index := 0

		worker.AddWorkloadStep(Step("foo", func() error { return nil }, ""))
		worker.AddWorkloadStep(Step("bar", func() error { return nil }, ""))
		worker.AddWorkloadStep(Step("barry", func() error { return nil }, ""))

		worker.Visit(func(workload WorkloadStep) {
			Ω(workload.Name).Should(Equal(experiements[index]))
			index++
		})

		Ω(index).Should(BeNumerically("==", len(experiements)))
	})

	Describe("When a single experiment is provided", func() {
		It("Validates a workload name", func() {
			worker := &defaultWorker{make(map[string]WorkloadStep)}
			worker.AddWorkloadStep(Step("foo", func() error { return nil }, ""))
			ok, err := worker.Validate("foo")
			Ω(err).Should(BeNil())
			Ω(ok).Should(BeTrue())
		})

		It("Rejects an invalid workload name", func() {
			worker := &defaultWorker{make(map[string]WorkloadStep)}
			worker.AddWorkloadStep(Step("foo", func() error { return nil }, ""))
			ok, err := worker.Validate("bar")
			Ω(err).ShouldNot(BeNil())
			Ω(err.Error()).Should(ContainSubstring("bar"))
			Ω(ok).Should(BeFalse())
		})
	})

	Describe("When multiple steps are provided separated by commas", func() {
		var worker *defaultWorker

		BeforeEach(func() {
			worker = &defaultWorker{make(map[string]WorkloadStep)}
			worker.AddWorkloadStep(Step("foo", func() error { time.Sleep(1 * time.Second); return nil }, ""))
			worker.AddWorkloadStep(Step("bar", func() error { time.Sleep(1 * time.Second); return nil }, ""))
		})

		It("Validates a workload list", func() {
			ok, err := worker.Validate("foo,foo,foo")
			Ω(err).Should(BeNil())
			Ω(ok).Should(BeTrue())
		})

		It("Rejects an invalid workload list", func() {
			ok, err := worker.Validate("foo,fake,foo")
			Ω(err).ShouldNot(BeNil())
			Ω(err.Error()).Should(ContainSubstring("fake"))
			Ω(ok).Should(BeFalse())
		})
	})
})
