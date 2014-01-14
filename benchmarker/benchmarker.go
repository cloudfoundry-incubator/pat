package benchmarker

import (
	"sync"
	"time"
)

type Timer struct {
	Runs []time.Duration
}

func Time(experiment func()) (totalTime time.Duration) {
	t0 := time.Now()
	experiment()
	t1 := time.Now()
	return t1.Sub(t0)
}

func Benchmark(experiment func(*Timer)) (totalTime time.Duration) {
	timer := &Timer{make([]time.Duration, 0)}
	experiment(timer)
	return 0
}

func Timed(out chan<- time.Duration, experiment func()) func() {
	return func() {
		out <- Time(experiment)
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
