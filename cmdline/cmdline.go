package cmdline

import (
	"fmt"
	"os"
	"strings"

	"github.com/julz/pat/benchmarker"
	"github.com/julz/pat/config"
	. "github.com/julz/pat/experiment"
	"github.com/julz/pat/experiments"
	. "github.com/julz/pat/laboratory"
	"github.com/julz/pat/store"
)

type Response struct {
	TotalTime int64
	Timestamp int64
}

var params = struct {
	iterations  int
	concurrency int
	silent      bool
	output      string
	workload    string
	interval    int
	stop        int
	csvDir      string
}{}

var restContext = experiments.NewContext()

func InitCommandLineFlags(config config.Config) {
	config.IntVar(&params.iterations, "iterations", 1, "number of pushes to attempt")
	config.IntVar(&params.concurrency, "concurrency", 1, "max number of pushes to attempt in parallel")
	config.BoolVar(&params.silent, "silent", false, "true to run the commands and print output the terminal")
	config.StringVar(&params.output, "output", "", "if specified, writes benchmark results to a CSV file")
	config.StringVar(&params.workload, "workload", "", "The set of operations a user should issue (ex. login,push,push)")
	config.IntVar(&params.interval, "interval", 0, "repeat a workload at n second interval, to be used with -stop")
	config.IntVar(&params.stop, "stop", 0, "stop a repeating interval after n second, to be used with -interval")
	config.StringVar(&params.csvDir, "csvDir", "output/csvs", "Directory to Store CSVs")
	restContext.DescribeParameters(config)
}

func RunCommandLine() error {
	lab := NewLaboratory(store.NewCsvStore("output/csvs"))
	worker := benchmarker.NewWorker()
	err := RunCommandLineWithLabAndWorker(lab, worker)

	for {
		in := make([]byte, 1)
		os.Stdin.Read(in)
		if string(in) == "q" {
			return err
		}
	}
}

func RunCommandLineWithLabAndWorker(lab Laboratory, worker benchmarker.Worker) (err error) {
	handlers := make([]func(<-chan *Sample), 0)

	if !params.silent {
		handlers = append(handlers, func(s <-chan *Sample) {
			display(params.concurrency, params.iterations, params.interval, params.stop, s)
		})
	}

	rest := restContext
	worker.AddExperiment("rest:login", rest.Login)
	worker.AddExperiment("rest:push", rest.Push)
	worker.AddExperiment("rest:target", rest.Target)
	worker.AddExperiment("gcf:push", experiments.Push)
	worker.AddExperiment("dummy", experiments.Dummy)
	worker.AddExperiment("dummyWithErrors", experiments.DummyWithErrors)

	lab.RunWithHandlers(
		NewRunnableExperiment(
			NewExperimentConfiguration(
				params.iterations, params.concurrency, params.interval, params.stop, worker, params.workload)), handlers)

	return nil
}

func display(concurrency int, iterations int, interval int, stop int, samples <-chan *Sample) {
	for s := range samples {
		fmt.Print("\033[2J\033[;H")
		fmt.Println("\x1b[32;1mCloud Foundry Performance Acceptance Tests\x1b[0m")
		fmt.Printf("Test underway. Concurrency: \x1b[36m%v\x1b[0m  Workload iterations: \x1b[36m%v\x1b[0m  Interval: \x1b[36m%v\x1b[0m  Stop: \x1b[36m%v\x1b[0m\n", concurrency, iterations, interval, stop)
		fmt.Println("┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄\n")

		fmt.Printf("\x1b[36mTotal iterations\x1b[0m:    %v  \x1b[36m%v\x1b[0m / %v\n", bar(s.Total, totalIterations(iterations, interval, stop), 25), s.Total, totalIterations(iterations, interval, stop))

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
			fmt.Printf("\x1b[1m\tCount\x1b[0m:                 \x1b[36m%v\x1b[0m\n", command.Count)
			fmt.Printf("\x1b[1m\tAverage\x1b[0m:               \x1b[36m%v\x1b[0m\n", command.Average)
			fmt.Printf("\x1b[1m\tLast time\x1b[0m:             \x1b[36m%v\x1b[0m\n", command.LastTime)
			fmt.Printf("\x1b[1m\tWorst time\x1b[0m:            \x1b[36m%v\x1b[0m\n", command.WorstTime)
			fmt.Printf("\x1b[1m\tTotal time\x1b[0m:            \x1b[36m%v\x1b[0m\n", command.TotalTime)
			fmt.Printf("\x1b[1m\tPer second throughput\x1b[0m: \x1b[36m%v\x1b[0m\n", command.Throughput)
		}
		fmt.Println("┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄")
		if s.TotalErrors > 0 {
			fmt.Printf("\nTotal errors: %d\n", s.TotalErrors)
			fmt.Printf("Last error: %v\n", "")
		}
		fmt.Println()
		fmt.Println("Type q <Enter> (or ctrl-c) to exit")
	}
}

func totalIterations(iterations int, interval int, stopTime int) int64 {
	var totalIterations int

	if stopTime > 0 && interval > 0 {
		totalIterations = ((stopTime / interval) + 1) * iterations
	} else {
		totalIterations = iterations
	}

	return int64(totalIterations)
}

func bar(n int64, total int64, size int) (bar string) {
	if n == 0 {
		n = 1
	}
	progress := int64(size) / (total / n)
	return "╞" + strings.Repeat("═", int(progress)) + strings.Repeat("┄", size-int(progress)) + "╡"
}
