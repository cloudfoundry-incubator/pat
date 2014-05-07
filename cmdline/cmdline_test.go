package cmdline_test

import (
	"fmt"
	"time"

	"github.com/cloudfoundry-incubator/pat/context"
	"github.com/cloudfoundry-incubator/pat/benchmarker"
	. "github.com/cloudfoundry-incubator/pat/cmdline"
	"github.com/cloudfoundry-incubator/pat/config"
	"github.com/cloudfoundry-incubator/pat/experiment"
	"github.com/cloudfoundry-incubator/pat/laboratory"
	"github.com/cloudfoundry-incubator/pat/workloads"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Cmdline", func() {
	var (
		flags config.Config
		args  []string
		lab   *dummyLab
		err   error
	)

	BeforeEach(func() {
		WithConfiguredWorkerAndSlaves = func(fn func(Worker benchmarker.Worker) error) error {
			worker := benchmarker.NewLocalWorker()
			worker.AddWorkloadStep(workloads.Step("login", nil, "description"))
			worker.AddWorkloadStep(workloads.Step("push", nil, "description"))
			worker.AddWorkloadStep(workloads.Step("gcf:push", nil, "description"))
			return fn(worker)
		}

		LaboratoryFactory = func(store laboratory.Store) (newLab laboratory.Laboratory) {
			lab = &dummyLab{}
			newLab = lab
			return
		}
	})

	JustBeforeEach(func() {
		flags = config.NewConfig()
		InitCommandLineFlags(flags)
		flags.Parse(args)

		BlockExit = func() {}
		err = RunCommandLine()
	})

	Describe("When -iterations is supplied", func() {
		BeforeEach(func() {
			args = []string{"-iterations", "3"}
		})

		It("configures the experiment with the parameter", func() {
			Ω(lab).Should(HaveBeenRunWith("iterations", 3))
		})
	})

	Describe("When -concurrency is supplied with a single integer", func() {
		BeforeEach(func() {
			args = []string{"-concurrency", "3"}
		})

		It("configures the experiment with the parameter", func() {
			Ω(lab).Should(HaveBeenRunWith("concurrency", []int{3}))
		})
	})

	Describe("When -concurrency is supplied with a range", func() {
		BeforeEach(func() {
			args = []string{"-concurrency", "1..3"}
		})

		It("configures the experiment with the parameter", func() {
			Ω(lab).Should(HaveBeenRunWith("concurrency", []int{1, 3}))
		})
	})

	Describe("When -concurrency is supplied with an incorrectly formatted input", func() {
		BeforeEach(func() {
			args = []string{"-concurrency", "1-3"}
		})

		It("throws an error", func() {
			Ω(err).ShouldNot(BeNil())
		})
	})

	Describe("When -workload is supplied", func() {
		BeforeEach(func() {
			args = []string{"-workload", "login,push"}
		})

		It("configures the experiment with the parameter", func() {
			Ω(lab).Should(HaveBeenRunWith("workload", "login,push"))
		})
	})

	Describe("When -list-workloads is supplied", func() {
		var (
			printCalledCount int
		)

		BeforeEach(func() {
			lab = nil
			args = []string{"-list-workloads"}

			printCalledCount = 0
			PrintWorkload = func(workload workloads.WorkloadStep) {
				printCalledCount++
			}
		})

		It("prints the list of available workloads and exits", func() {
			Ω(printCalledCount).Should(BeNumerically("==", 3))
			Ω(lab).Should(BeNil())
		})
	})

	Describe("When -interval is supplied", func() {
		BeforeEach(func() {
			args = []string{"-interval", "10"}
		})

		It("configures the experiment with the parameter", func() {
			Ω(lab).Should(HaveBeenRunWith("interval", 10))
		})
	})

	Describe("When -stop is supplied", func() {
		BeforeEach(func() {
			args = []string{"-stop", "11"}
		})

		It("configures the experiment with the parameter", func() {
			Ω(lab).Should(HaveBeenRunWith("stop", 11))
		})
	})

	Describe("When -concurrency:timeBetweenSteps is supplied", func() {
		BeforeEach(func() {
			args = []string{"-concurrency:timeBetweenSteps", "3"}
		})

		It("configures the experiment with the parameter", func() {
			Ω(lab).Should(HaveBeenRunWith("concurrencysteptime", 3*time.Second))
		})
	})
})

type runWithMatcher struct {
	field     string
	value     interface{}
	lastMatch interface{}
}

func HaveBeenRunWith(field string, value interface{}) OmegaMatcher {
	return &runWithMatcher{field, value, nil}
}

func (m *runWithMatcher) Match(actualLab interface{}) (bool, error) {
	runWith := actualLab.(*dummyLab).lastRunWith
	var actual interface{}
	switch m.field {
	case "iterations":
		actual = runWith.Iterations
	case "concurrency":
		actual = runWith.Concurrency
	case "workload":
		actual = runWith.Workload
	case "interval":
		actual = runWith.Interval
	case "stop":
		actual = runWith.Stop
	case "concurrencysteptime":
		actual = runWith.ConcurrencyStepTime
	}
	m.lastMatch = actual
	return Equal(actual).Match(m.value)
}

func (m *runWithMatcher) NegatedFailureMessage(actual interface{}) string {
	return fmt.Sprintf("Expected\n\t%#v\n not to to have been run with \n\t-%#v: %#v (but was run with: %#v)", actual, m.field, m.value, m.lastMatch)
}

func (m *runWithMatcher) FailureMessage(actual interface{}) string {
	return fmt.Sprintf("Expected\n\t%#v\nto to have been run with \n\t-%#v: %#v", actual, m.field, m.value, m.lastMatch)
}

type dummyLab struct {
	lastRunWith *experiment.RunnableExperiment
}

func (d *dummyLab) GetData(guid string) ([]*experiment.Sample, error) {
	return nil, nil
}

func (d *dummyLab) Run(runnable laboratory.Runnable, workloadCtx context.Context) (string, error) {
	return "", nil
}

func (d *dummyLab) RunWithHandlers(runnable laboratory.Runnable, handlers []func(<-chan *experiment.Sample), workloadCtx context.Context) (string, error) {
	d.lastRunWith = runnable.(*experiment.RunnableExperiment)
	return "", nil
}

func (d *dummyLab) Visit(func(experiment.Experiment)) {
}
