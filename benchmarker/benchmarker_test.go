package benchmarker

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"time"
)

var _ = Describe("Benchmarker", func() {
	Describe("#Time", func() {
		It("times an arbitrary function", func() {
			time := Time(func() { time.Sleep(2 * time.Second) })
			Ω(time.Seconds()).Should(BeNumerically("~", 2, 0.5))
		})
	})

	Describe("Timed", func() {
		It("sends the timing information of a function to a channel", func() {
			ch := make(chan time.Duration)
			var seconds float64
			go func() {
				defer close(ch)
				for t := range ch {
					seconds = t.Seconds()
				}
			}()

			Timed(ch, func() { time.Sleep(1 * time.Second) })()

			Ω(seconds).Should(BeNumerically("~", 1, 0.5))
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
				time := Time(func() {
					ExecuteConcurrently(1, Repeat(3, func() { time.Sleep(1 * time.Second) }))
				})
				Ω(time.Seconds()).Should(BeNumerically("~", 3, 1))
			})
		})

		Context("With 3 workers", func() {
			It("Runs in parallel", func() {
				time := Time(func() {
					ExecuteConcurrently(3, Repeat(3, func() { time.Sleep(2 * time.Second) }))
				})
				Ω(time.Seconds()).Should(BeNumerically("~", 2, 1))
			})
		})
	})
})
