package pat

import (
	"fmt"
	"strings"

	"github.com/julz/pat/benchmarker"
	. "github.com/julz/pat/experiment"
	"github.com/julz/pat/experiments"
	. "github.com/julz/pat/laboratory"
	"github.com/julz/pat/store"
)

type Response struct {
	TotalTime int64
	Timestamp int64
}

func RunCommandLine(concurrency int, iterations int, silent bool, name string, interval int, stop int, workload string) (err error) {
	handlers := make([]func(<-chan *Sample), 0)

	if !silent {
		handlers = append(handlers, func(s <-chan *Sample) { display(concurrency, iterations, interval, stop, s) })
	}

	worker := benchmarker.NewWorker()
	worker.AddExperiment("login", experiments.Dummy)
	worker.AddExperiment("push", experiments.Dummy)

	NewLaboratory(store.NewCsvStore("output/csvs")).RunWithHandlers(
		NewRunnableExperiment(
			NewExperimentConfiguration(
				iterations, concurrency, interval, stop, worker, workload)), handlers)

	return nil
}

func display(concurrency int, iterations int, interval int, stop int, samples <-chan *Sample) {
	for s := range samples {
		fmt.Print("\033[2J\033[;H")
		fmt.Println("\x1b[32;1mCloud Foundry Performance Acceptance Tests\x1b[0m")
		fmt.Printf("Test underway. Concurrency: \x1b[36m%v\x1b[0m  Workload iterations: \x1b[36m%v\x1b[0m  Interval: \x1b[36m%v\x1b[0m  Stop: \x1b[36m%v\x1b[0m\n", concurrency, iterations, interval, stop)
		fmt.Println("┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄\n")
		if stop > 0 && interval > 0 {
			fmt.Printf("\x1b[36mTotal iterations\x1b[0m:    %v  \x1b[36m%v\x1b[0m / %v\n", bar(s.Total, int64((iterations*stop)/interval), 25), s.Total, int64((iterations*stop)/interval))
		} else {
			fmt.Printf("\x1b[36mTotal iterations\x1b[0m:    %v  \x1b[36m%v\x1b[0m / %v\n", bar(s.Total, int64(iterations), 25), s.Total, int64(iterations))
		}
		fmt.Println()
		fmt.Printf("\x1b[1mLatest iteration\x1b[0m:  \x1b[36m%v\x1b[0m\n", s.LastResult)
		fmt.Printf("\x1b[1mWorst iteration\x1b[0m:   \x1b[36m%v\x1b[0m\n", s.WorstResult)
		fmt.Printf("\x1b[1mAverage iteration\x1b[0m: \x1b[36m%v\x1b[0m\n", s.Average)
		fmt.Printf("\x1b[1mTotal time\x1b[0m:        \x1b[36m%v\x1b[0m\n", s.TotalTime)
		fmt.Printf("\x1b[1mWall time\x1b[0m:         \x1b[36m%v\x1b[0m\n", s.WallTime)
		fmt.Printf("\x1b[1mRunning Workers\x1b[0m:   \x1b[36m%v\x1b[0m\n", s.TotalWorkers)
		fmt.Println()
		fmt.Println("\x1b[32;1mCommands Issued:\x1b[0m")
		fmt.Println()
		for key, command := range s.Commands {
			fmt.Printf("\x1b[1m%v\x1b[0m:\n", key)
			fmt.Printf("\x1b[1m\tCount\x1b[0m:      \x1b[36m%v\x1b[0m\n", command.Count)
			fmt.Printf("\x1b[1m\tAverage\x1b[0m:    \x1b[36m%v\x1b[0m\n", command.Average)
			fmt.Printf("\x1b[1m\tLast time\x1b[0m:  \x1b[36m%v\x1b[0m\n", command.LastTime)
			fmt.Printf("\x1b[1m\tWorst time\x1b[0m: \x1b[36m%v\x1b[0m\n", command.WorstTime)
			fmt.Printf("\x1b[1m\tTotal time\x1b[0m: \x1b[36m%v\x1b[0m\n", command.TotalTime)
		}
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
