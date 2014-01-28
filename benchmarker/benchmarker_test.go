package benchmarker

import (
	"errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"time"
)

var _ = Describe("Benchmarker", func() {
	Describe("#Time", func() {
		It("times an arbitrary function", func() {
			time, _ := Time(func() error { time.Sleep(2 * time.Second); return nil })
			Ω(time.Duration.Seconds()).Should(BeNumerically("~", 2, 0.5))
		})
	})

	Describe("TimeWorker", func() {
		It("allows a worker to execute a set of operations to run and sends timing information for each command in the operation set to a channel", func() {
			chBench := make(chan BenchmarkResult)
			resultBench := make(chan BenchmarkResult)
			go func(resultBench chan BenchmarkResult) {
				defer close(chBench)
				for t := range chBench {
					resultBench <- t
				}
			}(resultBench)

			worker := NewWorker()
			worker.AddExperiment("one", func() error { time.Sleep(1 * time.Second); return nil })
			worker.AddExperiment("two", func() error { time.Sleep(1 * time.Second); return nil })
			operations := []string{"one", "two"}

			go TimeWorker(nil, resultBench, nil, worker, operations)()
			Ω((<-resultBench).Duration.Seconds()).Should(BeNumerically("~", 1, 0.5))
			Ω((<-resultBench).Duration.Seconds()).Should(BeNumerically("~", 1, 0.5))
		})

		It("times how long it takes a worker to issues all operations.", func() {
			ch := make(chan IterationResult)
			result := make(chan IterationResult)
			go func(result chan IterationResult) {
				defer close(ch)
				for t := range ch {
					result <- t
				}
			}(result)

			chBench := make(chan BenchmarkResult)
			resultBench := make(chan BenchmarkResult)
			go func(resultBench chan BenchmarkResult) {
				defer close(chBench)
				for t := range chBench {
					resultBench <- t
				}
			}(resultBench)

			worker := NewWorker()
			worker.AddExperiment("one", func() error { time.Sleep(1 * time.Second); return nil })
			worker.AddExperiment("two", func() error { time.Sleep(1 * time.Second); return nil })
			operations := []string{"one", "two"}

			go TimeWorker(result, resultBench, nil, worker, operations)()
			Ω((<-resultBench).Duration.Seconds()).Should(BeNumerically("~", 1, 0.5))
			Ω((<-resultBench).Duration.Seconds()).Should(BeNumerically("~", 1, 0.5))
			Ω((<-result).Duration.Seconds()).Should(BeNumerically("~", 2, 0.5))
		})
	})

	Describe("TimedWithWorker", func() {
		It("sends the timing information retrieved from a worker to a channel", func() {
			ch := make(chan BenchmarkResult)
			result := make(chan time.Duration)
			go func(result chan time.Duration) {
				defer close(ch)
				for t := range ch {
					result <- t.Duration
				}
			}(result)

			TimedWithWorker(ch, nil, &DummyWorker{}, "three")()
			Ω((<-result).Seconds()).Should(BeNumerically("==", 3))
		})
	})

	Describe("LocalWorker", func() {
		It("Sets a function by name", func() {
			worker := NewWorker()
			worker.AddExperiment("foo", func() error { time.Sleep(1 * time.Second); return nil })
			Ω(worker.Experiments["foo"]).ShouldNot(BeNil())
		})

		It("Times a function by name", func() {
			worker := NewWorker().AddExperiment("foo", func() error { time.Sleep(1 * time.Second); return nil })
			result, _ := worker.Time("foo")
			Ω(result.Duration.Seconds()).Should(BeNumerically("~", 1, 0.1))
		})

		It("Sets the function command name in the response struct", func() {
			worker := NewWorker().AddExperiment("foo", func() error { time.Sleep(1 * time.Second); return nil })
			result, _ := worker.Time("foo")
			Ω(result.Command).Should(Equal("foo"))
		})

		It("Returns any errors", func() {
			worker := NewWorker().AddExperiment("foo", func() error { return errors.New("Foo") })
			_, err := worker.Time("foo")
			Ω(err).Should(HaveOccurred())
		})
	})

	Describe("Counted", func() {
		It("Sends +1 when the function is called, and -1 when it ends", func() {
			ch := make(chan int)
			go Counted(ch, func() {})()
			Ω(<-ch).Should(Equal(+1))
			Ω(<-ch).Should(Equal(-1))
		})
	})

	Describe("Once", func() {
		It("repeats a function once", func() {
			called := 0
			Execute(Once(func() { called = called + 1 }))
			Ω(called).Should(Equal(1))
		})
	})

	Describe("Repeat", func() {
		It("repeats a function N times", func() {
			called := 0
			Execute(Repeat(3, func() { called = called + 1 }))
			Ω(called).Should(Equal(3))
		})
	})

	Describe("RepeatEveryUntil", func() {
		It("repeats a function at n seconds interval", func() {
			start := time.Now()
			var end time.Time
			n := 2
			Execute(RepeatEveryUntil(n, 3, func() { end = time.Now() }, nil))
			elapsed := end.Sub(start)
			elapsed = (elapsed / time.Second)
			Ω(int(elapsed)).Should(Equal(n))
		})

		It("repeats a function at n seconds interval and stops at s second", func() {
			var total int = 0
			n := 2
			s := 11
			Execute(RepeatEveryUntil(n, s, func() { total += 1 }, nil))
			Ω(total).Should(Equal((s / n) + 1))
		})

		It("repeats a function at n seconds interval and stops when channel quit is set to true", func() {
			quit := make(chan bool)
			var total int = 0
			n := 2
			s := 11
			stop := 5
			time.AfterFunc(time.Duration(stop)*time.Second, func() { quit <- true })
			Execute(RepeatEveryUntil(n, s, func() { total += 1 }, quit))
			Ω(total).Should(Equal((stop / n) + 1))
		})

		It("runs a function once if n = 0 or s = 0", func() {
			var total int = 0
			n := 0
			s := 1
			Execute(RepeatEveryUntil(n, s, func() { total += 1 }, nil))
			Ω(total).Should(Equal(1))

			total = 0
			n = 3
			s = 0
			Execute(RepeatEveryUntil(n, s, func() { total += 1 }, nil))
			Ω(total).Should(Equal(1))
		})
	})

	Describe("Repeat Concurrently", func() {
		Context("with 1 worker", func() {
			It("Runs in series", func() {
				result, _ := Time(func() error {
					ExecuteConcurrently(1, Repeat(3, func() { time.Sleep(1 * time.Second) }))
					return nil
				})
				Ω(result.Duration.Seconds()).Should(BeNumerically("~", 3, 1))
			})
		})

		Context("With 3 workers", func() {
			It("Runs in parallel", func() {
				result, _ := Time(func() error {
					ExecuteConcurrently(3, Repeat(3, func() { time.Sleep(2 * time.Second) }))
					return nil
				})
				Ω(result.Duration.Seconds()).Should(BeNumerically("~", 2, 1))
			})
		})
	})
})

type DummyWorker struct{}

func (*DummyWorker) Time(experiment string) (BenchmarkResult, error) {
	var result BenchmarkResult
	if experiment == "three" {
		result.Duration = 3 * time.Second
		return result, nil
	}
	result.Duration = 0 * time.Second
	return result, nil
}
