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

		samples := make(chan *Sample)
		go Track(samples, results, errors, workers, quit)
		go func() { results <- 2 * time.Second }()
		go func() { results <- 4 * time.Second }()
		go func() { results <- 6 * time.Second }()

		Ω((<-samples).Average).Should(Equal(2 * time.Second))
		Ω((<-samples).Average).Should(Equal(3 * time.Second))
		Ω((<-samples).Average).Should(Equal(4 * time.Second))
	})

	PIt("Worst, errors, workers etc.", func() {})
})
