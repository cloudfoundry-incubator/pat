package output

import (
	"github.com/julz/pat/experiment"
)

type Multiplexer []func(chan *experiment.Sample)

func (out Multiplexer) Multiplex(in chan *experiment.Sample) {
	channels := make([]chan *experiment.Sample, 0)
	for _, f := range out {
		ch := make(chan *experiment.Sample)
		channels = append(channels, ch)
		go f(ch)
	}
	for i := range in {
		for _, o := range channels {
			o <- i
		}
	}
}
