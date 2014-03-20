package laboratory

import (
	. "github.com/julz/pat/experiment"
	"github.com/nu7hatch/gouuid"
)

type lab struct {
	store   Store
	running []Experiment
}

type Laboratory interface {
	Run(ex Runnable) (Experiment, error)
	RunWithHandlers(ex Runnable, fns []func(samples <-chan *Sample)) (Experiment, error)
	Visit(fn func(ex Experiment))
	GetData(name string) ([]*Sample, error)
}

type Runnable interface {
	Run(handler func(samples <-chan *Sample)) error
}

type Store interface {
	Writer(guid string) func(samples <-chan *Sample)
	LoadAll() ([]Experiment, error)
}

func NewLaboratory(history Store) Laboratory {
	lab := &lab{history, make([]Experiment, 0)}
	lab.reload()
	return lab
}

func (self *lab) reload() {
	self.running, _ = self.store.LoadAll()
}

func (self *lab) Run(ex Runnable) (Experiment, error) {
	return self.RunWithHandlers(ex, make([]func(<-chan *Sample), 0))
}

func (self *lab) RunWithHandlers(ex Runnable, additionalHandlers []func(<-chan *Sample)) (Experiment, error) {
	guid, _ := uuid.NewV4()
	buffered := &buffered{guid.String(), make([]*Sample, 0)}
	handlers := make([]func(<-chan *Sample), 2)
	handlers[0] = self.store.Writer(guid.String())
	handlers[1] = func(samples <-chan *Sample) {
		self.buffer(buffered, samples)
	}
	for _, h := range additionalHandlers {
		handlers = append(handlers, h)
	}
	go ex.Run(Multiplexer(handlers).Multiplex)
	return buffered, nil
}

func (self *lab) Visit(fn func(ex Experiment)) {
	for _, e := range self.running {
		fn(e)
	}
}

func (self *lab) GetData(name string) ([]*Sample, error) {
	for _, e := range self.running {
		if e.GetGuid() == name {
			return e.GetData()
		}
	}

	return nil, nil
}
