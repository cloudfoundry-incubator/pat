package experiment

import (
	. "github.com/julz/pat/benchmarker"
	"github.com/julz/pat/experiments"
	"time"
	"strings"
)

type SampleType int

const (
	ResultSample SampleType = iota
	WorkerSample
	ErrorSample
	OtherSample
)

type Command struct {
	Count     int64
	Average   time.Duration
	TotalTime time.Duration
	LastTime  time.Duration
	WorstTime time.Duration
}

type Sample struct {
	Commands     map[string]Command
	Average      time.Duration
	TotalTime    time.Duration
	Total        int64
	TotalErrors  int
	TotalWorkers int
	LastResult   time.Duration
	LastError    error
	WorstResult  time.Duration
	WallTime     time.Duration
	Type         SampleType
}

type RunningExperiment struct {
	iteration <-chan IterationResult
	benchmark <-chan BenchmarkResult
	errors    <-chan error
	workers   <-chan int
	samples   chan<- *Sample
	quit      <-chan bool
}

func Run(concurrency int, iterations int, interval int, stop int, workload string, tracker func(chan *Sample)) error {
	iteration := make(chan IterationResult)
	benchmark := make(chan BenchmarkResult)
	errors := make(chan error)
	workers := make(chan int)
	samples := make(chan *Sample)
	quit := make(chan bool)
	go Track(iteration, samples, benchmark, errors, workers, quit)
	go tracker(samples)

	worker := NewWorker()

	operations := strings.Split(workload, ",")
	if len(operations) == 1 && operations[0] == "" {
		operations[0] = "push"
	}

	for _, operation := range operations {
		if operation == "push" {
			worker.AddExperiment(operation, experiments.Dummy) //(dan) change to push later
		} else {
			worker.AddExperiment(operation, experiments.Dummy)
		}
	}

	Execute(RepeatEveryUntil(interval, stop, func() {
		ExecuteConcurrently(concurrency, Repeat(iterations, Counted(workers, TimeWorker(iteration, benchmark, errors, worker, operations))))
	}, quit))
	time.Sleep(1 * time.Second) //(dan) until we drain the channels, add a simple sleep. Print can close to fast and mess up terminal colors
	quit <- true
	return nil
}

func Track(iteration <-chan IterationResult, samplesOut chan<- *Sample, benchmark <-chan BenchmarkResult, errors <-chan error, workers <-chan int, quit <-chan bool) {
	ex := &RunningExperiment{iteration, benchmark, errors, workers, samplesOut, quit}
	ex.run()
}

func (ex *RunningExperiment) run() {
	commands := make(map[string]Command)
	var iterations int64
	var totalTime time.Duration
	var avg time.Duration
	var lastError error
	var lastResult time.Duration
	var totalErrors int
	var workers int
	var worstResult time.Duration
	startTime := time.Now()

	for {
		sampleType := OtherSample
		select {
		case iteration := <-ex.iteration:
			sampleType = ResultSample
			iterations = iterations + 1
			totalTime = totalTime + iteration.Duration
			avg = time.Duration(totalTime.Nanoseconds() / iterations)
			lastResult = iteration.Duration
			if iteration.Duration > worstResult {
				worstResult = iteration.Duration
			}
		case benchmark := <-ex.benchmark:
			cmd := commands[benchmark.Command]
			cmd.Count = cmd.Count + 1
			cmd.TotalTime = cmd.TotalTime + benchmark.Duration
			cmd.LastTime = benchmark.Duration
			cmd.Average = time.Duration(cmd.TotalTime.Nanoseconds() / cmd.Count)
			if benchmark.Duration > cmd.WorstTime {
				cmd.WorstTime = benchmark.Duration
			}

			commands[benchmark.Command] = cmd
		case e := <-ex.errors:
			lastError = e
			totalErrors = totalErrors + 1
		case w := <-ex.workers:
			workers = workers + w
		case <-ex.quit:
			close(ex.samples)
			return // FIXME(jz) maybe we need to drain the errors and results channels here?
		}

		ex.samples <- &Sample{commands, avg, totalTime, iterations, totalErrors, workers, lastResult, lastError, worstResult, time.Now().Sub(startTime), sampleType}
	}
}
