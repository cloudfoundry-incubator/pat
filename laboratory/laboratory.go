package laboratory

import (
	"github.com/cloudfoundry-community/pat/experiment"
	"github.com/nu7hatch/gouuid"
)

type lab struct {
	store   Store
	running []experiment.Experiment
}

type Laboratory interface {
	Run(ex Runnable) (experiment.Experiment, error)
	RunWithHandlers(ex Runnable, fns []func(samples <-chan *experiment.Sample)) (experiment.Experiment, error)
	Visit(fn func(ex experiment.Experiment))
	GetData(name string) ([]*experiment.Sample, error)
}

type Runnable interface {
	Run(handler func(samples <-chan *experiment.Sample)) error
}

type Store interface {
	Writer(guid string) func(samples <-chan *experiment.Sample)
	LoadAll() ([]experiment.Experiment, error)
}

func NewLaboratory(history Store) Laboratory {
	lab := &lab{history, make([]experiment.Experiment, 0)}
	lab.reload()
	return lab
}

func (self *lab) reload() {
	self.running, _ = self.store.LoadAll()
}

func (self *lab) Run(ex Runnable) (experiment.Experiment, error) {
	return self.RunWithHandlers(ex, make([]func(<-chan *experiment.Sample), 0))
}

func (self *lab) RunWithHandlers(ex Runnable, additionalHandlers []func(<-chan *experiment.Sample)) (experiment.Experiment, error) {
	guid, _ := uuid.NewV4()
	buffered := &buffered{guid.String(), make([]*experiment.Sample, 0)}
	handlers := make([]func(<-chan *experiment.Sample), 2)
	handlers[0] = self.store.Writer(guid.String())
	handlers[1] = func(samples <-chan *experiment.Sample) {
		self.buffer(buffered, samples)
	}
	for _, h := range additionalHandlers {
		handlers = append(handlers, h)
	}
	go ex.Run(Multiplexer(handlers).Multiplex)
	return buffered, nil
}

func (self *lab) Visit(fn func(ex experiment.Experiment)) {
	for _, e := range self.running {
		fn(e)
	}
}

func (self *lab) GetData(name string) ([]*experiment.Sample, error) {
	for _, e := range self.running {
		if e.GetGuid() == name {
			return e.GetData()
		}
	}

	return nil, nil
}
