package laboratory

import (
	. "github.com/julz/pat/experiment"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Laboratory", func() {
	Describe("Running an experiment", func() {
		var (
			lab             Laboratory
			store           *dummyStore
			experiment1     Runnable
			experiment2     Runnable
			run1            Experiment
			run2            Experiment
			handlerRecieved []*Sample
		)

		BeforeEach(func() {
			store = &dummyStore{make(map[string][]*Sample), make([]Experiment, 0)}
		})

		JustBeforeEach(func() {
			lab = NewLaboratory(store)
			experiment1 = &dummyExperiment{"1", []*Sample{&Sample{}}}
			experiment2 = &dummyExperiment{"2", []*Sample{&Sample{}, &Sample{}}}
			run1, _ = lab.Run(experiment1)
			handlerRecieved = nil
			run2, _ = lab.RunWithHandlers(experiment2, []func(<-chan *Sample){func(samples <-chan *Sample) {
				for s := range samples {
					handlerRecieved = append(handlerRecieved, s)
				}
			}})

			Eventually(func() int {
				got := make([]Experiment, 0)
				lab.Visit(func(e Experiment) {
					got = append(got, e)
				})
				return len(got)
			}).Should(Equal(2))
		})

		It("lists running experiments", func() {
			got := make([]Experiment, 0)
			lab.Visit(func(e Experiment) {
				got = append(got, e)
			})

			Ω(got[0]).Should(Equal(run1))
			Ω(got[1]).Should(Equal(run2))
		})

		It("gives experiments a GUID", func() {
			Ω(run1.GetGuid()).ShouldNot(Equal("1"))
			Ω(run1.GetGuid()).ShouldNot(Equal(run2.GetGuid()))
		})

		It("saves running experiments to the store", func() {
			Ω(store.store).Should(HaveLen(2))
			Ω(store.store[run1.GetGuid()]).Should(HaveLen(3))
			Ω(store.store[run2.GetGuid()]).Should(HaveLen(3))
		})

		It("retrieves data from a running experiment (by buffering in memory)", func() {
			Ω(data(lab.GetData(run1.GetGuid()))).Should(HaveLen(3))
			Ω(data(lab.GetData(run2.GetGuid()))).Should(HaveLen(3))
		})

		It("sends data from a running experiment to a handler function", func() {
			Ω(handlerRecieved).Should(HaveLen(3))
		})

		Describe("Loading previous experiment at startup", func() {
			var (
				loadedExperiment1 Experiment
				loadedExperiment2 Experiment
				loadedData1       []*Sample
				loadedData2       []*Sample
			)

			BeforeEach(func() {
				loadedData1 = []*Sample{&Sample{}, &Sample{}}
				loadedData2 = []*Sample{&Sample{}, &Sample{}, &Sample{}}
				loadedExperiment1 = &dummyExperiment{"load1", loadedData1}
				loadedExperiment2 = &dummyExperiment{"load2", loadedData2}
				store.previous = append(store.previous, loadedExperiment1)
				store.previous = append(store.previous, loadedExperiment2)
			})

			It("retrieves data from a loaded experiment (by calling GetData())", func() {
				Ω(data(lab.GetData("load1"))).Should(HaveLen(len(loadedData1)))
				Ω(data(lab.GetData("load2"))).Should(HaveLen(len(loadedData2)))
			})

			It("lists saved and running experiments", func() {
				var got []Experiment
				Eventually(func() []Experiment {
					got = make([]Experiment, 0)
					lab.Visit(func(e Experiment) {
						got = append(got, e)
					})
					return got
				}).Should(HaveLen(4))
				Ω(got[0]).Should(Equal(loadedExperiment1))
				Ω(got[1]).Should(Equal(loadedExperiment2))
				Ω(got[2].GetGuid()).Should(Equal(run1.GetGuid()))
				Ω(got[3].GetGuid()).Should(Equal(run2.GetGuid()))
			})
		})
	})
})

func data(s []*Sample, e error) []*Sample {
	Ω(e).ShouldNot(HaveOccurred())
	return s
}

type dummyStore struct {
	store    map[string][]*Sample
	previous []Experiment
}

type dummyExperiment struct {
	name string
	data []*Sample
}

func (store *dummyStore) Writer(guid string) func(samples <-chan *Sample) {
	return func(samples <-chan *Sample) {
		for s := range samples {
			store.store[guid] = append(store.store[guid], s)
		}
	}
}

func (store *dummyStore) LoadAll() ([]Experiment, error) {
	return store.previous, nil
}

func (e *dummyExperiment) Run(fn func(samples <-chan *Sample)) error {
	ch := make(chan *Sample)
	done := make(chan bool)
	go func() {
		fn(ch)
		done <- true
	}()
	ch <- &Sample{}
	ch <- &Sample{}
	ch <- &Sample{}
	close(ch)
	<-done
	return nil
}

func (e *dummyExperiment) GetData() ([]*Sample, error) {
	return e.data, nil
}

func (e *dummyExperiment) GetGuid() string {
	return e.name
}
