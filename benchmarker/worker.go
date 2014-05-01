package benchmarker

import (
	"errors"
	"strings"

	"github.com/cloudfoundry-incubator/pat/workloads"
)

type Worker interface {
	Time(experiment string, workerIndex int) IterationResult
	AddWorkloadStep(workloads workloads.WorkloadStep)
	Visit(fn func(workloads.WorkloadStep))
	Validate(name string) (result bool, err error)
}

type defaultWorker struct {
	Experiments map[string]workloads.WorkloadStep
}

func (self *defaultWorker) AddWorkloadStep(workload workloads.WorkloadStep) {
	self.Experiments[workload.Name] = workload
}

func (self *defaultWorker) Visit(fn func(workloads.WorkloadStep)) {
	for _, e := range self.Experiments {
		fn(e)
	}
}

func (self *defaultWorker) Validate(name string) (ok bool, err error) {
	ok = true
	ws := strings.Split(name, ",")
	for _, w := range ws {
		var valid = false
		self.Visit(func(workload workloads.WorkloadStep) {
			if workload.Name == w {
				valid = true
			}
		})
		if !valid {
			ok = false
			err = errors.New(w)
			break
		}
	}
	return
}
