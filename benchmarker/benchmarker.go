package benchmarker

import (
	"sync"
	"time"
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

func Counted(out chan<- int, fn func(map[string]interface{})) func(map[string]interface{}) {
	return func(workloadCtx map[string]interface{}) {
		out <- 1
		fn(workloadCtx)
		out <- -1
	}
}

func TimedWithWorker(out chan<- IterationResult, worker Worker, experiment string) func(map[string]interface{}) {
	return func(workloadCtx map[string]interface{}) {
		time := worker.Time(experiment, workloadCtx)
		out <- time
	}
}

func Once(fn func(map[string]interface{})) <-chan func(map[string]interface{}) {
	return Repeat(1, fn)
}

func RepeatEveryUntil(repeatInterval int, runTime int, fn func(map[string]interface{}), quit <-chan bool) <-chan func(map[string]interface{}) {
	if repeatInterval == 0 || runTime == 0 {
		return Once(fn)
	} else {
		ch := make(chan func(map[string]interface{}))
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

func Repeat(n int, fn func(map[string]interface{})) <-chan func(map[string]interface{}) {
	ch := make(chan func(map[string]interface{}))
	go func() {
		defer close(ch)
		for i := 0; i < n; i++ {
			ch <- fn
		}
	}()
	return ch
}

func Execute(tasks <-chan func(map[string]interface{}), workloadCtx map[string]interface{}) {
	for task := range tasks {		
		task(workloadCtx)
	}
}

func ExecuteConcurrently(workers int, tasks <-chan func(map[string]interface{}), workloadCtx map[string]interface{}) {
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		workloadCtx["workerIndex"] = i
		go func(t <-chan func(map[string]interface{}), ctx map[string]interface{}) {
			defer wg.Done()
			for task := range t {
				task(ctx)
			}
		}(tasks, clone(workloadCtx))
	}
	wg.Wait()
}

func clone(src map[string]interface{}) map[string]interface{} {
	var clone = make(map[string]interface{})
	for k, v := range src {
    	clone[k] = v
	}
	return clone
}