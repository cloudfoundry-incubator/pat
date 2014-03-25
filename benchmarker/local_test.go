package benchmarker

import (
	"errors"
	"time"
	. "github.com/cloudfoundry-community/pat/workloads"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("LocalWorker", func() {
	Describe("When a single experiment is provided", func() {
		It("Times a function by name", func() {
			worker := NewLocalWorker()
			worker.AddWorkloadStep(Step("foo", func() error { time.Sleep(1 * time.Second); return nil }, ""))
			result := worker.Time("foo")
			Ω(result.Duration.Seconds()).Should(BeNumerically("~", 1, 0.1))
		})

		It("Sets the function command name in the response struct", func() {
			worker := NewLocalWorker()
			worker.AddWorkloadStep(Step("foo", func() error { time.Sleep(1 * time.Second); return nil }, ""))
			result := worker.Time("foo")
			Ω(result.Steps[0].Command).Should(Equal("foo"))
		})

		It("Returns any errors", func() {
			worker := NewLocalWorker()
			worker.AddWorkloadStep(Step("foo", func() error { return errors.New("Foo") }, ""))
			result := worker.Time("foo")
			Ω(result.Error).Should(HaveOccurred())
		})

		It("Passes context to each step", func() {
			var context map[string]interface{}
			worker := NewLocalWorker()
			worker.AddWorkloadStep(StepWithContext("foo", func(ctx map[string]interface{}) error { context = ctx; ctx["a"] = 1; return nil }, ""))
			worker.AddWorkloadStep(StepWithContext("bar", func(ctx map[string]interface{}) error { ctx["a"] = ctx["a"].(int) + 2; return nil }, ""))
			worker.Time("foo")
			Ω(context).Should(HaveKey("a"))
		})
	})

	Describe("When multiple steps are provided separated by commas", func() {
		var result IterationResult
		var worker Worker

		BeforeEach(func() {
			worker = NewLocalWorker()
			worker.AddWorkloadStep(Step("foo", func() error { time.Sleep(1 * time.Second); return nil }, ""))
			worker.AddWorkloadStep(Step("bar", func() error { time.Sleep(1 * time.Second); return nil }, ""))
			result = worker.Time("foo,bar")
		})

		It("Reports the total time", func() {
			Ω(result.Duration.Seconds()).Should(BeNumerically("~", 2, 0.1))
		})

		It("Records each step seperately", func() {
			Ω(result.Steps).Should(HaveLen(2))
			Ω(result.Steps[0].Command).Should(Equal("foo"))
			Ω(result.Steps[1].Command).Should(Equal("bar"))
		})

		It("Times each step seperately", func() {
			Ω(result.Steps).Should(HaveLen(2))
			Ω(result.Steps[0].Duration.Seconds()).Should(BeNumerically("~", 1, 0.1))
			Ω(result.Steps[1].Duration.Seconds()).Should(BeNumerically("~", 1, 0.1))
		})
	})

	Describe("When a step returns an error", func() {
		var worker Worker
		var result IterationResult

		BeforeEach(func() {
			worker = NewLocalWorker()
			worker.AddWorkloadStep(Step("foo", func() error { time.Sleep(1 * time.Second); return nil }, ""))
			worker.AddWorkloadStep(Step("bar", func() error { time.Sleep(1 * time.Second); return nil }, ""))
			worker.AddWorkloadStep(Step("errors", func() error { return errors.New("fishfinger system overflow") }, ""))
			result = worker.Time("foo,errors,bar")
		})

		It("Records the error", func() {
			Ω(result.Error).Should(HaveOccurred())
		})

		It("Records all steps up to the error step", func() {
			Ω(result.Steps).Should(HaveLen(2))
			Ω(result.Steps[0].Command).Should(Equal("foo"))
			Ω(result.Steps[1].Command).Should(Equal("errors"))
		})

		It("Reports the time as the time up to the error", func() {
			Ω(result.Duration.Seconds()).Should(BeNumerically("~", 1, 0.1))
		})
	})
})
