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
			Ω(time.Seconds()).Should(BeNumerically("~", 2, 0.5))
		})
	})

	Describe("Timed", func() {
		It("sends the timing information of a function to a channel", func() {
			ch := make(chan time.Duration)
			result := make(chan float64)
			go func(result chan float64) {
				defer close(ch)
				for t := range ch {
					result <- t.Seconds()
				}
			}(result)

			Timed(ch, nil, func() error { time.Sleep(1 * time.Second); return nil })()

			Ω(<-result).Should(BeNumerically("~", 1, 0.5))
		})
	})

	Describe("TimedWithWorker", func() {
		It("sends the timing information retrieved from a worker to a channel", func() {
			ch := make(chan time.Duration)
			result := make(chan time.Duration)
			go func(result chan time.Duration) {
				defer close(ch)
				for t := range ch {
					result <- t
				}
			}(result)

			TimedWithWorker(ch, nil, &DummyWorker{}, "three")()
			Ω((<-result).Seconds()).Should(BeNumerically("==", 3))
		})
	})

	Describe("LocalWorker", func() {
		It("Times a function by name", func() {
			worker := NewWorker().withExperiment("foo", func() error { time.Sleep(1 * time.Second); return nil })
			time, _ := worker.Time("foo")
			Ω(time.Seconds()).Should(BeNumerically("~", 1, 0.1))
		})

		It("Returns any errors", func() {
			worker := NewWorker().withExperiment("foo", func() error { return errors.New("Foo") })
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

	Describe("Repeat Concurrently", func() {
		Context("with 1 worker", func() {
			It("Runs in series", func() {
				time, _ := Time(func() error {
					ExecuteConcurrently(1, Repeat(3, func() { time.Sleep(1 * time.Second) }))
					return nil
				})
				Ω(time.Seconds()).Should(BeNumerically("~", 3, 1))
			})
		})

		Context("With 3 workers", func() {
			It("Runs in parallel", func() {
				time, _ := Time(func() error {
					ExecuteConcurrently(3, Repeat(3, func() { time.Sleep(2 * time.Second) }))
					return nil
				})
				Ω(time.Seconds()).Should(BeNumerically("~", 2, 1))
			})
		})
	})
})

type DummyWorker struct{}

func (*DummyWorker) Time(experiment string) (time.Duration, error) {
	if experiment == "three" {
		return 3 * time.Second, nil
	}
	return 0 * time.Second, nil
}
