package benchmarker

import (
	"errors"
	"time"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Benchmarker", func() {
	Describe("#Time", func() {
		It("times an arbitrary function", func() {
			time, _ := Time(func() error { time.Sleep(2 * time.Second); return nil })
			Ω(time.Seconds()).Should(BeNumerically("~", 2, 0.5))
		})
	})

	Describe("TimedWithWorker", func() {
		It("sends the timing information retrieved from a worker to a channel", func() {
			ch := make(chan IterationResult)
			result := make(chan time.Duration)
			go func(result chan time.Duration) {
				defer close(ch)
				for t := range ch {
					result <- t.Duration
				}
			}(result)

			TimedWithWorker(ch, &DummyWorker{}, "three")()
			Ω((<-result).Seconds()).Should(BeNumerically("==", 3))
		})
	})

	Describe("LocalWorker", func() {
		It("Sets a function by name", func() {
			worker := NewWorker()
			worker.AddExperiment("foo", func() error { time.Sleep(1 * time.Second); return nil })
			Ω(worker.Experiments["foo"]).ShouldNot(BeNil())
		})

		Describe("When a single experiment is provided", func() {
			It("Times a function by name", func() {
				worker := NewWorker().AddExperiment("foo", func() error { time.Sleep(1 * time.Second); return nil })
				result := worker.Time("foo")
				Ω(result.Duration.Seconds()).Should(BeNumerically("~", 1, 0.1))
			})

			It("Sets the function command name in the response struct", func() {
				worker := NewWorker().AddExperiment("foo", func() error { time.Sleep(1 * time.Second); return nil })
				result := worker.Time("foo")
				Ω(result.Steps[0].Command).Should(Equal("foo"))
			})

			It("Returns any errors", func() {
				worker := NewWorker().AddExperiment("foo", func() error { return errors.New("Foo") })
				result := worker.Time("foo")
				Ω(result.Error).Should(HaveOccurred())
			})
		})

		Describe("When multiple steps are provided separated by commas", func() {
			var worker *LocalWorker
			var result IterationResult

			BeforeEach(func() {
				worker = NewWorker().AddExperiment("foo", func() error { time.Sleep(1 * time.Second); return nil })
				worker.AddExperiment("bar", func() error { time.Sleep(1 * time.Second); return nil })
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
			var worker *LocalWorker
			var result IterationResult

			BeforeEach(func() {
				worker = NewWorker().AddExperiment("foo", func() error { time.Sleep(1 * time.Second); return nil })
				worker.AddExperiment("bar", func() error { time.Sleep(1 * time.Second); return nil })
				worker.AddExperiment("errors", func() error { return errors.New("fishfinger system overflow") })
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
				Ω(result.Seconds()).Should(BeNumerically("~", 3, 1))
			})
		})

		Context("With 3 workers", func() {
			It("Runs in parallel", func() {
				result, _ := Time(func() error {
					ExecuteConcurrently(3, Repeat(3, func() { time.Sleep(2 * time.Second) }))
					return nil
				})
				Ω(result.Seconds()).Should(BeNumerically("~", 2, 1))
			})
		})
	})
})

type DummyWorker struct{}

func (*DummyWorker) Time(experiment string) IterationResult {
	var result IterationResult
	if experiment == "three" {
		result.Duration = 3 * time.Second
		return result
	}
	result.Duration = 0 * time.Second
	return result
}
