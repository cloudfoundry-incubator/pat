package benchmarker

import (
	"strings"
	"time"
	"fmt"

	"github.com/cloudfoundry-incubator/pat/workloads"
)

type LocalWorker struct {
	defaultWorker
}

func NewLocalWorker() *LocalWorker {
	return &LocalWorker{defaultWorker{make(map[string]workloads.WorkloadStep)}}
}

func (self *LocalWorker) Time(experiment string, workloadCtx map[string]interface{}) (result IterationResult) {		
	fmt.Println("====== go func() 1a" + experiment)
	experiments := strings.Split(experiment, ",")
	fmt.Println("====== go func() 1b " )
	var start = time.Now()
	for _, e := range experiments {		
		stepTime, err := Time(func() error { return self.Experiments[e].Fn(workloadCtx) })		
		result.Steps = append(result.Steps, StepResult{e, stepTime})		
		if err != nil {
			result.Error = encodeError(err)
			break
		}
	}
	result.Duration = time.Now().Sub(start)
	return
}
