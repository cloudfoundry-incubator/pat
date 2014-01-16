package pat

import (
	"fmt"
	. "github.com/julz/pat/benchmarker"
	"github.com/julz/pat/experiments"
	"strings"
	"time"
)

type Response struct {
	TotalTime int64
	Timestamp int64
}

func RunCommandLine(pushes int, concurrency int, silent bool) *Response {
	result := make(chan time.Duration)
	errors := make(chan error)
	workers := make(chan int)
	samples := make(chan *Sample)
	go Track(samples, result, errors, workers)

	// TODO(jz) move this out to a presenter
	go func(samples chan *Sample, target int) {
		for s := range samples {
			if !silent {
				fmt.Print("\033[2J\033[;H")
				fmt.Println("\x1b[32;1mCloud Foundry Performance Acceptance Tests\x1b[0m")
				fmt.Printf("Test underway.  Pushes: \x1b[36m%v\x1b[0m  Concurrency: \x1b[36m%v\x1b[0m\n", pushes, concurrency)
				fmt.Println("┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄\n")
				fmt.Printf("\x1b[36mTotal pushes\x1b[0m:    %v  \x1b[36m%v\x1b[0m / %v\n", bar(s.Total, int64(target), 25), s.Total, int64(target))
				fmt.Println()
				fmt.Printf("Latest Push\x1b[0m:     \x1b[36m%v\x1b[0m\n", s.LastResult)
				fmt.Printf("Worst Push\x1b[0m:      \x1b[36m%v\x1b[0m\n", s.WorstResult)
				fmt.Printf("Average\x1b[0m:         \x1b[36m%v\x1b[0m\n", s.Average)
				fmt.Printf("Total time\x1b[0m:      \x1b[36m%v\x1b[0m\n", s.TotalTime)
				fmt.Printf("Wall time\x1b[0m:       \x1b[36m%v\x1b[0m\n", s.WallTime)
				fmt.Printf("\x1b[1mRunning Workers\x1b[0m: \x1b[36m%v\x1b[0m\n", s.TotalWorkers)
				fmt.Println()
				fmt.Println("┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄\n")
				if s.TotalErrors > 0 {
					fmt.Printf("Total errors: %d\n", s.TotalErrors)
					fmt.Printf("Last error: %v\n", "")
				}
			}
		}
	}(samples, pushes)

	totalTime, _ := Time(func() error {
		ExecuteConcurrently(concurrency, Repeat(pushes, Counted(workers, Timed(result, errors, experiments.Dummy))))
		return nil
	})

	return &Response{totalTime.Nanoseconds(), time.Now().UnixNano()}
}

func bar(n int64, total int64, size int) (bar string) {
	if n == 0 {
		n = 1
	}
	progress := int64(size) / (total / n)
	return "╞" + strings.Repeat("═", int(progress)) + strings.Repeat("┄", size-int(progress)) + "╡"
}
