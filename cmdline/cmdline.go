package cmdline

import (
	"fmt"
	"os"
	"strings"

	"github.com/cloudfoundry-incubator/pat/context"
	"github.com/cloudfoundry-incubator/pat/benchmarker"
	"github.com/cloudfoundry-incubator/pat/config"
	. "github.com/cloudfoundry-incubator/pat/experiment"
	. "github.com/cloudfoundry-incubator/pat/laboratory"
	"github.com/cloudfoundry-incubator/pat/store"
	"github.com/cloudfoundry-incubator/pat/workloads"
)

var params = struct {
	iterations    int
	listWorkloads bool
	concurrency   int
	silent        bool
	output        string
	workload      string
	interval      int
	stop          int
}{}

func InitCommandLineFlags(config config.Config) {	
	config.IntVar(&params.iterations, "iterations", 1, "number of pushes to attempt")
	config.IntVar(&params.concurrency, "concurrency", 1, "max number of pushes to attempt in parallel")
	config.BoolVar(&params.silent, "silent", false, "true to run the commands and print output the terminal")
	config.StringVar(&params.output, "output", "", "if specified, writes benchmark results to a CSV file")
	config.StringVar(&params.workload, "workload", "gcf:push", "a comma-separated list of operations a user should issue (use -list-workloads to see available workload options)")
	config.IntVar(&params.interval, "interval", 0, "repeat a workload at n second interval, to be used with -stop")
	config.IntVar(&params.stop, "stop", 0, "stop a repeating interval after n second, to be used with -interval")
	config.BoolVar(&params.listWorkloads, "list-workloads", false, "Lists the available workloads")
	benchmarker.DescribeParameters(config)
	store.DescribeParameters(config)
}

func RunCommandLine() error {	
	params.workload = strings.Replace(params.workload, " ", "", -1)
	return WithConfiguredWorkerAndSlaves(func(worker benchmarker.Worker) error {
		return validateParameters(worker, func() error {
			return store.WithStore(func(store Store) error {

				lab := LaboratoryFactory(store)

				handlers := make([]func(<-chan *Sample), 0)
				if !params.silent {
					handlers = append(handlers, func(s <-chan *Sample) {
						display(params.concurrency, params.iterations, params.interval, params.stop, s)
					})
				}

				//workloadContext := context.WorkloadContext( context.NewWorkloadContent() )
				workloadContext := context.New()

				lab.RunWithHandlers(
					NewRunnableExperiment(
						NewExperimentConfiguration(
							params.iterations, params.concurrency, params.interval, params.stop, worker, params.workload)), handlers, workloadContext)

				BlockExit()
				return nil
			})
		})
	})
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
