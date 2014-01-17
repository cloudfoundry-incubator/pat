package benchmarker

import (
	"sync"
	"time"
)

type Worker interface {
	Time(experiment string) (time.Duration, error)
}

type LocalWorker struct {
	Experiments map[string]func() error
}

func NewWorker() *LocalWorker {
	return &LocalWorker{make(map[string]func() error)}
}

func (self *LocalWorker) withExperiment(name string, fn func() error) *LocalWorker {
	self.Experiments[name] = fn
	return self
}

func (self *LocalWorker) Time(experiment string) (time.Duration, error) {
	return Time(self.Experiments[experiment])
}

func Time(experiment func() error) (totalTime time.Duration, err error) {
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

func Timed(out chan<- time.Duration, errOut chan<- error, experiment func() error) func() {
	return TimedWithWorker(out, errOut, NewWorker().withExperiment("*", experiment), "*")
}

func TimedWithWorker(out chan<- time.Duration, errOut chan<- error, worker Worker, experiment string) func() {
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
