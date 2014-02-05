package experiment

import (
	. "github.com/julz/pat/benchmarker"
	"math"
	"strings"
	"time"
)

type SampleType int

const (
	ResultSample SampleType = iota
	WorkerSample
	ThroughputSample
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

type Throughput struct {
	Total         float64
	Commands      map[string]float64
	TimedCommands map[float64]map[string]int
}

type Sample struct {
	Throughput   Throughput
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

type Experiment interface {
	GetGuid() string
	GetData() ([]*Sample, error)
}

type ExperimentConfiguration struct {
	Iterations  int
	Concurrency int
	Interval    int
	Stop        int
	Worker      Worker
	Workload    string
}

type RunnableExperiment struct {
	ExperimentConfiguration
	executerFactory func(iterationResults chan IterationResult, benchmarkResults chan BenchmarkResult, errors chan error, workers chan int, quit chan bool, end chan bool) Executable
	samplerFactory  func(iterationResults chan IterationResult, benchmarkResults chan BenchmarkResult, errors chan error, workers chan int, samples chan *Sample, quit chan bool, end chan bool) Samplable
}

type ExecutableExperiment struct {
	ExperimentConfiguration
	iteration chan IterationResult
	benchmark chan BenchmarkResult
	errors    chan error
	workers   chan int
	quit      chan bool
	end       chan bool
}

type SamplableExperiment struct {
	iteration chan IterationResult
	benchmark chan BenchmarkResult
	errors    chan error
	workers   chan int
	samples   chan *Sample
	ticks     <-chan float64
	quit      chan bool
}

type Executable interface {
	Execute()
}

type Samplable interface {
	Sample()
}

func NewExperimentConfiguration(iterations int, concurrency int, interval int, stop int, worker Worker, workload string) ExperimentConfiguration {
	return ExperimentConfiguration{iterations, concurrency, interval, stop, worker, workload}
}

func NewRunnableExperiment(config ExperimentConfiguration) *RunnableExperiment {
	return &RunnableExperiment{config, config.newExecutableExperiment, newRunningExperiment}
}

func (c ExperimentConfiguration) newExecutableExperiment(iterationResults chan IterationResult, benchmarkResults chan BenchmarkResult, errors chan error, workers chan int, quit chan bool, end chan bool) Executable {
	return &ExecutableExperiment{c, iterationResults, benchmarkResults, errors, workers, quit, end}
}

func newRunningExperiment(iterationResults chan IterationResult, benchmarkResults chan BenchmarkResult, errors chan error, workers chan int, samples chan *Sample, quit chan bool, end chan bool) Samplable {
	return &SamplableExperiment{iterationResults, benchmarkResults, errors, workers, samples, newTicker(end), quit}
}

func newTicker(end chan bool) <-chan float64 {
	t := make(chan float64)
	go func(end chan bool) {
		startTime := time.Now()
		ticker := time.NewTicker(1 * time.Second).C
		for {
			select {
			case curTime := <-ticker:
				t <- curTime.Sub(startTime).Seconds()
			case <-end:
				t <- time.Now().Sub(startTime).Seconds()
				return
			}
		}
	}(end)
	return t
}

func (config *RunnableExperiment) Run(tracker func(<-chan *Sample)) error {
	iteration := make(chan IterationResult)
	benchmark := make(chan BenchmarkResult)
	errors := make(chan error)
	workers := make(chan int)
	samples := make(chan *Sample)
	quit := make(chan bool)
	done := make(chan bool)
	end := make(chan bool)
	sampler := config.samplerFactory(iteration, benchmark, errors, workers, samples, quit, end)
	go sampler.Sample()
	go func(d chan bool) {
		tracker(samples)
		d <- true
	}(done)

	config.executerFactory(iteration, benchmark, errors, workers, quit, end).Execute()
	<-done
	return nil
}

func (ex *ExecutableExperiment) Execute() {
	operations := strings.Split(ex.Workload, ",")
	if len(operations) == 1 && operations[0] == "" {
		operations[0] = "push"
	}

	Execute(RepeatEveryUntil(ex.Interval, ex.Stop, func() {
		ExecuteConcurrently(ex.Concurrency, Repeat(ex.Iterations, Counted(ex.workers, TimeWorker(ex.iteration, ex.benchmark, ex.errors, ex.Worker, operations))))
	}, ex.quit))
	ex.end <- true
	time.Sleep(1 * time.Second)
	close(ex.iteration)
}

func (ex *SamplableExperiment) Sample() {
	commands := make(map[string]Command)
	throughput := Throughput{0, make(map[string]float64), make(map[float64]map[string]int)}
	var count float64
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
		case iteration, ok := <-ex.iteration:
			if !ok {
				close(ex.samples)
				return
			}
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

			inner, ok := throughput.TimedCommands[math.Floor(benchmark.StopTime.Sub(startTime).Seconds())]
			if !ok {
				inner = make(map[string]int)
				throughput.TimedCommands[math.Floor(benchmark.StopTime.Sub(startTime).Seconds())] = inner
			}
			inner[benchmark.Command]++

			commands[benchmark.Command] = cmd
		case e := <-ex.errors:
			lastError = e
			totalErrors = totalErrors + 1
		case w := <-ex.workers:
			workers = workers + w
		case seconds := <-ex.ticks:
			sampleType = ThroughputSample
			count = 0
			for key, _ := range commands {
				cmd := commands[key]
				throughput.Commands[key] = float64(cmd.Count) / float64(seconds)
				count += float64(cmd.Count)
			}

			throughput.Total = count / float64(seconds)
		}

		ex.samples <- &Sample{throughput, commands, avg, totalTime, iterations, totalErrors, workers, lastResult, lastError, worstResult, time.Now().Sub(startTime), sampleType}
	}
}
