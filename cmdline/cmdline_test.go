package cmdline_test

import (
	"fmt"

	"github.com/cloudfoundry-community/pat/benchmarker"
	. "github.com/cloudfoundry-community/pat/cmdline"
	"github.com/cloudfoundry-community/pat/config"
	"github.com/cloudfoundry-community/pat/experiment"
	"github.com/cloudfoundry-community/pat/laboratory"
	"github.com/cloudfoundry-community/pat/workloads"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Cmdline", func() {
	var (
		flags config.Config
		args  []string
		lab   *dummyLab
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
		RunCommandLine()
	})

	Describe("When -iterations is supplied", func() {
		BeforeEach(func() {
			args = []string{"-iterations", "3"}
		})

		It("configures the experiment with the parameter", func() {
			Ω(lab).Should(HaveBeenRunWith("iterations", 3))
		})
	})

	Describe("When -concurrency is supplied", func() {
		BeforeEach(func() {
			args = []string{"-concurrency", "3"}
		})

		It("configures the experiment with the parameter", func() {
			Ω(lab).Should(HaveBeenRunWith("concurrency", 3))
		})
	})

	Describe("When -workload is supplied", func() {
		Describe("When -workload contains no white spaces", func() {
			BeforeEach(func() {
				args = []string{"-workload", "login,push"}
			})

			It("configures the experiment with the parameter", func() {
				Ω(lab).Should(HaveBeenRunWith("workload", "login,push"))
			})
		})

		Describe("When -workload contains white spaces", func() {
			BeforeEach(func() {
				args = []string{"-workload", "  login ,  push , gcf:push"}
			})

			It("removes white spaces in the parameter", func() {
				Ω(lab).Should(HaveBeenRunWith("workload", "login,push,gcf:push"))
			})
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

func (d *dummyLab) Run(runnable laboratory.Runnable, workloadCtx map[string]interface{}) (string, error) {
	return "", nil
}

func (d *dummyLab) RunWithHandlers(runnable laboratory.Runnable, handlers []func(<-chan *experiment.Sample), workloadCtx map[string]interface{}) (string, error) {
	d.lastRunWith = runnable.(*experiment.RunnableExperiment)
	return "", nil
}

func (d *dummyLab) Visit(func(experiment.Experiment)) {
}
