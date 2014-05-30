package benchmarker

import (
	"sync"
	"time"

	"github.com/cloudfoundry-incubator/pat/context"
)

type StepResult struct {
	Command  string
	Duration time.Duration
}

type IterationResult struct {
	Duration time.Duration
	Steps    []StepResult
	Error    *EncodableError
}

func Time(experiment func() error) (result time.Duration, err error) {
	t0 := time.Now()
	err = experiment()
	t1 := time.Now()
	return t1.Sub(t0), err
}

func Counted(out chan<- int, fn func(context.Context)) func(context.Context) {
	return func(workloadCtx context.Context) {
		out <- 1
		fn(workloadCtx)
		out <- -1
	}
}

func TimedWithWorker(out chan<- IterationResult, worker Worker, experiment string) func(context.Context) {
	return func(workloadCtx context.Context) {
		time := worker.Time(experiment, workloadCtx)
		out <- time
	}
}

func Once(fn func(context.Context)) <-chan func(context.Context) {
	return Repeat(1, fn)
}

func RepeatEveryUntil(repeatInterval int, runTime int, fn func(context.Context), quit <-chan bool) <-chan func(context.Context) {
	if repeatInterval == 0 || runTime == 0 {
		return Once(fn)
	} else {
		ch := make(chan func(context.Context))
		ticker := time.NewTicker(time.Duration(repeatInterval) * time.Second)
		go func() {
			defer close(ch)
			ch <- fn
			repeats := 0
			for {
				select {
				case <-ticker.C:
					repeats++
					if repeats*repeatInterval > runTime {
						ticker.Stop()
						return
					}
					ch <- fn
				case <-quit:
					ticker.Stop()
					return
				}
			}
		}()
		return ch
	}
}

func Repeat(n int, fn func(context.Context)) <-chan func(context.Context) {
	ch := make(chan func(context.Context))
	go func() {
		defer close(ch)
		for i := 0; i < n; i++ {
			ch <- fn
		}
	}()
	return ch
}

func Execute(tasks <-chan func(context.Context), workloadCtx context.Context) {
	for task := range tasks {
		task(workloadCtx)
	}
}

func ExecuteConcurrently(schedule <-chan int, tasks <-chan func(context.Context), workloadCtx context.Context) {
	var wg sync.WaitGroup
	indexCounter := 0

	for increment := range schedule {

		for i := 0; i < increment; i++ {
			wg.Add(1)
			go func(t <-chan func(context.Context), ctx context.Context) {
				defer wg.Done()
				for task := range t {
					ctx.PutInt("iterationIndex", indexCounter)
					indexCounter++
					task(ctx)
				}
			}(tasks, workloadCtx.Clone())
		}
	}
	wg.Wait()
}
