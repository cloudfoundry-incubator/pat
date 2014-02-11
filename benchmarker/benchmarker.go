package benchmarker

import (
	"strings"
	"sync"
	"time"
)

type StepResult struct {
	Command  string
	Duration time.Duration
}

type IterationResult struct {
	Duration time.Duration
	Steps    []StepResult
	Error    error
}

type Worker interface {
	Time(experiment string) IterationResult
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

func (self *LocalWorker) Time(experiment string) (result IterationResult) {
	experiments := strings.Split(experiment, ",")
	var start = time.Now()
	for _, e := range experiments {
		stepTime, err := Time(self.Experiments[e])
		result.Steps = append(result.Steps, StepResult{e, stepTime})
		if err != nil {
			result.Error = err
			break
		}
	}
	result.Duration = time.Now().Sub(start)
	return
}

func Time(experiment func() error) (result time.Duration, err error) {
	t0 := time.Now()
	err = experiment()
	t1 := time.Now()
	return t1.Sub(t0), err
}

func Counted(out chan<- int, fn func()) func() {
	return func() {
		out <- 1
		fn()
		out <- -1
	}
}

func TimedWithWorker(out chan<- IterationResult, worker Worker, experiment string) func() {
	return func() {
		time := worker.Time(experiment)
		out <- time
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
