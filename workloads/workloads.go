package workloads

import (
	"github.com/julz/pat/config"
)

type WorkloadAdder interface {
	AddWorkloadStep(WorkloadStep)
}

type WorkloadStep struct {
	Name         string
	Fn           func() error
	Description  string
}


type WorkloadList struct {
	workloads []WorkloadStep
}

var restContext = NewRestWorkloadContext()

func DefaultWorkloadList() *WorkloadList {
	return &WorkloadList{[]WorkloadStep {
		WorkloadStep{"rest:target", restContext.Target,"Sets the CF target"},
		WorkloadStep{"rest:login", restContext.Login, "Performs a login to the REST api. This option requires rest:target to be included in the list of workloads"},
		WorkloadStep{"rest:push", restContext.Push, "Pushes a simple Ruby application using the REST api. This option requires both rest:target and rest:login to be included in the list of workloads"},
		WorkloadStep{"gcf:push", Push, "Pushes a simple Ruby application using the CF command-line"},
		WorkloadStep{"dummy", Dummy, "An empty workload that can be used when a CF environment is not available"},
		WorkloadStep{"dummyWithErrors", DummyWithErrors, "An empty workload that generates errors. This can be used when a CF environment is not available"},
	}}
}

func (self *WorkloadList) DescribeWorkloads(to WorkloadAdder) {
	for _,workload := range self.workloads {
		to.AddWorkloadStep(workload);
	}
}

func (self *WorkloadList) DescribeParameters(config config.Config) {
	restContext.DescribeParameters(config)
}

