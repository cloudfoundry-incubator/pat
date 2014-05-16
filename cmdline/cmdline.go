package cmdline

import (
	"os"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cloudfoundry-incubator/pat/benchmarker"
	"github.com/cloudfoundry-incubator/pat/config"
	. "github.com/cloudfoundry-incubator/pat/experiment"
	. "github.com/cloudfoundry-incubator/pat/laboratory"
	"github.com/cloudfoundry-incubator/pat/store"
	"github.com/cloudfoundry-incubator/pat/workloads"
)

var params = struct {
	iterations          int
	listWorkloads       bool
	concurrency         string
	concurrencyStepTime int
	silent              bool
	output              string
	workload            string
	interval            int
	stop                int
	note                string
}{}

func InitCommandLineFlags(config config.Config) {
	config.IntVar(&params.iterations, "iterations", 1, "number of pushes to attempt")
	config.StringVar(&params.concurrency, "concurrency", "1", "number of workers to execute the workload in parallel, can be static or ramping up, i.e. 1..3")
	config.IntVar(&params.concurrencyStepTime, "concurrency:timeBetweenSteps", 60, "seconds between adding additonal workers when ramping works up")
	config.BoolVar(&params.silent, "silent", false, "true to run the commands and print output the terminal")
	config.StringVar(&params.output, "output", "", "if specified, writes benchmark results to a CSV file")
	config.StringVar(&params.workload, "workload", "gcf:push", "a comma-separated list of operations a user should issue (use -list-workloads to see available workload options)")
	config.IntVar(&params.interval, "interval", 0, "repeat a workload at n second interval, to be used with -stop")
	config.IntVar(&params.stop, "stop", 0, "stop a repeating interval after n second, to be used with -interval")
	config.BoolVar(&params.listWorkloads, "list-workloads", false, "Lists the available workloads")
	config.StringVar(&params.note, "note", "", "Description about the experiment")
	benchmarker.DescribeParameters(config)
	store.DescribeParameters(config)
}

func RunCommandLine() error {
	return WithConfiguredWorkerAndSlaves(func(worker benchmarker.Worker) error {
		return validateParameters(worker, func() error {
			return store.WithStore(func(store Store) error {

				parsedConcurrency, err := parseConcurrency(params.concurrency)
				parsedConcurrencyStepTime := parseConcurrencyStepTime(params.concurrencyStepTime)

				lab := LaboratoryFactory(store)

				handlers := make([]func(<-chan *Sample), 0)
				if !params.silent {
					handlers = append(handlers, func(s <-chan *Sample) {
						display(params.concurrency, params.iterations, params.interval, params.stop, params.concurrencyStepTime, s)
					})
				}

				lab.RunWithHandlers(
					NewRunnableExperiment(
						NewExperimentConfiguration(
							params.iterations, parsedConcurrency, parsedConcurrencyStepTime, params.interval, params.stop, worker, params.workload, params.note)), handlers)

				BlockExit()
				return err
			})
		})
	})
}

func parseConcurrency(concurrency string) ([]int, error) {
	rawConcurrency := strings.SplitN(concurrency, "..", 2)
	parsedConcurrency := make([]int, len(rawConcurrency))
	for i, v := range rawConcurrency {
		intV, err := strconv.Atoi(v)
		if err != nil {
			parsedConcurrency = []int{1}
			return parsedConcurrency, err
		} else {
			parsedConcurrency[i] = intV
		}

	}
	return parsedConcurrency, nil
}

func parseConcurrencyStepTime(concurrencyStepTime int) time.Duration {
	parsedConcurrencyStepTime := time.Duration(concurrencyStepTime) * time.Second
	return parsedConcurrencyStepTime
}

func validateParameters(worker benchmarker.Worker, then func() error) error {
	if params.listWorkloads {
		worker.Visit(PrintWorkload)
		return nil
	}

	var ok, err = worker.Validate(params.workload)

	if !ok {
		fmt.Printf("Invalid workload: '%s'\n\n", err)
		fmt.Println("Available workloads:\n")
		worker.Visit(PrintWorkload)
		return err
	}

	return then()
}

var WithConfiguredWorkerAndSlaves = func(fn func(worker benchmarker.Worker) error) error {
	return benchmarker.WithConfiguredWorkerAndSlaves(fn)
}

var LaboratoryFactory = func(store Store) (lab Laboratory) {
	lab = NewLaboratory(store)
	return
}

var BlockExit = func() {
	for {
		in := make([]byte, 1)
		os.Stdin.Read(in)
		if string(in) == "q" {
			return
		}
	}
}

var PrintWorkload = func(workload workloads.WorkloadStep) {
	fmt.Printf("\x1b[1m%s\x1b[0m\n\t%s\n", workload.Name, workload.Description)
}
