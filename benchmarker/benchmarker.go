package benchmarker

import (
	"sync"
	"time"
)

type BenchmarkResult struct {
	Command  string
	Duration time.Duration
}

type IterationResult struct {
	Duration time.Duration
}

type Worker interface {
	Time(experiment string) (BenchmarkResult, error)
}

type LocalWorker struct {
	Experiments map[string]func() error
}

func NewWorker() *LocalWorker {
	return &LocalWorker{make(map[string]func() error)}
}

func (self *LocalWorker) AddExperiment(name string, fn func() error) *LocalWorker {
	self.Experiments[name] = fn
	return self
}

func (self *LocalWorker) Time(experiment string) (BenchmarkResult, error) {
	benchmark, err := Time(self.Experiments[experiment])
	benchmark.Command = experiment
	return benchmark, err
}

func Time(experiment func() error) (benchmark BenchmarkResult, err error) {
	t0 := time.Now()
	err = experiment()
	t1 := time.Now()
	benchmark.Duration = t1.Sub(t0)
	return benchmark, err
}

func Counted(out chan<- int, fn func()) func() {
	return func() {
		out <- 1
		fn()
		out <- -1
	}
}

func TimeWorker(out chan<- IterationResult, bench chan<- BenchmarkResult, errOut chan<- error, worker Worker, operations []string) func() {
	return func() {
		tStart := time.Now()
		for _, operation := range operations {
			result, err := worker.Time(operation)

			if err == nil {
				bench <- result
			} else {
				errOut <- err
			}
		}

		iter := IterationResult{time.Now().Sub(tStart)}
		out <- iter
	}
}

func TimedWithWorker(out chan<- BenchmarkResult, errOut chan<- error, worker Worker, experiment string) func() {
	return func() {
		time, err := worker.Time(experiment)
		if err == nil {
			out <- time
		} else {
			errOut <- err
		}
	}
}

func Once(fn func()) <-chan func() {
	return Repeat(1, fn)
}

func RepeatEveryUntil(repeatInterval int, runTime int, fn func(), quit <-chan bool) <-chan func() {
	if repeatInterval == 0 || runTime == 0 {
		return Once(fn)
	} else {
		ch := make(chan func())
		var tickerQuit *time.Ticker
		ticker := time.NewTicker(time.Duration(repeatInterval) * time.Second)
		if runTime > 0 {
			tickerQuit = time.NewTicker(time.Duration(runTime) * time.Second)
		}
		go func() {
			ch <- fn
			defer close(ch)
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

func Repeat(n int, fn func()) <-chan func() {
	ch := make(chan func())
	go func() {
		defer close(ch)
		for i := 0; i < n; i++ {
			ch <- fn
		}
	}()
	return ch
}

func Execute(tasks <-chan func()) {
	for task := range tasks {
		task()
	}
}

func ExecuteConcurrently(workers int, tasks <-chan func()) {
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(t <-chan func()) {
			defer wg.Done()
			for task := range t {
				task()
			}
		}(tasks)
	}
	wg.Wait()
}
