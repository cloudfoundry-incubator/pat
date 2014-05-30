package benchmarker

import (
	"time"

	"github.com/cloudfoundry-incubator/pat/context"
	. "github.com/cloudfoundry-incubator/pat/workloads"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Benchmarker", func() {

	workloadCtx := context.New()

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
			go Counted(ch, func(context.Context) {})(workloadCtx)
			Ω(<-ch).Should(Equal(+1))
			Ω(<-ch).Should(Equal(-1))
		})
	})

	Describe("Once", func() {
		It("repeats a function once", func() {
			called := 0
			Execute(Once(func(context.Context) { called = called + 1 }), workloadCtx)
			Ω(called).Should(Equal(1))
		})
	})

	Describe("Repeat", func() {
		It("repeats a function N times", func() {
			called := 0
			Execute(Repeat(3, func(context.Context) { called = called + 1 }), workloadCtx)
			Ω(called).Should(Equal(3))
		})
	})

	Describe("RepeatEveryUntil", func() {
		It("repeats a function every interval seconds", func() {
			start := time.Now()
			var end time.Time
			interval := 2
			Execute(RepeatEveryUntil(interval, 3, func(context.Context) { end = time.Now() }, nil), workloadCtx)
			elapsed := end.Sub(start)
			elapsed = (elapsed / time.Second)
			Ω(int(elapsed)).Should(Equal(interval))
		})

		It("repeats a function every interval seconds until stop seconds", func() {
			var total int = 0
			interval := 2
			stop := 11
			Execute(RepeatEveryUntil(interval, stop, func(context.Context) { total += 1 }, nil), workloadCtx)
			Ω(total).Should(Equal((stop / interval) + 1))
		})

		It("repeats a function every interval seconds and stops when channel quit is set to true", func() {
			quit := make(chan bool)
			var total int = 0
			interval := 2
			stop := 11
			quitTime := 5
			time.AfterFunc(time.Duration(quitTime)*time.Second, func() { quit <- true })
			Execute(RepeatEveryUntil(interval, stop, func(context.Context) { total += 1 }, quit), workloadCtx)
			Ω(total).Should(Equal((quitTime / interval) + 1))
		})

		It("runs a function once if interval = 0 or stop = 0", func() {
			var total int = 0
			interval := 0
			stop := 1
			Execute(RepeatEveryUntil(interval, stop, func(context.Context) { total += 1 }, nil), workloadCtx)
			Ω(total).Should(Equal(1))

			total = 0
			interval = 3
			stop = 0
			Execute(RepeatEveryUntil(interval, stop, func(context.Context) { total += 1 }, nil), workloadCtx)
			Ω(total).Should(Equal(1))
		})

		It("runs a function stop/interval + 1 times if s is a multiple of n", func() {
			var total int = 0
			interval := 2
			stop := 10
			Execute(RepeatEveryUntil(interval, stop, func(context.Context) { total += 1 }, nil), workloadCtx)
			Ω(total).Should(Equal((stop / interval) + 1))
		})
	})

	Describe("#ExecuteConcurrently", func() {
		Context("When a single one is pushed", func() {
			var (
				schedule chan int
				tasks    chan func(context.Context)
			)

			BeforeEach(func() {
				schedule = make(chan int)
				tasks = make(chan func(context.Context))
			})
			It("Creates a new goroutine that executes the tasks serially", func() {
				orderWasExecuted := 0
				go func() {
					orderWasQueued := make(chan int, 1)
					defer close(tasks)
					for i := 0; i < 10; i++ {
						orderWasQueued <- i
						tasks <- func(n context.Context) {
							defer GinkgoRecover()
							Ω(orderWasExecuted).Should(Equal(<-orderWasQueued))
							orderWasExecuted++
						}
					}
				}()
				go func() {
					defer close(schedule)
					schedule <- 1
				}()
				ExecuteConcurrently(schedule, tasks, workloadCtx)
			})

			It("Pushes an iterationIndex into the context map", func() {
				go func() {
					defer close(tasks)
					for n := 0; n < 2; n++ {
						tasks <- func(ctx context.Context) {
							var tmp = n
							defer GinkgoRecover()
							index, exists := ctx.GetInt("iterationIndex")
							Ω(exists).ShouldNot(BeFalse())
							Ω(index).Should(Equal(tmp))
							time.Sleep(1 * time.Second)
						}
					}
				}()
				go func() {
					defer close(schedule)
					schedule <- 1
				}()
				ExecuteConcurrently(schedule, tasks, workloadCtx)
			})
		})

		Context("When an event larger than one is pushed", func() {
			It("Creates mutlple new goroutines that execute the tasks concurrent", func() {
				schedule := make(chan int)
				tasks := make(chan func(context.Context))
				executedInOrder := 0
				orderWasExecuted := 0
				go func() {
					defer close(tasks)
					for i := 0; i < 10; i++ {
						tasks <- func(n context.Context) {
							if orderWasExecuted == i {
								executedInOrder++
							}
							orderWasExecuted++
						}
					}
				}()
				go func() {
					defer close(schedule)
					schedule <- 2
				}()
				ExecuteConcurrently(schedule, tasks, workloadCtx)
				Ω(executedInOrder).Should(BeNumerically("~", 5, 1))
			})
		})

		Context("When multiple ones are pushed", func() {
			It("Creates a goroutine each time an event is pushed that executes the tasks concurrently", func() {
				schedule := make(chan int)
				tasks := make(chan func(context.Context))
				executedInOrder := 0
				orderWasExecuted := 0
				go func() {
					defer close(tasks)
					for i := 0; i < 10; i++ {
						tasks <- func(n context.Context) {
							if orderWasExecuted == i {
								executedInOrder++
							}
							orderWasExecuted++
						}
					}
				}()
				go func() {
					defer close(schedule)
					schedule <- 1
					schedule <- 1
				}()
				ExecuteConcurrently(schedule, tasks, workloadCtx)
				Ω(executedInOrder).Should(BeNumerically("~", 5, 1))
			})
		})
	})
})

type DummyWorker struct{}

func (*DummyWorker) Time(experiment string, workloadCtx context.Context) IterationResult {
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
