package laboratory

import "github.com/cloudfoundry-community/pat/experiment"

// An in-memory buffer so that the currently running experiment
// can be served in-memory rather than round-tripping to the data
// store
type buffered struct {
	name    string
	samples []*experiment.Sample
}

func (self *lab) buffer(buffered *buffered, samples <-chan *experiment.Sample) {
	self.running = append(self.running, buffered)
	for s := range samples {
		buffered.samples = append(buffered.samples, s)
	}
}

func (b *buffered) GetGuid() string {
	return b.name
}

func (b *buffered) GetData() ([]*experiment.Sample, error) {
	return b.samples, nil
}
