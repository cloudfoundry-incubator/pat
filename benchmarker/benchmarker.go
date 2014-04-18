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

func Counted(out chan<- int, fn func(int)) func(int) {
	return func(workerIndex int) {
		out <- 1
		fn(workerIndex)
		out <- -1
	}
}

func TimedWithWorker(out chan<- IterationResult, worker Worker, experiment string) func(int) {
	return func(workerIndex int) {
		time := worker.Time(experiment, workerIndex)
		out <- time
	}
}

func Once(fn func(int)) <-chan func(int) {
	return Repeat(1, fn)
}

func RepeatEveryUntil(repeatInterval int, runTime int, fn func(int), quit <-chan bool) <-chan func(int) {
	if repeatInterval == 0 || runTime == 0 {
		return Once(fn)
	} else {
		ch := make(chan func(int))
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

func Repeat(n int, fn func(int)) <-chan func(int) {
	ch := make(chan func(int))
	go func() {
		defer close(ch)
		for i := 0; i < n; i++ {
			ch <- fn
		}
	}()
	return ch
}

func Execute(tasks <-chan func(int)) {
	for task := range tasks {
		task(1)
	}
}

func ExecuteConcurrently(workers int, tasks <-chan func(int)) {
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(t <-chan func(int), n int) {
			defer wg.Done()
			for task := range t {
				task(n)
			}
		}(tasks, i)
	}
	wg.Wait()
}
