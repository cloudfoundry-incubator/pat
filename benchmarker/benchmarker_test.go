package benchmarker

import (
	. "github.com/cloudfoundry-community/pat/workloads"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"time"
)

var _ = Describe("Benchmarker", func() {
	workloadCtx := make(map[string]interface{})

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

			TimedWithWorker(ch, &DummyWorker{}, "three")(workloadCtx)
			Ω((<-result).Seconds()).Should(BeNumerically("==", 3))
		})
	})

	Describe("Counted", func() {
		It("Sends +1 when the function is called, and -1 when it ends", func() {
			ch := make(chan int)
			go Counted(ch, func(map[string]interface{}) {})(workloadCtx)
			Ω(<-ch).Should(Equal(+1))
			Ω(<-ch).Should(Equal(-1))
		})
	})

	Describe("Once", func() {
		It("repeats a function once", func() {
			called := 0
			Execute(Once(func(map[string]interface{}) { called = called + 1 }), workloadCtx)
			Ω(called).Should(Equal(1))
		})
	})

	Describe("Repeat", func() {
		It("repeats a function N times", func() {
			called := 0
			Execute(Repeat(3, func(map[string]interface{}) { called = called + 1 }), workloadCtx)
			Ω(called).Should(Equal(3))
		})
	})

	Describe("RepeatEveryUntil", func() {
		It("repeats a function at n seconds interval", func() {
			start := time.Now()
			var end time.Time
			n := 2
			Execute(RepeatEveryUntil(n, 3, func(map[string]interface{}) { end = time.Now() }, nil), workloadCtx)
			elapsed := end.Sub(start)
			elapsed = (elapsed / time.Second)
			Ω(int(elapsed)).Should(Equal(n))
		})

		It("repeats a function at n seconds interval and stops at s second", func() {
			var total int = 0
			n := 2
			s := 11
			Execute(RepeatEveryUntil(n, s, func(map[string]interface{}) { total += 1 }, nil), workloadCtx)
			Ω(total).Should(Equal((s / n) + 1))
		})

		It("repeats a function at n seconds interval and stops when channel quit is set to true", func() {
			quit := make(chan bool)
			var total int = 0
			n := 2
			s := 11
			stop := 5
			time.AfterFunc(time.Duration(stop)*time.Second, func() { quit <- true })
			Execute(RepeatEveryUntil(n, s, func(map[string]interface{}) { total += 1 }, quit), workloadCtx)
			Ω(total).Should(Equal((stop / n) + 1))
		})

		It("runs a function once if n = 0 or s = 0", func() {
			var total int = 0
			n := 0
			s := 1
			Execute(RepeatEveryUntil(n, s, func(map[string]interface{}) { total += 1 }, nil), workloadCtx)
			Ω(total).Should(Equal(1))

			total = 0
			n = 3
			s = 0
			Execute(RepeatEveryUntil(n, s, func(map[string]interface{}) { total += 1 }, nil), workloadCtx)
			Ω(total).Should(Equal(1))
		})
	})

	Describe("Execute", func() {
		AfterEach(func() {
			workloadCtx["cfUsername"] = ""
		})

		It("passes workloadCtx map to the test functions", func() {
			var cfUsername = ""

			workloadCtx["cfUsername"] = "user1,user2"

 			var fn = func(ctx map[string]interface{}) { 
	 			cfUsername = ctx["cfUsername"].(string)
 			}

 			Execute(RepeatEveryUntil(0, 0, func(map[string]interface{}) { ExecuteConcurrently(1, Repeat(1, fn), workloadCtx) }, nil), workloadCtx)
			 			
 			Ω(cfUsername).Should(Equal("user1,user2"))
		})
	})

	Describe("ExecuteConcurrently", func() {
		AfterEach(func() {
			workloadCtx["cfTarget"] = ""
		})

		It("passes workloadCtx map to the test functions", func() {			
 			var cfTarget = ""
			workloadCtx["cfTarget"] = "http://localhost/"
 			var fn = func(ctx map[string]interface{}) { 
	 			cfTarget = ctx["cfTarget"].(string)
 			}
			
 			ExecuteConcurrently(1, Repeat(1, fn), workloadCtx)
 			Ω(cfTarget).Should(Equal("http://localhost/"))
		})
	})

	Describe("Repeat Concurrently", func() {
		Context("with 1 worker", func() {
			It("Runs in series", func() {
				result, _ := Time(func() error {
					ExecuteConcurrently(1, Repeat(3, func(map[string]interface{}) { time.Sleep(1 * time.Second) }), workloadCtx)
					return nil
				})
				Ω(result.Seconds()).Should(BeNumerically("~", 3, 1))
			})
		})

		Context("With 3 workers", func() {
			It("Runs in parallel", func() {
				result, _ := Time(func() error {
					ExecuteConcurrently(3, Repeat(3, func(map[string]interface{}) { time.Sleep(2 * time.Second) }), workloadCtx)
					return nil
				})
				Ω(result.Seconds()).Should(BeNumerically("~", 2, 1))
			})
		})
	})

})

type DummyWorker struct{}

func (*DummyWorker) Time(experiment string, workloadCtx map[string]interface{}) IterationResult {
	var result IterationResult
	if experiment == "three" {
		result.Duration = 3 * time.Second
		return result
	}
	result.Duration = 0 * time.Second
	return result
}

func (d *DummyWorker) AddWorkloadStep(workload WorkloadStep) {
	return
}

func (d *DummyWorker) Visit(fn func(WorkloadStep)) {
}

func (d *DummyWorker) Validate(name string) (result bool, err error) {
	return
}
