package experiment

import (
	. "github.com/julz/pat/benchmarker"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"time"
)

var _ = Describe("Experiment", func() {
	It("Calculates the running average", func() {
		iteration := make(chan IterationResult)
		results := make(chan BenchmarkResult)
		errors := make(chan error)
		workers := make(chan int)
		quit := make(chan bool)

		samples := make(chan *Sample)
		go Track(iteration, samples, results, errors, workers, quit)
		go func() { iteration <- IterationResult{2 * time.Second} }()
		go func() { iteration <- IterationResult{4 * time.Second} }()
		go func() { iteration <- IterationResult{6 * time.Second} }()

		Ω((<-samples).Average).Should(Equal(2 * time.Second))
		Ω((<-samples).Average).Should(Equal(3 * time.Second))
		Ω((<-samples).Average).Should(Equal(4 * time.Second))
	})

	PIt("Worst, errors, workers etc.", func() {})
})
