package experiment

import (
	"time"
	. "github.com/julz/pat/benchmarker"
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
	Count      int64
	Throughput float64
	Average    time.Duration
	TotalTime  time.Duration
	LastTime   time.Duration
	WorstTime  time.Duration
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
	executerFactory func(iterationResults chan IterationResult, errors chan error, workers chan int, quit chan bool) Executable
	samplerFactory  func(iterationResults chan IterationResult, errors chan error, workers chan int, samples chan *Sample, quit chan bool) Samplable
}

type ExecutableExperiment struct {
	ExperimentConfiguration
	iteration chan IterationResult
	workers   chan int
	quit      chan bool
}

type SamplableExperiment struct {
	iteration chan IterationResult
	workers   chan int
	samples   chan *Sample
	ticks     <-chan int
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

func (c ExperimentConfiguration) newExecutableExperiment(iterationResults chan IterationResult, errors chan error, workers chan int, quit chan bool) Executable {
	return &ExecutableExperiment{c, iterationResults, workers, quit}
}

func newRunningExperiment(iterationResults chan IterationResult, errors chan error, workers chan int, samples chan *Sample, quit chan bool) Samplable {
	return &SamplableExperiment{iterationResults, workers, samples, newTicker(), quit}
}

func newTicker() <-chan int {
	t := make(chan int)
	go func() {
		seconds := 0
		for _ = range time.NewTicker(1 * time.Second).C {
			seconds = seconds + 1
			t <- seconds
		}
	}()
	return t
}

func (config *RunnableExperiment) Run(tracker func(<-chan *Sample)) error {
	iteration := make(chan IterationResult)
	errors := make(chan error)
	workers := make(chan int)
	samples := make(chan *Sample)
	quit := make(chan bool)
	done := make(chan bool)
	sampler := config.samplerFactory(iteration, errors, workers, samples, quit)
	go sampler.Sample()
	go func(d chan bool) {
		tracker(samples)
		d <- true
	}(done)

	config.executerFactory(iteration, errors, workers, quit).Execute()
	<-done
	return nil
}

func (ex *ExecutableExperiment) Execute() {
	Execute(RepeatEveryUntil(ex.Interval, ex.Stop, func() {
		ExecuteConcurrently(ex.Concurrency, Repeat(ex.Iterations, Counted(ex.workers, TimedWithWorker(ex.iteration, ex.Worker, ex.Workload))))
	}, ex.quit))

	close(ex.iteration)
}

func (ex *SamplableExperiment) Sample() {
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

			for _, step := range iteration.Steps {
				cmd := commands[step.Command]
				cmd.Count = cmd.Count + 1
				cmd.TotalTime = cmd.TotalTime + step.Duration
				cmd.LastTime = step.Duration
				cmd.Average = time.Duration(cmd.TotalTime.Nanoseconds() / cmd.Count)
				if step.Duration > cmd.WorstTime {
					cmd.WorstTime = step.Duration
				}

				commands[step.Command] = cmd
			}

			if iteration.Error != nil {
				lastError = iteration.Error
				totalErrors = totalErrors + 1
			}
		case w := <-ex.workers:
			workers = workers + w
		case seconds := <-ex.ticks:
			sampleType = ThroughputSample
			for key, _ := range commands {
				cmd := commands[key]
				cmd.Throughput = float64(cmd.Count) / float64(seconds)
				commands[key] = cmd
			}
		}

		ex.samples <- &Sample{commands, avg, totalTime, iterations, totalErrors, workers, lastResult, lastError, worstResult, time.Now().Sub(startTime), sampleType}
	}
}
