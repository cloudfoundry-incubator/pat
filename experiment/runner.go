package experiment

import (
	"math"
	"time"
	. "github.com/cloudfoundry-incubator/pat/benchmarker"
)

type SampleType int
type schedule chan int

const (
	ResultSample SampleType = iota
	WorkerSample
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
	Commands              map[string]Command
	Average               time.Duration
	TotalTime             time.Duration
	Total                 int64
	TotalErrors           int
	TotalWorkers          int
	LastResult            time.Duration
	LastError             error
	WorstResult           time.Duration
	NinetyfifthPercentile time.Duration
	WallTime              time.Duration
	Type                  SampleType
}

type Experiment interface {
	GetGuid() string
	GetData() ([]*Sample, error)
}

type ExperimentConfiguration struct {
	Iterations          int
	Concurrency         []int
	ConcurrencyStepTime time.Duration
	Interval            int
	Stop                int
	Worker              Worker
	Workload            string
	Note		    string
}

type RunnableExperiment struct {
	ExperimentConfiguration
	executerFactory func(iterationResults chan IterationResult, errors chan error, workers chan int, quit chan bool) Executable
	samplerFactory  func(iterations int, iterationResults chan IterationResult, errors chan error, workers chan int, samples chan *Sample, quit chan bool) Samplable
}

type ExecutableExperiment struct {
	ExperimentConfiguration
	iteration chan IterationResult
	workers   chan int
	quit      chan bool
	schedule  chan int
}

type SamplableExperiment struct {
	maxIterations int
	iteration     chan IterationResult
	workers       chan int
	samples       chan *Sample
	quit          chan bool
}

type Executable interface {
	Execute()
}

type Samplable interface {
	Sample()
}

func NewExperimentConfiguration(iterations int, concurrency []int, concurrencyStepTime time.Duration, interval int, stop int, worker Worker, workload string, note string) ExperimentConfiguration {
	return ExperimentConfiguration{iterations, concurrency, concurrencyStepTime, interval, stop, worker, workload, note}
}

func NewRunnableExperiment(config ExperimentConfiguration) *RunnableExperiment {
	return &RunnableExperiment{config, config.newExecutableExperiment, newRunningExperiment}
}

func (c ExperimentConfiguration) newExecutableExperiment(iterationResults chan IterationResult, errors chan error, workers chan int, quit chan bool) Executable {
	startingWorkers := c.Concurrency[0]
	totalWorkers := startingWorkers
	if len(c.Concurrency) > 1 {
		totalWorkers = c.Concurrency[1]
	}
	schedule := linearSchedule(startingWorkers, totalWorkers, c.ConcurrencyStepTime)
	return &ExecutableExperiment{c, iterationResults, workers, quit, schedule}
}

func newRunningExperiment(iterations int, iterationResults chan IterationResult, errors chan error, workers chan int, samples chan *Sample, quit chan bool) Samplable {
	return &SamplableExperiment{iterations, iterationResults, workers, samples, quit}
}

func (config *RunnableExperiment) GetExperimentConfiguration() ExperimentConfiguration {
	return config.ExperimentConfiguration
}

func (config *RunnableExperiment) Run(tracker func(<-chan *Sample)) error {
	iteration := make(chan IterationResult)
	errors := make(chan error)
	workers := make(chan int)
	samples := make(chan *Sample)
	quit := make(chan bool)
	done := make(chan bool)
	maxIterations := config.Iterations
	if config.Stop != 0 && config.Interval != 0 && config.Interval < config.Stop {
		maxIterations *= config.Stop / config.Interval
	}
	sampler := config.samplerFactory(maxIterations, iteration, errors, workers, samples, quit)
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
	Execute(RepeatEveryUntil(ex.Interval, ex.Stop, func(int) {
		ExecuteConcurrently(ex.schedule, Repeat(ex.Iterations, Counted(ex.workers, TimedWithWorker(ex.iteration, ex.Worker, ex.Workload))))
	}, ex.quit))

	close(ex.iteration)
}

func clone(src map[string]Command) map[string]Command {
	var clone = make(map[string]Command)
	for k, v := range src {
		clone[k] = v
	}
	return clone
}

func linearSchedule(startingWorkers int, totalWorkers int, concurrencyStepTime time.Duration) chan int {
	ch := make(chan int)
	go func() {
		defer close(ch)
		for i := 0; i < startingWorkers; i++ {
			ch <- 1
		}
		if concurrencyStepTime > 0 && startingWorkers < totalWorkers {
			tick := time.NewTicker(concurrencyStepTime)
			for _ = range tick.C {
				ch <- 1
				startingWorkers++
				if startingWorkers >= totalWorkers {
					tick.Stop()
					break
				}
			}
		}
	}()
	return ch
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
	var ninetyfifthPercentile time.Duration
	var percentileLength = int(math.Floor(float64(ex.maxIterations)*.05 + 0.95))
	var percentile = make([]time.Duration, percentileLength, percentileLength)
	var heartbeat = time.NewTicker(1 * time.Second)
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

			if lastResult > percentile[0] {
				percentile[0] = lastResult
				for i := 0; i < percentileLength-1 && lastResult > percentile[i+1]; i++ {
					percentile[i] = percentile[i+1]
					percentile[i+1] = lastResult
				}
			}

			ninetyfifthPercentile = percentile[percentileLength-int(math.Floor(float64(iterations)*.05+0.95))]

			for _, step := range iteration.Steps {
				cmd := commands[step.Command]
				cmd.Count = cmd.Count + 1
				cmd.TotalTime = cmd.TotalTime + step.Duration
				cmd.LastTime = step.Duration
				cmd.Average = time.Duration(cmd.TotalTime.Nanoseconds() / cmd.Count)
				cmd.Throughput = float64(cmd.Count) / cmd.TotalTime.Seconds()
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
		case _ = <-heartbeat.C:
			//heartbeat for updating CLI Walltime every second
		}
		ex.samples <- &Sample{clone(commands), avg, totalTime, iterations, totalErrors, workers, lastResult, lastError, worstResult, ninetyfifthPercentile, time.Now().Sub(startTime), sampleType}
	}
}
