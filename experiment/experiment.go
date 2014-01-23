package experiment

import (
  . "github.com/julz/pat/benchmarker"
  "github.com/julz/pat/experiments"
  "time"
)

type SampleType int

const (
  ResultSample SampleType = iota
  WorkerSample
  ErrorSample
  OtherSample
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
  WallTime     time.Duration
  Type         SampleType
}

type RunningExperiment struct {
  results chan time.Duration
  errors  chan error
  workers chan int
  samples chan *Sample
}

func Run(pushes int, concurrency int, tracker func(chan *Sample)) error {
  result := make(chan time.Duration)
  errors := make(chan error)
  workers := make(chan int)
  samples := make(chan *Sample)
  go Track(samples, result, errors, workers)
  go tracker(samples)
  ExecuteConcurrently(concurrency, Repeat(pushes, Counted(workers, Timed(result, errors, experiments.Dummy))))
  return nil
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
  startTime := time.Now()

  for {
    sampleType := OtherSample
    select {
    case result := <-ex.results:
      sampleType = ResultSample
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

    ex.samples <- &Sample{avg, totalTime, n, totalErrors, workers, lastResult, lastError, worstResult, time.Now().Sub(startTime), sampleType}
  }
}
