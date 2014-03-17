package experiment

import (
	"errors"
	"time"
	. "github.com/julz/pat/benchmarker"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ExperimentConfiguration and Sampler", func() {
	Describe("Running an Experiment and Sampling", func() {
		var (
			sampler      *DummySampler
			executor     *DummyExecutor
			config       *RunnableExperiment
			sampleFunc   func(*DummySampler)
			executorFunc func(*DummyExecutor)
			sample1      *Sample
			sample2      *Sample
			worker       Worker
		)

		BeforeEach(func() {
			sample1 = &Sample{}
			sample2 = &Sample{}
			worker = NewWorker()
			executorFactory := func(iterationResults chan IterationResult, errors chan error, workers chan int, quit chan bool) Executable {
				executor = &DummyExecutor{iterationResults, workers, errors, executorFunc}
				return executor
			}
			samplerFactory := func(maxIterations int, iterationResults chan IterationResult, errors chan error, workers chan int, samples chan *Sample, quit chan bool) Samplable {
				sampler = &DummySampler{samples, iterationResults, workers, errors, sampleFunc}
				return sampler
			}
			config = &RunnableExperiment{ExperimentConfiguration{5, 2, 1, 1, worker, "push"}, executorFactory, samplerFactory}
		})

		It("Sends Samples from Sampler to the passed tracker function", func() {
			executorFunc = func(e *DummyExecutor) {}
			sampleFunc = func(s *DummySampler) {
				defer close(s.samples)
				s.samples <- sample1
				s.samples <- sample2
			}

			got := make([]*Sample, 0)
			config.Run(func(samples <-chan *Sample) {
				for s := range samples {
					got = append(got, s)
				}
			})

			Ω(got).Should(HaveLen(2))
		})

		It("Sends IterationResults from Executor to Sampler", func() {
			executorFunc = func(e *DummyExecutor) {
				e.IterationResults <- IterationResult{}
				e.IterationResults <- IterationResult{}
				e.IterationResults <- IterationResult{}
				close(e.IterationResults)
			}

			got := make([]IterationResult, 0)
			sampleFunc = func(s *DummySampler) {
				defer close(s.samples)
				for r := range s.IterationResults {
					got = append(got, r)
				}
			}

			config.Run(func(samples <-chan *Sample) {
				for _ = range samples {
				}
			})
			Ω(got).Should(HaveLen(3))
		})

		It("Sends Worker events from Executor to the Sampler", func() {
			executorFunc = func(e *DummyExecutor) {
				e.Workers <- 2
				e.Workers <- -1
				close(e.Workers)
			}

			got := make([]int, 0)
			sampleFunc = func(s *DummySampler) {
				defer close(s.samples)
				for r := range s.Workers {
					got = append(got, r)
				}
			}

			config.Run(func(samples <-chan *Sample) {
				for _ = range samples {
				}
			})
			Ω(got).Should(Equal([]int{2, -1}))
		})

		It("Sends Error events from Executor to the Sampler", func() {
			executorFunc = func(e *DummyExecutor) {
				e.Errors <- errors.New("Foo")
				close(e.Errors)
			}

			got := make([]error, 0)
			sampleFunc = func(s *DummySampler) {
				defer close(s.samples)
				for r := range s.Errors {
					got = append(got, r)
				}
			}

			config.Run(func(samples <-chan *Sample) {
				for _ = range samples {
				}
			})
			Ω(got).Should(HaveLen(1))
			Ω(got[0].Error()).Should(Equal("Foo"))
		})
	})

	Describe("Executing", func() {
		PIt("Closes the iterationResults channel when the executorFunc has finished", func() {})
		PIt("Runs a given number of times", func() {})
		PIt("Uses the passed worker", func() {})
	})

	Describe("Sampling", func() {
		var (
			maxIterations int 
			iteration chan IterationResult
			workers   chan int
			quit      chan bool
			ticks     chan int
			samples   chan *Sample
		)

		BeforeEach(func() {
			maxIterations = 3
			iteration = make(chan IterationResult)
			workers = make(chan int)
			quit = make(chan bool)
			samples = make(chan *Sample)
			ticks = make(chan int)
			go (&SamplableExperiment{maxIterations, iteration, workers, samples, ticks, quit}).Sample()
		})

		It("Calculates the running average", func() {
			go func() { iteration <- IterationResult{2 * time.Second, nil, nil} }()
			go func() { iteration <- IterationResult{4 * time.Second, nil, nil} }()
			go func() { iteration <- IterationResult{6 * time.Second, nil, nil} }()

			Ω((<-samples).Average).Should(Equal(2 * time.Second))
			Ω((<-samples).Average).Should(Equal(3 * time.Second))
			Ω((<-samples).Average).Should(Equal(4 * time.Second))
		})

		It("Closes the samples channel when there are no more iterationResults", func() {
			go func() {
				iteration <- IterationResult{2 * time.Second, nil, nil}
				close(iteration)
			}()

			Ω((<-samples).Average).Should(Equal(2 * time.Second))
			Ω(samples).Should(BeClosed())
			return
		})

		It("Updates throughput when the timer ticks", func() {
			go func() { ticks <- 1 }()
			Eventually((<-samples).Type).Should(Equal(ThroughputSample))
		})

		It("Counts errors", func() {
			go func() {
				iteration <- IterationResult{0, nil, errors.New("fishfingers burnt")}
				iteration <- IterationResult{0, nil, errors.New("toast not buttered")}
			}()

			Ω((<-samples).TotalErrors).Should(Equal(1))
			Ω((<-samples).TotalErrors).Should(Equal(2))
		})

		It("Calculates the throughput for a command every tick", func() {
			go func() {
				iteration <- IterationResult{0, []StepResult{StepResult{Command: "push"}}, nil}
				iteration <- IterationResult{0, []StepResult{StepResult{Command: "push"}}, nil}
				ticks <- 2
				iteration <- IterationResult{0, []StepResult{StepResult{Command: "push"}}, nil}
				ticks <- 3
				iteration <- IterationResult{0, []StepResult{StepResult{Command: "push"}}, nil}
				ticks <- 6
			}()

			<-samples // ResultSample
			<-samples // ResultSample
			Ω((<-samples).Commands["push"].Throughput).Should(BeNumerically("==", 1))
			<-samples // ResultSample
			Ω((<-samples).Commands["push"].Throughput).Should(BeNumerically("==", 1))
			<-samples // ResultSample
			Ω((<-samples).Commands["push"].Throughput).Should(BeNumerically("==", 4.0/6.0))
		})
	})

	Describe("Sampling Percentile", func() {
		var (
			maxIterations int 	
			iteration chan IterationResult
			workers   chan int
			quit      chan bool
			ticks     chan int
			samples   chan *Sample
		)

		BeforeEach(func() {
			maxIterations = 21
			iteration = make(chan IterationResult)
			workers = make(chan int)
			quit = make(chan bool)
			samples = make(chan *Sample)
			ticks = make(chan int)
			go (&SamplableExperiment{maxIterations, iteration, workers, samples, ticks, quit}).Sample()
		})

		It("Calculates the 95th percentile", func() {
			for i := 1; i <= maxIterations; i++ {
				currentIteration := i
				go func() { iteration <- IterationResult{time.Duration(currentIteration) * time.Second, nil, nil} }()
			}
			for q := 1; q <= 20; q++ {
				Ω((<-samples).NinetyfifthPercentile).Should(Equal(time.Duration(q) * time.Second))
			}
			for q := 21; q <= 21; q++ {
				Ω((<-samples).NinetyfifthPercentile).Should(Equal(time.Duration(q-1) * time.Second))
			}
		})
	})
})

type DummySampler struct {
	samples          chan *Sample
	IterationResults chan IterationResult
	Workers          chan int
	Errors           chan error
	sampleFunc       func(*DummySampler)
}

type DummyExecutor struct {
	IterationResults chan IterationResult
	Workers          chan int
	Errors           chan error
	executorFunc     func(*DummyExecutor)
}

func (s *DummySampler) Sample() {
	s.sampleFunc(s)
}

func (e *DummyExecutor) Execute() {
	e.executorFunc(e)
}
