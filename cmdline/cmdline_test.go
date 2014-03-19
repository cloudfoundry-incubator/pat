package cmdline_test

import (
	. "github.com/julz/pat/cmdline"
	"github.com/julz/pat/benchmarker"
	"github.com/julz/pat/config"
	"github.com/julz/pat/experiment"
	"github.com/julz/pat/laboratory"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Cmdline", func() {
	var (
		flags  config.Config
		args   []string
		lab    *dummyLab
		worker benchmarker.Worker
	)

	JustBeforeEach(func() {
		flags = config.NewConfig()
		InitCommandLineFlags(flags)
		flags.Parse(args)
		lab = &dummyLab{}
		worker = benchmarker.NewWorker()
		RunCommandLineWithLabAndWorker(lab, worker)
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
		BeforeEach(func() {
			args = []string{"-workload", "login,push"}
		})

		It("configures the experiment with the parameter", func() {
			Ω(lab).Should(HaveBeenRunWith("workload", "login,push"))
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
	field string
	value interface{}
}

func HaveBeenRunWith(field string, value interface{}) OmegaMatcher {
	return &runWithMatcher{field, value}
}

func (m *runWithMatcher) Match(actualLab interface{}) (bool, string, error) {
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
	return Equal(actual).Match(m.value)
}

type dummyLab struct {
	lastRunWith *experiment.RunnableExperiment
}

func (d *dummyLab) GetData(guid string) ([]*experiment.Sample, error) {
	return nil, nil
}

func (d *dummyLab) Run(runnable laboratory.Runnable) (experiment.Experiment, error) {
	return nil, nil
}

func (d *dummyLab) RunWithHandlers(runnable laboratory.Runnable, handlers []func(<-chan *experiment.Sample)) (experiment.Experiment, error) {
	d.lastRunWith = runnable.(*experiment.RunnableExperiment)
	return nil, nil
}

func (d *dummyLab) Visit(func(experiment.Experiment)) {
}
