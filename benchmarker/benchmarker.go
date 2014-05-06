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
	Duration    time.Duration
	Steps       []StepResult
	Error       *EncodableError
}

func Time(experiment func() error) (result time.Duration, err error) {
	t0 := time.Now()	
	err = experiment()
	t1 := time.Now()
	return t1.Sub(t0), err
}

func Counted(out chan<- int, fn func(context.WorkloadContext)) func(context.WorkloadContext) {
	return func(workloadCtx context.WorkloadContext) {
		out <- 1
		fn(workloadCtx)
		out <- -1
	}
}

func TimedWithWorker(out chan<- IterationResult, worker Worker, experiment string) func(context.WorkloadContext) {
	return func(workloadCtx context.WorkloadContext) {
		time := worker.Time(experiment, workloadCtx)
		out <- time
	}
}

func Once(fn func(context.WorkloadContext)) <-chan func(context.WorkloadContext) {
	return Repeat(1, fn)
}

func RepeatEveryUntil(repeatInterval int, runTime int, fn func(context.WorkloadContext), quit <-chan bool) <-chan func(context.WorkloadContext) {
	if repeatInterval == 0 || runTime == 0 {
		return Once(fn)
	} else {
		ch := make(chan func(context.WorkloadContext))
		var tickerQuit *time.Ticker
		ticker := time.NewTicker(time.Duration(repeatInterval) * time.Second)
		if runTime > 0 {
			tickerQuit = time.NewTicker(time.Duration(runTime) * time.Second)
		}
		go func() {
			defer close(ch)
			ch <- fn
			for {
				select {
				case <-ticker.C:
					ch <- fn
				case <-quit:
					ticker.Stop()
					return
				case <-tickerQuit.C:
					ticker.Stop()
					return
				}
			}
		}()
		return ch
	}
}

func Repeat(n int, fn func(context.WorkloadContext)) <-chan func(context.WorkloadContext) {
	ch := make(chan func(context.WorkloadContext))
	go func() {
		defer close(ch)
		for i := 0; i < n; i++ {
			ch <- fn
		}
	}()
	return ch
}

func Execute(tasks <-chan func(context.WorkloadContext), workloadCtx context.WorkloadContext) {
	for task := range tasks {		
		task(workloadCtx)
	}
}

func ExecuteConcurrently(workers int, tasks <-chan func(context.WorkloadContext), workloadCtx context.WorkloadContext) {
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		workloadCtx.PutInt("workerIndex", i)
		go func(t <-chan func(context.WorkloadContext), ctx context.WorkloadContext) {
			defer wg.Done()
			for task := range t {
				task(ctx)
			}
		}(tasks, workloadCtx.Clone())
	}
	wg.Wait()
}
