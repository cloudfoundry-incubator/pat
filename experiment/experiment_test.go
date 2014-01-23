package experiment

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"time"
)

var _ = Describe("Experiment", func() {
	It("Calculates the running average", func() {
		results := make(chan time.Duration)
		errors := make(chan error)
		workers := make(chan int)
		quit := make(chan bool)
		workload_total := 10

		samples := make(chan *Sample)
		go Track(samples, results, errors, workers, quit, workload_total)
		go func() { results <- 2 * time.Second }()
		go func() { results <- 4 * time.Second }()
		go func() { results <- 6 * time.Second }()

		Ω((<-samples).Average).Should(Equal(2 * time.Second))
		Ω((<-samples).Average).Should(Equal(3 * time.Second))
		Ω((<-samples).Average).Should(Equal(4 * time.Second))
	})

	PIt("Worst, errors, workers etc.", func() {})

	Describe("RunExperiment.run()", func() {
		It("Resets task total when a workload is finished", func() {
			results := make(chan time.Duration)
			errors := make(chan error)
			workers := make(chan int)
			quit := make(chan bool)
			workload_total := 3

			samples := make(chan *Sample)
			go Track(samples, results, errors, workers, quit, workload_total)
			go func() { results <- 1 * time.Second }()
			go func() { results <- 2 * time.Second }()
			go func() { results <- 3 * time.Second }()
			go func() { results <- 4 * time.Second }()

			Ω((<-samples).Total).Should(Equal(int64(1)))
			Ω((<-samples).Total).Should(Equal(int64(2)))
			Ω((<-samples).Total).Should(Equal(int64(3)))
			Ω((<-samples).Total).Should(Equal(int64(1)))
		})
	})
})
