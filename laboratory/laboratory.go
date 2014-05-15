package laboratory

import (
	"github.com/cloudfoundry-incubator/pat/experiment"
	"github.com/nu7hatch/gouuid"
)

type lab struct {
	store  Store
	loaded []experiment.Experiment
}

type Laboratory interface {
	Run(ex Runnable) (string, error)
	RunWithHandlers(ex Runnable, fns []func(samples <-chan *experiment.Sample)) (string, error)
	Visit(fn func(ex experiment.Experiment))
	GetData(name string) ([]*experiment.Sample, error)
}

type Runnable interface {
	Run(handler func(samples <-chan *experiment.Sample)) error
	GetExperimentConfiguration() experiment.ExperimentConfiguration
}

type Store interface {
	Writer(guid string, ex experiment.ExperimentConfiguration) func(samples <-chan *experiment.Sample)
	LoadAll() ([]experiment.Experiment, error)
}

func NewLaboratory(history Store) Laboratory {
	lab := &lab{history, make([]experiment.Experiment, 0)}
	lab.reload()
	return lab
}

func (self *lab) reload() {
	self.loaded, _ = self.store.LoadAll()
}

func (self *lab) Run(ex Runnable) (string, error) {
	return self.RunWithHandlers(ex, make([]func(<-chan *experiment.Sample), 0))
}

func (self *lab) RunWithHandlers(ex Runnable, additionalHandlers []func(<-chan *experiment.Sample)) (string, error) {
	guid, _ := uuid.NewV4()

	handlers := make([]func(<-chan *experiment.Sample), 1)
	handlers[0] = self.store.Writer(guid.String(), ex.GetExperimentConfiguration())
	for _, h := range additionalHandlers {
		handlers = append(handlers, h)
	}
	go ex.Run(Multiplexer(handlers).Multiplex)
	return guid.String(), nil
}

func (self *lab) Visit(fn func(ex experiment.Experiment)) {
	self.reload()
	for _, e := range self.loaded {
		fn(e)
	}
}

func (self *lab) GetData(name string) ([]*experiment.Sample, error) {
	self.reload()
	for _, e := range self.loaded {
		if e.GetGuid() == name {
			return e.GetData()
		}
	}

	return nil, nil
}
