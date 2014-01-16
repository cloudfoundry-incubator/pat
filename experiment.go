package pat

import (
	"time"
)

type Sample struct {
	Average      time.Duration
	TotalTime    time.Duration
	Total        int64
	TotalErrors  int
	TotalWorkers int
	LastResult   time.Duration
	LastError    error
	WorstResult  time.Duration
	WallTime     time.Time
}

type RunningExperiment struct {
	results chan time.Duration
	errors  chan error
	workers chan int
	samples chan *Sample
}

func Track(samples chan *Sample, results chan time.Duration, errors chan error, workers chan int) {
	ex := &RunningExperiment{results, errors, workers, samples}
	ex.run()
}

func (ex *RunningExperiment) run() {
	var n int64
	var totalTime time.Duration
	var avg time.Duration
	var lastError error
	var lastResult time.Duration
	var totalErrors int
	var workers int
	var worstResult time.Duration
	for {
		select {
		case result := <-ex.results:
			n = n + 1
			totalTime = totalTime + result
			avg = time.Duration(totalTime.Nanoseconds() / n)
			lastResult = result
			if result > worstResult {
				worstResult = result
			}
		case e := <-ex.errors:
			lastError = e
			totalErrors = totalErrors + 1
		case w := <-ex.workers:
			workers = workers + w
		}

		ex.samples <- &Sample{avg, totalTime, n, totalErrors, workers, lastResult, lastError, worstResult, time.Now()}
	}
}
