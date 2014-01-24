package pat

import (
  "fmt"
  "github.com/julz/pat/experiment"
  "github.com/julz/pat/output"
  "strings"
)

type Response struct {
  TotalTime int64
  Timestamp int64
}

func RunCommandLine(pushes int, concurrency int, silent bool, name string, interval int, stop int) error {
  handlers := make([]func(chan *experiment.Sample), 0)

  if !silent {
    handlers = append(handlers, func(s chan *experiment.Sample) { display(int64(pushes), concurrency, interval, stop, s) })
  }

  if len(name) > 0 {
    handlers = append(handlers, output.NewCsvWriter(name).Write)
  }

  return experiment.Run(pushes, concurrency, interval, stop, output.Multiplexer(handlers).Multiplex)
}

func display(target int64, concurrency int, interval int, stop int, samples chan *experiment.Sample) {
  //temp workaround:(simon): with a repeating workload,  we don't know the total target of pushes until the interval stops, so we set target = current s.Total
  for s := range samples {
    if s.Total > target {
      target = s.Total
    }
    fmt.Print("\033[2J\033[;H")
    fmt.Println("\x1b[32;1mCloud Foundry Performance Acceptance Tests\x1b[0m")
    fmt.Printf("Test underway.  Pushes: \x1b[36m%v\x1b[0m  Concurrency: \x1b[36m%v\x1b[0m\n", target, concurrency)
    if interval > 0 && stop > 0 {
      fmt.Printf("\x1b[31mCurrent workload repeats at %d sec. interval, stops after %d sec.\x1b[0m\n", interval, stop)
    }
    fmt.Println("┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄\n")
    fmt.Printf("\x1b[36mTotal pushes\x1b[0m:    %v  \x1b[36m%v\x1b[0m / %v\n", bar(s.Total, int64(target), 25), s.Total, int64(target))
    fmt.Println()
    fmt.Printf("\x1b[1mLatest Push\x1b[0m:     \x1b[36m%v\x1b[0m\n", s.LastResult)
    fmt.Printf("\x1b[1mWorst Push\x1b[0m:      \x1b[36m%v\x1b[0m\n", s.WorstResult)
    fmt.Printf("\x1b[1mAverage\x1b[0m:         \x1b[36m%v\x1b[0m\n", s.Average)
    fmt.Printf("\x1b[1mTotal time\x1b[0m:      \x1b[36m%v\x1b[0m\n", s.TotalTime)
    fmt.Printf("\x1b[1mWall time\x1b[0m:       \x1b[36m%v\x1b[0m\n", s.WallTime)
    fmt.Printf("\x1b[1mRunning Workers\x1b[0m: \x1b[36m%v\x1b[0m\n", s.TotalWorkers)
    fmt.Println()
    fmt.Println("┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄")
    if s.TotalErrors > 0 {
      fmt.Printf("\nTotal errors: %d\n", s.TotalErrors)
      fmt.Printf("Last error: %v\n", "")
    }
  }
}

func bar(n int64, total int64, size int) (bar string) {
  if n == 0 {
    n = 1
  }
  progress := int64(size) / (total / n)
  return "╞" + strings.Repeat("═", int(progress)) + strings.Repeat("┄", size-int(progress)) + "╡"
}
